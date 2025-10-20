package auth

import (
	"slices"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/rs/zerolog/log"
)

var enforcer *casbin.Enforcer

// InitAuth initializes the casbin enforcer with the provided adapter and sets
// up default policies. It creates the casbin model, loads policies, and ensures
// the enclaveAdmin group and policy exist.
func InitAuth(adapter persist.Adapter) *casbin.Enforcer {
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
		m = (g(r.sub, p.sub) || p.sub == "*") && (g2(r.obj, p.obj) || p.obj == "*") && (r.act == p.act || p.act == "*")
	`

	m, err := model.NewModelFromString(modelContent)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create casbin model from string")
	}

	enforcer, err = casbin.NewEnforcer(m, adapter)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create casbin enforcer")
	}

	err = enforcer.LoadPolicy()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load casbin policy")
	}

	policies, err := enforcer.GetPolicy()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get casbin policies")
	}

	containsAdminPolicy := slices.IndexFunc(
		policies,
		func(policy []string) bool {
			return policy[0] == enclaveAdminGroup && policy[1] == "*" &&
				policy[2] == "*"
		},
	) != -1

	if !containsAdminPolicy {
		_, err = enforcer.AddPolicy(enclaveAdminGroup, "*", "*")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to add enclaveAdmin casbin policy")
		}
		log.Info().Msg("Added enclaveAdmin casbin policy")
	}

	userGroups, err := enforcer.GetNamedGroupingPolicy("g")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get casbin user groups")
	}

	containsAdminGroup := slices.IndexFunc(
		userGroups,
		func(group []string) bool {
			return group[1] == enclaveAdminGroup
		},
	) != -1

	if !containsAdminGroup {
		_, err = enforcer.AddNamedGroupingPolicy("g", nullUser, enclaveAdminGroup)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to add admin to enclaveAdmin casbin group")
		}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to save casbin policy")
	}

	log.Debug().Msg("Casbin enforcer initialized")

	return enforcer
}
