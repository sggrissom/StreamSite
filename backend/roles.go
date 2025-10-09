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
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type ListUsersRequest struct {
	// Empty for now, can add pagination later
}

type UserListInfo struct {
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Role      UserRole `json:"role"`
	RoleName  string   `json:"roleName"`
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
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	if caller.Role != RoleSiteAdmin {
		resp.Success = false
		resp.Error = "Only site admins can change user roles"
		return
	}

	// Validate role value
	if req.Role < RoleUser || req.Role > RoleSiteAdmin {
		resp.Success = false
		resp.Error = "Invalid role specified"
		return
	}

	// Get target user
	user := GetUser(ctx.Tx, req.UserId)
	if user.Id == 0 {
		resp.Success = false
		resp.Error = "User not found"
		return
	}

	// Prevent changing your own role (safety check)
	if caller.Id == req.UserId {
		resp.Success = false
		resp.Error = "Cannot change your own role"
		return
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

	resp.Success = true
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
