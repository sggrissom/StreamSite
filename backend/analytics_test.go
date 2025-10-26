package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbeam"
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

// TestIncrementRoomViewerCount verifies session-based view counting
func TestIncrementRoomViewerCount(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	// Setup: Create test studio and room
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
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
		vbolt.TxCommit(tx)
	})

	t.Run("NewViewer", func(t *testing.T) {
		// First connection from a new viewer
		IncrementRoomViewerCount(db, 10, "user:1")

		// Verify all counters incremented including unique viewers
		var analytics RoomAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		})

		if analytics.CurrentViewers != 1 {
			t.Errorf("CurrentViewers should be 1, got %d", analytics.CurrentViewers)
		}
		if analytics.TotalViewsAllTime != 1 {
			t.Errorf("TotalViewsAllTime should be 1, got %d", analytics.TotalViewsAllTime)
		}
		if analytics.TotalViewsThisMonth != 1 {
			t.Errorf("TotalViewsThisMonth should be 1, got %d", analytics.TotalViewsThisMonth)
		}
		if analytics.UniqueViewersAllTime != 1 {
			t.Errorf("UniqueViewersAllTime should be 1, got %d", analytics.UniqueViewersAllTime)
		}
		if analytics.UniqueViewersThisMonth != 1 {
			t.Errorf("UniqueViewersThisMonth should be 1, got %d", analytics.UniqueViewersThisMonth)
		}
		if analytics.PeakViewers != 1 {
			t.Errorf("PeakViewers should be 1, got %d", analytics.PeakViewers)
		}

		// Verify session was created
		var session ViewerSession
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, ViewerSessionsBkt, "user:1:10", &session)
		})
		if session.SessionKey != "user:1:10" {
			t.Errorf("Session should exist with key 'user:1:10', got '%s'", session.SessionKey)
		}
		if session.ViewerId != "user:1" {
			t.Errorf("Session viewerId should be 'user:1', got '%s'", session.ViewerId)
		}
		if session.RoomId != 10 {
			t.Errorf("Session roomId should be 10, got %d", session.RoomId)
		}
	})

	t.Run("ReconnectWithinTimeout", func(t *testing.T) {
		// Same viewer reconnects within 5 minutes (simulated by immediate reconnect)
		// Decrement first to simulate disconnect
		DecrementRoomViewerCount(db, 10)

		// Get current totals before reconnection
		var beforeAnalytics RoomAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, RoomAnalyticsBkt, 10, &beforeAnalytics)
		})

		// Reconnect same viewer
		IncrementRoomViewerCount(db, 10, "user:1")

		// Verify totals did NOT increase (session still valid)
		var afterAnalytics RoomAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, RoomAnalyticsBkt, 10, &afterAnalytics)
		})

		if afterAnalytics.TotalViewsAllTime != beforeAnalytics.TotalViewsAllTime {
			t.Errorf("TotalViewsAllTime should not change on reconnect within timeout, was %d, now %d",
				beforeAnalytics.TotalViewsAllTime, afterAnalytics.TotalViewsAllTime)
		}
		if afterAnalytics.UniqueViewersAllTime != beforeAnalytics.UniqueViewersAllTime {
			t.Errorf("UniqueViewersAllTime should not change on reconnect, was %d, now %d",
				beforeAnalytics.UniqueViewersAllTime, afterAnalytics.UniqueViewersAllTime)
		}
		if afterAnalytics.CurrentViewers != 1 {
			t.Errorf("CurrentViewers should be 1, got %d", afterAnalytics.CurrentViewers)
		}
	})

	t.Run("ReconnectAfterTimeout", func(t *testing.T) {
		// Simulate session timeout by manually updating LastSeenAt
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			var session ViewerSession
			vbolt.Read(tx, ViewerSessionsBkt, "user:1:10", &session)
			session.LastSeenAt = time.Now().Add(-10 * time.Minute) // 10 minutes ago
			vbolt.Write(tx, ViewerSessionsBkt, "user:1:10", &session)
			vbolt.TxCommit(tx)
		})

		// Get totals before expired reconnection
		var beforeAnalytics RoomAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, RoomAnalyticsBkt, 10, &beforeAnalytics)
		})

		// Decrement and reconnect after timeout
		DecrementRoomViewerCount(db, 10)
		IncrementRoomViewerCount(db, 10, "user:1")

		// Verify TotalViews increased but UniqueViewers did NOT
		var afterAnalytics RoomAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, RoomAnalyticsBkt, 10, &afterAnalytics)
		})

		if afterAnalytics.TotalViewsAllTime != beforeAnalytics.TotalViewsAllTime+1 {
			t.Errorf("TotalViewsAllTime should increase by 1 after timeout, was %d, now %d",
				beforeAnalytics.TotalViewsAllTime, afterAnalytics.TotalViewsAllTime)
		}
		if afterAnalytics.UniqueViewersAllTime != beforeAnalytics.UniqueViewersAllTime {
			t.Errorf("UniqueViewersAllTime should NOT change (already counted), was %d, now %d",
				beforeAnalytics.UniqueViewersAllTime, afterAnalytics.UniqueViewersAllTime)
		}
	})

	t.Run("MultipleUniqueViewers", func(t *testing.T) {
		// Add second viewer
		IncrementRoomViewerCount(db, 10, "user:2")

		// Add third viewer
		IncrementRoomViewerCount(db, 10, "code:abc123")

		// Verify analytics
		var analytics RoomAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, RoomAnalyticsBkt, 10, &analytics)
		})

		if analytics.CurrentViewers != 3 {
			t.Errorf("CurrentViewers should be 3, got %d", analytics.CurrentViewers)
		}
		if analytics.UniqueViewersAllTime != 3 {
			t.Errorf("UniqueViewersAllTime should be 3 (user:1, user:2, code:abc123), got %d",
				analytics.UniqueViewersAllTime)
		}

		// Verify sessions exist
		var session2, session3 ViewerSession
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, ViewerSessionsBkt, "user:2:10", &session2)
			vbolt.Read(tx, ViewerSessionsBkt, "code:abc123:10", &session3)
		})

		if session2.ViewerId != "user:2" {
			t.Errorf("Session2 viewerId should be 'user:2', got '%s'", session2.ViewerId)
		}
		if session3.ViewerId != "code:abc123" {
			t.Errorf("Session3 viewerId should be 'code:abc123', got '%s'", session3.ViewerId)
		}

		// Verify studio analytics aggregation
		var studioAnalytics StudioAnalytics
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			vbolt.Read(tx, StudioAnalyticsBkt, 1, &studioAnalytics)
		})

		if studioAnalytics.UniqueViewersAllTime != 3 {
			t.Errorf("Studio UniqueViewersAllTime should be 3, got %d", studioAnalytics.UniqueViewersAllTime)
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

// TestGetRoomAnalytics tests the GetRoomAnalytics API procedure
func TestGetRoomAnalytics(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	// Setup test data
	var ownerUser, viewerUser, nonMemberUser User
	var studio Studio
	var room Room

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner user
		ownerUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Owner User",
			Email:    "owner@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, ownerUser.Id, &ownerUser)
		vbolt.Write(tx, EmailBkt, ownerUser.Email, &ownerUser.Id)

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
			OwnerId:     ownerUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for owner
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		ownerMembership := StudioMembership{
			UserId:   ownerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, ownerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

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
			IsActive:   true,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create room analytics
		analytics := RoomAnalytics{
			RoomId:              room.Id,
			TotalViewsAllTime:   150,
			TotalViewsThisMonth: 45,
			CurrentViewers:      7,
			PeakViewers:         12,
			PeakViewersAt:       time.Now().Add(-24 * time.Hour),
			LastStreamAt:        time.Now().Add(-1 * time.Hour),
			TotalStreamMinutes:  300,
		}
		vbolt.Write(tx, RoomAnalyticsBkt, room.Id, &analytics)

		vbolt.TxCommit(tx)
	})

	// Test 1: Unauthenticated request
	t.Run("Unauthenticated", func(t *testing.T) {
		var err error
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: ""}
			req := GetRoomAnalyticsRequest{RoomId: room.Id}
			_, err = GetRoomAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure for unauthenticated request")
		}
		if err != nil && err.Error() != "Authentication required" {
			t.Errorf("Expected 'Authentication required' error, got: %s", err.Error())
		}
	})

	// Test 2: Invalid room ID
	t.Run("InvalidRoomId", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetRoomAnalyticsRequest{RoomId: -1}
			_, err = GetRoomAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure for invalid room ID")
		}
		if err != nil && err.Error() != "Invalid room ID" {
			t.Errorf("Expected 'Invalid room ID' error, got: %s", err.Error())
		}
	})

	// Test 3: Room not found
	t.Run("RoomNotFound", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetRoomAnalyticsRequest{RoomId: 99999}
			_, err = GetRoomAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure when room not found")
		}
		if err != nil && err.Error() != "Room not found" {
			t.Errorf("Expected 'Room not found' error, got: %s", err.Error())
		}
	})

	// Test 4: Access denied (non-member)
	t.Run("AccessDenied", func(t *testing.T) {
		token, err := createTestToken(nonMemberUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetRoomAnalyticsRequest{RoomId: room.Id}
			_, err = GetRoomAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure for non-member access")
		}
		if err != nil && err.Error() != "Access denied" {
			t.Errorf("Expected 'Access denied' error, got: %s", err.Error())
		}
	})

	// Test 5: Success - studio owner
	t.Run("SuccessOwner", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetRoomAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetRoomAnalyticsRequest{RoomId: room.Id}
			resp, err = GetRoomAnalytics(ctx, req)
		})

		if err != nil {
			t.Errorf("Expected no error, got error: %s", err.Error())
		}
		if resp.Analytics == nil {
			t.Fatal("Expected analytics to be returned")
		}
		if resp.Analytics.RoomId != room.Id {
			t.Errorf("Expected RoomId %d, got %d", room.Id, resp.Analytics.RoomId)
		}
		if resp.Analytics.CurrentViewers != 7 {
			t.Errorf("Expected CurrentViewers 7, got %d", resp.Analytics.CurrentViewers)
		}
		if resp.Analytics.TotalViewsAllTime != 150 {
			t.Errorf("Expected TotalViewsAllTime 150, got %d", resp.Analytics.TotalViewsAllTime)
		}
		if resp.Analytics.PeakViewers != 12 {
			t.Errorf("Expected PeakViewers 12, got %d", resp.Analytics.PeakViewers)
		}
		if !resp.IsStreaming {
			t.Errorf("Expected IsStreaming to be true")
		}
		if resp.RoomName != "Test Room" {
			t.Errorf("Expected RoomName 'Test Room', got %s", resp.RoomName)
		}
	})

	// Test 6: Success - viewer role
	t.Run("SuccessViewer", func(t *testing.T) {
		token, err := createTestToken(viewerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetRoomAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetRoomAnalyticsRequest{RoomId: room.Id}
			resp, err = GetRoomAnalytics(ctx, req)
		})

		if err != nil {
			t.Errorf("Expected no error for viewer role, got error: %s", err.Error())
		}
		if resp.Analytics == nil {
			t.Fatal("Expected analytics to be returned")
		}
	})

	// Test 7: New room with no analytics
	t.Run("NoAnalytics", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create a new room without analytics
		var newRoom Room
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			newRoom = Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: 2,
				Name:       "New Room",
				StreamKey:  "new-stream-key",
				IsActive:   false,
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, newRoom.Id, &newRoom)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, newRoom.Id, studio.Id)
			vbolt.TxCommit(tx)
		})

		var resp GetRoomAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetRoomAnalyticsRequest{RoomId: newRoom.Id}
			resp, err = GetRoomAnalytics(ctx, req)
		})

		if err != nil {
			t.Errorf("Expected no error, got error: %s", err.Error())
		}
		if resp.Analytics == nil {
			t.Fatal("Expected analytics to be returned")
		}
		if resp.Analytics.RoomId != newRoom.Id {
			t.Errorf("Expected RoomId %d, got %d", newRoom.Id, resp.Analytics.RoomId)
		}
		if resp.Analytics.CurrentViewers != 0 {
			t.Errorf("Expected CurrentViewers 0 for new room, got %d", resp.Analytics.CurrentViewers)
		}
	})
}

// TestGetStudioAnalytics tests the GetStudioAnalytics API procedure
func TestGetStudioAnalytics(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	// Setup test data
	var ownerUser, viewerUser, nonMemberUser User
	var studio Studio

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner user
		ownerUser = User{
			Id:       vbolt.NextIntId(tx, UsersBkt),
			Name:     "Owner User",
			Email:    "owner@test.com",
			Role:     RoleUser,
			Creation: time.Now(),
		}
		vbolt.Write(tx, UsersBkt, ownerUser.Id, &ownerUser)
		vbolt.Write(tx, EmailBkt, ownerUser.Email, &ownerUser.Id)

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
			OwnerId:     ownerUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create studio membership for owner
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		ownerMembership := StudioMembership{
			UserId:   ownerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, ownerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

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

		// Create multiple rooms with analytics
		for i := 1; i <= 3; i++ {
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: i,
				Name:       "Room " + string(rune('0'+i)),
				StreamKey:  "stream-key-" + string(rune('0'+i)),
				IsActive:   i <= 2, // First 2 rooms are active
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

			analytics := RoomAnalytics{
				RoomId:              room.Id,
				TotalViewsAllTime:   100 * i,
				TotalViewsThisMonth: 30 * i,
				CurrentViewers:      i * 2,
				PeakViewers:         i * 5,
				TotalStreamMinutes:  60 * i,
			}
			vbolt.Write(tx, RoomAnalyticsBkt, room.Id, &analytics)
		}

		// Create pre-aggregated studio analytics
		studioAnalytics := StudioAnalytics{
			StudioId:            studio.Id,
			TotalViewsAllTime:   600,
			TotalViewsThisMonth: 180,
			CurrentViewers:      12,
			TotalRooms:          3,
			ActiveRooms:         2,
			TotalStreamMinutes:  360,
		}
		vbolt.Write(tx, StudioAnalyticsBkt, studio.Id, &studioAnalytics)

		vbolt.TxCommit(tx)
	})

	// Test 1: Unauthenticated request
	t.Run("Unauthenticated", func(t *testing.T) {
		var err error
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: ""}
			req := GetStudioAnalyticsRequest{StudioId: studio.Id}
			_, err = GetStudioAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure for unauthenticated request")
		}
		if err != nil && err.Error() != "Authentication required" {
			t.Errorf("Expected 'Authentication required' error, got: %s", err.Error())
		}
	})

	// Test 2: Invalid studio ID
	t.Run("InvalidStudioId", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetStudioAnalyticsRequest{StudioId: -1}
			_, err = GetStudioAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure for invalid studio ID")
		}
		if err != nil && err.Error() != "Invalid studio ID" {
			t.Errorf("Expected 'Invalid studio ID' error, got: %s", err.Error())
		}
	})

	// Test 3: Access denied (non-member)
	t.Run("AccessDenied", func(t *testing.T) {
		token, err := createTestToken(nonMemberUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetStudioAnalyticsRequest{StudioId: studio.Id}
			_, err = GetStudioAnalytics(ctx, req)
		})

		if err == nil {
			t.Errorf("Expected failure for non-member access")
		}
		if err != nil && err.Error() != "Access denied" {
			t.Errorf("Expected 'Access denied' error, got: %s", err.Error())
		}
	})

	// Test 4: Success - studio owner
	t.Run("SuccessOwner", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetStudioAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetStudioAnalyticsRequest{StudioId: studio.Id}
			resp, err = GetStudioAnalytics(ctx, req)
		})

		if err != nil {
			t.Errorf("Expected no error, got error: %s", err.Error())
		}
		if resp.Analytics == nil {
			t.Fatal("Expected analytics to be returned")
		}
		if resp.Analytics.StudioId != studio.Id {
			t.Errorf("Expected StudioId %d, got %d", studio.Id, resp.Analytics.StudioId)
		}
		if resp.Analytics.TotalRooms != 3 {
			t.Errorf("Expected TotalRooms 3, got %d", resp.Analytics.TotalRooms)
		}
		if resp.Analytics.ActiveRooms != 2 {
			t.Errorf("Expected ActiveRooms 2, got %d", resp.Analytics.ActiveRooms)
		}
		if resp.StudioName != "Test Studio" {
			t.Errorf("Expected StudioName 'Test Studio', got %s", resp.StudioName)
		}
	})

	// Test 5: Success - viewer role
	t.Run("SuccessViewer", func(t *testing.T) {
		token, err := createTestToken(viewerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GetStudioAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetStudioAnalyticsRequest{StudioId: studio.Id}
			resp, err = GetStudioAnalytics(ctx, req)
		})

		if err != nil {
			t.Errorf("Expected no error for viewer role, got error: %s", err.Error())
		}
		if resp.Analytics == nil {
			t.Fatal("Expected analytics to be returned")
		}
	})

	// Test 6: New studio with no analytics (auto-aggregation)
	t.Run("NoAnalyticsAutoAggregate", func(t *testing.T) {
		token, err := createTestToken(ownerUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create a new studio
		var newStudio Studio
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			newStudio = Studio{
				Id:          vbolt.NextIntId(tx, StudiosBkt),
				Name:        "New Studio",
				Description: "A new studio",
				MaxRooms:    5,
				OwnerId:     ownerUser.Id,
				Creation:    time.Now(),
			}
			vbolt.Write(tx, StudiosBkt, newStudio.Id, &newStudio)

			// Create studio membership for owner
			newOwnerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
			newOwnerMembership := StudioMembership{
				UserId:   ownerUser.Id,
				StudioId: newStudio.Id,
				Role:     StudioRoleOwner,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, newOwnerMembershipId, &newOwnerMembership)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, newOwnerMembershipId, ownerUser.Id)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, newOwnerMembershipId, newStudio.Id)

			vbolt.TxCommit(tx)
		})

		var resp GetStudioAnalyticsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GetStudioAnalyticsRequest{StudioId: newStudio.Id}
			resp, err = GetStudioAnalytics(ctx, req)
		})

		if err != nil {
			t.Errorf("Expected no error, got error: %s", err.Error())
		}
		if resp.Analytics == nil {
			t.Fatal("Expected analytics to be returned")
		}
		if resp.Analytics.StudioId != newStudio.Id {
			t.Errorf("Expected StudioId %d, got %d", newStudio.Id, resp.Analytics.StudioId)
		}
		if resp.Analytics.TotalRooms != 0 {
			t.Errorf("Expected TotalRooms 0 for new studio, got %d", resp.Analytics.TotalRooms)
		}
	})
}

// TestResetMonthlyAnalytics tests the monthly reset functionality
func TestResetMonthlyAnalytics(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio with rooms
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create 3 rooms with analytics
		for i := 1; i <= 3; i++ {
			room := Room{Id: i, StudioId: 1, Name: "Room", Creation: time.Now()}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

			analytics := RoomAnalytics{
				RoomId:              i,
				TotalViewsAllTime:   1000 + i*100,
				TotalViewsThisMonth: 100 + i*10,
				CurrentViewers:      i,
				PeakViewers:         i * 5,
				TotalStreamMinutes:  i * 60,
			}
			vbolt.Write(tx, RoomAnalyticsBkt, i, &analytics)
		}

		// Create studio analytics
		studioAnalytics := StudioAnalytics{
			StudioId:            1,
			TotalViewsAllTime:   3600,
			TotalViewsThisMonth: 360,
			CurrentViewers:      6,
			TotalRooms:          3,
			TotalStreamMinutes:  360,
		}
		vbolt.Write(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		vbolt.TxCommit(tx)
	})

	// Run the monthly reset
	resetCount := ResetMonthlyAnalytics(db)

	// Verify reset count (3 rooms + 1 studio)
	if resetCount != 4 {
		t.Errorf("Expected resetCount 4, got %d", resetCount)
	}

	// Verify room analytics were reset
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		for i := 1; i <= 3; i++ {
			var analytics RoomAnalytics
			vbolt.Read(tx, RoomAnalyticsBkt, i, &analytics)

			if analytics.TotalViewsThisMonth != 0 {
				t.Errorf("Room %d: Expected TotalViewsThisMonth 0, got %d", i, analytics.TotalViewsThisMonth)
			}

			// Verify other fields were not reset
			expectedAllTime := 1000 + i*100
			if analytics.TotalViewsAllTime != expectedAllTime {
				t.Errorf("Room %d: TotalViewsAllTime should remain %d, got %d", i, expectedAllTime, analytics.TotalViewsAllTime)
			}
			if analytics.CurrentViewers != i {
				t.Errorf("Room %d: CurrentViewers should remain %d, got %d", i, i, analytics.CurrentViewers)
			}
		}

		// Verify studio analytics were reset
		var studioAnalytics StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		if studioAnalytics.TotalViewsThisMonth != 0 {
			t.Errorf("Studio: Expected TotalViewsThisMonth 0, got %d", studioAnalytics.TotalViewsThisMonth)
		}

		// Verify other fields were not reset
		if studioAnalytics.TotalViewsAllTime != 3600 {
			t.Errorf("Studio: TotalViewsAllTime should remain 3600, got %d", studioAnalytics.TotalViewsAllTime)
		}
		if studioAnalytics.CurrentViewers != 6 {
			t.Errorf("Studio: CurrentViewers should remain 6, got %d", studioAnalytics.CurrentViewers)
		}
	})
}

// TestResetMonthlyAnalyticsEmptyDB tests reset with no data
func TestResetMonthlyAnalyticsEmptyDB(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	// Run reset on empty database
	resetCount := ResetMonthlyAnalytics(db)

	if resetCount != 0 {
		t.Errorf("Expected resetCount 0 for empty DB, got %d", resetCount)
	}
}

// TestResetMonthlyAnalyticsAlreadyZero tests reset when counters are already zero
func TestResetMonthlyAnalyticsAlreadyZero(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create room with TotalViewsThisMonth already at 0
		room := Room{Id: 1, StudioId: 1, Name: "Room", Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		analytics := RoomAnalytics{
			RoomId:              1,
			TotalViewsAllTime:   500,
			TotalViewsThisMonth: 0, // Already zero
			CurrentViewers:      3,
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 1, &analytics)

		// Studio analytics also at zero
		studioAnalytics := StudioAnalytics{
			StudioId:            1,
			TotalViewsAllTime:   500,
			TotalViewsThisMonth: 0, // Already zero
			CurrentViewers:      3,
		}
		vbolt.Write(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		vbolt.TxCommit(tx)
	})

	// Run reset - should skip already-zero entries
	resetCount := ResetMonthlyAnalytics(db)

	if resetCount != 0 {
		t.Errorf("Expected resetCount 0 when counters already zero, got %d", resetCount)
	}
}

// TestResetMonthlyAnalyticsPartialData tests reset with mixed data
func TestResetMonthlyAnalyticsPartialData(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Room 1: Has monthly views
		room1 := Room{Id: 1, StudioId: 1, Name: "Room 1", Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room1.Id, &room1)
		analytics1 := RoomAnalytics{
			RoomId:              1,
			TotalViewsAllTime:   500,
			TotalViewsThisMonth: 50,
			CurrentViewers:      2,
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 1, &analytics1)

		// Room 2: Already zero
		room2 := Room{Id: 2, StudioId: 1, Name: "Room 2", Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
		analytics2 := RoomAnalytics{
			RoomId:              2,
			TotalViewsAllTime:   300,
			TotalViewsThisMonth: 0,
			CurrentViewers:      1,
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 2, &analytics2)

		// Studio: Has monthly views
		studioAnalytics := StudioAnalytics{
			StudioId:            1,
			TotalViewsAllTime:   800,
			TotalViewsThisMonth: 50,
			CurrentViewers:      3,
		}
		vbolt.Write(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		vbolt.TxCommit(tx)
	})

	// Run reset - should only reset room1 and studio (2 resets)
	resetCount := ResetMonthlyAnalytics(db)

	if resetCount != 2 {
		t.Errorf("Expected resetCount 2 (room1 + studio), got %d", resetCount)
	}

	// Verify
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		var analytics1 RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, 1, &analytics1)
		if analytics1.TotalViewsThisMonth != 0 {
			t.Errorf("Room 1: Expected TotalViewsThisMonth 0, got %d", analytics1.TotalViewsThisMonth)
		}

		var analytics2 RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, 2, &analytics2)
		if analytics2.TotalViewsThisMonth != 0 {
			t.Errorf("Room 2: Expected TotalViewsThisMonth 0, got %d", analytics2.TotalViewsThisMonth)
		}

		var studioAnalytics StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, 1, &studioAnalytics)
		if studioAnalytics.TotalViewsThisMonth != 0 {
			t.Errorf("Studio: Expected TotalViewsThisMonth 0, got %d", studioAnalytics.TotalViewsThisMonth)
		}
	})
}

// TestResetMonthlyAnalyticsWithUniqueViewers tests that UniqueViewersThisMonth is also reset
func TestResetMonthlyAnalyticsWithUniqueViewers(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio and room with unique viewer counts
		studio := Studio{Id: 1, Name: "Test Studio", OwnerId: 100, Creation: time.Now()}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		room := Room{Id: 1, StudioId: 1, Name: "Room", Creation: time.Now()}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		analytics := RoomAnalytics{
			RoomId:                 1,
			TotalViewsAllTime:      1000,
			TotalViewsThisMonth:    100,
			UniqueViewersAllTime:   50,
			UniqueViewersThisMonth: 20,
			CurrentViewers:         5,
		}
		vbolt.Write(tx, RoomAnalyticsBkt, 1, &analytics)

		studioAnalytics := StudioAnalytics{
			StudioId:               1,
			TotalViewsAllTime:      1000,
			TotalViewsThisMonth:    100,
			UniqueViewersAllTime:   50,
			UniqueViewersThisMonth: 20,
			CurrentViewers:         5,
		}
		vbolt.Write(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		vbolt.TxCommit(tx)
	})

	// Run reset
	resetCount := ResetMonthlyAnalytics(db)

	if resetCount != 2 {
		t.Errorf("Expected resetCount 2 (room + studio), got %d", resetCount)
	}

	// Verify both TotalViewsThisMonth and UniqueViewersThisMonth were reset
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		var analytics RoomAnalytics
		vbolt.Read(tx, RoomAnalyticsBkt, 1, &analytics)

		if analytics.TotalViewsThisMonth != 0 {
			t.Errorf("Expected TotalViewsThisMonth 0, got %d", analytics.TotalViewsThisMonth)
		}
		if analytics.UniqueViewersThisMonth != 0 {
			t.Errorf("Expected UniqueViewersThisMonth 0, got %d", analytics.UniqueViewersThisMonth)
		}
		// All-time counters should remain
		if analytics.TotalViewsAllTime != 1000 {
			t.Errorf("Expected TotalViewsAllTime 1000, got %d", analytics.TotalViewsAllTime)
		}
		if analytics.UniqueViewersAllTime != 50 {
			t.Errorf("Expected UniqueViewersAllTime 50, got %d", analytics.UniqueViewersAllTime)
		}

		var studioAnalytics StudioAnalytics
		vbolt.Read(tx, StudioAnalyticsBkt, 1, &studioAnalytics)

		if studioAnalytics.TotalViewsThisMonth != 0 {
			t.Errorf("Studio: Expected TotalViewsThisMonth 0, got %d", studioAnalytics.TotalViewsThisMonth)
		}
		if studioAnalytics.UniqueViewersThisMonth != 0 {
			t.Errorf("Studio: Expected UniqueViewersThisMonth 0, got %d", studioAnalytics.UniqueViewersThisMonth)
		}
		if studioAnalytics.TotalViewsAllTime != 1000 {
			t.Errorf("Studio: Expected TotalViewsAllTime 1000, got %d", studioAnalytics.TotalViewsAllTime)
		}
		if studioAnalytics.UniqueViewersAllTime != 50 {
			t.Errorf("Studio: Expected UniqueViewersAllTime 50, got %d", studioAnalytics.UniqueViewersAllTime)
		}
	})
}

// TestViewerSessionCleanup tests the session cleanup functionality
func TestViewerSessionCleanup(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create 3 sessions with different ages
		now := time.Now()

		// Session 1: Recent (5 minutes ago) - should NOT be cleaned
		session1 := ViewerSession{
			SessionKey:  "user:1:10",
			ViewerId:    "user:1",
			RoomId:      10,
			FirstSeenAt: now.Add(-5 * time.Minute),
			LastSeenAt:  now.Add(-5 * time.Minute),
		}
		vbolt.Write(tx, ViewerSessionsBkt, session1.SessionKey, &session1)
		vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, session1.SessionKey, session1.RoomId)

		// Session 2: Old (15 minutes ago) - SHOULD be cleaned
		session2 := ViewerSession{
			SessionKey:  "user:2:10",
			ViewerId:    "user:2",
			RoomId:      10,
			FirstSeenAt: now.Add(-20 * time.Minute),
			LastSeenAt:  now.Add(-15 * time.Minute),
		}
		vbolt.Write(tx, ViewerSessionsBkt, session2.SessionKey, &session2)
		vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, session2.SessionKey, session2.RoomId)

		// Session 3: Very old (30 minutes ago) - SHOULD be cleaned
		session3 := ViewerSession{
			SessionKey:  "code:abc:20",
			ViewerId:    "code:abc",
			RoomId:      20,
			FirstSeenAt: now.Add(-30 * time.Minute),
			LastSeenAt:  now.Add(-30 * time.Minute),
		}
		vbolt.Write(tx, ViewerSessionsBkt, session3.SessionKey, &session3)
		vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, session3.SessionKey, session3.RoomId)

		vbolt.TxCommit(tx)
	})

	// Manually run the cleanup logic (same as in StartViewerSessionCleanup)
	cleanupCount := 0
	cutoffTime := time.Now().Add(-SESSION_CLEANUP_AGE)

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		vbolt.IterateAll(tx, ViewerSessionsBkt, func(sessionKey string, session ViewerSession) bool {
			if session.LastSeenAt.Before(cutoffTime) {
				vbolt.Delete(tx, ViewerSessionsBkt, sessionKey)
				vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, sessionKey, -1)
				cleanupCount++
			}
			return true
		})

		vbolt.TxCommit(tx)
	})

	// Should have cleaned 2 old sessions
	if cleanupCount != 2 {
		t.Errorf("Expected cleanupCount 2 (session2 and session3), got %d", cleanupCount)
	}

	// Verify session1 still exists
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		var session1 ViewerSession
		vbolt.Read(tx, ViewerSessionsBkt, "user:1:10", &session1)
		if session1.SessionKey != "user:1:10" {
			t.Errorf("Session1 should still exist, got empty SessionKey")
		}

		// Verify session2 was deleted
		var session2 ViewerSession
		vbolt.Read(tx, ViewerSessionsBkt, "user:2:10", &session2)
		if session2.SessionKey != "" {
			t.Errorf("Session2 should be deleted, but SessionKey is '%s'", session2.SessionKey)
		}

		// Verify session3 was deleted
		var session3 ViewerSession
		vbolt.Read(tx, ViewerSessionsBkt, "code:abc:20", &session3)
		if session3.SessionKey != "" {
			t.Errorf("Session3 should be deleted, but SessionKey is '%s'", session3.SessionKey)
		}
	})
}

// TestViewerSessionCleanupEmptyDB tests cleanup with no sessions
func TestViewerSessionCleanupEmptyDB(t *testing.T) {
	db := setupTestAnalyticsDB(t)
	defer db.Close()

	// Run cleanup on empty database
	cleanupCount := 0
	cutoffTime := time.Now().Add(-SESSION_CLEANUP_AGE)

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		vbolt.IterateAll(tx, ViewerSessionsBkt, func(sessionKey string, session ViewerSession) bool {
			if session.LastSeenAt.Before(cutoffTime) {
				vbolt.Delete(tx, ViewerSessionsBkt, sessionKey)
				vbolt.SetTargetSingleTerm(tx, SessionsByRoomIndex, sessionKey, -1)
				cleanupCount++
			}
			return true
		})

		vbolt.TxCommit(tx)
	})

	if cleanupCount != 0 {
		t.Errorf("Expected cleanupCount 0 for empty DB, got %d", cleanupCount)
	}
}
