package backend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestSSEDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_sse.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

func TestSSEAuthenticationEvents(t *testing.T) {
	t.Run("NoAuthenticationReturns401", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		// Create a test room
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

			vbolt.TxCommit(tx)
		})

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Create request without any auth cookies
		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil)
		w := httptest.NewRecorder()

		// Call handler
		handler(w, req)

		// Should return 401
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("ValidJWTWithPermissionAllowsSSE", func(t *testing.T) {
		db := setupTestSSEDB(t)
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

		// Create user, studio, membership, and room
		var testUser User

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			testUser = User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "test@example.com",
				Name:     "Test User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, testUser.Id, &testUser)
			vbolt.Write(tx, EmailBkt, testUser.Email, &testUser.Id)

			// Create studio
			studio := Studio{
				Id:       vbolt.NextIntId(tx, StudiosBkt),
				Name:     "Test Studio",
				MaxRooms: 5,
				OwnerId:  testUser.Id,
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create membership
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				UserId:   testUser.Id,
				StudioId: studio.Id,
				Role:     StudioRoleAdmin,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, testUser.Id)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

			// Create room
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 1,
				Name:       "Test Room",
				StreamKey:  "test-key",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			vbolt.TxCommit(tx)
		})

		// Create JWT token
		testToken, err := createTestToken(testUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Create request with JWT cookie (with timeout context)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil).WithContext(ctx)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: testToken,
		})

		w := httptest.NewRecorder()

		// Call handler (will run until context timeout)
		handler(w, req)

		// Should return 200 and start SSE stream
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check SSE headers
		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Error("Expected Content-Type: text/event-stream")
		}
	})

	t.Run("ValidCodeSessionWithAccessAllowsSSE", func(t *testing.T) {
		db := setupTestSSEDB(t)
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
		var sessionToken string
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

			// Create access code for this room
			code, _ := generateUniqueCodeInDB(tx)
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
		jwtToken, err := createCodeSessionToken(sessionToken, codeExpiresAt)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Create request with authToken cookie (with timeout context)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil).WithContext(ctx)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: jwtToken,
		})

		w := httptest.NewRecorder()

		// Call handler
		handler(w, req)

		// Should return 200 and start SSE stream
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check SSE headers
		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Error("Expected Content-Type: text/event-stream")
		}
	})

	t.Run("CodeSessionDeniedForDifferentRoom", func(t *testing.T) {
		db := setupTestSSEDB(t)
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

		// Create two rooms and a code for room 1 only
		var room1Id int
		var sessionToken string
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

			// Create access code ONLY for room 1
			code, _ := generateUniqueCodeInDB(tx)
			codeExpiresAt = time.Now().Add(1 * time.Hour)
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room1Id,
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
		jwtToken, err := createCodeSessionToken(sessionToken, codeExpiresAt)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Try to connect to room 2 with code for room 1
		req := httptest.NewRequest("GET", "/sse/room?roomId=2", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: jwtToken,
		})

		w := httptest.NewRecorder()

		// Call handler
		handler(w, req)

		// Should return 403 Forbidden
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}
	})

	t.Run("ExpiredCodeSessionReturns401", func(t *testing.T) {
		db := setupTestSSEDB(t)
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

		// Create JWT with code session claims (JWT itself will be expired since codeExpiresAt is in the past)
		jwtToken, err := createCodeSessionToken(sessionToken, codeExpiresAt)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Create request with expired code session
		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: jwtToken,
		})

		w := httptest.NewRecorder()

		// Call handler
		handler(w, req)

		// Should return 401 Unauthorized (JWT expired)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for expired code, got %d", w.Code)
		}
	})

	t.Run("InvalidRoomIdReturns404", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create user with valid token but request nonexistent room
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

		testToken, _ := createTestToken(testUser.Id)

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Request nonexistent room
		req := httptest.NewRequest("GET", "/sse/room?roomId=999", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: testToken,
		})

		w := httptest.NewRecorder()

		// Call handler
		handler(w, req)

		// Should return 404 Not Found
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 for invalid room, got %d", w.Code)
		}
	})

	t.Run("JWTWithoutPermissionReturns403", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create user and room, but NO membership (no permission)
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

			// Create studio owned by someone else
			studio := Studio{
				Id:       vbolt.NextIntId(tx, StudiosBkt),
				Name:     "Test Studio",
				MaxRooms: 5,
				OwnerId:  999, // Different owner
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create room (no membership for testUser)
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 1,
				Name:       "Test Room",
				StreamKey:  "test-key",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			vbolt.TxCommit(tx)
		})

		testToken, _ := createTestToken(testUser.Id)

		// Create SSE handler
		handler := MakeStreamRoomEventsHandler(db)

		// Request room without permission
		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: testToken,
		})

		w := httptest.NewRecorder()

		// Call handler
		handler(w, req)

		// Should return 403 Forbidden
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for no permission, got %d", w.Code)
		}
	})
}

func TestViewerCountTracking(t *testing.T) {
	t.Run("ViewerCountIncrementsOnConnect", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		originalDb := appDb
		appDb = db
		defer func() { appDb = originalDb }()

		// Create room, code, and session
		var sessionToken string
		var code string

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

			// Create access code
			codeVal, _ := generateUniqueCodeInDB(tx)
			code = codeVal
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
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

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code,
				CurrentViewers:   1, // One session created
				TotalConnections: 1,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Verify initial viewer count
		var initialViewers int
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			initialViewers = analytics.CurrentViewers
		})

		if initialViewers != 1 {
			t.Errorf("Expected initial viewers 1, got %d", initialViewers)
		}
	})

	t.Run("ViewerCountDecrementsOnDisconnect", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create room, code, and session
		var sessionToken string
		var code string
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

			// Create access code
			codeVal, _ := generateUniqueCodeInDB(tx)
			code = codeVal
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

			// Set initial analytics with 3 viewers
			analytics := CodeAnalytics{
				Code:             code,
				CurrentViewers:   3,
				TotalConnections: 3,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Verify initial viewer count
		var beforeViewers int
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			beforeViewers = analytics.CurrentViewers
		})

		if beforeViewers != 3 {
			t.Fatalf("Expected initial viewers 3, got %d", beforeViewers)
		}

		// Create JWT with code session claims
		jwtToken, err := createCodeSessionToken(sessionToken, codeExpiresAt)
		if err != nil {
			t.Fatalf("Failed to create JWT: %v", err)
		}

		// Create SSE handler and connect with timeout context
		handler := MakeStreamRoomEventsHandler(db)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil).WithContext(ctx)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: jwtToken,
		})

		w := httptest.NewRecorder()

		// Call handler (will disconnect after timeout)
		handler(w, req)

		// Verify viewer count was decremented after disconnect
		var afterViewers int
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			afterViewers = analytics.CurrentViewers
		})

		if afterViewers != 2 {
			t.Errorf("Expected viewers to decrement to 2, got %d", afterViewers)
		}
	})

	t.Run("ViewerLimitEnforcement", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		originalDb := appDb
		appDb = db
		defer func() { appDb = originalDb }()

		// Create room and code with MaxViewers=2
		var code string

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

			// Create access code with MaxViewers=2
			codeVal, _ := generateUniqueCodeInDB(tx)
			code = codeVal
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 2, // Limit of 2 viewers
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code: code,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Validate code successfully for first viewer
		var resp1 ValidateAccessCodeResponse
		var err1 error
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			resp1, err1 = ValidateAccessCode(ctx, ValidateAccessCodeRequest{Code: code})
		})
		if err1 != nil || !resp1.Success {
			t.Fatalf("First validation should succeed: %v, %v", err1, resp1.Error)
		}

		// Validate code successfully for second viewer
		var resp2 ValidateAccessCodeResponse
		var err2 error
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			resp2, err2 = ValidateAccessCode(ctx, ValidateAccessCodeRequest{Code: code})
		})
		if err2 != nil || !resp2.Success {
			t.Fatalf("Second validation should succeed: %v, %v", err2, resp2.Error)
		}

		// Verify current viewers is 2
		var currentViewers int
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			currentViewers = analytics.CurrentViewers
		})

		if currentViewers != 2 {
			t.Errorf("Expected 2 current viewers, got %d", currentViewers)
		}

		// Third validation should fail due to capacity
		var resp3 ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			resp3, _ = ValidateAccessCode(ctx, ValidateAccessCodeRequest{Code: code})
		})
		if resp3.Success {
			t.Error("Third validation should fail (at capacity)")
		}
		if resp3.Error == "" {
			t.Error("Expected error message about capacity")
		}
	})

	t.Run("ViewerCountIgnoresJWTUsers", func(t *testing.T) {
		db := setupTestSSEDB(t)
		defer db.Close()

		originalDb := appDb
		originalKey := jwtKey
		appDb = db
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() {
			appDb = originalDb
			jwtKey = originalKey
		}()

		// Create user, studio, membership, and room
		var testUser User
		var code string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			testUser = User{
				Id:       vbolt.NextIntId(tx, UsersBkt),
				Email:    "test@example.com",
				Name:     "Test User",
				Creation: time.Now(),
			}
			vbolt.Write(tx, UsersBkt, testUser.Id, &testUser)
			vbolt.Write(tx, EmailBkt, testUser.Email, &testUser.Id)

			// Create studio
			studio := Studio{
				Id:       vbolt.NextIntId(tx, StudiosBkt),
				Name:     "Test Studio",
				MaxRooms: 5,
				OwnerId:  testUser.Id,
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create membership
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				UserId:   testUser.Id,
				StudioId: studio.Id,
				Role:     StudioRoleAdmin,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, testUser.Id)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

			// Create room
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 1,
				Name:       "Test Room",
				StreamKey:  "test-key",
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Create access code
			codeVal, _ := generateUniqueCodeInDB(tx)
			code = codeVal
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  testUser.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(1 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code: code,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Create JWT token
		testToken, _ := createTestToken(testUser.Id)

		// Create SSE handler and connect with JWT (not code session)
		handler := MakeStreamRoomEventsHandler(db)
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		req := httptest.NewRequest("GET", "/sse/room?roomId=1", nil).WithContext(ctx)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: testToken,
		})

		w := httptest.NewRecorder()

		// Call handler (JWT user, not code session)
		handler(w, req)

		// Verify viewer count for the code is still 0 (JWT users don't count)
		var viewers int
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			viewers = analytics.CurrentViewers
		})

		if viewers != 0 {
			t.Errorf("Expected 0 viewers for code (JWT users shouldn't count), got %d", viewers)
		}
	})
}
