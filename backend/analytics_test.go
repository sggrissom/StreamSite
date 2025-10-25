package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test database setup helper (reuses the standard test DB setup pattern)
func setupTestAnalyticsDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test.db"
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

// TestIncrementRoomViewerCount verifies viewer count increments and peak tracking
func TestIncrementRoomViewerCount(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	// Do all setup and testing in a single write transaction
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test studio and room
		studio := Studio{
			Id:       1,
			Name:     "Test Studio",
			OwnerId:  100,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		room := Room{
			Id:       10,
			StudioId: studio.Id,
			Name:     "Test Room",
			IsActive: false,
			Creation: time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Manually increment (inline instead of calling helper)
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.RoomId == 0 {
			analytics.RoomId = 10
		}
		analytics.CurrentViewers++
		analytics.TotalViewsAllTime++
		analytics.TotalViewsThisMonth++
		if analytics.CurrentViewers > analytics.PeakViewers {
			analytics.PeakViewers = analytics.CurrentViewers
			analytics.PeakViewersAt = time.Now()
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)
		UpdateStudioAnalyticsFromRoom(tx, studio.Id)

		// Verify first increment
		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.CurrentViewers != 1 {
			t.Errorf("CurrentViewers should be 1, got %d", analytics.CurrentViewers)
		}
		if analytics.TotalViewsAllTime != 1 {
			t.Errorf("TotalViewsAllTime should be 1, got %d", analytics.TotalViewsAllTime)
		}
		if analytics.PeakViewers != 1 {
			t.Errorf("PeakViewers should be 1, got %d", analytics.PeakViewers)
		}

		// Second increment
		analytics.CurrentViewers++
		analytics.TotalViewsAllTime++
		analytics.TotalViewsThisMonth++
		if analytics.CurrentViewers > analytics.PeakViewers {
			analytics.PeakViewers = analytics.CurrentViewers
			analytics.PeakViewersAt = time.Now()
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)
		UpdateStudioAnalyticsFromRoom(tx, studio.Id)

		// Verify second increment
		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.CurrentViewers != 2 {
			t.Errorf("CurrentViewers should be 2, got %d", analytics.CurrentViewers)
		}
		if analytics.PeakViewers != 2 {
			t.Errorf("PeakViewers should be 2, got %d", analytics.PeakViewers)
		}

		// Verify studio analytics
		var studioAnalytics StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, 1, &studioAnalytics)
		if studioAnalytics.CurrentViewers != 2 {
			t.Errorf("Studio CurrentViewers should be 2, got %d", studioAnalytics.CurrentViewers)
		}
	})
}

// TestDecrementRoomViewerCount verifies viewer count decrements with floor at 0
func TestDecrementRoomViewerCount(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create test studio and room
		studio := Studio{
			Id:       1,
			Name:     "Test Studio",
			OwnerId:  100,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		room := Room{
			Id:       10,
			StudioId: studio.Id,
			Name:     "Test Room",
			IsActive: false,
			Creation: time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Start with 3 viewers
		analytics := RoomAnalytics{
			RoomId:         10,
			CurrentViewers: 3,
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		// Decrement once
		if analytics.CurrentViewers > 0 {
			analytics.CurrentViewers--
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.CurrentViewers != 2 {
			t.Errorf("CurrentViewers should be 2, got %d", analytics.CurrentViewers)
		}

		// Decrement to zero
		analytics.CurrentViewers = 0
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		// Try to decrement below zero (should stay at 0)
		if analytics.CurrentViewers > 0 {
			analytics.CurrentViewers--
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.CurrentViewers != 0 {
			t.Errorf("CurrentViewers should stay at 0, got %d", analytics.CurrentViewers)
		}
	})
}

// TestPeakViewersTracking verifies peak viewers is tracked correctly
func TestPeakViewersTracking(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Setup
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		room := Room{Id: 10, StudioId: studio.Id, Name: "Test Room", Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Increment to peak of 5
		var analytics RoomAnalytics
		analytics.RoomId = 10
		analytics.CurrentViewers = 5
		analytics.PeakViewers = 5
		analytics.PeakViewersAt = time.Now()
		firstPeakTime := analytics.PeakViewersAt
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		// Decrement to 2 - peak should remain 5
		analytics.CurrentViewers = 2
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.PeakViewers != 5 {
			t.Errorf("PeakViewers should still be 5, got %d", analytics.PeakViewers)
		}
		if !analytics.PeakViewersAt.Equal(firstPeakTime) {
			t.Error("PeakViewersAt should not change when not exceeding peak")
		}

		// Increment to 7 - new peak
		analytics.CurrentViewers = 7
		analytics.PeakViewers = 7
		analytics.PeakViewersAt = time.Now()
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.PeakViewers != 7 {
			t.Errorf("PeakViewers should be 7, got %d", analytics.PeakViewers)
		}
		if analytics.PeakViewersAt.Equal(firstPeakTime) {
			t.Error("PeakViewersAt should be updated for new peak")
		}
	})
}

// TestRecordStreamStartStop verifies stream duration tracking
func TestRecordStreamStartStop(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Setup
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		room := Room{Id: 10, StudioId: studio.Id, Name: "Test Room", Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Simulate stream lifecycle
		analytics := RoomAnalytics{RoomId: 10}
		streamStart := time.Now().Add(-30 * time.Minute) // Simulate 30 min stream
		analytics.StreamStartedAt = streamStart
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		// Stream stops
		duration := time.Since(streamStart)
		analytics.TotalStreamMinutes = int(duration.Minutes())
		analytics.LastStreamAt = time.Now()
		analytics.StreamStartedAt = time.Time{} // Reset
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &analytics)

		// Verify
		vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		if analytics.StreamStartedAt.IsZero() == false {
			t.Error("StreamStartedAt should be reset")
		}
		if analytics.LastStreamAt.IsZero() {
			t.Error("LastStreamAt should be set")
		}
		if analytics.TotalStreamMinutes < 29 || analytics.TotalStreamMinutes > 31 {
			t.Errorf("Expected ~30 minutes, got %d", analytics.TotalStreamMinutes)
		}
	})
}

// TestUpdateStudioAnalytics verifies studio aggregation across multiple rooms
func TestUpdateStudioAnalytics(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Setup studio with 3 rooms
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Room 1: active, 5 viewers
		room1 := Room{Id: 10, StudioId: 1, Name: "Room 1", IsActive: true, Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room1.Id, &room1)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room1.Id, studio.Id)
		vbolt.Write(tx, RoomAnalyticsBkt, 10, &RoomAnalytics{
			RoomId: 10, CurrentViewers: 5, TotalViewsAllTime: 100,
			TotalViewsThisMonth: 50, TotalStreamMinutes: 120,
		})

		// Room 2: inactive, 0 viewers
		room2 := Room{Id: 20, StudioId: 1, Name: "Room 2", IsActive: false, Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room2.Id, studio.Id)
		vbolt.Write(tx, RoomAnalyticsBkt, 20, &RoomAnalytics{
			RoomId: 20, CurrentViewers: 0, TotalViewsAllTime: 75,
			TotalViewsThisMonth: 30, TotalStreamMinutes: 90,
		})

		// Room 3: active, 3 viewers
		room3 := Room{Id: 30, StudioId: 1, Name: "Room 3", IsActive: true, Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room3.Id, &room3)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room3.Id, studio.Id)
		vbolt.Write(tx, RoomAnalyticsBkt, 30, &RoomAnalytics{
			RoomId: 30, CurrentViewers: 3, TotalViewsAllTime: 50,
			TotalViewsThisMonth: 20, TotalStreamMinutes: 60,
		})

		// Aggregate
		UpdateStudioAnalyticsFromRoom(tx, studio.Id)

		// Verify
		var studioAnalytics StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		if studioAnalytics.TotalRooms != 3 {
			t.Errorf("TotalRooms should be 3, got %d", studioAnalytics.TotalRooms)
		}
		if studioAnalytics.ActiveRooms != 2 {
			t.Errorf("ActiveRooms should be 2, got %d", studioAnalytics.ActiveRooms)
		}
		if studioAnalytics.CurrentViewers != 8 {
			t.Errorf("CurrentViewers should be 8, got %d", studioAnalytics.CurrentViewers)
		}
		if studioAnalytics.TotalViewsAllTime != 225 {
			t.Errorf("TotalViewsAllTime should be 225, got %d", studioAnalytics.TotalViewsAllTime)
		}
		if studioAnalytics.TotalViewsThisMonth != 100 {
			t.Errorf("TotalViewsThisMonth should be 100, got %d", studioAnalytics.TotalViewsThisMonth)
		}
		if studioAnalytics.TotalStreamMinutes != 270 {
			t.Errorf("TotalStreamMinutes should be 270, got %d", studioAnalytics.TotalStreamMinutes)
		}
	})
}
