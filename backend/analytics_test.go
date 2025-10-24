package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestAnalyticsDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_analytics.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

// TestPackRoomAnalytics verifies RoomAnalytics serialization/deserialization
func TestPackRoomAnalytics(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := RoomAnalytics{
			RoomId:              42,
			TotalViewsAllTime:   1250,
			TotalViewsThisMonth: 350,
			CurrentViewers:      12,
			PeakViewers:         45,
			PeakViewersAt:       time.Now().Truncate(time.Second),
			LastStreamAt:        time.Now().Add(-2 * time.Hour).Truncate(time.Second),
			StreamStartedAt:     time.Now().Add(-30 * time.Minute).Truncate(time.Second),
			TotalStreamMinutes:  18420, // ~307 hours
		}

		// Write and read back
		vbolt.Write(tx, RoomAnalyticsBkt, original.RoomId, &original)
		var retrieved RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, original.RoomId, &retrieved)

		// Verify all fields
		if retrieved.RoomId != original.RoomId {
			t.Errorf("RoomId mismatch: got %d, want %d", retrieved.RoomId, original.RoomId)
		}
		if retrieved.TotalViewsAllTime != original.TotalViewsAllTime {
			t.Errorf("TotalViewsAllTime mismatch: got %d, want %d", retrieved.TotalViewsAllTime, original.TotalViewsAllTime)
		}
		if retrieved.TotalViewsThisMonth != original.TotalViewsThisMonth {
			t.Errorf("TotalViewsThisMonth mismatch: got %d, want %d", retrieved.TotalViewsThisMonth, original.TotalViewsThisMonth)
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
		if !retrieved.LastStreamAt.Equal(original.LastStreamAt) {
			t.Errorf("LastStreamAt mismatch: got %v, want %v", retrieved.LastStreamAt, original.LastStreamAt)
		}
		if !retrieved.StreamStartedAt.Equal(original.StreamStartedAt) {
			t.Errorf("StreamStartedAt mismatch: got %v, want %v", retrieved.StreamStartedAt, original.StreamStartedAt)
		}
		if retrieved.TotalStreamMinutes != original.TotalStreamMinutes {
			t.Errorf("TotalStreamMinutes mismatch: got %d, want %d", retrieved.TotalStreamMinutes, original.TotalStreamMinutes)
		}
	})
}

// TestPackStudioAnalytics verifies StudioAnalytics serialization/deserialization
func TestPackStudioAnalytics(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := StudioAnalytics{
			StudioId:            7,
			TotalViewsAllTime:   5400,
			TotalViewsThisMonth: 1200,
			CurrentViewers:      28,
			TotalRooms:          5,
			ActiveRooms:         2,
			TotalStreamMinutes:  72000, // ~1200 hours
		}

		// Write and read back
		vbolt.Write(tx, StudioAnalyticsBkt, original.StudioId, &original)
		var retrieved StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, original.StudioId, &retrieved)

		// Verify all fields
		if retrieved.StudioId != original.StudioId {
			t.Errorf("StudioId mismatch: got %d, want %d", retrieved.StudioId, original.StudioId)
		}
		if retrieved.TotalViewsAllTime != original.TotalViewsAllTime {
			t.Errorf("TotalViewsAllTime mismatch: got %d, want %d", retrieved.TotalViewsAllTime, original.TotalViewsAllTime)
		}
		if retrieved.TotalViewsThisMonth != original.TotalViewsThisMonth {
			t.Errorf("TotalViewsThisMonth mismatch: got %d, want %d", retrieved.TotalViewsThisMonth, original.TotalViewsThisMonth)
		}
		if retrieved.CurrentViewers != original.CurrentViewers {
			t.Errorf("CurrentViewers mismatch: got %d, want %d", retrieved.CurrentViewers, original.CurrentViewers)
		}
		if retrieved.TotalRooms != original.TotalRooms {
			t.Errorf("TotalRooms mismatch: got %d, want %d", retrieved.TotalRooms, original.TotalRooms)
		}
		if retrieved.ActiveRooms != original.ActiveRooms {
			t.Errorf("ActiveRooms mismatch: got %d, want %d", retrieved.ActiveRooms, original.ActiveRooms)
		}
		if retrieved.TotalStreamMinutes != original.TotalStreamMinutes {
			t.Errorf("TotalStreamMinutes mismatch: got %d, want %d", retrieved.TotalStreamMinutes, original.TotalStreamMinutes)
		}
	})
}

// TestRoomAnalyticsZeroValues verifies zero values are handled correctly
func TestRoomAnalyticsZeroValues(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := RoomAnalytics{
			RoomId: 99,
			// All other fields are zero values
		}

		// Write and read back
		vbolt.Write(tx, RoomAnalyticsBkt, original.RoomId, &original)
		var retrieved RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, original.RoomId, &retrieved)

		// Verify zero values persisted correctly
		if retrieved.RoomId != original.RoomId {
			t.Errorf("RoomId mismatch: got %d, want %d", retrieved.RoomId, original.RoomId)
		}
		if retrieved.TotalViewsAllTime != 0 {
			t.Errorf("TotalViewsAllTime should be 0, got %d", retrieved.TotalViewsAllTime)
		}
		if retrieved.CurrentViewers != 0 {
			t.Errorf("CurrentViewers should be 0, got %d", retrieved.CurrentViewers)
		}
		if !retrieved.PeakViewersAt.IsZero() {
			t.Errorf("PeakViewersAt should be zero time, got %v", retrieved.PeakViewersAt)
		}
	})
}

// TestStudioAnalyticsZeroValues verifies zero values are handled correctly
func TestStudioAnalyticsZeroValues(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		original := StudioAnalytics{
			StudioId: 88,
			// All other fields are zero values
		}

		// Write and read back
		vbolt.Write(tx, StudioAnalyticsBkt, original.StudioId, &original)
		var retrieved StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, original.StudioId, &retrieved)

		// Verify zero values persisted correctly
		if retrieved.StudioId != original.StudioId {
			t.Errorf("StudioId mismatch: got %d, want %d", retrieved.StudioId, original.StudioId)
		}
		if retrieved.TotalViewsAllTime != 0 {
			t.Errorf("TotalViewsAllTime should be 0, got %d", retrieved.TotalViewsAllTime)
		}
		if retrieved.CurrentViewers != 0 {
			t.Errorf("CurrentViewers should be 0, got %d", retrieved.CurrentViewers)
		}
		if retrieved.TotalRooms != 0 {
			t.Errorf("TotalRooms should be 0, got %d", retrieved.TotalRooms)
		}
	})
}
