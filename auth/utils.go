package auth

import "context"

func AuthenticatedUserFromContext(ctx context.Context) string {
	user, ok := ctx.Value(AuthenticatedUser).(string)
	if !ok {
		return UnauthenticatedUser
	}

	return user
}
