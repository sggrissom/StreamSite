package backend

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// Admin procedures for manual system maintenance and troubleshooting

// RecalculateViewerCountsRequest is the request for RecalculateViewerCounts
type RecalculateViewerCountsRequest struct {
	// StudioId restricts recalculation to a specific studio (0 = all studios)
	StudioId int `json:"studioId"`
}

// RecalculateViewerCountsResponse is the response for RecalculateViewerCounts
type RecalculateViewerCountsResponse struct {
	RoomsUpdated   int    `json:"roomsUpdated"`
	StudiosUpdated int    `json:"studiosUpdated"`
	CodesUpdated   int    `json:"codesUpdated"`
	Message        string `json:"message,omitempty"`
}

// RecalculateViewerCounts manually recalculates CurrentViewers counts by querying
// actual SSE connections. This is useful for fixing stuck counts after crashes,
// ungraceful shutdowns, or bugs in the increment/decrement logic.
//
// This procedure:
//  1. Queries SSE manager for actual active connections per room
//  2. Updates RoomAnalytics.CurrentViewers to match reality
//  3. Recalculates StudioAnalytics by aggregating room counts
//  4. Updates CodeAnalytics by counting active code sessions in SSE
//
// Requires admin permission for at least one studio (or global admin).
func RecalculateViewerCounts(ctx *vbeam.Context, req RecalculateViewerCountsRequest) (resp RecalculateViewerCountsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Get actual viewer counts from SSE manager
	actualCounts := sseManager.GetCurrentViewerCounts()

	vbeam.UseWriteTx(ctx)

	roomsUpdated := 0
	studiosUpdated := 0

	// Update room analytics with actual counts
	vbolt.IterateAll(ctx.Tx, RoomAnalyticsBkt, func(roomId int, analytics RoomAnalytics) bool {
		// Get room to check studio access
		room := GetRoom(ctx.Tx, roomId)
		if room.Id == 0 {
			return true // Skip invalid rooms
		}

		// If filtering by studio, skip rooms not in that studio
		if req.StudioId > 0 && room.StudioId != req.StudioId {
			return true
		}

		// Check if caller has admin permission for this room's studio
		if !HasStudioPermission(ctx.Tx, caller.Id, room.StudioId, StudioRoleAdmin) {
			return true // Skip rooms where caller is not admin
		}

		// Get actual count from SSE manager (0 if no connections)
		actualCount := actualCounts[roomId]

		// Update if different
		if analytics.CurrentViewers != actualCount {
			analytics.CurrentViewers = actualCount
			vbolt.Write(ctx.Tx, RoomAnalyticsBkt, roomId, &analytics)
			roomsUpdated++
		}

		return true // continue iteration
	})

	// Recalculate studio analytics by aggregating room counts
	if req.StudioId > 0 {
		// Update specific studio
		UpdateStudioAnalyticsFromRoom(ctx.Tx, req.StudioId)
		studiosUpdated = 1
	} else {
		// Update all studios where caller is admin
		vbolt.IterateAll(ctx.Tx, StudiosBkt, func(studioId int, studio Studio) bool {
			if HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleAdmin) {
				UpdateStudioAnalyticsFromRoom(ctx.Tx, studioId)
				studiosUpdated++
			}
			return true
		})
	}

	// Recalculate code analytics by counting active code sessions in SSE
	codesUpdated := 0
	sessionTokenCounts := sseManager.GetCodeSessionCounts()

	// Map session tokens to codes
	codeCountsMap := make(map[string]int)
	for sessionToken := range sessionTokenCounts {
		var session CodeSession
		vbolt.Read(ctx.Tx, CodeSessionsBkt, sessionToken, &session)
		if session.Token != "" {
			codeCountsMap[session.Code]++
		}
	}

	vbolt.IterateAll(ctx.Tx, CodeAnalyticsBkt, func(code string, analytics CodeAnalytics) bool {
		// Get the access code to check permissions
		var accessCode AccessCode
		vbolt.Read(ctx.Tx, AccessCodesBkt, code, &accessCode)
		if accessCode.Code == "" {
			return true // Skip if code not found
		}

		// Determine studio ID for permission check
		var studioId int
		if accessCode.Type == CodeTypeRoom {
			room := GetRoom(ctx.Tx, accessCode.TargetId)
			if room.Id == 0 {
				return true
			}
			studioId = room.StudioId
		} else {
			studioId = accessCode.TargetId
		}

		// If filtering by studio, skip codes not in that studio
		if req.StudioId > 0 && studioId != req.StudioId {
			return true
		}

		// Check if caller has admin permission
		if !HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleAdmin) {
			return true
		}

		// Get actual count from mapped codes
		actualCount := codeCountsMap[code]

		// Update if different
		if analytics.CurrentViewers != actualCount {
			analytics.CurrentViewers = actualCount
			vbolt.Write(ctx.Tx, CodeAnalyticsBkt, code, &analytics)
			codesUpdated++
		}

		return true
	})

	vbolt.TxCommit(ctx.Tx)

	// Log the recalculation
	LogInfo(LogCategorySystem, "Manual viewer count recalculation completed", map[string]interface{}{
		"roomsUpdated":   roomsUpdated,
		"studiosUpdated": studiosUpdated,
		"codesUpdated":   codesUpdated,
		"requestedBy":    caller.Id,
		"userEmail":      caller.Email,
		"studioFilter":   req.StudioId,
	})

	resp.RoomsUpdated = roomsUpdated
	resp.StudiosUpdated = studiosUpdated
	resp.CodesUpdated = codesUpdated
	resp.Message = "Viewer counts recalculated successfully"
	return
}

// GetSystemLogsRequest is the request for GetSystemLogs
type GetSystemLogsRequest struct {
	Level    *string `json:"level,omitempty"`    // Optional: filter by log level (INFO, WARN, ERROR, DEBUG)
	Category *string `json:"category,omitempty"` // Optional: filter by category (AUTH, STREAM, API, SYSTEM)
	Search   *string `json:"search,omitempty"`   // Optional: text search in message
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Category  string                 `json:"category"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	UserID    *int                   `json:"userId,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"userAgent,omitempty"`
}

// GetSystemLogsResponse is the response for GetSystemLogs
type GetSystemLogsResponse struct {
	Logs       []LogEntry `json:"logs"`
	TotalCount int        `json:"totalCount"`
}

// GetSystemLogs retrieves and filters system logs. Only accessible by site admins.
func GetSystemLogs(ctx *vbeam.Context, req GetSystemLogsRequest) (resp GetSystemLogsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Only site admins can access logs
	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can access system logs")
	}

	// Open log file
	logFile := "logs/stream.log"
	file, err := os.Open(logFile)
	if err != nil {
		LogWarn(LogCategorySystem, "Failed to open log file", map[string]interface{}{
			"error":       err.Error(),
			"requestedBy": caller.Id,
		})
		return resp, errors.New("Failed to open log file")
	}
	defer file.Close()

	// Read and parse log entries
	var allLogs []LogEntry
	scanner := bufio.NewScanner(file)

	// Increase buffer size to handle long log lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Extract JSON portion (log lines may have date prefix before JSON)
		jsonStart := strings.Index(line, "{")
		if jsonStart == -1 {
			// No JSON in this line
			continue
		}

		// Parse JSON portion only
		var entry LogEntry
		if err := json.Unmarshal([]byte(line[jsonStart:]), &entry); err != nil {
			// Skip non-JSON lines (like HTTP server logs)
			continue
		}

		// Apply filters
		if req.Level != nil && *req.Level != "" && entry.Level != *req.Level {
			continue
		}

		if req.Category != nil && *req.Category != "" && entry.Category != *req.Category {
			continue
		}

		if req.Search != nil && *req.Search != "" {
			searchLower := strings.ToLower(*req.Search)
			if !strings.Contains(strings.ToLower(entry.Message), searchLower) {
				continue
			}
		}

		allLogs = append(allLogs, entry)
	}

	if err := scanner.Err(); err != nil {
		LogErrorSimple(LogCategorySystem, "Error reading log file", map[string]interface{}{
			"error":       err.Error(),
			"requestedBy": caller.Id,
		})
		return resp, errors.New("Error reading log file")
	}

	// Reverse array to show newest first (file is chronological oldest-to-newest)
	for i, j := 0, len(allLogs)-1; i < j; i, j = i+1, j-1 {
		allLogs[i], allLogs[j] = allLogs[j], allLogs[i]
	}

	// Limit to last 100 entries
	totalCount := len(allLogs)
	if len(allLogs) > 100 {
		allLogs = allLogs[:100]
	}

	resp.Logs = allLogs
	resp.TotalCount = totalCount

	// Log the access
	LogInfo(LogCategorySystem, "System logs accessed", map[string]interface{}{
		"requestedBy":  caller.Id,
		"userEmail":    caller.Email,
		"filterLevel":  req.Level,
		"filterCat":    req.Category,
		"filterSearch": req.Search,
		"resultCount":  len(allLogs),
	})

	return
}

// StudioPerformanceMetrics contains performance metrics for a single studio
type StudioPerformanceMetrics struct {
	StudioId            int     `json:"studioId"`
	StudioName          string  `json:"studioName"`
	TotalRooms          int     `json:"totalRooms"`
	AvgTimeToFirstFrame int     `json:"avgTimeToFirstFrame"` // ms
	StartupSuccessRate  float64 `json:"startupSuccessRate"`  // percentage
	AvgRebufferRatio    float64 `json:"avgRebufferRatio"`    // percentage
	AvgBitrateMbps      float64 `json:"avgBitrateMbps"`      // Mbps
	TotalErrors         int     `json:"totalErrors"`
	AvgErrorsPerSession float64 `json:"avgErrorsPerSession"` // average errors per viewing session
}

// SitePerformanceMetrics contains aggregated performance metrics across the entire site
type SitePerformanceMetrics struct {
	// Site-wide aggregated metrics
	AvgTimeToFirstFrame  int     `json:"avgTimeToFirstFrame"` // ms
	StartupSuccessRate   float64 `json:"startupSuccessRate"`  // percentage
	AvgRebufferRatio     float64 `json:"avgRebufferRatio"`    // percentage
	AvgBitrateMbps       float64 `json:"avgBitrateMbps"`      // Mbps
	TotalStartupAttempts int     `json:"totalStartupAttempts"`
	TotalStartupFailures int     `json:"totalStartupFailures"`
	TotalRebufferEvents  int     `json:"totalRebufferEvents"`
	TotalRebufferSeconds int     `json:"totalRebufferSeconds"`
	TotalErrors          int     `json:"totalErrors"`
	NetworkErrors        int     `json:"networkErrors"`
	MediaErrors          int     `json:"mediaErrors"`
	AvgErrorsPerSession  float64 `json:"avgErrorsPerSession"` // average errors per viewing session

	// Quality distribution
	Quality480pSeconds  int     `json:"quality480pSeconds"`
	Quality720pSeconds  int     `json:"quality720pSeconds"`
	Quality1080pSeconds int     `json:"quality1080pSeconds"`
	Quality480pPercent  float64 `json:"quality480pPercent"`
	Quality720pPercent  float64 `json:"quality720pPercent"`
	Quality1080pPercent float64 `json:"quality1080pPercent"`

	// Count of rooms with data
	TotalRoomsWithData int `json:"totalRoomsWithData"`
}

// GetSitePerformanceMetricsRequest is the request for GetSitePerformanceMetrics
type GetSitePerformanceMetricsRequest struct {
	// Empty for now, could add date range filters in future
}

// GetSitePerformanceMetricsResponse is the response for GetSitePerformanceMetrics
type GetSitePerformanceMetricsResponse struct {
	SiteWide  SitePerformanceMetrics     `json:"siteWide"`
	PerStudio []StudioPerformanceMetrics `json:"perStudio"`
}

// GetSitePerformanceMetrics retrieves aggregated performance metrics across the entire site.
// Only accessible by site admins.
func GetSitePerformanceMetrics(ctx *vbeam.Context, req GetSitePerformanceMetricsRequest) (resp GetSitePerformanceMetricsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Only site admins can access site-wide performance metrics
	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can access site-wide performance metrics")
	}

	// Aggregate site-wide metrics
	var siteWide SitePerformanceMetrics
	var totalWeightedTTFF, totalWeightedRebuffer, totalWeightedBitrate float64
	var totalStartupAttempts int

	// Map to accumulate per-studio metrics
	studioMetricsMap := make(map[int]*StudioPerformanceMetrics)

	// Iterate through all room analytics
	vbolt.IterateAll(ctx.Tx, RoomAnalyticsBkt, func(roomId int, analytics RoomAnalytics) bool {
		// Skip rooms with no performance data
		if analytics.StartupAttempts == 0 {
			return true
		}

		// Get room to find studio
		room := GetRoom(ctx.Tx, roomId)
		if room.Id == 0 {
			return true
		}

		// Accumulate site-wide metrics (weighted averages)
		weight := float64(analytics.StartupAttempts)
		totalWeightedTTFF += float64(analytics.AvgTimeToFirstFrame) * weight
		totalWeightedRebuffer += analytics.AvgRebufferRatio * weight
		totalWeightedBitrate += analytics.AvgBitrateMbps * weight
		totalStartupAttempts += analytics.StartupAttempts

		// Accumulate totals
		siteWide.TotalStartupAttempts += analytics.StartupAttempts
		siteWide.TotalStartupFailures += analytics.StartupFailures
		siteWide.TotalRebufferEvents += analytics.TotalRebufferEvents
		siteWide.TotalRebufferSeconds += analytics.TotalRebufferSeconds
		siteWide.TotalErrors += analytics.TotalErrors
		siteWide.NetworkErrors += analytics.NetworkErrors
		siteWide.MediaErrors += analytics.MediaErrors
		siteWide.Quality480pSeconds += analytics.Quality480pSeconds
		siteWide.Quality720pSeconds += analytics.Quality720pSeconds
		siteWide.Quality1080pSeconds += analytics.Quality1080pSeconds
		siteWide.TotalRoomsWithData++

		// Accumulate per-studio metrics
		studioMetrics, exists := studioMetricsMap[room.StudioId]
		if !exists {
			studio := GetStudioById(ctx.Tx, room.StudioId)
			studioMetrics = &StudioPerformanceMetrics{
				StudioId:   room.StudioId,
				StudioName: studio.Name,
			}
			studioMetricsMap[room.StudioId] = studioMetrics
		}

		studioMetrics.TotalRooms++
		// We'll calculate averages after the loop using studio analytics

		return true // continue
	})

	// Calculate site-wide weighted averages
	if totalStartupAttempts > 0 {
		siteWide.AvgTimeToFirstFrame = int(totalWeightedTTFF / float64(totalStartupAttempts))
		siteWide.AvgRebufferRatio = totalWeightedRebuffer / float64(totalStartupAttempts)
		siteWide.AvgBitrateMbps = totalWeightedBitrate / float64(totalStartupAttempts)
		siteWide.StartupSuccessRate = float64(siteWide.TotalStartupAttempts-siteWide.TotalStartupFailures) / float64(siteWide.TotalStartupAttempts) * 100
		siteWide.AvgErrorsPerSession = float64(siteWide.TotalErrors) / float64(siteWide.TotalStartupAttempts)
	}

	// Calculate quality distribution percentages
	totalQualitySeconds := siteWide.Quality480pSeconds + siteWide.Quality720pSeconds + siteWide.Quality1080pSeconds
	if totalQualitySeconds > 0 {
		siteWide.Quality480pPercent = float64(siteWide.Quality480pSeconds) / float64(totalQualitySeconds) * 100
		siteWide.Quality720pPercent = float64(siteWide.Quality720pSeconds) / float64(totalQualitySeconds) * 100
		siteWide.Quality1080pPercent = float64(siteWide.Quality1080pSeconds) / float64(totalQualitySeconds) * 100
	}

	// Fill in per-studio metrics from studio analytics
	for studioId, studioMetrics := range studioMetricsMap {
		var studioAnalytics StudioAnalytics
		vbolt.Read(ctx.Tx, StudioAnalyticsBkt, studioId, &studioAnalytics)

		if studioAnalytics.StartupAttempts > 0 {
			studioMetrics.AvgTimeToFirstFrame = studioAnalytics.AvgTimeToFirstFrame
			studioMetrics.AvgRebufferRatio = studioAnalytics.AvgRebufferRatio
			studioMetrics.AvgBitrateMbps = studioAnalytics.AvgBitrateMbps
			studioMetrics.TotalErrors = studioAnalytics.TotalErrors
			studioMetrics.StartupSuccessRate = float64(studioAnalytics.StartupAttempts-studioAnalytics.StartupFailures) / float64(studioAnalytics.StartupAttempts) * 100
			studioMetrics.AvgErrorsPerSession = float64(studioAnalytics.TotalErrors) / float64(studioAnalytics.StartupAttempts)
		}
	}

	// Convert map to slice for response
	resp.SiteWide = siteWide
	resp.PerStudio = make([]StudioPerformanceMetrics, 0, len(studioMetricsMap))
	for _, studioMetrics := range studioMetricsMap {
		resp.PerStudio = append(resp.PerStudio, *studioMetrics)
	}

	// Log the access
	LogInfo(LogCategorySystem, "Site performance metrics accessed", map[string]interface{}{
		"requestedBy": caller.Id,
		"userEmail":   caller.Email,
		"studioCount": len(resp.PerStudio),
		"roomCount":   siteWide.TotalRoomsWithData,
	})

	return
}

// BucketInfo describes a database bucket or index
type BucketInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	KeyType     string `json:"keyType"`   // "int" or "string"
	ValueType   string `json:"valueType"` // e.g., "User", "Room", "[]byte"
	IsIndex     bool   `json:"isIndex"`
}

// ListBucketsRequest is the request for ListBuckets
type ListBucketsRequest struct{}

// ListBucketsResponse is the response for ListBuckets
type ListBucketsResponse struct {
	Buckets []BucketInfo `json:"buckets"`
}

// ListBuckets returns a list of all database buckets and indexes.
// Only accessible by site admins.
func ListBuckets(ctx *vbeam.Context, req ListBucketsRequest) (resp ListBucketsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Only site admins can access bucket viewer
	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can access bucket viewer")
	}

	// Hardcoded list of all buckets and indexes
	resp.Buckets = []BucketInfo{
		// Studios & Rooms
		{Name: "studios", Description: "Studios (organizations)", KeyType: "int", ValueType: "Studio", IsIndex: false},
		{Name: "rooms", Description: "Streaming rooms", KeyType: "int", ValueType: "Room", IsIndex: false},
		{Name: "streams", Description: "Stream history records", KeyType: "int", ValueType: "Stream", IsIndex: false},
		{Name: "room_stream_keys", Description: "Stream key lookup", KeyType: "string", ValueType: "int (roomId)", IsIndex: false},

		// Analytics
		{Name: "room_analytics", Description: "Room viewing analytics", KeyType: "int", ValueType: "RoomAnalytics", IsIndex: false},
		{Name: "studio_analytics", Description: "Studio aggregated analytics", KeyType: "int", ValueType: "StudioAnalytics", IsIndex: false},
		{Name: "viewer_sessions", Description: "Active viewer sessions", KeyType: "string", ValueType: "ViewerSession", IsIndex: false},

		// Camera & Membership
		{Name: "camera_config", Description: "RTSP camera configurations", KeyType: "int", ValueType: "CameraConfig", IsIndex: false},
		{Name: "memberships", Description: "Studio memberships", KeyType: "int", ValueType: "StudioMembership", IsIndex: false},

		// Code Access System
		{Name: "access_codes", Description: "Temporary access codes", KeyType: "string", ValueType: "AccessCode", IsIndex: false},
		{Name: "code_sessions", Description: "Code-based sessions", KeyType: "string", ValueType: "CodeSession", IsIndex: false},
		{Name: "code_analytics", Description: "Code usage analytics", KeyType: "string", ValueType: "CodeAnalytics", IsIndex: false},
		{Name: "user_code_sessions", Description: "User code session mapping", KeyType: "int", ValueType: "string (token)", IsIndex: false},

		// Users & Auth
		{Name: "users", Description: "User accounts", KeyType: "int", ValueType: "User", IsIndex: false},
		{Name: "emails", Description: "Email to userId lookup", KeyType: "string", ValueType: "int (userId)", IsIndex: false},
		{Name: "refresh_tokens", Description: "User refresh tokens", KeyType: "int", ValueType: "RefreshToken", IsIndex: false},
		{Name: "refresh_token_lookup", Description: "Token string to userId lookup", KeyType: "string", ValueType: "int (userId)", IsIndex: false},

		// Indexes
		{Name: "sessions_by_room", Description: "Index: roomId -> sessionKeys", KeyType: "int", ValueType: "[]string", IsIndex: true},
		{Name: "sessions_by_code", Description: "Index: code -> sessionKeys", KeyType: "string", ValueType: "[]string", IsIndex: true},
		{Name: "refresh_tokens_by_user", Description: "Index: userId -> tokenIds", KeyType: "int", ValueType: "[]int", IsIndex: true},
		{Name: "rooms_by_studio", Description: "Index: studioId -> roomIds", KeyType: "int", ValueType: "[]int", IsIndex: true},
		{Name: "streams_by_studio", Description: "Index: studioId -> streamIds", KeyType: "int", ValueType: "[]int", IsIndex: true},
		{Name: "streams_by_room", Description: "Index: roomId -> streamIds", KeyType: "int", ValueType: "[]int", IsIndex: true},
		{Name: "codes_by_room", Description: "Index: roomId -> codes", KeyType: "int", ValueType: "[]string", IsIndex: true},
		{Name: "codes_by_studio", Description: "Index: studioId -> codes", KeyType: "int", ValueType: "[]string", IsIndex: true},
		{Name: "codes_by_creator", Description: "Index: userId -> codes", KeyType: "int", ValueType: "[]string", IsIndex: true},
		{Name: "memberships_by_user", Description: "Index: userId -> membershipIds", KeyType: "int", ValueType: "[]int", IsIndex: true},
		{Name: "memberships_by_studio", Description: "Index: studioId -> membershipIds", KeyType: "int", ValueType: "[]int", IsIndex: true},
	}

	// Log the access
	LogInfo(LogCategorySystem, "Bucket list accessed", map[string]interface{}{
		"requestedBy": caller.Id,
		"userEmail":   caller.Email,
	})

	return
}

// BucketEntry represents a key-value entry in a bucket
type BucketEntry struct {
	Key   interface{} `json:"key"`   // Can be int or string
	Value interface{} `json:"value"` // JSON representation of the value
}

// GetBucketDataRequest is the request for GetBucketData
type GetBucketDataRequest struct {
	BucketName string `json:"bucketName"`
}

// GetBucketDataResponse is the response for GetBucketData
type GetBucketDataResponse struct {
	Entries []BucketEntry `json:"entries"`
	Total   int           `json:"total"`
}

// GetBucketData retrieves all data from a specific bucket.
// Only accessible by site admins.
func GetBucketData(ctx *vbeam.Context, req GetBucketDataRequest) (resp GetBucketDataResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("Authentication required")
	}

	// Only site admins can access bucket data
	if caller.Role != RoleSiteAdmin {
		return resp, errors.New("Only site admins can access bucket data")
	}

	// Initialize response
	resp.Entries = []BucketEntry{}

	// Check if this is an index (indexes cannot be directly iterated)
	switch req.BucketName {
	case "sessions_by_room", "sessions_by_code", "refresh_tokens_by_user",
		"rooms_by_studio", "streams_by_studio", "streams_by_room",
		"codes_by_room", "codes_by_studio", "codes_by_creator",
		"memberships_by_user", "memberships_by_studio":
		return resp, errors.New("Indexes cannot be directly iterated. Use the corresponding bucket instead.")
	}

	// Use switch statement to handle different buckets
	switch req.BucketName {
	case "studios":
		vbolt.IterateAll(ctx.Tx, StudiosBkt, func(id int, studio Studio) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: studio})
			return true
		})

	case "rooms":
		vbolt.IterateAll(ctx.Tx, RoomsBkt, func(id int, room Room) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: room})
			return true
		})

	case "streams":
		vbolt.IterateAll(ctx.Tx, StreamsBkt, func(id int, stream Stream) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: stream})
			return true
		})

	case "room_stream_keys":
		vbolt.IterateAll(ctx.Tx, RoomStreamKeyBkt, func(key string, roomId int) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: key, Value: roomId})
			return true
		})

	case "room_analytics":
		vbolt.IterateAll(ctx.Tx, RoomAnalyticsBkt, func(id int, analytics RoomAnalytics) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: analytics})
			return true
		})

	case "studio_analytics":
		vbolt.IterateAll(ctx.Tx, StudioAnalyticsBkt, func(id int, analytics StudioAnalytics) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: analytics})
			return true
		})

	case "viewer_sessions":
		vbolt.IterateAll(ctx.Tx, ViewerSessionsBkt, func(key string, session ViewerSession) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: key, Value: session})
			return true
		})

	case "camera_config":
		vbolt.IterateAll(ctx.Tx, CameraConfigBkt, func(id int, config CameraConfig) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: config})
			return true
		})

	case "memberships":
		vbolt.IterateAll(ctx.Tx, MembershipBkt, func(id int, membership StudioMembership) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: membership})
			return true
		})

	case "access_codes":
		vbolt.IterateAll(ctx.Tx, AccessCodesBkt, func(code string, accessCode AccessCode) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: code, Value: accessCode})
			return true
		})

	case "code_sessions":
		vbolt.IterateAll(ctx.Tx, CodeSessionsBkt, func(token string, session CodeSession) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: token, Value: session})
			return true
		})

	case "code_analytics":
		vbolt.IterateAll(ctx.Tx, CodeAnalyticsBkt, func(code string, analytics CodeAnalytics) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: code, Value: analytics})
			return true
		})

	case "user_code_sessions":
		vbolt.IterateAll(ctx.Tx, UserCodeSessionsBkt, func(userId int, token string) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: userId, Value: token})
			return true
		})

	case "users":
		vbolt.IterateAll(ctx.Tx, UsersBkt, func(id int, user User) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: user})
			return true
		})

	case "emails":
		vbolt.IterateAll(ctx.Tx, EmailBkt, func(email string, userId int) bool {
			resp.Entries = append(resp.Entries, BucketEntry{Key: email, Value: userId})
			return true
		})

	case "refresh_tokens":
		vbolt.IterateAll(ctx.Tx, RefreshTokenBkt, func(id int, token RefreshToken) bool {
			// Mask sensitive token strings
			maskedToken := token
			if len(maskedToken.Token) > 8 {
				maskedToken.Token = maskedToken.Token[:8] + "..." + maskedToken.Token[len(maskedToken.Token)-8:]
			}
			resp.Entries = append(resp.Entries, BucketEntry{Key: id, Value: maskedToken})
			return true
		})

	case "refresh_token_lookup":
		vbolt.IterateAll(ctx.Tx, RefreshTokenByTokenBkt, func(token string, userId int) bool {
			// Mask token in key
			maskedKey := token
			if len(maskedKey) > 16 {
				maskedKey = maskedKey[:8] + "..." + maskedKey[len(maskedKey)-8:]
			}
			resp.Entries = append(resp.Entries, BucketEntry{Key: maskedKey, Value: userId})
			return true
		})

	default:
		return resp, errors.New("Unknown bucket: " + req.BucketName)
	}

	resp.Total = len(resp.Entries)

	// Log the access
	LogInfo(LogCategorySystem, "Bucket data accessed", map[string]interface{}{
		"requestedBy": caller.Id,
		"userEmail":   caller.Email,
		"bucketName":  req.BucketName,
		"entryCount":  resp.Total,
	})

	return
}

// RegisterAdminMethods registers all admin-related API procedures
func RegisterAdminMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, RecalculateViewerCounts)
	vbeam.RegisterProc(app, GetSystemLogs)
	vbeam.RegisterProc(app, GetSitePerformanceMetrics)
	vbeam.RegisterProc(app, ListBuckets)
	vbeam.RegisterProc(app, GetBucketData)
}
