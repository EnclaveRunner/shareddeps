// enforcer
//
//nolint:paralleltest,dupl,gosec // Tests are not run in parallel due to global
package auth

import (
	"os"
	"path/filepath"
	"testing"

	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAdapter(t *testing.T) *fileadapter.Adapter {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_policy.csv")

	// Create the file to avoid "no such file" error
	file, err := os.Create(tempFile)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	adapter := fileadapter.NewAdapter(tempFile)

	return adapter
}

func TestInitAuth(t *testing.T) {
	adapter := setupTestAdapter(t)

	// Initialize auth
	enforcer := InitAuth(adapter)

	// Verify enforcer is not nil
	require.NotNil(t, enforcer)

	// Verify the global enforcer is set
	assert.NotNil(t, enforcer)

	// Verify that enclaveAdmin policy exists
	policies, err := enforcer.GetPolicy()
	require.NoError(t, err)

	foundAdminPolicy := false
	for _, policy := range policies {
		if len(policy) >= 3 && policy[0] == enclaveAdminGroup && policy[1] == "*" &&
			policy[2] == "*" {
			foundAdminPolicy = true

			break
		}
	}
	assert.True(t, foundAdminPolicy, "enclaveAdmin policy should exist")

	// Verify that nullUser is in enclaveAdmin group
	userGroups, err := enforcer.GetNamedGroupingPolicy("g")
	require.NoError(t, err)

	foundAdminGroup := false
	for _, group := range userGroups {
		if len(group) >= 2 && group[0] == nullUser &&
			group[1] == enclaveAdminGroup {
			foundAdminGroup = true

			break
		}
	}
	assert.True(t, foundAdminGroup, "nullUser should be in enclaveAdmin group")
}

func TestCreateUserGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	testCases := []struct {
		name      string
		groupName string
		wantErr   bool
	}{
		{
			name:      "create new user group",
			groupName: "testUserGroup",
			wantErr:   false,
		},
		{
			name:      "create existing user group",
			groupName: "testUserGroup",
			wantErr:   false, // Should not error if group already exists
		},
		{
			name:      "create group with special name",
			groupName: "user-group_123",
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CreateUserGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify group exists
				exists, err := UserGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.True(t, exists)
			}
		})
	}
}

func TestRemoveUserGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create a test group
	err := CreateUserGroup("testRemoveGroup")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		groupName string
		wantErr   bool
		errType   error
	}{
		{
			name:      "remove existing group",
			groupName: "testRemoveGroup",
			wantErr:   false,
		},
		{
			name:      "remove non-existing group",
			groupName: "nonExistentGroup",
			wantErr:   true,
			errType:   &NotFoundError{},
		},
		{
			name:      "cannot remove enclaveAdmin group",
			groupName: enclaveAdminGroup,
			wantErr:   true,
			errType:   &ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RemoveUserGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify group no longer exists
				exists, err := UserGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestAddUserToGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test groups
	err := CreateUserGroup("testGroup1")
	require.NoError(t, err)
	err = CreateUserGroup("testGroup2")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		userName   string
		groupNames []string
		wantErr    bool
		errType    error
	}{
		{
			name:       "add user to single group",
			userName:   "testUser1",
			groupNames: []string{"testGroup1"},
			wantErr:    false,
		},
		{
			name:       "add user to multiple groups",
			userName:   "testUser2",
			groupNames: []string{"testGroup1", "testGroup2"},
			wantErr:    false,
		},
		{
			name:       "add user to non-existent group",
			userName:   "testUser3",
			groupNames: []string{"nonExistentGroup"},
			wantErr:    true,
			errType:    &NotFoundError{},
		},
		{
			name:       "cannot add nullUser",
			userName:   nullUser,
			groupNames: []string{"testGroup1"},
			wantErr:    true,
			errType:    &ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AddUserToGroup(tc.userName, tc.groupNames...)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify user is in groups
				userGroups, err := GetGroupsForUser(tc.userName)
				assert.NoError(t, err)
				for _, groupName := range tc.groupNames {
					assert.Contains(t, userGroups, groupName)
				}
			}
		})
	}
}

func TestRemoveUserFromGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test group and add user
	err := CreateUserGroup("testGroup")
	require.NoError(t, err)
	err = CreateUserGroup("testGroup1")
	require.NoError(t, err)
	err = CreateUserGroup("testGroup2")
	require.NoError(t, err)
	err = CreateUserGroup("testGroup3")
	require.NoError(t, err)
	err = AddUserToGroup("testUser", "testGroup")
	require.NoError(t, err)

	testCases := []struct {
		name       string
		userName   string
		groupNames []string
		wantErr    bool
		errType    error
	}{
		{
			name:       "remove user from group",
			userName:   "testUser",
			groupNames: []string{"testGroup"},
			wantErr:    false,
		},
		{
			name:       "remove user from non-existent group",
			userName:   "testUser",
			groupNames: []string{"nonExistentGroup"},
			wantErr:    true,
			errType:    &NotFoundError{},
		},
		{
			name:       "cannot remove nullUser",
			userName:   nullUser,
			groupNames: []string{"testGroup"},
			wantErr:    true,
			errType:    &ConflictError{},
		},
		{
			name:       "remove user from multiple groups",
			userName:   "testUser",
			groupNames: []string{"testGroup1", "testGroup2", "testGroup3"},
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RemoveUserFromGroup(tc.userName, tc.groupNames...)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify user is no longer in group
				userGroups, err := GetGroupsForUser(tc.userName)
				assert.NoError(t, err)
				for _, groupName := range tc.groupNames {
					assert.NotContains(t, userGroups, groupName)
				}
			}
		})
	}
}

func TestCreateResourceGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	testCases := []struct {
		name      string
		groupName string
		wantErr   bool
	}{
		{
			name:      "create new resource group",
			groupName: "testResourceGroup",
			wantErr:   false,
		},
		{
			name:      "create existing resource group",
			groupName: "testResourceGroup",
			wantErr:   false, // Should not error if group already exists
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := CreateResourceGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify group exists
				exists, err := ResourceGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.True(t, exists)
			}
		})
	}
}

func TestRemoveResourceGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create a test resource group
	err := CreateResourceGroup("testRemoveResourceGroup")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		groupName string
		wantErr   bool
		errType   error
	}{
		{
			name:      "remove existing resource group",
			groupName: "testRemoveResourceGroup",
			wantErr:   false,
		},
		{
			name:      "remove non-existing resource group",
			groupName: "nonExistentResourceGroup",
			wantErr:   true,
			errType:   &NotFoundError{},
		},
		{
			name:      "cannot remove enclaveAdmin resource group",
			groupName: enclaveAdminGroup,
			wantErr:   true,
			errType:   &ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RemoveResourceGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify group no longer exists
				exists, err := ResourceGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestAddResourceToGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test resource groups
	err := CreateResourceGroup("testResourceGroup1")
	require.NoError(t, err)
	err = CreateResourceGroup("testResourceGroup2")
	require.NoError(t, err)

	testCases := []struct {
		name         string
		resourceName string
		groupNames   []string
		wantErr      bool
		errType      error
	}{
		{
			name:         "add resource to single group",
			resourceName: "testResource1",
			groupNames:   []string{"testResourceGroup1"},
			wantErr:      false,
		},
		{
			name:         "add resource to multiple groups",
			resourceName: "testResource2",
			groupNames:   []string{"testResourceGroup1", "testResourceGroup2"},
			wantErr:      false,
		},
		{
			name:         "add resource to non-existent group",
			resourceName: "testResource3",
			groupNames:   []string{"nonExistentResourceGroup"},
			wantErr:      true,
			errType:      &NotFoundError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AddResourceToGroup(tc.resourceName, tc.groupNames...)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify resource is in groups
				resourceGroups, err := GetGroupsForResource(tc.resourceName)
				assert.NoError(t, err)
				for _, groupName := range tc.groupNames {
					assert.Contains(t, resourceGroups, groupName)
				}
			}
		})
	}
}

func TestAddPolicy(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test groups
	err := CreateUserGroup("testUserGroup")
	require.NoError(t, err)
	err = CreateResourceGroup("testResourceGroup")
	require.NoError(t, err)

	testCases := []struct {
		name          string
		userGroup     string
		resourceGroup string
		method        string
		wantErr       bool
		errType       error
	}{
		{
			name:          "add valid policy",
			userGroup:     "testUserGroup",
			resourceGroup: "testResourceGroup",
			method:        "GET",
			wantErr:       false,
		},
		{
			name:          "add duplicate policy",
			userGroup:     "testUserGroup",
			resourceGroup: "testResourceGroup",
			method:        "GET",
			wantErr:       false, // Should not error for duplicate
		},
		{
			name:          "add policy with non-existent user group",
			userGroup:     "nonExistentUserGroup",
			resourceGroup: "testResourceGroup",
			method:        "POST",
			wantErr:       true,
			errType:       &NotFoundError{},
		},
		{
			name:          "add policy with non-existent resource group",
			userGroup:     "testUserGroup",
			resourceGroup: "nonExistentResourceGroup",
			method:        "POST",
			wantErr:       true,
			errType:       &NotFoundError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AddPolicy(tc.userGroup, tc.resourceGroup, tc.method)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify policy exists
				policies, err := enforcer.GetFilteredPolicy(0, tc.userGroup, tc.resourceGroup, tc.method)
				assert.NoError(t, err)
				assert.Len(t, policies, 1)
			}
		})
	}
}

func TestRemovePolicy(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test groups and add a policy
	err := CreateUserGroup("testUserGroup")
	require.NoError(t, err)
	err = CreateResourceGroup("testResourceGroup")
	require.NoError(t, err)
	err = AddPolicy("testUserGroup", "testResourceGroup", "GET")
	require.NoError(t, err)

	testCases := []struct {
		name          string
		userGroup     string
		resourceGroup string
		method        string
		wantErr       bool
		errType       error
	}{
		{
			name:          "remove existing policy",
			userGroup:     "testUserGroup",
			resourceGroup: "testResourceGroup",
			method:        "GET",
			wantErr:       false,
		},
		{
			name:          "remove non-existent policy",
			userGroup:     "testUserGroup",
			resourceGroup: "testResourceGroup",
			method:        "DELETE",
			wantErr:       false, // Should not error for non-existent policy
		},
		{
			name:          "cannot remove enclaveAdmin policy",
			userGroup:     enclaveAdminGroup,
			resourceGroup: "*",
			method:        "*",
			wantErr:       true,
			errType:       &ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := RemovePolicy(tc.userGroup, tc.resourceGroup, tc.method)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorIs(t, err, tc.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserGroups(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test groups
	err := CreateUserGroup("testGroup1")
	require.NoError(t, err)
	err = CreateUserGroup("testGroup2")
	require.NoError(t, err)

	groups, err := GetUserGroups()
	assert.NoError(t, err)
	assert.NotEmpty(t, groups)

	// Should contain at least our test groups and enclaveAdmin
	groupNames := make([]string, len(groups))
	for i, group := range groups {
		groupNames[i] = group.GroupName
	}
	assert.Contains(t, groupNames, "testGroup1")
	assert.Contains(t, groupNames, "testGroup2")
	assert.Contains(t, groupNames, enclaveAdminGroup)
}

func TestGetResourceGroups(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test resource groups
	err := CreateResourceGroup("testResourceGroup1")
	require.NoError(t, err)
	err = CreateResourceGroup("testResourceGroup2")
	require.NoError(t, err)

	groups, err := GetResourceGroups()
	assert.NoError(t, err)
	assert.NotEmpty(t, groups)

	// Should contain at least our test groups
	groupNames := make([]string, len(groups))
	for i, group := range groups {
		groupNames[i] = group.GroupName
	}
	assert.Contains(t, groupNames, "testResourceGroup1")
	assert.Contains(t, groupNames, "testResourceGroup2")
}

func TestGetUserGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test group and add users
	err := CreateUserGroup("testGroup")
	require.NoError(t, err)
	err = AddUserToGroup("user1", "testGroup")
	require.NoError(t, err)
	err = AddUserToGroup("user2", "testGroup")
	require.NoError(t, err)

	users, err := GetUserGroup("testGroup")
	assert.NoError(t, err)
	assert.Contains(t, users, "user1")
	assert.Contains(t, users, "user2")
	assert.NotContains(t, users, nullUser) // nullUser should be filtered out
}

func TestGetResourceGroup(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test resource group and add resources
	err := CreateResourceGroup("testResourceGroup")
	require.NoError(t, err)
	err = AddResourceToGroup("resource1", "testResourceGroup")
	require.NoError(t, err)
	err = AddResourceToGroup("resource2", "testResourceGroup")
	require.NoError(t, err)

	resources, err := GetResourceGroup("testResourceGroup")
	assert.NoError(t, err)
	assert.Contains(t, resources, "resource1")
	assert.Contains(t, resources, "resource2")
}

func TestRemoveUser(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test groups and add user
	err := CreateUserGroup("testGroup1")
	require.NoError(t, err)
	err = CreateUserGroup("testGroup2")
	require.NoError(t, err)
	err = AddUserToGroup("testUser", "testGroup1", "testGroup2")
	require.NoError(t, err)

	// Remove user
	err = RemoveUser("testUser")
	assert.NoError(t, err)

	// Verify user is removed from all groups
	userGroups, err := GetGroupsForUser("testUser")
	assert.NoError(t, err)
	assert.Empty(t, userGroups)

	// Test removing nullUser
	err = RemoveUser(nullUser)
	assert.ErrorIs(t, err, &ConflictError{})
}

func TestRemoveResource(t *testing.T) {
	adapter := setupTestAdapter(t)
	InitAuth(adapter)

	// Create test resource groups and add resource
	err := CreateResourceGroup("testResourceGroup1")
	require.NoError(t, err)
	err = CreateResourceGroup("testResourceGroup2")
	require.NoError(t, err)
	err = AddResourceToGroup(
		"testResource",
		"testResourceGroup1",
		"testResourceGroup2",
	)
	require.NoError(t, err)

	// Remove resource
	err = RemoveResource("testResource")
	assert.NoError(t, err)

	// Verify resource is removed from all groups
	resourceGroups, err := GetGroupsForResource("testResource")
	assert.NoError(t, err)
	assert.Empty(t, resourceGroups)
}

func TestGroupManagerInterface(t *testing.T) {
	// Test that UserGroup implements group interface
	ug := UserGroup{UserName: "testUser", GroupName: "testGroup"}
	assert.Equal(t, "testUser", ug.GetName())
	assert.Equal(t, "testGroup", ug.GetGroupName())

	// Test that ResourceGroup implements group interface
	rg := ResourceGroup{ResourceName: "testResource", GroupName: "testGroup"}
	assert.Equal(t, "testResource", rg.GetName())
	assert.Equal(t, "testGroup", rg.GetGroupName())
}

// Benchmark tests
func BenchmarkInitAuth(b *testing.B) {
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "bench_policy.csv")

	for b.Loop() {
		adapter := fileadapter.NewAdapter(tempFile)
		InitAuth(adapter)
	}
}

func BenchmarkAddUserToGroup(b *testing.B) {
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "bench_policy.csv")

	// Create the file
	file, err := os.Create(tempFile)
	require.NoError(b, err)
	err = file.Close()
	require.NoError(b, err)

	adapter := fileadapter.NewAdapter(tempFile)
	InitAuth(adapter)
	err = CreateUserGroup("benchGroup")
	require.NoError(b, err)

	for b.Loop() {
		err = AddUserToGroup("benchUser", "benchGroup")
		require.NoError(b, err)
		err = RemoveUser("benchUser")
		require.NoError(b, err)
	}
}

func BenchmarkAddPolicy(b *testing.B) {
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "bench_policy.csv")

	// Create the file
	file, err := os.Create(tempFile)
	require.NoError(b, err)
	err = file.Close()
	require.NoError(b, err)

	adapter := fileadapter.NewAdapter(tempFile)
	InitAuth(adapter)
	err = CreateUserGroup("benchUserGroup")
	require.NoError(b, err)
	err = CreateResourceGroup("benchResourceGroup")
	require.NoError(b, err)

	for b.Loop() {
		err = AddPolicy("benchUserGroup", "benchResourceGroup", "GET")
		require.NoError(b, err)
		err = RemovePolicy("benchUserGroup", "benchResourceGroup", "GET")
		require.NoError(b, err)
	}
}
