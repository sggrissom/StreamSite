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
