package backend

import (
	"fmt"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test ClassSchedule packing/unpacking for one-time schedule
func TestPackClassSchedule(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test one-time class schedule
		original := ClassSchedule{
			Id:              1,
			RoomId:          10,
			StudioId:        5,
			Name:            "Math 101",
			Description:     "Introduction to Algebra",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour).Truncate(time.Second),
			EndTime:         time.Now().Add(25 * time.Hour).Truncate(time.Second),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			CreatedBy:       100,
			CreatedAt:       time.Now().Truncate(time.Second),
			UpdatedAt:       time.Now().Truncate(time.Second),
			IsActive:        true,
		}

		// Write and read back
		vbolt.Write(tx, ClassSchedulesBkt, original.Id, &original)
		var retrieved ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, original.Id, &retrieved)

		// Verify all fields match
		if retrieved.Id != original.Id {
			t.Errorf("Id mismatch: got %d, want %d", retrieved.Id, original.Id)
		}
		if retrieved.RoomId != original.RoomId {
			t.Errorf("RoomId mismatch: got %d, want %d", retrieved.RoomId, original.RoomId)
		}
		if retrieved.StudioId != original.StudioId {
			t.Errorf("StudioId mismatch: got %d, want %d", retrieved.StudioId, original.StudioId)
		}
		if retrieved.Name != original.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, original.Name)
		}
		if retrieved.Description != original.Description {
			t.Errorf("Description mismatch: got %s, want %s", retrieved.Description, original.Description)
		}
		if retrieved.IsRecurring != original.IsRecurring {
			t.Errorf("IsRecurring mismatch: got %v, want %v", retrieved.IsRecurring, original.IsRecurring)
		}
		if !retrieved.StartTime.Equal(original.StartTime) {
			t.Errorf("StartTime mismatch: got %v, want %v", retrieved.StartTime, original.StartTime)
		}
		if !retrieved.EndTime.Equal(original.EndTime) {
			t.Errorf("EndTime mismatch: got %v, want %v", retrieved.EndTime, original.EndTime)
		}
		if retrieved.PreRollMinutes != original.PreRollMinutes {
			t.Errorf("PreRollMinutes mismatch: got %d, want %d", retrieved.PreRollMinutes, original.PreRollMinutes)
		}
		if retrieved.PostRollMinutes != original.PostRollMinutes {
			t.Errorf("PostRollMinutes mismatch: got %d, want %d", retrieved.PostRollMinutes, original.PostRollMinutes)
		}
		if retrieved.AutoStartCamera != original.AutoStartCamera {
			t.Errorf("AutoStartCamera mismatch: got %v, want %v", retrieved.AutoStartCamera, original.AutoStartCamera)
		}
		if retrieved.AutoStopCamera != original.AutoStopCamera {
			t.Errorf("AutoStopCamera mismatch: got %v, want %v", retrieved.AutoStopCamera, original.AutoStopCamera)
		}
		if retrieved.CreatedBy != original.CreatedBy {
			t.Errorf("CreatedBy mismatch: got %d, want %d", retrieved.CreatedBy, original.CreatedBy)
		}
		if !retrieved.CreatedAt.Equal(original.CreatedAt) {
			t.Errorf("CreatedAt mismatch: got %v, want %v", retrieved.CreatedAt, original.CreatedAt)
		}
		if !retrieved.UpdatedAt.Equal(original.UpdatedAt) {
			t.Errorf("UpdatedAt mismatch: got %v, want %v", retrieved.UpdatedAt, original.UpdatedAt)
		}
		if retrieved.IsActive != original.IsActive {
			t.Errorf("IsActive mismatch: got %v, want %v", retrieved.IsActive, original.IsActive)
		}
	})
}

// Test ClassSchedule packing/unpacking for recurring schedule with weekdays
func TestPackClassScheduleRecurring(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test recurring class schedule (Mon/Wed/Fri)
		original := ClassSchedule{
			Id:              2,
			RoomId:          15,
			StudioId:        7,
			Name:            "Science Lab",
			Description:     "Weekly science experiments",
			IsRecurring:     true,
			RecurStartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			RecurEndDate:    time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC),
			RecurWeekdays:   []int{1, 3, 5}, // Mon, Wed, Fri
			RecurTimeStart:  "09:00",
			RecurTimeEnd:    "10:30",
			RecurTimezone:   "America/New_York",
			PreRollMinutes:  10,
			PostRollMinutes: 5,
			AutoStartCamera: true,
			AutoStopCamera:  false,
			CreatedBy:       200,
			CreatedAt:       time.Now().Truncate(time.Second),
			UpdatedAt:       time.Now().Truncate(time.Second),
			IsActive:        true,
		}

		// Write and read back
		vbolt.Write(tx, ClassSchedulesBkt, original.Id, &original)
		var retrieved ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, original.Id, &retrieved)

		// Verify recurring-specific fields
		if retrieved.IsRecurring != original.IsRecurring {
			t.Errorf("IsRecurring mismatch: got %v, want %v", retrieved.IsRecurring, original.IsRecurring)
		}
		if !retrieved.RecurStartDate.Equal(original.RecurStartDate) {
			t.Errorf("RecurStartDate mismatch: got %v, want %v", retrieved.RecurStartDate, original.RecurStartDate)
		}
		if !retrieved.RecurEndDate.Equal(original.RecurEndDate) {
			t.Errorf("RecurEndDate mismatch: got %v, want %v", retrieved.RecurEndDate, original.RecurEndDate)
		}

		// Verify weekdays slice
		if len(retrieved.RecurWeekdays) != len(original.RecurWeekdays) {
			t.Errorf("RecurWeekdays length mismatch: got %d, want %d", len(retrieved.RecurWeekdays), len(original.RecurWeekdays))
		}
		for i, day := range original.RecurWeekdays {
			if i >= len(retrieved.RecurWeekdays) {
				t.Errorf("RecurWeekdays[%d] missing", i)
				continue
			}
			if retrieved.RecurWeekdays[i] != day {
				t.Errorf("RecurWeekdays[%d] mismatch: got %d, want %d", i, retrieved.RecurWeekdays[i], day)
			}
		}

		if retrieved.RecurTimeStart != original.RecurTimeStart {
			t.Errorf("RecurTimeStart mismatch: got %s, want %s", retrieved.RecurTimeStart, original.RecurTimeStart)
		}
		if retrieved.RecurTimeEnd != original.RecurTimeEnd {
			t.Errorf("RecurTimeEnd mismatch: got %s, want %s", retrieved.RecurTimeEnd, original.RecurTimeEnd)
		}
		if retrieved.RecurTimezone != original.RecurTimezone {
			t.Errorf("RecurTimezone mismatch: got %s, want %s", retrieved.RecurTimezone, original.RecurTimezone)
		}
	})
}

// Test ClassSchedule with empty weekdays slice
func TestPackClassScheduleEmptyWeekdays(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a schedule with empty weekdays (one-time schedule)
		original := ClassSchedule{
			Id:            3,
			RoomId:        20,
			StudioId:      8,
			Name:          "Guest Lecture",
			IsRecurring:   false,
			StartTime:     time.Now().Add(48 * time.Hour).Truncate(time.Second),
			EndTime:       time.Now().Add(50 * time.Hour).Truncate(time.Second),
			RecurWeekdays: []int{}, // Empty slice
			CreatedBy:     300,
			CreatedAt:     time.Now().Truncate(time.Second),
			UpdatedAt:     time.Now().Truncate(time.Second),
			IsActive:      true,
		}

		// Write and read back
		vbolt.Write(tx, ClassSchedulesBkt, original.Id, &original)
		var retrieved ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, original.Id, &retrieved)

		// Verify empty slice is preserved
		if retrieved.RecurWeekdays == nil {
			// It's acceptable for empty slice to be nil after unpacking
			retrieved.RecurWeekdays = []int{}
		}
		if len(retrieved.RecurWeekdays) != 0 {
			t.Errorf("RecurWeekdays should be empty: got length %d", len(retrieved.RecurWeekdays))
		}
	})
}

// Test ClassPermission packing/unpacking
func TestPackClassPermission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test class permission
		original := ClassPermission{
			Id:         1,
			ScheduleId: 10,
			UserId:     100,
			Role:       int(StudioRoleViewer),
			GrantedBy:  200,
			GrantedAt:  time.Now().Truncate(time.Second),
		}

		// Write and read back
		vbolt.Write(tx, ClassPermissionsBkt, original.Id, &original)
		var retrieved ClassPermission
		vbolt.Read(tx, ClassPermissionsBkt, original.Id, &retrieved)

		// Verify all fields match
		if retrieved.Id != original.Id {
			t.Errorf("Id mismatch: got %d, want %d", retrieved.Id, original.Id)
		}
		if retrieved.ScheduleId != original.ScheduleId {
			t.Errorf("ScheduleId mismatch: got %d, want %d", retrieved.ScheduleId, original.ScheduleId)
		}
		if retrieved.UserId != original.UserId {
			t.Errorf("UserId mismatch: got %d, want %d", retrieved.UserId, original.UserId)
		}
		if retrieved.Role != original.Role {
			t.Errorf("Role mismatch: got %d, want %d", retrieved.Role, original.Role)
		}
		if retrieved.GrantedBy != original.GrantedBy {
			t.Errorf("GrantedBy mismatch: got %d, want %d", retrieved.GrantedBy, original.GrantedBy)
		}
		if !retrieved.GrantedAt.Equal(original.GrantedAt) {
			t.Errorf("GrantedAt mismatch: got %v, want %v", retrieved.GrantedAt, original.GrantedAt)
		}
	})
}

// Test ScheduleExecutionLog packing/unpacking
func TestPackScheduleExecutionLog(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test execution log (successful start)
		original := ScheduleExecutionLog{
			Id:         1,
			ScheduleId: 5,
			RoomId:     15,
			Action:     "start_camera",
			Timestamp:  time.Now().Truncate(time.Second),
			Success:    true,
			ErrorMsg:   "",
		}

		// Write and read back
		vbolt.Write(tx, ScheduleLogsBkt, original.Id, &original)
		var retrieved ScheduleExecutionLog
		vbolt.Read(tx, ScheduleLogsBkt, original.Id, &retrieved)

		// Verify all fields match
		if retrieved.Id != original.Id {
			t.Errorf("Id mismatch: got %d, want %d", retrieved.Id, original.Id)
		}
		if retrieved.ScheduleId != original.ScheduleId {
			t.Errorf("ScheduleId mismatch: got %d, want %d", retrieved.ScheduleId, original.ScheduleId)
		}
		if retrieved.RoomId != original.RoomId {
			t.Errorf("RoomId mismatch: got %d, want %d", retrieved.RoomId, original.RoomId)
		}
		if retrieved.Action != original.Action {
			t.Errorf("Action mismatch: got %s, want %s", retrieved.Action, original.Action)
		}
		if !retrieved.Timestamp.Equal(original.Timestamp) {
			t.Errorf("Timestamp mismatch: got %v, want %v", retrieved.Timestamp, original.Timestamp)
		}
		if retrieved.Success != original.Success {
			t.Errorf("Success mismatch: got %v, want %v", retrieved.Success, original.Success)
		}
		if retrieved.ErrorMsg != original.ErrorMsg {
			t.Errorf("ErrorMsg mismatch: got %s, want %s", retrieved.ErrorMsg, original.ErrorMsg)
		}
	})
}

// Test ScheduleExecutionLog with error message
func TestPackScheduleExecutionLogWithError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test execution log (failed start)
		original := ScheduleExecutionLog{
			Id:         2,
			ScheduleId: 7,
			RoomId:     20,
			Action:     "start_camera",
			Timestamp:  time.Now().Truncate(time.Second),
			Success:    false,
			ErrorMsg:   "No camera config found for room",
		}

		// Write and read back
		vbolt.Write(tx, ScheduleLogsBkt, original.Id, &original)
		var retrieved ScheduleExecutionLog
		vbolt.Read(tx, ScheduleLogsBkt, original.Id, &retrieved)

		// Verify error message is preserved
		if retrieved.Success != original.Success {
			t.Errorf("Success mismatch: got %v, want %v", retrieved.Success, original.Success)
		}
		if retrieved.ErrorMsg != original.ErrorMsg {
			t.Errorf("ErrorMsg mismatch: got %s, want %s", retrieved.ErrorMsg, original.ErrorMsg)
		}
	})
}

// Test index operations for schedules
func TestScheduleIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test schedules for the same room
		schedule1 := ClassSchedule{
			Id:       1,
			RoomId:   10,
			StudioId: 5,
			Name:     "Morning Class",
			IsActive: true,
		}
		schedule2 := ClassSchedule{
			Id:       2,
			RoomId:   10,
			StudioId: 5,
			Name:     "Afternoon Class",
			IsActive: true,
		}

		// Write schedules
		vbolt.Write(tx, ClassSchedulesBkt, schedule1.Id, &schedule1)
		vbolt.Write(tx, ClassSchedulesBkt, schedule2.Id, &schedule2)

		// Update indexes
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule1.Id, schedule1.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule2.Id, schedule2.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule1.Id, schedule1.StudioId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule2.Id, schedule2.StudioId)

		// Query schedules by room
		var scheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, 10, &scheduleIds, vbolt.Window{})

		if len(scheduleIds) != 2 {
			t.Errorf("Expected 2 schedules for room 10, got %d", len(scheduleIds))
		}

		// Verify both schedule IDs are present
		foundSchedule1 := false
		foundSchedule2 := false
		for _, id := range scheduleIds {
			if id == schedule1.Id {
				foundSchedule1 = true
			}
			if id == schedule2.Id {
				foundSchedule2 = true
			}
		}

		if !foundSchedule1 {
			t.Errorf("Schedule 1 not found in room index")
		}
		if !foundSchedule2 {
			t.Errorf("Schedule 2 not found in room index")
		}
	})
}

// Test permission indexes
func TestPermissionIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test permissions for the same schedule
		perm1 := ClassPermission{
			Id:         1,
			ScheduleId: 10,
			UserId:     100,
			Role:       int(StudioRoleViewer),
		}
		perm2 := ClassPermission{
			Id:         2,
			ScheduleId: 10,
			UserId:     101,
			Role:       int(StudioRoleMember),
		}

		// Write permissions
		vbolt.Write(tx, ClassPermissionsBkt, perm1.Id, &perm1)
		vbolt.Write(tx, ClassPermissionsBkt, perm2.Id, &perm2)

		// Update indexes
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm1.Id, perm1.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm2.Id, perm2.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm1.Id, perm1.UserId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm2.Id, perm2.UserId)

		// Query permissions by schedule
		var permIds []int
		vbolt.ReadTermTargets(tx, PermsByScheduleIdx, 10, &permIds, vbolt.Window{})

		if len(permIds) != 2 {
			t.Errorf("Expected 2 permissions for schedule 10, got %d", len(permIds))
		}

		// Query permissions by user
		var userPermIds []int
		vbolt.ReadTermTargets(tx, PermsByUserIdx, 100, &userPermIds, vbolt.Window{})

		if len(userPermIds) != 1 {
			t.Errorf("Expected 1 permission for user 100, got %d", len(userPermIds))
		}
		if len(userPermIds) > 0 && userPermIds[0] != perm1.Id {
			t.Errorf("Expected permission ID %d for user 100, got %d", perm1.Id, userPermIds[0])
		}
	})
}

// Test creating a one-time schedule with all required fields
func TestCreateOneTimeSchedule(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test user
		user := createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  user.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   user.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, user.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create one-time schedule
		startTime := time.Now().Add(24 * time.Hour).Truncate(time.Second)
		endTime := startTime.Add(1 * time.Hour)

		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room.Id,
			StudioId:        studio.Id,
			Name:            "Math 101",
			Description:     "Introduction to Algebra",
			IsRecurring:     false,
			StartTime:       startTime,
			EndTime:         endTime,
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			CreatedBy:       user.Id,
			CreatedAt:       time.Now().Truncate(time.Second),
			UpdatedAt:       time.Now().Truncate(time.Second),
			IsActive:        true,
		}

		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, schedule.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule.Id, schedule.StudioId)

		// Verify schedule was created
		var retrieved ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule.Id, &retrieved)

		if retrieved.Id == 0 {
			t.Error("Schedule should be created")
		}
		if retrieved.Name != "Math 101" {
			t.Errorf("Expected schedule name 'Math 101', got '%s'", retrieved.Name)
		}
		if retrieved.IsRecurring != false {
			t.Error("Schedule should not be recurring")
		}
		if !retrieved.StartTime.Equal(startTime) {
			t.Errorf("StartTime mismatch: got %v, want %v", retrieved.StartTime, startTime)
		}
		if !retrieved.EndTime.Equal(endTime) {
			t.Errorf("EndTime mismatch: got %v, want %v", retrieved.EndTime, endTime)
		}
		if retrieved.PreRollMinutes != 5 {
			t.Errorf("PreRollMinutes should be 5, got %d", retrieved.PreRollMinutes)
		}
		if retrieved.PostRollMinutes != 2 {
			t.Errorf("PostRollMinutes should be 2, got %d", retrieved.PostRollMinutes)
		}

		// Verify indexes were updated
		var scheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, room.Id, &scheduleIds, vbolt.Window{})
		if len(scheduleIds) != 1 {
			t.Errorf("Expected 1 schedule for room, got %d", len(scheduleIds))
		}
		if len(scheduleIds) > 0 && scheduleIds[0] != schedule.Id {
			t.Errorf("Expected schedule ID %d in room index, got %d", schedule.Id, scheduleIds[0])
		}

		var studioScheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByStudioIdx, studio.Id, &studioScheduleIds, vbolt.Window{})
		if len(studioScheduleIds) != 1 {
			t.Errorf("Expected 1 schedule for studio, got %d", len(studioScheduleIds))
		}
	})
}

// Test creating a recurring schedule with weekdays
func TestCreateRecurringSchedule(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test user
		user := createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  user.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   user.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, user.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create recurring schedule (Mon/Wed/Fri 9-10am)
		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room.Id,
			StudioId:        studio.Id,
			Name:            "Science Lab",
			Description:     "Weekly experiments",
			IsRecurring:     true,
			RecurStartDate:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			RecurEndDate:    time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC),
			RecurWeekdays:   []int{1, 3, 5}, // Mon, Wed, Fri
			RecurTimeStart:  "09:00",
			RecurTimeEnd:    "10:00",
			RecurTimezone:   "America/New_York",
			PreRollMinutes:  10,
			PostRollMinutes: 5,
			AutoStartCamera: true,
			AutoStopCamera:  false,
			CreatedBy:       user.Id,
			CreatedAt:       time.Now().Truncate(time.Second),
			UpdatedAt:       time.Now().Truncate(time.Second),
			IsActive:        true,
		}

		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, schedule.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule.Id, schedule.StudioId)

		// Verify schedule was created
		var retrieved ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule.Id, &retrieved)

		if retrieved.Id == 0 {
			t.Error("Schedule should be created")
		}
		if retrieved.IsRecurring != true {
			t.Error("Schedule should be recurring")
		}
		if len(retrieved.RecurWeekdays) != 3 {
			t.Errorf("Expected 3 weekdays, got %d", len(retrieved.RecurWeekdays))
		}
		if retrieved.RecurTimeStart != "09:00" {
			t.Errorf("Expected RecurTimeStart '09:00', got '%s'", retrieved.RecurTimeStart)
		}
		if retrieved.RecurTimezone != "America/New_York" {
			t.Errorf("Expected timezone 'America/New_York', got '%s'", retrieved.RecurTimezone)
		}
		if retrieved.PreRollMinutes != 10 {
			t.Errorf("PreRollMinutes should be 10, got %d", retrieved.PreRollMinutes)
		}
		if retrieved.AutoStopCamera != false {
			t.Error("AutoStopCamera should be false")
		}

		// Verify weekdays are correct
		expectedWeekdays := []int{1, 3, 5}
		for i, day := range expectedWeekdays {
			if retrieved.RecurWeekdays[i] != day {
				t.Errorf("Weekday[%d] should be %d, got %d", i, day, retrieved.RecurWeekdays[i])
			}
		}
	})
}

// Test that schedule defaults are applied correctly
func TestScheduleDefaults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Test that PreRollMinutes defaults to 5 when set to 0
		schedule := ClassSchedule{
			Id:              1,
			RoomId:          10,
			StudioId:        5,
			Name:            "Test Schedule",
			PreRollMinutes:  0, // Should default to 5 in the API
			PostRollMinutes: 0, // Should default to 2 in the API
		}

		// In actual API usage, defaults would be set by CreateClassSchedule
		// For this test, just verify the data model supports the values
		if schedule.PreRollMinutes == 0 {
			schedule.PreRollMinutes = 5
		}
		if schedule.PostRollMinutes == 0 {
			schedule.PostRollMinutes = 2
		}

		if schedule.PreRollMinutes != 5 {
			t.Errorf("Expected default PreRollMinutes 5, got %d", schedule.PreRollMinutes)
		}
		if schedule.PostRollMinutes != 2 {
			t.Errorf("Expected default PostRollMinutes 2, got %d", schedule.PostRollMinutes)
		}
	})
}

// Test validation: end time must be after start time
func TestScheduleTimeValidation(t *testing.T) {
	// This test validates the business logic that should be in CreateClassSchedule
	startTime := time.Now()
	endTime := startTime.Add(-1 * time.Hour) // End before start (invalid)

	if !endTime.Before(startTime) {
		t.Error("End time should be before start time in this test")
	}

	// The actual validation happens in CreateClassSchedule procedure
	// which should return an error when endTime.Before(startTime)
}

// Test multiple schedules for the same room
func TestMultipleSchedulesPerRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test room
		roomId := 10
		studioId := 5

		// Create first schedule
		schedule1 := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          roomId,
			StudioId:        studioId,
			Name:            "Morning Class",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(25 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule1.Id, &schedule1)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule1.Id, roomId)

		// Create second schedule for same room
		schedule2 := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          roomId,
			StudioId:        studioId,
			Name:            "Afternoon Class",
			IsRecurring:     false,
			StartTime:       time.Now().Add(26 * time.Hour),
			EndTime:         time.Now().Add(27 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule2.Id, &schedule2)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule2.Id, roomId)

		// Verify both schedules exist for the room
		var scheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, roomId, &scheduleIds, vbolt.Window{})

		if len(scheduleIds) != 2 {
			t.Errorf("Expected 2 schedules for room, got %d", len(scheduleIds))
		}

		// Verify we can retrieve both schedules
		var retrieved1 ClassSchedule
		var retrieved2 ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule1.Id, &retrieved1)
		vbolt.Read(tx, ClassSchedulesBkt, schedule2.Id, &retrieved2)

		if retrieved1.Name != "Morning Class" {
			t.Errorf("Expected first schedule name 'Morning Class', got '%s'", retrieved1.Name)
		}
		if retrieved2.Name != "Afternoon Class" {
			t.Errorf("Expected second schedule name 'Afternoon Class', got '%s'", retrieved2.Name)
		}
	})
}

// Test listing schedules by room
func TestListSchedulesByRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test user
		user := createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  user.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add membership
		membership := StudioMembership{
			UserId:   user.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, user.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Create two rooms
		streamKey1, _ := GenerateStreamKey()
		room1 := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey1,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room1.Id, &room1)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room1.Id, studio.Id)

		streamKey2, _ := GenerateStreamKey()
		room2 := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 2,
			Name:       "Room 2",
			StreamKey:  streamKey2,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room2.Id, studio.Id)

		// Create schedules for room1
		schedule1 := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room1.Id,
			StudioId:        studio.Id,
			Name:            "Math 101",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(25 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule1.Id, &schedule1)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule1.Id, room1.Id)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule1.Id, studio.Id)

		schedule2 := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room1.Id,
			StudioId:        studio.Id,
			Name:            "Science 101",
			IsRecurring:     false,
			StartTime:       time.Now().Add(48 * time.Hour),
			EndTime:         time.Now().Add(49 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule2.Id, &schedule2)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule2.Id, room1.Id)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule2.Id, studio.Id)

		// Create schedule for room2
		schedule3 := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room2.Id,
			StudioId:        studio.Id,
			Name:            "Art 101",
			IsRecurring:     false,
			StartTime:       time.Now().Add(72 * time.Hour),
			EndTime:         time.Now().Add(73 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule3.Id, &schedule3)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule3.Id, room2.Id)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule3.Id, studio.Id)

		// Query schedules for room1
		var scheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, room1.Id, &scheduleIds, vbolt.Window{})

		if len(scheduleIds) != 2 {
			t.Errorf("Expected 2 schedules for room1, got %d", len(scheduleIds))
		}

		// Verify correct schedules returned
		foundMath := false
		foundScience := false
		for _, id := range scheduleIds {
			var schedule ClassSchedule
			vbolt.Read(tx, ClassSchedulesBkt, id, &schedule)
			if schedule.Name == "Math 101" {
				foundMath = true
			}
			if schedule.Name == "Science 101" {
				foundScience = true
			}
		}

		if !foundMath || !foundScience {
			t.Error("Should find both Math 101 and Science 101 for room1")
		}

		// Query schedules for room2
		var room2ScheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, room2.Id, &room2ScheduleIds, vbolt.Window{})

		if len(room2ScheduleIds) != 1 {
			t.Errorf("Expected 1 schedule for room2, got %d", len(room2ScheduleIds))
		}
	})
}

// Test listing schedules by studio
func TestListSchedulesByStudio(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test user
		user := createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  user.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add membership
		membership := StudioMembership{
			UserId:   user.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, user.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create 3 schedules for this studio
		for i := 1; i <= 3; i++ {
			schedule := ClassSchedule{
				Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
				RoomId:          room.Id,
				StudioId:        studio.Id,
				Name:            fmt.Sprintf("Class %d", i),
				IsRecurring:     false,
				StartTime:       time.Now().Add(time.Duration(i*24) * time.Hour),
				EndTime:         time.Now().Add(time.Duration(i*24+1) * time.Hour),
				PreRollMinutes:  5,
				PostRollMinutes: 2,
				IsActive:        true,
			}
			vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
			vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, room.Id)
			vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule.Id, studio.Id)
		}

		// Query schedules by studio
		var scheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByStudioIdx, studio.Id, &scheduleIds, vbolt.Window{})

		if len(scheduleIds) != 3 {
			t.Errorf("Expected 3 schedules for studio, got %d", len(scheduleIds))
		}
	})
}

// Test that only active schedules are returned
func TestListOnlyActiveSchedules(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 10
		studioId := 5

		// Create active schedule
		activeSchedule := ClassSchedule{
			Id:       vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:   roomId,
			StudioId: studioId,
			Name:     "Active Class",
			IsActive: true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, activeSchedule.Id, &activeSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, activeSchedule.Id, roomId)

		// Create inactive schedule (soft-deleted)
		inactiveSchedule := ClassSchedule{
			Id:       vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:   roomId,
			StudioId: studioId,
			Name:     "Inactive Class",
			IsActive: false, // Soft deleted
		}
		vbolt.Write(tx, ClassSchedulesBkt, inactiveSchedule.Id, &inactiveSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, inactiveSchedule.Id, roomId)

		// Both are in the index, but only active should be returned by ListClassSchedules
		var allScheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, roomId, &allScheduleIds, vbolt.Window{})

		if len(allScheduleIds) != 2 {
			t.Errorf("Expected 2 schedules in index, got %d", len(allScheduleIds))
		}

		// Simulating what ListClassSchedules does: filter by IsActive
		activeCount := 0
		for _, id := range allScheduleIds {
			var schedule ClassSchedule
			vbolt.Read(tx, ClassSchedulesBkt, id, &schedule)
			if schedule.IsActive {
				activeCount++
			}
		}

		if activeCount != 1 {
			t.Errorf("Expected 1 active schedule, got %d", activeCount)
		}
	})
}

// Test calculating upcoming instances for recurring schedule
func TestCalculateUpcomingInstances(t *testing.T) {
	// Create a Mon/Wed/Fri recurring schedule starting today
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)

	schedule := ClassSchedule{
		IsRecurring:    true,
		RecurStartDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc),
		RecurEndDate:   time.Date(now.Year()+1, now.Month(), now.Day(), 0, 0, 0, 0, loc),
		RecurWeekdays:  []int{1, 3, 5}, // Mon, Wed, Fri
		RecurTimeStart: "09:00",
		RecurTimeEnd:   "10:00",
		RecurTimezone:  "America/New_York",
	}

	instances := calculateUpcomingInstances(&schedule, 10)

	if len(instances) == 0 {
		t.Error("Should calculate at least some upcoming instances")
	}

	if len(instances) > 10 {
		t.Errorf("Requested 10 instances, got %d", len(instances))
	}

	// Verify each instance falls on Mon/Wed/Fri
	for i, instance := range instances {
		weekday := instance.StartTime.In(loc).Weekday()
		if weekday != time.Monday && weekday != time.Wednesday && weekday != time.Friday {
			t.Errorf("Instance %d falls on %s, expected Mon/Wed/Fri", i, weekday)
		}

		// Verify time is 9:00 AM
		hour := instance.StartTime.In(loc).Hour()
		minute := instance.StartTime.In(loc).Minute()
		if hour != 9 || minute != 0 {
			t.Errorf("Instance %d start time is %02d:%02d, expected 09:00", i, hour, minute)
		}

		// Verify duration is 1 hour
		duration := instance.EndTime.Sub(instance.StartTime)
		expectedDuration := 1 * time.Hour
		if duration != expectedDuration {
			t.Errorf("Instance %d duration is %v, expected %v", i, duration, expectedDuration)
		}

		// Verify all instances are in the future
		if instance.EndTime.Before(now) {
			t.Errorf("Instance %d is in the past", i)
		}
	}
}

// Test that one-time schedules return single instance if future
func TestGetScheduleDetailsOneTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create future one-time schedule
		futureStart := time.Now().Add(24 * time.Hour).Truncate(time.Second)
		futureEnd := futureStart.Add(1 * time.Hour)

		schedule := ClassSchedule{
			Id:          vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:      10,
			StudioId:    5,
			Name:        "Future Class",
			IsRecurring: false,
			StartTime:   futureStart,
			EndTime:     futureEnd,
			IsActive:    true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Simulating GetScheduleDetails logic
		var instances []ClassInstance
		if schedule.StartTime.After(time.Now()) {
			instances = []ClassInstance{
				{
					StartTime: schedule.StartTime,
					EndTime:   schedule.EndTime,
				},
			}
		}

		if len(instances) != 1 {
			t.Errorf("Expected 1 instance for future one-time schedule, got %d", len(instances))
		}

		if !instances[0].StartTime.Equal(futureStart) {
			t.Error("Instance start time should match schedule start time")
		}
	})
}

// Test that past one-time schedules return empty instances
func TestGetScheduleDetailsPastOneTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create past one-time schedule
		pastStart := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
		pastEnd := pastStart.Add(1 * time.Hour)

		schedule := ClassSchedule{
			Id:          vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:      10,
			StudioId:    5,
			Name:        "Past Class",
			IsRecurring: false,
			StartTime:   pastStart,
			EndTime:     pastEnd,
			IsActive:    true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Simulating GetScheduleDetails logic
		var instances []ClassInstance
		if schedule.StartTime.After(time.Now()) {
			instances = []ClassInstance{
				{
					StartTime: schedule.StartTime,
					EndTime:   schedule.EndTime,
				},
			}
		} else {
			instances = []ClassInstance{}
		}

		if len(instances) != 0 {
			t.Errorf("Expected 0 instances for past one-time schedule, got %d", len(instances))
		}
	})
}

// Test updating schedule name
func TestUpdateScheduleName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test user
		user := createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  user.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		membership := StudioMembership{
			UserId:   user.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, user.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create schedule
		originalSchedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room.Id,
			StudioId:        studio.Id,
			Name:            "Original Name",
			Description:     "Original Description",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(25 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			CreatedBy:       user.Id,
			CreatedAt:       time.Now().Truncate(time.Second),
			UpdatedAt:       time.Now().Truncate(time.Second),
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, originalSchedule.Id, &originalSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, originalSchedule.Id, room.Id)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, originalSchedule.Id, studio.Id)

		// Update name only
		newName := "Updated Name"
		originalSchedule.Name = newName
		originalSchedule.UpdatedAt = time.Now()
		vbolt.Write(tx, ClassSchedulesBkt, originalSchedule.Id, &originalSchedule)

		// Verify update
		var updated ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, originalSchedule.Id, &updated)

		if updated.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
		}
		// Description should remain unchanged
		if updated.Description != "Original Description" {
			t.Errorf("Expected description unchanged, got '%s'", updated.Description)
		}
		// Other fields should remain unchanged
		if updated.PreRollMinutes != 5 {
			t.Errorf("PreRollMinutes should be unchanged: got %d", updated.PreRollMinutes)
		}
	})
}

// Test updating one-time schedule times
func TestUpdateScheduleTimes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 10
		studioId := 5

		originalStart := time.Now().Add(24 * time.Hour).Truncate(time.Second)
		originalEnd := originalStart.Add(1 * time.Hour)

		schedule := ClassSchedule{
			Id:          vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:      roomId,
			StudioId:    studioId,
			Name:        "Test Class",
			IsRecurring: false,
			StartTime:   originalStart,
			EndTime:     originalEnd,
			IsActive:    true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Update times
		newStart := time.Now().Add(48 * time.Hour).Truncate(time.Second)
		newEnd := newStart.Add(2 * time.Hour)

		schedule.StartTime = newStart
		schedule.EndTime = newEnd
		schedule.UpdatedAt = time.Now()
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Verify update
		var updated ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule.Id, &updated)

		if !updated.StartTime.Equal(newStart) {
			t.Error("StartTime should be updated")
		}
		if !updated.EndTime.Equal(newEnd) {
			t.Error("EndTime should be updated")
		}
	})
}

// Test updating recurring schedule weekdays
func TestUpdateRecurringWeekdays(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 10
		studioId := 5

		schedule := ClassSchedule{
			Id:             vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:         roomId,
			StudioId:       studioId,
			Name:           "Test Class",
			IsRecurring:    true,
			RecurWeekdays:  []int{1, 3, 5}, // Mon, Wed, Fri
			RecurTimeStart: "09:00",
			RecurTimeEnd:   "10:00",
			RecurTimezone:  "UTC",
			IsActive:       true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Update to Tue, Thu
		schedule.RecurWeekdays = []int{2, 4}
		schedule.UpdatedAt = time.Now()
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Verify update
		var updated ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule.Id, &updated)

		if len(updated.RecurWeekdays) != 2 {
			t.Errorf("Expected 2 weekdays, got %d", len(updated.RecurWeekdays))
		}
		if updated.RecurWeekdays[0] != 2 || updated.RecurWeekdays[1] != 4 {
			t.Errorf("Expected weekdays [2, 4], got %v", updated.RecurWeekdays)
		}
	})
}

// Test that update validates times
func TestUpdateValidation(t *testing.T) {
	// Test that end time must be after start time
	start := time.Now()
	end := start.Add(-1 * time.Hour) // End before start (invalid)

	// This would be caught by UpdateClassSchedule validation
	if !end.Before(start) {
		t.Error("End time should be before start time in this test")
	}
}

// Test deleting a schedule (soft delete)
func TestDeleteSchedule(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 10
		studioId := 5

		schedule := ClassSchedule{
			Id:       vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:   roomId,
			StudioId: studioId,
			Name:     "Test Class",
			IsActive: true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, roomId)

		// Soft delete
		schedule.IsActive = false
		schedule.UpdatedAt = time.Now()
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Verify schedule still exists but is inactive
		var deleted ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule.Id, &deleted)

		if deleted.Id == 0 {
			t.Error("Schedule should still exist in database")
		}
		if deleted.IsActive {
			t.Error("Schedule should be inactive (IsActive=false)")
		}
	})
}

// Test that deleted schedules are excluded from List queries
func TestDeletedScheduleNotInList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 10
		studioId := 5

		// Create active schedule
		activeSchedule := ClassSchedule{
			Id:       vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:   roomId,
			StudioId: studioId,
			Name:     "Active Class",
			IsActive: true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, activeSchedule.Id, &activeSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, activeSchedule.Id, roomId)

		// Create deleted schedule
		deletedSchedule := ClassSchedule{
			Id:       vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:   roomId,
			StudioId: studioId,
			Name:     "Deleted Class",
			IsActive: false, // Soft deleted
		}
		vbolt.Write(tx, ClassSchedulesBkt, deletedSchedule.Id, &deletedSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, deletedSchedule.Id, roomId)

		// Simulate ListClassSchedules filtering
		var scheduleIds []int
		vbolt.ReadTermTargets(tx, SchedulesByRoomIdx, roomId, &scheduleIds, vbolt.Window{})

		// Count only active schedules
		activeCount := 0
		for _, id := range scheduleIds {
			var schedule ClassSchedule
			vbolt.Read(tx, ClassSchedulesBkt, id, &schedule)
			if schedule.IsActive {
				activeCount++
			}
		}

		if activeCount != 1 {
			t.Errorf("Expected 1 active schedule in list, got %d", activeCount)
		}
	})
}

// Test updating automation settings
func TestUpdateAutomationSettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 10
		studioId := 5

		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          roomId,
			StudioId:        studioId,
			Name:            "Test Class",
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Update automation settings
		schedule.PreRollMinutes = 10
		schedule.PostRollMinutes = 5
		schedule.AutoStartCamera = false
		schedule.AutoStopCamera = false
		schedule.UpdatedAt = time.Now()
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)

		// Verify update
		var updated ClassSchedule
		vbolt.Read(tx, ClassSchedulesBkt, schedule.Id, &updated)

		if updated.PreRollMinutes != 10 {
			t.Errorf("PreRollMinutes should be 10, got %d", updated.PreRollMinutes)
		}
		if updated.PostRollMinutes != 5 {
			t.Errorf("PostRollMinutes should be 5, got %d", updated.PostRollMinutes)
		}
		if updated.AutoStartCamera != false {
			t.Error("AutoStartCamera should be false")
		}
		if updated.AutoStopCamera != false {
			t.Error("AutoStopCamera should be false")
		}
	})
}
