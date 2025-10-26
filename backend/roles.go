package backend

import (
	"errors"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

func RegisterRoleMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, SetUserRole)
	vbeam.RegisterProc(app, ListUsers)
}

// Request/Response types
type SetUserRoleRequest struct {
	UserId int      `json:"userId"`
	Role   UserRole `json:"role"`
}

type SetUserRoleResponse struct {
}

type ListUsersRequest struct {
	// Empty for now, can add pagination later
}

type UserListInfo struct {
	Id       int      `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Role     UserRole `json:"role"`
	RoleName string   `json:"roleName"`
}

type ListUsersResponse struct {
	Users []UserListInfo `json:"users"`
}

// Helper function to get role name as string
func GetRoleName(role UserRole) string {
	switch role {
	case RoleUser:
		return "User"
	case RoleStreamAdmin:
		return "Stream Admin"
	case RoleSiteAdmin:
		return "Site Admin"
	default:
		return "Unknown"
	}
}

// vbeam procedures
func SetUserRole(ctx *vbeam.Context, req SetUserRoleRequest) (resp SetUserRoleResponse, err error) {
	// Check if caller is site admin
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can change user roles")
	}

	// Validate role value
	if req.Role < RoleUser || req.Role > RoleSiteAdmin {
		return resp, errors.New("Invalid role specified")
	}

	// Get target user
	user := GetUser(ctx.Tx, req.UserId)
	if user.Id == 0 {
		return resp, errors.New("User not found")
	}

	// Prevent changing your own role (safety check)
	if caller.Id == req.UserId {
		return resp, errors.New("Cannot change your own role")
	}

	// Update user role
	vbeam.UseWriteTx(ctx)
	user.Role = req.Role
	vbolt.Write(ctx.Tx, UsersBkt, user.Id, &user)
	vbolt.TxCommit(ctx.Tx)

	// Log the role change
	LogInfo(LogCategoryAuth, "User role changed", map[string]interface{}{
		"adminId":    caller.Id,
		"adminEmail": caller.Email,
		"userId":     user.Id,
		"userEmail":  user.Email,
		"newRole":    GetRoleName(req.Role),
	})

	return
}

func ListUsers(ctx *vbeam.Context, req ListUsersRequest) (resp ListUsersResponse, err error) {
	// Check if caller is site admin
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		err = errors.New("authentication required")
		return
	}

	if caller.Role != RoleSiteAdmin {
		err = errors.New("only site admins can list users")
		return
	}

	// Get all users from database
	resp.Users = make([]UserListInfo, 0)

	vbolt.IterateAll(ctx.Tx, UsersBkt, func(userId int, user User) bool {
		resp.Users = append(resp.Users, UserListInfo{
			Id:       user.Id,
			Name:     user.Name,
			Email:    user.Email,
			Role:     user.Role,
			RoleName: GetRoleName(user.Role),
		})
		return true // continue iteration
	})

	return
}
