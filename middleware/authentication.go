package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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
			log.Debug().Str("user", username).Msg("Authenticating user with BasicAuth")
			// BasicAuth provided, validate it
			userID, err := basicAuthAuthenticator(username, password)

			if err == nil {
				authenticatedUser = userID
			} else {
				authorizationFailed = true
			}
		} else {
			// No authorization provided continue as anonymous user
			log.Debug().Msg("No authentication provided. Proceeding as unauthenticated user")
			authenticatedUser = UnauthenticatedUser
		}

		c.Request.SetBasicAuth(authenticatedUser, "")

		user, _, _ := c.Request.BasicAuth()
		method := c.Request.Method
		path := c.Request.URL.Path

		log.Debug().
			Str("user", user).
			Str("method", method).
			Str("path", path).
			Msg("Authentication middleware processed request")

		if authorizationFailed {
			// Authentication failed. Return 401 and abort the request.
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Next()
		}
	}
}
