package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

func InsertAuthenticatedUser(c *gin.Context, user string) {
	c.Set(ContextKeyAuthenticatedUser, user)
}

func RetrieveAuthenticatedUser(ctx context.Context) string {
	user, ok := ctx.Value(ContextKeyAuthenticatedUser).(string)
	if !ok {
		return UnauthenticatedUser
	}

	return user
}
