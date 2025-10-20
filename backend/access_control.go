package backend

import (
	"time"

	"go.hasen.dev/vbolt"
)

// RoomAccessResult encapsulates the result of a room access check
type RoomAccessResult struct {
	Allowed       bool       // Whether access is granted
	Role          StudioRole // User's role in the studio (if allowed)
	IsCodeAuth    bool       // Whether access is via code (vs studio membership)
	CodeExpiresAt *time.Time // When code access expires (if IsCodeAuth)
	DenialReason  string     // Human-readable reason for denial
}

// CheckRoomAccess is the single source of truth for room access permissions
// Handles all access types: anonymous code sessions, logged-in code sessions,
// studio membership, and site admin privileges
//
// Parameters:
//   - tx: Database transaction
//   - user: The user attempting access (can have Id=-1 for anonymous)
//   - roomId: The room being accessed
//   - anonymousSessionToken: Session token for anonymous users (userId=-1), empty for logged-in
//
// Returns: RoomAccessResult with access decision and metadata
func CheckRoomAccess(tx *vbolt.Tx, user User, roomId int, anonymousSessionToken string) RoomAccessResult {
	// Get room
	room := GetRoom(tx, roomId)
	if room.Id == 0 {
		return RoomAccessResult{
			Allowed:      false,
			DenialReason: "Room not found",
		}
	}

	// Get studio
	studio := GetStudioById(tx, room.StudioId)
	if studio.Id == 0 {
		return RoomAccessResult{
			Allowed:      false,
			DenialReason: "Studio not found",
		}
	}

	// 1. Check code-based access (works for both anonymous and logged-in users)
	codeAccess := checkCodeAccessForRoom(tx, user.Id, anonymousSessionToken, roomId, room.StudioId)
	if codeAccess.Allowed {
		return codeAccess
	}

	// 2. Check studio membership (only for logged-in users)
	if user.Id > 0 {
		role := GetUserStudioRole(tx, user.Id, studio.Id)

		// Site admins can view all rooms
		if user.Role == RoleSiteAdmin {
			if role == -1 {
				role = StudioRoleOwner // Give admins owner role for display
			}
			return RoomAccessResult{
				Allowed:    true,
				Role:       role,
				IsCodeAuth: false,
			}
		}

		// Regular member access
		if role != -1 {
			return RoomAccessResult{
				Allowed:    true,
				Role:       role,
				IsCodeAuth: false,
			}
		}
	}

	// No access granted
	return RoomAccessResult{
		Allowed:      false,
		DenialReason: "You do not have permission to view this room",
	}
}

// checkCodeAccessForRoom checks if a user has code-based access to a specific room
// Handles both anonymous (userId=-1) and logged-in (userId>0) users
func checkCodeAccessForRoom(tx *vbolt.Tx, userId int, anonymousSessionToken string, roomId int, studioId int) RoomAccessResult {
	var sessionToken string

	if userId == -1 {
		// Anonymous user - use provided session token
		sessionToken = anonymousSessionToken
	} else {
		// Logged-in user - check UserCodeSessionsBkt
		vbolt.Read(tx, UserCodeSessionsBkt, userId, &sessionToken)
	}

	if sessionToken == "" {
		return RoomAccessResult{Allowed: false}
	}

	// Load session
	var session CodeSession
	vbolt.Read(tx, CodeSessionsBkt, sessionToken, &session)

	if session.Token == "" {
		return RoomAccessResult{Allowed: false}
	}

	// Load access code
	var accessCode AccessCode
	vbolt.Read(tx, AccessCodesBkt, session.Code, &accessCode)

	if accessCode.Code == "" || accessCode.IsRevoked {
		return RoomAccessResult{Allowed: false}
	}

	// Validate code grants access to this specific room
	var grantsAccess bool

	if accessCode.Type == CodeTypeRoom {
		// Room-specific code: must match requested room
		grantsAccess = (accessCode.TargetId == roomId)
	} else if accessCode.Type == CodeTypeStudio {
		// Studio-wide code: must match room's studio
		grantsAccess = (accessCode.TargetId == studioId)
	}

	if !grantsAccess {
		return RoomAccessResult{Allowed: false}
	}

	// Access granted via code
	return RoomAccessResult{
		Allowed:       true,
		Role:          StudioRoleViewer,
		IsCodeAuth:    true,
		CodeExpiresAt: &accessCode.ExpiresAt,
	}
}

// GetUserCodeSession loads a user's active code session if they have one
// Returns the session and access code, or empty structs if no active session
func GetUserCodeSession(tx *vbolt.Tx, userId int) (CodeSession, AccessCode) {
	var session CodeSession
	var accessCode AccessCode

	if userId <= 0 {
		return session, accessCode
	}

	// Check if user has active code session
	var sessionToken string
	vbolt.Read(tx, UserCodeSessionsBkt, userId, &sessionToken)

	if sessionToken == "" {
		return session, accessCode
	}

	// Load session
	vbolt.Read(tx, CodeSessionsBkt, sessionToken, &session)
	if session.Token == "" {
		return session, accessCode
	}

	// Load access code
	vbolt.Read(tx, AccessCodesBkt, session.Code, &accessCode)

	return session, accessCode
}

// GetCodeSessionFromToken loads a code session and access code from a session token
// Used for both anonymous and logged-in users who have code access
// Returns empty structs if session or code not found
func GetCodeSessionFromToken(tx *vbolt.Tx, sessionToken string) (CodeSession, AccessCode) {
	var session CodeSession
	var accessCode AccessCode

	if sessionToken == "" {
		return session, accessCode
	}

	// Load session
	vbolt.Read(tx, CodeSessionsBkt, sessionToken, &session)
	if session.Token == "" {
		return session, accessCode
	}

	// Load access code
	vbolt.Read(tx, AccessCodesBkt, session.Code, &accessCode)

	return session, accessCode
}

// GetRoomsAccessibleViaCode returns all rooms accessible via an access code
// Handles both room-specific and studio-wide codes
func GetRoomsAccessibleViaCode(tx *vbolt.Tx, accessCode AccessCode) []RoomWithStudio {
	rooms := make([]RoomWithStudio, 0)

	if accessCode.Code == "" || accessCode.IsRevoked {
		return rooms
	}

	if accessCode.Type == CodeTypeRoom {
		// Room-specific code: return just that room
		room := GetRoom(tx, accessCode.TargetId)
		if room.Id != 0 {
			studio := GetStudioById(tx, room.StudioId)
			rooms = append(rooms, RoomWithStudio{
				Room:       room,
				StudioName: studio.Name,
			})
		}
	} else if accessCode.Type == CodeTypeStudio {
		// Studio-wide code: return all rooms in that studio
		studio := GetStudioById(tx, accessCode.TargetId)
		if studio.Id != 0 {
			studioRooms := ListStudioRooms(tx, studio.Id)
			for _, room := range studioRooms {
				rooms = append(rooms, RoomWithStudio{
					Room:       room,
					StudioName: studio.Name,
				})
			}
		}
	}

	return rooms
}
