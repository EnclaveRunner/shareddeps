package auth

// AddPolicy adds a policy to the enforcer if it does not already exist.
//
// It checks if the user group and resource group exist before adding the policy and throws if they
// do not.
func AddPolicy(userGroup, resourceGroup, method string) error {
	ugExists, err := UserGroupExists(userGroup)
	if err != nil {
		return err
	}
	if !ugExists {
		return makeErrUserGroupNotFound(userGroup)
	}

	rgExists, err := ResourceGroupExists(resourceGroup)
	if err != nil {
		return err
	}
	if !rgExists {
		return makeResourceGroupNotFoundError(resourceGroup)
	}

	filteredPolicies, err := enforcer.GetFilteredPolicy(0, userGroup, resourceGroup, method)
	if err != nil {
		return makeErrCasbinConnection("AddPolicy", err)
	}
	if len(filteredPolicies) > 0 {
		return nil
	}

	_, err = enforcer.AddPolicy(userGroup, resourceGroup, method)
	if err != nil {
		return makeErrCasbinConnection("AddPolicy", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("AddPolicy", err)
	}

	return nil
}

// RemovePolicy removes a policy from the enforcer.
//
// It prevents the removal of the enclaveAdmin policy to ensure that enclaveAdmins always have full
// access.
func RemovePolicy(userGroup, resourceGroup, method string) error {
	if userGroup == enclaveAdminGroup && resourceGroup == "*" && method == "*" {
		return errEnclaveAdminPolicy
	}

	_, err := enforcer.RemovePolicy(userGroup, resourceGroup, method)
	if err != nil {
		return makeErrCasbinConnection("RemovePolicy", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("RemovePolicy", err)
	}

	return nil
}
