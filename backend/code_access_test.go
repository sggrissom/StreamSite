package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestCodeDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_code.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
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
