package main

import "github.com/gin-gonic/gin"

func authHeader(c *gin.Context) string {
  slice := c.Request.Header["Ttlset-Auth"] // yes gin capitalizes it
  if len(slice) > 0 { return slice[0] }
  return ""
}

// middleware for >= role
func RequireRole(minRole Role) gin.HandlerFunc {
  return func(c *gin.Context) {
    token := authHeader(c)
    if len(token) == 0 {
      c.String(401, "missing ttlset-auth header")
      c.Abort()
    } else if acct, ok := accounts[token]; ok && acct.Role >= minRole {
      c.Next()
    } else if ok {
      c.String(403, "requires greater permission")
      c.Abort()
    } else {
      c.String(401, "unknown account")
      c.Abort()
    }
  }
}
