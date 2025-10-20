package middleware

import (
	"context"
	"net/http"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type BasicAuthenticator func(ctx context.Context, username, password string) (string, error)

func Authentication(basicAuthAuthenticator BasicAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, hasBasicAuth := c.Request.BasicAuth()

		authenticatedUser := ""
		authorizationFailed := false
		if hasBasicAuth {
			log.Debug().
				Str("user", username).
				Msg("Authenticating user with BasicAuth")
			// BasicAuth provided, validate it
			userID, err := basicAuthAuthenticator(c.Request.Context(), username, password)

			if err == nil {
				authenticatedUser = userID
			} else {
				authorizationFailed = true
			}
		} else {
			// No authorization provided continue as anonymous user
			log.Debug().Msg("No authentication provided. Proceeding as unauthenticated user")
			authenticatedUser = auth.UnauthenticatedUser
		}

		c.Request.SetBasicAuth(authenticatedUser, "")
		c.Request = c.Request.WithContext(
			context.WithValue(
				c.Request.Context(),
				auth.AuthenticatedUser,
				authenticatedUser,
			),
		)

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
