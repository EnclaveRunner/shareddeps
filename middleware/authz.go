package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-contrib/authz"
	"github.com/gin-gonic/gin"
)

func Authz(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return authz.NewAuthorizer(enforcer)
}
