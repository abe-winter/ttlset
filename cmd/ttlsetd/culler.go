package main

import "time"

// periodic worker that culls stale items in TtlSets and removes empty TtlSets from global SETS
func cullSets() {
  empties := func () []string {
    empties := make([]string, 0)
    // todo: locking here is bad. this blocks creation of new sets for the whole iteration + cull
    // need to either copy the map or periodically unlock to drain lock queue
    sets_lock.RLock()
    defer sets_lock.RUnlock()
    for k, v := range SETS {
      v.Cull(time.Now())
      if v.Len() == 0 {
        empties = append(empties, k)
      }
    }
    return empties
  }()

  // todo: locking too aggressive here
  sets_lock.Lock()
  defer sets_lock.Unlock()
  for _, key := range empties {
    // todo: race condition. (Len, delete) is a compare-and-swap operation. Needs to be locked
    if ts, found := SETS[key]; found && ts.Len() == 0 {
      delete(SETS, key)
    }
  }
}

// start a ticker and call cullSets() periodically
func startCuller() {
  // note: documented ticker behavior is to skip missed intervals this is right for us
  // todo: measure cull time + warn if > tickerInterval / 10
  ticker := time.NewTicker(tickerInterval)
  go func() {
    for range ticker.C {
      cullSets() // todo: catch errors + report
    }
  }()
}
