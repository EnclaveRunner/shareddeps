package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gin-contrib/authz"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Authz(
	adapter persist.Adapter,
	defaultPolicies, defaultUserGroups, defaultRessourceGroups [][]string,
) gin.HandlerFunc {
	modelContent := `
		[request_definition]
		r = sub, obj, act

		[policy_definition]
		p = sub, obj, act

		[role_definition]
		g = _, _
		g2 = _, _

		[policy_effect]
		e = some(where (p.eft == allow))

		[matchers]
		m = g(r.sub, p.sub) && g2(r.obj, p.obj) && (r.act == p.act || p.act == "*")
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

	userGroups, err := enforcer.GetNamedGroupingPolicy("g")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get existing user groups")
	}

	if len(userGroups) == 0 && len(defaultUserGroups) > 0 {
		// No existing groups, load default groups
		_, err = enforcer.AddNamedGroupingPolicies("g", defaultUserGroups)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to add default user groups")
		} else {
			log.Info().Msg("No groups found. Added default user groups")
		}
	}

	ressourceGroups, err := enforcer.GetNamedGroupingPolicy("g2")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get existing ressource groups")
	}

	if len(ressourceGroups) == 0 && len(defaultRessourceGroups) > 0 {
		// No existing groups, load default groups
		_, err = enforcer.AddNamedGroupingPolicies("g2", defaultRessourceGroups)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to add default ressource groups")
		} else {
			log.Info().Msg("No groups found. Added default ressource groups")
		}
	}

	err = enforcer.LoadPolicy()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load policies")
	}

	enforcer.EnableLog(true)

	rule, _ := enforcer.GetFilteredNamedGroupingPolicy("g", 0, "__unauthenticated__")
	for _, r := range rule {
		log.Debug().Strs("rule", r).Msg("__unauthenticated__ grouping policy")
	}

	rule, _ = enforcer.GetFilteredNamedGroupingPolicy("g2", 0, "/ready")
	for _, r := range rule {
		log.Debug().Strs("rule", r).Msg("/ready grouping policy")
	}

	result, _ := enforcer.Enforce("__unauthenticated__", "/ready", "GET")
	log.Debug().Bool("unauthenticated_ready", result).Msg("Unauthenticated access to /ready")

	return authz.NewAuthorizer(enforcer)
}
