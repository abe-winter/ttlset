package main

import "github.com/gin-gonic/gin"
import "github.com/abe-winter/ttlset/m/v2/pkg/ttlset"

var SETS map[string]ttlset.TtlSet = make(map[string]ttlset.TtlSet)

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

  r.Run()
}
