package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

// Test helper to create a test user
func createTestUser(t *testing.T, tx *vbolt.Tx, email string, role UserRole) User {
	user := User{
		Id:        vbolt.NextIntId(tx, UsersBkt),
		Name:      "Test User",
		Email:     email,
		Role:      role,
		Creation:  time.Now(),
		LastLogin: time.Now(),
	}
	vbolt.Write(tx, UsersBkt, user.Id, &user)
	vbolt.Write(tx, EmailBkt, user.Email, &user.Id)
	return user
}

// Test Studio packing/unpacking
func TestPackStudio(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test studio
		original := Studio{
			Id:          1,
			Name:        "Test Studio",
			Description: "A test studio for testing",
			MaxRooms:    5,
			OwnerId:     100,
			Creation:    time.Now().Truncate(time.Second), // Truncate for comparison
		}

		// Write and read back
		vbolt.Write(tx, StudiosBkt, original.Id, &original)
		var retrieved Studio
		vbolt.Read(tx, StudiosBkt, original.Id, &retrieved)

		// Verify all fields match
		if retrieved.Id != original.Id {
			t.Errorf("Id mismatch: got %d, want %d", retrieved.Id, original.Id)
		}
		if retrieved.Name != original.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, original.Name)
		}
		if retrieved.Description != original.Description {
			t.Errorf("Description mismatch: got %s, want %s", retrieved.Description, original.Description)
		}
		if retrieved.MaxRooms != original.MaxRooms {
			t.Errorf("MaxRooms mismatch: got %d, want %d", retrieved.MaxRooms, original.MaxRooms)
		}
		if retrieved.OwnerId != original.OwnerId {
			t.Errorf("OwnerId mismatch: got %d, want %d", retrieved.OwnerId, original.OwnerId)
		}
		if !retrieved.Creation.Equal(original.Creation) {
			t.Errorf("Creation mismatch: got %v, want %v", retrieved.Creation, original.Creation)
		}
	})
}

// Test Room packing/unpacking
func TestPackRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test room
		original := Room{
			Id:         1,
			StudioId:   10,
			RoomNumber: 3,
			Name:       "Main Stage",
			StreamKey:  "test-stream-key-12345",
			IsActive:   true,
			Creation:   time.Now().Truncate(time.Second),
		}

		// Write and read back
		vbolt.Write(tx, RoomsBkt, original.Id, &original)
		var retrieved Room
		vbolt.Read(tx, RoomsBkt, original.Id, &retrieved)

		// Verify all fields match
		if retrieved.Id != original.Id {
			t.Errorf("Id mismatch: got %d, want %d", retrieved.Id, original.Id)
		}
		if retrieved.StudioId != original.StudioId {
			t.Errorf("StudioId mismatch: got %d, want %d", retrieved.StudioId, original.StudioId)
		}
		if retrieved.RoomNumber != original.RoomNumber {
			t.Errorf("RoomNumber mismatch: got %d, want %d", retrieved.RoomNumber, original.RoomNumber)
		}
		if retrieved.Name != original.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, original.Name)
		}
		if retrieved.StreamKey != original.StreamKey {
			t.Errorf("StreamKey mismatch: got %s, want %s", retrieved.StreamKey, original.StreamKey)
		}
		if retrieved.IsActive != original.IsActive {
			t.Errorf("IsActive mismatch: got %v, want %v", retrieved.IsActive, original.IsActive)
		}
		if !retrieved.Creation.Equal(original.Creation) {
			t.Errorf("Creation mismatch: got %v, want %v", retrieved.Creation, original.Creation)
		}
	})
}

// Test StudioMembership packing/unpacking
func TestPackStudioMembership(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Test all role types
		roles := []StudioRole{StudioRoleViewer, StudioRoleMember, StudioRoleAdmin, StudioRoleOwner}

		for i, role := range roles {
			original := StudioMembership{
				UserId:   100 + i,
				StudioId: 50,
				Role:     role,
				JoinedAt: time.Now().Truncate(time.Second),
			}

			membershipId := i + 1
			vbolt.Write(tx, MembershipBkt, membershipId, &original)
			var retrieved StudioMembership
			vbolt.Read(tx, MembershipBkt, membershipId, &retrieved)

			if retrieved.UserId != original.UserId {
				t.Errorf("UserId mismatch: got %d, want %d", retrieved.UserId, original.UserId)
			}
			if retrieved.StudioId != original.StudioId {
				t.Errorf("StudioId mismatch: got %d, want %d", retrieved.StudioId, original.StudioId)
			}
			if retrieved.Role != original.Role {
				t.Errorf("Role mismatch: got %d, want %d", retrieved.Role, original.Role)
			}
			if !retrieved.JoinedAt.Equal(original.JoinedAt) {
				t.Errorf("JoinedAt mismatch: got %v, want %v", retrieved.JoinedAt, original.JoinedAt)
			}
		}
	})
}

// Test Stream packing/unpacking with zero-value EndTime
func TestPackStream(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Test active stream (EndTime is zero value)
		activeStream := Stream{
			Id:          1,
			StudioId:    10,
			RoomId:      5,
			Title:       "Live Stream",
			Description: "Currently streaming",
			StartTime:   time.Now().Truncate(time.Second),
			EndTime:     time.Time{}, // Active stream has zero-value end time
			CreatedBy:   100,
		}

		vbolt.Write(tx, StreamsBkt, activeStream.Id, &activeStream)
		var retrievedActive Stream
		vbolt.Read(tx, StreamsBkt, activeStream.Id, &retrievedActive)

		if !retrievedActive.StartTime.Equal(activeStream.StartTime) {
			t.Errorf("StartTime mismatch: got %v, want %v", retrievedActive.StartTime, activeStream.StartTime)
		}
		if !retrievedActive.EndTime.IsZero() {
			t.Errorf("EndTime should be zero for active stream, got %v", retrievedActive.EndTime)
		}

		// Test completed stream (EndTime is set)
		endTime := time.Now().Add(2 * time.Hour).Truncate(time.Second)
		completedStream := Stream{
			Id:          2,
			StudioId:    10,
			RoomId:      5,
			Title:       "Past Stream",
			Description: "Already ended",
			StartTime:   time.Now().Truncate(time.Second),
			EndTime:     endTime,
			CreatedBy:   100,
		}

		vbolt.Write(tx, StreamsBkt, completedStream.Id, &completedStream)
		var retrievedCompleted Stream
		vbolt.Read(tx, StreamsBkt, completedStream.Id, &retrievedCompleted)

		if retrievedCompleted.EndTime.IsZero() {
			t.Error("EndTime should not be zero for completed stream")
		}
		if !retrievedCompleted.EndTime.Equal(endTime) {
			t.Errorf("EndTime mismatch: got %v, want %v", retrievedCompleted.EndTime, endTime)
		}
	})
}

// Test GenerateStreamKey
func TestGenerateStreamKey(t *testing.T) {
	key1, err := GenerateStreamKey()
	if err != nil {
		t.Fatalf("Failed to generate stream key: %v", err)
	}
	if len(key1) == 0 {
		t.Error("Generated stream key is empty")
	}

	key2, err := GenerateStreamKey()
	if err != nil {
		t.Fatalf("Failed to generate second stream key: %v", err)
	}

	// Keys should be unique
	if key1 == key2 {
		t.Error("Generated stream keys are not unique")
	}
}

// Test GetStudio
func TestGetStudio(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		studio := Studio{
			Id:          1,
			Name:        "Test Studio",
			Description: "Description",
			MaxRooms:    3,
			OwnerId:     100,
			Creation:    time.Now(),
		}

		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Test retrieval
		retrieved := GetStudio(tx, studio.Id)
		if retrieved.Id != studio.Id {
			t.Errorf("GetStudio failed: got ID %d, want %d", retrieved.Id, studio.Id)
		}
		if retrieved.Name != studio.Name {
			t.Errorf("GetStudio failed: got name %s, want %s", retrieved.Name, studio.Name)
		}

		// Test non-existent studio
		notFound := GetStudio(tx, 999)
		if notFound.Id != 0 {
			t.Errorf("Expected zero-value studio for non-existent ID, got %v", notFound)
		}
	})
}

// Test GetRoom and GetRoomByStreamKey
func TestGetRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		room := Room{
			Id:         1,
			StudioId:   10,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  "secret-key-123",
			IsActive:   false,
			Creation:   time.Now(),
		}

		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.Write(tx, RoomStreamKeyBkt, room.StreamKey, &room.Id)

		// Test GetRoom
		retrieved := GetRoom(tx, room.Id)
		if retrieved.Id != room.Id || retrieved.StreamKey != room.StreamKey {
			t.Errorf("GetRoom failed: got %v, want %v", retrieved, room)
		}

		// Test GetRoomByStreamKey
		retrievedByKey := GetRoomByStreamKey(tx, room.StreamKey)
		if retrievedByKey.Id != room.Id {
			t.Errorf("GetRoomByStreamKey failed: got ID %d, want %d", retrievedByKey.Id, room.Id)
		}

		// Test with invalid key
		notFound := GetRoomByStreamKey(tx, "invalid-key")
		if notFound.Id != 0 {
			t.Errorf("Expected zero-value room for invalid key, got %v", notFound)
		}
	})
}

// Test RoomsByStudioIdx index
func TestRoomsByStudioIndex(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		studioId := 10

		// Create multiple rooms for a studio
		for i := 1; i <= 3; i++ {
			room := Room{
				Id:         i,
				StudioId:   studioId,
				RoomNumber: i,
				Name:       "Room " + string(rune('0'+i)),
				StreamKey:  "key-" + string(rune('0'+i)),
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, room.StudioId)
		}

		// Retrieve all rooms for the studio
		rooms := ListStudioRooms(tx, studioId)
		if len(rooms) != 3 {
			t.Errorf("Expected 3 rooms, got %d", len(rooms))
		}

		// Verify room IDs
		roomIds := make(map[int]bool)
		for _, room := range rooms {
			roomIds[room.Id] = true
		}
		for i := 1; i <= 3; i++ {
			if !roomIds[i] {
				t.Errorf("Room %d not found in results", i)
			}
		}
	})
}

// Test MembershipByUserIdx and MembershipByStudioIdx
func TestMembershipIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		userId := 100
		studioIds := []int{10, 20, 30}

		// Create memberships for one user across multiple studios
		for i, studioId := range studioIds {
			membershipId := i + 1
			membership := StudioMembership{
				UserId:   userId,
				StudioId: studioId,
				Role:     StudioRoleMember,
				JoinedAt: time.Now(),
			}
			vbolt.Write(tx, MembershipBkt, membershipId, &membership)
			vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)
			vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studioId)
		}

		// Test ListUserStudios
		studios := ListUserStudios(tx, userId)
		if len(studios) != 0 {
			// Studios don't exist yet, but memberships do
			t.Logf("Note: ListUserStudios returns empty because studios aren't created")
		}

		// Test ListStudioMembers
		members := ListStudioMembers(tx, studioIds[0])
		if len(members) != 1 {
			t.Errorf("Expected 1 member for studio %d, got %d", studioIds[0], len(members))
		}
		if len(members) > 0 && members[0].UserId != userId {
			t.Errorf("Expected userId %d, got %d", userId, members[0].UserId)
		}
	})
}

// Test GetUserStudioRole
func TestGetUserStudioRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		userId := 100
		studioId := 10
		membershipId := 1

		// Create membership with Admin role
		membership := StudioMembership{
			UserId:   userId,
			StudioId: studioId,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, userId)

		// Test getting role
		role := GetUserStudioRole(tx, userId, studioId)
		if role != StudioRoleAdmin {
			t.Errorf("Expected StudioRoleAdmin, got %d", role)
		}

		// Test non-member
		nonMemberRole := GetUserStudioRole(tx, 999, studioId)
		if nonMemberRole != -1 {
			t.Errorf("Expected -1 for non-member, got %d", nonMemberRole)
		}
	})
}

// Test HasStudioPermission
func TestHasStudioPermission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a regular user with Member role
		memberUser := createTestUser(t, tx, "member@test.com", RoleUser)
		studioId := 10
		membershipId := 1

		membership := StudioMembership{
			UserId:   memberUser.Id,
			StudioId: studioId,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, memberUser.Id)

		// Member should have Member permission
		if !HasStudioPermission(tx, memberUser.Id, studioId, StudioRoleMember) {
			t.Error("Member should have Member permission")
		}

		// Member should not have Admin permission
		if HasStudioPermission(tx, memberUser.Id, studioId, StudioRoleAdmin) {
			t.Error("Member should not have Admin permission")
		}

		// Create a site admin user
		siteAdmin := createTestUser(t, tx, "admin@test.com", RoleSiteAdmin)

		// Site admin should have permission even without membership
		if !HasStudioPermission(tx, siteAdmin.Id, studioId, StudioRoleOwner) {
			t.Error("Site admin should have full permissions in all studios")
		}

		// Non-member should not have permission
		nonMember := createTestUser(t, tx, "nonmember@test.com", RoleUser)
		if HasStudioPermission(tx, nonMember.Id, studioId, StudioRoleViewer) {
			t.Error("Non-member should not have any permissions")
		}
	})
}

// Test GetStudioRoleName
func TestGetStudioRoleName(t *testing.T) {
	tests := []struct {
		role     StudioRole
		expected string
	}{
		{StudioRoleViewer, "Viewer"},
		{StudioRoleMember, "Member"},
		{StudioRoleAdmin, "Admin"},
		{StudioRoleOwner, "Owner"},
		{StudioRole(99), "Unknown"},
	}

	for _, tt := range tests {
		result := GetStudioRoleName(tt.role)
		if result != tt.expected {
			t.Errorf("GetStudioRoleName(%d) = %s, want %s", tt.role, result, tt.expected)
		}
	}
}

// Test StreamsByStudioIdx and StreamsByRoomIdx
func TestStreamIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		studioId := 10
		roomId := 5

		// Create multiple streams
		for i := 1; i <= 3; i++ {
			stream := Stream{
				Id:          i,
				StudioId:    studioId,
				RoomId:      roomId,
				Title:       "Stream " + string(rune('0'+i)),
				Description: "Test stream",
				StartTime:   time.Now(),
				EndTime:     time.Time{}, // Active streams (zero value)
				CreatedBy:   100,
			}
			vbolt.Write(tx, StreamsBkt, stream.Id, &stream)
			vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, stream.Id, stream.StudioId)
			vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, stream.Id, stream.RoomId)
		}

		// Test StreamsByStudioIdx
		var streamIds []int
		vbolt.ReadTermTargets(tx, StreamsByStudioIdx, studioId, &streamIds, vbolt.Window{})
		if len(streamIds) != 3 {
			t.Errorf("Expected 3 streams for studio, got %d", len(streamIds))
		}

		// Test StreamsByRoomIdx
		var roomStreamIds []int
		vbolt.ReadTermTargets(tx, StreamsByRoomIdx, roomId, &roomStreamIds, vbolt.Window{})
		if len(roomStreamIds) != 3 {
			t.Errorf("Expected 3 streams for room, got %d", len(roomStreamIds))
		}
	})
}
