package main

import "flag"
import "time"
import "sync"
import "github.com/gin-gonic/gin"
import "github.com/abe-winter/ttlset/m/v2/pkg/ttlset"

var version = "(unset)" // ldflags
var fVersion = flag.Bool("version", false, "print version and exit")

// todo: move SETS and associated logic to pkg or internal
var SETS map[string]*ttlset.TtlSet = make(map[string]*ttlset.TtlSet)
var sets_lock sync.RWMutex = sync.RWMutex{}

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

func main() {
  flag.Parse()

  if *fVersion {
    println(version)
    return
  }

  r := gin.Default()

  // add item to set (or update its last-seen time)
  r.POST("/set/:key/item/:item", RequireRole(ReadWrite), func(c *gin.Context) {
    // todo: make sure these params get unescaped
    set := getSet(c.Param("key"), true)
    existed, prevTime, newLen := set.Add(c.Param("item"), time.Now())
    c.JSON(200, gin.H{"key_existed": existed, "prevTime": prevTime, "newLen": newLen})
  })

  // set TTL on set
  r.PATCH("/set/:key/ttl", RequireRole(ReadWrite), func(c *gin.Context) {
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
  r.GET("/set/:key/count", RequireRole(ReadOnly), func(c *gin.Context) {
    set := getSet(c.Param("key"), false)
    count := 0
    if set != nil { count = set.Len() }
    c.JSON(200, gin.H{"exists": set != nil, "count": count})
  })

  // remove entire set
  r.DELETE("/set/:key", RequireRole(ReadWrite), func(c *gin.Context) {
    sets_lock.Lock()
    defer sets_lock.Unlock()
    delete(SETS, c.Param("key"))
    c.String(200, "ok")
  })

  // remove just an item
  r.DELETE("/set/:key/item/:item", RequireRole(ReadWrite), func(c *gin.Context) {
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
