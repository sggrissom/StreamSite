package backend

import (
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

// Helper functions for analytics tracking

func IncrementRoomViewerCount(db *vbolt.DB, roomId int) {
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Get room to find studioId
		room := GetRoom(tx, roomId)
		if room.Id == 0 {
			return
		}

		// Load or create room analytics
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, roomId, &analytics)
		if analytics.RoomId == 0 {
			analytics.RoomId = roomId
		}

		// Increment counters
		analytics.CurrentViewers++
		analytics.TotalViewsAllTime++
		analytics.TotalViewsThisMonth++

		// Update peak if necessary
		if analytics.CurrentViewers > analytics.PeakViewers {
			analytics.PeakViewers = analytics.CurrentViewers
			analytics.PeakViewersAt = time.Now()
		}

		// Save
		vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)

		incrementStudioViewerCount(tx, room.StudioId)

		vbolt.TxCommit(tx)
	})
}

func DecrementRoomViewerCount(db *vbolt.DB, roomId int) {
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := GetRoom(tx, roomId)
		if room.Id == 0 {
			return
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
	})
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

func incrementStudioViewerCount(tx *vbolt.Tx, studioId int) {
	var studioAnalytics StudioAnalytics
	vbolt.Read(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)

	// Initialize if first viewer
	if studioAnalytics.StudioId == 0 {
		studioAnalytics.StudioId = studioId
	}

	// Increment viewer counts
	studioAnalytics.CurrentViewers++
	studioAnalytics.TotalViewsAllTime++
	studioAnalytics.TotalViewsThisMonth++

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
	Success     bool           `json:"success"`
	Error       string         `json:"error,omitempty"`
	Analytics   *RoomAnalytics `json:"analytics,omitempty"`
	IsStreaming bool           `json:"isStreaming"`
	RoomName    string         `json:"roomName"`
}

func GetRoomAnalytics(ctx *vbeam.Context, req GetRoomAnalyticsRequest) (resp GetRoomAnalyticsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Validate room ID
	if req.RoomId <= 0 {
		resp.Success = false
		resp.Error = "Invalid room ID"
		return
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		resp.Success = false
		resp.Error = "Room not found"
		return
	}

	// Check studio access (requires at least viewer role)
	access := CheckStudioAccess(ctx.Tx, caller, room.StudioId)
	if !access.Allowed {
		resp.Success = false
		resp.Error = "Access denied"
		return
	}

	// Load analytics
	var analytics RoomAnalytics
	vbolt.Read(ctx.Tx, RoomAnalyticsBkt, req.RoomId, &analytics)

	// Initialize if not found
	if analytics.RoomId == 0 {
		analytics.RoomId = req.RoomId
	}

	resp.Success = true
	resp.Analytics = &analytics
	resp.IsStreaming = room.IsActive
	resp.RoomName = room.Name
	return
}

type GetStudioAnalyticsRequest struct {
	StudioId int `json:"studioId"`
}

type GetStudioAnalyticsResponse struct {
	Success    bool             `json:"success"`
	Error      string           `json:"error,omitempty"`
	Analytics  *StudioAnalytics `json:"analytics,omitempty"`
	StudioName string           `json:"studioName"`
}

func GetStudioAnalytics(ctx *vbeam.Context, req GetStudioAnalyticsRequest) (resp GetStudioAnalyticsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Validate studio ID
	if req.StudioId <= 0 {
		resp.Success = false
		resp.Error = "Invalid studio ID"
		return
	}

	// Check studio access (requires at least viewer role)
	access := CheckStudioAccess(ctx.Tx, caller, req.StudioId)
	if !access.Allowed {
		resp.Success = false
		resp.Error = "Access denied"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Load studio analytics
	var analytics StudioAnalytics
	vbolt.Read(ctx.Tx, StudioAnalyticsBkt, req.StudioId, &analytics)

	// Initialize if not found
	if analytics.StudioId == 0 {
		analytics.StudioId = req.StudioId
	}

	resp.Success = true
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

// ResetMonthlyAnalytics resets the TotalViewsThisMonth counter for all rooms and studios
// Should be called on the first day of each month
func ResetMonthlyAnalytics(db *vbolt.DB) int {
	resetCount := 0

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Reset room analytics
		vbolt.IterateAll(tx, RoomAnalyticsBkt, func(roomId int, analytics RoomAnalytics) bool {
			if analytics.TotalViewsThisMonth > 0 {
				analytics.TotalViewsThisMonth = 0
				vbolt.Write(tx, RoomAnalyticsBkt, roomId, &analytics)
				resetCount++
			}
			return true // continue iteration
		})

		// Reset studio analytics
		vbolt.IterateAll(tx, StudioAnalyticsBkt, func(studioId int, analytics StudioAnalytics) bool {
			if analytics.TotalViewsThisMonth > 0 {
				analytics.TotalViewsThisMonth = 0
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
