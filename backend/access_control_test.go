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
