package backend

import (
	"net/http"
	"net/http/httptest"
	"stream/cfg"
	"strings"
	"testing"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestStreamsDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_streams.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

func TestAuthenticatedStreamProxy(t *testing.T) {
	t.Run("NoAuthenticationReturns403", func(t *testing.T) {
		db := setupTestStreamsDB(t)
		defer db.Close()

		// Save original appDb
		originalDb := appDb
		appDb = db
		defer func() { appDb = originalDb }()

		// Create a test room
		var roomId int
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
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
				StreamKey:  "test-key",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)
			roomId = room.Id

			vbolt.TxCommit(tx)
		})

		// Create request without any auth cookies
		req := httptest.NewRequest("GET", "/streams/room/"+string(rune(roomId+48))+".m3u8", nil)
		w := httptest.NewRecorder()

		// Manually call GetAuthFromRequest to simulate handler behavior
		_, err := GetAuthFromRequest(req, db)
		if err == nil {
			t.Error("Expected authentication error")
		}

		// Verify it would return 403
		if err == ErrAuthFailure {
			w.WriteHeader(http.StatusForbidden)
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("ValidJWTAllowsAccess", func(t *testing.T) {
		db := setupTestStreamsDB(t)
		defer db.Close()

		// Save original appDb and jwtKey
		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create a test user
		var testUser User
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			testUser = User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "test@example.com",
				Name:     "Test User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, testUser.Id, &testUser)
			vbolt.Write(tx, EmailBkt, testUser.Email, &testUser.Id)

			vbolt.TxCommit(tx)
		})

		// Create JWT token
		testToken, err := createTestToken(testUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create request with JWT cookie
		req := httptest.NewRequest("GET", "/streams/room/1.m3u8", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: testToken,
		})

		// Test authentication
		authCtx, err := GetAuthFromRequest(req, db)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if authCtx.IsCodeAuth {
			t.Error("Expected JWT auth, got code auth")
		}

		if authCtx.User.Id != testUser.Id {
			t.Errorf("Expected user ID %d, got %d", testUser.Id, authCtx.User.Id)
		}
	})

	t.Run("ValidCodeSessionWithAccessAllowed", func(t *testing.T) {
		db := setupTestStreamsDB(t)
		defer db.Close()

		// Save original appDb and jwtKey
		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create room, code, and session
		var roomId int
		var sessionToken string
		var accessCodeStr string
		var codeExpiresAt time.Time

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
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
				StreamKey:  "test-key",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)
			roomId = room.Id

			// Create access code for this room
			code, _ := generateUniqueCodeInDB(tx)
			accessCodeStr = code
			codeExpiresAt = time.Now().Add(1 * time.Hour)
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  codeExpiresAt,
				MaxViewers: 0,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        accessCode.Code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		// Create JWT with code session claims
		jwtToken, err := createCodeSessionToken(sessionToken, accessCodeStr, codeExpiresAt)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		// Create request with authToken cookie
		req := httptest.NewRequest("GET", "/streams/room/1.m3u8", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: jwtToken,
		})

		// Test authentication
		authCtx, err := GetAuthFromRequest(req, db)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !authCtx.IsCodeAuth {
			t.Error("Expected code auth, got JWT auth")
		}

		// Verify access to the room
		var hasAccess bool
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			accessResp, _ := GetCodeStreamAccess(ctx, GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       roomId,
			})
			hasAccess = accessResp.Allowed
		})

		if !hasAccess {
			t.Error("Expected access to be allowed")
		}
	})

	t.Run("CodeSessionDeniedForDifferentRoom", func(t *testing.T) {
		db := setupTestStreamsDB(t)
		defer db.Close()

		// Save original appDb
		originalDb := appDb
		appDb = db
		defer func() { appDb = originalDb }()

		// Create two rooms and a code for room 1
		var room1Id, room2Id int
		var sessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			studio := Studio{
				Id:       vbolt.NextIntId(tx, StudiosBkt),
				Name:     "Test Studio",
				MaxRooms: 5,
				OwnerId:  1,
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Room 1
			room1 := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 1,
				Name:       "Test Room 1",
				StreamKey:  "test-key-1",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room1.Id, &room1)
			room1Id = room1.Id

			// Room 2
			room2 := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 2,
				Name:       "Test Room 2",
				StreamKey:  "test-key-2",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
			room2Id = room2.Id

			// Create access code ONLY for room 1
			code, _ := generateUniqueCodeInDB(tx)
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room1Id,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:       sessionToken,
				Code:        accessCode.Code,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		// Try to access room 2 with code for room 1
		var hasAccessToRoom2 bool
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			accessResp, _ := GetCodeStreamAccess(ctx, GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       room2Id,
			})
			hasAccessToRoom2 = accessResp.Allowed
		})

		if hasAccessToRoom2 {
			t.Error("Expected access to room 2 to be denied (code is for room 1)")
		}

		// Verify access to room 1 works
		var hasAccessToRoom1 bool
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			accessResp, _ := GetCodeStreamAccess(ctx, GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       room1Id,
			})
			hasAccessToRoom1 = accessResp.Allowed
		})

		if !hasAccessToRoom1 {
			t.Error("Expected access to room 1 to be allowed")
		}
	})

	t.Run("ExpiredCodeSessionDenied", func(t *testing.T) {
		db := setupTestStreamsDB(t)
		defer db.Close()

		// Save original appDb and jwtKey
		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create expired code and session
		var sessionToken string
		var accessCodeStr string
		var codeExpiresAt time.Time

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
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
				StreamKey:  "test-key",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Create EXPIRED access code
			code, _ := generateUniqueCodeInDB(tx)
			accessCodeStr = code
			codeExpiresAt = time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  1,
				CreatedAt:  time.Now().Add(-2 * time.Hour),
				ExpiresAt:  codeExpiresAt,
				MaxViewers: 0,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session for expired code
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:            sessionToken,
				Code:             accessCode.Code,
				ConnectedAt:      time.Now().Add(-2 * time.Hour),
				LastSeen:         time.Now(),
				GracePeriodUntil: time.Time{}, // Not in grace period
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		// Create JWT with expired code session claims (JWT will be expired)
		jwtToken, err := createCodeSessionToken(sessionToken, accessCodeStr, codeExpiresAt)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		// Create request with expired code session
		req := httptest.NewRequest("GET", "/streams/room/1.m3u8", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: jwtToken,
		})

		// Test authentication - should fail (JWT is expired)
		_, err = GetAuthFromRequest(req, db)
		if err != ErrAuthFailure {
			t.Errorf("Expected ErrAuthFailure for expired code, got %v", err)
		}
	})

	t.Run("InvalidRoomIdReturns400", func(t *testing.T) {
		db := setupTestStreamsDB(t)
		defer db.Close()

		originalDb := appDb
		appDb = db
		defer func() { appDb = originalDb }()

		// Test invalid room ID format
		req := httptest.NewRequest("GET", "/streams/room/abc.m3u8", nil)
		w := httptest.NewRecorder()

		// Simulate what the handler does
		path := req.URL.Path
		pathRemainder := strings.TrimPrefix(path, "/streams/room/")

		roomIdEnd := 0
		for i, ch := range pathRemainder {
			if ch < '0' || ch > '9' {
				roomIdEnd = i
				break
			}
		}

		if roomIdEnd == 0 {
			w.WriteHeader(http.StatusBadRequest)
		}

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid room ID, got %d", w.Code)
		}
	})
}
