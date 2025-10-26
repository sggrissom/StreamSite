package backend

import (
	"net/http"
	"net/http/httptest"
	"stream/cfg"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"golang.org/x/crypto/bcrypt"
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
		testToken, tokenErr := createTestToken(testUser.Id)
		if tokenErr != nil {
			t.Fatalf("Failed to create test token: %v", tokenErr)
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

		// Create JWT with code session claims (userId=-1)
		claims := &Claims{
			UserId:       -1,
			SessionToken: sessionToken,
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
			UserId:       -1,
			SessionToken: "invalid-session-token",
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
			UserId:       -1,
			SessionToken: sessionToken,
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

// TestUnifiedJWT_AnonymousCodeSession tests JWT issuance with userId=-1 for anonymous code users
func TestUnifiedJWT_AnonymousCodeSession(t *testing.T) {
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

	// Create access code and session
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

		// Generate code
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

	// Create JWT with userId=-1 and sessionToken
	claims := &Claims{
		UserId:       -1,
		SessionToken: sessionToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(codeExpiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}

	// Create request with JWT
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

	// Verify userId=-1
	if authCtx.User.Id != -1 {
		t.Errorf("Expected userId=-1, got %d", authCtx.User.Id)
	}

	// Verify IsCodeAuth
	if !authCtx.IsCodeAuth {
		t.Error("Expected IsCodeAuth=true for anonymous code session")
	}

	// Verify CodeSession populated
	if authCtx.CodeSession == nil {
		t.Fatal("Expected CodeSession, got nil")
	}

	if authCtx.CodeSession.Token != sessionToken {
		t.Errorf("Expected session token %s, got %s", sessionToken, authCtx.CodeSession.Token)
	}

	// Verify NO entry in UserCodeSessionsBkt (this is anonymous)
	var storedSessionToken string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, -1, &storedSessionToken)
	})
	if storedSessionToken != "" {
		t.Error("Expected no entry in UserCodeSessionsBkt for anonymous session")
	}
}

// TestUnifiedJWT_LoggedInUserWithCode tests logged-in user adding code access
func TestUnifiedJWT_LoggedInUserWithCode(t *testing.T) {
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

	// Create test user
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

	// Create access code and session
	var sessionToken string
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio/room
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     1,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

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

		// Create code
		code, _ := generateUniqueCodeInDB(tx)
		accessCode := AccessCode{
			Code:       code,
			Type:       CodeTypeRoom,
			TargetId:   room.Id,
			CreatedBy:  1,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
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

		// Store mapping in UserCodeSessionsBkt
		vbolt.Write(tx, UserCodeSessionsBkt, testUser.Id, &sessionToken)

		vbolt.TxCommit(tx)
	})

	// Create JWT with real userId (no sessionToken in claims)
	claims := &Claims{
		UserId: testUser.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}

	// Create request with JWT
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

	// Verify real user populated
	if authCtx.User.Id != testUser.Id {
		t.Errorf("Expected userId=%d, got %d", testUser.Id, authCtx.User.Id)
	}

	if authCtx.User.Email != testUser.Email {
		t.Errorf("Expected email=%s, got %s", testUser.Email, authCtx.User.Email)
	}

	// Verify UserCodeSession populated (user has code access)
	if authCtx.UserCodeSession == nil {
		t.Fatal("Expected UserCodeSession, got nil")
	}

	if authCtx.UserCodeSession.Token != sessionToken {
		t.Errorf("Expected session token %s, got %s", sessionToken, authCtx.UserCodeSession.Token)
	}

	// Verify entry exists in UserCodeSessionsBkt
	var storedSessionToken string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, testUser.Id, &storedSessionToken)
	})
	if storedSessionToken != sessionToken {
		t.Errorf("Expected UserCodeSessionsBkt[%d]=%s, got %s", testUser.Id, sessionToken, storedSessionToken)
	}
}

// TestUnifiedJWT_LogoutClearsCodeSession tests that logout clears UserCodeSessionsBkt
func TestUnifiedJWT_LogoutClearsCodeSession(t *testing.T) {
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

	// Create test user
	var testUser User
	var sessionToken string
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		testUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Email:    "test@example.com",
			Name:     "Test User",
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, testUser.Id, &testUser)
		vbolt.Write(tx, EmailBkt, testUser.Email, &testUser.Id)

		// Add code session mapping
		sessionToken = "test-session-token"
		vbolt.Write(tx, UserCodeSessionsBkt, testUser.Id, &sessionToken)

		vbolt.TxCommit(tx)
	})

	// Verify entry exists before logout
	var storedBefore string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, testUser.Id, &storedBefore)
	})
	if storedBefore != sessionToken {
		t.Fatalf("Test setup error: expected session token in DB before logout")
	}

	// Create JWT for user
	claims := &Claims{
		UserId: testUser.Id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}

	// Create logout request
	req := httptest.NewRequest("POST", "/api/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "authToken",
		Value: tokenString,
	})

	// Create response recorder
	w := httptest.NewRecorder()

	// Call logout handler
	logoutHandler(w, req)

	// Verify UserCodeSessionsBkt entry is cleared
	var storedAfter string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, testUser.Id, &storedAfter)
	})
	if storedAfter != "" {
		t.Errorf("Expected UserCodeSessionsBkt entry to be cleared after logout, got %s", storedAfter)
	}
}

// TestMigrateCodeSession_AnonymousLogin tests that anonymous code session users
// can login and have their session migrated to their user account
func TestMigrateCodeSession_AnonymousLogin(t *testing.T) {
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

	// Create a test user with password
	var testUser User
	testPassword := "password123"
	var passHash []byte
	var err error
	passHash, err = bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		testUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Email:    "test@example.com",
			Name:     "Test User",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, testUser.Id, &testUser)
		vbolt.Write(tx, EmailBkt, testUser.Email, &testUser.Id)
		vbolt.Write(tx, PasswdBkt, testUser.Id, &passHash)
		vbolt.TxCommit(tx)
	})

	// Create an access code and code session
	var code AccessCode
	var sessionToken string
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		code = AccessCode{
			Code:       "12345",
			Type:       CodeTypeRoom,
			TargetId:   1,
			CreatedBy:  testUser.Id,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
			MaxViewers: 0,
			IsRevoked:  false,
		}
		vbolt.Write(tx, AccessCodesBkt, code.Code, &code)

		sessionToken = "test-session-uuid"
		session := CodeSession{
			Token:       sessionToken,
			Code:        code.Code,
			ConnectedAt: time.Now(),
			LastSeen:    time.Now(),
		}
		vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)
		vbolt.TxCommit(tx)
	})

	// Create anonymous JWT with code session
	anonClaims := &Claims{
		UserId:       -1,
		SessionToken: sessionToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	anonToken := jwt.NewWithClaims(jwt.SigningMethodHS256, anonClaims)
	anonTokenString, err := anonToken.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to create anonymous JWT: %v", err)
	}

	// Create login request with anonymous JWT cookie
	loginReq := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(loginReq))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "authToken",
		Value: anonTokenString,
	})

	// Create response recorder
	w := httptest.NewRecorder()

	// Call login handler
	loginHandler(w, req)

	// Verify response is successful
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify UserCodeSessionsBkt has the migrated session
	var migratedSession string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, testUser.Id, &migratedSession)
	})

	if migratedSession != sessionToken {
		t.Errorf("Expected migrated session %s, got %s", sessionToken, migratedSession)
	}

	// Verify new JWT does NOT have sessionToken claim
	cookies := w.Result().Cookies()
	var newAuthToken string
	for _, cookie := range cookies {
		if cookie.Name == "authToken" {
			newAuthToken = cookie.Value
			break
		}
	}

	if newAuthToken == "" {
		t.Fatal("Expected new authToken cookie")
	}

	// Parse new token and verify claims
	newToken, err := jwt.ParseWithClaims(newAuthToken, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		t.Fatalf("Failed to parse new JWT: %v", err)
	}

	newClaims, ok := newToken.Claims.(*Claims)
	if !ok {
		t.Fatal("Failed to extract claims from new JWT")
	}

	if newClaims.UserId != testUser.Id {
		t.Errorf("Expected userId %d, got %d", testUser.Id, newClaims.UserId)
	}

	if newClaims.SessionToken != "" {
		t.Errorf("Expected no sessionToken in new JWT, got %s", newClaims.SessionToken)
	}
}

// TestMigrateCodeSession_AnonymousRegister tests that anonymous code session users
// can create an account and have their session migrated
func TestMigrateCodeSession_AnonymousRegister(t *testing.T) {
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

	// Create an access code and code session
	var code AccessCode
	var sessionToken string
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		code = AccessCode{
			Code:       "12345",
			Type:       CodeTypeRoom,
			TargetId:   1,
			CreatedBy:  1,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
			MaxViewers: 0,
			IsRevoked:  false,
		}
		vbolt.Write(tx, AccessCodesBkt, code.Code, &code)

		sessionToken = "test-session-uuid"
		session := CodeSession{
			Token:       sessionToken,
			Code:        code.Code,
			ConnectedAt: time.Now(),
			LastSeen:    time.Now(),
		}
		vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)
		vbolt.TxCommit(tx)
	})

	// Create anonymous JWT with code session
	anonClaims := &Claims{
		UserId:       -1,
		SessionToken: sessionToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	anonToken := jwt.NewWithClaims(jwt.SigningMethodHS256, anonClaims)
	anonTokenString, err := anonToken.SignedString(jwtKey)
	if err != nil {
		t.Fatalf("Failed to create anonymous JWT: %v", err)
	}

	// Call CreateAccount procedure with anonymous JWT token
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{
			Tx:    tx,
			Token: anonTokenString,
		}
		_, err = CreateAccount(ctx, CreateAccountRequest{
			Name:            "New User",
			Email:           "newuser@example.com",
			Password:        "password123",
			ConfirmPassword: "password123",
		})
	})

	if err != nil {
		t.Fatalf("CreateAccount returned error: %v", err)
	}

	if err != nil {
		t.Fatalf("Expected no error, got error: %s", err.Error())
	}

	// Get new user ID from database
	var newUserId int
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		newUserId = GetUserId(tx, "newuser@example.com")
	})

	if newUserId == 0 {
		t.Fatal("Expected new user to be created")
	}

	// Verify UserCodeSessionsBkt has the migrated session
	var migratedSession string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, newUserId, &migratedSession)
	})

	if migratedSession != sessionToken {
		t.Errorf("Expected migrated session %s, got %s", sessionToken, migratedSession)
	}
}

// TestMigrateCodeSession_NoMigrationNeeded tests that users without code sessions
// can login/register normally without migration
func TestMigrateCodeSession_NoMigrationNeeded(t *testing.T) {
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

	// Create a test user with password
	var testUser User
	testPassword := "password123"
	var passHash []byte
	var err error
	passHash, err = bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		testUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Email:    "test@example.com",
			Name:     "Test User",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, testUser.Id, &testUser)
		vbolt.Write(tx, EmailBkt, testUser.Email, &testUser.Id)
		vbolt.Write(tx, PasswdBkt, testUser.Id, &passHash)
		vbolt.TxCommit(tx)
	})

	// Create login request WITHOUT any JWT cookie
	loginReq := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(loginReq))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call login handler
	loginHandler(w, req)

	// Verify response is successful
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify UserCodeSessionsBkt does NOT have any entry (no migration occurred)
	var storedSession string
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		vbolt.Read(tx, UserCodeSessionsBkt, testUser.Id, &storedSession)
	})

	if storedSession != "" {
		t.Errorf("Expected no UserCodeSessionsBkt entry, got %s", storedSession)
	}

	// Verify JWT was issued correctly
	cookies := w.Result().Cookies()
	var newAuthToken string
	for _, cookie := range cookies {
		if cookie.Name == "authToken" {
			newAuthToken = cookie.Value
			break
		}
	}

	if newAuthToken == "" {
		t.Fatal("Expected new authToken cookie")
	}
}
