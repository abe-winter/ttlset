package main

import "time"
import "sync"
import "github.com/gin-gonic/gin"
import "github.com/abe-winter/ttlset/m/v2/pkg/ttlset"

// todo: move SETS and associated logic to pkg or internal
var SETS map[string]*ttlset.TtlSet = make(map[string]*ttlset.TtlSet)
var sets_lock sync.RWMutex = sync.RWMutex{}

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

// map lookup with lock
func getSet(key string, create_missing bool) *ttlset.TtlSet {
  set := func () *ttlset.TtlSet {
    sets_lock.RLock()
    defer sets_lock.RUnlock()
    return SETS[key]
  }()
  if set == nil && create_missing {
    sets_lock.Lock()
    defer sets_lock.Unlock()
    set = SETS[key] // i.e. check that someone else hasn't created it
    if set == nil {
      set = ttlset.New()
      SETS[key] = set
    }
  }
  return set
}

func startCuller() {
  ticker := time.NewTicker(time.Second * 10) // todo: durt from config
  go func() {
    for range ticker.C {
      cullSets() // todo: catch errors + report
    }
  }()
}

func main() {
	r := gin.Default()

  // add item to set (or update its last-seen time)
  r.POST("/set/:key/item/:item", func(c *gin.Context) {
    // todo: make sure these params get unescaped
    set := getSet(c.Param("key"), true)
    existed, prevTime := set.Add(c.Param("item"), time.Now())
    c.JSON(200, gin.H{"key_existed": existed, "prevTime": prevTime})
  })

  // set TTL on set
  r.PATCH("/set/:key/ttl", func(c *gin.Context) {
    // huge design flaw here; when the set is culled, it loses its ttl setting. this needs thought
    query := struct {
      TtlSeconds uint
    }{}
    // todo: check min / max
    c.BindQuery(&query)
    set := getSet(c.Param("key"), false)
    if set == nil {
      c.String(404, "set not found")
    } else {
      set.Ttl = time.Duration(query.TtlSeconds) * time.Second
      c.String(200, "ok")
    }
  })

  // get size of set
  r.GET("/set/:key/count", func(c *gin.Context) {
    set := getSet(c.Param("key"), false)
    count := 0
    if set != nil { count = set.Len() }
    c.JSON(200, gin.H{"exists": set != nil, "count": count})
  })

  // remove entire set
  r.DELETE("/set/:key", func(c *gin.Context) {
    sets_lock.Lock()
    defer sets_lock.Unlock()
    delete(SETS, c.Param("key"))
    c.String(200, "ok")
  })

  // remove just an item
  r.DELETE("/set/:key/item/:item", func(c *gin.Context) {
    set := getSet(c.Param("key"), false)
    removed, prevTime := false, time.Time{}
    if set != nil {
      removed, prevTime = set.Remove(c.Param("item"), false, time.Time{})
    }
    c.JSON(200, gin.H{"set_exists": set != nil, "removed": removed, "prevTime": prevTime})
  })

  startCuller()
  r.Run()
}
