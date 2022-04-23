package main

import "time"
import "github.com/gin-gonic/gin"
import "github.com/abe-winter/ttlset/m/v2/pkg/ttlset"

var SETS map[string]ttlset.TtlSet = make(map[string]ttlset.TtlSet)

// periodic worker that culls stale items in TtlSets and removes empty TtlSets from global SETS
func cullSets() {
  empties := make([]string, 0)
  for k, v := range SETS {
    v.Cull(time.Now())
    if v.Len() == 0 {
      empties = append(empties, k)
    }
  }
  for _, key := range empties {
    // rechecking len to mostly prevent race, but actually need a RWMutex probably
    if ts, found := SETS[key]; found && ts.Len() == 0 {
      delete(SETS, key)
    }
  }
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

  r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

  r.POST("/set/:key/item/:val", func(c *gin.Context) {
    panic("notimp")
  })

  r.PATCH("/set/:key/ttl", func(c *gin.Context) {
    panic("notimp")
  })

  r.GET("/set/:key/count", func(c *gin.Context) {
    panic("notimp")
  })

  r.DELETE("/set/:key", func(c *gin.Context) {
    panic("notimp")
  })

  r.DELETE("/set/:key/item/:val", func(c *gin.Context) {
    panic("notimp")
  })

  startCuller()
  r.Run()
}
