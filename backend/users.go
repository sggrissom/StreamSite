package backend

import (
	"errors"
	"stream/cfg"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
	"golang.org/x/crypto/bcrypt"
)

// UserRole defines the authorization level of a user
type UserRole int

const (
	RoleUser        UserRole = 0 // Base user - can access basic features
	RoleStreamAdmin UserRole = 1 // Stream/Class/Studio admin - can manage streams
	RoleSiteAdmin   UserRole = 2 // Site admin - full access to all features
)

func RegisterUserMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, CreateAccount)
	vbeam.RegisterProc(app, GetAuthContext)
	vbeam.RegisterProc(app, ListAllUsers)
	vbeam.RegisterProc(app, UpdateUserRole)
}

// Request/Response types
type CreateAccountRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountResponse struct {
	Token string       `json:"token,omitempty"`
	Auth  AuthResponse `json:"auth,omitempty"`
}

type LoginResponse struct {
	Token string       `json:"token,omitempty"`
	Auth  AuthResponse `json:"auth,omitempty"`
}

type AuthResponse struct {
	Id               int      `json:"id"`
	Name             string   `json:"name"`
	Email            string   `json:"email"`
	Role             UserRole `json:"role"`
	IsStreamAdmin    bool     `json:"isStreamAdmin"`    // Quick check for stream admin or higher
	IsSiteAdmin      bool     `json:"isSiteAdmin"`      // Quick check for site admin
	CanManageStudios bool     `json:"canManageStudios"` // Has Member+ role in any studio
}

// Database types
type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	Creation  time.Time `json:"creation"`
	LastLogin time.Time `json:"lastLogin"`
}

// Packing functions for vbolt serialization
func PackUser(self *User, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.String(&self.Name, buf)
	vpack.String(&self.Email, buf)
	vpack.Int((*int)(&self.Role), buf)
	vpack.Time(&self.Creation, buf)
	vpack.Time(&self.LastLogin, buf)
}

// Buckets for vbolt database storage
var UsersBkt = vbolt.Bucket(&cfg.Info, "users", vpack.FInt, PackUser)

// user id => hashed password
var PasswdBkt = vbolt.Bucket(&cfg.Info, "passwd", vpack.FInt, vpack.ByteSlice)

// email => user id
var EmailBkt = vbolt.Bucket(&cfg.Info, "email", vpack.StringZ, vpack.Int)

// Database helper functions
func GetUserId(tx *vbolt.Tx, email string) (userId int) {
	vbolt.Read(tx, EmailBkt, email, &userId)
	return
}

func GetUser(tx *vbolt.Tx, userId int) (user User) {
	vbolt.Read(tx, UsersBkt, userId, &user)
	return
}

func GetPassHash(tx *vbolt.Tx, userId int) (hash []byte) {
	vbolt.Read(tx, PasswdBkt, userId, &hash)
	return
}

func AddUserTx(tx *vbolt.Tx, req CreateAccountRequest, hash []byte) User {
	var user User
	user.Id = vbolt.NextIntId(tx, UsersBkt)
	user.Name = req.Name
	user.Email = req.Email
	user.Role = RoleUser // New users default to base user role

	// First user (ID=1) is automatically set as Site Admin
	if user.Id == 1 {
		user.Role = RoleSiteAdmin
	}

	user.Creation = time.Now()
	user.LastLogin = time.Now()

	// Save user data
	vbolt.Write(tx, UsersBkt, user.Id, &user)
	// Store password hash (can be empty for OAuth users)
	vbolt.Write(tx, PasswdBkt, user.Id, &hash)
	vbolt.Write(tx, EmailBkt, user.Email, &user.Id)

	return user
}

func GetAuthResponseFromUser(tx *vbolt.Tx, user User) AuthResponse {
	// Site admins can manage everything
	canManageStudios := user.Role == RoleSiteAdmin

	// Otherwise, check if user has Member+ role in any studio
	if !canManageStudios {
		canManageStudios = UserHasStudioManagementAccess(tx, user.Id)
	}

	return AuthResponse{
		Id:               user.Id,
		Name:             user.Name,
		Email:            user.Email,
		Role:             user.Role,
		IsStreamAdmin:    user.Role >= RoleStreamAdmin, // Stream admin or site admin
		IsSiteAdmin:      user.Role == RoleSiteAdmin,   // Site admin only
		CanManageStudios: canManageStudios,
	}
}

// vbeam procedures
func CreateAccount(ctx *vbeam.Context, req CreateAccountRequest) (resp CreateAccountResponse, err error) {
	// Validate request
	if err = validateCreateAccountRequest(req); err != nil {
		return
	}

	// Check if email already exists
	userId := GetUserId(ctx.Tx, req.Email)
	if userId != 0 {
		return resp, errors.New("Email already registered")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return resp, errors.New("Failed to process password")
	}

	// Create user
	vbeam.UseWriteTx(ctx)
	user := AddUserTx(ctx.Tx, req, hash)

	// Check if request has anonymous code session (userId=-1) and migrate it
	if len(ctx.Token) > 0 {
		// Parse JWT to check for anonymous session
		token, parseErr := jwt.ParseWithClaims(ctx.Token, &Claims{}, func(token *jwt.Token) (any, error) {
			return jwtKey, nil
		})

		if parseErr == nil && token.Valid {
			if claims, ok := token.Claims.(*Claims); ok && claims.UserId == -1 && claims.SessionToken != "" {
				// Migrate anonymous code session to new user account
				vbolt.Write(ctx.Tx, UserCodeSessionsBkt, user.Id, &claims.SessionToken)

				LogInfo(LogCategoryAuth, "Migrated anonymous code session to new account", map[string]interface{}{
					"userId":       user.Id,
					"sessionToken": claims.SessionToken,
				})
			}
		}
	}

	resp.Auth = GetAuthResponseFromUser(ctx.Tx, user)
	vbolt.TxCommit(ctx.Tx)

	return
}

func GetAuthContext(ctx *vbeam.Context, req Empty) (resp AuthResponse, err error) {
	user, authErr := GetAuthUser(ctx)
	if authErr == nil && user.Id != 0 {
		resp = GetAuthResponseFromUser(ctx.Tx, user)
	}
	return
}

func ListAllUsers(ctx *vbeam.Context, req ListAllUsersRequest) (resp ListAllUsersResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Only site admins can list all users
	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can access this feature")
	}

	// Get all users from the bucket
	var allUsers []User
	vbolt.IterateAllReverse(ctx.Tx, UsersBkt, func(userId int, user User) bool {
		allUsers = append(allUsers, user)
		return true // continue iteration
	})

	// Build response with studio count information
	resp.Users = make([]UserWithStats, 0, len(allUsers))
	for _, user := range allUsers {
		var membershipIds []int
		vbolt.ReadTermTargets(ctx.Tx, MembershipByUserIdx, user.Id, &membershipIds, vbolt.Window{})

		resp.Users = append(resp.Users, UserWithStats{
			User:        user,
			StudioCount: len(membershipIds),
		})
	}

	return
}

func UpdateUserRole(ctx *vbeam.Context, req UpdateUserRoleRequest) (resp UpdateUserRoleResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Only site admins can update user roles
	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can change user roles")
	}

	// Cannot change your own role
	if caller.Id == req.UserId {
		return resp, errors.New("You cannot change your own role")
	}

	// Get the user to update
	user := GetUser(ctx.Tx, req.UserId)
	if user.Id == 0 {
		return resp, errors.New("User not found")
	}

	// Validate new role
	if req.NewRole < RoleUser || req.NewRole > RoleSiteAdmin {
		return resp, errors.New("Invalid role value")
	}

	// Update user role
	vbeam.UseWriteTx(ctx)
	user.Role = req.NewRole
	vbolt.Write(ctx.Tx, UsersBkt, user.Id, &user)
	vbolt.TxCommit(ctx.Tx)

	// Log role change
	LogInfo(LogCategorySystem, "User role updated", map[string]interface{}{
		"userId":    user.Id,
		"userEmail": user.Email,
		"newRole":   req.NewRole,
		"updatedBy": caller.Id,
	})

	resp.User = user
	return
}

type Empty struct{}

type ListAllUsersRequest struct {
	// Empty - site admin only
}

type UserWithStats struct {
	User
	StudioCount int `json:"studioCount"` // Number of studios user is a member of
}

type ListAllUsersResponse struct {
	Users []UserWithStats `json:"users,omitempty"`
}

type UpdateUserRoleRequest struct {
	UserId  int      `json:"userId"`
	NewRole UserRole `json:"newRole"`
}

type UpdateUserRoleResponse struct {
	User User `json:"user,omitempty"`
}

func validateCreateAccountRequest(req CreateAccountRequest) error {
	if req.Name == "" {
		return errors.New("Name is required")
	}
	if req.Email == "" {
		return errors.New("Email is required")
	}

	// Allow empty passwords for OAuth users
	if req.Password != "" {
		if len(req.Password) < 8 {
			return errors.New("Password must be at least 8 characters")
		}
		if req.Password != req.ConfirmPassword {
			return errors.New("Passwords do not match")
		}
	}
	return nil
}
