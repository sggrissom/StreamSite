package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// SSEClient represents a single SSE connection
type SSEClient struct {
	RoomID int
	Writer http.ResponseWriter
	Done   chan bool
}

// SSEManager manages all active SSE connections
type SSEManager struct {
	mu sync.RWMutex
	// Map: roomId -> list of clients watching that room
	clients map[int][]*SSEClient
}

// Global SSE manager instance
var sseManager = &SSEManager{
	clients: make(map[int][]*SSEClient),
}

// AddClient adds a client to a room's subscriber list
func (m *SSEManager) AddClient(roomID int, client *SSEClient) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[roomID] == nil {
		m.clients[roomID] = []*SSEClient{}
	}
	m.clients[roomID] = append(m.clients[roomID], client)

	LogDebug(LogCategorySystem, "SSE client connected", map[string]interface{}{
		"roomId": roomID,
		"total":  len(m.clients[roomID]),
	})
}

// RemoveClient removes a client from a room's subscriber list
func (m *SSEManager) RemoveClient(roomID int, client *SSEClient) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.clients[roomID]
	for i, c := range clients {
		if c == client {
			m.clients[roomID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	// Cleanup empty room lists
	if len(m.clients[roomID]) == 0 {
		delete(m.clients, roomID)
	}

	LogDebug(LogCategorySystem, "SSE client disconnected", map[string]interface{}{
		"roomId":    roomID,
		"remaining": len(m.clients[roomID]),
	})
}

// flushSSE flushes an SSE response using the best available method
func flushSSE(w http.ResponseWriter) {
	// Try ResponseController first (works with wrapped writers)
	rc := http.NewResponseController(w)
	if rc.Flush() == nil {
		return
	}

	// Fallback to direct Flusher interface
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// BroadcastRoomStatus sends a status update to all clients watching a room
func (m *SSEManager) BroadcastRoomStatus(roomID int, isActive bool) {
	m.mu.RLock()
	clients := m.clients[roomID]
	m.mu.RUnlock()

	if len(clients) == 0 {
		return // No one watching
	}

	event := map[string]interface{}{
		"isActive":  isActive,
		"timestamp": time.Now().Unix(),
	}

	data, _ := json.Marshal(event)
	message := fmt.Sprintf("event: status\ndata: %s\n\n", data)

	LogDebug(LogCategorySystem, "Broadcasting SSE update", map[string]interface{}{
		"roomId":   roomID,
		"isActive": isActive,
		"clients":  len(clients),
	})

	// Send to all connected clients
	for _, client := range clients {
		select {
		case <-client.Done:
			// Client already disconnected
			continue
		default:
			if _, err := fmt.Fprint(client.Writer, message); err == nil {
				flushSSE(client.Writer)
			}
		}
	}
}

// MakeStreamRoomEventsHandler creates an HTTP handler for SSE connections
// Note: This is a special handler that doesn't follow the normal vbeam RPC pattern
// because SSE requires keeping the connection open
func MakeStreamRoomEventsHandler(db *vbolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get room ID from query params
		roomIDStr := r.URL.Query().Get("roomId")
		if roomIDStr == "" {
			http.Error(w, "roomId parameter required", http.StatusBadRequest)
			return
		}

		var roomID int
		_, err := fmt.Sscanf(roomIDStr, "%d", &roomID)
		if err != nil {
			http.Error(w, "invalid roomId", http.StatusBadRequest)
			return
		}

		// Validate room exists
		var room Room
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			room = GetRoom(tx, roomID)
		})

		if room.Id == 0 {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}

		// Authenticate user (supports both JWT and code sessions)
		authCtx, authErr := GetAuthFromRequest(r, db)
		if authErr != nil {
			LogWarnWithRequest(r, LogCategoryAuth, "Unauthorized SSE connection attempt", map[string]interface{}{
				"roomId": roomID,
			})
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Check permissions based on authentication type
		var hasPermission bool

		if authCtx.IsCodeAuth {
			// For code-based auth, verify room access
			vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
				ctx := &vbeam.Context{Tx: tx}
				accessResp, _ := GetCodeStreamAccess(ctx, GetCodeStreamAccessRequest{
					SessionToken: authCtx.CodeSession.Token,
					RoomId:       roomID,
				})
				hasPermission = accessResp.Allowed
			})

			if !hasPermission {
				LogWarnWithRequest(r, LogCategoryAuth, "Code session denied SSE access to room", map[string]interface{}{
					"roomId":       roomID,
					"code":         authCtx.AccessCode.Code,
					"sessionToken": authCtx.CodeSession.Token,
				})
				http.Error(w, "Access denied to this room", http.StatusForbidden)
				return
			}

			LogInfo(LogCategoryAuth, "Code session connected to SSE", map[string]interface{}{
				"roomId": roomID,
				"code":   authCtx.AccessCode.Code,
			})
		} else {
			// For JWT auth, check studio permissions
			vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
				hasPermission = HasStudioPermission(tx, authCtx.User.Id, room.StudioId, StudioRoleViewer)
			})

			if !hasPermission {
				LogWarnWithRequest(r, LogCategoryAuth, "User denied SSE access to room", map[string]interface{}{
					"roomId": roomID,
					"userId": authCtx.User.Id,
					"email":  authCtx.User.Email,
				})
				http.Error(w, "You do not have permission to view this room", http.StatusForbidden)
				return
			}

			LogInfo(LogCategoryAuth, "User connected to SSE", map[string]interface{}{
				"roomId": roomID,
				"userId": authCtx.User.Id,
				"email":  authCtx.User.Email,
			})
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Create SSE client
		client := &SSEClient{
			RoomID: roomID,
			Writer: w,
			Done:   make(chan bool),
		}

		// Register client
		sseManager.AddClient(roomID, client)
		defer sseManager.RemoveClient(roomID, client)

		// Send initial status immediately
		initialEvent := map[string]interface{}{
			"isActive":  room.IsActive,
			"timestamp": time.Now().Unix(),
		}
		initialData, _ := json.Marshal(initialEvent)
		fmt.Fprintf(w, "event: status\ndata: %s\n\n", initialData)
		flushSSE(w)

		// Keep connection alive with periodic pings
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// Wait for disconnect
		notify := r.Context().Done()
		for {
			select {
			case <-notify:
				// Client disconnected
				close(client.Done)
				return
			case <-ticker.C:
				// Send keepalive comment
				fmt.Fprintf(w, ": keepalive\n\n")
				flushSSE(w)
			}
		}
	}
}
