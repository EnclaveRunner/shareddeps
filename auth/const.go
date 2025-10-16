package auth

import (
	"errors"
	"fmt"
)

const (
	nullUser          = "nullUser"
	nullResource      = "nullResource"
	enclaveAdminGroup = "enclaveAdmin"
)

var errCasbinConnection = errors.New("casbin returned an error during action")

func makeErrCasbinConnection(action string, casbinErr error) error {
	//nolint:errorlint // casbin error is not wrapped in favor of errCasbinConnection
	return fmt.Errorf("%w \"%s\", casbin error: %v", errCasbinConnection, action, casbinErr)
}

var errUserGroupNotFound = errors.New("user group not found")

func makeErrUserGroupNotFound(groupName string) error {
	return fmt.Errorf("%w: %s", errUserGroupNotFound, groupName)
}

var errResourceGroupNotFound = errors.New("resource group not found")

func makeResourceGroupNotFoundError(groupName string) error {
	return fmt.Errorf("%w: %s", errResourceGroupNotFound, groupName)
}

var errEnclaveAdminPolicy = errors.New("cannot modify enclaveAdmin policy related objects")

var errNullUser = fmt.Errorf("user %s is reserved and cannot be used", nullUser)
