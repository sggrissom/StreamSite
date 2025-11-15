package backend

import (
	"fmt"
	"stream/cfg"
	"sync"
	"time"

	"go.hasen.dev/vbolt"
)

// Constants for scheduler configuration
const (
	SCHEDULER_CHECK_INTERVAL = 30 * time.Second
	DUPLICATE_ACTION_WINDOW  = 1 * time.Minute // Prevent duplicate actions within this window
)

// In-memory cache for tracking recent actions to prevent duplicates
var (
	recentActionsMu sync.RWMutex
	recentActions   = make(map[string]time.Time) // "scheduleId:action" -> timestamp
)

// StartClassScheduler starts the background scheduler that evaluates class schedules
// and automatically starts/stops cameras based on schedule timing
func StartClassScheduler(db *vbolt.DB) {
	LogInfo(LogCategorySystem, "Starting class schedule background job", map[string]interface{}{
		"checkInterval": SCHEDULER_CHECK_INTERVAL.String(),
	})

	go func() {
		ticker := time.NewTicker(SCHEDULER_CHECK_INTERVAL)
		defer ticker.Stop()

		for range ticker.C {
			db.View(func(tx *vbolt.Tx) error {
				evaluateSchedules(db, tx)
				return nil
			})
		}
	}()

	// Start cleanup goroutine for recent actions cache
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			cleanupRecentActions()
		}
	}()
}

// evaluateSchedules checks all active schedules and performs camera actions as needed
func evaluateSchedules(db *vbolt.DB, tx *vbolt.Tx) {
	now := time.Now()

	scheduleCount := 0

	// Iterate through all schedules
	vbolt.IterateAll(tx, ClassSchedulesBkt, func(scheduleId int, schedule ClassSchedule) bool {
		scheduleCount++

		// Skip inactive schedules
		if !schedule.IsActive {
			return true // continue iteration
		}

		// Check if camera should start
		if shouldStartCamera(&schedule, now) {
			executeScheduleAction(db, tx, &schedule, "start_camera", now)
		}

		// Check if camera should stop
		if shouldStopCamera(&schedule, now) {
			executeScheduleAction(db, tx, &schedule, "stop_camera", now)
		}

		return true // continue iteration
	})

	LogDebug(LogCategorySystem, "Evaluated class schedules", map[string]interface{}{
		"scheduleCount": scheduleCount,
		"currentTime":   now.Format(time.RFC3339),
	})
}

// shouldStartCamera determines if a camera should be started for the given schedule
func shouldStartCamera(schedule *ClassSchedule, now time.Time) bool {
	// Must have auto-start enabled
	if !schedule.AutoStartCamera {
		return false
	}

	// Check if camera is already running
	if cameraManager.IsRunning(schedule.RoomId) {
		return false
	}

	// Check if we recently tried to start (avoid duplicates)
	if recentlyExecutedAction(schedule.Id, "start_camera", now) {
		return false
	}

	// Calculate the time window when camera should be running
	startWindow, endWindow := getScheduleTimeWindow(schedule, now)
	if startWindow.IsZero() {
		return false // No valid time window for today
	}

	// Camera should start if we're within the window
	return now.After(startWindow) && now.Before(endWindow)
}

// shouldStopCamera determines if a camera should be stopped for the given schedule
func shouldStopCamera(schedule *ClassSchedule, now time.Time) bool {
	// Must have auto-stop enabled
	if !schedule.AutoStopCamera {
		return false
	}

	// Camera must be running
	if !cameraManager.IsRunning(schedule.RoomId) {
		return false
	}

	// Check if we recently tried to stop (avoid duplicates)
	if recentlyExecutedAction(schedule.Id, "stop_camera", now) {
		return false
	}

	// Calculate the time window when camera should be running
	startWindow, endWindow := getScheduleTimeWindow(schedule, now)
	if startWindow.IsZero() {
		return false // No valid time window for today
	}

	// Camera should stop if we're outside the window
	// Only stop if we're past the end window
	if now.After(endWindow) {
		// Check if another schedule is keeping this camera active
		if hasOverlappingActiveSchedule(schedule.RoomId, schedule.Id, now) {
			return false
		}
		return true
	}

	return false
}

// getScheduleTimeWindow returns the start and end time (including pre/post roll) for a schedule
// Returns zero times if the schedule is not active today
func getScheduleTimeWindow(schedule *ClassSchedule, now time.Time) (time.Time, time.Time) {
	if !schedule.IsRecurring {
		// One-time schedule
		startWithPreRoll := schedule.StartTime.Add(-time.Duration(schedule.PreRollMinutes) * time.Minute)
		endWithPostRoll := schedule.EndTime.Add(time.Duration(schedule.PostRollMinutes) * time.Minute)

		// Check if the schedule is in the past or future (not relevant today)
		if now.Before(startWithPreRoll.Add(-24*time.Hour)) || now.After(endWithPostRoll.Add(24*time.Hour)) {
			return time.Time{}, time.Time{}
		}

		return startWithPreRoll, endWithPostRoll
	}

	// Recurring schedule - check if today matches
	loc, err := time.LoadLocation(schedule.RecurTimezone)
	if err != nil {
		LogWarn(LogCategorySystem, "Invalid timezone in schedule", map[string]interface{}{
			"scheduleId": schedule.Id,
			"timezone":   schedule.RecurTimezone,
			"error":      err.Error(),
		})
		return time.Time{}, time.Time{}
	}

	nowInScheduleTz := now.In(loc)
	weekday := int(nowInScheduleTz.Weekday())

	// Check if today is a scheduled weekday
	isScheduledDay := false
	for _, day := range schedule.RecurWeekdays {
		if day == weekday {
			isScheduledDay = true
			break
		}
	}

	if !isScheduledDay {
		return time.Time{}, time.Time{}
	}

	// Check if we're within the recurring date range
	if !schedule.RecurStartDate.IsZero() && nowInScheduleTz.Before(schedule.RecurStartDate) {
		return time.Time{}, time.Time{}
	}
	if !schedule.RecurEndDate.IsZero() && nowInScheduleTz.After(schedule.RecurEndDate) {
		return time.Time{}, time.Time{}
	}

	// Parse the recurring time strings
	startTime, err := parseTimeOfDay(schedule.RecurTimeStart)
	if err != nil {
		LogWarn(LogCategorySystem, "Invalid start time in schedule", map[string]interface{}{
			"scheduleId": schedule.Id,
			"timeString": schedule.RecurTimeStart,
			"error":      err.Error(),
		})
		return time.Time{}, time.Time{}
	}

	endTime, err := parseTimeOfDay(schedule.RecurTimeEnd)
	if err != nil {
		LogWarn(LogCategorySystem, "Invalid end time in schedule", map[string]interface{}{
			"scheduleId": schedule.Id,
			"timeString": schedule.RecurTimeEnd,
			"error":      err.Error(),
		})
		return time.Time{}, time.Time{}
	}

	// Build today's start and end times in the schedule's timezone
	year, month, day := nowInScheduleTz.Date()
	todayStart := time.Date(year, month, day, startTime.Hour(), startTime.Minute(), 0, 0, loc)
	todayEnd := time.Date(year, month, day, endTime.Hour(), endTime.Minute(), 0, 0, loc)

	// Apply pre-roll and post-roll
	startWithPreRoll := todayStart.Add(-time.Duration(schedule.PreRollMinutes) * time.Minute)
	endWithPostRoll := todayEnd.Add(time.Duration(schedule.PostRollMinutes) * time.Minute)

	return startWithPreRoll, endWithPostRoll
}

// parseTimeOfDay parses a time string like "09:00" or "14:30"
func parseTimeOfDay(timeStr string) (time.Time, error) {
	return time.Parse("15:04", timeStr)
}

// recentlyExecutedAction checks if an action was executed recently to prevent duplicates
func recentlyExecutedAction(scheduleId int, action string, now time.Time) bool {
	key := fmt.Sprintf("%d:%s", scheduleId, action)

	recentActionsMu.RLock()
	lastTime, exists := recentActions[key]
	recentActionsMu.RUnlock()

	if !exists {
		return false
	}

	// Check if within duplicate window
	return now.Sub(lastTime) < DUPLICATE_ACTION_WINDOW
}

// hasOverlappingActiveSchedule checks if another schedule is keeping the camera active
func hasOverlappingActiveSchedule(roomId, excludeScheduleId int, now time.Time) bool {
	// This would query all schedules for the room and check if any overlap with current time
	// For the initial implementation, we'll return false (conservative approach)
	// TODO: Implement overlap detection in future iteration
	return false
}

// executeScheduleAction performs the actual camera start/stop action and logs it
func executeScheduleAction(db *vbolt.DB, tx *vbolt.Tx, schedule *ClassSchedule, action string, now time.Time) {
	// Get camera configuration
	var cameraConfig CameraConfig
	vbolt.Read(tx, CameraConfigBkt, schedule.RoomId, &cameraConfig)
	if cameraConfig.RoomId == 0 {
		// Camera not configured - log and skip
		vbolt.WithWriteTx(db, func(wtx *vbolt.Tx) {
			logExecutionResult(wtx, schedule, action, now, false, "Camera not configured")
			vbolt.TxCommit(wtx)
		})
		return
	}

	// Get room stream key for RTMP output
	var room Room
	vbolt.Read(tx, RoomsBkt, schedule.RoomId, &room)
	if room.Id == 0 {
		vbolt.WithWriteTx(db, func(wtx *vbolt.Tx) {
			logExecutionResult(wtx, schedule, action, now, false, "Room not found")
			vbolt.TxCommit(wtx)
		})
		return
	}

	var success bool
	var errorMsg string

	if action == "start_camera" {
		// Check one more time if camera is already running
		if cameraManager.IsRunning(schedule.RoomId) {
			vbolt.WithWriteTx(db, func(wtx *vbolt.Tx) {
				logExecutionResult(wtx, schedule, "skip_already_running", now, true, "")
				vbolt.TxCommit(wtx)
			})
			return
		}

		// Build RTMP output URL
		rtmpOut := fmt.Sprintf("%s/%s", cfg.SRSRTMPBase, room.StreamKey)

		// Start the camera
		err := cameraManager.Start(nil, schedule.RoomId, cameraConfig.RTSPURL, rtmpOut)
		if err != nil {
			success = false
			errorMsg = err.Error()
		} else {
			success = true
			LogInfo(LogCategorySystem, "Class scheduler started camera", map[string]interface{}{
				"scheduleId":   schedule.Id,
				"scheduleName": schedule.Name,
				"roomId":       schedule.RoomId,
				"studioId":     schedule.StudioId,
			})
		}
	} else if action == "stop_camera" {
		// Stop the camera
		err := cameraManager.Stop(schedule.RoomId)
		if err != nil {
			success = false
			errorMsg = err.Error()
		} else {
			success = true
			LogInfo(LogCategorySystem, "Class scheduler stopped camera", map[string]interface{}{
				"scheduleId":   schedule.Id,
				"scheduleName": schedule.Name,
				"roomId":       schedule.RoomId,
				"studioId":     schedule.StudioId,
			})
		}
	}

	// Record action to prevent duplicates
	key := fmt.Sprintf("%d:%s", schedule.Id, action)
	recentActionsMu.Lock()
	recentActions[key] = now
	recentActionsMu.Unlock()

	// Log the execution - open write transaction
	vbolt.WithWriteTx(db, func(wtx *vbolt.Tx) {
		logExecutionResult(wtx, schedule, action, now, success, errorMsg)
		vbolt.TxCommit(wtx)
	})
}

// logExecutionResult creates a log entry in the ScheduleExecutionLog
func logExecutionResult(tx *vbolt.Tx, schedule *ClassSchedule, action string, timestamp time.Time, success bool, errorMsg string) {
	log := ScheduleExecutionLog{
		Id:         vbolt.NextIntId(tx, ScheduleLogsBkt),
		ScheduleId: schedule.Id,
		RoomId:     schedule.RoomId,
		Action:     action,
		Timestamp:  timestamp,
		Success:    success,
		ErrorMsg:   errorMsg,
	}

	vbolt.Write(tx, ScheduleLogsBkt, log.Id, &log)

	// Update indexes
	vbolt.SetTargetSingleTerm(tx, LogsByScheduleIdx, log.Id, log.ScheduleId)
	vbolt.SetTargetSingleTerm(tx, LogsByRoomIdx, log.Id, log.RoomId)
}

// cleanupRecentActions removes old entries from the recent actions cache
func cleanupRecentActions() {
	now := time.Now()
	cutoff := now.Add(-2 * DUPLICATE_ACTION_WINDOW) // Keep 2x the window for safety

	recentActionsMu.Lock()
	defer recentActionsMu.Unlock()

	for key, timestamp := range recentActions {
		if timestamp.Before(cutoff) {
			delete(recentActions, key)
		}
	}
}
