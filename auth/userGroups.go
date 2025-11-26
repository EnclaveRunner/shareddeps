//nolint:dupl // Duplicated code is reduced to a minimum with groupManager
package auth

type UserGroup struct {
	UserName  string
	GroupName string
}

// CreateUserGroup creates a new user group with the specified name.
// If the group already exists, the function returns without error.
func (auth *AuthModule) CreateUserGroup(groupName string) error {
	return auth.userGroupManager.CreateGroup(groupName)
}

// RemoveUserGroup removes a user group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func (auth *AuthModule) RemoveUserGroup(groupName string) error {
	return auth.userGroupManager.RemoveGroup(groupName)
}

// GetUserGroups returns all user groups as a slice of UserGroup structs.
func (auth *AuthModule) GetUserGroups() ([]UserGroup, error) {
	return auth.userGroupManager.GetGroups()
}

// UserGroupExists checks if a user group with the specified name exists.
func (auth *AuthModule) UserGroupExists(groupName string) (bool, error) {
	return auth.userGroupManager.GroupExists(groupName)
}

// AddUserToGroup adds a user to one or more groups.
// It validates that all specified groups exist before adding the user.
func (auth *AuthModule) AddUserToGroup(
	userName string,
	groupName ...string,
) error {
	return auth.userGroupManager.AddToGroup(userName, groupName...)
}

// RemoveUserFromGroup removes a user from one or more groups.
// It validates that all specified groups exist before removing the user.
func (auth *AuthModule) RemoveUserFromGroup(
	userName string,
	groupName ...string,
) error {
	return auth.userGroupManager.RemoveFromGroup(userName, groupName...)
}

// RemoveUser removes a user from all groups they belong to.
func (auth *AuthModule) RemoveUser(userName string) error {
	return auth.userGroupManager.RemoveEntity(userName)
}

// GetGroupsForUser returns all groups that a specific user belongs to.
func (auth *AuthModule) GetGroupsForUser(userName string) ([]string, error) {
	return auth.userGroupManager.GetGroupsForEntity(userName)
}

// GetUserGroup returns all users that belong to a specific group.
func (auth *AuthModule) GetUserGroup(groupName string) ([]string, error) {
	return auth.userGroupManager.GetEntitiesInGroup(groupName)
}
