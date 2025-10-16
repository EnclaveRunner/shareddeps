package auth

type ResourceGroup struct {
	ResourceName string
	GroupName    string
}

// CreateResourceGroup creates a new resource group with the specified name.
// If the group already exists, the function returns without error.
func CreateResourceGroup(groupName string) error {
	resourceGroupExists, err := ResourceGroupExists(groupName)
	if err != nil {
		return err
	}

	if resourceGroupExists {
		return nil
	}

	_, err = enforcer.AddNamedGroupingPolicy("g2", groupName, groupName)
	if err != nil {
		return makeErrCasbinConnection("CreateResourceGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("CreateResourceGroup", err)
	}

	return nil
}

// RemoveResourceGroup removes a resource group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func RemoveResourceGroup(groupName string) error {
	if groupName == enclaveAdminGroup {
		return errEnclaveAdminPolicy
	}

	resourceGroupExists, err := ResourceGroupExists(groupName)
	if err != nil {
		return err
	}
	if !resourceGroupExists {
		return makeResourceGroupNotFoundError(groupName)
	}

	_, err = enforcer.RemoveFilteredNamedGroupingPolicy("g2", 1, groupName)
	if err != nil {
		return makeErrCasbinConnection("RemoveResourceGroup", err)
	}

	_, err = enforcer.RemoveFilteredPolicy(0, groupName)
	if err != nil {
		return makeErrCasbinConnection("RemoveResourceGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("RemoveResourceGroup", err)
	}

	return nil
}

// GetResourceGroups returns all resource groups as a slice of ResourceGroup structs.
func GetResourceGroups() ([]ResourceGroup, error) {
	groups, err := enforcer.GetNamedGroupingPolicy("g2")
	if err != nil {
		return nil, makeErrCasbinConnection("GetResourceGroups", err)
	}

	groupsStructured := make([]ResourceGroup, len(groups))
	for i, group := range groups {
		groupsStructured[i] = ResourceGroup{
			ResourceName: group[0],
			GroupName:    group[1],
		}
	}

	return groupsStructured, nil
}

// ResourceGroupExists checks if a resource group with the specified name exists.
func ResourceGroupExists(groupName string) (bool, error) {
	filtered, err := enforcer.GetFilteredNamedGroupingPolicy("g2", 1, groupName)
	if err != nil {
		return false, makeErrCasbinConnection("ResourceGroupExists", err)
	}

	return len(filtered) > 0, nil
}

// AddResourceToGroup adds a resource to one or more groups.
// It validates that all specified groups exist before adding the resource.
func AddResourceToGroup(resourceName string, groupName ...string) error {
	if resourceName == nullUser {
		return errNullUser
	}

	// Check that all groups exist
	for _, group := range groupName {
		groupExists, err := ResourceGroupExists(group)
		if err != nil {
			return err
		}
		if !groupExists {
			return makeResourceGroupNotFoundError(group)
		}
	}

	_, err := enforcer.AddNamedGroupingPolicies(
		"g2",
		[][]string{append([]string{resourceName}, groupName...)},
	)
	if err != nil {
		return makeErrCasbinConnection("AddResourceToGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("AddResourceToGroup", err)
	}

	return nil
}

// RemoveResourceFromGroup removes a resource from one or more groups.
// It validates that all specified groups exist before removing the resource.
func RemoveResourceFromGroup(resourceName string, groupName ...string) error {
	if resourceName == nullUser {
		return errNullUser
	}

	// Check that all groups exist
	for _, group := range groupName {
		groupExists, err := ResourceGroupExists(group)
		if err != nil {
			return err
		}
		if !groupExists {
			return makeResourceGroupNotFoundError(group)
		}
	}

	_, err := enforcer.RemoveNamedGroupingPolicies(
		"g2",
		[][]string{append([]string{resourceName}, groupName...)},
	)
	if err != nil {
		return makeErrCasbinConnection("RemoveResourceFromGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("RemoveResourceFromGroup", err)
	}

	return nil
}

// RemoveResource removes a resource from all groups it belongs to.
func RemoveResource(resourceName string) error {
	if resourceName == nullUser {
		return errNullUser
	}

	_, err := enforcer.RemoveFilteredNamedGroupingPolicy("g2", 0, resourceName)
	if err != nil {
		return makeErrCasbinConnection("RemoveResource", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("RemoveResource", err)
	}

	return nil
}

// GetGroupsForResource returns all groups that a specific resource belongs to.
func GetGroupsForResource(resourceName string) ([]string, error) {
	resourceGroups, err := enforcer.GetFilteredNamedGroupingPolicy("g2", 0, resourceName)
	if err != nil {
		return nil, makeErrCasbinConnection("GetGroupsForResource", err)
	}

	groupNames := make([]string, 0, len(resourceGroups))
	for _, group := range resourceGroups {
		groupNames = append(groupNames, group[1])
	}

	return groupNames, nil
}

// GetResourceGroup returns all resources that belong to a specific group.
func GetResourceGroup(groupName string) ([]string, error) {
	groupExists, err := ResourceGroupExists(groupName)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, makeResourceGroupNotFoundError(groupName)
	}

	resourceGroups, err := enforcer.GetFilteredNamedGroupingPolicy("g2", 1, groupName)
	if err != nil {
		return nil, makeErrCasbinConnection("GetResourcesInGroup", err)
	}

	resourceNames := make([]string, 0, len(resourceGroups))
	for _, group := range resourceGroups {
		if group[0] == nullUser {
			continue
		}
		resourceNames = append(resourceNames, group[0])
	}

	return resourceNames, nil
}
