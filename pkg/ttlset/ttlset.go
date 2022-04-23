package ttlset

import "hash/adler32"
import "sync"
import "time"
import "github.com/emirpasic/gods/trees/btree"

var defaultTtl time.Duration = time.Minute // todo: from config
var defaultOrder int = 3 // todo: from config
// for locking without being single-threaded
var keyMutexes []sync.Mutex = make([]sync.Mutex, 50)
var treeMutexes []sync.Mutex = make([]sync.Mutex, 50)

// stores details about each key in the set
type TtlVal struct {
  t time.Time
}

type TtlSet struct {
  Ttl time.Duration
  Byval map[string]TtlVal
  Bytime *btree.Tree
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

func New() TtlSet {
  return TtlSet{
    Ttl: defaultTtl,
    Byval: make(map[string]TtlVal),
    Bytime: btree.NewWith(defaultOrder, timeComparator),
  }
}

// return a lock specific to the key and its has collisions.
// this protects against a race condition, without blocking all concurrency.
// race condition is dupes in the Bytime tree.
func getMutex(key string, mutexes []sync.Mutex) *sync.Mutex {
  return &mutexes[adler32.Checksum([]byte(key)) % uint32(len(mutexes))]
}

// remove key from slice at Bytime[oldTime] (with locking)
func (ts *TtlSet) rmTreeKey(key string, oldTime time.Time) {
  mutex := getMutex(oldTime.String(), treeMutexes)
  mutex.Lock()
  defer mutex.Unlock()
  if slice, found := ts.Bytime.Get(oldTime); found {
    removed := sRemove(key, slice.([]string))
    if len(removed) == 0 {
      ts.Bytime.Remove(oldTime)
    } else {
      ts.Bytime.Put(oldTime, removed)
    }
  } else {
    // this shouldn't happen
    panic("todo: recover from not found")
  }
}

// add key to slice at Bytime[now] (with locking)
func (ts *TtlSet) addTreeKey(key string, now time.Time) {
  mutex := getMutex(now.String(), treeMutexes)
  mutex.Lock()
  defer mutex.Unlock()
  if slice, found := ts.Bytime.Get(now); found {
    ts.Bytime.Put(now, append(slice.([]string), key))
  } else {
    ts.Bytime.Put(now, []string{key})
  }
}

// add key to TtlSet. returns (existed, prevTime)
func (ts *TtlSet) Add(key string, now time.Time) (bool, time.Time) {
  mutex := getMutex(key, keyMutexes)
  mutex.Lock()
  defer mutex.Unlock()
  if elem, ok := ts.Byval[key]; ok {
    // exists case
    oldTime := elem.t
    if !oldTime.Equal(now) {
      ts.rmTreeKey(key, oldTime)
      ts.addTreeKey(key, now)
    }
    return true, oldTime
  } else {
    ts.Byval[key] = TtlVal{t: now}
    ts.addTreeKey(key, now)
    return false, time.Time{}
  }
}

// remove key. return (existed, prevTime)
func (ts *TtlSet) Remove(key string, now time.Time) (bool, time.Time) {
  mutex := getMutex(key, keyMutexes)
  mutex.Lock()
  defer mutex.Unlock()
  if elem, ok := ts.Byval[key]; ok {
    // exists case
    oldTime := elem.t
    ts.rmTreeKey(key, oldTime)
    delete(ts.Byval, key)
    return true, oldTime
  }
  return false, time.Time{}
}

// length of TtlSet aka # of keys
func (ts *TtlSet) Len() int {
  return len(ts.Byval)
}

// remove keys from TtlSet that are older than ttl
func (ts *TtlSet) Cull(now time.Time) {
  // cutoff := ts.ttl
  panic("notimp")
}
