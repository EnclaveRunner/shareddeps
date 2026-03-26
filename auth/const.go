package auth

import (
	"fmt"
)

const (
	nullUser            = "null_user"
	nullResource        = "null_resource"
	enclaveAdminGroup   = "enclave_admin"
	UnauthenticatedUser = "__unauthenticated__"
)

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

func (e *CasbinError) Is(target error) bool {
	_, ok := target.(*CasbinError)

	return ok
}

type NotFoundError struct {
	ResourceType string
	Name         string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.ResourceType, e.Name)
}

func (e *NotFoundError) Is(target error) bool {
	_, ok := target.(*NotFoundError)

	return ok
}

type ConflictError struct {
	Reason string
}

func (e *ConflictError) Error() string {
	return "conflict: " + e.Reason
}

func (e *ConflictError) Is(target error) bool {
	_, ok := target.(*ConflictError)

	return ok
}
