package backend

import (
	"stream/cfg"
	"time"

	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// RoomAnalytics tracks viewing statistics and streaming metrics for a room
type RoomAnalytics struct {
	RoomId              int       `json:"roomId"`
	TotalViewsAllTime   int       `json:"totalViewsAllTime"`   // Total connection count
	TotalViewsThisMonth int       `json:"totalViewsThisMonth"` // Reset monthly
	CurrentViewers      int       `json:"currentViewers"`      // Live count
	PeakViewers         int       `json:"peakViewers"`         // Historical max
	PeakViewersAt       time.Time `json:"peakViewersAt"`
	LastStreamAt        time.Time `json:"lastStreamAt"`       // Last IsActive=true
	StreamStartedAt     time.Time `json:"streamStartedAt"`    // When current stream began
	TotalStreamMinutes  int       `json:"totalStreamMinutes"` // Lifetime streaming duration
}

// StudioAnalytics tracks aggregated statistics across all rooms in a studio
type StudioAnalytics struct {
	StudioId            int `json:"studioId"`
	TotalViewsAllTime   int `json:"totalViewsAllTime"`
	TotalViewsThisMonth int `json:"totalViewsThisMonth"`
	CurrentViewers      int `json:"currentViewers"` // Sum across all rooms
	TotalRooms          int `json:"totalRooms"`
	ActiveRooms         int `json:"activeRooms"` // Currently streaming
	TotalStreamMinutes  int `json:"totalStreamMinutes"`
}

// Packing functions for vbolt serialization

func PackRoomAnalytics(self *RoomAnalytics, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.Int(&self.TotalViewsAllTime, buf)
	vpack.Int(&self.TotalViewsThisMonth, buf)
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
	vpack.Int(&self.CurrentViewers, buf)
	vpack.Int(&self.TotalRooms, buf)
	vpack.Int(&self.ActiveRooms, buf)
	vpack.Int(&self.TotalStreamMinutes, buf)
}

// Buckets for entity storage

// RoomAnalyticsBkt: roomId (int) -> RoomAnalytics
var RoomAnalyticsBkt = vbolt.Bucket(&cfg.Info, "room_analytics", vpack.FInt, PackRoomAnalytics)

// StudioAnalyticsBkt: studioId (int) -> StudioAnalytics
var StudioAnalyticsBkt = vbolt.Bucket(&cfg.Info, "studio_analytics", vpack.FInt, PackStudioAnalytics)

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
