//nolint:gosec,dupl // Test code
package auth_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/EnclaveRunner/shareddeps/auth"
	fileadapter "github.com/casbin/casbin/v3/persist/file-adapter"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	enclaveAdminGroup = "enclave_admin"
	nullUser          = "null_user"
	nullResource      = "null_resource"
)

func setupTestAuth(t *testing.T) auth.AuthModule {
	t.Helper()

	// Create a temporary file for testing
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_policy.csv")

	// Create the file to avoid "no such file" error
	file, err := os.Create(tempFile)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	adapter := fileadapter.NewAdapter(tempFile)
	authModule := auth.NewModule(adapter)

	return authModule
}

func TestInitAuth(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

	// Verify the auth module is properly initialized
	// Verify that enclaveAdmin policy exists
	policies, err := authModule.ListPolicies()
	require.NoError(t, err)

	foundAdminPolicy := false
	for _, policy := range policies {
		if policy.UserGroup == enclaveAdminGroup && policy.ResourceGroup == "*" &&
			policy.Permission == "*" {
			foundAdminPolicy = true

			break
		}
	}
	assert.True(t, foundAdminPolicy, "enclave_admin policy should exist")

	// Verify that enclave_admin group is empty
	userGroups, err := authModule.GetUserGroup(enclaveAdminGroup)
	assert.NoError(t, err)
	assert.Empty(t, userGroups, "enclave_admin group should be empty")
}

func TestCreateUserGroup(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

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
			t.Parallel()

			err := authModule.CreateUserGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify group exists
				exists, err := authModule.UserGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.True(t, exists)

				assignedGroups, err := authModule.GetUserGroup(tc.groupName)
				assert.NoError(t, err)
				assert.Len(t, assignedGroups, 0)
			}
		})
	}
}

func TestRemoveUserGroup(t *testing.T) {
	t.Parallel()

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
			errType:   &auth.NotFoundError{},
		},
		{
			name:      "cannot remove enclave_admin group",
			groupName: enclaveAdminGroup,
			wantErr:   true,
			errType:   &auth.ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create a test group for the removal test
			if tc.groupName == "testRemoveGroup" {
				err := authModule.CreateUserGroup("testRemoveGroup")
				require.NoError(t, err)
			}

			err := authModule.RemoveUserGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify group no longer exists
				exists, err := authModule.UserGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestAddUserToGroup(t *testing.T) {
	t.Parallel()

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
			errType:    &auth.NotFoundError{},
		},
		{
			name:       "cannot add nullUser",
			userName:   nullUser,
			groupNames: []string{"testGroup1"},
			wantErr:    true,
			errType:    &auth.ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create test groups
			err := authModule.CreateUserGroup("testGroup1")
			require.NoError(t, err)
			err = authModule.CreateUserGroup("testGroup2")
			require.NoError(t, err)

			err = authModule.AddUserToGroup(tc.userName, tc.groupNames...)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify user is in groups
				userGroups, err := authModule.GetGroupsForUser(tc.userName)
				assert.NoError(t, err)
				for _, groupName := range tc.groupNames {
					assert.Contains(t, userGroups, groupName)
				}
			}
		})
	}
}

func TestRemoveUserFromGroup(t *testing.T) {
	t.Parallel()

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
			errType:    &auth.NotFoundError{},
		},
		{
			name:       "cannot remove nullUser",
			userName:   nullUser,
			groupNames: []string{"testGroup"},
			wantErr:    true,
			errType:    &auth.ConflictError{},
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
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create test groups
			err := authModule.CreateUserGroup("testGroup")
			require.NoError(t, err)
			err = authModule.CreateUserGroup("testGroup1")
			require.NoError(t, err)
			err = authModule.CreateUserGroup("testGroup2")
			require.NoError(t, err)
			err = authModule.CreateUserGroup("testGroup3")
			require.NoError(t, err)

			// Add user to groups for removal tests
			if tc.userName == "testUser" && tc.name == "remove user from group" {
				err = authModule.AddUserToGroup("testUser", "testGroup")
				require.NoError(t, err)
			}

			err = authModule.RemoveUserFromGroup(tc.userName, tc.groupNames...)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify user is no longer in group
				userGroups, err := authModule.GetGroupsForUser(tc.userName)
				assert.NoError(t, err)
				for _, groupName := range tc.groupNames {
					assert.NotContains(t, userGroups, groupName)
				}
			}
		})
	}
}

func TestCreateResourceGroup(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

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
			t.Parallel()

			err := authModule.CreateResourceGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify group exists
				exists, err := authModule.ResourceGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.True(t, exists)

				assignedResources, err := authModule.GetResourceGroup(tc.groupName)
				assert.NoError(t, err)
				assert.Len(t, assignedResources, 0)
			}
		})
	}
}

func TestRemoveResourceGroup(t *testing.T) {
	t.Parallel()

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
			errType:   &auth.NotFoundError{},
		},
		{
			name:      "cannot remove enclave_admin resource group",
			groupName: enclaveAdminGroup,
			wantErr:   true,
			errType:   &auth.ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create a test resource group for the removal test
			if tc.groupName == "testRemoveResourceGroup" {
				err := authModule.CreateResourceGroup("testRemoveResourceGroup")
				require.NoError(t, err)
			}

			err := authModule.RemoveResourceGroup(tc.groupName)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify group no longer exists
				exists, err := authModule.ResourceGroupExists(tc.groupName)
				assert.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

func TestAddResourceToGroup(t *testing.T) {
	t.Parallel()

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
			errType:      &auth.NotFoundError{},
		},
		{
			name:         "add null resource to group",
			resourceName: nullResource,
			groupNames:   []string{"testResourceGroup1"},
			wantErr:      true,
			errType:      &auth.ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create test resource groups
			err := authModule.CreateResourceGroup("testResourceGroup1")
			require.NoError(t, err)
			err = authModule.CreateResourceGroup("testResourceGroup2")
			require.NoError(t, err)

			err = authModule.AddResourceToGroup(tc.resourceName, tc.groupNames...)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify resource is in groups
				resourceGroups, err := authModule.GetGroupsForResource(tc.resourceName)
				assert.NoError(t, err)
				for _, groupName := range tc.groupNames {
					assert.Contains(t, resourceGroups, groupName)
				}
			}
		})
	}
}

func TestAddPolicy(t *testing.T) {
	t.Parallel()

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
			errType:       &auth.NotFoundError{},
		},
		{
			name:          "add policy with non-existent resource group",
			userGroup:     "testUserGroup",
			resourceGroup: "nonExistentResourceGroup",
			method:        "POST",
			wantErr:       true,
			errType:       &auth.NotFoundError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create test groups
			err := authModule.CreateUserGroup("testUserGroup")
			require.NoError(t, err)
			err = authModule.CreateResourceGroup("testResourceGroup")
			require.NoError(t, err)

			err = authModule.AddPolicy(tc.userGroup, tc.resourceGroup, tc.method)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)

				// Verify policy exists
				policies, err := authModule.ListPolicies()
				assert.NoError(t, err)

				foundPolicy := false
				for _, policy := range policies {
					if policy.UserGroup == tc.userGroup &&
						policy.ResourceGroup == tc.resourceGroup &&
						policy.Permission == tc.method {
						foundPolicy = true

						break
					}
				}
				assert.True(t, foundPolicy)
			}
		})
	}
}

func TestRemovePolicy(t *testing.T) {
	t.Parallel()

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
			name:          "cannot remove enclave_admin policy",
			userGroup:     enclaveAdminGroup,
			resourceGroup: "*",
			method:        "*",
			wantErr:       true,
			errType:       &auth.ConflictError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			authModule := setupTestAuth(t)

			// Create test groups and add a policy
			err := authModule.CreateUserGroup("testUserGroup")
			require.NoError(t, err)
			err = authModule.CreateResourceGroup("testResourceGroup")
			require.NoError(t, err)

			if tc.name == "remove existing policy" {
				err = authModule.AddPolicy("testUserGroup", "testResourceGroup", "GET")
				require.NoError(t, err)
			}

			err = authModule.RemovePolicy(tc.userGroup, tc.resourceGroup, tc.method)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errType != nil {
					assert.ErrorAs(t, err, &tc.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUserGroups(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

	// Create test groups
	err := authModule.CreateUserGroup("testGroup1")
	require.NoError(t, err)
	err = authModule.CreateUserGroup("testGroup2")
	require.NoError(t, err)

	groups, err := authModule.GetUserGroups()
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
	t.Parallel()

	authModule := setupTestAuth(t)

	// Create test resource groups
	err := authModule.CreateResourceGroup("testResourceGroup1")
	require.NoError(t, err)
	err = authModule.CreateResourceGroup("testResourceGroup2")
	require.NoError(t, err)

	groups, err := authModule.GetResourceGroups()
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
	t.Parallel()

	authModule := setupTestAuth(t)

	// Create test group and add users
	err := authModule.CreateUserGroup("testGroup")
	require.NoError(t, err)
	err = authModule.AddUserToGroup("user1", "testGroup")
	require.NoError(t, err)
	err = authModule.AddUserToGroup("user2", "testGroup")
	require.NoError(t, err)

	users, err := authModule.GetUserGroup("testGroup")
	assert.NoError(t, err)
	assert.Contains(t, users, "user1")
	assert.Contains(t, users, "user2")
	assert.NotContains(t, users, nullUser) // nullUser should be filtered out
}

func TestGetResourceGroup(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

	// Create test resource group and add resources
	err := authModule.CreateResourceGroup("testResourceGroup")
	require.NoError(t, err)
	err = authModule.AddResourceToGroup("resource1", "testResourceGroup")
	require.NoError(t, err)
	err = authModule.AddResourceToGroup("resource2", "testResourceGroup")
	require.NoError(t, err)

	resources, err := authModule.GetResourceGroup("testResourceGroup")
	assert.NoError(t, err)
	assert.Contains(t, resources, "resource1")
	assert.Contains(t, resources, "resource2")
}

func TestRemoveUser(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

	// Create test groups and add user
	err := authModule.CreateUserGroup("testGroup1")
	require.NoError(t, err)
	err = authModule.CreateUserGroup("testGroup2")
	require.NoError(t, err)
	err = authModule.AddUserToGroup("testUser", "testGroup1", "testGroup2")
	require.NoError(t, err)

	// Remove user
	err = authModule.RemoveUser("testUser")
	assert.NoError(t, err)

	// Verify user is removed from all groups
	userGroups, err := authModule.GetGroupsForUser("testUser")
	assert.NoError(t, err)
	assert.Empty(t, userGroups)

	// Test removing nullUser
	err = authModule.RemoveUser(nullUser)
	var errConflict *auth.ConflictError
	assert.ErrorAs(t, err, &errConflict)
}

func TestRemoveResource(t *testing.T) {
	t.Parallel()

	authModule := setupTestAuth(t)

	// Create test resource groups and add resource
	err := authModule.CreateResourceGroup("testResourceGroup1")
	require.NoError(t, err)
	err = authModule.CreateResourceGroup("testResourceGroup2")
	require.NoError(t, err)
	err = authModule.AddResourceToGroup(
		"testResource",
		"testResourceGroup1",
		"testResourceGroup2",
	)
	require.NoError(t, err)

	// Remove resource
	err = authModule.RemoveResource("testResource")
	assert.NoError(t, err)

	// Verify resource is removed from all groups
	resourceGroups, err := authModule.GetGroupsForResource("testResource")
	assert.NoError(t, err)
	assert.Empty(t, resourceGroups)
}

func TestGroupManagerInterface(t *testing.T) {
	t.Parallel()

	// Test that UserGroup implements group interface
	ug := auth.UserGroup{UserName: "testUser", GroupName: "testGroup"}
	assert.Equal(t, "testUser", ug.GetName())
	assert.Equal(t, "testGroup", ug.GetGroupName())

	// Test that ResourceGroup implements group interface
	rg := auth.ResourceGroup{ResourceName: "testResource", GroupName: "testGroup"}
	assert.Equal(t, "testResource", rg.GetName())
	assert.Equal(t, "testGroup", rg.GetGroupName())
}

// Benchmark tests
func BenchmarkInitAuth(b *testing.B) {
	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "bench_policy.csv")

	for range b.N {
		adapter := fileadapter.NewAdapter(tempFile)
		auth.NewModule(adapter)
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
	authModule := auth.NewModule(adapter)
	err = authModule.CreateUserGroup("benchGroup")
	require.NoError(b, err)

	b.ResetTimer()
	for range b.N {
		err = authModule.AddUserToGroup("benchUser", "benchGroup")
		require.NoError(b, err)
		err = authModule.RemoveUser("benchUser")
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
	authModule := auth.NewModule(adapter)
	err = authModule.CreateUserGroup("benchUserGroup")
	require.NoError(b, err)
	err = authModule.CreateResourceGroup("benchResourceGroup")
	require.NoError(b, err)

	b.ResetTimer()
	for range b.N {
		err = authModule.AddPolicy("benchUserGroup", "benchResourceGroup", "GET")
		require.NoError(b, err)
		err = authModule.RemovePolicy("benchUserGroup", "benchResourceGroup", "GET")
		require.NoError(b, err)
	}
}

func TestSetAuthenticatedUser(t *testing.T) {
	t.Parallel()

	user := "testUserID"

	r := &http.Request{}
	r = r.WithContext(auth.SetAuthenticatedUser(r.Context(), user))

	actual := auth.GetAuthenticatedUser(r.Context())

	assert.Equal(t, user, actual)
}

func TestGetAuthenticatedUser_Unauthenticated(t *testing.T) {
	t.Parallel()

	r := &http.Request{}

	actual := auth.GetAuthenticatedUser(r.Context())

	assert.Equal(t, auth.UnauthenticatedUser, actual)
}

func TestSetAuthenticatedUser_GinContext(t *testing.T) {
	t.Parallel()

	user := "testUserID"

	c, engine := gin.CreateTestContext(nil)
	engine.ContextWithFallback = true
	c.Request, _ = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/",
		http.NoBody,
	)
	c.Request = c.Request.WithContext(
		auth.SetAuthenticatedUser(c.Request.Context(), user),
	)

	actual := auth.GetAuthenticatedUser(c)

	assert.Equal(t, user, actual)
}
