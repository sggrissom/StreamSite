package backend

import (
	"net/http"
	"net/http/httptest"
	"stream/cfg"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestAuthDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_auth.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

func TestGetAuthFromRequest(t *testing.T) {
	t.Run("JWTAuthentication", func(t *testing.T) {
		db := setupTestAuthDB(t)
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

		// Create a JWT token for the user
		testToken, err := createTestToken(testUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create a request with JWT cookie
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: testToken,
		})

		// Test GetAuthFromRequest
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

		if authCtx.User.Email != testUser.Email {
			t.Errorf("Expected email %s, got %s", testUser.Email, authCtx.User.Email)
		}

		if authCtx.CodeSession != nil {
			t.Error("Expected nil CodeSession for JWT auth")
		}

		if authCtx.AccessCode != nil {
			t.Error("Expected nil AccessCode for JWT auth")
		}
	})

	t.Run("CodeAuthentication", func(t *testing.T) {
		db := setupTestAuthDB(t)
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

		// Create a valid access code and session
		var validCode string
		var sessionToken string
		var codeExpiresAt time.Time

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create studio
			studio := Studio{
				Id:          vbolt.NextIntId(tx, StudiosBkt),
				Name:        "Test Studio",
				Description: "Test",
				MaxRooms:    5,
				OwnerId:     1,
				Creation:    time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create room
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 1,
				Name:       "Test Room",
				StreamKey:  "test-key",
				IsActive:   false,
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Generate unique code
			code, _ := generateUniqueCodeInDB(tx)
			validCode = code
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
				Label:      "Test Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Create session
			token, _ := generateSessionToken()
			sessionToken = token
			session := CodeSession{
				Token:            sessionToken,
				Code:             accessCode.Code,
				ConnectedAt:      time.Now(),
				LastSeen:         time.Now(),
				GracePeriodUntil: time.Time{},
				ClientIP:         "127.0.0.1",
				UserAgent:        "test-agent",
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		// Create JWT with code session claims
		claims := &Claims{
			IsCodeSession: true,
			SessionToken:  sessionToken,
			Code:          validCode,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(codeExpiresAt),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			t.Fatalf("Failed to create test JWT: %v", err)
		}

		// Create a request with authToken cookie containing the JWT
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: tokenString,
		})

		// Test GetAuthFromRequest
		authCtx, err := GetAuthFromRequest(req, db)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !authCtx.IsCodeAuth {
			t.Error("Expected code auth, got JWT auth")
		}

		if authCtx.User.Id != -1 {
			t.Errorf("Expected pseudo-user ID -1, got %d", authCtx.User.Id)
		}

		if authCtx.User.Email != "anonymous@code-session" {
			t.Errorf("Expected anonymous email, got %s", authCtx.User.Email)
		}

		if authCtx.CodeSession == nil {
			t.Fatal("Expected CodeSession, got nil")
		}

		if authCtx.CodeSession.Token != sessionToken {
			t.Errorf("Expected session token %s, got %s", sessionToken, authCtx.CodeSession.Token)
		}

		if authCtx.CodeSession.Code != validCode {
			t.Errorf("Expected code %s, got %s", validCode, authCtx.CodeSession.Code)
		}

		if authCtx.AccessCode == nil {
			t.Fatal("Expected AccessCode, got nil")
		}

		if authCtx.AccessCode.Code != validCode {
			t.Errorf("Expected access code %s, got %s", validCode, authCtx.AccessCode.Code)
		}
	})

	t.Run("NoAuthentication", func(t *testing.T) {
		db := setupTestAuthDB(t)
		defer db.Close()

		// Create a request with no cookies
		req := httptest.NewRequest("GET", "/test", nil)

		// Test GetAuthFromRequest
		_, err := GetAuthFromRequest(req, db)
		if err != ErrAuthFailure {
			t.Errorf("Expected ErrAuthFailure, got %v", err)
		}
	})

	t.Run("InvalidJWT", func(t *testing.T) {
		db := setupTestAuthDB(t)
		defer db.Close()

		// Create a request with invalid JWT cookie
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: "invalid-jwt-token",
		})

		// Test GetAuthFromRequest
		_, err := GetAuthFromRequest(req, db)
		if err != ErrAuthFailure {
			t.Errorf("Expected ErrAuthFailure, got %v", err)
		}
	})

	t.Run("InvalidCodeSession", func(t *testing.T) {
		db := setupTestAuthDB(t)
		defer db.Close()

		// Save original jwtKey
		originalKey := jwtKey
		jwtKey = []byte("test-secret-key-for-jwt-testing")
		defer func() { jwtKey = originalKey }()

		// Create JWT with invalid/non-existent session token
		claims := &Claims{
			IsCodeSession: true,
			SessionToken:  "invalid-session-token",
			Code:          "12345",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(jwtKey)

		// Create a request with authToken cookie
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: tokenString,
		})

		// Test GetAuthFromRequest
		_, err := GetAuthFromRequest(req, db)
		if err != ErrAuthFailure {
			t.Errorf("Expected ErrAuthFailure, got %v", err)
		}
	})

	t.Run("ExpiredCodeSession", func(t *testing.T) {
		db := setupTestAuthDB(t)
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

		// Create an expired access code and session
		var sessionToken string
		var validCode string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create studio
			studio := Studio{
				Id:          vbolt.NextIntId(tx, StudiosBkt),
				Name:        "Test Studio",
				Description: "Test",
				MaxRooms:    5,
				OwnerId:     1,
				Creation:    time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create room
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 1,
				Name:       "Test Room",
				StreamKey:  "test-key",
				IsActive:   false,
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Generate unique code (expired)
			code, _ := generateUniqueCodeInDB(tx)
			validCode = code
			accessCode := AccessCode{
				Code:       code,
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  1,
				CreatedAt:  time.Now().Add(-2 * time.Hour),
				ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code",
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
				ClientIP:         "127.0.0.1",
				UserAgent:        "test-agent",
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)

			vbolt.TxCommit(tx)
		})

		// Create JWT with code session claims (but code is expired in DB)
		claims := &Claims{
			IsCodeSession: true,
			SessionToken:  sessionToken,
			Code:          validCode,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // JWT not expired, but code is
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(jwtKey)

		// Create a request with authToken cookie
		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  "authToken",
			Value: tokenString,
		})

		// Test GetAuthFromRequest - should succeed even though code is expired
		// (expiration is enforced at stream access level, not auth level)
		authCtx, err := GetAuthFromRequest(req, db)
		if err != nil {
			t.Errorf("Expected authentication to succeed for expired code (expiration checked elsewhere), got error: %v", err)
		}

		if !authCtx.IsCodeAuth {
			t.Error("Expected code auth")
		}

		if authCtx.CodeSession == nil {
			t.Fatal("Expected CodeSession")
		}

		if authCtx.AccessCode == nil {
			t.Fatal("Expected AccessCode")
		}

		// Verify that the access code is indeed expired (test setup is correct)
		if !authCtx.AccessCode.ExpiresAt.Before(time.Now()) {
			t.Error("Test setup error: code should be expired")
		}
	})

}
