package auth

type Policy struct {
	UserGroup     string
	ResourceGroup string
	Permission    string
}

// AddPolicy adds a policy to the enforcer if it does not already exist.
//
// It checks if the user group and resource group exist before adding the policy
// and throws if they
// do not.
func (auth *AuthModule) AddPolicy(
	userGroup, resourceGroup, method string,
) error {
	if userGroup != "*" {
		ugExists, err := auth.UserGroupExists(userGroup)
		if err != nil {
			return err
		}
		if !ugExists {
			return &NotFoundError{"userGroup", userGroup}
		}
	}

	if resourceGroup != "*" {
		rgExists, err := auth.ResourceGroupExists(resourceGroup)
		if err != nil {
			return err
		}
		if !rgExists {
			return &NotFoundError{"resourceGroup", resourceGroup}
		}
	}

	filteredPolicies, err := auth.enforcer.GetFilteredPolicy(
		0,
		userGroup,
		resourceGroup,
		method,
	)
	if err != nil {
		return &CasbinError{"GetFilteredPolicy", err}
	}
	if len(filteredPolicies) > 0 {
		return nil
	}

	_, err = auth.enforcer.AddPolicy(userGroup, resourceGroup, method)
	if err != nil {
		return &CasbinError{"AddPolicy", err}
	}

	err = auth.enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"AddPolicy", err}
	}

	return nil
}

func (auth *AuthModule) ListPolicies() ([]Policy, error) {
	rawPolicies, err := auth.enforcer.GetPolicy()
	if err != nil {
		return nil, &CasbinError{"GetPolicy", err}
	}

	policies := make([]Policy, 0, len(rawPolicies))
	for _, rawPolicy := range rawPolicies {
		policy := Policy{
			UserGroup:     rawPolicy[0],
			ResourceGroup: rawPolicy[1],
			Permission:    rawPolicy[2],
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// RemovePolicy removes a policy from the enforcer.
//
// It prevents the removal of the enclaveAdmin policy to ensure that
// enclaveAdmins always have full
// access.
func (auth *AuthModule) RemovePolicy(
	userGroup, resourceGroup, method string,
) error {
	if userGroup == enclaveAdminGroup && resourceGroup == "*" && method == "*" {
		return &ConflictError{"The provided policy cannot be removed"}
	}

	_, err := auth.enforcer.RemovePolicy(userGroup, resourceGroup, method)
	if err != nil {
		return &CasbinError{"RemovePolicy", err}
	}

	err = auth.enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"RemovePolicy", err}
	}

	return nil
}
