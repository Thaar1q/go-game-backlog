package service

import (
	"strings"

	"game_backlog_api/app/model"
	"game_backlog_api/helper"
)

// 1. User Resource Authorization
func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

// 2. Backlog Ownership Authorization
func CanAccessBacklog(
	current model.AuthUser,
	targetOwnerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetOwnerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

// 3. Role Assignment Validation
func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	req model.AssignRoleRequest,
	perms *helper.PermissionSet,
) map[string]string {
	errs := map[string]string{}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		errs["role"] = "must be filled"
		return errs
	}

	if !perms.IsKnownRole(role) {
		errs["role"] = "role unknown, pick one from: " +
			strings.Join(perms.KnownRoles(), ", ")
	}

	if current.UserID == targetID {
		errs["role"] = "cannot change own role"
	}

	return errs
}
