package middleware

import "github.com/gin-gonic/gin"

const (
	UnauthenticatedUser = "__unauthenticated__"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, ok := c.Request.BasicAuth()
		if ok {
			// Authorize using basic auth
		} else {
			// No auhthorization provided continue as anonymous user
			c.Request.SetBasicAuth(UnauthenticatedUser, "")
		}

		c.Next()
	}
}