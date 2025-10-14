package backend

import (
	"stream/cfg"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// StudioRole defines the authorization level of a user within a studio
type StudioRole int

const (
	StudioRoleViewer StudioRole = 0 // Can watch streams
	StudioRoleMember StudioRole = 1 // Can stream
	StudioRoleAdmin  StudioRole = 2 // Can manage rooms/members
	StudioRoleOwner  StudioRole = 3 // Full control
)

// GetStudioRoleName returns the human-readable name for a studio role
func GetStudioRoleName(role StudioRole) string {
	switch role {
	case StudioRoleViewer:
		return "Viewer"
	case StudioRoleMember:
		return "Member"
	case StudioRoleAdmin:
		return "Admin"
	case StudioRoleOwner:
		return "Owner"
	default:
		return "Unknown"
	}
}

// StudioMembership represents a user's membership and role in a studio
type StudioMembership struct {
	UserId   int        `json:"userId"`
	StudioId int        `json:"studioId"`
	Role     StudioRole `json:"role"`
	JoinedAt time.Time  `json:"joinedAt"`
}

// Packing function for vbolt serialization

func PackStudioMembership(self *StudioMembership, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.UserId, buf)
	vpack.Int(&self.StudioId, buf)
	vpack.Int((*int)(&self.Role), buf)
	vpack.Time(&self.JoinedAt, buf)
}

// Buckets and indexes for membership storage

// Membership bucket: (userId, studioId) composite key -> StudioMembership
var MembershipBkt = vbolt.Bucket(&cfg.Info, "studio_membership", vpack.FInt, PackStudioMembership)

// MembershipByUserIdx: userId (term) -> membershipId (target)
// Find all studios a user belongs to
var MembershipByUserIdx = vbolt.Index(&cfg.Info, "membership_by_user", vpack.FInt, vpack.FInt)

// MembershipByStudioIdx: studioId (term) -> membershipId (target)
// Find all members of a studio
var MembershipByStudioIdx = vbolt.Index(&cfg.Info, "membership_by_studio", vpack.FInt, vpack.FInt)

// Helper functions

// GetMembership retrieves a membership by ID
func GetMembership(tx *vbolt.Tx, membershipId int) (membership StudioMembership) {
	vbolt.Read(tx, MembershipBkt, membershipId, &membership)
	return
}

// GetUserStudioRole returns the user's role in a studio, or -1 if not a member
func GetUserStudioRole(tx *vbolt.Tx, userId int, studioId int) StudioRole {
	// Get all memberships for the user
	var membershipIds []int
	vbolt.ReadTermTargets(tx, MembershipByUserIdx, userId, &membershipIds, vbolt.Window{})

	// Find the membership for this studio
	for _, membershipId := range membershipIds {
		membership := GetMembership(tx, membershipId)
		if membership.StudioId == studioId {
			return membership.Role
		}
	}

	// Not a member
	return -1
}

// HasStudioPermission checks if a user has at least the specified role in a studio
func HasStudioPermission(tx *vbolt.Tx, userId int, studioId int, minRole StudioRole) bool {
	// Site admins have full access to all studios
	user := GetUser(tx, userId)
	if user.Role == RoleSiteAdmin {
		return true
	}

	// Check studio-specific role
	userRole := GetUserStudioRole(tx, userId, studioId)
	return userRole >= minRole
}

// ListUserStudios returns all studios the user is a member of
func ListUserStudios(tx *vbolt.Tx, userId int) []Studio {
	var studios []Studio
	var membershipIds []int

	vbolt.ReadTermTargets(tx, MembershipByUserIdx, userId, &membershipIds, vbolt.Window{})

	for _, membershipId := range membershipIds {
		membership := GetMembership(tx, membershipId)
		studio := GetStudioById(tx, membership.StudioId)
		if studio.Id > 0 {
			studios = append(studios, studio)
		}
	}

	return studios
}

// ListStudioMembers returns all memberships for a studio
func ListStudioMembers(tx *vbolt.Tx, studioId int) []StudioMembership {
	var memberships []StudioMembership
	var membershipIds []int

	vbolt.ReadTermTargets(tx, MembershipByStudioIdx, studioId, &membershipIds, vbolt.Window{})

	for _, membershipId := range membershipIds {
		membership := GetMembership(tx, membershipId)
		if membership.UserId > 0 {
			memberships = append(memberships, membership)
		}
	}

	return memberships
}

// API Request/Response types

type AddStudioMemberRequest struct {
	StudioId  int        `json:"studioId"`
	UserId    int        `json:"userId,omitempty"`    // Optional: use this OR userEmail
	UserEmail string     `json:"userEmail,omitempty"` // Optional: use this OR userId
	Role      StudioRole `json:"role"`
}

type AddStudioMemberResponse struct {
	Success    bool             `json:"success"`
	Error      string           `json:"error,omitempty"`
	Membership StudioMembership `json:"membership,omitempty"`
}

type RemoveStudioMemberRequest struct {
	StudioId int `json:"studioId"`
	UserId   int `json:"userId"`
}

type RemoveStudioMemberResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type UpdateStudioMemberRoleRequest struct {
	StudioId int        `json:"studioId"`
	UserId   int        `json:"userId"`
	NewRole  StudioRole `json:"newRole"`
}

type UpdateStudioMemberRoleResponse struct {
	Success    bool             `json:"success"`
	Error      string           `json:"error,omitempty"`
	Membership StudioMembership `json:"membership,omitempty"`
}

type MemberWithDetails struct {
	StudioMembership
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
	RoleName  string `json:"roleName"`
}

type ListStudioMembersRequest struct {
	StudioId int `json:"studioId"`
}

type ListStudioMembersResponse struct {
	Success bool                `json:"success"`
	Error   string              `json:"error,omitempty"`
	Members []MemberWithDetails `json:"members,omitempty"`
}

type LeaveStudioRequest struct {
	StudioId int `json:"studioId"`
}

type LeaveStudioResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// API Procedures

func AddStudioMember(ctx *vbeam.Context, req AddStudioMemberRequest) (resp AddStudioMemberResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if caller has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can add members"
		return
	}

	// Determine target user ID
	targetUserId := req.UserId
	if req.UserEmail != "" {
		// Look up user by email
		vbolt.Read(ctx.Tx, EmailBkt, req.UserEmail, &targetUserId)
		if targetUserId == 0 {
			resp.Success = false
			resp.Error = "No user found with that email address"
			return
		}
	}

	// Validate target user exists
	targetUser := GetUser(ctx.Tx, targetUserId)
	if targetUser.Id == 0 {
		resp.Success = false
		resp.Error = "User not found"
		return
	}

	// Validate role
	if req.Role < StudioRoleViewer || req.Role > StudioRoleOwner {
		resp.Success = false
		resp.Error = "Invalid role"
		return
	}

	// Only owners can add other owners
	if req.Role == StudioRoleOwner && !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleOwner) {
		resp.Success = false
		resp.Error = "Only studio owners can add other owners"
		return
	}

	// Check if user is already a member
	existingRole := GetUserStudioRole(ctx.Tx, targetUserId, studio.Id)
	if existingRole != -1 {
		resp.Success = false
		resp.Error = "User is already a member of this studio"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Create membership
	membership := StudioMembership{
		UserId:   targetUserId,
		StudioId: studio.Id,
		Role:     req.Role,
		JoinedAt: time.Now(),
	}
	membershipId := vbolt.NextIntId(ctx.Tx, MembershipBkt)
	vbolt.Write(ctx.Tx, MembershipBkt, membershipId, &membership)
	vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByUserIdx, membershipId, targetUserId)
	vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByStudioIdx, membershipId, studio.Id)

	vbolt.TxCommit(ctx.Tx)

	// Log member addition
	LogInfo(LogCategorySystem, "Studio member added", map[string]interface{}{
		"studioId":      studio.Id,
		"studioName":    studio.Name,
		"addedUserId":   targetUserId,
		"addedUserName": targetUser.Name,
		"role":          GetStudioRoleName(req.Role),
		"addedBy":       caller.Id,
		"addedByEmail":  caller.Email,
	})

	resp.Success = true
	resp.Membership = membership
	return
}

func RemoveStudioMember(ctx *vbeam.Context, req RemoveStudioMemberRequest) (resp RemoveStudioMemberResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if caller has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can remove members"
		return
	}

	// Check if target user is a member
	targetRole := GetUserStudioRole(ctx.Tx, req.UserId, studio.Id)
	if targetRole == -1 {
		resp.Success = false
		resp.Error = "User is not a member of this studio"
		return
	}

	// Cannot remove the studio owner
	if studio.OwnerId == req.UserId {
		resp.Success = false
		resp.Error = "Cannot remove the studio owner"
		return
	}

	// Only owners can remove admins or other owners
	if targetRole >= StudioRoleAdmin && !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleOwner) {
		resp.Success = false
		resp.Error = "Only studio owners can remove admins or owners"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Find and delete the membership
	var membershipIds []int
	vbolt.ReadTermTargets(ctx.Tx, MembershipByUserIdx, req.UserId, &membershipIds, vbolt.Window{})
	for _, membershipId := range membershipIds {
		membership := GetMembership(ctx.Tx, membershipId)
		if membership.StudioId == studio.Id {
			// Unindex from both user and studio indexes
			vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByUserIdx, membershipId, -1)
			vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByStudioIdx, membershipId, -1)
			// Delete membership
			vbolt.Delete(ctx.Tx, MembershipBkt, membershipId)
			break
		}
	}

	vbolt.TxCommit(ctx.Tx)

	// Get target user info for logging
	targetUser := GetUser(ctx.Tx, req.UserId)

	// Log member removal
	LogInfo(LogCategorySystem, "Studio member removed", map[string]interface{}{
		"studioId":        studio.Id,
		"studioName":      studio.Name,
		"removedUserId":   req.UserId,
		"removedUserName": targetUser.Name,
		"removedRole":     GetStudioRoleName(targetRole),
		"removedBy":       caller.Id,
		"removedByEmail":  caller.Email,
	})

	resp.Success = true
	return
}

func UpdateStudioMemberRole(ctx *vbeam.Context, req UpdateStudioMemberRoleRequest) (resp UpdateStudioMemberRoleResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if caller has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can update member roles"
		return
	}

	// Validate new role
	if req.NewRole < StudioRoleViewer || req.NewRole > StudioRoleOwner {
		resp.Success = false
		resp.Error = "Invalid role"
		return
	}

	// Only owners can assign owner role
	if req.NewRole == StudioRoleOwner && !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleOwner) {
		resp.Success = false
		resp.Error = "Only studio owners can assign owner role"
		return
	}

	// Check if target user is a member
	oldRole := GetUserStudioRole(ctx.Tx, req.UserId, studio.Id)
	if oldRole == -1 {
		resp.Success = false
		resp.Error = "User is not a member of this studio"
		return
	}

	// Cannot change the studio owner's role
	if studio.OwnerId == req.UserId {
		resp.Success = false
		resp.Error = "Cannot change the role of the studio owner"
		return
	}

	// Only owners can modify admin/owner roles
	if (oldRole >= StudioRoleAdmin || req.NewRole >= StudioRoleAdmin) &&
		!HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleOwner) {
		resp.Success = false
		resp.Error = "Only studio owners can modify admin or owner roles"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Find and update the membership
	var membershipIds []int
	vbolt.ReadTermTargets(ctx.Tx, MembershipByUserIdx, req.UserId, &membershipIds, vbolt.Window{})
	var updatedMembership StudioMembership
	for _, membershipId := range membershipIds {
		membership := GetMembership(ctx.Tx, membershipId)
		if membership.StudioId == studio.Id {
			// Update role
			membership.Role = req.NewRole
			vbolt.Write(ctx.Tx, MembershipBkt, membershipId, &membership)
			updatedMembership = membership
			break
		}
	}

	// Get target user info for logging (before committing the transaction)
	targetUser := GetUser(ctx.Tx, req.UserId)

	vbolt.TxCommit(ctx.Tx)

	// Log role update
	LogInfo(LogCategorySystem, "Studio member role updated", map[string]interface{}{
		"studioId":       studio.Id,
		"studioName":     studio.Name,
		"userId":         req.UserId,
		"userName":       targetUser.Name,
		"oldRole":        GetStudioRoleName(oldRole),
		"newRole":        GetStudioRoleName(req.NewRole),
		"updatedBy":      caller.Id,
		"updatedByEmail": caller.Email,
	})

	resp.Success = true
	resp.Membership = updatedMembership
	return
}

func ListStudioMembersAPI(ctx *vbeam.Context, req ListStudioMembersRequest) (resp ListStudioMembersResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if caller has permission to view members (Viewer+)
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleViewer) {
		resp.Success = false
		resp.Error = "You do not have permission to view this studio's members"
		return
	}

	// Get all memberships for the studio
	memberships := ListStudioMembers(ctx.Tx, studio.Id)

	// Enhance with user details
	resp.Members = make([]MemberWithDetails, 0, len(memberships))
	for _, membership := range memberships {
		user := GetUser(ctx.Tx, membership.UserId)
		if user.Id > 0 {
			resp.Members = append(resp.Members, MemberWithDetails{
				StudioMembership: membership,
				UserName:         user.Name,
				UserEmail:        user.Email,
				RoleName:         GetStudioRoleName(membership.Role),
			})
		}
	}

	resp.Success = true
	return
}

func LeaveStudio(ctx *vbeam.Context, req LeaveStudioRequest) (resp LeaveStudioResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if caller is a member
	callerRole := GetUserStudioRole(ctx.Tx, caller.Id, studio.Id)
	if callerRole == -1 {
		resp.Success = false
		resp.Error = "You are not a member of this studio"
		return
	}

	// Cannot leave if you're the owner
	if studio.OwnerId == caller.Id {
		resp.Success = false
		resp.Error = "Studio owner cannot leave. Transfer ownership or delete the studio."
		return
	}

	vbeam.UseWriteTx(ctx)

	// Find and delete the membership
	var membershipIds []int
	vbolt.ReadTermTargets(ctx.Tx, MembershipByUserIdx, caller.Id, &membershipIds, vbolt.Window{})
	for _, membershipId := range membershipIds {
		membership := GetMembership(ctx.Tx, membershipId)
		if membership.StudioId == studio.Id {
			// Unindex from both user and studio indexes
			vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByUserIdx, membershipId, -1)
			vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByStudioIdx, membershipId, -1)
			// Delete membership
			vbolt.Delete(ctx.Tx, MembershipBkt, membershipId)
			break
		}
	}

	vbolt.TxCommit(ctx.Tx)

	// Log studio leave
	LogInfo(LogCategorySystem, "User left studio", map[string]interface{}{
		"studioId":   studio.Id,
		"studioName": studio.Name,
		"userId":     caller.Id,
		"userName":   caller.Name,
		"userEmail":  caller.Email,
		"role":       GetStudioRoleName(callerRole),
	})

	resp.Success = true
	return
}

// RegisterStudioMembershipMethods registers studio membership-related API procedures
func RegisterStudioMembershipMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, AddStudioMember)
	vbeam.RegisterProc(app, RemoveStudioMember)
	vbeam.RegisterProc(app, UpdateStudioMemberRole)
	vbeam.RegisterProc(app, ListStudioMembersAPI)
	vbeam.RegisterProc(app, LeaveStudio)
}
