package service_test

import (
	"testing"

	"game_backlog_api/app/model"
	"game_backlog_api/app/service"
	"game_backlog_api/helper"
)

func TestCanAccessBacklog(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin":   {"backlog:read:any", "backlog:update:any"},
		"curator": {},
		"user":    {},
	})

	callerUser := model.AuthUser{UserID: 1, Role: "user"}
	callerAdmin := model.AuthUser{UserID: 99, Role: "admin"}

	// Own record access
	if !service.CanAccessBacklog(callerUser, 1, perms, "backlog:read:any") {
		t.Errorf("expected true for backlog owner")
	}

	// Other record without permission (fail closed)
	if service.CanAccessBacklog(callerUser, 2, perms, "backlog:read:any") {
		t.Errorf("expected false for non-owner user")
	}

	// Admin access with :any bypass
	if !service.CanAccessBacklog(callerAdmin, 2, perms, "backlog:read:any") {
		t.Errorf("expected true for admin with read:any permission")
	}
}

func TestValidateAssignRole(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin":   {"role:assign"},
		"curator": {},
		"user":    {},
	})

	admin := model.AuthUser{UserID: 10, Role: "admin"}

	// Error: self-role change
	errs := service.ValidateAssignRole(admin, 10, model.AssignRoleRequest{Role: "curator"}, perms)
	if errs["role"] != "cannot change own role" {
		t.Errorf("expected cannot change own role error, got: %v", errs)
	}

	// Error: unknown role
	errs = service.ValidateAssignRole(admin, 11, model.AssignRoleRequest{Role: "wizard"}, perms)
	if errs["role"] == "" {
		t.Errorf("expected unknown role error")
	}

	// Success
	errs = service.ValidateAssignRole(admin, 11, model.AssignRoleRequest{Role: "curator"}, perms)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
