package auth

import (
	"errors"
	"strings"
)

// Role represents a user role in the system
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

// Permission represents an action that can be performed
type Permission string

const (
	PermissionReadTools   Permission = "read:tools"
	PermissionWriteTools  Permission = "write:tools"
	PermissionManageUsers Permission = "manage:users"
	PermissionManageRoles Permission = "manage:roles"
)

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionReadTools,
		PermissionWriteTools,
		PermissionManageUsers,
		PermissionManageRoles,
	},
	RoleUser: {
		PermissionReadTools,
		PermissionWriteTools,
	},
	RoleViewer: {
		PermissionReadTools,
	},
}

// HasPermission checks if a user with given roles has a specific permission
func HasPermission(userRoles []string, requiredPermission Permission) bool {
	for _, roleStr := range userRoles {
		role := Role(strings.ToLower(roleStr))
		permissions, exists := RolePermissions[role]
		if !exists {
			continue
		}

		for _, permission := range permissions {
			if permission == requiredPermission {
				return true
			}
		}
	}
	return false
}

// ValidateRoles checks if all provided roles are valid
func ValidateRoles(roles []string) error {
	for _, roleStr := range roles {
		role := Role(strings.ToLower(roleStr))
		if _, exists := RolePermissions[role]; !exists {
			return errors.New("invalid role: " + roleStr)
		}
	}
	return nil
}

// GetUserPermissions returns all permissions for given roles
func GetUserPermissions(roles []string) []Permission {
	permissionMap := make(map[Permission]bool)

	for _, roleStr := range roles {
		role := Role(strings.ToLower(roleStr))
		if permissions, exists := RolePermissions[role]; exists {
			for _, permission := range permissions {
				permissionMap[permission] = true
			}
		}
	}

	var permissions []Permission
	for permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	return permissions
}
