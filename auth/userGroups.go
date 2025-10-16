package auth

type UserGroup struct {
	UserName  string
	GroupName string
}

// CreateUserGroup creates a new user group with the specified name.
// If the group already exists, the function returns without error.
func CreateUserGroup(groupName string) error {
	userGroupExists, err := UserGroupExists(groupName)
	if err != nil {
		return err
	}

	if userGroupExists {
		return nil
	}

	_, err = enforcer.AddNamedGroupingPolicy("g", groupName, groupName)
	if err != nil {
		return makeErrCasbinConnection("CreateUserGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("CreateUserGroup", err)
	}

	return nil
}

// RemoveUserGroup removes a user group and all associated policies.
// It prevents removal of the enclaveAdmin group to maintain system security.
func RemoveUserGroup(groupName string) error {
	if groupName == enclaveAdminGroup {
		return errEnclaveAdminPolicy
	}

	userGroupExists, err := UserGroupExists(groupName)
	if err != nil {
		return err
	}
	if !userGroupExists {
		return makeErrUserGroupNotFound(groupName)
	}

	_, err = enforcer.RemoveFilteredNamedGroupingPolicy("g", 1, groupName)
	if err != nil {
		return makeErrCasbinConnection("RemoveUserGroup", err)
	}

	_, err = enforcer.RemoveFilteredPolicy(0, groupName)
	if err != nil {
		return makeErrCasbinConnection("RemoveUserGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("RemoveUserGroup", err)
	}

	return nil
}

// GetUserGroups returns all user groups as a slice of UserGroup structs.
func GetUserGroups() ([]UserGroup, error) {
	groups, err := enforcer.GetNamedGroupingPolicy("g")
	if err != nil {
		return nil, makeErrCasbinConnection("GetUserGroups", err)
	}

	groupsStructured := make([]UserGroup, len(groups))
	for i, group := range groups {
		groupsStructured[i] = UserGroup{
			UserName:  group[0],
			GroupName: group[1],
		}
	}

	return groupsStructured, nil
}

// UserGroupExists checks if a user group with the specified name exists.
func UserGroupExists(groupName string) (bool, error) {
	filtered, err := enforcer.GetFilteredNamedGroupingPolicy("g", 1, groupName)
	if err != nil {
		return false, makeErrCasbinConnection("UserGroupExists", err)
	}

	return len(filtered) > 0, nil
}

// AddUserToGroup adds a user to one or more groups.
// It validates that all specified groups exist before adding the user.
func AddUserToGroup(userName string, groupName ...string) error {
	if userName == nullUser {
		return errNullUser
	}

	// Check that all groups exist
	for _, group := range groupName {
		groupExists, err := UserGroupExists(group)
		if err != nil {
			return err
		}
		if !groupExists {
			return makeErrUserGroupNotFound(group)
		}
	}

	_, err := enforcer.AddNamedGroupingPolicies(
		"g",
		[][]string{append([]string{userName}, groupName...)},
	)
	if err != nil {
		return makeErrCasbinConnection("AddUserToGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("AddUserToGroup", err)
	}

	return nil
}

// RemoveUserFromGroup removes a user from one or more groups.
// It validates that all specified groups exist before removing the user.
func RemoveUserFromGroup(userName string, groupName ...string) error {
	if userName == nullUser {
		return errNullUser
	}

	// Check that all groups exist
	for _, group := range groupName {
		groupExists, err := UserGroupExists(group)
		if err != nil {
			return err
		}
		if !groupExists {
			return makeErrUserGroupNotFound(group)
		}
	}

	_, err := enforcer.RemoveNamedGroupingPolicies(
		"g",
		[][]string{append([]string{userName}, groupName...)},
	)
	if err != nil {
		return makeErrCasbinConnection("removeUserFromGroup", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("removeUserFromGroup", err)
	}

	return nil
}

// RemoveUser removes a user from all groups they belong to.
func RemoveUser(userName string) error {
	if userName == nullUser {
		return errNullUser
	}

	_, err := enforcer.RemoveFilteredNamedGroupingPolicy("g", 0, userName)
	if err != nil {
		return makeErrCasbinConnection("removeUser", err)
	}

	err = enforcer.SavePolicy()
	if err != nil {
		return makeErrCasbinConnection("removeUser", err)
	}

	return nil
}

// GetGroupsForUser returns all groups that a specific user belongs to.
func GetGroupsForUser(userName string) ([]string, error) {
	userGroups, err := enforcer.GetFilteredNamedGroupingPolicy("g", 0, userName)
	if err != nil {
		return nil, makeErrCasbinConnection("GetGroupsForUser", err)
	}

	groupNames := make([]string, 0, len(userGroups))
	for _, group := range userGroups {
		groupNames = append(groupNames, group[1])
	}

	return groupNames, nil
}

// GetUserGroup returns all users that belong to a specific group.
func GetUserGroup(groupName string) ([]string, error) {
	groupExists, err := UserGroupExists(groupName)
	if err != nil {
		return nil, err
	}
	if !groupExists {
		return nil, makeErrUserGroupNotFound(groupName)
	}

	userGroups, err := enforcer.GetFilteredNamedGroupingPolicy("g", 1, groupName)
	if err != nil {
		return nil, makeErrCasbinConnection("GetUsersInGroup", err)
	}

	userNames := make([]string, 0, len(userGroups))
	for _, group := range userGroups {
		if group[0] == nullUser {
			continue
		}
		userNames = append(userNames, group[0])
	}

	return userNames, nil
}
