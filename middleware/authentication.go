package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	UnauthenticatedUser = "__unauthenticated__"
)

type BasicAuthenticator func(username, password string) (string, error)

func Authentication(basicAuthAuthenticator BasicAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasBasicAuth := c.Request.BasicAuth()

		authenticatedUser := ""
		authorizationFailed := false
		if hasBasicAuth {
			// BasicAuth provided, validate it
			userID, err := basicAuthAuthenticator(username, password)

			if err == nil {
				authenticatedUser = userID
			} else {
				authorizationFailed = true
			}
		} else {
			// No authorization provided continue as anonymous user
			authenticatedUser = UnauthenticatedUser
		}

		c.Request.SetBasicAuth(authenticatedUser, "")

		if authorizationFailed {
			// Authentication failed. Return 401 and abort the request.
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Next()
		}
	}
}
