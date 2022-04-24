package main

import "github.com/gin-gonic/gin"

func authHeader(c *gin.Context) string {
  slice := c.Request.Header["ttlset-auth"]
  if len(slice) > 0 { return slice[0] }
  return ""
}

// middleware for >= role
func RequireRole(minRole Role) gin.HandlerFunc {
  return func(c *gin.Context) {
    if acct, ok := accounts[authHeader(c)]; ok {
      if acct.Role >= minRole {
        c.Next()
      } else {
        c.String(403, "requires greater permission")
      }
    } else {
      c.String(401, "missing or unknown account")
    }
  }
}
