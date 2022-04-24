package ttlset

import "hash/adler32"
import "sync"
import "time"
import "github.com/emirpasic/gods/trees/btree"

var keyMutexes []sync.Mutex = make([]sync.Mutex, 50)

// stores details about each key in the set
type TtlVal struct {
  t time.Time
}

type TtlSet struct {
  Ttl time.Duration
  Byval map[string]TtlVal
  Bytime *btree.Tree
  valLock sync.RWMutex
  timeLock sync.Mutex
}

// helper for btree
func timeComparator(a, b interface{}) int {
  ta := a.(time.Time)
  tb := b.(time.Time)
  if ta.Equal(tb) {
    return 0
  } else if ta.Before(tb) {
    return -1
  } else {
    return 1
  }
}

func New() *TtlSet {
  return &TtlSet{
    Ttl: defaultTtl,
    // note: TtlVal is a value rather than a pointer here because GC is faster per https://www.komu.engineer/blogs/go-gc-maps
    Byval: make(map[string]TtlVal),
    Bytime: btree.NewWith(treeOrder, timeComparator),
  }
}

// return a lock specific to the key and its has collisions.
// this protects against a race condition, without blocking all concurrency.
// race condition is dupes in the Bytime tree.
func getMutex(key string, mutexes []sync.Mutex) *sync.Mutex {
  return &mutexes[adler32.Checksum([]byte(key)) % uint32(len(mutexes))]
}

// btree helper, gets entry from within node
func GetEntry(tree *btree.Tree, key time.Time) *btree.Entry {
  // todo: 1.18 generics
  if node := tree.GetNode(key); node != nil {
    for _, entry := range node.Entries {
      if entry.Key.(time.Time).Equal(key) {
        return entry
      }
    }
  }
  return nil
}

// remove key from slice at Bytime[oldTime]
// MUST hold timeLock outside or you risk race
func (ts *TtlSet) rmTreeKey(key string, oldTime time.Time) {
  if entry := GetEntry(ts.Bytime, oldTime); entry != nil {
    removed := sRemove(key, entry.Value.([]string))
    if len(removed) == 0 {
      ts.Bytime.Remove(oldTime)
    } else {
      entry.Value = removed
    }
  } else {
    // this shouldn't happen, but also make it an http error so the client knows what's up
    panic("todo: recover from not found")
  }
}

// add key to slice at Bytime[now]
// MUST hold timeLock outside or you risk race
func (ts *TtlSet) addTreeKey(key string, now time.Time) {
  if entry := GetEntry(ts.Bytime, now); entry != nil {
    entry.Value = append(entry.Value.([]string), key)
  } else {
    ts.Bytime.Put(now, []string{key})
  }
}


// safely RUnlock a RWMutex ugh why is this necessary
type CancelableLocker struct {
  Locked bool
  locker sync.RWMutex // todo 1.18 make this generic ugh
}

// factory for CancelableLocker
func cancelableLock(locker sync.RWMutex) CancelableLocker {
  locker.Lock()
  return CancelableLocker{Locked: true, locker: locker}
}

func (cl *CancelableLocker) Unlock() {
  if cl.Locked {
    cl.locker.Unlock()
    cl.Locked = false
  }
}

// add key to TtlSet. returns (existed, prevTime)
func (ts *TtlSet) Add(key string, now time.Time) (bool, time.Time) {
  mutex := getMutex(key, keyMutexes)
  mutex.Lock()
  defer mutex.Unlock()
  oldTime := time.Time{}

  elem, found := func () (TtlVal, bool) {
    ts.valLock.Lock()
    defer ts.valLock.Unlock()
    elem, found := ts.Byval[key]
    ts.Byval[key] = TtlVal{t: now}
    return elem, found
  }()

  ts.timeLock.Lock()
  defer ts.timeLock.Unlock()
  if found {
    oldTime = elem.t
    ts.rmTreeKey(key, oldTime)
  }
  ts.addTreeKey(key, now)
  return found, oldTime
}

// remove key. return (removed, prevTime)
func (ts *TtlSet) Remove(key string, cullMode bool, cullCutoff time.Time) (bool, time.Time) {
  mutex := getMutex(key, keyMutexes)
  mutex.Lock()
  defer mutex.Unlock()

  // note: could do read-then-write lock, but:
  // 1) there's no read-only case in this function
  // 2) not sure how much concurrency the read-only time opens up
  canceler := cancelableLock(ts.valLock)
  defer canceler.Unlock()

  if elem, ok := ts.Byval[key]; ok {
    // exists case
    oldTime := elem.t
    if cullMode && oldTime.After(cullCutoff) {
      return false, oldTime
    }
    delete(ts.Byval, key)
    canceler.Unlock()
    ts.timeLock.Lock()
    defer ts.timeLock.Unlock()
    ts.rmTreeKey(key, oldTime)
    return true, oldTime
  }
  return false, time.Time{}
}

// length of TtlSet aka # of keys
func (ts *TtlSet) Len() int {
  // todo: is the read lock necessary here or is len(map) safe?
  // if not, don't need an RWMutex
  ts.valLock.RLock()
  defer ts.valLock.RUnlock()
  return len(ts.Byval)
}

// remove keys from TtlSet that are older than ttl
func (ts *TtlSet) Cull(now time.Time) int {
  cutoff := now.Add(-ts.Ttl)
  candidates := make([][]string, 0)
  func () {
    ts.timeLock.Lock()
    defer ts.timeLock.Unlock()
    iter := ts.Bytime.Iterator()
    for ; iter.Next(); {
      if iter.Key().(time.Time).After(cutoff) { break }
      candidates = append(candidates, iter.Value().([]string))
    }
  }()
  nculled := 0
  for _, slice := range candidates {
    for _, key := range slice {
      // todo: lock around large chunks instead of inside Remove
      removed, _ := ts.Remove(key, true, cutoff)
      if removed { nculled += 1 }
    }
  }
  return nculled
}
