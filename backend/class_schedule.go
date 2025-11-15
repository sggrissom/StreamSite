package backend

import (
	"errors"
	"stream/cfg"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// ClassSchedule represents a single class schedule (one-time or recurring)
type ClassSchedule struct {
	Id          int    `json:"id"`
	RoomId      int    `json:"roomId"`      // Which room this class uses
	StudioId    int    `json:"studioId"`    // Parent studio (for permissions)
	Name        string `json:"name"`        // e.g., "Math 101"
	Description string `json:"description"` // e.g., "Algebra basics"

	// Schedule Type
	IsRecurring bool `json:"isRecurring"` // false=one-time, true=recurring

	// One-Time Schedule (IsRecurring=false)
	StartTime time.Time `json:"startTime"` // Single class start time (UTC)
	EndTime   time.Time `json:"endTime"`   // Single class end time (UTC)

	// Recurring Schedule (IsRecurring=true)
	RecurStartDate time.Time `json:"recurStartDate"` // When recurrence starts (date only)
	RecurEndDate   time.Time `json:"recurEndDate"`   // When recurrence ends (date only, nullable)
	RecurWeekdays  []int     `json:"recurWeekdays"`  // [0,2,4] = Sun, Tue, Thu (0=Sun, 6=Sat)
	RecurTimeStart string    `json:"recurTimeStart"` // e.g., "09:00" (HH:MM in local timezone)
	RecurTimeEnd   string    `json:"recurTimeEnd"`   // e.g., "10:00"
	RecurTimezone  string    `json:"recurTimezone"`  // e.g., "America/New_York"

	// Camera Automation
	PreRollMinutes  int  `json:"preRollMinutes"`  // Start camera N minutes early (default: 5)
	PostRollMinutes int  `json:"postRollMinutes"` // Stop camera N minutes late (default: 2)
	AutoStartCamera bool `json:"autoStartCamera"` // Enable/disable auto-start (default: true)
	AutoStopCamera  bool `json:"autoStopCamera"`  // Enable/disable auto-stop (default: true)

	// Metadata
	CreatedBy int       `json:"createdBy"` // User ID who created schedule
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsActive  bool      `json:"isActive"` // Soft delete flag
}

// ClassPermission maps users to classes with specific permission levels
type ClassPermission struct {
	Id         int       `json:"id"`
	ScheduleId int       `json:"scheduleId"` // Which class schedule
	UserId     int       `json:"userId"`     // Which user
	Role       int       `json:"role"`       // StudioRoleViewer, Member, Admin, Owner
	GrantedBy  int       `json:"grantedBy"`  // User ID who granted permission
	GrantedAt  time.Time `json:"grantedAt"`
}

// ScheduleExecutionLog logs each time the scheduler takes an action
type ScheduleExecutionLog struct {
	Id         int       `json:"id"`
	ScheduleId int       `json:"scheduleId"`
	RoomId     int       `json:"roomId"`
	Action     string    `json:"action"` // "start_camera", "stop_camera", "skip_already_running"
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"errorMsg"` // Empty if success
}

// Packing functions for vbolt serialization

func PackClassSchedule(self *ClassSchedule, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.Int(&self.StudioId, buf)
	vpack.String(&self.Name, buf)
	vpack.String(&self.Description, buf)
	vpack.Bool(&self.IsRecurring, buf)
	vpack.Time(&self.StartTime, buf)
	vpack.Time(&self.EndTime, buf)
	vpack.Time(&self.RecurStartDate, buf)
	vpack.Time(&self.RecurEndDate, buf)
	vpack.Slice(&self.RecurWeekdays, vpack.Int, buf)
	vpack.String(&self.RecurTimeStart, buf)
	vpack.String(&self.RecurTimeEnd, buf)
	vpack.String(&self.RecurTimezone, buf)
	vpack.Int(&self.PreRollMinutes, buf)
	vpack.Int(&self.PostRollMinutes, buf)
	vpack.Bool(&self.AutoStartCamera, buf)
	vpack.Bool(&self.AutoStopCamera, buf)
	vpack.Int(&self.CreatedBy, buf)
	vpack.Time(&self.CreatedAt, buf)
	vpack.Time(&self.UpdatedAt, buf)
	vpack.Bool(&self.IsActive, buf)
}

func PackClassPermission(self *ClassPermission, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.ScheduleId, buf)
	vpack.Int(&self.UserId, buf)
	vpack.Int(&self.Role, buf)
	vpack.Int(&self.GrantedBy, buf)
	vpack.Time(&self.GrantedAt, buf)
}

func PackScheduleExecutionLog(self *ScheduleExecutionLog, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.ScheduleId, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.String(&self.Action, buf)
	vpack.Time(&self.Timestamp, buf)
	vpack.Bool(&self.Success, buf)
	vpack.String(&self.ErrorMsg, buf)
}

// Buckets for entity storage

// ClassSchedulesBkt: scheduleId -> ClassSchedule
var ClassSchedulesBkt = vbolt.Bucket(&cfg.Info, "class_schedules", vpack.FInt, PackClassSchedule)

// ClassPermissionsBkt: permissionId -> ClassPermission
var ClassPermissionsBkt = vbolt.Bucket(&cfg.Info, "class_permissions", vpack.FInt, PackClassPermission)

// ScheduleLogsBkt: logId -> ScheduleExecutionLog
var ScheduleLogsBkt = vbolt.Bucket(&cfg.Info, "schedule_logs", vpack.FInt, PackScheduleExecutionLog)

// Indexes for efficient queries

// SchedulesByRoomIdx: roomId (term) -> scheduleId (target)
// Find all schedules for a given room
var SchedulesByRoomIdx = vbolt.Index(&cfg.Info, "schedules_by_room", vpack.FInt, vpack.FInt)

// SchedulesByStudioIdx: studioId (term) -> scheduleId (target)
// Find all schedules for a studio
var SchedulesByStudioIdx = vbolt.Index(&cfg.Info, "schedules_by_studio", vpack.FInt, vpack.FInt)

// PermsByScheduleIdx: scheduleId (term) -> permissionId (target)
// Find all permissions for a given schedule
var PermsByScheduleIdx = vbolt.Index(&cfg.Info, "perms_by_schedule", vpack.FInt, vpack.FInt)

// PermsByUserIdx: userId (term) -> permissionId (target)
// Find all class permissions for a given user
var PermsByUserIdx = vbolt.Index(&cfg.Info, "perms_by_user", vpack.FInt, vpack.FInt)

// LogsByScheduleIdx: scheduleId (term) -> logId (target)
// Find all execution logs for a given schedule
var LogsByScheduleIdx = vbolt.Index(&cfg.Info, "logs_by_schedule", vpack.FInt, vpack.FInt)

// LogsByRoomIdx: roomId (term) -> logId (target)
// Find all execution logs for a given room
var LogsByRoomIdx = vbolt.Index(&cfg.Info, "logs_by_room", vpack.FInt, vpack.FInt)

// API Request/Response types

type CreateClassScheduleRequest struct {
	RoomId      int    `json:"roomId"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Schedule Type
	IsRecurring bool `json:"isRecurring"`

	// One-Time Fields
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`

	// Recurring Fields
	RecurStartDate time.Time `json:"recurStartDate"`
	RecurEndDate   time.Time `json:"recurEndDate"`   // Optional
	RecurWeekdays  []int     `json:"recurWeekdays"`  // [0,2,4] for Sun,Tue,Thu
	RecurTimeStart string    `json:"recurTimeStart"` // "09:00"
	RecurTimeEnd   string    `json:"recurTimeEnd"`   // "10:00"
	RecurTimezone  string    `json:"recurTimezone"`  // "America/New_York"

	// Automation Settings
	PreRollMinutes  int  `json:"preRollMinutes"`  // Default: 5
	PostRollMinutes int  `json:"postRollMinutes"` // Default: 2
	AutoStartCamera bool `json:"autoStartCamera"` // Default: true
	AutoStopCamera  bool `json:"autoStopCamera"`  // Default: true
}

type CreateClassScheduleResponse struct {
	ScheduleId int `json:"scheduleId"`
}

// API Procedures

func CreateClassSchedule(ctx *vbeam.Context, req CreateClassScheduleRequest) (resp CreateClassScheduleResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Get room
	var room Room
	vbolt.Read(ctx.Tx, RoomsBkt, req.RoomId, &room)
	if room.Id == 0 {
		return resp, errors.New("room not found")
	}

	// Check permission (Admin+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, room.StudioId, StudioRoleAdmin) {
		return resp, errors.New("admin permission required")
	}

	// Validate request
	if req.Name == "" {
		return resp, errors.New("name required")
	}

	if req.IsRecurring {
		// Recurring schedule validation
		if len(req.RecurWeekdays) == 0 {
			return resp, errors.New("recurring schedule requires weekdays")
		}
		if req.RecurTimeStart == "" || req.RecurTimeEnd == "" {
			return resp, errors.New("recurring schedule requires start/end times")
		}
		if req.RecurTimezone == "" {
			req.RecurTimezone = "UTC"
		}
		if req.RecurStartDate.IsZero() {
			return resp, errors.New("recurring schedule requires start date")
		}
	} else {
		// One-time schedule validation
		if req.StartTime.IsZero() || req.EndTime.IsZero() {
			return resp, errors.New("one-time schedule requires start/end times")
		}
		if req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime) {
			return resp, errors.New("end time must be after start time")
		}
	}

	// Set defaults
	if req.PreRollMinutes == 0 {
		req.PreRollMinutes = 5
	}
	if req.PostRollMinutes == 0 {
		req.PostRollMinutes = 2
	}

	// Create schedule
	vbeam.UseWriteTx(ctx)

	schedule := ClassSchedule{
		Id:              vbolt.NextIntId(ctx.Tx, ClassSchedulesBkt),
		RoomId:          req.RoomId,
		StudioId:        room.StudioId,
		Name:            req.Name,
		Description:     req.Description,
		IsRecurring:     req.IsRecurring,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		RecurStartDate:  req.RecurStartDate,
		RecurEndDate:    req.RecurEndDate,
		RecurWeekdays:   req.RecurWeekdays,
		RecurTimeStart:  req.RecurTimeStart,
		RecurTimeEnd:    req.RecurTimeEnd,
		RecurTimezone:   req.RecurTimezone,
		PreRollMinutes:  req.PreRollMinutes,
		PostRollMinutes: req.PostRollMinutes,
		AutoStartCamera: req.AutoStartCamera,
		AutoStopCamera:  req.AutoStopCamera,
		CreatedBy:       caller.Id,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsActive:        true,
	}

	// Write to database
	vbolt.Write(ctx.Tx, ClassSchedulesBkt, schedule.Id, &schedule)

	// Update indexes
	vbolt.SetTargetSingleTerm(ctx.Tx, SchedulesByRoomIdx, schedule.Id, schedule.RoomId)
	vbolt.SetTargetSingleTerm(ctx.Tx, SchedulesByStudioIdx, schedule.Id, schedule.StudioId)

	vbolt.TxCommit(ctx.Tx)

	resp.ScheduleId = schedule.Id
	return
}

// RegisterClassScheduleMethods registers class schedule API procedures
func RegisterClassScheduleMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, CreateClassSchedule)
}
