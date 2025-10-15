package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-contrib/authz"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Authz(adapter persist.Adapter, defaultPolicies, defaultGroups [][]string) gin.HandlerFunc {
	modelContent := `
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
		m = ug(r.sub, p.sub) && rg(r.obj, p.obj) && (r.act == p.act || p.act == "*")
		`

	m, err := model.NewModelFromString(modelContent)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create model from string")
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create enforcer")
	}

	policies, err := enforcer.GetPolicy()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get existing policies")
	}

	if len(policies) == 0 && len(defaultPolicies) > 0 {
		// No existing policies, load default policies
		_, err = enforcer.AddPolicies(defaultPolicies)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to add default policies")
		} else {
			log.Info().Msg("No plicies found. Added default policies")
		}
	}

	groups, err := enforcer.GetGroupingPolicy()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get existing groups")
	}

	if len(groups) == 0 && len(defaultGroups) > 0 {
		// No existing groups, load default groups
		_, err = enforcer.AddGroupingPolicies(defaultGroups)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to add default groups")
		} else {
			log.Info().Msg("No groups found. Added default groups")
		}
	}

	return authz.NewAuthorizer(enforcer)
}
