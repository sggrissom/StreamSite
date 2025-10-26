package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"stream/backend/ingest"
	"strings"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// Global IngestManager instance
var ingestManager *ingest.IngestManager

// extractRoomIdFromIngestPath extracts room ID from paths like "/api/rooms/{id}/ingest/..."
func extractRoomIdFromIngestPath(path string) (roomId int, err error) {
	// Expected format: /api/rooms/{id}/ingest/{action}
	if !strings.HasPrefix(path, "/api/rooms/") {
		return 0, fmt.Errorf("invalid path: must start with /api/rooms/")
	}

	// Remove prefix
	remainder := strings.TrimPrefix(path, "/api/rooms/")

	// Find the next slash
	slashIdx := strings.Index(remainder, "/")
	if slashIdx == -1 {
		return 0, fmt.Errorf("invalid path: no room ID found")
	}

	roomIdStr := remainder[:slashIdx]
	roomId, err = strconv.Atoi(roomIdStr)
	if err != nil {
		return 0, fmt.Errorf("invalid room ID: %w", err)
	}

	return roomId, nil
}

// StartIngestHandler handles POST /api/rooms/{id}/ingest/start
func StartIngestHandler(db *vbolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract room ID from path
		roomId, err := extractRoomIdFromIngestPath(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid room ID in path", http.StatusBadRequest)
			return
		}

		// Check authentication
		authCtx, authErr := GetAuthFromRequest(r, db)
		if authErr != nil {
			LogWarnWithRequest(r, LogCategoryAuth, "Unauthorized ingest start attempt", map[string]interface{}{
				"roomId": roomId,
			})
			http.Error(w, "Authentication required", http.StatusForbidden)
			return
		}

		// Get room to verify it exists and get studio ID
		var room Room
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			room = GetRoom(tx, roomId)
		})

		if room.Id == 0 {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		// Check permissions - require Admin or higher on studio
		var hasPermission bool
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			hasPermission = HasStudioPermission(tx, authCtx.User.Id, room.StudioId, StudioRoleAdmin)
		})

		if !hasPermission {
			LogWarnWithRequest(r, LogCategoryAuth, "Permission denied for ingest start", map[string]interface{}{
				"roomId":   roomId,
				"studioId": room.StudioId,
				"userId":   authCtx.User.Id,
			})
			http.Error(w, "Only studio admins can control camera ingest", http.StatusForbidden)
			return
		}

		// Get camera configuration
		var cameraConfig CameraConfig
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			cameraConfig = GetCameraConfig(tx, roomId)
		})

		if cameraConfig.RoomId == 0 || cameraConfig.RTSPURL == "" {
			http.Error(w, "Camera RTSP URL not configured for this room", http.StatusBadRequest)
			return
		}

		// Build RTMP target URL
		rtmpOut := fmt.Sprintf("rtmp://127.0.0.1/live/%s", room.StreamKey)

		// Start the ingest process
		err = ingestManager.Start(context.Background(), roomId, cameraConfig.RTSPURL, rtmpOut)
		if err != nil {
			if strings.Contains(err.Error(), "already running") {
				http.Error(w, "Ingest already running for this room", http.StatusConflict)
			} else if strings.Contains(err.Error(), "FFmpeg binary not found") {
				http.Error(w, "FFmpeg binary not found (check FFMPEG_BIN)", http.StatusInternalServerError)
			} else {
				LogErrorSimple(LogCategorySystem, "Failed to start ingest", map[string]interface{}{
					"roomId": roomId,
					"error":  err.Error(),
				})
				http.Error(w, fmt.Sprintf("Failed to start ingest: %s", err.Error()), http.StatusInternalServerError)
			}
			return
		}

		// Log successful start
		LogInfo(LogCategorySystem, "Ingest started via API", map[string]interface{}{
			"roomId":   roomId,
			"studioId": room.StudioId,
			"userId":   authCtx.User.Id,
			"rtspURL":  cameraConfig.RTSPURL,
		})

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "started",
			"roomId": roomId,
		})
	}
}

// StopIngestHandler handles POST /api/rooms/{id}/ingest/stop
func StopIngestHandler(db *vbolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract room ID from path
		roomId, err := extractRoomIdFromIngestPath(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid room ID in path", http.StatusBadRequest)
			return
		}

		// Check authentication
		authCtx, authErr := GetAuthFromRequest(r, db)
		if authErr != nil {
			LogWarnWithRequest(r, LogCategoryAuth, "Unauthorized ingest stop attempt", map[string]interface{}{
				"roomId": roomId,
			})
			http.Error(w, "Authentication required", http.StatusForbidden)
			return
		}

		// Get room to verify it exists and get studio ID
		var room Room
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			room = GetRoom(tx, roomId)
		})

		if room.Id == 0 {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		// Check permissions - require Admin or higher on studio
		var hasPermission bool
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			hasPermission = HasStudioPermission(tx, authCtx.User.Id, room.StudioId, StudioRoleAdmin)
		})

		if !hasPermission {
			LogWarnWithRequest(r, LogCategoryAuth, "Permission denied for ingest stop", map[string]interface{}{
				"roomId":   roomId,
				"studioId": room.StudioId,
				"userId":   authCtx.User.Id,
			})
			http.Error(w, "Only studio admins can control camera ingest", http.StatusForbidden)
			return
		}

		// Stop the ingest process
		err = ingestManager.Stop(roomId)
		if err != nil {
			if strings.Contains(err.Error(), "no ingest running") {
				http.Error(w, "No ingest running for this room", http.StatusNotFound)
			} else {
				LogErrorSimple(LogCategorySystem, "Failed to stop ingest", map[string]interface{}{
					"roomId": roomId,
					"error":  err.Error(),
				})
				http.Error(w, fmt.Sprintf("Failed to stop ingest: %s", err.Error()), http.StatusInternalServerError)
			}
			return
		}

		// Log successful stop
		LogInfo(LogCategorySystem, "Ingest stopped via API", map[string]interface{}{
			"roomId":   roomId,
			"studioId": room.StudioId,
			"userId":   authCtx.User.Id,
		})

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "stopped",
			"roomId": roomId,
		})
	}
}

// IngestStatusHandler handles GET /api/rooms/{id}/ingest/status
func IngestStatusHandler(db *vbolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept GET
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract room ID from path
		roomId, err := extractRoomIdFromIngestPath(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid room ID in path", http.StatusBadRequest)
			return
		}

		// Check authentication
		authCtx, authErr := GetAuthFromRequest(r, db)
		if authErr != nil {
			LogWarnWithRequest(r, LogCategoryAuth, "Unauthorized ingest status check", map[string]interface{}{
				"roomId": roomId,
			})
			http.Error(w, "Authentication required", http.StatusForbidden)
			return
		}

		// Get room to verify it exists and get studio ID
		var room Room
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			room = GetRoom(tx, roomId)
		})

		if room.Id == 0 {
			http.Error(w, "Room not found", http.StatusNotFound)
			return
		}

		// Check permissions - require Viewer or higher on studio
		var hasPermission bool
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			hasPermission = HasStudioPermission(tx, authCtx.User.Id, room.StudioId, StudioRoleViewer)
		})

		if !hasPermission {
			LogWarnWithRequest(r, LogCategoryAuth, "Permission denied for ingest status", map[string]interface{}{
				"roomId":   roomId,
				"studioId": room.StudioId,
				"userId":   authCtx.User.Id,
			})
			http.Error(w, "Permission denied", http.StatusForbidden)
			return
		}

		// Get ingest status
		running, startTime, rtspURL := ingestManager.GetStatus(roomId)

		// Build response
		response := map[string]interface{}{
			"roomId":  roomId,
			"running": running,
		}

		if running {
			response["startedAt"] = startTime.Format("2006-01-02T15:04:05Z07:00")
			response["rtspURL"] = rtspURL
		}

		// Return status response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// RegisterIngestMethods registers HTTP handlers for camera ingest control
func RegisterIngestMethods(app *vbeam.Application) {
	// Initialize global IngestManager
	ingestManager = ingest.NewIngestManager()

	// Get database reference from app
	db := app.DB

	// Register HTTP handlers
	app.HandleFunc("/api/rooms/", func(w http.ResponseWriter, r *http.Request) {
		// Route to appropriate handler based on path suffix
		if strings.HasSuffix(r.URL.Path, "/ingest/start") {
			StartIngestHandler(db)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/ingest/stop") {
			StopIngestHandler(db)(w, r)
		} else if strings.HasSuffix(r.URL.Path, "/ingest/status") {
			IngestStatusHandler(db)(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
