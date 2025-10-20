package auth

import (
	"fmt"
)

const (
	nullUser                        = "nullUser"
	nullResource                    = "nullResource"
	enclaveAdminGroup               = "enclaveAdmin"
	AuthenticatedUser   ConstextKey = "authenticatedUser"
	UnauthenticatedUser             = "__unauthenticated__"
)

type ConstextKey string

type CasbinError struct {
	Action string
	Err    error
}

func (e *CasbinError) Error() string {
	return fmt.Sprintf("casbin error during %s: %v", e.Action, e.Err)
}

func (e *CasbinError) Unwrap() error {
	return e.Err
}

type NotFoundError struct {
	ResourceType string
	Name         string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.Name)
}

type ConflictError struct {
	Reason string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s", e.Reason)
}
