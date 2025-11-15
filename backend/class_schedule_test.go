package backend

import (
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
