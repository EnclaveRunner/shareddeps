//nolint:dupl // Duplicated code is reduced to a minimum with groupManager
package auth

var userGroupManager = newUserGroupManager()

type UserGroup struct {
	UserName  string
	GroupName string
}

// CreateUserGroup creates a new user group with the specified name.
// If the group already exists, the function returns without error.
func CreateUserGroup(groupName string) error {
	return userGroupManager.CreateGroup(groupName)
}

// RemoveUserGroup removes a user group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func RemoveUserGroup(groupName string) error {
	return userGroupManager.RemoveGroup(groupName)
}

// GetUserGroups returns all user groups as a slice of UserGroup structs.
func GetUserGroups() ([]UserGroup, error) {
	return userGroupManager.GetGroups()
}

// UserGroupExists checks if a user group with the specified name exists.
func UserGroupExists(groupName string) (bool, error) {
	return userGroupManager.GroupExists(groupName)
}

// AddUserToGroup adds a user to one or more groups.
// It validates that all specified groups exist before adding the user.
func AddUserToGroup(userName string, groupName ...string) error {
	return userGroupManager.AddToGroup(userName, groupName...)
}

// RemoveUserFromGroup removes a user from one or more groups.
// It validates that all specified groups exist before removing the user.
func RemoveUserFromGroup(userName string, groupName ...string) error {
	return userGroupManager.RemoveFromGroup(userName, groupName...)
}

// RemoveUser removes a user from all groups they belong to.
func RemoveUser(userName string) error {
	return userGroupManager.RemoveEntity(userName)
}

// GetGroupsForUser returns all groups that a specific user belongs to.
func GetGroupsForUser(userName string) ([]string, error) {
	return userGroupManager.GetGroupsForEntity(userName)
}

// GetUserGroup returns all users that belong to a specific group.
func GetUserGroup(groupName string) ([]string, error) {
	return userGroupManager.GetEntitiesInGroup(groupName)
}
