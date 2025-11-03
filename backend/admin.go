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

// RegisterAdminMethods registers all admin-related API procedures
func RegisterAdminMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, RecalculateViewerCounts)
	vbeam.RegisterProc(app, GetSystemLogs)
}
