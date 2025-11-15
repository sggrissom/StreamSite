package backend

import (
	"fmt"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Initialize camera manager for tests
func init() {
	if cameraManager == nil {
		cameraManager = NewCameraManager()
	}
}

// clearRecentActionsCache clears the global recent actions cache for testing
func clearRecentActionsCache() {
	recentActionsMu.Lock()
	defer recentActionsMu.Unlock()
	recentActions = make(map[string]time.Time)
}

// TestShouldStartCamera_OneTime tests camera start logic for one-time schedules
func TestShouldStartCamera_OneTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a one-time schedule for 2 hours from now
	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	endTime := startTime.Add(1 * time.Hour)

	schedule := ClassSchedule{
		RoomId:          1,
		IsRecurring:     false,
		StartTime:       startTime,
		EndTime:         endTime,
		PreRollMinutes:  5,
		PostRollMinutes: 2,
		AutoStartCamera: true,
		AutoStopCamera:  true,
		IsActive:        true,
	}

	// Test 1: Too early (before pre-roll)
	testTime := startTime.Add(-10 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should not start camera 10 minutes before pre-roll")
		}
	})

	// Test 2: During pre-roll window (should start)
	testTime = startTime.Add(-3 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should start camera during pre-roll window")
		}
	})

	// Test 3: During class time
	testTime = startTime.Add(30 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should start camera during class time")
		}
	})

	// Test 4: During post-roll
	testTime = endTime.Add(1 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should start camera during post-roll")
		}
	})

	// Test 5: After post-roll (too late)
	testTime = endTime.Add(5 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should not start camera after post-roll")
		}
	})

	// Test 6: Auto-start disabled
	schedule.AutoStartCamera = false
	testTime = startTime
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should not start camera when AutoStartCamera is false")
		}
	})
}

// TestShouldStartCamera_Recurring tests camera start logic for recurring schedules
func TestShouldStartCamera_Recurring(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().In(loc)
	currentWeekday := int(now.Weekday())

	// Create a recurring schedule for today at a specific time
	schedule := ClassSchedule{
		RoomId:          1,
		IsRecurring:     true,
		RecurWeekdays:   []int{currentWeekday},
		RecurTimeStart:  "14:00",
		RecurTimeEnd:    "15:00",
		RecurTimezone:   "America/New_York",
		PreRollMinutes:  5,
		PostRollMinutes: 2,
		AutoStartCamera: true,
		AutoStopCamera:  true,
		IsActive:        true,
	}

	// Test 1: Correct weekday, during pre-roll
	year, month, day := now.Date()
	testTime := time.Date(year, month, day, 13, 57, 0, 0, loc) // 3 minutes before start
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should start camera on correct weekday during pre-roll")
		}
	})

	// Test 2: Correct weekday, during class
	testTime = time.Date(year, month, day, 14, 30, 0, 0, loc)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should start camera on correct weekday during class")
		}
	})

	// Test 3: Wrong weekday
	wrongWeekday := (currentWeekday + 1) % 7
	schedule.RecurWeekdays = []int{wrongWeekday}
	testTime = time.Date(year, month, day, 14, 0, 0, 0, loc)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should not start camera on wrong weekday")
		}
	})

	// Test 4: Multiple weekdays including today
	schedule.RecurWeekdays = []int{currentWeekday, wrongWeekday}
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStartCamera(tx, &schedule, testTime) {
			t.Error("Should start camera when today is in weekday list")
		}
	})
}

// TestShouldStopCamera_OutsideWindow tests camera stop logic
func TestShouldStopCamera_OutsideWindow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	startTime := now.Add(-2 * time.Hour) // Class was 2 hours ago
	endTime := startTime.Add(1 * time.Hour)

	schedule := ClassSchedule{
		Id:              1,
		RoomId:          1,
		IsRecurring:     false,
		StartTime:       startTime,
		EndTime:         endTime,
		PreRollMinutes:  5,
		PostRollMinutes: 2,
		AutoStartCamera: true,
		AutoStopCamera:  true,
		IsActive:        true,
	}

	// Start the camera manually
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Set up test room and camera config
		room := Room{Id: 1, StudioId: 1, StreamKey: "test-key"}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		cameraConfig := CameraConfig{RoomId: 1, RTSPURL: "rtsp://127.0.0.1:65535/test"}
		vbolt.Write(tx, CameraConfigBkt, cameraConfig.RoomId, &cameraConfig)
		vbolt.TxCommit(tx)
	})

	// Simulate camera running
	cameraManager.Start(nil, 1, "rtsp://127.0.0.1:65535/test", "rtmp://127.0.0.1:65535/test")
	defer cameraManager.Stop(1)

	// Test 1: During class (should not stop)
	testTime := startTime.Add(30 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStopCamera(tx, &schedule, testTime) {
			t.Error("Should not stop camera during class time")
		}
	})

	// Test 2: During post-roll (should not stop)
	testTime = endTime.Add(1 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStopCamera(tx, &schedule, testTime) {
			t.Error("Should not stop camera during post-roll")
		}
	})

	// Test 3: After post-roll (should stop)
	testTime = endTime.Add(5 * time.Minute)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if !shouldStopCamera(tx, &schedule, testTime) {
			t.Error("Should stop camera after post-roll")
		}
	})

	// Test 4: Camera not running
	cameraManager.Stop(1)
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStopCamera(tx, &schedule, testTime) {
			t.Error("Should not try to stop camera that's not running")
		}
	})

	// Test 5: Auto-stop disabled
	cameraManager.Start(nil, 1, "rtsp://127.0.0.1:65535/test", "rtmp://127.0.0.1:65535/test")
	schedule.AutoStopCamera = false
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		if shouldStopCamera(tx, &schedule, testTime) {
			t.Error("Should not stop camera when AutoStopCamera is false")
		}
	})
}

// TestGetScheduleTimeWindow_OneTime tests time window calculation for one-time schedules
func TestGetScheduleTimeWindow_OneTime(t *testing.T) {
	now := time.Now()
	startTime := now.Add(1 * time.Hour)
	endTime := startTime.Add(1 * time.Hour)

	schedule := ClassSchedule{
		IsRecurring:     false,
		StartTime:       startTime,
		EndTime:         endTime,
		PreRollMinutes:  5,
		PostRollMinutes: 2,
	}

	start, end := getScheduleTimeWindow(&schedule, now)

	expectedStart := startTime.Add(-5 * time.Minute)
	expectedEnd := endTime.Add(2 * time.Minute)

	if !start.Equal(expectedStart) {
		t.Errorf("Start time mismatch. Got %v, want %v", start, expectedStart)
	}

	if !end.Equal(expectedEnd) {
		t.Errorf("End time mismatch. Got %v, want %v", end, expectedEnd)
	}
}

// TestGetScheduleTimeWindow_Recurring tests time window calculation for recurring schedules
func TestGetScheduleTimeWindow_Recurring(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().In(loc)
	currentWeekday := int(now.Weekday())

	schedule := ClassSchedule{
		IsRecurring:     true,
		RecurWeekdays:   []int{currentWeekday},
		RecurTimeStart:  "14:00",
		RecurTimeEnd:    "15:00",
		RecurTimezone:   "America/New_York",
		PreRollMinutes:  5,
		PostRollMinutes: 2,
	}

	start, end := getScheduleTimeWindow(&schedule, now)

	if start.IsZero() || end.IsZero() {
		t.Error("Should return valid time window for today")
	}

	// Verify times are in the schedule timezone
	if start.Location().String() != loc.String() {
		t.Errorf("Start time should be in schedule timezone. Got %s, want %s", start.Location().String(), loc.String())
	}

	// Verify pre-roll and post-roll applied
	duration := end.Sub(start)
	expectedDuration := 1*time.Hour + 5*time.Minute + 2*time.Minute // class + pre + post
	if duration != expectedDuration {
		t.Errorf("Duration mismatch. Got %v, want %v", duration, expectedDuration)
	}

	// Test wrong weekday
	wrongWeekday := (currentWeekday + 1) % 7
	schedule.RecurWeekdays = []int{wrongWeekday}
	start, end = getScheduleTimeWindow(&schedule, now)
	if !start.IsZero() || !end.IsZero() {
		t.Error("Should return zero times for wrong weekday")
	}
}

// TestSchedulerIgnoresDeleted tests that inactive schedules are skipped
func TestSchedulerIgnoresDeleted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var scheduleId int

	// Create a deleted schedule
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          1,
			StudioId:        1,
			Name:            "Deleted Class",
			IsRecurring:     false,
			StartTime:       time.Now(),
			EndTime:         time.Now().Add(1 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			IsActive:        false, // Deleted
		}
		scheduleId = schedule.Id
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.TxCommit(tx)
	})

	// Set up camera config
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := Room{Id: 1, StudioId: 1, StreamKey: "test-key"}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		cameraConfig := CameraConfig{RoomId: 1, RTSPURL: "rtsp://127.0.0.1:65535/test"}
		vbolt.Write(tx, CameraConfigBkt, cameraConfig.RoomId, &cameraConfig)
		vbolt.TxCommit(tx)
	})

	// Run scheduler
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		evaluateSchedules(db, tx)
	})

	// Verify camera was NOT started
	if cameraManager.IsRunning(1) {
		t.Error("Scheduler should ignore inactive schedules")
	}

	// Verify no execution logs were created
	var logs []int
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.ReadTermTargets(tx, LogsByScheduleIdx, scheduleId, &logs, vbolt.Window{})
	})

	if len(logs) > 0 {
		t.Error("Should not create execution logs for inactive schedules")
	}
}

// TestSchedulerLogsActions tests that scheduler actions are logged
func TestSchedulerLogsActions(t *testing.T) {
	clearRecentActionsCache() // Clear cache to avoid test pollution
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	var scheduleId int

	// Create an active schedule that should start camera now
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          1,
			StudioId:        1,
			Name:            "Test Class",
			IsRecurring:     false,
			StartTime:       now.Add(-10 * time.Minute), // Started 10 min ago
			EndTime:         now.Add(50 * time.Minute),  // Ends in 50 min
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			IsActive:        true,
		}
		scheduleId = schedule.Id
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.TxCommit(tx)
	})

	// Set up room and camera config
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := Room{Id: 1, StudioId: 1, StreamKey: "test-key"}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		cameraConfig := CameraConfig{RoomId: 1, RTSPURL: "rtsp://127.0.0.1:65535/test"}
		vbolt.Write(tx, CameraConfigBkt, cameraConfig.RoomId, &cameraConfig)
		vbolt.TxCommit(tx)
	})

	// Run scheduler
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		evaluateSchedules(db, tx)
	})

	// Camera should be started
	if !cameraManager.IsRunning(1) {
		t.Error("Scheduler should have started camera")
	}
	defer cameraManager.Stop(1)

	// Wait for async write to complete
	time.Sleep(100 * time.Millisecond)

	// Verify execution log was created
	var logIds []int
	var executionLog ScheduleExecutionLog

	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.ReadTermTargets(tx, LogsByScheduleIdx, scheduleId, &logIds, vbolt.Window{})

		if len(logIds) == 0 {
			t.Error("Expected execution log to be created")
			return
		}

		vbolt.Read(tx, ScheduleLogsBkt, logIds[0], &executionLog)
	})

	if executionLog.ScheduleId != scheduleId {
		t.Errorf("Log schedule ID mismatch. Got %d, want %d", executionLog.ScheduleId, scheduleId)
	}

	if executionLog.RoomId != 1 {
		t.Errorf("Log room ID mismatch. Got %d, want 1", executionLog.RoomId)
	}

	if executionLog.Action != "start_camera" {
		t.Errorf("Log action mismatch. Got %s, want start_camera", executionLog.Action)
	}

	if !executionLog.Success {
		t.Errorf("Log should indicate success, got error: %s", executionLog.ErrorMsg)
	}
}

// TestSchedulerHandlesNoCamera tests graceful handling when camera is not configured
func TestSchedulerHandlesNoCamera(t *testing.T) {
	clearRecentActionsCache() // Clear cache to avoid test pollution
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	var scheduleId int

	// Create a schedule for a room without camera config
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          99, // Room with no camera
			StudioId:        1,
			Name:            "Test Class",
			IsRecurring:     false,
			StartTime:       now,
			EndTime:         now.Add(1 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			IsActive:        true,
		}
		scheduleId = schedule.Id
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.TxCommit(tx)
	})

	// Set up room but NO camera config
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := Room{Id: 99, StudioId: 1, StreamKey: "test-key"}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.TxCommit(tx)
	})

	// Run scheduler - should not crash
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		evaluateSchedules(db, tx)
	})

	// Wait for async write to complete
	time.Sleep(100 * time.Millisecond)

	// Verify execution log shows failure
	var logIds []int
	var executionLog ScheduleExecutionLog

	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.ReadTermTargets(tx, LogsByScheduleIdx, scheduleId, &logIds, vbolt.Window{})

		if len(logIds) == 0 {
			t.Error("Expected execution log even for failed attempt")
			return
		}

		vbolt.Read(tx, ScheduleLogsBkt, logIds[0], &executionLog)
	})

	if executionLog.Success {
		t.Error("Log should indicate failure when camera not configured")
	}

	if executionLog.ErrorMsg != "Camera not configured" {
		t.Errorf("Log error message mismatch. Got %s, want 'Camera not configured'", executionLog.ErrorMsg)
	}
}

// TestSchedulerAvoidsDuplicates tests that camera is not started if already running
func TestSchedulerAvoidsDuplicates(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()

	// Create a schedule
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          1,
			StudioId:        1,
			Name:            "Test Class",
			IsRecurring:     false,
			StartTime:       now,
			EndTime:         now.Add(1 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.TxCommit(tx)
	})

	// Set up room and camera
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := Room{Id: 1, StudioId: 1, StreamKey: "test-key"}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		cameraConfig := CameraConfig{RoomId: 1, RTSPURL: "rtsp://127.0.0.1:65535/test"}
		vbolt.Write(tx, CameraConfigBkt, cameraConfig.RoomId, &cameraConfig)
		vbolt.TxCommit(tx)
	})

	// Manually start camera first
	cameraManager.Start(nil, 1, "rtsp://127.0.0.1:65535/test", "rtmp://127.0.0.1:65535/test")
	defer cameraManager.Stop(1)

	// Run scheduler
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		evaluateSchedules(db, tx)
	})

	// Check execution log - should have "skip_already_running" action
	var executionLog ScheduleExecutionLog
	logFound := false

	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.IterateAll(tx, ScheduleLogsBkt, func(logId int, log ScheduleExecutionLog) bool {
			executionLog = log
			logFound = true
			return false // stop after first log
		})
		vbolt.TxCommit(tx)
	})

	if logFound {
		if executionLog.Action != "skip_already_running" {
			t.Errorf("Expected skip_already_running action, got %s", executionLog.Action)
		}
	}
}

// TestRecentActionsCachePreventsRapidDuplicates tests that the recentActions cache
// prevents the same action from being executed multiple times in rapid succession
func TestRecentActionsCachePreventsRapidDuplicates(t *testing.T) {
	clearRecentActionsCache() // Clear cache to avoid test pollution
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	var scheduleId int

	// Create a schedule that should trigger camera start
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		schedule := ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          1,
			StudioId:        1,
			Name:            "Test Class",
			IsRecurring:     false,
			StartTime:       now.Add(-10 * time.Minute), // Started 10 min ago
			EndTime:         now.Add(50 * time.Minute),  // Ends in 50 min
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			IsActive:        true,
		}
		scheduleId = schedule.Id
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.TxCommit(tx)
	})

	// Set up room and camera config
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := Room{Id: 1, StudioId: 1, StreamKey: "test-key"}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		cameraConfig := CameraConfig{RoomId: 1, RTSPURL: "rtsp://127.0.0.1:65535/test"}
		vbolt.Write(tx, CameraConfigBkt, cameraConfig.RoomId, &cameraConfig)
		vbolt.TxCommit(tx)
	})

	// Run scheduler FIRST time
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		evaluateSchedules(db, tx)
	})

	// Stop the camera that was started
	if cameraManager.IsRunning(1) {
		cameraManager.Stop(1)
	}

	// Wait a tiny bit to ensure logs are written
	time.Sleep(100 * time.Millisecond)

	// Count logs after first run
	var firstRunLogCount int
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		var logIds []int
		vbolt.ReadTermTargets(tx, LogsByScheduleIdx, scheduleId, &logIds, vbolt.Window{})
		firstRunLogCount = len(logIds)
	})

	if firstRunLogCount != 1 {
		t.Errorf("Expected 1 log after first run, got %d", firstRunLogCount)
	}

	// Run scheduler SECOND time (within DUPLICATE_ACTION_WINDOW)
	// This should be blocked by the recentActions cache
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		evaluateSchedules(db, tx)
	})

	// Wait for any potential writes
	time.Sleep(100 * time.Millisecond)

	// Count logs after second run - should still be 1 (not 2!)
	var secondRunLogCount int
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		var logIds []int
		vbolt.ReadTermTargets(tx, LogsByScheduleIdx, scheduleId, &logIds, vbolt.Window{})
		secondRunLogCount = len(logIds)
		vbolt.TxCommit(tx)
	})

	if secondRunLogCount != 1 {
		t.Errorf("Expected 1 log after second run (duplicate prevented), got %d", secondRunLogCount)
	}

	// Verify the cache actually has an entry
	key := fmt.Sprintf("%d:start_camera", scheduleId)
	recentActionsMu.RLock()
	_, exists := recentActions[key]
	recentActionsMu.RUnlock()

	if !exists {
		t.Error("Expected action to be in recent actions cache")
	}
}

// TestParseTimeOfDay tests the time parsing helper
func TestParseTimeOfDay(t *testing.T) {
	tests := []struct {
		input      string
		expectedOk bool
		expectedH  int
		expectedM  int
	}{
		{"09:00", true, 9, 0},
		{"14:30", true, 14, 30},
		{"00:00", true, 0, 0},
		{"23:59", true, 23, 59},
		{"9:00", true, 9, 0},      // Single digit hour is valid
		{"25:00", false, 0, 0},    // Invalid hour
		{"12:60", false, 0, 0},    // Invalid minute
		{"12:00:00", false, 0, 0}, // Wrong format
	}

	for _, test := range tests {
		result, err := parseTimeOfDay(test.input)
		if test.expectedOk && err != nil {
			t.Errorf("parseTimeOfDay(%s) failed: %v", test.input, err)
		}
		if !test.expectedOk && err == nil {
			t.Errorf("parseTimeOfDay(%s) should have failed", test.input)
		}
		if test.expectedOk {
			if result.Hour() != test.expectedH || result.Minute() != test.expectedM {
				t.Errorf("parseTimeOfDay(%s) = %02d:%02d, want %02d:%02d",
					test.input, result.Hour(), result.Minute(), test.expectedH, test.expectedM)
			}
		}
	}
}

// TestOverlappingSchedules tests that consecutive classes with pre/post-roll overlap
// don't prematurely stop the camera when the second class is using it
func TestOverlappingSchedules(t *testing.T) {
	clearRecentActionsCache()
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	roomId := 1
	studioId := 1

	// Create camera config and room
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		cameraConfig := CameraConfig{
			RoomId:  roomId,
			RTSPURL: "rtsp://127.0.0.1:65535/test",
		}
		vbolt.Write(tx, CameraConfigBkt, roomId, &cameraConfig)

		room := Room{
			Id:        roomId,
			StreamKey: "test-key",
		}
		vbolt.Write(tx, RoomsBkt, roomId, &room)
		vbolt.TxCommit(tx)
	})

	// Current time: 10:00 AM
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	// Schedule 1: 9:00-10:00 with 15 min post-roll
	// Time window: 9:00-10:15
	// At 10:00, this schedule has passed its end time but is still in post-roll
	schedule1Id := 1
	schedule1 := ClassSchedule{
		Id:              schedule1Id,
		StudioId:        studioId,
		RoomId:          roomId,
		Name:            "Morning Class",
		IsActive:        true,
		AutoStartCamera: true,
		AutoStopCamera:  true,
		IsRecurring:     false,
		StartTime:       time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		PreRollMinutes:  0,
		PostRollMinutes: 15,
	}

	// Schedule 2: 10:00-11:00 with 15 min pre-roll
	// Time window: 9:45-11:00
	// At 10:00, this schedule is in its pre-roll phase
	schedule2Id := 2
	schedule2 := ClassSchedule{
		Id:              schedule2Id,
		StudioId:        studioId,
		RoomId:          roomId,
		Name:            "Late Morning Class",
		IsActive:        true,
		AutoStartCamera: true,
		AutoStopCamera:  true,
		IsRecurring:     false,
		StartTime:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		PreRollMinutes:  15,
		PostRollMinutes: 0,
	}

	// Save schedules
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		vbolt.Write(tx, ClassSchedulesBkt, schedule1Id, &schedule1)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule1Id, roomId)

		vbolt.Write(tx, ClassSchedulesBkt, schedule2Id, &schedule2)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule2Id, roomId)

		vbolt.TxCommit(tx)
	})

	// Simulate that camera was started earlier by schedule2 (at 9:45)
	cameraManager.Start(nil, roomId, "rtsp://127.0.0.1:65535/test", "rtmp://127.0.0.1:1935/live/test-key")

	// Mark schedule1's start action as already executed to prevent re-starting
	recentActionsMu.Lock()
	recentActions[fmt.Sprintf("%d:start_camera", schedule1Id)] = now.Add(-30 * time.Minute)
	recentActionsMu.Unlock()

	// Evaluate schedules at 10:00
	// Schedule 1 should want to stop (past its end time)
	// BUT Schedule 2 is still active (in pre-roll window 9:45-11:00)
	// So camera should NOT be stopped
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		// Check schedule1's shouldStopCamera
		shouldStop1 := shouldStopCamera(tx, &schedule1, now)
		if shouldStop1 {
			t.Error("Schedule 1 should not want to stop camera because Schedule 2 is overlapping")
		}

		// Verify hasOverlappingActiveSchedule returns true for schedule1
		hasOverlap := hasOverlappingActiveSchedule(tx, roomId, schedule1Id, now)
		if !hasOverlap {
			t.Error("Expected hasOverlappingActiveSchedule to return true for schedule1 at 10:00")
		}
	})

	// Now test at 10:16 - after schedule1's window but still in schedule2's window
	laterTime := time.Date(2024, 1, 15, 10, 16, 0, 0, time.UTC)

	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		// Schedule1 should still not stop because schedule2 is active
		shouldStop1 := shouldStopCamera(tx, &schedule1, laterTime)
		if shouldStop1 {
			t.Error("Schedule 1 should not want to stop camera at 10:16 because Schedule 2 is still active")
		}
	})

	// Test at 11:01 - after both schedules' windows
	muchLaterTime := time.Date(2024, 1, 15, 11, 1, 0, 0, time.UTC)

	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		// Schedule2 should want to stop now
		shouldStop2 := shouldStopCamera(tx, &schedule2, muchLaterTime)
		if !shouldStop2 {
			t.Error("Schedule 2 should want to stop camera at 11:01 (past its window)")
		}

		// No overlap at this time
		hasOverlap := hasOverlappingActiveSchedule(tx, roomId, schedule2Id, muchLaterTime)
		if hasOverlap {
			t.Error("Expected no overlapping schedule at 11:01")
		}
	})

	// Clean up camera
	cameraManager.Stop(roomId)
}
