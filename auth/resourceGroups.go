//nolint:dupl // Duplicated code is reduced to a minimum with groupManager
package auth

type ResourceGroup struct {
	ResourceName string
	GroupName    string
}

// CreateResourceGroup creates a new resource group with the specified name.
// If the group already exists, the function returns without error.
func (auth *AuthModule) CreateResourceGroup(groupName string) error {
	return auth.resourceGroupManager.CreateGroup(groupName)
}

// RemoveResourceGroup removes a resource group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func (auth *AuthModule) RemoveResourceGroup(groupName string) error {
	return auth.resourceGroupManager.RemoveGroup(groupName)
}

// GetResourceGroups returns all resource groups as a slice of ResourceGroup
// structs.
func (auth *AuthModule) GetResourceGroups() ([]ResourceGroup, error) {
	return auth.resourceGroupManager.GetGroups()
}

// ResourceGroupExists checks if a resource group with the specified name
// exists.
func (auth *AuthModule) ResourceGroupExists(groupName string) (bool, error) {
	return auth.resourceGroupManager.GroupExists(groupName)
}

// AddResourceToGroup adds a resource to one or more groups.
// It validates that all specified groups exist before adding the resource.
func (auth *AuthModule) AddResourceToGroup(
	resourceName string,
	groupName ...string,
) error {
	return auth.resourceGroupManager.AddToGroup(resourceName, groupName...)
}

// RemoveResourceFromGroup removes a resource from one or more groups.
// It validates that all specified groups exist before removing the resource.
func (auth *AuthModule) RemoveResourceFromGroup(
	resourceName string,
	groupName ...string,
) error {
	return auth.resourceGroupManager.RemoveFromGroup(resourceName, groupName...)
}

// RemoveResource removes a resource from all groups it belongs to.
func (auth *AuthModule) RemoveResource(resourceName string) error {
	return auth.resourceGroupManager.RemoveEntity(resourceName)
}

// GetGroupsForResource returns all groups that a specific resource belongs to.
func (auth *AuthModule) GetGroupsForResource(
	resourceName string,
) ([]string, error) {
	return auth.resourceGroupManager.GetGroupsForEntity(resourceName)
}

// GetResourceGroup returns all resources that belong to a specific group.
func (auth *AuthModule) GetResourceGroup(groupName string) ([]string, error) {
	return auth.resourceGroupManager.GetEntitiesInGroup(groupName)
}
