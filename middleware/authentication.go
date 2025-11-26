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
			userID, err := basicAuthAuthenticator(
				c.Request.Context(),
				username,
				password,
			)

			if err == nil {
				authenticatedUser = userID
			} else {
				log.Debug().Err(err).Msg("Basic authentication failed")
				authorizationFailed = true
			}
		} else {
			// No authorization provided continue as anonymous user
			log.Debug().Msg("No authentication provided. Proceeding as unauthenticated user")
			authenticatedUser = auth.UnauthenticatedUser
		}

		auth.SetAuthenticatedUser(c, authenticatedUser)

		if authorizationFailed {
			// Authentication failed. Return 401 and abort the request.
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.Next()
		}
	}
}
