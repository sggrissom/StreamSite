package backend

import (
	"fmt"
	"stream/cfg"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestCodeDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_code.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

// createTestToken generates a JWT token for testing
func createTestToken(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// TestPackAccessCode verifies AccessCode serialization/deserialization
func TestPackAccessCode(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := AccessCode{
			Code:       "12345",
			Type:       CodeTypeRoom,
			TargetId:   42,
			CreatedBy:  1,
			CreatedAt:  time.Now().Truncate(time.Second),
			ExpiresAt:  time.Now().Add(2 * time.Hour).Truncate(time.Second),
			MaxViewers: 30,
			IsRevoked:  false,
			Label:      "Physics 101",
		}

		// Write and read back
		vbolt.Write(tx, AccessCodesBkt, original.Code, &original)
		var retrieved AccessCode
		vbolt.Read(tx, AccessCodesBkt, original.Code, &retrieved)

		// Verify
		if retrieved.Code != original.Code {
			t.Errorf("Code mismatch: got %s, want %s", retrieved.Code, original.Code)
		}
		if retrieved.Type != original.Type {
			t.Errorf("Type mismatch: got %d, want %d", retrieved.Type, original.Type)
		}
		if retrieved.TargetId != original.TargetId {
			t.Errorf("TargetId mismatch: got %d, want %d", retrieved.TargetId, original.TargetId)
		}
		if retrieved.CreatedBy != original.CreatedBy {
			t.Errorf("CreatedBy mismatch: got %d, want %d", retrieved.CreatedBy, original.CreatedBy)
		}
		if !retrieved.CreatedAt.Equal(original.CreatedAt) {
			t.Errorf("CreatedAt mismatch: got %v, want %v", retrieved.CreatedAt, original.CreatedAt)
		}
		if !retrieved.ExpiresAt.Equal(original.ExpiresAt) {
			t.Errorf("ExpiresAt mismatch: got %v, want %v", retrieved.ExpiresAt, original.ExpiresAt)
		}
		if retrieved.MaxViewers != original.MaxViewers {
			t.Errorf("MaxViewers mismatch: got %d, want %d", retrieved.MaxViewers, original.MaxViewers)
		}
		if retrieved.IsRevoked != original.IsRevoked {
			t.Errorf("IsRevoked mismatch: got %v, want %v", retrieved.IsRevoked, original.IsRevoked)
		}
		if retrieved.Label != original.Label {
			t.Errorf("Label mismatch: got %s, want %s", retrieved.Label, original.Label)
		}
	})
}

// TestPackCodeSession verifies CodeSession serialization/deserialization
func TestPackCodeSession(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := CodeSession{
			Token:            "session-uuid-123",
			Code:             "12345",
			ConnectedAt:      time.Now().Truncate(time.Second),
			LastSeen:         time.Now().Truncate(time.Second),
			GracePeriodUntil: time.Now().Add(15 * time.Minute).Truncate(time.Second),
			ClientIP:         "192.168.1.100",
			UserAgent:        "Mozilla/5.0",
		}

		// Write and read back
		vbolt.Write(tx, CodeSessionsBkt, original.Token, &original)
		var retrieved CodeSession
		vbolt.Read(tx, CodeSessionsBkt, original.Token, &retrieved)

		// Verify
		if retrieved.Token != original.Token {
			t.Errorf("Token mismatch: got %s, want %s", retrieved.Token, original.Token)
		}
		if retrieved.Code != original.Code {
			t.Errorf("Code mismatch: got %s, want %s", retrieved.Code, original.Code)
		}
		if !retrieved.ConnectedAt.Equal(original.ConnectedAt) {
			t.Errorf("ConnectedAt mismatch: got %v, want %v", retrieved.ConnectedAt, original.ConnectedAt)
		}
		if !retrieved.LastSeen.Equal(original.LastSeen) {
			t.Errorf("LastSeen mismatch: got %v, want %v", retrieved.LastSeen, original.LastSeen)
		}
		if !retrieved.GracePeriodUntil.Equal(original.GracePeriodUntil) {
			t.Errorf("GracePeriodUntil mismatch: got %v, want %v", retrieved.GracePeriodUntil, original.GracePeriodUntil)
		}
		if retrieved.ClientIP != original.ClientIP {
			t.Errorf("ClientIP mismatch: got %s, want %s", retrieved.ClientIP, original.ClientIP)
		}
		if retrieved.UserAgent != original.UserAgent {
			t.Errorf("UserAgent mismatch: got %s, want %s", retrieved.UserAgent, original.UserAgent)
		}
	})
}

// TestPackCodeAnalytics verifies CodeAnalytics serialization/deserialization
func TestPackCodeAnalytics(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := CodeAnalytics{
			Code:             "12345",
			TotalConnections: 50,
			CurrentViewers:   12,
			PeakViewers:      25,
			PeakViewersAt:    time.Now().Truncate(time.Second),
			LastConnectionAt: time.Now().Truncate(time.Second),
		}

		// Write and read back
		vbolt.Write(tx, CodeAnalyticsBkt, original.Code, &original)
		var retrieved CodeAnalytics
		vbolt.Read(tx, CodeAnalyticsBkt, original.Code, &retrieved)

		// Verify
		if retrieved.Code != original.Code {
			t.Errorf("Code mismatch: got %s, want %s", retrieved.Code, original.Code)
		}
		if retrieved.TotalConnections != original.TotalConnections {
			t.Errorf("TotalConnections mismatch: got %d, want %d", retrieved.TotalConnections, original.TotalConnections)
		}
		if retrieved.CurrentViewers != original.CurrentViewers {
			t.Errorf("CurrentViewers mismatch: got %d, want %d", retrieved.CurrentViewers, original.CurrentViewers)
		}
		if retrieved.PeakViewers != original.PeakViewers {
			t.Errorf("PeakViewers mismatch: got %d, want %d", retrieved.PeakViewers, original.PeakViewers)
		}
		if !retrieved.PeakViewersAt.Equal(original.PeakViewersAt) {
			t.Errorf("PeakViewersAt mismatch: got %v, want %v", retrieved.PeakViewersAt, original.PeakViewersAt)
		}
		if !retrieved.LastConnectionAt.Equal(original.LastConnectionAt) {
			t.Errorf("LastConnectionAt mismatch: got %v, want %v", retrieved.LastConnectionAt, original.LastConnectionAt)
		}
	})
}

// TestGenerateUniqueCode verifies code generation
func TestGenerateUniqueCode(t *testing.T) {
	for i := 0; i < 100; i++ {
		code, err := GenerateUniqueCode()
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}

		// Check length
		if len(code) != 5 {
			t.Errorf("Code has wrong length: got %d, want 5 (code: %s)", len(code), code)
		}

		// Check range
		if code < "10000" || code > "99999" {
			t.Errorf("Code out of range: %s", code)
		}

		// Check it's not a bad pattern
		if isBadPattern(code) {
			t.Errorf("Generated code has bad pattern: %s", code)
		}
	}
}

// TestIsBadPattern verifies pattern detection
func TestIsBadPattern(t *testing.T) {
	tests := []struct {
		code   string
		isBad  bool
		reason string
	}{
		{"11111", true, "all same digit"},
		{"22222", true, "all same digit"},
		{"12345", true, "sequential ascending"},
		{"23456", true, "sequential ascending"},
		{"54321", true, "sequential descending"},
		{"43210", true, "sequential descending"},
		{"10000", false, "valid code"},
		{"42857", false, "valid code"},
		{"98765", true, "sequential descending"},
		{"13579", false, "non-sequential pattern"},
		{"24680", false, "non-sequential pattern"},
		{"19283", false, "random digits"},
	}

	for _, tt := range tests {
		result := isBadPattern(tt.code)
		if result != tt.isBad {
			t.Errorf("isBadPattern(%s) = %v, want %v (%s)", tt.code, result, tt.isBad, tt.reason)
		}
	}
}

// TestCodeTypeConstants verifies code type enum values
func TestCodeTypeConstants(t *testing.T) {
	if CodeTypeRoom != 0 {
		t.Errorf("CodeTypeRoom should be 0, got %d", CodeTypeRoom)
	}
	if CodeTypeStudio != 1 {
		t.Errorf("CodeTypeStudio should be 1, got %d", CodeTypeStudio)
	}
}

// TestGenerateAccessCode tests the GenerateAccessCode procedure
func TestGenerateAccessCode(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	// Setup test data
	var adminUser, regularUser User
	var studio Studio
	var room Room

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Admin User",
			Email:    "admin@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)
		vbolt.Write(tx, EmailBkt, adminUser.Email, &adminUser.Id)

		// Create regular user
		regularUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Regular User",
			Email:    "user@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, regularUser.Id, &regularUser)
		vbolt.Write(tx, EmailBkt, regularUser.Email, &regularUser.Id)

		// Create studio
		studio = Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "A test studio",
			MaxRooms:    5,
			OwnerId:     adminUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for admin (owner role)
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create studio membership for regular user (viewer role)
		viewerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		viewerMembership := StudioMembership{
			UserId:   regularUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, viewerMembershipId, &viewerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, viewerMembershipId, regularUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, viewerMembershipId, studio.Id)

		// Create room
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  "test-stream-key",
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.TxCommit(tx)
	})

	// Test 1: Generate room code as admin (should succeed)
	t.Run("RoomCodeAsAdmin", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GenerateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeRoom),
				TargetId:        room.Id,
				DurationMinutes: 120,
				MaxViewers:      30,
				Label:           "Physics 101",
			}

			resp, err = GenerateAccessCode(ctx, req)
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.Code == "" {
			t.Error("Expected code to be generated")
		}
		if len(resp.Code) != 5 {
			t.Errorf("Expected 5-digit code, got %s", resp.Code)
		}
		if resp.ExpiresAt.IsZero() {
			t.Error("Expected expiration time to be set")
		}
		if resp.ShareURL != "/watch/"+resp.Code {
			t.Errorf("Expected share URL /watch/%s, got %s", resp.Code, resp.ShareURL)
		}

		// Verify code was saved to database (in a new read transaction)
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var savedCode AccessCode
			vbolt.Read(tx, AccessCodesBkt, resp.Code, &savedCode)
			if savedCode.Code != resp.Code {
				t.Error("Code was not saved to database")
			}
			if savedCode.Type != CodeTypeRoom {
				t.Errorf("Expected type %d, got %d", CodeTypeRoom, savedCode.Type)
			}
			if savedCode.TargetId != room.Id {
				t.Errorf("Expected targetId %d, got %d", room.Id, savedCode.TargetId)
			}
			if savedCode.MaxViewers != 30 {
				t.Errorf("Expected maxViewers 30, got %d", savedCode.MaxViewers)
			}
			if savedCode.Label != "Physics 101" {
				t.Errorf("Expected label 'Physics 101', got '%s'", savedCode.Label)
			}

			// Verify analytics was initialized
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, resp.Code, &analytics)
			if analytics.Code != resp.Code {
				t.Error("Analytics was not initialized")
			}
			if analytics.TotalConnections != 0 {
				t.Errorf("Expected 0 total connections, got %d", analytics.TotalConnections)
			}
		})
	})

	// Test 2: Generate studio code as admin (should succeed)
	t.Run("StudioCodeAsAdmin", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GenerateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeStudio),
				TargetId:        studio.Id,
				DurationMinutes: 60,
				MaxViewers:      0, // unlimited
				Label:           "All Classes Today",
			}

			resp, err = GenerateAccessCode(ctx, req)
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.Code == "" {
			t.Error("Expected code to be generated")
		}
	})

	// Test 3: Generate code as non-admin (should fail)
	t.Run("CodeAsNonAdmin", func(t *testing.T) {
		regularToken, err := createTestToken(regularUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GenerateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: regularToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeRoom),
				TargetId:        room.Id,
				DurationMinutes: 60,
				MaxViewers:      0,
				Label:           "",
			}

			resp, err = GenerateAccessCode(ctx, req)
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.Success {
			t.Error("Expected failure for non-admin user")
		}
		if resp.Error != "Only studio admins can generate access codes" {
			t.Errorf("Expected permission error, got: %s", resp.Error)
		}
	})

	// Test 4: Invalid room ID (should fail)
	t.Run("InvalidRoomId", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GenerateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeRoom),
				TargetId:        99999, // non-existent
				DurationMinutes: 60,
				MaxViewers:      0,
				Label:           "",
			}

			resp, err = GenerateAccessCode(ctx, req)
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.Success {
			t.Error("Expected failure for invalid room ID")
		}
		if resp.Error != "Room not found" {
			t.Errorf("Expected 'Room not found', got: %s", resp.Error)
		}
	})

	// Test 5: Invalid duration (should fail)
	t.Run("InvalidDuration", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GenerateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeRoom),
				TargetId:        room.Id,
				DurationMinutes: 0, // invalid
				MaxViewers:      0,
				Label:           "",
			}

			resp, err = GenerateAccessCode(ctx, req)
		})

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.Success {
			t.Error("Expected failure for invalid duration")
		}
		if resp.Error != "Duration must be greater than 0" {
			t.Errorf("Expected duration error, got: %s", resp.Error)
		}
	})
}

// TestValidateAccessCode tests the ValidateAccessCode procedure
func TestValidateAccessCode(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	// Setup: Create a valid code first
	var validCode, expiredCode, revokedCode, capacityCode AccessCode
	var testRoom Room
	var testStudio Studio

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio
		testStudio = Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     1,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, testStudio.Id, &testStudio)

		// Create room
		testRoom = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   testStudio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  "test-key",
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, testRoom.Id, &testRoom)

		// Create valid code (expires in 1 hour)
		validCode = AccessCode{
			Code:       "12345",
			Type:       CodeTypeRoom,
			TargetId:   testRoom.Id,
			CreatedBy:  1,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
			MaxViewers: 0,
			IsRevoked:  false,
			Label:      "Valid Code",
		}
		vbolt.Write(tx, AccessCodesBkt, validCode.Code, &validCode)

		// Initialize analytics for valid code
		analytics := CodeAnalytics{
			Code:             validCode.Code,
			TotalConnections: 0,
			CurrentViewers:   0,
			PeakViewers:      0,
		}
		vbolt.Write(tx, CodeAnalyticsBkt, validCode.Code, &analytics)

		// Create expired code
		expiredCode = AccessCode{
			Code:       "11111",
			Type:       CodeTypeRoom,
			TargetId:   testRoom.Id,
			CreatedBy:  1,
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			MaxViewers: 0,
			IsRevoked:  false,
			Label:      "Expired Code",
		}
		vbolt.Write(tx, AccessCodesBkt, expiredCode.Code, &expiredCode)

		// Create revoked code
		revokedCode = AccessCode{
			Code:       "22222",
			Type:       CodeTypeRoom,
			TargetId:   testRoom.Id,
			CreatedBy:  1,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
			MaxViewers: 0,
			IsRevoked:  true, // Revoked
			Label:      "Revoked Code",
		}
		vbolt.Write(tx, AccessCodesBkt, revokedCode.Code, &revokedCode)

		// Create code with viewer limit
		capacityCode = AccessCode{
			Code:       "33333",
			Type:       CodeTypeRoom,
			TargetId:   testRoom.Id,
			CreatedBy:  1,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
			MaxViewers: 2, // Limit of 2 viewers
			IsRevoked:  false,
			Label:      "Capacity Code",
		}
		vbolt.Write(tx, AccessCodesBkt, capacityCode.Code, &capacityCode)

		// Initialize analytics with 2 current viewers (at capacity)
		capacityAnalytics := CodeAnalytics{
			Code:             capacityCode.Code,
			TotalConnections: 2,
			CurrentViewers:   2, // At capacity
			PeakViewers:      2,
		}
		vbolt.Write(tx, CodeAnalyticsBkt, capacityCode.Code, &capacityAnalytics)

		vbolt.TxCommit(tx)
	})

	// Test 1: Valid code (should succeed)
	t.Run("ValidCode", func(t *testing.T) {
		var resp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "12345"}
			resp, _ = ValidateAccessCode(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.SessionToken == "" {
			t.Error("Expected session token to be generated")
		}
		if len(resp.SessionToken) != 44 { // base64 encoding of 32 bytes
			t.Errorf("Session token has unexpected length: got %d, want 44 (token: %s)", len(resp.SessionToken), resp.SessionToken)
		}
		if resp.RedirectTo != "/stream/"+fmt.Sprint(testRoom.Id) {
			t.Errorf("Expected redirect to /stream/%d, got %s", testRoom.Id, resp.RedirectTo)
		}
		if resp.Type != int(CodeTypeRoom) {
			t.Errorf("Expected type %d, got %d", CodeTypeRoom, resp.Type)
		}
		if resp.TargetId != testRoom.Id {
			t.Errorf("Expected targetId %d, got %d", testRoom.Id, resp.TargetId)
		}

		// Verify session was created in database
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var session CodeSession
			vbolt.Read(tx, CodeSessionsBkt, resp.SessionToken, &session)
			if session.Token != resp.SessionToken {
				t.Error("Session was not saved to database")
			}
			if session.Code != "12345" {
				t.Errorf("Expected session code 12345, got %s", session.Code)
			}

			// Verify analytics was updated
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, "12345", &analytics)
			if analytics.TotalConnections != 1 {
				t.Errorf("Expected 1 total connection, got %d", analytics.TotalConnections)
			}
			if analytics.CurrentViewers != 1 {
				t.Errorf("Expected 1 current viewer, got %d", analytics.CurrentViewers)
			}
			if analytics.PeakViewers != 1 {
				t.Errorf("Expected peak viewers 1, got %d", analytics.PeakViewers)
			}
		})
	})

	// Test 2: Invalid code (doesn't exist)
	t.Run("InvalidCode", func(t *testing.T) {
		var resp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "99999"}
			resp, _ = ValidateAccessCode(ctx, req)
		})

		if resp.Success {
			t.Error("Expected failure for invalid code")
		}
		if resp.Error != "Invalid code" {
			t.Errorf("Expected 'Invalid code', got: %s", resp.Error)
		}
	})

	// Test 3: Invalid format (not 5 digits)
	t.Run("InvalidFormat", func(t *testing.T) {
		var resp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "123"}
			resp, _ = ValidateAccessCode(ctx, req)
		})

		if resp.Success {
			t.Error("Expected failure for invalid format")
		}
		if resp.Error != "Invalid code format" {
			t.Errorf("Expected 'Invalid code format', got: %s", resp.Error)
		}
	})

	// Test 4: Expired code
	t.Run("ExpiredCode", func(t *testing.T) {
		var resp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "11111"}
			resp, _ = ValidateAccessCode(ctx, req)
		})

		if resp.Success {
			t.Error("Expected failure for expired code")
		}
		if resp.Error != "Code has expired" {
			t.Errorf("Expected 'Code has expired', got: %s", resp.Error)
		}
	})

	// Test 5: Revoked code
	t.Run("RevokedCode", func(t *testing.T) {
		var resp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "22222"}
			resp, _ = ValidateAccessCode(ctx, req)
		})

		if resp.Success {
			t.Error("Expected failure for revoked code")
		}
		if resp.Error != "Code has been revoked" {
			t.Errorf("Expected 'Code has been revoked', got: %s", resp.Error)
		}
	})

	// Test 6: At capacity
	t.Run("AtCapacity", func(t *testing.T) {
		var resp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "33333"}
			resp, _ = ValidateAccessCode(ctx, req)
		})

		if resp.Success {
			t.Error("Expected failure for code at capacity")
		}
		if resp.Error != "Stream is at capacity (2/2 viewers)" {
			t.Errorf("Expected capacity error, got: %s", resp.Error)
		}
	})

	// Test 7: Multiple validations update analytics correctly
	t.Run("MultipleValidations", func(t *testing.T) {
		// Validate the valid code 3 more times
		for i := 0; i < 3; i++ {
			var resp ValidateAccessCodeResponse
			vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
				ctx := &vbeam.Context{Tx: tx}
				req := ValidateAccessCodeRequest{Code: "12345"}
				resp, _ = ValidateAccessCode(ctx, req)
			})

			if !resp.Success {
				t.Fatalf("Validation %d failed: %s", i+1, resp.Error)
			}
		}

		// Check analytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, "12345", &analytics)
			// Should be 4 total now (1 from first test + 3 from this test)
			if analytics.TotalConnections != 4 {
				t.Errorf("Expected 4 total connections, got %d", analytics.TotalConnections)
			}
			if analytics.CurrentViewers != 4 {
				t.Errorf("Expected 4 current viewers, got %d", analytics.CurrentViewers)
			}
			if analytics.PeakViewers != 4 {
				t.Errorf("Expected peak viewers 4, got %d", analytics.PeakViewers)
			}
		})
	})
}

func TestGetCodeStreamAccess(t *testing.T) {
	t.Run("ValidSessionRoomCode", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var roomId int = 1

		// Setup: Create user, studio, room, and code
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create user
			user := User{
				Id:    1,
				Name:  "Test User",
				Email: "test@example.com",
			}
			vbolt.Write(tx, UsersBkt, user.Id, &user)
			vbolt.Write(tx, EmailBkt, user.Email, &user.Id)

			// Create studio
			studio := Studio{
				Id:       1,
				Name:     "Test Studio",
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create membership
			membershipId := vbolt.NextIntId(tx, MembershipBkt)
			membership := StudioMembership{
				UserId:   user.Id,
				StudioId: studio.Id,
				Role:     StudioRoleAdmin,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, user.Id)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

			// Create room
			room := Room{
				Id:       roomId,
				StudioId: studio.Id,
				Name:     "Test Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Create access code
			code := AccessCode{
				Code:       "12345",
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  user.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Test Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)

			// Add to index
			vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, code.Code, roomId)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Use ValidateAccessCode to create session (this is the normal flow)
		var validateResp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "12345"}
			validateResp, _ = ValidateAccessCode(ctx, req)
		})

		if !validateResp.Success {
			t.Fatalf("Failed to validate code: %s", validateResp.Error)
		}
		sessionToken = validateResp.SessionToken

		// Test: Access the room with the session
		var resp GetCodeStreamAccessResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       roomId,
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if !resp.Allowed {
			t.Fatalf("Expected access to be allowed, got: %s", resp.Message)
		}
		if resp.RoomId != roomId {
			t.Errorf("Expected roomId %d, got %d", roomId, resp.RoomId)
		}
		if resp.StudioId != 1 {
			t.Errorf("Expected studioId 1, got %d", resp.StudioId)
		}
		if resp.GracePeriod {
			t.Errorf("Expected not in grace period")
		}
	})

	t.Run("ValidSessionStudioCode", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var studioId int = 1
		var room1Id int = 10
		var room2Id int = 20

		// Setup: Create studio with multiple rooms and studio-wide code
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create studio
			studio := Studio{
				Id:       studioId,
				Name:     "Test Studio",
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

			// Create room 1
			room1 := Room{
				Id:       room1Id,
				StudioId: studioId,
				Name:     "Room 1",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room1.Id, &room1)

			// Create room 2
			room2 := Room{
				Id:       room2Id,
				StudioId: studioId,
				Name:     "Room 2",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room2.Id, &room2)

			// Create studio-wide access code
			code := AccessCode{
				Code:       "99999",
				Type:       CodeTypeStudio,
				TargetId:   studioId,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Studio Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)

			// Add to index
			vbolt.SetTargetSingleTerm(tx, CodesByStudioIdx, code.Code, studioId)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Use ValidateAccessCode to create session
		var validateResp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "99999"}
			validateResp, _ = ValidateAccessCode(ctx, req)
		})

		if !validateResp.Success {
			t.Fatalf("Failed to validate code: %s", validateResp.Error)
		}
		sessionToken = validateResp.SessionToken

		// Test: Access both rooms with studio-wide code
		for _, roomId := range []int{room1Id, room2Id} {
			var resp GetCodeStreamAccessResponse
			vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
				ctx := &vbeam.Context{Tx: tx}
				req := GetCodeStreamAccessRequest{
					SessionToken: sessionToken,
					RoomId:       roomId,
				}
				resp, _ = GetCodeStreamAccess(ctx, req)
			})

			if !resp.Allowed {
				t.Fatalf("Expected access to room %d to be allowed, got: %s", roomId, resp.Message)
			}
			if resp.RoomId != roomId {
				t.Errorf("Expected roomId %d, got %d", roomId, resp.RoomId)
			}
			if resp.StudioId != studioId {
				t.Errorf("Expected studioId %d, got %d", studioId, resp.StudioId)
			}
		}
	})

	t.Run("InvalidSessionToken", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var resp GetCodeStreamAccessResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: "invalid-token-xyz",
				RoomId:       1,
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if resp.Allowed {
			t.Errorf("Expected access to be denied for invalid token")
		}
		if resp.Message != "Invalid session token" {
			t.Errorf("Expected 'Invalid session token' message, got: %s", resp.Message)
		}
	})

	t.Run("RevokedCode", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var roomId int = 1

		// Setup: Create code and session
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create room
			room := Room{
				Id:       roomId,
				StudioId: 1,
				Name:     "Test Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Create access code
			code := AccessCode{
				Code:       "12345",
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false, // Not revoked yet
				Label:      "Test Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)
			vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, code.Code, roomId)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Create session using ValidateAccessCode
		var validateResp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "12345"}
			validateResp, _ = ValidateAccessCode(ctx, req)
		})

		if !validateResp.Success {
			t.Fatalf("Failed to validate code: %s", validateResp.Error)
		}
		sessionToken = validateResp.SessionToken

		// Now revoke the code
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			var code AccessCode
			vbolt.Read(tx, AccessCodesBkt, "12345", &code)
			code.IsRevoked = true
			vbolt.Write(tx, AccessCodesBkt, "12345", &code)
			vbolt.TxCommit(tx)
		})

		// Test: Try to access with revoked code
		var resp GetCodeStreamAccessResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       roomId,
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if resp.Allowed {
			t.Errorf("Expected access to be denied for revoked code")
		}
		if resp.Message != "Access code has been revoked" {
			t.Errorf("Expected 'Access code has been revoked' message, got: %s", resp.Message)
		}
	})

	t.Run("ExpiredCodeGrantsGracePeriod", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var roomId int = 1

		// Setup: Create code and session, then expire the code
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create room
			room := Room{
				Id:       roomId,
				StudioId: 1,
				Name:     "Test Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Create access code that will be expired
			code := AccessCode{
				Code:       "12345",
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  1,
				CreatedAt:  time.Now().Add(-3 * time.Hour),
				ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)
			vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, code.Code, roomId)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Manually create a session (since ValidateAccessCode would reject expired code)
		token, _ := generateSessionToken()
		sessionToken = token
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			session := CodeSession{
				Token:            sessionToken,
				Code:             "12345",
				ConnectedAt:      time.Now().Add(-2 * time.Hour),
				LastSeen:         time.Now().Add(-1 * time.Minute),
				GracePeriodUntil: time.Time{}, // No grace period yet
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)
			vbolt.SetTargetSingleTerm(tx, SessionsByCodeIdx, sessionToken, "12345")
			vbolt.TxCommit(tx)
		})

		// Test: Access should be granted with grace period
		var resp GetCodeStreamAccessResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       roomId,
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if !resp.Allowed {
			t.Fatalf("Expected access to be allowed with grace period, got: %s", resp.Message)
		}
		if !resp.GracePeriod {
			t.Errorf("Expected to be in grace period")
		}
	})

	t.Run("GracePeriodExpired", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var roomId int = 1

		// Setup: Create expired code
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create room
			room := Room{
				Id:       roomId,
				StudioId: 1,
				Name:     "Test Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)

			// Create expired access code
			code := AccessCode{
				Code:       "12345",
				Type:       CodeTypeRoom,
				TargetId:   roomId,
				CreatedBy:  1,
				CreatedAt:  time.Now().Add(-3 * time.Hour),
				ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)
			vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, code.Code, roomId)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Manually create session with expired grace period
		token, _ := generateSessionToken()
		sessionToken = token
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			session := CodeSession{
				Token:            sessionToken,
				Code:             "12345",
				ConnectedAt:      time.Now().Add(-2 * time.Hour),
				LastSeen:         time.Now().Add(-1 * time.Minute),
				GracePeriodUntil: time.Now().Add(-5 * time.Minute), // Grace period expired 5 min ago
			}
			vbolt.Write(tx, CodeSessionsBkt, sessionToken, &session)
			vbolt.SetTargetSingleTerm(tx, SessionsByCodeIdx, sessionToken, "12345")
			vbolt.TxCommit(tx)
		})

		// Test: Access should be denied
		var resp GetCodeStreamAccessResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       roomId,
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if resp.Allowed {
			t.Errorf("Expected access to be denied after grace period expired")
		}
		if resp.Message != "Access code has expired" {
			t.Errorf("Expected 'Access code has expired' message, got: %s", resp.Message)
		}
	})

	t.Run("RoomCodeWrongRoom", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var targetRoomId int = 1
		var wrongRoomId int = 2

		// Setup: Create two rooms and room-specific code
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create target room
			room1 := Room{
				Id:       targetRoomId,
				StudioId: 1,
				Name:     "Target Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room1.Id, &room1)

			// Create other room
			room2 := Room{
				Id:       wrongRoomId,
				StudioId: 1,
				Name:     "Other Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room2.Id, &room2)

			// Create room-specific access code for room 1
			code := AccessCode{
				Code:       "12345",
				Type:       CodeTypeRoom,
				TargetId:   targetRoomId, // Only valid for room 1
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Room 1 Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)
			vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, code.Code, targetRoomId)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Create session using ValidateAccessCode
		var validateResp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "12345"}
			validateResp, _ = ValidateAccessCode(ctx, req)
		})

		if !validateResp.Success {
			t.Fatalf("Failed to validate code: %s", validateResp.Error)
		}
		sessionToken = validateResp.SessionToken

		// Test: Try to access wrong room
		var resp GetCodeStreamAccessResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       wrongRoomId, // Wrong room!
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if resp.Allowed {
			t.Errorf("Expected access to be denied for wrong room")
		}
		if resp.Message != "This code is not valid for this room" {
			t.Errorf("Expected 'This code is not valid for this room' message, got: %s", resp.Message)
		}
	})

	t.Run("StudioCodeWrongStudio", func(t *testing.T) {
		db := setupTestCodeDB(t)
		defer db.Close()

		var sessionToken string
		var studio1Id int = 1
		var studio2Id int = 2
		var room2Id int = 20

		// Setup: Create two studios with studio-wide code for studio 1
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create studio 1
			studio1 := Studio{
				Id:       studio1Id,
				Name:     "Studio 1",
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio1.Id, &studio1)

			// Create studio 2
			studio2 := Studio{
				Id:       studio2Id,
				Name:     "Studio 2",
				Creation: time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, studio2.Id, &studio2)

			// Create room in studio 2
			room2 := Room{
				Id:       room2Id,
				StudioId: studio2Id,
				Name:     "Studio 2 Room",
				Creation: time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room2.Id, &room2)

			// Create studio-wide access code for studio 1
			code := AccessCode{
				Code:       "99999",
				Type:       CodeTypeStudio,
				TargetId:   studio1Id, // Only valid for studio 1
				CreatedBy:  1,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Studio 1 Code",
			}
			vbolt.Write(tx, AccessCodesBkt, code.Code, &code)
			vbolt.SetTargetSingleTerm(tx, CodesByStudioIdx, code.Code, studio1Id)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             code.Code,
				TotalConnections: 0,
				CurrentViewers:   0,
				PeakViewers:      0,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, code.Code, &analytics)

			vbolt.TxCommit(tx)
		})

		// Create session using ValidateAccessCode
		var validateResp ValidateAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: "99999"}
			validateResp, _ = ValidateAccessCode(ctx, req)
		})

		if !validateResp.Success {
			t.Fatalf("Failed to validate code: %s", validateResp.Error)
		}
		sessionToken = validateResp.SessionToken

		// Test: Try to access room in studio 2
		var resp GetCodeStreamAccessResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := GetCodeStreamAccessRequest{
				SessionToken: sessionToken,
				RoomId:       room2Id, // Room in wrong studio!
			}
			resp, _ = GetCodeStreamAccess(ctx, req)
		})

		if resp.Allowed {
			t.Errorf("Expected access to be denied for room in wrong studio")
		}
		if resp.Message != "This code is not valid for this studio" {
			t.Errorf("Expected 'This code is not valid for this studio' message, got: %s", resp.Message)
		}
	})
}

func TestRevokeAccessCode(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	// Setup test data
	var adminUser, regularUser User
	var studio Studio
	var room Room

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Admin User",
			Email:    "admin@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)
		vbolt.Write(tx, EmailBkt, adminUser.Email, &adminUser.Id)

		// Create regular user (viewer role)
		regularUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Regular User",
			Email:    "viewer@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, regularUser.Id, &regularUser)
		vbolt.Write(tx, EmailBkt, regularUser.Email, &regularUser.Id)

		// Create studio
		studio = Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "A test studio",
			MaxRooms:    5,
			OwnerId:     adminUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for admin (Admin role)
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create studio membership for regular user (Viewer role)
		viewerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		viewerMembership := StudioMembership{
			UserId:   regularUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, viewerMembershipId, &viewerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, viewerMembershipId, regularUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, viewerMembershipId, studio.Id)

		// Create room
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  "test-stream-key",
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		vbolt.TxCommit(tx)
	})

	// Test 1: Revoke non-existent code
	t.Run("NonExistentCode", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp RevokeAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := RevokeAccessCodeRequest{
				Code: "00000",
			}
			resp, _ = RevokeAccessCode(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when revoking non-existent code")
		}
		if resp.Error != "Access code not found" {
			t.Errorf("Expected 'Access code not found' error, got: %s", resp.Error)
		}
	})

	// Test 2: Unauthenticated revocation attempt
	t.Run("UnauthenticatedRequest", func(t *testing.T) {
		// First create a code to revoke
		var code string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			accessCode := AccessCode{
				Code:       "11111",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Test Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			code = accessCode.Code
			vbolt.TxCommit(tx)
		})

		var resp RevokeAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: "", // No token
			}

			req := RevokeAccessCodeRequest{
				Code: code,
			}
			resp, _ = RevokeAccessCode(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when revoking without authentication")
		}
		if resp.Error != "Authentication required" {
			t.Errorf("Expected 'Authentication required' error, got: %s", resp.Error)
		}
	})

	// Test 3: Permission denied (viewer trying to revoke)
	t.Run("PermissionDenied", func(t *testing.T) {
		// Create a code
		var code string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			accessCode := AccessCode{
				Code:       "22222",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Test Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			code = accessCode.Code
			vbolt.TxCommit(tx)
		})

		// Try to revoke as regular viewer
		viewerToken, err := createTestToken(regularUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp RevokeAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: viewerToken,
			}

			req := RevokeAccessCodeRequest{
				Code: code,
			}
			resp, _ = RevokeAccessCode(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when viewer tries to revoke code")
		}
		if resp.Error != "Admin permission required" {
			t.Errorf("Expected 'Admin permission required' error, got: %s", resp.Error)
		}
	})

	// Test 4: Successfully revoke valid code with active sessions
	t.Run("SuccessfulRevocation", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create a code using the procedure
		var code string
		var sessionToken1, sessionToken2 string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeRoom),
				TargetId:        room.Id,
				DurationMinutes: 120,
				MaxViewers:      0,
				Label:           "Code to be revoked",
			}
			resp, _ := GenerateAccessCode(ctx, req)
			if !resp.Success {
				t.Fatalf("Failed to generate code: %s", resp.Error)
			}
			code = resp.Code
		})

		// Create two active sessions (each in separate transaction)
		// Session 1
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: code}
			resp, _ := ValidateAccessCode(ctx, req)
			if !resp.Success {
				t.Fatalf("Failed to validate code for session 1: %s", resp.Error)
			}
			sessionToken1 = resp.SessionToken
		})

		// Session 2
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx}
			req := ValidateAccessCodeRequest{Code: code}
			resp, _ := ValidateAccessCode(ctx, req)
			if !resp.Success {
				t.Fatalf("Failed to validate code for session 2: %s", resp.Error)
			}
			sessionToken2 = resp.SessionToken
		})

		// Verify sessions exist and analytics show 2 current viewers
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var session1 CodeSession
			vbolt.Read(tx, CodeSessionsBkt, sessionToken1, &session1)
			if session1.Token == "" {
				t.Errorf("Session 1 should exist before revocation")
			}

			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			if analytics.CurrentViewers != 2 {
				t.Errorf("Expected 2 current viewers, got: %d", analytics.CurrentViewers)
			}
			if analytics.TotalConnections != 2 {
				t.Errorf("Expected 2 total connections, got: %d", analytics.TotalConnections)
			}
		})

		// Revoke the code
		var revokeResp RevokeAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := RevokeAccessCodeRequest{Code: code}
			revokeResp, _ = RevokeAccessCode(ctx, req)
		})

		if !revokeResp.Success {
			t.Fatalf("Expected successful revocation, got error: %s", revokeResp.Error)
		}
		if revokeResp.SessionsKilled != 2 {
			t.Errorf("Expected 2 sessions killed, got: %d", revokeResp.SessionsKilled)
		}

		// Verify code is revoked
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var accessCode AccessCode
			vbolt.Read(tx, AccessCodesBkt, code, &accessCode)
			if !accessCode.IsRevoked {
				t.Errorf("Code should be marked as revoked")
			}
		})

		// Verify sessions are deleted
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var session1 CodeSession
			vbolt.Read(tx, CodeSessionsBkt, sessionToken1, &session1)
			if session1.Token != "" {
				t.Errorf("Session 1 should be deleted after revocation")
			}

			var session2 CodeSession
			vbolt.Read(tx, CodeSessionsBkt, sessionToken2, &session2)
			if session2.Token != "" {
				t.Errorf("Session 2 should be deleted after revocation")
			}
		})

		// Verify analytics updated (CurrentViewers should be 0)
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var analytics CodeAnalytics
			vbolt.Read(tx, CodeAnalyticsBkt, code, &analytics)
			if analytics.CurrentViewers != 0 {
				t.Errorf("Expected 0 current viewers after revocation, got: %d", analytics.CurrentViewers)
			}
			// TotalConnections should remain at 2
			if analytics.TotalConnections != 2 {
				t.Errorf("Expected 2 total connections to remain, got: %d", analytics.TotalConnections)
			}
		})

		// Verify session index is cleared
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var sessionTokens []string
			vbolt.ReadTermTargets(tx, SessionsByCodeIdx, code, &sessionTokens, vbolt.Window{})
			if len(sessionTokens) != 0 {
				t.Errorf("Expected 0 sessions in index after revocation, got: %d", len(sessionTokens))
			}
		})
	})

	// Test 5: Attempt to revoke already-revoked code
	t.Run("AlreadyRevoked", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create and immediately revoke a code
		var code string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			accessCode := AccessCode{
				Code:       "33333",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  true, // Already revoked
				Label:      "Already revoked code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			code = accessCode.Code
			vbolt.TxCommit(tx)
		})

		// Try to revoke again
		var resp RevokeAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := RevokeAccessCodeRequest{Code: code}
			resp, _ = RevokeAccessCode(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when revoking already-revoked code")
		}
		if resp.Error != "Access code is already revoked" {
			t.Errorf("Expected 'Access code is already revoked' error, got: %s", resp.Error)
		}
	})

	// Test 6: Revoke studio code
	t.Run("RevokeStudioCode", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create a studio code
		var code string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GenerateAccessCodeRequest{
				Type:            int(CodeTypeStudio),
				TargetId:        studio.Id,
				DurationMinutes: 120,
				MaxViewers:      0,
				Label:           "Studio code to revoke",
			}
			resp, _ := GenerateAccessCode(ctx, req)
			if !resp.Success {
				t.Fatalf("Failed to generate studio code: %s", resp.Error)
			}
			code = resp.Code
		})

		// Revoke the studio code
		var revokeResp RevokeAccessCodeResponse
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := RevokeAccessCodeRequest{Code: code}
			revokeResp, _ = RevokeAccessCode(ctx, req)
		})

		if !revokeResp.Success {
			t.Fatalf("Expected successful revocation of studio code, got error: %s", revokeResp.Error)
		}

		// Verify code is revoked
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var accessCode AccessCode
			vbolt.Read(tx, AccessCodesBkt, code, &accessCode)
			if !accessCode.IsRevoked {
				t.Errorf("Studio code should be marked as revoked")
			}
		})
	})
}

func TestListAccessCodes(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	// Setup test data
	var adminUser, viewerUser, nonMemberUser User
	var studio Studio
	var room Room

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Admin User",
			Email:    "admin@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)
		vbolt.Write(tx, EmailBkt, adminUser.Email, &adminUser.Id)

		// Create viewer user
		viewerUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Viewer User",
			Email:    "viewer@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, viewerUser.Id, &viewerUser)
		vbolt.Write(tx, EmailBkt, viewerUser.Email, &viewerUser.Id)

		// Create non-member user
		nonMemberUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Non-Member User",
			Email:    "nonmember@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, nonMemberUser.Id, &nonMemberUser)
		vbolt.Write(tx, EmailBkt, nonMemberUser.Email, &nonMemberUser.Id)

		// Create studio
		studio = Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "A test studio",
			MaxRooms:    5,
			OwnerId:     adminUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for admin (Admin role)
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create studio membership for viewer (Viewer role)
		viewerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		viewerMembership := StudioMembership{
			UserId:   viewerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, viewerMembershipId, &viewerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, viewerMembershipId, viewerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, viewerMembershipId, studio.Id)

		// Create room
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  "test-stream-key",
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		vbolt.TxCommit(tx)
	})

	// Test 1: List room codes (empty list)
	t.Run("EmptyList", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeRoom),
				TargetId: room.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if len(resp.Codes) != 0 {
			t.Errorf("Expected empty list, got %d codes", len(resp.Codes))
		}
	})

	// Test 2: Authentication required
	t.Run("AuthenticationRequired", func(t *testing.T) {
		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: "", // No token
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeRoom),
				TargetId: room.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when listing without authentication")
		}
		if resp.Error != "Authentication required" {
			t.Errorf("Expected 'Authentication required' error, got: %s", resp.Error)
		}
	})

	// Test 3: Room not found
	t.Run("RoomNotFound", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeRoom),
				TargetId: 99999, // Non-existent room
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when room not found")
		}
		if resp.Error != "Room not found" {
			t.Errorf("Expected 'Room not found' error, got: %s", resp.Error)
		}
	})

	// Test 4: Studio not found
	t.Run("StudioNotFound", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeStudio),
				TargetId: 99999, // Non-existent studio
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when studio not found")
		}
		if resp.Error != "Studio not found" {
			t.Errorf("Expected 'Studio not found' error, got: %s", resp.Error)
		}
	})

	// Test 5: Permission denied (non-member)
	t.Run("PermissionDenied", func(t *testing.T) {
		nonMemberToken, err := createTestToken(nonMemberUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: nonMemberToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeRoom),
				TargetId: room.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when non-member tries to list codes")
		}
		if resp.Error != "You do not have permission to view access codes for this room" {
			t.Errorf("Expected permission error, got: %s", resp.Error)
		}
	})

	// Create some test codes for the remaining tests
	var code1, code2, code3 string
	adminToken, _ := createTestToken(adminUser.Email)

	// Code 1: Active room code
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{Tx: tx, Token: adminToken}
		req := GenerateAccessCodeRequest{
			Type:            int(CodeTypeRoom),
			TargetId:        room.Id,
			DurationMinutes: 120,
			MaxViewers:      10,
			Label:           "First Code",
		}
		resp, _ := GenerateAccessCode(ctx, req)
		code1 = resp.Code
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	})

	// Code 2: Expired room code
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		accessCode := AccessCode{
			Code:       "11111",
			Type:       CodeTypeRoom,
			TargetId:   room.Id,
			CreatedBy:  adminUser.Id,
			CreatedAt:  time.Now().Add(-3 * time.Hour),
			ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			MaxViewers: 0,
			IsRevoked:  false,
			Label:      "Expired Code",
		}
		vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
		vbolt.SetTargetSingleTerm(tx, CodesByRoomIdx, accessCode.Code, room.Id)

		// Initialize analytics
		analytics := CodeAnalytics{
			Code:             accessCode.Code,
			TotalConnections: 5,
			CurrentViewers:   0,
			PeakViewers:      3,
			PeakViewersAt:    time.Now().Add(-2 * time.Hour),
		}
		vbolt.Write(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)
		code2 = accessCode.Code
		vbolt.TxCommit(tx)
		time.Sleep(10 * time.Millisecond)
	})

	// Code 3: Revoked studio code
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		accessCode := AccessCode{
			Code:       "22222",
			Type:       CodeTypeStudio,
			TargetId:   studio.Id,
			CreatedBy:  adminUser.Id,
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(2 * time.Hour),
			MaxViewers: 0,
			IsRevoked:  true, // Revoked
			Label:      "Revoked Studio Code",
		}
		vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
		vbolt.SetTargetSingleTerm(tx, CodesByStudioIdx, accessCode.Code, studio.Id)

		// Initialize analytics
		analytics := CodeAnalytics{
			Code:             accessCode.Code,
			TotalConnections: 2,
			CurrentViewers:   0,
			PeakViewers:      2,
		}
		vbolt.Write(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)
		code3 = accessCode.Code
		vbolt.TxCommit(tx)
	})

	// Test 6: List room codes as admin
	t.Run("ListRoomCodesAsAdmin", func(t *testing.T) {
		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeRoom),
				TargetId: room.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if len(resp.Codes) != 2 {
			t.Fatalf("Expected 2 room codes, got %d", len(resp.Codes))
		}

		// Verify sorting (newest first)
		if resp.Codes[0].Code != code1 {
			t.Errorf("Expected first code to be %s, got %s", code1, resp.Codes[0].Code)
		}
		if resp.Codes[1].Code != code2 {
			t.Errorf("Expected second code to be %s, got %s", code2, resp.Codes[1].Code)
		}

		// Verify first code details
		if resp.Codes[0].Type != int(CodeTypeRoom) {
			t.Errorf("Expected type Room, got %d", resp.Codes[0].Type)
		}
		if resp.Codes[0].Label != "First Code" {
			t.Errorf("Expected label 'First Code', got %s", resp.Codes[0].Label)
		}
		if resp.Codes[0].IsRevoked {
			t.Errorf("Expected code1 not to be revoked")
		}
		if resp.Codes[0].IsExpired {
			t.Errorf("Expected code1 not to be expired")
		}
		if resp.Codes[0].CurrentViewers != 0 {
			t.Errorf("Expected 0 current viewers, got %d", resp.Codes[0].CurrentViewers)
		}

		// Verify second code is expired
		if !resp.Codes[1].IsExpired {
			t.Errorf("Expected code2 to be expired")
		}
		if resp.Codes[1].TotalViews != 5 {
			t.Errorf("Expected 5 total views, got %d", resp.Codes[1].TotalViews)
		}
	})

	// Test 7: List room codes as viewer
	t.Run("ListRoomCodesAsViewer", func(t *testing.T) {
		viewerToken, err := createTestToken(viewerUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: viewerToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeRoom),
				TargetId: room.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success for viewer, got error: %s", resp.Error)
		}
		if len(resp.Codes) != 2 {
			t.Errorf("Viewer should see same codes as admin, got %d", len(resp.Codes))
		}
	})

	// Test 8: List studio codes
	t.Run("ListStudioCodes", func(t *testing.T) {
		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := ListAccessCodesRequest{
				Type:     int(CodeTypeStudio),
				TargetId: studio.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if len(resp.Codes) != 1 {
			t.Fatalf("Expected 1 studio code, got %d", len(resp.Codes))
		}

		// Verify studio code details
		if resp.Codes[0].Code != code3 {
			t.Errorf("Expected code %s, got %s", code3, resp.Codes[0].Code)
		}
		if resp.Codes[0].Type != int(CodeTypeStudio) {
			t.Errorf("Expected type Studio, got %d", resp.Codes[0].Type)
		}
		if !resp.Codes[0].IsRevoked {
			t.Errorf("Expected code to be revoked")
		}
		if resp.Codes[0].IsExpired {
			t.Errorf("Expected code not to be expired (only revoked)")
		}
		if resp.Codes[0].TotalViews != 2 {
			t.Errorf("Expected 2 total views, got %d", resp.Codes[0].TotalViews)
		}
	})

	// Test 9: Invalid code type
	t.Run("InvalidCodeType", func(t *testing.T) {
		var resp ListAccessCodesResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := ListAccessCodesRequest{
				Type:     99, // Invalid type
				TargetId: room.Id,
			}
			resp, _ = ListAccessCodes(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure for invalid code type")
		}
		if resp.Error != "Invalid code type (must be 0 for room or 1 for studio)" {
			t.Errorf("Expected invalid type error, got: %s", resp.Error)
		}
	})
}

func TestGetCodeAnalytics(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	// Setup test data
	var adminUser, viewerUser, nonMemberUser User
	var studio Studio
	var room Room

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Admin User",
			Email:    "admin@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)
		vbolt.Write(tx, EmailBkt, adminUser.Email, &adminUser.Id)

		// Create viewer user
		viewerUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Viewer User",
			Email:    "viewer@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, viewerUser.Id, &viewerUser)
		vbolt.Write(tx, EmailBkt, viewerUser.Email, &viewerUser.Id)

		// Create non-member user
		nonMemberUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Non-Member User",
			Email:    "nonmember@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, nonMemberUser.Id, &nonMemberUser)
		vbolt.Write(tx, EmailBkt, nonMemberUser.Email, &nonMemberUser.Id)

		// Create studio
		studio = Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "A test studio",
			MaxRooms:    5,
			OwnerId:     adminUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for admin
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create studio membership for viewer
		viewerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		viewerMembership := StudioMembership{
			UserId:   viewerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, viewerMembershipId, &viewerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, viewerMembershipId, viewerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, viewerMembershipId, studio.Id)

		// Create room
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  "test-stream-key",
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		vbolt.TxCommit(tx)
	})

	// Test 1: Code not found
	t.Run("CodeNotFound", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GetCodeAnalyticsRequest{Code: "99999"}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when code not found")
		}
		if resp.Error != "Access code not found" {
			t.Errorf("Expected 'Access code not found' error, got: %s", resp.Error)
		}
	})

	// Test 2: Authentication required
	t.Run("AuthenticationRequired", func(t *testing.T) {
		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: "", // No token
			}

			req := GetCodeAnalyticsRequest{Code: "12345"}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when not authenticated")
		}
		if resp.Error != "Authentication required" {
			t.Errorf("Expected 'Authentication required' error, got: %s", resp.Error)
		}
	})

	// Test 3: Invalid code format
	t.Run("InvalidCodeFormat", func(t *testing.T) {
		adminToken, err := createTestToken(adminUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GetCodeAnalyticsRequest{Code: "123"} // Too short
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure for invalid code format")
		}
		if resp.Error != "Invalid code format" {
			t.Errorf("Expected 'Invalid code format' error, got: %s", resp.Error)
		}
	})

	// Create a test code for remaining tests
	var testCode string
	adminToken, _ := createTestToken(adminUser.Email)

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{Tx: tx, Token: adminToken}
		req := GenerateAccessCodeRequest{
			Type:            int(CodeTypeRoom),
			TargetId:        room.Id,
			DurationMinutes: 120,
			MaxViewers:      10,
			Label:           "Analytics Test Code",
		}
		resp, _ := GenerateAccessCode(ctx, req)
		testCode = resp.Code
	})

	// Test 4: Permission denied (viewer)
	t.Run("PermissionDenied", func(t *testing.T) {
		viewerToken, err := createTestToken(viewerUser.Email)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: viewerToken,
			}

			req := GetCodeAnalyticsRequest{Code: testCode}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if resp.Success {
			t.Errorf("Expected failure when viewer tries to access analytics")
		}
		if resp.Error != "Only studio admins can view code analytics" {
			t.Errorf("Expected admin permission error, got: %s", resp.Error)
		}
	})

	// Test 5: Successful analytics for active code (no sessions)
	t.Run("ActiveCodeNoSessions", func(t *testing.T) {
		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GetCodeAnalyticsRequest{Code: testCode}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.Code != testCode {
			t.Errorf("Expected code %s, got %s", testCode, resp.Code)
		}
		if resp.Type != int(CodeTypeRoom) {
			t.Errorf("Expected type Room, got %d", resp.Type)
		}
		if resp.Label != "Analytics Test Code" {
			t.Errorf("Expected label 'Analytics Test Code', got %s", resp.Label)
		}
		if resp.Status != "active" {
			t.Errorf("Expected status 'active', got %s", resp.Status)
		}
		if resp.TotalConnections != 0 {
			t.Errorf("Expected 0 total connections, got %d", resp.TotalConnections)
		}
		if resp.CurrentViewers != 0 {
			t.Errorf("Expected 0 current viewers, got %d", resp.CurrentViewers)
		}
		if len(resp.Sessions) != 0 {
			t.Errorf("Expected 0 sessions, got %d", len(resp.Sessions))
		}
	})

	// Create some sessions for testing
	var sessionToken1, sessionToken2 string
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{Tx: tx}

		// Session 1
		req := ValidateAccessCodeRequest{Code: testCode}
		resp, _ := ValidateAccessCode(ctx, req)
		sessionToken1 = resp.SessionToken
	})

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{Tx: tx}

		// Session 2
		req := ValidateAccessCodeRequest{Code: testCode}
		resp, _ := ValidateAccessCode(ctx, req)
		sessionToken2 = resp.SessionToken
	})

	// Manually set ClientIP and UserAgent for testing
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		var session1 CodeSession
		vbolt.Read(tx, CodeSessionsBkt, sessionToken1, &session1)
		session1.ClientIP = "192.168.1.100"
		session1.UserAgent = "Mozilla/5.0 Test Browser"
		vbolt.Write(tx, CodeSessionsBkt, sessionToken1, &session1)

		var session2 CodeSession
		vbolt.Read(tx, CodeSessionsBkt, sessionToken2, &session2)
		session2.ClientIP = "10.0.0.50"
		session2.UserAgent = "Mobile Safari Test"
		// Set LastSeen to 11 minutes ago to make it inactive
		session2.LastSeen = time.Now().Add(-11 * time.Minute)
		vbolt.Write(tx, CodeSessionsBkt, sessionToken2, &session2)

		vbolt.TxCommit(tx)
	})

	// Test 6: Analytics with active sessions
	t.Run("ActiveCodeWithSessions", func(t *testing.T) {
		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GetCodeAnalyticsRequest{Code: testCode}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.TotalConnections != 2 {
			t.Errorf("Expected 2 total connections, got %d", resp.TotalConnections)
		}
		if resp.CurrentViewers != 2 {
			t.Errorf("Expected 2 current viewers, got %d", resp.CurrentViewers)
		}
		if len(resp.Sessions) != 2 {
			t.Fatalf("Expected 2 sessions, got %d", len(resp.Sessions))
		}

		// Find sessions by their characteristics (order may vary)
		var activeSession, inactiveSession *SessionInfo
		for i := range resp.Sessions {
			if resp.Sessions[i].ClientIP == "192.168.1.xxx" {
				activeSession = &resp.Sessions[i]
			} else if resp.Sessions[i].ClientIP == "10.0.0.xxx" {
				inactiveSession = &resp.Sessions[i]
			}
		}

		// Verify active session (192.168.1.100)
		if activeSession == nil {
			t.Fatal("Expected to find session with IP 192.168.1.xxx")
		}
		if activeSession.UserAgent != "Mozilla/5.0 Test Browser" {
			t.Errorf("Expected user agent 'Mozilla/5.0 Test Browser', got %s", activeSession.UserAgent)
		}
		if !activeSession.IsActive {
			t.Errorf("Expected active session to have IsActive=true")
		}

		// Verify inactive session (10.0.0.50)
		if inactiveSession == nil {
			t.Fatal("Expected to find session with IP 10.0.0.xxx")
		}
		if inactiveSession.UserAgent != "Mobile Safari Test" {
			t.Errorf("Expected user agent 'Mobile Safari Test', got %s", inactiveSession.UserAgent)
		}
		if inactiveSession.IsActive {
			t.Errorf("Expected inactive session to have IsActive=false (LastSeen > 10 min ago)")
		}
	})

	// Test 7: Expired code
	t.Run("ExpiredCode", func(t *testing.T) {
		// Create an expired code
		var expiredCode string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			accessCode := AccessCode{
				Code:       "88888",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now().Add(-3 * time.Hour),
				ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             accessCode.Code,
				TotalConnections: 10,
				CurrentViewers:   0,
				PeakViewers:      5,
				PeakViewersAt:    time.Now().Add(-2 * time.Hour),
			}
			vbolt.Write(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)
			expiredCode = accessCode.Code
			vbolt.TxCommit(tx)
		})

		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GetCodeAnalyticsRequest{Code: expiredCode}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.Status != "expired" {
			t.Errorf("Expected status 'expired', got %s", resp.Status)
		}
		if resp.TotalConnections != 10 {
			t.Errorf("Expected 10 total connections, got %d", resp.TotalConnections)
		}
		if resp.PeakViewers != 5 {
			t.Errorf("Expected 5 peak viewers, got %d", resp.PeakViewers)
		}
	})

	// Test 8: Revoked code
	t.Run("RevokedCode", func(t *testing.T) {
		// Create a revoked code
		var revokedCode string
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			accessCode := AccessCode{
				Code:       "77777",
				Type:       CodeTypeStudio,
				TargetId:   studio.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  true, // Revoked
				Label:      "Revoked Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)

			// Initialize analytics
			analytics := CodeAnalytics{
				Code:             accessCode.Code,
				TotalConnections: 3,
				CurrentViewers:   0,
				PeakViewers:      2,
			}
			vbolt.Write(tx, CodeAnalyticsBkt, accessCode.Code, &analytics)
			revokedCode = accessCode.Code
			vbolt.TxCommit(tx)
		})

		var resp GetCodeAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{
				Tx:    tx,
				Token: adminToken,
			}

			req := GetCodeAnalyticsRequest{Code: revokedCode}
			resp, _ = GetCodeAnalytics(ctx, req)
		})

		if !resp.Success {
			t.Fatalf("Expected success, got error: %s", resp.Error)
		}
		if resp.Status != "revoked" {
			t.Errorf("Expected status 'revoked', got %s", resp.Status)
		}
		if resp.Type != int(CodeTypeStudio) {
			t.Errorf("Expected type Studio, got %d", resp.Type)
		}
	})

	// Test 9: IP anonymization edge cases
	t.Run("IPAnonymization", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"192.168.1.100", "192.168.1.xxx"},
			{"10.0.0.1", "10.0.0.xxx"},
			{"", ""},
			{"invalid", "xxx"},
			{"2001:db8::1", "2001:db8::xxx"}, // IPv6 keeps the :: notation
		}

		for _, tc := range testCases {
			result := anonymizeIP(tc.input)
			if result != tc.expected {
				t.Errorf("anonymizeIP(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		}
	})
}

func TestValidateCodeSession(t *testing.T) {
	db := setupTestCodeDB(t)
	defer db.Close()

	// Setup test data
	var adminUser User
	var studio Studio
	var room Room

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Admin User",
			Email:    "admin@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)
		vbolt.Write(tx, EmailBkt, adminUser.Email, &adminUser.Id)

		// Create studio
		studio = Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "A test studio",
			MaxRooms:    5,
			OwnerId:     adminUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for admin
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create room
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  "test-stream-key",
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		vbolt.TxCommit(tx)
	})

	// Test 1: Session not found
	t.Run("SessionNotFound", func(t *testing.T) {
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, _, _, msg := ValidateCodeSession(tx, "nonexistent-token")
			if valid {
				t.Errorf("Expected invalid for nonexistent session")
			}
			if msg != "Session not found" {
				t.Errorf("Expected 'Session not found' message, got: %s", msg)
			}
		})
	})

	// Create a test code and session for remaining tests
	var testCode string
	var sessionToken string
	adminToken, _ := createTestToken(adminUser.Email)

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{Tx: tx, Token: adminToken}
		req := GenerateAccessCodeRequest{
			Type:            int(CodeTypeRoom),
			TargetId:        room.Id,
			DurationMinutes: 120,
			MaxViewers:      0,
			Label:           "Test Code",
		}
		resp, _ := GenerateAccessCode(ctx, req)
		testCode = resp.Code
	})

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ctx := &vbeam.Context{Tx: tx}
		req := ValidateAccessCodeRequest{Code: testCode}
		resp, _ := ValidateAccessCode(ctx, req)
		sessionToken = resp.SessionToken
	})

	// Test 2: Valid session for active code
	t.Run("ValidActiveSession", func(t *testing.T) {
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, session, code, msg := ValidateCodeSession(tx, sessionToken)
			if !valid {
				t.Errorf("Expected valid session, got error: %s", msg)
			}
			if session.Token != sessionToken {
				t.Errorf("Expected session token %s, got %s", sessionToken, session.Token)
			}
			if code.Code != testCode {
				t.Errorf("Expected code %s, got %s", testCode, code.Code)
			}
			if msg != "" {
				t.Errorf("Expected empty message for valid session, got: %s", msg)
			}
		})
	})

	// Test 3: Code not found (orphaned session)
	t.Run("CodeNotFound", func(t *testing.T) {
		// Create orphaned session (session with invalid code)
		orphanedToken := "orphaned-token"
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			orphanedSession := CodeSession{
				Token:       orphanedToken,
				Code:        "99999", // Non-existent code
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, orphanedToken, &orphanedSession)
			vbolt.TxCommit(tx)
		})

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, _, _, msg := ValidateCodeSession(tx, orphanedToken)
			if valid {
				t.Errorf("Expected invalid for orphaned session")
			}
			if msg != "Access code not found" {
				t.Errorf("Expected 'Access code not found' message, got: %s", msg)
			}
		})
	})

	// Test 4: Revoked code
	t.Run("RevokedCode", func(t *testing.T) {
		// Create a revoked code with session
		var revokedCode string
		var revokedSessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create revoked code
			accessCode := AccessCode{
				Code:       "55555",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(2 * time.Hour),
				MaxViewers: 0,
				IsRevoked:  true, // Revoked
				Label:      "Revoked Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			revokedCode = accessCode.Code

			// Create session for revoked code
			revokedSessionToken = "revoked-session-token"
			session := CodeSession{
				Token:       revokedSessionToken,
				Code:        revokedCode,
				ConnectedAt: time.Now(),
				LastSeen:    time.Now(),
			}
			vbolt.Write(tx, CodeSessionsBkt, revokedSessionToken, &session)
			vbolt.TxCommit(tx)
		})

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, _, _, msg := ValidateCodeSession(tx, revokedSessionToken)
			if valid {
				t.Errorf("Expected invalid for revoked code")
			}
			if msg != "Access code has been revoked" {
				t.Errorf("Expected 'Access code has been revoked' message, got: %s", msg)
			}
		})
	})

	// Test 5: Expired code (no grace period set)
	t.Run("ExpiredCodeNoGrace", func(t *testing.T) {
		// Create expired code with session
		var expiredCode string
		var expiredSessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create expired code
			accessCode := AccessCode{
				Code:       "66666",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now().Add(-3 * time.Hour),
				ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			expiredCode = accessCode.Code

			// Create session for expired code (no grace period set)
			expiredSessionToken = "expired-session-token"
			session := CodeSession{
				Token:            expiredSessionToken,
				Code:             expiredCode,
				ConnectedAt:      time.Now().Add(-2 * time.Hour),
				LastSeen:         time.Now(),
				GracePeriodUntil: time.Time{}, // No grace period
			}
			vbolt.Write(tx, CodeSessionsBkt, expiredSessionToken, &session)
			vbolt.TxCommit(tx)
		})

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, _, _, msg := ValidateCodeSession(tx, expiredSessionToken)
			if valid {
				t.Errorf("Expected invalid for expired code without grace period")
			}
			if msg != "Access code has expired (grace period available)" {
				t.Errorf("Expected 'Access code has expired (grace period available)' message, got: %s", msg)
			}
		})
	})

	// Test 6: Expired code within grace period
	t.Run("ExpiredCodeWithinGrace", func(t *testing.T) {
		// Create expired code with active grace period
		var expiredCode string
		var graceSessionToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create expired code
			accessCode := AccessCode{
				Code:       "77777",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now().Add(-3 * time.Hour),
				ExpiresAt:  time.Now().Add(-5 * time.Minute), // Expired 5 min ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code with Grace",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			expiredCode = accessCode.Code

			// Create session with active grace period
			graceSessionToken = "grace-session-token"
			session := CodeSession{
				Token:            graceSessionToken,
				Code:             expiredCode,
				ConnectedAt:      time.Now().Add(-2 * time.Hour),
				LastSeen:         time.Now(),
				GracePeriodUntil: time.Now().Add(10 * time.Minute), // Grace period ends in 10 min
			}
			vbolt.Write(tx, CodeSessionsBkt, graceSessionToken, &session)
			vbolt.TxCommit(tx)
		})

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, session, code, msg := ValidateCodeSession(tx, graceSessionToken)
			if !valid {
				t.Errorf("Expected valid for session within grace period, got error: %s", msg)
			}
			if code.Code != expiredCode {
				t.Errorf("Expected code %s, got %s", expiredCode, code.Code)
			}
			if session.Token != graceSessionToken {
				t.Errorf("Expected session token %s, got %s", graceSessionToken, session.Token)
			}
		})
	})

	// Test 7: Expired code past grace period
	t.Run("ExpiredCodePastGrace", func(t *testing.T) {
		// Create expired code with expired grace period
		var expiredCode string
		var expiredGraceToken string

		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			// Create expired code
			accessCode := AccessCode{
				Code:       "88889",
				Type:       CodeTypeRoom,
				TargetId:   room.Id,
				CreatedBy:  adminUser.Id,
				CreatedAt:  time.Now().Add(-3 * time.Hour),
				ExpiresAt:  time.Now().Add(-20 * time.Minute), // Expired 20 min ago
				MaxViewers: 0,
				IsRevoked:  false,
				Label:      "Expired Code Past Grace",
			}
			vbolt.Write(tx, AccessCodesBkt, accessCode.Code, &accessCode)
			expiredCode = accessCode.Code

			// Create session with expired grace period
			expiredGraceToken = "expired-grace-token"
			session := CodeSession{
				Token:            expiredGraceToken,
				Code:             expiredCode,
				ConnectedAt:      time.Now().Add(-2 * time.Hour),
				LastSeen:         time.Now(),
				GracePeriodUntil: time.Now().Add(-5 * time.Minute), // Grace ended 5 min ago
			}
			vbolt.Write(tx, CodeSessionsBkt, expiredGraceToken, &session)
			vbolt.TxCommit(tx)
		})

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, _, _, msg := ValidateCodeSession(tx, expiredGraceToken)
			if valid {
				t.Errorf("Expected invalid for expired grace period")
			}
			if msg != "Access code has expired" {
				t.Errorf("Expected 'Access code has expired' message, got: %s", msg)
			}
		})
	})

	// Test 8: Valid session returns correct data
	t.Run("ReturnsCorrectData", func(t *testing.T) {
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			valid, session, code, _ := ValidateCodeSession(tx, sessionToken)
			if !valid {
				t.Fatalf("Expected valid session")
			}

			// Verify session data
			if session.Code != testCode {
				t.Errorf("Expected session.Code to be %s, got %s", testCode, session.Code)
			}
			if session.Token != sessionToken {
				t.Errorf("Expected session.Token to be %s, got %s", sessionToken, session.Token)
			}

			// Verify code data
			if code.Code != testCode {
				t.Errorf("Expected code.Code to be %s, got %s", testCode, code.Code)
			}
			if code.Type != CodeTypeRoom {
				t.Errorf("Expected code.Type to be Room, got %d", code.Type)
			}
			if code.TargetId != room.Id {
				t.Errorf("Expected code.TargetId to be %d, got %d", room.Id, code.TargetId)
			}
			if code.Label != "Test Code" {
				t.Errorf("Expected code.Label to be 'Test Code', got %s", code.Label)
			}
		})
	})
}
