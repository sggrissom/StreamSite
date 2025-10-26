package backend

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// extractRoomIdFromPath parses a path like "/streams/room/{roomId}{suffix}"
// and extracts the room ID and suffix portion.
// Returns an error if the path format is invalid.
func extractRoomIdFromPath(path string) (roomId int, suffix string, err error) {
	if !strings.HasPrefix(path, "/streams/room/") {
		return 0, "", fmt.Errorf("invalid path: must start with /streams/room/")
	}

	// Get everything after /streams/room/
	pathRemainder := strings.TrimPrefix(path, "/streams/room/")
	if pathRemainder == "" {
		return 0, "", fmt.Errorf("invalid path: no content after /streams/room/")
	}

	// Find where roomId ends (first non-digit character)
	roomIdEnd := 0
	for i, ch := range pathRemainder {
		if ch < '0' || ch > '9' {
			roomIdEnd = i
			break
		}
	}

	// If no non-digit found, the entire string is just a roomId (invalid - needs suffix)
	if roomIdEnd == 0 {
		return 0, "", fmt.Errorf("invalid path: no suffix after room ID")
	}

	roomIdStr := pathRemainder[:roomIdEnd]
	suffix = pathRemainder[roomIdEnd:]

	// Parse room ID
	roomId, err = strconv.Atoi(roomIdStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid room ID: %w", err)
	}

	return roomId, suffix, nil
}

// RegisterRoomStreamProxy sets up a proxy that maps /streams/room/{roomId}/* to the
// actual stream key path on SRS, hiding stream keys from viewers
func RegisterRoomStreamProxy(app *vbeam.Application) {
	srsURL, _ := url.Parse("http://127.0.0.1:8080")
	srsProxy := httputil.NewSingleHostReverseProxy(srsURL)

	// Capture database instance for lookups
	db := app.DB

	// Custom director to rewrite paths from room ID to stream key
	origDirector := srsProxy.Director
	srsProxy.Director = func(r *http.Request) {
		origDirector(r) // sets scheme/host to 127.0.0.1:8080

		// Extract roomId and suffix from path: /streams/room/{roomId}{suffix}
		// Examples:
		//   /streams/room/1.m3u8 -> roomId=1, suffix=.m3u8
		//   /streams/room/1-0.ts -> roomId=1, suffix=-0.ts
		path := r.URL.Path
		roomId, suffix, err := extractRoomIdFromPath(path)
		if err != nil {
			LogWarn(LogCategorySystem, "Invalid room stream path",
				"path", path, "error", err.Error())
			return
		}

		// Look up the room to get its stream key
		var room Room
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			room = GetRoom(tx, roomId)
		})

		if room.Id == 0 {
			LogWarn(LogCategorySystem, "Room not found for stream request",
				"roomId", roomId, "path", path)
			return
		}

		// Rewrite path to use stream key
		// /streams/room/{roomId}{suffix} -> /streams/live/{streamKey}{suffix}
		// Examples:
		//   /streams/room/1.m3u8 -> /streams/live/{streamKey}.m3u8
		//   /streams/room/1-0.ts -> /streams/live/{streamKey}-0.ts
		// Note: Production SRS serves at /streams/live/, local at /live/
		newPath := fmt.Sprintf("/streams/live/%s%s", room.StreamKey, suffix)
		r.URL.Path = newPath

		// Store roomId and streamKey in headers for use in ModifyResponse
		r.Header.Set("X-Room-Id", strconv.Itoa(roomId))
		r.Header.Set("X-Stream-Key", room.StreamKey)

		LogInfo(LogCategorySystem, "Proxying room stream request",
			"roomId", roomId,
			"originalPath", path,
			"newPath", newPath,
			"streamKey", room.StreamKey,
			"suffix", suffix)
	}

	// ModifyResponse rewrites URLs in m3u8 playlists to use room-based paths
	srsProxy.ModifyResponse = func(res *http.Response) error {
		p := res.Request.URL.Path
		roomId := res.Request.Header.Get("X-Room-Id")
		streamKey := res.Request.Header.Get("X-Stream-Key")

		if strings.HasSuffix(p, ".m3u8") {
			res.Header.Set("Content-Type", "application/vnd.apple.mpegurl")
			res.Header.Set("Cache-Control", "no-store, must-revalidate")
			res.Header.Set("Pragma", "no-cache")
			res.Header.Set("Expires", "0")

			// Rewrite URLs in m3u8 playlists if we have room context
			if roomId != "" && streamKey != "" && res.Body != nil {
				// Read the entire response body
				bodyBytes, err := io.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					return err
				}

				body := string(bodyBytes)

				// Rewrite absolute paths: /streams/live/{streamKey} -> /streams/room/{roomId}
				body = strings.ReplaceAll(body, "/streams/live/"+streamKey, "/streams/room/"+roomId)

				// Rewrite relative paths: {streamKey}-N.ts -> /streams/room/{roomId}-N.ts
				// Split by newlines and process each line
				lines := strings.Split(body, "\n")
				for i, line := range lines {
					trimmed := strings.TrimSpace(line)
					// If line starts with the stream key (segment reference), make it absolute
					if strings.HasPrefix(trimmed, streamKey+"-") || strings.HasPrefix(trimmed, streamKey+".") {
						lines[i] = "/streams/room/" + roomId + strings.TrimPrefix(trimmed, streamKey)
					}
				}
				body = strings.Join(lines, "\n")

				// Create new response body
				newBody := io.NopCloser(bytes.NewBufferString(body))
				res.Body = newBody
				res.ContentLength = int64(len(body))
				res.Header.Set("Content-Length", strconv.FormatInt(int64(len(body)), 10))

				LogInfo(LogCategorySystem, "Rewrote m3u8 playlist URLs",
					"roomId", roomId,
					"streamKey", streamKey,
					"path", p)
			}
		} else if strings.HasSuffix(p, ".ts") || strings.HasSuffix(p, ".m4s") {
			if res.Header.Get("Content-Type") == "" {
				res.Header.Set("Content-Type", "video/mp2t")
			}
			res.Header.Set("Cache-Control", "public, max-age=60")
		}
		return nil
	}

	// Create authenticated wrapper for the proxy
	authenticatedHandler := func(w http.ResponseWriter, r *http.Request) {
		// Extract roomId from path for authentication check
		roomId, _, err := extractRoomIdFromPath(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid stream path", http.StatusBadRequest)
			return
		}

		// Check authentication
		authCtx, authErr := GetAuthFromRequest(r, db)
		if authErr != nil {
			LogWarnWithRequest(r, LogCategoryAuth, "Unauthorized stream access attempt", map[string]interface{}{
				"roomId": roomId,
				"path":   r.URL.Path,
			})
			http.Error(w, "Authentication required", http.StatusForbidden)
			return
		}

		// Verify user has access to this room using unified access control
		var anonymousSessionToken string
		if authCtx.User.Id == -1 && authCtx.CodeSession != nil {
			anonymousSessionToken = authCtx.CodeSession.Token
		}

		var hasAccess bool
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			access := CheckRoomAccess(tx, authCtx.User, roomId, anonymousSessionToken)
			hasAccess = access.Allowed
		})

		if !hasAccess {
			LogWarnWithRequest(r, LogCategoryAuth, "Access denied to room", map[string]interface{}{
				"roomId": roomId,
				"userId": authCtx.User.Id,
			})
			http.Error(w, "Access denied to this room", http.StatusForbidden)
			return
		}

		// Log successful access
		logData := map[string]interface{}{
			"roomId": roomId,
			"userId": authCtx.User.Id,
		}
		if authCtx.IsCodeAuth && authCtx.AccessCode != nil {
			logData["code"] = authCtx.AccessCode.Code
			LogInfo(LogCategoryAuth, "Code session accessing stream", logData)
		} else {
			logData["email"] = authCtx.User.Email
			LogInfo(LogCategoryAuth, "User accessing stream", logData)
		}

		// Authentication successful - proxy the request
		srsProxy.ServeHTTP(w, r)
	}

	// Register the authenticated handler for room-based streams
	app.HandleFunc("/streams/room/", authenticatedHandler)
}
