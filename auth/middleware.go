package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (auth *AuthModule) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetAuthenticatedUser(c.Request.Context())
		method := c.Request.Method
		path := c.Request.URL.Path

		allowed, err := auth.enforcer.Enforce(user, path, method)
		if err != nil {
			log.Error().Err(err).Msg("Authorization check failed")
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}
		if !allowed {
			log.Warn().
				Str("user", user).
				Str("path", path).
				Str("method", method).
				Msg("Unauthorized access attempt")
			c.AbortWithStatus(http.StatusForbidden)

			return
		}
		c.Next()
	}
}
