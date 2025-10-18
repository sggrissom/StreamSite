package backend

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"stream/cfg"
	"time"

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
