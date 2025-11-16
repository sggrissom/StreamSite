package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestAccessControlDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_access_control.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

// Helper to create test studio and room
func createTestStudioAndRoom(tx *vbolt.Tx) (Studio, Room) {
	studio := Studio{
		Id:       vbolt.NextIntId(tx, StudiosBkt),
		Name:     "Test Studio",
		MaxRooms: 5,
		OwnerId:  1,
		Creation: time.Now(),
	}
	vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

	room := Room{
		Id:         vbolt.NextIntId(tx, RoomsBkt),
		StudioId:   studio.Id,
		RoomNumber: 1,
		Name:       "Test Room",
		StreamKey:  "test-key-123",
		Creation:   time.Now(),
	}
	vbolt.Write(tx, RoomsBkt, room.Id, &room)

	// Index room by studio
	vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

	return studio, room
}

func TestCheckRoomAccess(t *testing.T) {
	t.Run("RoomNotFound", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		user := User{Id: 1, Email: "test@example.com", Name: "Test User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, user, 99999, "")
		})

		if result.Allowed {
			t.Error("Expected access to be denied for non-existent room")
		}

		if result.DenialReason != "Room not found" {
			t.Errorf("Expected 'Room not found', got '%s'", result.DenialReason)
		}
	})

	t.Run("AnonymousWithValidRoomCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var roomId int
		var sessionToken string
		var codeExpiresAt time.Time

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			_, room := createTestStudioAndRoom(tx)
			roomId = room.Id

			// Create access code for this room
			code := "12345"
			codeExpiresAt = time.Now().Add(1 * time.Hour)
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  codeExpiresAt,
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		// Test anonymous user with code
		anonymousUser := User{Id: -1, Email: "anonymous@code-session", Name: "Access Code User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, anonymousUser, roomId, sessionToken)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for valid anonymous code session")
		}

		if !result.IsCodeAuth {
			t.Error("Expected IsCodeAuth to be true")
		}

		if result.Role != StudioRoleViewer {
			t.Errorf("Expected role Viewer, got %d", result.Role)
		}

		if result.CodeExpiresAt == nil {
			t.Error("Expected CodeExpiresAt to be set")
		}
	})

	t.Run("AnonymousWithValidStudioCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var roomId int
		var sessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			studio, room := createTestStudioAndRoom(tx)
			roomId = room.Id

			// Create studio-wide access code
			code := "99999"
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeStudio,
				TargetId:   studio.Id, // Studio ID, not room ID
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		anonymousUser := User{Id: -1, Email: "anonymous@code-session", Name: "Access Code User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, anonymousUser, roomId, sessionToken)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for valid studio code")
		}

		if !result.IsCodeAuth {
			t.Error("Expected IsCodeAuth to be true for studio code")
		}
	})

	t.Run("AnonymousWithRevokedCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var roomId int
		var sessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			_, room := createTestStudioAndRoom(tx)
			roomId = room.Id

			// Create REVOKED access code
			code := "12345"
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  true, // REVOKED
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		anonymousUser := User{Id: -1, Email: "anonymous@code-session", Name: "Access Code User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, anonymousUser, roomId, sessionToken)
		})

		if result.Allowed {
			t.Error("Expected access to be denied for revoked code")
		}
	})

	t.Run("AnonymousWithCodeForWrongRoom", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var room1Id, room2Id int
		var sessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			studio, room1 := createTestStudioAndRoom(tx)
			room1Id = room1.Id

			// Create second room in same studio
			room2 := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 2,
				Name:       "Test Room 2",
				StreamKey:  "test-key-456",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room2.Id, studio.Id)
			room2Id = room2.Id

			// Create code for room1 only
			code := "12345"
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room1Id, // Code is for room1
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		anonymousUser := User{Id: -1, Email: "anonymous@code-session", Name: "Access Code User"}

		// Try to access room2 with code for room1
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, anonymousUser, room2Id, sessionToken)
		})

		if result.Allowed {
			t.Error("Expected access to be denied when code is for different room")
		}
	})

	t.Run("LoggedInUserWithCodeAccess", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, roomId int
		var sessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "user@example.com",
				Name:     "Test User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio and room
			_, room := createTestStudioAndRoom(tx)
			roomId = room.Id

			// Create access code
			code := "12345"
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			// Map user to session in UserCodeSessionsBkt
			vbolt.Write(tx, UserCodeSessionsBkt, userId, &sessionToken)

			vbolt.TxCommit(tx)
		})

		loggedInUser := User{Id: userId, Email: "user@example.com", Name: "Test User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, loggedInUser, roomId, "")
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for logged-in user with code")
		}

		if !result.IsCodeAuth {
			t.Error("Expected IsCodeAuth to be true for logged-in user with code")
		}

		if result.Role != StudioRoleViewer {
			t.Errorf("Expected role Viewer, got %d", result.Role)
		}
	})

	t.Run("LoggedInUserWithStudioMembership", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, roomId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "member@example.com",
				Name:     "Member User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio and room
			studio, room := createTestStudioAndRoom(tx)
			roomId = room.Id
			studioId = studio.Id

			// Add user as member
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				StudioId: studioId,
				UserId:   userId,
				Role:     StudioRoleMember,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)

			// Index membership
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

			vbolt.TxCommit(tx)
		})

		memberUser := User{Id: userId, Email: "member@example.com", Name: "Member User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, memberUser, roomId, "")
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for studio member")
		}

		if result.IsCodeAuth {
			t.Error("Expected IsCodeAuth to be false for membership-based access")
		}

		if result.Role != StudioRoleMember {
			t.Errorf("Expected role Member, got %d", result.Role)
		}
	})

	t.Run("SiteAdminWithNoMembership", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, roomId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create site admin user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "admin@example.com",
				Name:     "Admin User",
				Role:     RoleSiteAdmin,
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio and room (admin is NOT a member)
			_, room := createTestStudioAndRoom(tx)
			roomId = room.Id

			vbolt.TxCommit(tx)
		})

		adminUser := User{Id: userId, Email: "admin@example.com", Name: "Admin User", Role: RoleSiteAdmin}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, adminUser, roomId, "")
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for site admin")
		}

		if result.Role != StudioRoleOwner {
			t.Errorf("Expected role Owner for site admin, got %d", result.Role)
		}
	})

	t.Run("NoAccess", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, roomId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create regular user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "user@example.com",
				Name:     "Regular User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio and room (user has no access)
			_, room := createTestStudioAndRoom(tx)
			roomId = room.Id

			vbolt.TxCommit(tx)
		})

		regularUser := User{Id: userId, Email: "user@example.com", Name: "Regular User"}

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, regularUser, roomId, "")
		})

		if result.Allowed {
			t.Error("Expected access to be denied for user with no access")
		}

		if result.DenialReason != "You do not have permission to view this room" {
			t.Errorf("Expected permission denied message, got '%s'", result.DenialReason)
		}
	})
}

func TestGetUserCodeSession(t *testing.T) {
	t.Run("UserHasCodeSession", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId int
		var expectedCode string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			userId = 1

			// Create access code
			expectedCode = "12345"
			accessCode := AccessCode{
				Code:       expectedCode,
				Type:       CodeTypeRoom,
				TargetId:   1,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			sessionToken, _ := generateSessionToken()
			session := CodeSession{
				Token:       sessionToken,
				Code:        expectedCode,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			// Map user to session
			vbolt.Write(tx, UserCodeSessionsBkt, userId, &sessionToken)

			vbolt.TxCommit(tx)
		})

		var session CodeSession
		var accessCode AccessCode
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			session, accessCode = GetUserCodeSession(tx, userId)
		})

		if session.Token == "" {
			t.Error("Expected session to be returned")
		}

		if accessCode.Code != expectedCode {
			t.Errorf("Expected code '%s', got '%s'", expectedCode, accessCode.Code)
		}
	})

	t.Run("UserHasNoCodeSession", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var session CodeSession
		var accessCode AccessCode
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			session, accessCode = GetUserCodeSession(tx, 999)
		})

		if session.Token != "" {
			t.Error("Expected empty session for user without code access")
		}

		if accessCode.Code != "" {
			t.Error("Expected empty access code for user without code access")
		}
	})

	t.Run("InvalidUserId", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var session CodeSession
		var accessCode AccessCode
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			session, accessCode = GetUserCodeSession(tx, -1)
		})

		if session.Token != "" {
			t.Error("Expected empty session for userId <= 0")
		}

		if accessCode.Code != "" {
			t.Error("Expected empty access code for userId <= 0")
		}
	})
}

func TestGetCodeSessionFromToken(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var sessionToken string
		var expectedCode string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create access code
			expectedCode = "67890"
			accessCode := AccessCode{
				Code:       expectedCode,
				Type:       CodeTypeRoom,
				TargetId:   1,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        expectedCode,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		var session CodeSession
		var accessCode AccessCode
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			session, accessCode = GetCodeSessionFromToken(tx, sessionToken)
		})

		if session.Token != sessionToken {
			t.Errorf("Expected session token '%s', got '%s'", sessionToken, session.Token)
		}

		if accessCode.Code != expectedCode {
			t.Errorf("Expected code '%s', got '%s'", expectedCode, accessCode.Code)
		}
	})

	t.Run("EmptyToken", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var session CodeSession
		var accessCode AccessCode
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			session, accessCode = GetCodeSessionFromToken(tx, "")
		})

		if session.Token != "" {
			t.Error("Expected empty session for empty token")
		}

		if accessCode.Code != "" {
			t.Error("Expected empty access code for empty token")
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var session CodeSession
		var accessCode AccessCode
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			session, accessCode = GetCodeSessionFromToken(tx, "invalid-token-xyz")
		})

		if session.Token != "" {
			t.Error("Expected empty session for invalid token")
		}

		if accessCode.Code != "" {
			t.Error("Expected empty access code for invalid token")
		}
	})
}

func TestGetRoomsAccessibleViaCode(t *testing.T) {
	t.Run("RoomCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var accessCode AccessCode
		var expectedRoomId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			_, room := createTestStudioAndRoom(tx)
			expectedRoomId = room.Id

			// Create room-specific code
			code := "12345"
			accessCode = AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			vbolt.TxCommit(tx)
		})

		var rooms []RoomWithStudio
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			rooms = GetRoomsAccessibleViaCode(tx, accessCode)
		})

		if len(rooms) != 1 {
			t.Errorf("Expected 1 room, got %d", len(rooms))
		}

		if len(rooms) > 0 && rooms[0].Room.Id != expectedRoomId {
			t.Errorf("Expected room ID %d, got %d", expectedRoomId, rooms[0].Room.Id)
		}
	})

	t.Run("StudioCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var accessCode AccessCode

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			studio, _ := createTestStudioAndRoom(tx)

			// Create second room in same studio
			room2 := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 2,
				Name:       "Test Room 2",
				StreamKey:  "test-key-456",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room2.Id, studio.Id)

			// Create studio-wide code
			code := "99999"
			accessCode = AccessCode{
				Code:       code,
				Type:       CodeTypeStudio,
				TargetId:   studio.Id,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			vbolt.TxCommit(tx)
		})

		var rooms []RoomWithStudio
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			rooms = GetRoomsAccessibleViaCode(tx, accessCode)
		})

		if len(rooms) != 2 {
			t.Errorf("Expected 2 rooms, got %d", len(rooms))
		}
	})

	t.Run("RevokedCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		accessCode := AccessCode{
			Code:       "12345",
			Type:       CodeTypeRoom,
			TargetId:   1,
			CreatedBy:  1,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
			MaxViewers: 10,
			IsRevoked:  true, // REVOKED
		}

		var rooms []RoomWithStudio
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			rooms = GetRoomsAccessibleViaCode(tx, accessCode)
		})

		if len(rooms) != 0 {
			t.Errorf("Expected 0 rooms for revoked code, got %d", len(rooms))
		}
	})

	t.Run("EmptyCode", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		accessCode := AccessCode{} // Empty code

		var rooms []RoomWithStudio
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			rooms = GetRoomsAccessibleViaCode(tx, accessCode)
		})

		if len(rooms) != 0 {
			t.Errorf("Expected 0 rooms for empty code, got %d", len(rooms))
		}
	})
}

func TestCheckStudioAccess(t *testing.T) {
	t.Run("StudioNotFound", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		user := User{Id: 1, Email: "test@example.com", Name: "Test User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, user, 99999)
		})

		if result.Allowed {
			t.Error("Expected access to be denied for non-existent studio")
		}

		if result.DenialReason != "Studio not found" {
			t.Errorf("Expected 'Studio not found', got '%s'", result.DenialReason)
		}
	})

	t.Run("AnonymousUserDenied", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id
			vbolt.TxCommit(tx)
		})

		anonymousUser := User{Id: -1, Email: "anonymous@code-session", Name: "Access Code User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, anonymousUser, studioId)
		})

		if result.Allowed {
			t.Error("Expected access to be denied for anonymous user")
		}

		if result.DenialReason != "You must be logged in to access studio management" {
			t.Errorf("Expected login required message, got '%s'", result.DenialReason)
		}
	})

	t.Run("LoggedInUserWithCodeSessionButNoMembership", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user (not a studio member)
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "user@example.com",
				Name:     "Test User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			// User has code session but NOT studio membership
			code := "12345"
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeStudio,
				TargetId:   studioId,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 10,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			sessionToken, _ := generateSessionToken()
			session := CodeSession{
				Token:       sessionToken,
				Code:        code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)
			vbolt.Write(tx, UserCodeSessionsBkt, userId, &sessionToken)

			vbolt.TxCommit(tx)
		})

		loggedInUser := User{Id: userId, Email: "user@example.com", Name: "Test User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, loggedInUser, studioId)
		})

		if result.Allowed {
			t.Error("Expected access to be denied for user with code session but no membership")
		}

		if result.DenialReason != "You do not have permission to view this studio" {
			t.Errorf("Expected permission denied message, got '%s'", result.DenialReason)
		}
	})

	t.Run("StudioViewer", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "viewer@example.com",
				Name:     "Viewer User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			// Add user as viewer
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				StudioId: studioId,
				UserId:   userId,
				Role:     StudioRoleViewer,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

			vbolt.TxCommit(tx)
		})

		viewerUser := User{Id: userId, Email: "viewer@example.com", Name: "Viewer User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, viewerUser, studioId)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for studio viewer")
		}

		if result.Role != StudioRoleViewer {
			t.Errorf("Expected role Viewer, got %d", result.Role)
		}

		if result.IsSiteAdmin {
			t.Error("Expected IsSiteAdmin to be false")
		}
	})

	t.Run("StudioMember", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "member@example.com",
				Name:     "Member User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			// Add user as member
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				StudioId: studioId,
				UserId:   userId,
				Role:     StudioRoleMember,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

			vbolt.TxCommit(tx)
		})

		memberUser := User{Id: userId, Email: "member@example.com", Name: "Member User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, memberUser, studioId)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for studio member")
		}

		if result.Role != StudioRoleMember {
			t.Errorf("Expected role Member, got %d", result.Role)
		}

		if result.IsSiteAdmin {
			t.Error("Expected IsSiteAdmin to be false")
		}
	})

	t.Run("StudioAdmin", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "admin@example.com",
				Name:     "Admin User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			// Add user as admin
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				StudioId: studioId,
				UserId:   userId,
				Role:     StudioRoleAdmin,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

			vbolt.TxCommit(tx)
		})

		adminUser := User{Id: userId, Email: "admin@example.com", Name: "Admin User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, adminUser, studioId)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for studio admin")
		}

		if result.Role != StudioRoleAdmin {
			t.Errorf("Expected role Admin, got %d", result.Role)
		}

		if result.IsSiteAdmin {
			t.Error("Expected IsSiteAdmin to be false")
		}
	})

	t.Run("StudioOwner", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "owner@example.com",
				Name:     "Owner User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio with this user as owner
			studio := Studio{
				Id:       vbolt.NextIntId(tx, StudiosBkt),
				Name:     "Test Studio",
				MaxRooms: 5,
				OwnerId:  userId,
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)
			studioId = studio.Id

			// Add owner membership
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				StudioId: studioId,
				UserId:   userId,
				Role:     StudioRoleOwner,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

			vbolt.TxCommit(tx)
		})

		ownerUser := User{Id: userId, Email: "owner@example.com", Name: "Owner User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, ownerUser, studioId)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for studio owner")
		}

		if result.Role != StudioRoleOwner {
			t.Errorf("Expected role Owner, got %d", result.Role)
		}

		if result.IsSiteAdmin {
			t.Error("Expected IsSiteAdmin to be false")
		}
	})

	t.Run("SiteAdminWithNoMembership", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create site admin user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "admin@example.com",
				Name:     "Site Admin",
				Role:     RoleSiteAdmin,
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio (admin is NOT a member)
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			vbolt.TxCommit(tx)
		})

		siteAdminUser := User{Id: userId, Email: "admin@example.com", Name: "Site Admin", Role: RoleSiteAdmin}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, siteAdminUser, studioId)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for site admin")
		}

		if result.Role != StudioRoleOwner {
			t.Errorf("Expected role Owner for site admin, got %d", result.Role)
		}

		if !result.IsSiteAdmin {
			t.Error("Expected IsSiteAdmin to be true")
		}
	})

	t.Run("SiteAdminWhoIsAlsoMember", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create site admin user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "admin@example.com",
				Name:     "Site Admin",
				Role:     RoleSiteAdmin,
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			// Add admin as member (not owner)
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				StudioId: studioId,
				UserId:   userId,
				Role:     StudioRoleMember,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

			vbolt.TxCommit(tx)
		})

		siteAdminUser := User{Id: userId, Email: "admin@example.com", Name: "Site Admin", Role: RoleSiteAdmin}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, siteAdminUser, studioId)
		})

		if !result.Allowed {
			t.Error("Expected access to be allowed for site admin who is also member")
		}

		// Should get their actual membership role, not Owner
		if result.Role != StudioRoleMember {
			t.Errorf("Expected role Member (actual membership), got %d", result.Role)
		}

		if !result.IsSiteAdmin {
			t.Error("Expected IsSiteAdmin to be true")
		}
	})

	t.Run("RegularUserWithNoAccess", func(t *testing.T) {
		db := setupTestAccessControlDB(t)
		defer db.Close()

		var userId, studioId int

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create regular user
			user := User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "user@example.com",
				Name:     "Regular User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			userId = user.Id

			// Create studio (user has no access)
			studio, _ := createTestStudioAndRoom(tx)
			studioId = studio.Id

			vbolt.TxCommit(tx)
		})

		regularUser := User{Id: userId, Email: "user@example.com", Name: "Regular User"}

		var result StudioAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckStudioAccess(tx, regularUser, studioId)
		})

		if result.Allowed {
			t.Error("Expected access to be denied for user with no access")
		}

		if result.DenialReason != "You do not have permission to view this studio" {
			t.Errorf("Expected permission denied message, got '%s'", result.DenialReason)
		}
	})
}

// TestCheckClassPermission tests the CheckClassPermission function
func TestCheckClassPermission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var testUser User
	var testRoom Room
	var activeSchedule ClassSchedule
	var permission ClassPermission

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create user
		testUser = createTestUser(t, tx, "user@test.com", RoleStreamAdmin)

		// Create studio and room
		studio, room := createTestStudioAndRoom(tx)
		testRoom = room

		// Create active schedule (happening now)
		now := time.Now()
		activeSchedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Active Class",
			IsRecurring:     false,
			StartTime:       now.Add(-30 * time.Minute), // Started 30 minutes ago
			EndTime:         now.Add(30 * time.Minute),  // Ends in 30 minutes
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, activeSchedule.Id, &activeSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, activeSchedule.Id, activeSchedule.RoomId)

		// Grant permission to user
		permission = ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: activeSchedule.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleViewer),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, permission.Id, &permission)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, permission.Id, permission.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, permission.Id, permission.UserId)

		vbolt.TxCommit(tx)
	})

	// Test 1: Access granted during active class
	t.Run("ActiveClass", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckClassPermission(tx, testUser.Id, testRoom.Id)
		})

		if !result.Allowed {
			t.Error("Expected access to be granted during active class")
		}
		if !result.IsClassAuth {
			t.Error("Expected IsClassAuth to be true")
		}
		if result.Role != StudioRoleViewer {
			t.Errorf("Expected role %d, got %d", StudioRoleViewer, result.Role)
		}
	})

	// Test 2: Anonymous user denied
	t.Run("AnonymousDenied", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckClassPermission(tx, -1, testRoom.Id)
		})

		if result.Allowed {
			t.Error("Expected anonymous user to be denied")
		}
	})

	// Test 3: Access denied for wrong room
	t.Run("WrongRoom", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckClassPermission(tx, testUser.Id, 99999)
		})

		if result.Allowed {
			t.Error("Expected access to be denied for wrong room")
		}
	})
}

// TestClassPermissionIntegration tests class permissions integrated into CheckRoomAccess
func TestClassPermissionIntegration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var testUser, otherUser User
	var testRoom Room
	var activeSchedule, futureSchedule, pastSchedule ClassSchedule

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		testUser = createTestUser(t, tx, "user@test.com", RoleStreamAdmin)
		otherUser = createTestUser(t, tx, "other@test.com", RoleStreamAdmin)

		// Create studio and room
		studio, room := createTestStudioAndRoom(tx)
		testRoom = room

		now := time.Now()

		// Create active schedule (happening now)
		activeSchedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Active Class",
			IsRecurring:     false,
			StartTime:       now.Add(-30 * time.Minute),
			EndTime:         now.Add(30 * time.Minute),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, activeSchedule.Id, &activeSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, activeSchedule.Id, activeSchedule.RoomId)

		// Create future schedule (starts in 2 hours)
		futureSchedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Future Class",
			IsRecurring:     false,
			StartTime:       now.Add(2 * time.Hour),
			EndTime:         now.Add(3 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, futureSchedule.Id, &futureSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, futureSchedule.Id, futureSchedule.RoomId)

		// Create past schedule (ended 20 minutes ago - outside grace period)
		pastSchedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Past Class",
			IsRecurring:     false,
			StartTime:       now.Add(-2 * time.Hour),
			EndTime:         now.Add(-20 * time.Minute), // Ended 20 min ago
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, pastSchedule.Id, &pastSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, pastSchedule.Id, pastSchedule.RoomId)

		// Grant active class permission to testUser
		activePerm := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: activeSchedule.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleMember),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, activePerm.Id, &activePerm)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, activePerm.Id, activePerm.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, activePerm.Id, activePerm.UserId)

		// Grant future class permission to testUser
		futurePerm := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: futureSchedule.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleViewer),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, futurePerm.Id, &futurePerm)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, futurePerm.Id, futurePerm.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, futurePerm.Id, futurePerm.UserId)

		// Grant past class permission to testUser
		pastPerm := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: pastSchedule.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleAdmin),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, pastPerm.Id, &pastPerm)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, pastPerm.Id, pastPerm.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, pastPerm.Id, pastPerm.UserId)

		vbolt.TxCommit(tx)
	})

	// Test 1: Access granted during active class
	t.Run("AccessDuringActiveClass", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, testUser, testRoom.Id, "")
		})

		if !result.Allowed {
			t.Error("Expected access to be granted during active class")
		}
		if !result.IsClassAuth {
			t.Error("Expected IsClassAuth to be true")
		}
		if result.Role != StudioRoleMember {
			t.Errorf("Expected role %d, got %d", StudioRoleMember, result.Role)
		}
	})

	// Test 2: Access denied before class starts
	t.Run("DeniedBeforeClassStarts", func(t *testing.T) {
		// Create a user with only future class permission
		var futureOnlyUser User
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			futureOnlyUser = createTestUser(t, tx, "future@test.com", RoleStreamAdmin)

			perm := ClassPermission{
				Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
				ScheduleId: futureSchedule.Id,
				UserId:     futureOnlyUser.Id,
				Role:       int(StudioRoleViewer),
				GrantedBy:  testUser.Id,
				GrantedAt:  time.Now(),
			}
			vbolt.Write(tx, ClassPermissionsBkt, perm.Id, &perm)
			vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm.Id, perm.ScheduleId)
			vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm.Id, perm.UserId)

			vbolt.TxCommit(tx)
		})

		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, futureOnlyUser, testRoom.Id, "")
		})

		if result.Allowed {
			t.Error("Expected access to be denied before class starts")
		}
	})

	// Test 3: User without class permission denied
	t.Run("NoPermissionDenied", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckRoomAccess(tx, otherUser, testRoom.Id, "")
		})

		if result.Allowed {
			t.Error("Expected access to be denied for user without permission")
		}
	})
}

// TestClassPermissionGracePeriod tests the 15-minute grace period after class ends
func TestClassPermissionGracePeriod(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var testUser User
	var testRoom Room
	var recentlyEndedSchedule ClassSchedule

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create user
		testUser = createTestUser(t, tx, "user@test.com", RoleStreamAdmin)

		// Create studio and room
		studio, room := createTestStudioAndRoom(tx)
		testRoom = room

		now := time.Now()

		// Create schedule that ended 10 minutes ago (within 15-minute grace period)
		recentlyEndedSchedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Recently Ended Class",
			IsRecurring:     false,
			StartTime:       now.Add(-70 * time.Minute),
			EndTime:         now.Add(-10 * time.Minute), // Ended 10 min ago
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, recentlyEndedSchedule.Id, &recentlyEndedSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, recentlyEndedSchedule.Id, recentlyEndedSchedule.RoomId)

		// Grant permission
		perm := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: recentlyEndedSchedule.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleViewer),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, perm.Id, &perm)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm.Id, perm.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm.Id, perm.UserId)

		vbolt.TxCommit(tx)
	})

	// Test: Access granted within grace period
	t.Run("WithinGracePeriod", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckClassPermission(tx, testUser.Id, testRoom.Id)
		})

		if !result.Allowed {
			t.Error("Expected access to be granted within grace period (10 min after class ended)")
		}
		if !result.IsClassAuth {
			t.Error("Expected IsClassAuth to be true")
		}
	})
}

// TestClassPermissionGracePeriodExpired tests access is denied after grace period expires
func TestClassPermissionGracePeriodExpired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var testUser User
	var testRoom Room
	var expiredSchedule ClassSchedule

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create user
		testUser = createTestUser(t, tx, "user@test.com", RoleStreamAdmin)

		// Create studio and room
		studio, room := createTestStudioAndRoom(tx)
		testRoom = room

		now := time.Now()

		// Create schedule that ended 20 minutes ago (past 15-minute grace period)
		expiredSchedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Expired Class",
			IsRecurring:     false,
			StartTime:       now.Add(-90 * time.Minute),
			EndTime:         now.Add(-20 * time.Minute), // Ended 20 min ago
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, expiredSchedule.Id, &expiredSchedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, expiredSchedule.Id, expiredSchedule.RoomId)

		// Grant permission
		perm := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: expiredSchedule.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleViewer),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, perm.Id, &perm)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm.Id, perm.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm.Id, perm.UserId)

		vbolt.TxCommit(tx)
	})

	// Test: Access denied after grace period expires
	t.Run("GracePeriodExpired", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckClassPermission(tx, testUser.Id, testRoom.Id)
		})

		if result.Allowed {
			t.Error("Expected access to be denied after grace period expires (20 min after class ended)")
		}
	})
}

// TestClassPermissionMultipleSchedules tests handling of multiple overlapping schedules
func TestClassPermissionMultipleSchedules(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var testUser User
	var testRoom Room
	var schedule1, schedule2 ClassSchedule

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create user
		testUser = createTestUser(t, tx, "user@test.com", RoleStreamAdmin)

		// Create studio and room
		studio, room := createTestStudioAndRoom(tx)
		testRoom = room

		now := time.Now()

		// Create first schedule (active now with Viewer role)
		schedule1 = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Class 1",
			IsRecurring:     false,
			StartTime:       now.Add(-30 * time.Minute),
			EndTime:         now.Add(30 * time.Minute),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule1.Id, &schedule1)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule1.Id, schedule1.RoomId)

		// Create second schedule (also active now with Admin role)
		schedule2 = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          testRoom.Id,
			StudioId:        studio.Id,
			Name:            "Class 2",
			IsRecurring:     false,
			StartTime:       now.Add(-20 * time.Minute),
			EndTime:         now.Add(40 * time.Minute),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			CreatedBy:       testUser.Id,
			CreatedAt:       now,
			UpdatedAt:       now,
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule2.Id, &schedule2)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule2.Id, schedule2.RoomId)

		// Grant permissions
		perm1 := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: schedule1.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleViewer),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, perm1.Id, &perm1)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm1.Id, perm1.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm1.Id, perm1.UserId)

		perm2 := ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: schedule2.Id,
			UserId:     testUser.Id,
			Role:       int(StudioRoleAdmin),
			GrantedBy:  testUser.Id,
			GrantedAt:  now,
		}
		vbolt.Write(tx, ClassPermissionsBkt, perm2.Id, &perm2)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm2.Id, perm2.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm2.Id, perm2.UserId)

		vbolt.TxCommit(tx)
	})

	// Test: Access granted with first found permission
	t.Run("MultipleActiveSchedules", func(t *testing.T) {
		var result RoomAccessResult
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			result = CheckClassPermission(tx, testUser.Id, testRoom.Id)
		})

		if !result.Allowed {
			t.Error("Expected access to be granted when user has multiple active class permissions")
		}
		if !result.IsClassAuth {
			t.Error("Expected IsClassAuth to be true")
		}
		// Should return the first matching permission's role
		if result.Role != StudioRoleViewer && result.Role != StudioRoleAdmin {
			t.Errorf("Expected role to be either Viewer or Admin, got %d", result.Role)
		}
	})
}
