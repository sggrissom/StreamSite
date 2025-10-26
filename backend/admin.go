package backend

import (
	"errors"

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

// RegisterAdminMethods registers all admin-related API procedures
func RegisterAdminMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, RecalculateViewerCounts)
}
