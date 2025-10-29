package backend

import (
	"errors"
	"strconv"
	"stream/cfg"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// Session timeout constants for smart view counting
const (
	SESSION_TIMEOUT          = 5 * time.Minute  // Window for same session (reconnects don't count as new views)
	SESSION_CLEANUP_INTERVAL = 5 * time.Minute  // How often to run cleanup job
	SESSION_CLEANUP_AGE      = 10 * time.Minute // Age at which to delete sessions
)

// RoomAnalytics tracks viewing statistics and streaming metrics for a room
type RoomAnalytics struct {
	RoomId                 int       `json:"roomId"`
	TotalViewsAllTime      int       `json:"totalViewsAllTime"`      // Total connection count
	TotalViewsThisMonth    int       `json:"totalViewsThisMonth"`    // Reset monthly
	UniqueViewersAllTime   int       `json:"uniqueViewersAllTime"`   // Unique viewers (lifetime)
	UniqueViewersThisMonth int       `json:"uniqueViewersThisMonth"` // Unique viewers (reset monthly)
	CurrentViewers         int       `json:"currentViewers"`         // Live count
	PeakViewers            int       `json:"peakViewers"`            // Historical max
	PeakViewersAt          time.Time `json:"peakViewersAt"`
	LastStreamAt           time.Time `json:"lastStreamAt"`       // Last IsActive=true
	StreamStartedAt        time.Time `json:"streamStartedAt"`    // When current stream began
	TotalStreamMinutes     int       `json:"totalStreamMinutes"` // Lifetime streaming duration
}

// StudioAnalytics tracks aggregated statistics across all rooms in a studio
type StudioAnalytics struct {
	StudioId               int `json:"studioId"`
	TotalViewsAllTime      int `json:"totalViewsAllTime"`
	TotalViewsThisMonth    int `json:"totalViewsThisMonth"`
	UniqueViewersAllTime   int `json:"uniqueViewersAllTime"`   // Unique viewers (lifetime)
	UniqueViewersThisMonth int `json:"uniqueViewersThisMonth"` // Unique viewers (reset monthly)
	CurrentViewers         int `json:"currentViewers"`         // Sum across all rooms
	TotalRooms             int `json:"totalRooms"`
	ActiveRooms            int `json:"activeRooms"` // Currently streaming
	TotalStreamMinutes     int `json:"totalStreamMinutes"`
}

// ViewerSession tracks individual viewer sessions for smart view counting
type ViewerSession struct {
	SessionKey  string    `json:"sessionKey"` // Composite key: "viewerId:roomId"
	ViewerId    string    `json:"viewerId"`   // userId (JWT) or sessionToken (code auth)
	RoomId      int       `json:"roomId"`
	Code        string    `json:"code"`        // Access code used (empty for regular users)
	FirstSeenAt time.Time `json:"firstSeenAt"` // When this viewer first connected to this room
	LastSeenAt  time.Time `json:"lastSeenAt"`  // Most recent activity (updated on reconnect)
}

// Packing functions for vbolt serialization

func PackRoomAnalytics(self *RoomAnalytics, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.Int(&self.TotalViewsAllTime, buf)
	vpack.Int(&self.TotalViewsThisMonth, buf)
	vpack.Int(&self.UniqueViewersAllTime, buf)
	vpack.Int(&self.UniqueViewersThisMonth, buf)
	vpack.Int(&self.CurrentViewers, buf)
	vpack.Int(&self.PeakViewers, buf)
	vpack.Time(&self.PeakViewersAt, buf)
	vpack.Time(&self.LastStreamAt, buf)
	vpack.Time(&self.StreamStartedAt, buf)
	vpack.Int(&self.TotalStreamMinutes, buf)
}

func PackStudioAnalytics(self *StudioAnalytics, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.StudioId, buf)
	vpack.Int(&self.TotalViewsAllTime, buf)
	vpack.Int(&self.TotalViewsThisMonth, buf)
	vpack.Int(&self.UniqueViewersAllTime, buf)
	vpack.Int(&self.UniqueViewersThisMonth, buf)
	vpack.Int(&self.CurrentViewers, buf)
	vpack.Int(&self.TotalRooms, buf)
	vpack.Int(&self.ActiveRooms, buf)
	vpack.Int(&self.TotalStreamMinutes, buf)
}

func PackViewerSession(self *ViewerSession, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.String(&self.SessionKey, buf)
	vpack.String(&self.ViewerId, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.String(&self.Code, buf)
	vpack.Time(&self.FirstSeenAt, buf)
	vpack.Time(&self.LastSeenAt, buf)
}

// Buckets for entity storage

// RoomAnalyticsBkt: roomId (int) -> RoomAnalytics
var RoomAnalyticsBkt = vbolt.Bucket(&cfg.Info, "room_analytics", vpack.FInt, PackRoomAnalytics)

// StudioAnalyticsBkt: studioId (int) -> StudioAnalytics
var StudioAnalyticsBkt = vbolt.Bucket(&cfg.Info, "studio_analytics", vpack.FInt, PackStudioAnalytics)

// ViewerSessionsBkt: sessionKey (string) -> ViewerSession
var ViewerSessionsBkt = vbolt.Bucket(&cfg.Info, "viewer_sessions", vpack.StringZ, PackViewerSession)

// Indexes for querying

// SessionsByRoomIndex: Term=roomId, Target=sessionKey
var SessionsByRoomIndex = vbolt.Index(&cfg.Info, "sessions_by_room", vpack.FInt, vpack.StringZ)

// SessionsByCodeIndex: Term=code, Target=sessionKey
// Find all viewer sessions using a specific access code
var SessionsByCodeIndex = vbolt.Index(&cfg.Info, "sessions_by_code", vpack.StringZ, vpack.StringZ)

// Helper functions for analytics tracking

func IncrementRoomViewerCount(db *vbolt.DB, roomId int, viewerId string, code string) {
	var currentCount int
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Get room to find studioId
		room := GetRoom(tx, roomId)
		if room.Id == 0 {
			return
		}

		// Build session key
		sessionKey := viewerId + ":" + strconv.Itoa(roomId)

		// Look up existing session
		var session ViewerSession
		vbolt.Read(tx, ViewerSessionsBkt, sessionKey, &session)

		now := time.Now()
		isNewSession := session.SessionKey == ""
		isExpiredSession := !isNewSession && now.Sub(session.LastSeenAt) > SESSION_TIMEOUT

		// Load or create room analytics
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, roomId, &analytics)
		if analytics.RoomId == 0 {
			analytics.RoomId = roomId
		}

		// Always increment current viewers (live count)
		analytics.CurrentViewers++

		// Only increment view totals if new or expired session
		if isNewSession || isExpiredSession {
			analytics.TotalViewsAllTime++
			analytics.TotalViewsThisMonth++

			// Track unique viewers
			if isNewSession {
				analytics.UniqueViewersAllTime++
				analytics.UniqueViewersThisMonth++
			}
		}

		// Update peak if necessary
		if analytics.CurrentViewers > analytics.PeakViewers {
			analytics.PeakViewers = analytics.CurrentViewers
			analytics.PeakViewersAt = now
		}

		// Update or create session
		if isNewSession {
			session.SessionKey = sessionKey
			session.ViewerId = viewerId
			session.RoomId = roomId
			session.Code = code
			session.FirstSeenAt = now
		}
		session.LastSeenAt = now

		// Save session and analytics
		vbolt.Write(tx, ViewerSessionsBkt, sessionKey, &session)
		vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, sessionKey, roomId)

		// Index by code if this is a code-based session
		if code != "" {
			vbolt.SetTargetSingleTerm(tx, SessionsByCodeIndex, sessionKey, code)

			// Increment code analytics
			var codeAnalytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &codeAnalytics)
			codeAnalytics.CurrentViewers++

			// Update peak if necessary
			if codeAnalytics.CurrentViewers > codeAnalytics.PeakViewers {
				codeAnalytics.PeakViewers = codeAnalytics.CurrentViewers
				codeAnalytics.PeakViewersAt = now
			}

			vbolt.Write(tx, CodeAnalyticsBkt, code, &codeAnalytics)
		}

		vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)

		incrementStudioViewerCount(tx, room.StudioId, viewerId, isNewSession, isExpiredSession)

		vbolt.TxCommit(tx)
		currentCount = analytics.CurrentViewers
	})

	sseManager.BroadcastViewerCount(roomId, currentCount)
}

func DecrementRoomViewerCount(db *vbolt.DB, roomId int, viewerId string, accessCode string) {
	var currentCount int
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := GetRoom(tx, roomId)
		if room.Id == 0 {
			return
		}

		// Determine the code to use for analytics
		// Prefer the passed accessCode, but fall back to session lookup for backwards compatibility
		codeToUse := accessCode

		// If no code was passed, try to look it up from the session
		if codeToUse == "" {
			sessionKey := viewerId + ":" + strconv.Itoa(roomId)
			var session ViewerSession
			vbolt.Read(tx, ViewerSessionsBkt, sessionKey, &session)
			codeToUse = session.Code
		}

		// If this is a code-based session, decrement code analytics
		if codeToUse != "" {
			var codeAnalytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, codeToUse, &codeAnalytics)

			if codeAnalytics.CurrentViewers > 0 {
				codeAnalytics.CurrentViewers--
			}

			vbolt.Write(tx, CodeAnalyticsBkt, codeToUse, &codeAnalytics)
		}

		// Load room analytics
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, roomId, &analytics)
		if analytics.RoomId == 0 {
			return
		}

		if analytics.CurrentViewers > 0 {
			analytics.CurrentViewers--
		}

		// Save
		vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)

		decrementStudioViewerCount(tx, room.StudioId)

		vbolt.TxCommit(tx)
		currentCount = analytics.CurrentViewers
	})

	// Broadcast updated viewer count to all connected clients
	sseManager.BroadcastViewerCount(roomId, currentCount)
}

func UpdateStudioAnalyticsFromRoom(tx *vbolt.Tx, studioId int) {
	// Get all rooms in studio
	rooms := ListStudioRooms(tx, studioId)

	// Aggregate stats
	var studioAnalytics StudioAnalytics
	studioAnalytics.StudioId = studioId
	studioAnalytics.TotalRooms = len(rooms)

	for _, room := range rooms {
		// Load room analytics
		var roomAnalytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, room.Id, &roomAnalytics)

		// Aggregate
		studioAnalytics.TotalViewsAllTime += roomAnalytics.TotalViewsAllTime
		studioAnalytics.TotalViewsThisMonth += roomAnalytics.TotalViewsThisMonth
		studioAnalytics.UniqueViewersAllTime += roomAnalytics.UniqueViewersAllTime
		studioAnalytics.UniqueViewersThisMonth += roomAnalytics.UniqueViewersThisMonth
		studioAnalytics.CurrentViewers += roomAnalytics.CurrentViewers
		studioAnalytics.TotalStreamMinutes += roomAnalytics.TotalStreamMinutes

		// Count active rooms
		if room.IsActive {
			studioAnalytics.ActiveRooms++
		}
	}

	// Save
	vbolt.Write(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)
}

func incrementStudioViewerCount(tx *vbolt.Tx, studioId int, viewerId string, isNewSession bool, isExpiredSession bool) {
	var studioAnalytics StudioAnalytics
	vbolt.Read(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)

	// Initialize if first viewer
	if studioAnalytics.StudioId == 0 {
		studioAnalytics.StudioId = studioId
	}

	// Always increment current viewers
	studioAnalytics.CurrentViewers++

	// Only increment totals if new or expired session
	if isNewSession || isExpiredSession {
		studioAnalytics.TotalViewsAllTime++
		studioAnalytics.TotalViewsThisMonth++

		// Track unique viewers at studio level
		if isNewSession {
			studioAnalytics.UniqueViewersAllTime++
			studioAnalytics.UniqueViewersThisMonth++
		}
	}

	vbolt.Write(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)
}

func decrementStudioViewerCount(tx *vbolt.Tx, studioId int) {
	var studioAnalytics StudioAnalytics
	vbolt.Read(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)

	if studioAnalytics.StudioId == 0 {
		return
	}

	if studioAnalytics.CurrentViewers > 0 {
		studioAnalytics.CurrentViewers--
	}

	vbolt.Write(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)
}

func RecordStreamStart(db *vbolt.DB, roomId int, studioId int) {
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, roomId, &analytics)
		if analytics.RoomId == 0 {
			analytics.RoomId = roomId
		}

		analytics.StreamStartedAt = time.Now()

		vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)

		UpdateStudioAnalyticsFromRoom(tx, studioId)

		vbolt.TxCommit(tx)
	})
}

func RecordStreamStop(db *vbolt.DB, roomId int, studioId int) {
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, roomId, &analytics)
		if analytics.RoomId == 0 {
			return
		}

		// Calculate duration if stream was started
		if !analytics.StreamStartedAt.IsZero() {
			duration := time.Since(analytics.StreamStartedAt)
			analytics.TotalStreamMinutes += int(duration.Minutes())
		}

		analytics.LastStreamAt = time.Now()
		analytics.StreamStartedAt = time.Time{} // Reset to zero

		vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)

		UpdateStudioAnalyticsFromRoom(tx, studioId)

		vbolt.TxCommit(tx)
	})
}

// API Procedures

type GetRoomAnalyticsRequest struct {
	RoomId int `json:"roomId"`
}

type GetRoomAnalyticsResponse struct {
	Analytics   *RoomAnalytics `json:"analytics,omitempty"`
	IsStreaming bool           `json:"isStreaming"`
	RoomName    string         `json:"roomName"`
}

func GetRoomAnalytics(ctx *vbeam.Context, req GetRoomAnalyticsRequest) (resp GetRoomAnalyticsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Validate room ID
	if req.RoomId <= 0 {
		return resp, errors.New("Invalid room ID")
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("Room not found")
	}

	// Check studio access (requires at least viewer role)
	access := CheckStudioAccess(ctx.Tx, caller, room.StudioId)
	if !access.Allowed {
		return resp, errors.New("Access denied")
	}

	// Load analytics
	var analytics RoomAnalytics
	vbolt.Read(ctx.Tx, RoomAnalyticsBkt, req.RoomId, &analytics)

	// Initialize if not found
	if analytics.RoomId == 0 {
		analytics.RoomId = req.RoomId
	}

	resp.Analytics = &analytics
	resp.IsStreaming = room.IsActive
	resp.RoomName = room.Name
	return
}

type GetStudioAnalyticsRequest struct {
	StudioId int `json:"studioId"`
}

type GetStudioAnalyticsResponse struct {
	Analytics  *StudioAnalytics `json:"analytics,omitempty"`
	StudioName string           `json:"studioName"`
}

func GetStudioAnalytics(ctx *vbeam.Context, req GetStudioAnalyticsRequest) (resp GetStudioAnalyticsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Validate studio ID
	if req.StudioId <= 0 {
		return resp, errors.New("Invalid studio ID")
	}

	// Check studio access (requires at least viewer role)
	access := CheckStudioAccess(ctx.Tx, caller, req.StudioId)
	if !access.Allowed {
		return resp, errors.New("Access denied")
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		return resp, errors.New("Studio not found")
	}

	// Load studio analytics
	var analytics StudioAnalytics
	vbolt.Read(ctx.Tx, StudioAnalyticsBkt, req.StudioId, &analytics)

	// Initialize if not found
	if analytics.StudioId == 0 {
		analytics.StudioId = req.StudioId
	}

	resp.Analytics = &analytics
	resp.StudioName = studio.Name
	return
}

// RegisterAnalyticsMethods registers all analytics-related API procedures
func RegisterAnalyticsMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, GetRoomAnalytics)
	vbeam.RegisterProc(app, GetStudioAnalytics)
}

// Background Jobs

var lastResetMonth = time.Now().Month()
var lastResetYear = time.Now().Year()

// ResetMonthlyAnalytics resets the TotalViewsThisMonth and UniqueViewersThisMonth counters for all rooms and studios
// Should be called on the first day of each month
func ResetMonthlyAnalytics(db *vbolt.DB) int {
	resetCount := 0

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Reset room analytics
		vbolt.IterateAll(tx, RoomAnalyticsBkt, func(roomId int, analytics RoomAnalytics) bool {
			if analytics.TotalViewsThisMonth > 0 || analytics.UniqueViewersThisMonth > 0 {
				analytics.TotalViewsThisMonth = 0
				analytics.UniqueViewersThisMonth = 0
				vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)
				resetCount++
			}
			return true // continue iteration
		})

		// Reset studio analytics
		vbolt.IterateAll(tx, StudioAnalyticsBkt, func(studioId int, analytics StudioAnalytics) bool {
			if analytics.TotalViewsThisMonth > 0 || analytics.UniqueViewersThisMonth > 0 {
				analytics.TotalViewsThisMonth = 0
				analytics.UniqueViewersThisMonth = 0
				vbolt.Write(tx, StudioAnalyticsBkt, studioId, &analytics)
				resetCount++
			}
			return true // continue iteration
		})

		vbolt.TxCommit(tx)
	})

	LogInfo(LogCategorySystem, "Monthly analytics reset completed", map[string]interface{}{
		"resetCount": resetCount,
	})

	return resetCount
}

// StartMonthlyAnalyticsReset starts a background goroutine that checks daily
// and resets monthly analytics counters on the first day of each month
func StartMonthlyAnalyticsReset(db *vbolt.DB) {
	LogInfo(LogCategorySystem, "Starting monthly analytics reset job", map[string]interface{}{
		"frequency": "daily check",
	})

	go func() {
		// Check immediately on startup
		now := time.Now()
		currentMonth := now.Month()
		currentYear := now.Year()

		// Reset if we're in a new month
		if currentMonth != lastResetMonth || currentYear != lastResetYear {
			ResetMonthlyAnalytics(db)
			lastResetMonth = currentMonth
			lastResetYear = currentYear
		}

		// Then check daily
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			now = time.Now()
			currentMonth = now.Month()
			currentYear = now.Year()

			// Reset on month change
			if currentMonth != lastResetMonth || currentYear != lastResetYear {
				ResetMonthlyAnalytics(db)
				lastResetMonth = currentMonth
				lastResetYear = currentYear
			}
		}
	}()
}

// StartViewerSessionCleanup starts a background goroutine that periodically
// removes stale viewer sessions to prevent database bloat
func StartViewerSessionCleanup(db *vbolt.DB) {
	LogInfo(LogCategorySystem, "Starting viewer session cleanup job", map[string]interface{}{
		"interval":   SESSION_CLEANUP_INTERVAL.String(),
		"cleanupAge": SESSION_CLEANUP_AGE.String(),
	})

	go func() {
		ticker := time.NewTicker(SESSION_CLEANUP_INTERVAL)
		defer ticker.Stop()

		for range ticker.C {
			cleanupCount := 0
			cutoffTime := time.Now().Add(-SESSION_CLEANUP_AGE)

			vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
				// Iterate all viewer sessions
				vbolt.IterateAll(tx, ViewerSessionsBkt, func(sessionKey string, session ViewerSession) bool {
					// Delete sessions older than cleanup age
					if session.LastSeenAt.Before(cutoffTime) {
						// Decrement viewer counts before deleting session
						// Get room to find studioId
						room := GetRoom(tx, session.RoomId)
						if room.Id != 0 {
							// Decrement room analytics
							var roomAnalytics RoomAnalytics
							vbolt.Read(tx, RoomAnalyticsBkt, session.RoomId, &roomAnalytics)
							if roomAnalytics.CurrentViewers > 0 {
								roomAnalytics.CurrentViewers--
								vbolt.Write(tx, RoomAnalyticsBkt, session.RoomId, &roomAnalytics)
							}

							// Decrement studio analytics
							var studioAnalytics StudioAnalytics
							vbolt.Read(tx, StudioAnalyticsBkt, room.StudioId, &studioAnalytics)
							if studioAnalytics.CurrentViewers > 0 {
								studioAnalytics.CurrentViewers--
								vbolt.Write(tx, StudioAnalyticsBkt, room.StudioId, &studioAnalytics)
							}
						}

						// If this is a code-based session, decrement code analytics
						if session.Code != "" {
							var codeAnalytics CodeAnalytics
							vbolt.Read(tx, CodeAnalyticsBkt, session.Code, &codeAnalytics)
							if codeAnalytics.CurrentViewers > 0 {
								codeAnalytics.CurrentViewers--
								vbolt.Write(tx, CodeAnalyticsBkt, session.Code, &codeAnalytics)
							}

							// Remove from code index
							vbolt.SetTargetSingleTerm(tx, SessionsByCodeIndex, sessionKey, "")
						}

						// Now delete the session
						vbolt.Delete(tx, ViewerSessionsBkt, sessionKey)
						vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, sessionKey, -1)
						cleanupCount++
					}
					return true // continue iteration
				})

				vbolt.TxCommit(tx)
			})

			if cleanupCount > 0 {
				LogInfo(LogCategorySystem, "Viewer session cleanup completed", map[string]interface{}{
					"cleanedSessions": cleanupCount,
				})
			}
		}
	}()
}

// ResetAllCurrentViewers resets all CurrentViewers counts to 0 across all analytics buckets
// This should be called on server startup since SSE connections cannot persist across restarts
func ResetAllCurrentViewers(db *vbolt.DB) {
	resetCount := 0
	sessionCount := 0

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Reset room analytics
		vbolt.IterateAll(tx, RoomAnalyticsBkt, func(roomId int, analytics RoomAnalytics) bool {
			if analytics.CurrentViewers > 0 {
				analytics.CurrentViewers = 0
				vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)
				resetCount++
			}
			return true // continue iteration
		})

		// Reset studio analytics
		vbolt.IterateAll(tx, StudioAnalyticsBkt, func(studioId int, analytics StudioAnalytics) bool {
			if analytics.CurrentViewers > 0 {
				analytics.CurrentViewers = 0
				vbolt.Write(tx, StudioAnalyticsBkt, studioId, &analytics)
				resetCount++
			}
			return true // continue iteration
		})

		vbolt.TxCommit(tx)
	})

	// Also reset code analytics (imported from code_access.go)
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		codeResetCount := 0
		vbolt.IterateAll(tx, CodeAnalyticsBkt, func(code string, analytics CodeAnalytics) bool {
			if analytics.CurrentViewers > 0 {
				analytics.CurrentViewers = 0
				vbolt.Write(tx, CodeAnalyticsBkt, code, &analytics)
				codeResetCount++
			}
			return true // continue iteration
		})
		vbolt.TxCommit(tx)
		resetCount += codeResetCount
	})

	// Delete all viewer sessions since SSE connections are closed on restart
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		vbolt.IterateAll(tx, ViewerSessionsBkt, func(sessionKey string, session ViewerSession) bool {
			// Delete session from bucket
			vbolt.Delete(tx, ViewerSessionsBkt, sessionKey)
			// Remove from room index
			vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, sessionKey, -1)
			// Remove from code index if applicable
			if session.Code != "" {
				vbolt.SetTargetSingleTerm(tx, SessionsByCodeIndex, sessionKey, "")
			}
			sessionCount++
			return true // continue iteration
		})
		vbolt.TxCommit(tx)
	})

	LogInfo(LogCategorySystem, "Reset all CurrentViewers to 0 on startup", map[string]interface{}{
		"resetCount":      resetCount,
		"deletedSessions": sessionCount,
	})
}
