package auth

import "fmt"

// GroupType represents the policy type for different group kinds
type GroupType string

const (
	UserGroupType     GroupType = "g"
	ResourceGroupType GroupType = "g2"
)

// group interface that both UserGroup and ResourceGroup should implement
type group interface {
	GetName() string
	GetGroupName() string
}

// Implement group interface for UserGroup
func (ug UserGroup) GetName() string {
	return ug.UserName
}

func (ug UserGroup) GetGroupName() string {
	return ug.GroupName
}

// Implement group interface for ResourceGroup
func (rg ResourceGroup) GetName() string {
	return rg.ResourceName
}

func (rg ResourceGroup) GetGroupName() string {
	return rg.GroupName
}

// groupManager handles common group operations for different group types
type groupManager[T group] struct {
	groupType       GroupType
	groupName       string
	nullName        string
	createGroupFunc func(T, []string) T
}

// newUserGroupManager creates a manager for user groups
func newUserGroupManager() *groupManager[UserGroup] {
	return &groupManager[UserGroup]{
		groupType: UserGroupType,
		groupName: "userGroup",
		nullName:  nullUser,
		createGroupFunc: func(group UserGroup, data []string) UserGroup {
			return UserGroup{
				UserName:  data[0],
				GroupName: data[1],
			}
		},
	}
}

// newResourceGroupManager creates a manager for resource groups
func newResourceGroupManager() *groupManager[ResourceGroup] {
	return &groupManager[ResourceGroup]{
		groupType: ResourceGroupType,
		groupName: "resourceGroup",
		nullName:  nullResource,
		createGroupFunc: func(group ResourceGroup, data []string) ResourceGroup {
			return ResourceGroup{
				ResourceName: data[0],
				GroupName:    data[1],
			}
		},
	}
}

// CreateGroup creates a new group with the specified name.
// If the group already exists, the function returns without error.
func (gm *groupManager[T]) CreateGroup(groupName string) error {
	groupExists, err := gm.GroupExists(groupName)
	if err != nil {
		return err
	}

	if groupExists {
		return nil
	}

	_, err = enforcer.AddNamedGroupingPolicy(
		string(gm.groupType),
		groupName,
		groupName,
	)
	if err != nil {
		return &CasbinError{"CreateGroup", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"CreateGroup", err}
	}

	return nil
}

// RemoveGroup removes a group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func (gm *groupManager[T]) RemoveGroup(groupName string) error {
	if groupName == enclaveAdminGroup {
		return &ConflictError{"Enclave admin group cannot be removed"}
	}

	groupExists, err := gm.GroupExists(groupName)
	if err != nil {
		return err
	}
	if !groupExists {
		return &NotFoundError{gm.groupName, groupName}
	}

	_, err = enforcer.RemoveFilteredNamedGroupingPolicy(
		string(gm.groupType),
		1,
		groupName,
	)
	if err != nil {
		return &CasbinError{"RemoveGroup", err}
	}

	_, err = enforcer.RemoveFilteredPolicy(0, groupName)
	if err != nil {
		return &CasbinError{"RemoveGroup", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"RemoveGroup", err}
	}

	return nil
}

// GetGroups returns all groups as a slice of group structs.
func (gm *groupManager[T]) GetGroups() ([]T, error) {
	groups, err := enforcer.GetNamedGroupingPolicy(string(gm.groupType))
	if err != nil {
		return nil, &CasbinError{"GetGroups", err}
	}

	var zero T
	groupsStructured := make([]T, len(groups))
	for i, group := range groups {
		groupsStructured[i] = gm.createGroupFunc(zero, group)
	}

	return groupsStructured, nil
}

// GroupExists checks if a group with the specified name exists.
func (gm *groupManager[T]) GroupExists(groupName string) (bool, error) {
	filtered, err := enforcer.GetFilteredNamedGroupingPolicy(
		string(gm.groupType),
		1,
		groupName,
	)
	if err != nil {
		return false, &CasbinError{"GroupExists", err}
	}

	return len(filtered) > 0, nil
}

// AddToGroup adds an entity to one or more groups.
// It validates that all specified groups exist before adding the entity.
func (gm *groupManager[T]) AddToGroup(
	entityName string,
	groupName ...string,
) error {
	if entityName == gm.nullName {
		return &ConflictError{fmt.Sprintf("Name %s is reserved", entityName)}
	}

	// Check that all groups exist
	for _, group := range groupName {
		groupExists, err := gm.GroupExists(group)
		if err != nil {
			return err
		}
		if !groupExists {
			return &NotFoundError{gm.groupName, group}
		}
	}

	groupingPolicies := make([][]string, len(groupName))
	for i, group := range groupName {
		groupingPolicies[i] = []string{entityName, group}
	}

	_, err := enforcer.AddNamedGroupingPolicies(
		string(gm.groupType),
		groupingPolicies,
	)
	if err != nil {
		return &CasbinError{"AddToGroup", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"AddToGroup", err}
	}

	return nil
}

// RemoveFromGroup removes an entity from one or more groups.
// It validates that all specified groups exist before removing the entity.
func (gm *groupManager[T]) RemoveFromGroup(
	entityName string,
	groupName ...string,
) error {
	if entityName == gm.nullName {
		return &ConflictError{fmt.Sprintf("Name %s is reserved", entityName)}
	}

	// Check that all groups exist
	for _, group := range groupName {
		groupExists, err := gm.GroupExists(group)
		if err != nil {
			return err
		}
		if !groupExists {
			return &NotFoundError{gm.groupName, group}
		}
	}

	groupingPolicies := make([][]string, len(groupName))
	for i, group := range groupName {
		groupingPolicies[i] = []string{entityName, group}
	}

	_, err := enforcer.RemoveNamedGroupingPolicies(
		string(gm.groupType),
		groupingPolicies,
	)
	if err != nil {
		return &CasbinError{"RemoveFromGroup", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"RemoveFromGroup", err}
	}

	return nil
}

// RemoveEntity removes an entity from all groups it belongs to.
func (gm *groupManager[T]) RemoveEntity(entityName string) error {
	if entityName == gm.nullName {
		return &ConflictError{fmt.Sprintf("Name %s is reserved", entityName)}
	}

	_, err := enforcer.RemoveFilteredNamedGroupingPolicy(
		string(gm.groupType),
		0,
		entityName,
	)
	if err != nil {
		return &CasbinError{"RemoveEntity", err}
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return &CasbinError{"RemoveEntity", err}
	}

	return nil
}

// GetGroupsForEntity returns all groups that a specific entity belongs to.
func (gm *groupManager[T]) GetGroupsForEntity(
	entityName string,
) ([]string, error) {
	entityGroups, err := enforcer.GetFilteredNamedGroupingPolicy(
		string(gm.groupType),
		0,
		entityName,
	)
	if err != nil {
		return nil, &CasbinError{"GetGroupsForEntity", err}
	}

	groupNames := make([]string, 0, len(entityGroups))
	for _, group := range entityGroups {
		groupNames = append(groupNames, group[1])
	}

	return groupNames, nil
}

// GetEntitiesInGroup returns all entities that belong to a specific group.
func (gm *groupManager[T]) GetEntitiesInGroup(
	groupName string,
) ([]string, error) {
	groupExists, err := gm.GroupExists(groupName)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, &NotFoundError{gm.groupName, groupName}
	}

	entityGroups, err := enforcer.GetFilteredNamedGroupingPolicy(
		string(gm.groupType),
		1,
		groupName,
	)
	if err != nil {
		return nil, &CasbinError{"GetEntitiesInGroup", err}
	}

	entityNames := make([]string, 0, len(entityGroups))
	for _, group := range entityGroups {
		if group[0] == nullUser {
			continue
		}
		entityNames = append(entityNames, group[0])
	}

	return entityNames, nil
}
