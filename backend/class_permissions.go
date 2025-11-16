package backend

import (
	"errors"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// GrantClassPermissionRequest grants a user permission to a specific class
type GrantClassPermissionRequest struct {
	ScheduleId int `json:"scheduleId"`
	UserId     int `json:"userId"`
	Role       int `json:"role"` // StudioRoleViewer, Member, Admin, Owner
}

type GrantClassPermissionResponse struct {
	PermissionId int `json:"permissionId"`
}

// GrantClassPermission grants or updates a user's permission for a class
func GrantClassPermission(ctx *vbeam.Context, req GrantClassPermissionRequest) (resp GrantClassPermissionResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Get schedule
	var schedule ClassSchedule
	vbolt.Read(ctx.Tx, ClassSchedulesBkt, req.ScheduleId, &schedule)
	if schedule.Id == 0 {
		return resp, errors.New("schedule not found")
	}

	// Check permission (Admin+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, schedule.StudioId, StudioRoleAdmin) {
		return resp, errors.New("admin permission required")
	}

	// Validate role
	if req.Role < int(StudioRoleViewer) || req.Role > int(StudioRoleOwner) {
		return resp, errors.New("invalid role")
	}

	// Upgrade to write transaction
	vbeam.UseWriteTx(ctx)

	// Check if permission already exists for this user+schedule
	var existingPermIds []int
	vbolt.ReadTermTargets(ctx.Tx, PermsByScheduleIdx, req.ScheduleId, &existingPermIds, vbolt.Window{})

	for _, permId := range existingPermIds {
		var perm ClassPermission
		vbolt.Read(ctx.Tx, ClassPermissionsBkt, permId, &perm)
		if perm.UserId == req.UserId {
			// Update existing permission
			perm.Role = req.Role
			perm.GrantedBy = caller.Id
			perm.GrantedAt = time.Now()
			vbolt.Write(ctx.Tx, ClassPermissionsBkt, perm.Id, &perm)
			vbolt.TxCommit(ctx.Tx)
			return GrantClassPermissionResponse{PermissionId: perm.Id}, nil
		}
	}

	// Create new permission
	perm := ClassPermission{
		Id:         vbolt.NextIntId(ctx.Tx, ClassPermissionsBkt),
		ScheduleId: req.ScheduleId,
		UserId:     req.UserId,
		Role:       req.Role,
		GrantedBy:  caller.Id,
		GrantedAt:  time.Now(),
	}

	vbolt.Write(ctx.Tx, ClassPermissionsBkt, perm.Id, &perm)

	// Update indexes
	vbolt.SetTargetSingleTerm(ctx.Tx, PermsByScheduleIdx, perm.Id, perm.ScheduleId)
	vbolt.SetTargetSingleTerm(ctx.Tx, PermsByUserIdx, perm.Id, perm.UserId)

	vbolt.TxCommit(ctx.Tx)

	return GrantClassPermissionResponse{PermissionId: perm.Id}, nil
}

// RevokeClassPermissionRequest removes a user's permission from a class
type RevokeClassPermissionRequest struct {
	PermissionId int `json:"permissionId"`
}

type RevokeClassPermissionResponse struct {
	Success bool `json:"success"`
}

// RevokeClassPermission removes a user's permission from a class
func RevokeClassPermission(ctx *vbeam.Context, req RevokeClassPermissionRequest) (resp RevokeClassPermissionResponse, err error) {
	// Validate authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Get permission
	var perm ClassPermission
	vbolt.Read(ctx.Tx, ClassPermissionsBkt, req.PermissionId, &perm)
	if perm.Id == 0 {
		return resp, errors.New("permission not found")
	}

	// Get schedule to check studio permission
	var schedule ClassSchedule
	vbolt.Read(ctx.Tx, ClassSchedulesBkt, perm.ScheduleId, &schedule)
	if schedule.Id == 0 {
		return resp, errors.New("schedule not found")
	}

	// Check permission (Admin+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, schedule.StudioId, StudioRoleAdmin) {
		return resp, errors.New("admin permission required")
	}

	// Upgrade to write transaction
	vbeam.UseWriteTx(ctx)

	// Delete permission
	vbolt.Delete(ctx.Tx, ClassPermissionsBkt, perm.Id)

	// Remove from indexes
	vbolt.SetTargetSingleTerm(ctx.Tx, PermsByScheduleIdx, perm.Id, -1)
	vbolt.SetTargetSingleTerm(ctx.Tx, PermsByUserIdx, perm.Id, -1)

	vbolt.TxCommit(ctx.Tx)

	return RevokeClassPermissionResponse{Success: true}, nil
}

// ListClassPermissionsRequest lists all permissions for a class
type ListClassPermissionsRequest struct {
	ScheduleId int `json:"scheduleId"`
}

type ClassPermissionWithUser struct {
	Permission ClassPermission `json:"permission"`
	UserName   string          `json:"userName"`
	UserEmail  string          `json:"userEmail"`
}

type ListClassPermissionsResponse struct {
	Permissions []ClassPermissionWithUser `json:"permissions"`
}

// ListClassPermissions lists all users with permission to a specific class
func ListClassPermissions(ctx *vbeam.Context, req ListClassPermissionsRequest) (resp ListClassPermissionsResponse, err error) {
	// Validate authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Get schedule
	var schedule ClassSchedule
	vbolt.Read(ctx.Tx, ClassSchedulesBkt, req.ScheduleId, &schedule)
	if schedule.Id == 0 {
		return resp, errors.New("schedule not found")
	}

	// Check permission (Admin+ required to list permissions)
	if !HasStudioPermission(ctx.Tx, caller.Id, schedule.StudioId, StudioRoleAdmin) {
		return resp, errors.New("admin permission required")
	}

	// Get all permission IDs for this schedule
	var permIds []int
	vbolt.ReadTermTargets(ctx.Tx, PermsByScheduleIdx, req.ScheduleId, &permIds, vbolt.Window{})

	// Load permissions and join with user data
	result := ListClassPermissionsResponse{
		Permissions: make([]ClassPermissionWithUser, 0, len(permIds)),
	}

	for _, permId := range permIds {
		var perm ClassPermission
		vbolt.Read(ctx.Tx, ClassPermissionsBkt, permId, &perm)

		// Get user details
		var user User
		vbolt.Read(ctx.Tx, UsersBkt, perm.UserId, &user)

		result.Permissions = append(result.Permissions, ClassPermissionWithUser{
			Permission: perm,
			UserName:   user.Name,
			UserEmail:  user.Email,
		})
	}

	return result, nil
}

// ListMyUpcomingClassesRequest gets upcoming classes across all accessible rooms
type ListMyUpcomingClassesRequest struct {
	Limit int `json:"limit"` // Max number of classes to return (default 20)
}

type UpcomingClassWithRoom struct {
	Schedule      ClassSchedule `json:"schedule"`
	InstanceStart time.Time     `json:"instanceStart"`
	InstanceEnd   time.Time     `json:"instanceEnd"`
	RoomId        int           `json:"roomId"`
	RoomName      string        `json:"roomName"`
	StudioId      int           `json:"studioId"`
	StudioName    string        `json:"studioName"`
}

type ListMyUpcomingClassesResponse struct {
	Classes []UpcomingClassWithRoom `json:"classes"`
}

// ListMyUpcomingClasses returns upcoming classes across all rooms the user has access to.
// This shows scheduled classes visible to anyone with room access - class permissions
// only matter when actually viewing the stream during class time.
func ListMyUpcomingClasses(ctx *vbeam.Context, req ListMyUpcomingClassesRequest) (resp ListMyUpcomingClassesResponse, err error) {
	// Validate authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Default limit
	limit := req.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	now := time.Now()

	// Get all studios the user is a member of
	var studioIds []int
	vbolt.ReadTermTargets(ctx.Tx, MembershipByUserIdx, caller.Id, &studioIds, vbolt.Window{})

	allClasses := make([]UpcomingClassWithRoom, 0)

	// For each studio, get all rooms and their upcoming classes
	for _, studioId := range studioIds {
		// Get studio details
		var studio Studio
		if !vbolt.Read(ctx.Tx, StudiosBkt, studioId, &studio) {
			continue
		}

		// Get all rooms in this studio
		var roomIds []int
		vbolt.ReadTermTargets(ctx.Tx, RoomsByStudioIdx, studioId, &roomIds, vbolt.Window{})

		for _, roomId := range roomIds {
			// Get room details
			var room Room
			if !vbolt.Read(ctx.Tx, RoomsBkt, roomId, &room) {
				continue
			}

			// Get upcoming classes for this room
			upcomingClasses := GetNextClassForRoom(ctx.Tx, roomId, now, limit)

			// Add room and studio info to each class
			for _, class := range upcomingClasses {
				allClasses = append(allClasses, UpcomingClassWithRoom{
					Schedule:      class.Schedule,
					InstanceStart: class.InstanceStart,
					InstanceEnd:   class.InstanceEnd,
					RoomId:        roomId,
					RoomName:      room.Name,
					StudioId:      studioId,
					StudioName:    studio.Name,
				})
			}
		}
	}

	// Sort all classes by start time
	for i := 0; i < len(allClasses); i++ {
		for j := i + 1; j < len(allClasses); j++ {
			if allClasses[j].InstanceStart.Before(allClasses[i].InstanceStart) {
				allClasses[i], allClasses[j] = allClasses[j], allClasses[i]
			}
		}
	}

	// Limit results
	if len(allClasses) > limit {
		allClasses = allClasses[:limit]
	}

	return ListMyUpcomingClassesResponse{Classes: allClasses}, nil
}

// RegisterClassPermissionMethods registers all class permission procedures
func RegisterClassPermissionMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, GrantClassPermission)
	vbeam.RegisterProc(app, RevokeClassPermission)
	vbeam.RegisterProc(app, ListClassPermissions)
	vbeam.RegisterProc(app, ListMyUpcomingClasses)
}
