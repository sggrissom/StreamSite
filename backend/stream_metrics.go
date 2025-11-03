package backend

import (
	"errors"
	"math"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// StreamMetrics represents performance metrics collected from a single viewing session
type StreamMetrics struct {
	RoomId int `json:"roomId"`

	// Startup Performance
	TimeToFirstFrame int  `json:"timeToFirstFrame"` // Milliseconds, -1 if failed
	StartupSucceeded bool `json:"startupSucceeded"` // Whether playback started successfully

	// Buffering Performance
	RebufferEvents  int `json:"rebufferEvents"`  // Number of buffering interruptions
	RebufferSeconds int `json:"rebufferSeconds"` // Total time spent buffering
	WatchSeconds    int `json:"watchSeconds"`    // Total viewing duration (for ratio calculation)

	// Quality Performance
	Seconds480p  int     `json:"seconds480p"`  // Time at 480p
	Seconds720p  int     `json:"seconds720p"`  // Time at 720p
	Seconds1080p int     `json:"seconds1080p"` // Time at 1080p
	AvgBitrate   float64 `json:"avgBitrate"`   // Average bitrate in Mbps

	// Error Tracking
	NetworkErrors int `json:"networkErrors"` // Network/loading errors
	MediaErrors   int `json:"mediaErrors"`   // Decoding/format errors
}

// ReportStreamMetricsRequest is the request for reporting metrics from the frontend
type ReportStreamMetricsRequest struct {
	Metrics StreamMetrics `json:"metrics"`
}

// ReportStreamMetricsResponse is the response after reporting metrics
type ReportStreamMetricsResponse struct {
}

// ReportStreamMetrics receives performance metrics from the frontend and updates analytics
// Note: This procedure allows metrics from any viewer (authenticated or anonymous)
// since performance data is valuable regardless of authentication status
func ReportStreamMetrics(ctx *vbeam.Context, req ReportStreamMetricsRequest) (resp ReportStreamMetricsResponse, err error) {
	// Validate room ID
	if req.Metrics.RoomId <= 0 {
		return resp, errors.New("Invalid room ID")
	}

	// Get room to validate it exists
	room := GetRoom(ctx.Tx, req.Metrics.RoomId)
	if room.Id == 0 {
		return resp, errors.New("Room not found")
	}

	// Upgrade to write transaction
	vbeam.UseWriteTx(ctx)

	// Update room analytics
	updateRoomPerformanceMetrics(ctx.Tx, req.Metrics)

	// Update studio analytics
	updateStudioPerformanceMetrics(ctx.Tx, room.StudioId, req.Metrics)

	vbolt.TxCommit(ctx.Tx)

	return resp, nil
}

// updateRoomPerformanceMetrics updates room analytics with new performance data using weighted averages
func updateRoomPerformanceMetrics(tx *vbolt.Tx, metrics StreamMetrics) {
	var analytics RoomAnalytics
	vbolt.Read(tx, RoomAnalyticsBkt, metrics.RoomId, &analytics)

	if analytics.RoomId == 0 {
		analytics.RoomId = metrics.RoomId
	}

	// Update startup metrics
	analytics.StartupAttempts++
	if metrics.StartupSucceeded {
		// Update time to first frame using weighted average (recent data weighted more)
		if analytics.AvgTimeToFirstFrame == 0 {
			analytics.AvgTimeToFirstFrame = metrics.TimeToFirstFrame
		} else if metrics.TimeToFirstFrame > 0 {
			// Weighted average: 70% old, 30% new (favors recent performance)
			analytics.AvgTimeToFirstFrame = int(float64(analytics.AvgTimeToFirstFrame)*0.7 + float64(metrics.TimeToFirstFrame)*0.3)
		}
	} else {
		analytics.StartupFailures++
	}

	// Update buffering metrics
	analytics.TotalRebufferEvents += metrics.RebufferEvents
	analytics.TotalRebufferSeconds += metrics.RebufferSeconds

	// Calculate rebuffer ratio for this session
	if metrics.WatchSeconds > 0 {
		sessionRebufferRatio := (float64(metrics.RebufferSeconds) / float64(metrics.WatchSeconds)) * 100.0
		// Cap at 100%
		if sessionRebufferRatio > 100.0 {
			sessionRebufferRatio = 100.0
		}

		// Update average rebuffer ratio using weighted average
		if analytics.AvgRebufferRatio == 0 {
			analytics.AvgRebufferRatio = sessionRebufferRatio
		} else {
			analytics.AvgRebufferRatio = analytics.AvgRebufferRatio*0.7 + sessionRebufferRatio*0.3
		}
	}

	// Update quality distribution
	analytics.Quality480pSeconds += metrics.Seconds480p
	analytics.Quality720pSeconds += metrics.Seconds720p
	analytics.Quality1080pSeconds += metrics.Seconds1080p

	// Update average bitrate using weighted average
	if metrics.AvgBitrate > 0 {
		if analytics.AvgBitrateMbps == 0 {
			analytics.AvgBitrateMbps = metrics.AvgBitrate
		} else {
			analytics.AvgBitrateMbps = analytics.AvgBitrateMbps*0.7 + metrics.AvgBitrate*0.3
		}
	}

	// Update error counts
	totalSessionErrors := metrics.NetworkErrors + metrics.MediaErrors
	analytics.TotalErrors += totalSessionErrors
	analytics.NetworkErrors += metrics.NetworkErrors
	analytics.MediaErrors += metrics.MediaErrors

	// Round floating point values to avoid precision drift
	analytics.AvgRebufferRatio = math.Round(analytics.AvgRebufferRatio*100) / 100
	analytics.AvgBitrateMbps = math.Round(analytics.AvgBitrateMbps*100) / 100

	vbolt.Write(tx, RoomAnalyticsBkt, metrics.RoomId, &analytics)

	LogInfo(LogCategorySystem, "Updated room performance metrics", map[string]interface{}{
		"roomId":           metrics.RoomId,
		"ttff":             metrics.TimeToFirstFrame,
		"startupSucceeded": metrics.StartupSucceeded,
		"rebufferEvents":   metrics.RebufferEvents,
		"watchSeconds":     metrics.WatchSeconds,
	})
}

// updateStudioPerformanceMetrics aggregates performance metrics from all rooms in a studio
func updateStudioPerformanceMetrics(tx *vbolt.Tx, studioId int, metrics StreamMetrics) {
	// Get all rooms in the studio
	rooms := ListStudioRooms(tx, studioId)

	// Aggregate performance metrics across all rooms
	var studioAnalytics StudioAnalytics
	studioAnalytics.StudioId = studioId
	studioAnalytics.TotalRooms = len(rooms)

	var totalStartupAttempts int
	var weightedTTFF float64
	var weightedRebufferRatio float64
	var weightedBitrate float64

	for _, room := range rooms {
		var roomAnalytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, room.Id, &roomAnalytics)

		// Aggregate viewership stats (existing logic from UpdateStudioAnalyticsFromRoom)
		studioAnalytics.TotalViewsAllTime += roomAnalytics.TotalViewsAllTime
		studioAnalytics.TotalViewsThisMonth += roomAnalytics.TotalViewsThisMonth
		studioAnalytics.UniqueViewersAllTime += roomAnalytics.UniqueViewersAllTime
		studioAnalytics.UniqueViewersThisMonth += roomAnalytics.UniqueViewersThisMonth
		studioAnalytics.CurrentViewers += roomAnalytics.CurrentViewers
		studioAnalytics.TotalStreamMinutes += roomAnalytics.TotalStreamMinutes

		if room.IsActive {
			studioAnalytics.ActiveRooms++
		}

		// Aggregate performance metrics (weighted by number of startup attempts)
		if roomAnalytics.StartupAttempts > 0 {
			weight := float64(roomAnalytics.StartupAttempts)
			weightedTTFF += float64(roomAnalytics.AvgTimeToFirstFrame) * weight
			weightedRebufferRatio += roomAnalytics.AvgRebufferRatio * weight
			weightedBitrate += roomAnalytics.AvgBitrateMbps * weight
			totalStartupAttempts += roomAnalytics.StartupAttempts
		}

		// Sum cumulative metrics
		studioAnalytics.StartupAttempts += roomAnalytics.StartupAttempts
		studioAnalytics.StartupFailures += roomAnalytics.StartupFailures
		studioAnalytics.TotalRebufferEvents += roomAnalytics.TotalRebufferEvents
		studioAnalytics.TotalRebufferSeconds += roomAnalytics.TotalRebufferSeconds
		studioAnalytics.Quality480pSeconds += roomAnalytics.Quality480pSeconds
		studioAnalytics.Quality720pSeconds += roomAnalytics.Quality720pSeconds
		studioAnalytics.Quality1080pSeconds += roomAnalytics.Quality1080pSeconds
		studioAnalytics.TotalErrors += roomAnalytics.TotalErrors
		studioAnalytics.NetworkErrors += roomAnalytics.NetworkErrors
		studioAnalytics.MediaErrors += roomAnalytics.MediaErrors
	}

	// Calculate weighted averages
	if totalStartupAttempts > 0 {
		studioAnalytics.AvgTimeToFirstFrame = int(weightedTTFF / float64(totalStartupAttempts))
		studioAnalytics.AvgRebufferRatio = math.Round((weightedRebufferRatio/float64(totalStartupAttempts))*100) / 100
		studioAnalytics.AvgBitrateMbps = math.Round((weightedBitrate/float64(totalStartupAttempts))*100) / 100
	}

	vbolt.Write(tx, StudioAnalyticsBkt, studioId, &studioAnalytics)
}

// RegisterStreamMetricsMethods registers all stream metrics API procedures
func RegisterStreamMetricsMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, ReportStreamMetrics)
}
