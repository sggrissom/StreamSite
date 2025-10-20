package backend

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"stream/cfg"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// CodeType defines whether the code grants access to a room or entire studio
type CodeType int

const (
	CodeTypeRoom   CodeType = 0 // Access to specific room only
	CodeTypeStudio CodeType = 1 // Access to all rooms in studio
)

// AccessCode represents a temporary access code for anonymous stream viewing
type AccessCode struct {
	Code       string    `json:"code"`       // 5-digit code (e.g., "42857")
	Type       CodeType  `json:"type"`       // 0=room, 1=studio
	TargetId   int       `json:"targetId"`   // Room ID or Studio ID
	CreatedBy  int       `json:"createdBy"`  // User ID who created the code
	CreatedAt  time.Time `json:"createdAt"`  // When code was created
	ExpiresAt  time.Time `json:"expiresAt"`  // Hard expiration time
	MaxViewers int       `json:"maxViewers"` // 0=unlimited, >0=max concurrent
	IsRevoked  bool      `json:"isRevoked"`  // Manual revocation flag
	Label      string    `json:"label"`      // Optional description (e.g., "Physics 101 - Oct 18")
}

// CodeSession represents an active viewing session using an access code
type CodeSession struct {
	Token            string    `json:"token"`                      // Session UUID
	Code             string    `json:"code"`                       // Which code was used
	ConnectedAt      time.Time `json:"connectedAt"`                // Initial connection time
	LastSeen         time.Time `json:"lastSeen"`                   // Last activity (for timeout detection)
	GracePeriodUntil time.Time `json:"gracePeriodUntil,omitempty"` // Set when code expires (zero if not in grace period)
	ClientIP         string    `json:"clientIP"`                   // For analytics/rate limiting
	UserAgent        string    `json:"userAgent"`                  // For analytics
}

// CodeAnalytics tracks usage statistics for an access code
type CodeAnalytics struct {
	Code             string    `json:"code"`             // The access code
	TotalConnections int       `json:"totalConnections"` // Lifetime connection count
	CurrentViewers   int       `json:"currentViewers"`   // Active sessions right now
	PeakViewers      int       `json:"peakViewers"`      // Historical maximum
	PeakViewersAt    time.Time `json:"peakViewersAt"`    // When peak occurred
	LastConnectionAt time.Time `json:"lastConnectionAt"` // Most recent connection
}

// Packing functions for vbolt serialization

func PackAccessCode(self *AccessCode, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.String(&self.Code, buf)
	vpack.Int((*int)(&self.Type), buf)
	vpack.Int(&self.TargetId, buf)
	vpack.Int(&self.CreatedBy, buf)
	vpack.Time(&self.CreatedAt, buf)
	vpack.Time(&self.ExpiresAt, buf)
	vpack.Int(&self.MaxViewers, buf)
	vpack.Bool(&self.IsRevoked, buf)
	vpack.String(&self.Label, buf)
}

func PackCodeSession(self *CodeSession, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.String(&self.Token, buf)
	vpack.String(&self.Code, buf)
	vpack.Time(&self.ConnectedAt, buf)
	vpack.Time(&self.LastSeen, buf)
	vpack.Time(&self.GracePeriodUntil, buf)
	vpack.String(&self.ClientIP, buf)
	vpack.String(&self.UserAgent, buf)
}

func PackCodeAnalytics(self *CodeAnalytics, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.String(&self.Code, buf)
	vpack.Int(&self.TotalConnections, buf)
	vpack.Int(&self.CurrentViewers, buf)
	vpack.Int(&self.PeakViewers, buf)
	vpack.Time(&self.PeakViewersAt, buf)
	vpack.Time(&self.LastConnectionAt, buf)
}

// Buckets for entity storage

// AccessCodesBkt: code (string) -> AccessCode
var AccessCodesBkt = vbolt.Bucket(&cfg.Info, "access_codes", vpack.StringZ, PackAccessCode)

// CodeSessionsBkt: sessionToken (string) -> CodeSession
var CodeSessionsBkt = vbolt.Bucket(&cfg.Info, "code_sessions", vpack.StringZ, PackCodeSession)

// CodeAnalyticsBkt: code (string) -> CodeAnalytics
var CodeAnalyticsBkt = vbolt.Bucket(&cfg.Info, "code_analytics", vpack.StringZ, PackCodeAnalytics)

// Indexes for relationship queries

// CodesByRoomIdx: roomId (term) -> code (target)
// Find all codes for a specific room
var CodesByRoomIdx = vbolt.Index(&cfg.Info, "codes_by_room", vpack.FInt, vpack.StringZ)

// CodesByStudioIdx: studioId (term) -> code (target)
// Find all codes for a specific studio
var CodesByStudioIdx = vbolt.Index(&cfg.Info, "codes_by_studio", vpack.FInt, vpack.StringZ)

// CodesByCreatorIdx: userId (term) -> code (target)
// Find all codes created by a user
var CodesByCreatorIdx = vbolt.Index(&cfg.Info, "codes_by_creator", vpack.FInt, vpack.StringZ)

// SessionsByCodeIdx: code (term) -> sessionToken (target)
// Find all active sessions for a code
var SessionsByCodeIdx = vbolt.Index(&cfg.Info, "sessions_by_code", vpack.StringZ, vpack.StringZ)

// Code generation utilities

// GenerateUniqueCode creates a random 5-digit code avoiding common patterns
func GenerateUniqueCode() (string, error) {
	// Generate random 5-digit number (10000-99999)
	// Avoid patterns like 11111, 12345, etc.

	for attempt := 0; attempt < 10; attempt++ {
		// Generate number between 10000 and 99999
		n, err := rand.Int(rand.Reader, big.NewInt(90000))
		if err != nil {
			return "", err
		}
		code := int(n.Int64()) + 10000
		codeStr := fmt.Sprintf("%05d", code)

		// Check for bad patterns
		if isBadPattern(codeStr) {
			continue
		}

		return codeStr, nil
	}

	return "", fmt.Errorf("failed to generate unique code after 10 attempts")
}

// isBadPattern checks if a code uses easily guessable patterns
func isBadPattern(code string) bool {
	// Check for all same digit (11111, 22222, etc.)
	allSame := true
	for i := 1; i < len(code); i++ {
		if code[i] != code[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return true
	}

	// Check for sequential ascending (12345, 23456, etc.)
	sequential := true
	for i := 1; i < len(code); i++ {
		if code[i] != code[i-1]+1 {
			sequential = false
			break
		}
	}
	if sequential {
		return true
	}

	// Check for sequential descending (54321, 43210, etc.)
	descending := true
	for i := 1; i < len(code); i++ {
		if code[i] != code[i-1]-1 {
			descending = false
			break
		}
	}
	if descending {
		return true
	}

	return false
}

// Request/Response types for API procedures

type GenerateAccessCodeRequest struct {
	Type            int    `json:"type"`            // 0=room, 1=studio
	TargetId        int    `json:"targetId"`        // Room ID or Studio ID
	DurationMinutes int    `json:"durationMinutes"` // How long code is valid
	MaxViewers      int    `json:"maxViewers"`      // 0=unlimited
	Label           string `json:"label"`           // Optional description
}

type GenerateAccessCodeResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Code      string    `json:"code,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
	ShareURL  string    `json:"shareUrl,omitempty"` // e.g., "/watch/42857"
}

type ValidateAccessCodeRequest struct {
	Code string `json:"code"` // 5-digit code
}

type ValidateAccessCodeResponse struct {
	Success      bool      `json:"success"`
	Error        string    `json:"error,omitempty"`
	SessionToken string    `json:"sessionToken,omitempty"` // UUID for the session
	RedirectTo   string    `json:"redirectTo,omitempty"`   // URL to redirect to (e.g., "/stream/123")
	ExpiresAt    time.Time `json:"expiresAt,omitempty"`    // When code expires
	Type         int       `json:"type,omitempty"`         // 0=room, 1=studio
	TargetId     int       `json:"targetId,omitempty"`     // Room or Studio ID
}

type GetCodeStreamAccessRequest struct {
	SessionToken string `json:"sessionToken"` // Session token from ValidateAccessCode
	RoomId       int    `json:"roomId"`       // Which room trying to access
}

type GetCodeStreamAccessResponse struct {
	Allowed     bool      `json:"allowed"`               // Whether access is granted
	RoomId      int       `json:"roomId,omitempty"`      // Room ID (echoed back)
	StudioId    int       `json:"studioId,omitempty"`    // Studio ID (for context)
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`   // When code expires
	GracePeriod bool      `json:"gracePeriod,omitempty"` // Whether in grace period
	Message     string    `json:"message,omitempty"`     // Human-readable status
}

type RevokeAccessCodeRequest struct {
	Code string `json:"code"` // 5-digit code to revoke
}

type RevokeAccessCodeResponse struct {
	Success        bool   `json:"success"`
	Error          string `json:"error,omitempty"`
	SessionsKilled int    `json:"sessionsKilled,omitempty"` // Number of active sessions terminated
}

type ListAccessCodesRequest struct {
	Type     int `json:"type"`     // 0=room, 1=studio
	TargetId int `json:"targetId"` // Room ID or Studio ID
}

type AccessCodeListItem struct {
	Code           string    `json:"code"`
	Type           int       `json:"type"`
	Label          string    `json:"label"`
	CreatedAt      time.Time `json:"createdAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
	IsRevoked      bool      `json:"isRevoked"`
	IsExpired      bool      `json:"isExpired"`
	CurrentViewers int       `json:"currentViewers"`
	TotalViews     int       `json:"totalViews"`
}

type ListAccessCodesResponse struct {
	Success bool                 `json:"success"`
	Error   string               `json:"error,omitempty"`
	Codes   []AccessCodeListItem `json:"codes,omitempty"`
}

type GetCodeAnalyticsRequest struct {
	Code string `json:"code"` // 5-digit code
}

type SessionInfo struct {
	ConnectedAt time.Time `json:"connectedAt"` // When session started
	Duration    int       `json:"duration"`    // Seconds since connection (or until LastSeen if inactive)
	ClientIP    string    `json:"clientIP"`    // Anonymized IP address
	UserAgent   string    `json:"userAgent"`   // Browser/device info
	IsActive    bool      `json:"isActive"`    // Whether session is still active (LastSeen < 10min ago)
}

type GetCodeAnalyticsResponse struct {
	Success          bool          `json:"success"`
	Error            string        `json:"error,omitempty"`
	Code             string        `json:"code,omitempty"`
	Type             int           `json:"type,omitempty"`             // 0=room, 1=studio
	Label            string        `json:"label,omitempty"`            // Description
	Status           string        `json:"status,omitempty"`           // "active", "expired", "revoked"
	CreatedAt        time.Time     `json:"createdAt,omitempty"`        // When code was created
	ExpiresAt        time.Time     `json:"expiresAt,omitempty"`        // When code expires
	TotalConnections int           `json:"totalConnections,omitempty"` // Lifetime connection count
	CurrentViewers   int           `json:"currentViewers,omitempty"`   // Active sessions right now
	PeakViewers      int           `json:"peakViewers,omitempty"`      // Historical maximum
	PeakViewersAt    time.Time     `json:"peakViewersAt,omitempty"`    // When peak occurred
	Sessions         []SessionInfo `json:"sessions,omitempty"`         // Current active sessions
}

// Helper functions

// codeExistsInDB checks if a code already exists in the database
func codeExistsInDB(tx *vbolt.Tx, code string) bool {
	var existing AccessCode
	vbolt.Read(tx, AccessCodesBkt, code, &existing)
	return existing.Code != ""
}

// generateUniqueCodeInDB generates a code and ensures it's unique in the database
func generateUniqueCodeInDB(tx *vbolt.Tx) (string, error) {
	for attempt := 0; attempt < 20; attempt++ {
		code, err := GenerateUniqueCode()
		if err != nil {
			return "", err
		}

		// Check if code already exists in database
		if !codeExistsInDB(tx, code) {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique code after 20 attempts")
}

// generateSessionToken creates a cryptographically secure random token for code sessions
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateCodeSession validates a code session token and returns session + code info
// Returns: isValid, session, code, errorMessage
func ValidateCodeSession(tx *vbolt.Tx, sessionToken string) (bool, CodeSession, AccessCode, string) {
	var session CodeSession
	var code AccessCode

	// Load session
	vbolt.Read(tx, CodeSessionsBkt, sessionToken, &session)
	if session.Token == "" {
		return false, session, code, "Session not found"
	}

	// Load associated code
	vbolt.Read(tx, AccessCodesBkt, session.Code, &code)
	if code.Code == "" {
		return false, session, code, "Access code not found"
	}

	// Check if code is revoked
	if code.IsRevoked {
		return false, session, code, "Access code has been revoked"
	}

	// Check expiration with grace period logic
	now := time.Now()
	if now.After(code.ExpiresAt) {
		// Code has expired - check grace period
		if !session.GracePeriodUntil.IsZero() && now.Before(session.GracePeriodUntil) {
			// Still within grace period - allow access
			return true, session, code, ""
		} else if session.GracePeriodUntil.IsZero() {
			// Code just expired, no grace period set yet
			// This means we need to grant grace period (will be set by caller if they want to)
			return false, session, code, "Access code has expired (grace period available)"
		} else {
			// Grace period has ended
			return false, session, code, "Access code has expired"
		}
	}

	// Code is valid and active
	return true, session, code, ""
}

// DecrementCodeViewerCount decrements the CurrentViewers count for a code session
// This should be called when SSE clients disconnect to accurately track active viewers
func DecrementCodeViewerCount(db *vbolt.DB, sessionToken string) error {
	if sessionToken == "" {
		return nil // Not a code session, nothing to decrement
	}

	var codeStr string

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Load the session to get the associated code
		var session CodeSession
		vbolt.Read(tx, CodeSessionsBkt, sessionToken, &session)

		if session.Token == "" {
			// Session not found, nothing to decrement
			return
		}

		codeStr = session.Code

		// Load the analytics for this code
		var analytics CodeAnalytics
		vbolt.Read(tx, CodeAnalyticsBkt, session.Code, &analytics)

		// Decrement CurrentViewers (with bounds check)
		if analytics.CurrentViewers > 0 {
			analytics.CurrentViewers--
			vbolt.Write(tx, CodeAnalyticsBkt, session.Code, &analytics)

			LogDebug(LogCategorySystem, "Decremented viewer count for code session", map[string]interface{}{
				"code":           session.Code,
				"currentViewers": analytics.CurrentViewers,
				"sessionToken":   sessionToken,
			})
		}

		vbolt.TxCommit(tx)
	})

	if codeStr != "" {
		LogInfo(LogCategorySystem, "Code session disconnected", map[string]interface{}{
			"code":         codeStr,
			"sessionToken": sessionToken,
		})
	}

	return nil
}

// CleanupInactiveSessions removes code sessions that haven't been active
// for more than 10 minutes. This prevents zombie sessions from keeping
// viewer counts inflated and database bloat.
// Returns the number of sessions cleaned up.
//
// Note: This iterates through all codes in the AccessCodesBkt. For better
// performance in production, consider adding an index of all codes or
// tracking recently active codes.
func CleanupInactiveSessions(db *vbolt.DB) int {
	cleanedCount := 0
	cutoffTime := time.Now().Add(-10 * time.Minute)

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Collect all unique codes that have sessions
		// We do this by reading all sessions from the SessionsByCodeIdx
		codesWithSessions := make(map[string]bool)

		// Iterate through all access codes and check their sessions
		vbolt.IterateAll(tx, AccessCodesBkt, func(code string, accessCode AccessCode) bool {
			codesWithSessions[code] = true
			return true
		})

		// For each code that has sessions, check for inactive ones
		for code := range codesWithSessions {
			var sessionTokens []string
			vbolt.ReadTermTargets(tx, SessionsByCodeIdx, code, &sessionTokens, vbolt.Window{})

			for _, token := range sessionTokens {
				var session CodeSession
				vbolt.Read(tx, CodeSessionsBkt, token, &session)

				if session.Token == "" {
					continue
				}

				// Check if session is inactive (LastSeen > 10 minutes ago)
				if session.LastSeen.Before(cutoffTime) {
					// Decrement viewer count for this session's code
					var analytics CodeAnalytics
					vbolt.Read(tx, CodeAnalyticsBkt, session.Code, &analytics)

					if analytics.CurrentViewers > 0 {
						analytics.CurrentViewers--
						vbolt.Write(tx, CodeAnalyticsBkt, session.Code, &analytics)
					}

					// Remove session from index
					vbolt.SetTargetSingleTerm(tx, SessionsByCodeIdx, session.Token, "")

					// Delete the session
					vbolt.Delete(tx, CodeSessionsBkt, session.Token)

					cleanedCount++

					LogDebug(LogCategorySystem, "Cleaned up inactive code session", map[string]interface{}{
						"sessionToken": token,
						"code":         session.Code,
						"lastSeen":     session.LastSeen,
						"inactiveFor":  time.Since(session.LastSeen).String(),
					})
				}
			}
		}

		vbolt.TxCommit(tx)
	})

	if cleanedCount > 0 {
		LogInfo(LogCategorySystem, "Session cleanup completed", map[string]interface{}{
			"sessionsRemoved": cleanedCount,
		})
	}

	return cleanedCount
}

// StartCodeSessionCleanup starts a background goroutine that periodically
// cleans up inactive code sessions every 5 minutes
func StartCodeSessionCleanup(db *vbolt.DB) {
	LogInfo(LogCategorySystem, "Starting code session cleanup job", map[string]interface{}{
		"frequency":       "5 minutes",
		"inactivityLimit": "10 minutes",
	})

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			CleanupInactiveSessions(db)
		}
	}()
}

// CleanupOldAccessCodes removes access codes that have been expired for more than
// the specified retention period. This prevents database bloat from accumulating
// old codes and their associated data.
//
// Returns the number of codes cleaned up.
//
// For each old code, this function:
//   - Deletes the code entry from AccessCodesBkt
//   - Deletes analytics from CodeAnalyticsBkt
//   - Deletes any remaining sessions from CodeSessionsBkt
//   - Removes the code from all indexes (room, studio, creator)
func CleanupOldAccessCodes(db *vbolt.DB, retentionDays int) int {
	cleanedCount := 0
	cutoffTime := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Iterate through all codes to find ones to delete
		var codesToDelete []string

		vbolt.IterateAll(tx, AccessCodesBkt, func(codeStr string, code AccessCode) bool {
			// Check if code is expired and beyond retention period
			if code.ExpiresAt.Before(cutoffTime) {
				codesToDelete = append(codesToDelete, codeStr)
			}
			return true
		})

		// Delete each old code and its associated data
		for _, codeStr := range codesToDelete {
			// Load the code to get its metadata for index removal
			var code AccessCode
			vbolt.Read(tx, AccessCodesBkt, codeStr, &code)
			if code.Code == "" {
				continue // Code already deleted
			}

			// Delete code from bucket
			vbolt.Delete(tx, AccessCodesBkt, codeStr)

			// Delete analytics
			vbolt.Delete(tx, CodeAnalyticsBkt, codeStr)

			// Delete any remaining sessions for this code
			// Use the SessionsByCodeIdx to find sessions efficiently
			var sessionTokens []string
			vbolt.ReadTermTargets(tx, SessionsByCodeIdx, codeStr, &sessionTokens, vbolt.Window{})
			for _, token := range sessionTokens {
				vbolt.Delete(tx, CodeSessionsBkt, token)
				vbolt.SetTargetSingleTerm(tx, SessionsByCodeIdx, token, "")
			}

			// Remove from indexes
			if code.Type == CodeTypeRoom {
				vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, codeStr, -1)
			} else {
				vbolt.SetTargetSingleTerm(tx, CodesByStudioIdx, codeStr, -1)
			}
			vbolt.SetTargetSingleTerm(tx, CodesByCreatorIdx, codeStr, -1)

			cleanedCount++
		}

		if cleanedCount > 0 {
			LogInfo(LogCategorySystem, "Cleaned up old access codes", map[string]interface{}{
				"codesDeleted":  cleanedCount,
				"retentionDays": retentionDays,
				"cutoffDate":    cutoffTime.Format("2006-01-02"),
			})
		}
	})

	return cleanedCount
}

// StartOldCodeCleanup starts a background goroutine that periodically
// cleans up access codes that have been expired for more than 7 days
func StartOldCodeCleanup(db *vbolt.DB) {
	const retentionDays = 7

	LogInfo(LogCategorySystem, "Starting old code cleanup job", map[string]interface{}{
		"frequency":     "daily",
		"retentionDays": retentionDays,
	})

	go func() {
		// Run immediately on startup
		CleanupOldAccessCodes(db, retentionDays)

		// Then run daily
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			CleanupOldAccessCodes(db, retentionDays)
		}
	}()
}

// HandleExpiredCodes handles access codes that have expired, setting grace periods
// for active sessions and disconnecting sessions whose grace period has ended.
// Returns the number of sessions affected (grace period set + disconnected).
//
// This function:
//  1. Finds codes where ExpiresAt <= now
//  2. For active sessions without grace period: sets GracePeriodUntil = now + 15min
//  3. Broadcasts CODE_EXPIRED_GRACE_PERIOD event to affected rooms
//  4. Finds sessions where GracePeriodUntil <= now (grace period ended)
//  5. Disconnects those sessions (deletes from DB, decrements viewer count)
//
// Note: Grace period is per-session, not per-code, to handle cases where
// a code expires while multiple people are watching.
func HandleExpiredCodes(db *vbolt.DB) int {
	affectedSessions := 0
	now := time.Now()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Track codes that need SSE broadcasts (map: code -> roomIds)
		codesToBroadcast := make(map[string][]int)

		// Iterate through all access codes to find expired ones
		vbolt.IterateAll(tx, AccessCodesBkt, func(codeStr string, code AccessCode) bool {
			// Skip if code is non-expired or revoked
			if code.ExpiresAt.After(now) || code.IsRevoked {
				return true
			}

			// Find all sessions for this expired code
			var sessionTokens []string
			vbolt.ReadTermTargets(tx, SessionsByCodeIdx, code.Code, &sessionTokens, vbolt.Window{})

			for _, token := range sessionTokens {
				var session CodeSession
				vbolt.Read(tx, CodeSessionsBkt, token, &session)

				if session.Token == "" {
					continue
				}

				// Case 1: Session doesn't have grace period yet - set it now
				if session.GracePeriodUntil.IsZero() {
					session.GracePeriodUntil = now.Add(15 * time.Minute)
					vbolt.Write(tx, CodeSessionsBkt, session.Token, &session)
					affectedSessions++

					// Track this code for SSE broadcast
					if code.Type == CodeTypeRoom {
						codesToBroadcast[code.Code] = []int{code.TargetId}
					} else {
						// Studio code - need to get all rooms
						var roomIds []int
						vbolt.ReadTermTargets(tx, RoomsByStudioIdx, code.TargetId, &roomIds, vbolt.Window{})
						codesToBroadcast[code.Code] = roomIds
					}

					LogDebug(LogCategorySystem, "Set grace period for expired code session", map[string]interface{}{
						"sessionToken":       token,
						"code":               code.Code,
						"gracePeriodUntil":   session.GracePeriodUntil,
						"gracePeriodMinutes": 15,
					})
				}

				// Case 2: Grace period has ended - disconnect session
				if !session.GracePeriodUntil.IsZero() && session.GracePeriodUntil.Before(now) {
					// Decrement viewer count
					var analytics CodeAnalytics
					vbolt.Read(tx, CodeAnalyticsBkt, session.Code, &analytics)
					if analytics.CurrentViewers > 0 {
						analytics.CurrentViewers--
						vbolt.Write(tx, CodeAnalyticsBkt, session.Code, &analytics)
					}

					// Remove session from index
					vbolt.SetTargetSingleTerm(tx, SessionsByCodeIdx, session.Token, "")

					// Delete the session
					vbolt.Delete(tx, CodeSessionsBkt, session.Token)

					affectedSessions++

					LogDebug(LogCategorySystem, "Disconnected session after grace period ended", map[string]interface{}{
						"sessionToken":     token,
						"code":             code.Code,
						"gracePeriodEnded": session.GracePeriodUntil,
					})
				}
			}

			return true
		})

		vbolt.TxCommit(tx)

		// Broadcast SSE events after committing transaction
		for code, roomIds := range codesToBroadcast {
			for _, roomId := range roomIds {
				sseManager.BroadcastCodeExpiredGracePeriod(roomId, 15)
				LogDebug(LogCategorySystem, "Broadcasted grace period event", map[string]interface{}{
					"code":   code,
					"roomId": roomId,
				})
			}
		}
	})

	if affectedSessions > 0 {
		LogInfo(LogCategorySystem, "Expired code handling completed", map[string]interface{}{
			"affectedSessions": affectedSessions,
		})
	}

	return affectedSessions
}

// StartExpiredCodeHandler starts a background goroutine that periodically
// handles expired access codes and grace periods every 1 minute
func StartExpiredCodeHandler(db *vbolt.DB) {
	LogInfo(LogCategorySystem, "Starting expired code handler job", map[string]interface{}{
		"frequency":   "1 minute",
		"gracePeriod": "15 minutes",
	})

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			HandleExpiredCodes(db)
		}
	}()
}

// API Procedures

// validateAccessCodeHandler is an HTTP handler that validates an access code
// and sets the authToken cookie with a JWT for authentication
func validateAccessCodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		vbeam.RespondError(w, errors.New("validate-access-code must be POST"))
		return
	}

	var req ValidateAccessCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		vbeam.RespondError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Call validation logic
	resp, err := validateAccessCodeLogic(appDb, req.Code)
	if err != nil {
		vbeam.RespondError(w, err)
		return
	}

	// If validation succeeded, generate JWT and set authToken cookie
	if resp.Success {
		// Generate JWT with code session claims
		expirationTime := resp.ExpiresAt
		claims := &Claims{
			IsCodeSession: true,
			SessionToken:  resp.SessionToken,
			Code:          req.Code,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, tokenErr := token.SignedString(jwtKey)
		if tokenErr != nil {
			LogErrorWithRequest(r, LogCategorySystem, "Failed to generate JWT for code session", map[string]interface{}{
				"code":         req.Code,
				"sessionToken": resp.SessionToken,
				"error":        tokenErr.Error(),
			})
			vbeam.RespondError(w, errors.New("failed to generate session token"))
			return
		}

		// Set authToken cookie (same as regular user auth)
		http.SetCookie(w, &http.Cookie{
			Name:     "authToken",
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   60 * 60 * 24, // 24 hours
		})
	}

	// Send JSON response
	json.NewEncoder(w).Encode(resp)
}

func RegisterCodeAccessMethods(app *vbeam.Application) {
	// Register HTTP handler for code validation (needs cookie setting)
	app.HandleFunc("/api/validate-access-code", validateAccessCodeHandler)

	// Register vbeam procedures
	vbeam.RegisterProc(app, GenerateAccessCode)
	vbeam.RegisterProc(app, GetCodeStreamAccess)
	vbeam.RegisterProc(app, RevokeAccessCode)
	vbeam.RegisterProc(app, ListAccessCodes)
	vbeam.RegisterProc(app, GetCodeAnalytics)
}

// GenerateAccessCode creates a new temporary access code for room or studio viewing
func GenerateAccessCode(ctx *vbeam.Context, req GenerateAccessCodeRequest) (resp GenerateAccessCodeResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Validate code type
	if req.Type != int(CodeTypeRoom) && req.Type != int(CodeTypeStudio) {
		resp.Success = false
		resp.Error = "Invalid code type (must be 0 for room or 1 for studio)"
		return
	}

	// Validate duration
	if req.DurationMinutes <= 0 {
		resp.Success = false
		resp.Error = "Duration must be greater than 0"
		return
	}

	// Validate max viewers
	if req.MaxViewers < 0 {
		resp.Success = false
		resp.Error = "Max viewers cannot be negative"
		return
	}

	// Validate label length
	if len(req.Label) > 200 {
		resp.Success = false
		resp.Error = "Label is too long (max 200 characters)"
		return
	}

	var studioId int
	var targetName string

	// Validate target and check permissions based on type
	if req.Type == int(CodeTypeRoom) {
		// Room code - validate room exists and check permission
		room := GetRoom(ctx.Tx, req.TargetId)
		if room.Id == 0 {
			resp.Success = false
			resp.Error = "Room not found"
			return
		}

		studioId = room.StudioId
		targetName = room.Name

	} else {
		// Studio code - validate studio exists and check permission
		studio := GetStudioById(ctx.Tx, req.TargetId)
		if studio.Id == 0 {
			resp.Success = false
			resp.Error = "Studio not found"
			return
		}

		studioId = studio.Id
		targetName = studio.Name

	}

	// Check if user has Admin+ permission for the studio
	if !HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can generate access codes"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Generate unique code
	code, err := generateUniqueCodeInDB(ctx.Tx)
	if err != nil {
		resp.Success = false
		resp.Error = "Failed to generate unique code"
		return
	}

	// Calculate expiration time
	now := time.Now()
	expiresAt := now.Add(time.Duration(req.DurationMinutes) * time.Minute)

	// Create access code
	accessCode := AccessCode{
		Code:       code,
		Type:       CodeType(req.Type),
		TargetId:   req.TargetId,
		CreatedBy:  caller.Id,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		MaxViewers: req.MaxViewers,
		IsRevoked:  false,
		Label:      req.Label,
	}

	// Save to database
	vbolt.Write(ctx.Tx, AccessCodesBkt, code, &accessCode)

	// Add to appropriate index
	if req.Type == int(CodeTypeRoom) {
		vbolt.SetTargetSingleTerm(ctx.Tx, CodesByRoomIdx, code, req.TargetId)
	} else {
		vbolt.SetTargetSingleTerm(ctx.Tx, CodesByStudioIdx, code, req.TargetId)
	}

	// Add to creator index
	vbolt.SetTargetSingleTerm(ctx.Tx, CodesByCreatorIdx, code, caller.Id)

	// Initialize analytics
	analytics := CodeAnalytics{
		Code:             code,
		TotalConnections: 0,
		CurrentViewers:   0,
		PeakViewers:      0,
		PeakViewersAt:    time.Time{},
		LastConnectionAt: time.Time{},
	}
	vbolt.Write(ctx.Tx, CodeAnalyticsBkt, code, &analytics)

	vbolt.TxCommit(ctx.Tx)

	// Log code generation
	LogInfo(LogCategorySystem, "Access code generated", map[string]interface{}{
		"code":      code,
		"type":      req.Type,
		"targetId":  req.TargetId,
		"target":    targetName,
		"studioId":  studioId,
		"duration":  req.DurationMinutes,
		"expiresAt": expiresAt,
		"createdBy": caller.Id,
		"userEmail": caller.Email,
		"label":     req.Label,
	})

	// Build share URL
	shareURL := fmt.Sprintf("/watch/%s", code)

	resp.Success = true
	resp.Code = code
	resp.ExpiresAt = expiresAt
	resp.ShareURL = shareURL
	return
}

// validateAccessCodeLogic contains the core validation logic for access codes.
// This is extracted as a helper to allow both HTTP handler and procedure usage.
func validateAccessCodeLogic(db *vbolt.DB, code string) (resp ValidateAccessCodeResponse, err error) {
	// Validate code format (5 digits)
	if len(code) != 5 {
		resp.Success = false
		resp.Error = "Invalid code format"
		return
	}

	var accessCode AccessCode
	var sessionToken string
	var validationFailed bool
	now := time.Now()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Look up code in database
		vbolt.Read(tx, AccessCodesBkt, code, &accessCode)

		// Check if code exists
		if accessCode.Code == "" {
			resp.Success = false
			resp.Error = "Invalid code"
			validationFailed = true
			return
		}

		// Check if code is revoked
		if accessCode.IsRevoked {
			resp.Success = false
			resp.Error = "Code has been revoked"
			validationFailed = true
			return
		}

		// Check if code is expired
		if now.After(accessCode.ExpiresAt) {
			resp.Success = false
			resp.Error = "Code has expired"
			validationFailed = true
			return
		}

		// Check viewer limit (if set)
		if accessCode.MaxViewers > 0 {
			// Load current analytics to check viewer count
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

			if analytics.CurrentViewers >= accessCode.MaxViewers {
				resp.Success = false
				resp.Error = fmt.Sprintf("Stream is at capacity (%d/%d viewers)", analytics.CurrentViewers, accessCode.MaxViewers)
				validationFailed = true
				return
			}
		}

		// Generate session token
		var tokenErr error
		sessionToken, tokenErr = generateSessionToken()
		if tokenErr != nil {
			resp.Success = false
			resp.Error = "Failed to generate session token"
			err = tokenErr
			validationFailed = true
			return
		}

		// Create code session
		session := CodeSession{
			Token:            sessionToken,
			Code:             accessCode.Code,
			ConnectedAt:      now,
			LastSeen:         now,
			GracePeriodUntil: time.Time{}, // Not in grace period yet
			ClientIP:         "",          // Will be set by middleware later
			UserAgent:        "",          // Will be set by middleware later
		}
		vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

		// Add to session index
		vbolt.SetTargetSingleTerm(tx, SessionsByCodeIdx, sessionToken, accessCode.Code)

		// Update analytics
		var analytics CodeAnalytics
		vbolt.Read(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

		analytics.TotalConnections++
		analytics.CurrentViewers++
		analytics.LastConnectionAt = now

		// Update peak viewers if necessary
		if analytics.CurrentViewers > analytics.PeakViewers {
			analytics.PeakViewers = analytics.CurrentViewers
			analytics.PeakViewersAt = now
		}

		vbolt.Write(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

		vbolt.TxCommit(tx)
	})

	// If validation failed, return early
	if validationFailed || err != nil {
		return
	}

	// Build redirect URL based on code type
	var redirectTo string
	if accessCode.Type == CodeTypeRoom {
		redirectTo = fmt.Sprintf("/stream/%d", accessCode.TargetId)
	} else {
		// For studio codes, redirect to studio page (they can choose which room)
		redirectTo = fmt.Sprintf("/studio/%d", accessCode.TargetId)
	}

	// Log validation success
	LogInfo(LogCategorySystem, "Access code validated", map[string]interface{}{
		"code":         accessCode.Code,
		"sessionToken": sessionToken,
		"type":         accessCode.Type,
		"targetId":     accessCode.TargetId,
	})

	resp.Success = true
	resp.SessionToken = sessionToken
	resp.RedirectTo = redirectTo
	resp.ExpiresAt = accessCode.ExpiresAt
	resp.Type = int(accessCode.Type)
	resp.TargetId = accessCode.TargetId
	return
}

// ValidateAccessCode validates a 5-digit code and creates a viewing session
// This vbeam procedure works within an existing transaction context
// Note: The HTTP handler uses validateAccessCodeHandler instead (sets cookies)
func ValidateAccessCode(ctx *vbeam.Context, req ValidateAccessCodeRequest) (resp ValidateAccessCodeResponse, err error) {
	// Validate code format (5 digits)
	if len(req.Code) != 5 {
		resp.Success = false
		resp.Error = "Invalid code format"
		return
	}

	// Look up code in database
	var accessCode AccessCode
	vbolt.Read(ctx.Tx, AccessCodesBkt, req.Code, &accessCode)

	// Check if code exists
	if accessCode.Code == "" {
		resp.Success = false
		resp.Error = "Invalid code"
		return
	}

	// Check if code is revoked
	if accessCode.IsRevoked {
		resp.Success = false
		resp.Error = "Code has been revoked"
		return
	}

	// Check if code is expired
	now := time.Now()
	if now.After(accessCode.ExpiresAt) {
		resp.Success = false
		resp.Error = "Code has expired"
		return
	}

	// Check viewer limit (if set)
	if accessCode.MaxViewers > 0 {
		// Load current analytics to check viewer count
		var analytics CodeAnalytics
		vbolt.Read(ctx.Tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

		if analytics.CurrentViewers >= accessCode.MaxViewers {
			resp.Success = false
			resp.Error = fmt.Sprintf("Stream is at capacity (%d/%d viewers)", analytics.CurrentViewers, accessCode.MaxViewers)
			return
		}
	}

	vbeam.UseWriteTx(ctx)

	// Generate session token
	sessionToken, err := generateSessionToken()
	if err != nil {
		resp.Success = false
		resp.Error = "Failed to generate session token"
		return
	}

	// Create code session
	session := CodeSession{
		Token:            sessionToken,
		Code:             accessCode.Code,
		ConnectedAt:      now,
		LastSeen:         now,
		GracePeriodUntil: time.Time{}, // Not in grace period yet
		ClientIP:         "",          // Will be set by middleware later
		UserAgent:        "",          // Will be set by middleware later
	}
	vbolt.Write(ctx.Tx, CodeSessionsBkt, sessionToken, &session)

	// Add to session index
	vbolt.SetTargetSingleTerm(ctx.Tx, SessionsByCodeIdx, sessionToken, accessCode.Code)

	// Update analytics
	var analytics CodeAnalytics
	vbolt.Read(ctx.Tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

	analytics.TotalConnections++
	analytics.CurrentViewers++
	analytics.LastConnectionAt = now

	// Update peak viewers if necessary
	if analytics.CurrentViewers > analytics.PeakViewers {
		analytics.PeakViewers = analytics.CurrentViewers
		analytics.PeakViewersAt = now
	}

	vbolt.Write(ctx.Tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

	vbolt.TxCommit(ctx.Tx)

	// Build redirect URL based on code type
	var redirectTo string
	if accessCode.Type == CodeTypeRoom {
		redirectTo = fmt.Sprintf("/stream/%d", accessCode.TargetId)
	} else {
		// For studio codes, redirect to studio page (they can choose which room)
		redirectTo = fmt.Sprintf("/studio/%d", accessCode.TargetId)
	}

	// Log validation success
	LogInfo(LogCategorySystem, "Access code validated", map[string]interface{}{
		"code":           accessCode.Code,
		"sessionToken":   sessionToken,
		"type":           accessCode.Type,
		"targetId":       accessCode.TargetId,
		"currentViewers": analytics.CurrentViewers,
		"peakViewers":    analytics.PeakViewers,
	})

	resp.Success = true
	resp.SessionToken = sessionToken
	resp.RedirectTo = redirectTo
	resp.ExpiresAt = accessCode.ExpiresAt
	resp.Type = int(accessCode.Type)
	resp.TargetId = accessCode.TargetId
	return
}

// GetCodeStreamAccess checks if a session token can access a specific room
func GetCodeStreamAccess(ctx *vbeam.Context, req GetCodeStreamAccessRequest) (resp GetCodeStreamAccessResponse, err error) {
	now := time.Now()

	// Look up session by token
	var session CodeSession
	vbolt.Read(ctx.Tx, CodeSessionsBkt, req.SessionToken, &session)

	if session.Token == "" {
		resp.Allowed = false
		resp.Message = "Invalid session token"
		return
	}

	// Load the associated code
	var accessCode AccessCode
	vbolt.Read(ctx.Tx, AccessCodesBkt, session.Code, &accessCode)

	if accessCode.Code == "" {
		resp.Allowed = false
		resp.Message = "Access code not found"
		return
	}

	// Check if code is revoked
	if accessCode.IsRevoked {
		resp.Allowed = false
		resp.Message = "Access code has been revoked"
		return
	}

	// Check expiration with grace period logic
	inGracePeriod := false
	if now.After(accessCode.ExpiresAt) {
		// Code has expired - check if we're in grace period
		if !session.GracePeriodUntil.IsZero() && now.Before(session.GracePeriodUntil) {
			// Still in grace period
			inGracePeriod = true
		} else if session.GracePeriodUntil.IsZero() {
			// Code just expired, grant 15-minute grace period
			vbeam.UseWriteTx(ctx)
			session.GracePeriodUntil = now.Add(15 * time.Minute)
			vbolt.Write(ctx.Tx, CodeSessionsBkt, session.Token, &session)
			inGracePeriod = true
		} else {
			// Grace period has ended
			resp.Allowed = false
			resp.Message = "Access code has expired"
			return
		}
	}

	// Validate room access based on code type
	if accessCode.Type == CodeTypeRoom {
		// Room-specific code: verify roomId matches exactly
		if accessCode.TargetId != req.RoomId {
			resp.Allowed = false
			resp.Message = "This code is not valid for this room"
			return
		}
	} else {
		// Studio-wide code: verify roomId belongs to the same studio
		room := GetRoom(ctx.Tx, req.RoomId)
		if room.Id == 0 {
			resp.Allowed = false
			resp.Message = "Room not found"
			return
		}
		if room.StudioId != accessCode.TargetId {
			resp.Allowed = false
			resp.Message = "This code is not valid for this studio"
			return
		}
	}

	// Access granted - update LastSeen timestamp
	vbeam.UseWriteTx(ctx)
	session.LastSeen = now
	vbolt.Write(ctx.Tx, CodeSessionsBkt, session.Token, &session)

	// Get room to return studioId
	room := GetRoom(ctx.Tx, req.RoomId)

	resp.Allowed = true
	resp.RoomId = req.RoomId
	resp.StudioId = room.StudioId
	resp.ExpiresAt = accessCode.ExpiresAt
	resp.GracePeriod = inGracePeriod
	resp.Message = "Access granted"
	return
}

// RevokeAccessCode revokes an access code and terminates all active sessions
func RevokeAccessCode(ctx *vbeam.Context, req RevokeAccessCodeRequest) (resp RevokeAccessCodeResponse, err error) {
	// Look up the access code
	var accessCode AccessCode
	vbolt.Read(ctx.Tx, AccessCodesBkt, req.Code, &accessCode)

	if accessCode.Code == "" {
		resp.Success = false
		resp.Error = "Access code not found"
		return
	}

	// Check if already revoked
	if accessCode.IsRevoked {
		resp.Success = false
		resp.Error = "Access code is already revoked"
		return
	}

	// Determine studio ID for permission check
	var studioId int
	if accessCode.Type == CodeTypeRoom {
		room := GetRoom(ctx.Tx, accessCode.TargetId)
		if room.Id == 0 {
			resp.Success = false
			resp.Error = "Room not found"
			return
		}
		studioId = room.StudioId
	} else {
		studioId = accessCode.TargetId
	}

	// Check admin permission
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	if !HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Admin permission required"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Mark code as revoked
	accessCode.IsRevoked = true
	vbolt.Write(ctx.Tx, AccessCodesBkt, accessCode.Code, &accessCode)

	// Find all active sessions for this code
	var sessionTokens []string
	vbolt.ReadTermTargets(ctx.Tx, SessionsByCodeIdx, accessCode.Code, &sessionTokens, vbolt.Window{})

	// Delete each session and update analytics
	var analytics CodeAnalytics
	vbolt.Read(ctx.Tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

	sessionsKilled := 0
	for _, token := range sessionTokens {
		// Delete session
		var emptySession CodeSession
		vbolt.Write(ctx.Tx, CodeSessionsBkt, token, &emptySession)

		// Remove from index
		vbolt.SetTargetSingleTerm(ctx.Tx, SessionsByCodeIdx, token, "")

		sessionsKilled++
	}

	// Update analytics to reflect terminated sessions
	if analytics.CurrentViewers >= sessionsKilled {
		analytics.CurrentViewers -= sessionsKilled
	} else {
		analytics.CurrentViewers = 0
	}
	vbolt.Write(ctx.Tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

	// For studio codes, fetch room IDs before committing transaction
	var roomIds []int
	if accessCode.Type == CodeTypeStudio {
		vbolt.ReadTermTargets(ctx.Tx, RoomsByStudioIdx, studioId, &roomIds, vbolt.Window{})
	}

	vbolt.TxCommit(ctx.Tx)

	// Broadcast CODE_REVOKED event to affected viewers only
	// Only viewers using THIS specific revoked code will receive the event
	// For room codes: broadcast to that specific room
	// For studio codes: broadcast to all rooms in the studio
	if accessCode.Type == CodeTypeRoom {
		sseManager.BroadcastCodeRevoked(accessCode.TargetId, sessionTokens)
	} else {
		// Studio code - broadcast to all rooms in the studio
		for _, roomId := range roomIds {
			sseManager.BroadcastCodeRevoked(roomId, sessionTokens)
		}
	}

	// Log revocation
	LogInfo(LogCategorySystem, "Access code revoked", map[string]interface{}{
		"code":           accessCode.Code,
		"type":           accessCode.Type,
		"targetId":       accessCode.TargetId,
		"studioId":       studioId,
		"sessionsKilled": sessionsKilled,
		"revokedBy":      caller.Id,
		"userEmail":      caller.Email,
	})

	resp.Success = true
	resp.SessionsKilled = sessionsKilled
	return
}

// ListAccessCodes returns all access codes for a room or studio
func ListAccessCodes(ctx *vbeam.Context, req ListAccessCodesRequest) (resp ListAccessCodesResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Validate code type
	if req.Type != int(CodeTypeRoom) && req.Type != int(CodeTypeStudio) {
		resp.Success = false
		resp.Error = "Invalid code type (must be 0 for room or 1 for studio)"
		return
	}

	var studioId int
	var targetName string

	// Validate target and check permissions based on type
	if req.Type == int(CodeTypeRoom) {
		// Room code list - validate room exists and check permission
		room := GetRoom(ctx.Tx, req.TargetId)
		if room.Id == 0 {
			resp.Success = false
			resp.Error = "Room not found"
			return
		}
		studioId = room.StudioId
		targetName = room.Name
	} else {
		// Studio code list - validate studio exists and check permission
		studio := GetStudioById(ctx.Tx, req.TargetId)
		if studio.Id == 0 {
			resp.Success = false
			resp.Error = "Studio not found"
			return
		}
		studioId = studio.Id
		targetName = studio.Name
	}

	// Check if user has Viewer+ permission for the studio
	if !HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleViewer) {
		resp.Success = false
		resp.Error = "You do not have permission to view access codes for this " +
			map[bool]string{true: "room", false: "studio"}[req.Type == int(CodeTypeRoom)]
		return
	}

	// Query appropriate index to get code strings
	var codes []string
	if req.Type == int(CodeTypeRoom) {
		vbolt.ReadTermTargets(ctx.Tx, CodesByRoomIdx, req.TargetId, &codes, vbolt.Window{})
	} else {
		vbolt.ReadTermTargets(ctx.Tx, CodesByStudioIdx, req.TargetId, &codes, vbolt.Window{})
	}

	// Build list items
	var items []AccessCodeListItem
	now := time.Now()

	for _, code := range codes {
		// Load access code
		var accessCode AccessCode
		vbolt.Read(ctx.Tx, AccessCodesBkt, code, &accessCode)
		if accessCode.Code == "" {
			continue // Skip if code not found
		}

		// Load analytics
		var analytics CodeAnalytics
		vbolt.Read(ctx.Tx, CodeAnalyticsBkt, code, &analytics)

		// Build list item
		item := AccessCodeListItem{
			Code:           accessCode.Code,
			Type:           int(accessCode.Type),
			Label:          accessCode.Label,
			CreatedAt:      accessCode.CreatedAt,
			ExpiresAt:      accessCode.ExpiresAt,
			IsRevoked:      accessCode.IsRevoked,
			IsExpired:      now.After(accessCode.ExpiresAt),
			CurrentViewers: analytics.CurrentViewers,
			TotalViews:     analytics.TotalConnections,
		}
		items = append(items, item)
	}

	// Sort by creation time (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	// Log access
	LogInfo(LogCategorySystem, "Access codes listed", map[string]interface{}{
		"type":      req.Type,
		"targetId":  req.TargetId,
		"target":    targetName,
		"studioId":  studioId,
		"codeCount": len(items),
		"requestBy": caller.Id,
		"userEmail": caller.Email,
	})

	resp.Success = true
	resp.Codes = items
	return
}

// anonymizeIP masks the last octet of an IP address for privacy
func anonymizeIP(ip string) string {
	if ip == "" {
		return ""
	}
	// Simple anonymization: replace last segment with "xxx"
	// e.g., "192.168.1.100" -> "192.168.1.xxx"
	lastDot := -1
	for i := len(ip) - 1; i >= 0; i-- {
		if ip[i] == '.' || ip[i] == ':' {
			lastDot = i
			break
		}
	}
	if lastDot > 0 {
		return ip[:lastDot+1] + "xxx"
	}
	return "xxx" // Fallback if no dot/colon found
}

// GetCodeAnalytics returns detailed analytics for a specific access code
func GetCodeAnalytics(ctx *vbeam.Context, req GetCodeAnalyticsRequest) (resp GetCodeAnalyticsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Validate code format
	if len(req.Code) != 5 {
		resp.Success = false
		resp.Error = "Invalid code format"
		return
	}

	// Look up code
	var accessCode AccessCode
	vbolt.Read(ctx.Tx, AccessCodesBkt, req.Code, &accessCode)

	if accessCode.Code == "" {
		resp.Success = false
		resp.Error = "Access code not found"
		return
	}

	// Determine studio ID for permission check
	var studioId int
	if accessCode.Type == CodeTypeRoom {
		room := GetRoom(ctx.Tx, accessCode.TargetId)
		if room.Id == 0 {
			resp.Success = false
			resp.Error = "Room not found"
			return
		}
		studioId = room.StudioId
	} else {
		studioId = accessCode.TargetId
	}

	// Check admin permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studioId, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can view code analytics"
		return
	}

	// Load analytics
	var analytics CodeAnalytics
	vbolt.Read(ctx.Tx, CodeAnalyticsBkt, accessCode.Code, &analytics)

	// Calculate status
	now := time.Now()
	status := "active"
	if accessCode.IsRevoked {
		status = "revoked"
	} else if now.After(accessCode.ExpiresAt) {
		status = "expired"
	}

	// Load active sessions
	var sessionTokens []string
	vbolt.ReadTermTargets(ctx.Tx, SessionsByCodeIdx, accessCode.Code, &sessionTokens, vbolt.Window{})

	var sessions []SessionInfo
	for _, token := range sessionTokens {
		var session CodeSession
		vbolt.Read(ctx.Tx, CodeSessionsBkt, token, &session)
		if session.Token == "" {
			continue // Skip if session not found
		}

		// Calculate duration
		duration := int(session.LastSeen.Sub(session.ConnectedAt).Seconds())

		// Determine if session is still active (LastSeen within last 10 minutes)
		isActive := now.Sub(session.LastSeen) < 10*time.Minute

		// Build session info
		info := SessionInfo{
			ConnectedAt: session.ConnectedAt,
			Duration:    duration,
			ClientIP:    anonymizeIP(session.ClientIP),
			UserAgent:   session.UserAgent,
			IsActive:    isActive,
		}
		sessions = append(sessions, info)
	}

	// Log access
	LogInfo(LogCategorySystem, "Code analytics viewed", map[string]interface{}{
		"code":         accessCode.Code,
		"type":         accessCode.Type,
		"targetId":     accessCode.TargetId,
		"studioId":     studioId,
		"status":       status,
		"sessionCount": len(sessions),
		"viewedBy":     caller.Id,
		"userEmail":    caller.Email,
	})

	resp.Success = true
	resp.Code = accessCode.Code
	resp.Type = int(accessCode.Type)
	resp.Label = accessCode.Label
	resp.Status = status
	resp.CreatedAt = accessCode.CreatedAt
	resp.ExpiresAt = accessCode.ExpiresAt
	resp.TotalConnections = analytics.TotalConnections
	resp.CurrentViewers = analytics.CurrentViewers
	resp.PeakViewers = analytics.PeakViewers
	resp.PeakViewersAt = analytics.PeakViewersAt
	resp.Sessions = sessions
	return
}
