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
func AddPolicy(userGroup, resourceGroup, method string) error {
	ugExists, err := UserGroupExists(userGroup)
	if err != nil {
		return err
	}
	if !ugExists {
		return &NotFoundError{"userGroup", userGroup}
	}

	rgExists, err := ResourceGroupExists(resourceGroup)
	if err != nil {
		return err
	}
	if !rgExists {
		return &NotFoundError{"resourceGroup", resourceGroup}
	}

	filteredPolicies, err := enforcer.GetFilteredPolicy(
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

	_, err = enforcer.AddPolicy(userGroup, resourceGroup, method)
	if err != nil {
		return &CasbinError{"AddPolicy", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"AddPolicy", err}
	}

	return nil
}

func ListPolicies() ([]Policy, error) {
	rawPolicies, err := enforcer.GetPolicy()
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
func RemovePolicy(userGroup, resourceGroup, method string) error {
	if userGroup == enclaveAdminGroup && resourceGroup == "*" && method == "*" {
		return &ConflictError{"The provided policy cannot be removed"}
	}

	_, err := enforcer.RemovePolicy(userGroup, resourceGroup, method)
	if err != nil {
		return &CasbinError{"RemovePolicy", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"RemovePolicy", err}
	}

	return nil
}
