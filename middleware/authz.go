package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-contrib/authz"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Authz(adapter persist.Adapter) gin.HandlerFunc {
	modelContent :=
		`
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[role_definition]
		ug = _, _
		rg = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = ug(r.sub, p.sub) && rg(r.obj, p.obj) && r.act == p.act
		`

	m, err := model.NewModelFromString(modelContent)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create model from string")
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create enforcer")
	}

	return authz.NewAuthorizer(enforcer)
}
