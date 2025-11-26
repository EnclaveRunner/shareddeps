package auth

import (
	"context"
)

type contextKey string

const authenticatedUser contextKey = "authenticatedUser"

func SetAuthenticatedUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, authenticatedUser, user)
}

func GetAuthenticatedUser(ctx context.Context) string {
	user, ok := ctx.Value(authenticatedUser).(string)
	if !ok {
		return UnauthenticatedUser
	}

	return user
}
