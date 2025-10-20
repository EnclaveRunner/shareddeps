//nolint:dupl // Duplicated code is reduced to a minimum with groupManager
package auth

var resourceGroupManager = newResourceGroupManager()

type ResourceGroup struct {
	ResourceName string
	GroupName    string
}

// CreateResourceGroup creates a new resource group with the specified name.
// If the group already exists, the function returns without error.
func CreateResourceGroup(groupName string) error {
	return resourceGroupManager.CreateGroup(groupName)
}

// RemoveResourceGroup removes a resource group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func RemoveResourceGroup(groupName string) error {
	return resourceGroupManager.RemoveGroup(groupName)
}

// GetResourceGroups returns all resource groups as a slice of ResourceGroup
// structs.
func GetResourceGroups() ([]ResourceGroup, error) {
	return resourceGroupManager.GetGroups()
}

// ResourceGroupExists checks if a resource group with the specified name
// exists.
func ResourceGroupExists(groupName string) (bool, error) {
	return resourceGroupManager.GroupExists(groupName)
}

// AddResourceToGroup adds a resource to one or more groups.
// It validates that all specified groups exist before adding the resource.
func AddResourceToGroup(resourceName string, groupName ...string) error {
	return resourceGroupManager.AddToGroup(resourceName, groupName...)
}

// RemoveResourceFromGroup removes a resource from one or more groups.
// It validates that all specified groups exist before removing the resource.
func RemoveResourceFromGroup(resourceName string, groupName ...string) error {
	return resourceGroupManager.RemoveFromGroup(resourceName, groupName...)
}

// RemoveResource removes a resource from all groups it belongs to.
func RemoveResource(resourceName string) error {
	return resourceGroupManager.RemoveEntity(resourceName)
}

// GetGroupsForResource returns all groups that a specific resource belongs to.
func GetGroupsForResource(resourceName string) ([]string, error) {
	return resourceGroupManager.GetGroupsForEntity(resourceName)
}

// GetResourceGroup returns all resources that belong to a specific group.
func GetResourceGroup(groupName string) ([]string, error) {
	return resourceGroupManager.GetEntitiesInGroup(groupName)
}
