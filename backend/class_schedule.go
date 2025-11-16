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

	// One-Time Fields - sent as ISO strings, parsed to time.Time
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`

	// Recurring Fields
	RecurStartDate string `json:"recurStartDate"` // ISO string, parsed to time.Time
	RecurEndDate   string `json:"recurEndDate"`   // Optional ISO string
	RecurWeekdays  []int  `json:"recurWeekdays"`  // [0,2,4] for Sun,Tue,Thu
	RecurTimeStart string `json:"recurTimeStart"` // "09:00"
	RecurTimeEnd   string `json:"recurTimeEnd"`   // "10:00"
	RecurTimezone  string `json:"recurTimezone"`  // "America/New_York"

	// Automation Settings
	PreRollMinutes  int  `json:"preRollMinutes"`  // Default: 5
	PostRollMinutes int  `json:"postRollMinutes"` // Default: 2
	AutoStartCamera bool `json:"autoStartCamera"` // Default: true
	AutoStopCamera  bool `json:"autoStopCamera"`  // Default: true
}

type CreateClassScheduleResponse struct {
	ScheduleId int `json:"scheduleId"`
}

type ListClassSchedulesRequest struct {
	StudioId *int `json:"studioId"` // Optional: filter by studio
	RoomId   *int `json:"roomId"`   // Optional: filter by room
}

type ListClassSchedulesResponse struct {
	Schedules []ClassSchedule `json:"schedules"`
}

type ClassInstance struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type GetScheduleDetailsRequest struct {
	ScheduleId int `json:"scheduleId"`
}

type GetScheduleDetailsResponse struct {
	Schedule          ClassSchedule   `json:"schedule"`
	UpcomingInstances []ClassInstance `json:"upcomingInstances"`
}

type UpdateClassScheduleRequest struct {
	ScheduleId int `json:"scheduleId"`

	// Optional fields - only update if provided
	Name        *string `json:"name"`
	Description *string `json:"description"`

	// One-Time Fields - sent as ISO strings, parsed to time.Time
	StartTime *string `json:"startTime"`
	EndTime   *string `json:"endTime"`

	// Recurring Fields
	RecurStartDate *string `json:"recurStartDate"` // ISO string, parsed to time.Time
	RecurEndDate   *string `json:"recurEndDate"`   // Optional ISO string
	RecurWeekdays  []int   `json:"recurWeekdays"`
	RecurTimeStart *string `json:"recurTimeStart"`
	RecurTimeEnd   *string `json:"recurTimeEnd"`
	RecurTimezone  *string `json:"recurTimezone"`

	// Automation Settings
	PreRollMinutes  *int  `json:"preRollMinutes"`
	PostRollMinutes *int  `json:"postRollMinutes"`
	AutoStartCamera *bool `json:"autoStartCamera"`
	AutoStopCamera  *bool `json:"autoStopCamera"`
}

type UpdateClassScheduleResponse struct {
	Success bool `json:"success"`
}

type DeleteClassScheduleRequest struct {
	ScheduleId int `json:"scheduleId"`
}

type DeleteClassScheduleResponse struct {
	Success bool `json:"success"`
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

	// Parse and validate time fields based on schedule type
	var startTime, endTime time.Time
	var recurStartDate, recurEndDate time.Time

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
		if req.RecurStartDate == "" {
			return resp, errors.New("recurring schedule requires start date")
		}

		// Parse recurring start date
		var err error
		recurStartDate, err = time.Parse(time.RFC3339, req.RecurStartDate)
		if err != nil {
			return resp, errors.New("invalid recur start date format")
		}

		// Parse recurring end date (optional)
		if req.RecurEndDate != "" {
			recurEndDate, err = time.Parse(time.RFC3339, req.RecurEndDate)
			if err != nil {
				return resp, errors.New("invalid recur end date format")
			}
		}
	} else {
		// One-time schedule validation
		if req.StartTime == "" || req.EndTime == "" {
			return resp, errors.New("one-time schedule requires start/end times")
		}

		// Parse one-time start and end times
		var err error
		startTime, err = time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return resp, errors.New("invalid start time format")
		}
		endTime, err = time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return resp, errors.New("invalid end time format")
		}

		// Validate time ordering
		if endTime.Before(startTime) || endTime.Equal(startTime) {
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
		StartTime:       startTime,      // Use parsed time
		EndTime:         endTime,        // Use parsed time
		RecurStartDate:  recurStartDate, // Use parsed time
		RecurEndDate:    recurEndDate,   // Use parsed time
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

func ListClassSchedules(ctx *vbeam.Context, req ListClassSchedulesRequest) (resp ListClassSchedulesResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Must specify at least one filter
	if req.StudioId == nil && req.RoomId == nil {
		return resp, errors.New("must specify studioId or roomId")
	}

	var scheduleIds []int
	var studioId int

	if req.RoomId != nil {
		// Filter by room
		vbolt.ReadTermTargets(ctx.Tx, SchedulesByRoomIdx, *req.RoomId, &scheduleIds, vbolt.Window{})

		// Get room to find studio for permission check
		var room Room
		vbolt.Read(ctx.Tx, RoomsBkt, *req.RoomId, &room)
		if room.Id == 0 {
			return resp, errors.New("room not found")
		}
		studioId = room.StudioId
	} else {
		// Filter by studio
		studioId = *req.StudioId
		vbolt.ReadTermTargets(ctx.Tx, SchedulesByStudioIdx, studioId, &scheduleIds, vbolt.Window{})

		// Verify studio exists
		var studio Studio
		vbolt.Read(ctx.Tx, StudiosBkt, studioId, &studio)
		if studio.Id == 0 {
			return resp, errors.New("studio not found")
		}
	}

	// Check permission (Viewer+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleViewer) {
		return resp, errors.New("viewer permission required")
	}

	// Load schedules
	resp.Schedules = make([]ClassSchedule, 0, len(scheduleIds))
	for _, scheduleId := range scheduleIds {
		var schedule ClassSchedule
		vbolt.Read(ctx.Tx, ClassSchedulesBkt, scheduleId, &schedule)
		if schedule.Id > 0 && schedule.IsActive {
			resp.Schedules = append(resp.Schedules, schedule)
		}
	}

	return
}

func GetScheduleDetails(ctx *vbeam.Context, req GetScheduleDetailsRequest) (resp GetScheduleDetailsResponse, err error) {
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

	// Check permission (Viewer+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, schedule.StudioId, StudioRoleViewer) {
		return resp, errors.New("viewer permission required")
	}

	resp.Schedule = schedule

	// Calculate upcoming instances for recurring schedules
	if schedule.IsRecurring {
		resp.UpcomingInstances = calculateUpcomingInstances(&schedule, 10)
	} else {
		// For one-time schedules, include the single instance if it's in the future
		if schedule.StartTime.After(time.Now()) {
			resp.UpcomingInstances = []ClassInstance{
				{
					StartTime: schedule.StartTime,
					EndTime:   schedule.EndTime,
				},
			}
		} else {
			resp.UpcomingInstances = []ClassInstance{}
		}
	}

	return
}

// Helper function to calculate upcoming instances for recurring schedules
func calculateUpcomingInstances(schedule *ClassSchedule, count int) []ClassInstance {
	if !schedule.IsRecurring {
		return []ClassInstance{}
	}

	// Parse timezone
	loc, err := time.LoadLocation(schedule.RecurTimezone)
	if err != nil {
		loc = time.UTC
	}

	// Parse class times
	startTime, err := time.Parse("15:04", schedule.RecurTimeStart)
	if err != nil {
		return []ClassInstance{}
	}
	endTime, err := time.Parse("15:04", schedule.RecurTimeEnd)
	if err != nil {
		return []ClassInstance{}
	}

	instances := make([]ClassInstance, 0, count)
	now := time.Now().In(loc)

	// Start searching from today
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	// Search up to 365 days in the future
	maxDate := currentDate.Add(365 * 24 * time.Hour)
	if !schedule.RecurEndDate.IsZero() {
		endDate := schedule.RecurEndDate.In(loc)
		if endDate.Before(maxDate) {
			maxDate = endDate
		}
	}

	for currentDate.Before(maxDate) && len(instances) < count {
		// Check if current date is before recurrence start
		if currentDate.Before(schedule.RecurStartDate.In(loc)) {
			currentDate = currentDate.Add(24 * time.Hour)
			continue
		}

		// Check if current weekday matches
		weekday := int(currentDate.Weekday())
		matchesWeekday := false
		for _, wd := range schedule.RecurWeekdays {
			if wd == weekday {
				matchesWeekday = true
				break
			}
		}

		if matchesWeekday {
			// Build instance for this day
			instanceStart := time.Date(
				currentDate.Year(), currentDate.Month(), currentDate.Day(),
				startTime.Hour(), startTime.Minute(), 0, 0, loc,
			)
			instanceEnd := time.Date(
				currentDate.Year(), currentDate.Month(), currentDate.Day(),
				endTime.Hour(), endTime.Minute(), 0, 0, loc,
			)

			// Only include if it's in the future
			if instanceEnd.After(now) {
				instances = append(instances, ClassInstance{
					StartTime: instanceStart,
					EndTime:   instanceEnd,
				})
			}
		}

		currentDate = currentDate.Add(24 * time.Hour)
	}

	return instances
}

// GetCurrentClassForRoom returns the currently active class for a room, or nil if no class is active.
// A class is considered active if the current time is within the class time + grace period (15 minutes).
func GetCurrentClassForRoom(tx *vbolt.Tx, roomId int, now time.Time) *ClassSchedule {
	// Get all active schedules for this room
	var scheduleIds []int
	vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, roomId, &scheduleIds, vbolt.Window{})

	gracePeriod := 15 * time.Minute

	for _, schedId := range scheduleIds {
		var sched ClassSchedule
		if !vbolt.Read(tx, ClassSchedulesBkt, schedId, &sched) {
			continue
		}

		// Skip inactive schedules
		if !sched.IsActive {
			continue
		}

		// One-time schedule: check if we're within the time window
		if !sched.IsRecurring {
			if now.After(sched.StartTime) && now.Before(sched.EndTime.Add(gracePeriod)) {
				return &sched
			}
			continue
		}

		// Recurring schedule: check if there's a current instance
		loc, err := time.LoadLocation(sched.RecurTimezone)
		if err != nil {
			loc = time.UTC
		}

		nowInTz := now.In(loc)
		weekday := int(nowInTz.Weekday())

		// Check if today is a scheduled weekday
		matchesWeekday := false
		for _, wd := range sched.RecurWeekdays {
			if wd == weekday {
				matchesWeekday = true
				break
			}
		}

		if !matchesWeekday {
			continue
		}

		// Check date range
		if nowInTz.Before(sched.RecurStartDate.In(loc)) {
			continue
		}
		if !sched.RecurEndDate.IsZero() && nowInTz.After(sched.RecurEndDate.In(loc)) {
			continue
		}

		// Parse class times
		startTime, err := time.Parse("15:04", sched.RecurTimeStart)
		if err != nil {
			continue
		}
		endTime, err := time.Parse("15:04", sched.RecurTimeEnd)
		if err != nil {
			continue
		}

		// Build today's instance
		instanceStart := time.Date(
			nowInTz.Year(), nowInTz.Month(), nowInTz.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, loc,
		)
		instanceEnd := time.Date(
			nowInTz.Year(), nowInTz.Month(), nowInTz.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, loc,
		)

		// Check if we're within the active window (including grace period)
		if nowInTz.After(instanceStart) && nowInTz.Before(instanceEnd.Add(gracePeriod)) {
			return &sched
		}
	}

	return nil
}

// GetNextClassForRoom returns the next N upcoming class instances for a room.
// Returns both one-time and recurring schedule instances, sorted by start time.
func GetNextClassForRoom(tx *vbolt.Tx, roomId int, now time.Time, limit int) []ClassScheduleWithInstance {
	// Get all active schedules for this room
	var scheduleIds []int
	vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, roomId, &scheduleIds, vbolt.Window{})

	allInstances := make([]ClassScheduleWithInstance, 0)

	for _, schedId := range scheduleIds {
		var sched ClassSchedule
		if !vbolt.Read(tx, ClassSchedulesBkt, schedId, &sched) {
			continue
		}

		// Skip inactive schedules
		if !sched.IsActive {
			continue
		}

		if !sched.IsRecurring {
			// One-time schedule: include if in the future
			if sched.StartTime.After(now) {
				allInstances = append(allInstances, ClassScheduleWithInstance{
					Schedule:      sched,
					InstanceStart: sched.StartTime,
					InstanceEnd:   sched.EndTime,
				})
			}
		} else {
			// Recurring schedule: get upcoming instances
			instances := calculateUpcomingInstances(&sched, limit)
			for _, inst := range instances {
				allInstances = append(allInstances, ClassScheduleWithInstance{
					Schedule:      sched,
					InstanceStart: inst.StartTime,
					InstanceEnd:   inst.EndTime,
				})
			}
		}
	}

	// Sort by start time
	for i := 0; i < len(allInstances); i++ {
		for j := i + 1; j < len(allInstances); j++ {
			if allInstances[j].InstanceStart.Before(allInstances[i].InstanceStart) {
				allInstances[i], allInstances[j] = allInstances[j], allInstances[i]
			}
		}
	}

	// Limit results
	if len(allInstances) > limit {
		allInstances = allInstances[:limit]
	}

	return allInstances
}

// ClassScheduleWithInstance combines a schedule with a specific instance time
type ClassScheduleWithInstance struct {
	Schedule      ClassSchedule `json:"schedule"`
	InstanceStart time.Time     `json:"instanceStart"`
	InstanceEnd   time.Time     `json:"instanceEnd"`
}

func UpdateClassSchedule(ctx *vbeam.Context, req UpdateClassScheduleRequest) (resp UpdateClassScheduleResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Get existing schedule
	var schedule ClassSchedule
	vbolt.Read(ctx.Tx, ClassSchedulesBkt, req.ScheduleId, &schedule)
	if schedule.Id == 0 {
		return resp, errors.New("schedule not found")
	}

	// Check permission (Admin+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, schedule.StudioId, StudioRoleAdmin) {
		return resp, errors.New("admin permission required")
	}

	// Update only provided fields
	if req.Name != nil {
		if *req.Name == "" {
			return resp, errors.New("name cannot be empty")
		}
		schedule.Name = *req.Name
	}

	if req.Description != nil {
		schedule.Description = *req.Description
	}

	// Update one-time schedule fields (parse from strings)
	if req.StartTime != nil {
		if *req.StartTime == "" {
			return resp, errors.New("start time cannot be empty")
		}
		parsedTime, err := time.Parse(time.RFC3339, *req.StartTime)
		if err != nil {
			return resp, errors.New("invalid start time format")
		}
		schedule.StartTime = parsedTime
	}
	if req.EndTime != nil {
		if *req.EndTime == "" {
			return resp, errors.New("end time cannot be empty")
		}
		parsedTime, err := time.Parse(time.RFC3339, *req.EndTime)
		if err != nil {
			return resp, errors.New("invalid end time format")
		}
		schedule.EndTime = parsedTime
	}

	// Validate one-time schedule times if both are being modified
	if !schedule.IsRecurring {
		if !schedule.StartTime.IsZero() && !schedule.EndTime.IsZero() {
			if schedule.EndTime.Before(schedule.StartTime) || schedule.EndTime.Equal(schedule.StartTime) {
				return resp, errors.New("end time must be after start time")
			}
		}
	}

	// Update recurring schedule fields (parse from strings)
	if req.RecurStartDate != nil {
		if *req.RecurStartDate == "" {
			return resp, errors.New("recur start date cannot be empty")
		}
		parsedTime, err := time.Parse(time.RFC3339, *req.RecurStartDate)
		if err != nil {
			return resp, errors.New("invalid recur start date format")
		}
		schedule.RecurStartDate = parsedTime
	}
	if req.RecurEndDate != nil {
		if *req.RecurEndDate != "" {
			parsedTime, err := time.Parse(time.RFC3339, *req.RecurEndDate)
			if err != nil {
				return resp, errors.New("invalid recur end date format")
			}
			schedule.RecurEndDate = parsedTime
		} else {
			// Allow clearing the end date by setting to zero time
			schedule.RecurEndDate = time.Time{}
		}
	}
	if req.RecurWeekdays != nil {
		if len(req.RecurWeekdays) == 0 && schedule.IsRecurring {
			return resp, errors.New("recurring schedule requires weekdays")
		}
		schedule.RecurWeekdays = req.RecurWeekdays
	}
	if req.RecurTimeStart != nil {
		if *req.RecurTimeStart == "" && schedule.IsRecurring {
			return resp, errors.New("recurring schedule requires start time")
		}
		schedule.RecurTimeStart = *req.RecurTimeStart
	}
	if req.RecurTimeEnd != nil {
		if *req.RecurTimeEnd == "" && schedule.IsRecurring {
			return resp, errors.New("recurring schedule requires end time")
		}
		schedule.RecurTimeEnd = *req.RecurTimeEnd
	}
	if req.RecurTimezone != nil {
		schedule.RecurTimezone = *req.RecurTimezone
	}

	// Update automation settings
	if req.PreRollMinutes != nil {
		schedule.PreRollMinutes = *req.PreRollMinutes
	}
	if req.PostRollMinutes != nil {
		schedule.PostRollMinutes = *req.PostRollMinutes
	}
	if req.AutoStartCamera != nil {
		schedule.AutoStartCamera = *req.AutoStartCamera
	}
	if req.AutoStopCamera != nil {
		schedule.AutoStopCamera = *req.AutoStopCamera
	}

	// Update timestamp
	schedule.UpdatedAt = time.Now()

	// Write to database
	vbeam.UseWriteTx(ctx)
	vbolt.Write(ctx.Tx, ClassSchedulesBkt, schedule.Id, &schedule)
	vbolt.TxCommit(ctx.Tx)

	resp.Success = true
	return
}

func DeleteClassSchedule(ctx *vbeam.Context, req DeleteClassScheduleRequest) (resp DeleteClassScheduleResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Get existing schedule
	var schedule ClassSchedule
	vbolt.Read(ctx.Tx, ClassSchedulesBkt, req.ScheduleId, &schedule)
	if schedule.Id == 0 {
		return resp, errors.New("schedule not found")
	}

	// Check permission (Admin+ required)
	if !HasStudioPermission(ctx.Tx, caller.Id, schedule.StudioId, StudioRoleAdmin) {
		return resp, errors.New("admin permission required")
	}

	// Soft delete - set IsActive to false
	vbeam.UseWriteTx(ctx)
	schedule.IsActive = false
	schedule.UpdatedAt = time.Now()
	vbolt.Write(ctx.Tx, ClassSchedulesBkt, schedule.Id, &schedule)
	vbolt.TxCommit(ctx.Tx)

	resp.Success = true
	return
}

// RegisterClassScheduleMethods registers class schedule API procedures
func RegisterClassScheduleMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, CreateClassSchedule)
	vbeam.RegisterProc(app, ListClassSchedules)
	vbeam.RegisterProc(app, GetScheduleDetails)
	vbeam.RegisterProc(app, UpdateClassSchedule)
	vbeam.RegisterProc(app, DeleteClassSchedule)
}
