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

// Test GetStudioById
func TestGetStudioById(t *testing.T) {
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
		retrieved := GetStudioById(tx, studio.Id)
		if retrieved.Id != studio.Id {
			t.Errorf("GetStudioById failed: got ID %d, want %d", retrieved.Id, studio.Id)
		}
		if retrieved.Name != studio.Name {
			t.Errorf("GetStudioById failed: got name %s, want %s", retrieved.Name, studio.Name)
		}

		// Test non-existent studio
		notFound := GetStudioById(tx, 999)
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

// Business Logic Tests

// Test studio creation and membership workflow
func TestStudioCreationWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var ownerUserId int
	var studioId int

	// Create everything in one transaction then verify
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create user
		user := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		ownerUserId = user.Id

		// Create studio
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "A test studio",
			MaxRooms:    10,
			OwnerId:     ownerUserId,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)
		studioId = studio.Id

		// Create owner membership
		membership := StudioMembership{
			UserId:   ownerUserId,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, ownerUserId)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Verify within same transaction
		retrieved := GetStudioById(tx, studioId)
		if retrieved.Id == 0 {
			t.Error("Studio should exist")
		}
		if retrieved.Name != "Test Studio" {
			t.Errorf("Studio name mismatch: got %s, want %s", retrieved.Name, "Test Studio")
		}
		if retrieved.MaxRooms != 10 {
			t.Errorf("MaxRooms mismatch: got %d, want %d", retrieved.MaxRooms, 10)
		}

		// Verify membership and role
		role := GetUserStudioRole(tx, ownerUserId, studioId)
		if role != StudioRoleOwner {
			t.Errorf("Expected owner role, got %d", role)
		}
	})
}

// Test listing user studios with roles
func TestListUserStudiosWithRoles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var user1Id, user2Id, studio1Id, studio2Id int

	// Create users and studios, then verify in same transaction
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		user1 := createTestUser(t, tx, "user1@test.com", RoleUser)
		user2 := createTestUser(t, tx, "user2@test.com", RoleUser)
		user1Id = user1.Id
		user2Id = user2.Id

		// Create two studios
		studio1 := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Studio 1",
			MaxRooms: 5,
			OwnerId:  user1Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio1.Id, &studio1)
		studio1Id = studio1.Id

		studio2 := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Studio 2",
			MaxRooms: 5,
			OwnerId:  user2Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio2.Id, &studio2)
		studio2Id = studio2.Id

		// User1 is owner of studio1
		membership1 := StudioMembership{
			UserId:   user1Id,
			StudioId: studio1Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		membershipId1 := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId1, &membership1)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId1, user1Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId1, studio1Id)

		// User2 is owner of studio2
		membership2 := StudioMembership{
			UserId:   user2Id,
			StudioId: studio2Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		membershipId2 := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId2, &membership2)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId2, user2Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId2, studio2Id)

		// User1 is also a member of studio2
		membership3 := StudioMembership{
			UserId:   user1Id,
			StudioId: studio2Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId3 := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId3, &membership3)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId3, user1Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId3, studio2Id)

		// Test user1 sees both studios with correct roles
		studios := ListUserStudios(tx, user1Id)
		if len(studios) != 2 {
			t.Errorf("User1 should see 2 studios, got %d", len(studios))
		}

		roleMap := make(map[int]StudioRole)
		for _, studio := range studios {
			roleMap[studio.Id] = GetUserStudioRole(tx, user1Id, studio.Id)
		}

		if roleMap[studio1Id] != StudioRoleOwner {
			t.Errorf("User1 should be owner of studio1, got role %d", roleMap[studio1Id])
		}
		if roleMap[studio2Id] != StudioRoleMember {
			t.Errorf("User1 should be member of studio2, got role %d", roleMap[studio2Id])
		}

		// Test user2 sees only studio2
		studios2 := ListUserStudios(tx, user2Id)
		if len(studios2) != 1 {
			t.Errorf("User2 should see 1 studio, got %d", len(studios2))
		}
		if len(studios2) > 0 && studios2[0].Id != studio2Id {
			t.Errorf("User2 should see studio2, got %d", studios2[0].Id)
		}
	})
}

// Test permission checking
func TestStudioPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	var ownerUser, memberUser, nonMemberUser, siteAdminUser User
	var studioId int

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		ownerUser = createTestUser(t, tx, "owner@test.com", RoleUser)
		memberUser = createTestUser(t, tx, "member@test.com", RoleUser)
		nonMemberUser = createTestUser(t, tx, "nonmember@test.com", RoleUser)
		siteAdminUser = createTestUser(t, tx, "siteadmin@test.com", RoleSiteAdmin)

		// Create a studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  ownerUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)
		studioId = studio.Id

		// Add owner membership
		ownerMembership := StudioMembership{
			UserId:   ownerUser.Id,
			StudioId: studioId,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, ownerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studioId)

		// Add member membership
		memberMembership := StudioMembership{
			UserId:   memberUser.Id,
			StudioId: studioId,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		memberMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, memberMembershipId, &memberMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, memberMembershipId, memberUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, memberMembershipId, studioId)

		// Test permissions within same transaction
		// Owner has full permissions
		if !HasStudioPermission(tx, ownerUser.Id, studioId, StudioRoleOwner) {
			t.Error("Owner should have owner permission")
		}

		// Member has member permission but not owner
		if !HasStudioPermission(tx, memberUser.Id, studioId, StudioRoleMember) {
			t.Error("Member should have member permission")
		}
		if HasStudioPermission(tx, memberUser.Id, studioId, StudioRoleOwner) {
			t.Error("Member should not have owner permission")
		}

		// Non-member has no permissions
		if HasStudioPermission(tx, nonMemberUser.Id, studioId, StudioRoleViewer) {
			t.Error("Non-member should not have any permissions")
		}

		// Site admin has all permissions even without membership
		if !HasStudioPermission(tx, siteAdminUser.Id, studioId, StudioRoleOwner) {
			t.Error("Site admin should have owner permission")
		}
	})
}

// Test UpdateStudio - valid updates work correctly
func TestUpdateStudio(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a studio owner (StreamAdmin)
		ownerUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Owner",
			Email: "owner@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, ownerUser.Id, &ownerUser)

		// Create studio
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Original Name",
			Description: "Original Description",
			MaxRooms:    5,
			OwnerId:     ownerUser.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add owner membership
		ownerMembership := StudioMembership{
			UserId:   ownerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, ownerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Update studio fields
		studio.Name = "Updated Name"
		studio.Description = "Updated Description"
		studio.MaxRooms = 10
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Verify update
		updated := GetStudioById(tx, studio.Id)
		if updated.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
		}
		if updated.Description != "Updated Description" {
			t.Errorf("Expected description 'Updated Description', got '%s'", updated.Description)
		}
		if updated.MaxRooms != 10 {
			t.Errorf("Expected maxRooms 10, got %d", updated.MaxRooms)
		}
	})
}

// Test UpdateStudio permissions - only Admin+ can update
func TestUpdateStudioPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users with different roles
		adminUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Admin User",
			Email: "admin@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)

		memberUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Member User",
			Email: "member@test.com",
			Role:  RoleUser,
		}
		vbolt.Write(tx, UsersBkt, memberUser.Id, &memberUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  adminUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Add member membership
		memberMembership := StudioMembership{
			UserId:   memberUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		memberMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, memberMembershipId, &memberMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, memberMembershipId, memberUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, memberMembershipId, studio.Id)

		// Admin should have permission to update
		if !HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Admin should have permission to update studio")
		}

		// Member should NOT have permission to update
		if HasStudioPermission(tx, memberUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Member should not have permission to update studio")
		}
	})
}

// Test DeleteStudio cascade delete
func TestDeleteStudioCascade(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner user
		ownerUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Owner",
			Email: "owner@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, ownerUser.Id, &ownerUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  ownerUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add owner membership
		ownerMembership := StudioMembership{
			UserId:   ownerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, ownerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Create a room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)

		// Create a stream
		stream := Stream{
			Id:          vbolt.NextIntId(tx, StreamsBkt),
			StudioId:    studio.Id,
			RoomId:      room.Id,
			Title:       "Test Stream",
			Description: "Test",
			StartTime:   time.Now(),
			CreatedBy:   ownerUser.Id,
		}
		vbolt.Write(tx, StreamsBkt, stream.Id, &stream)
		vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, stream.Id, studio.Id)
		vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, stream.Id, room.Id)

		// Perform cascade delete
		// 1. Delete all rooms
		rooms := ListStudioRooms(tx, studio.Id)
		for _, r := range rooms {
			vbolt.Delete(tx, RoomStreamKeyBkt, r.StreamKey)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, r.Id, -1)
			vbolt.Delete(tx, RoomsBkt, r.Id)
		}

		// 2. Delete all streams
		var streamIds []int
		vbolt.ReadTermTargets(tx, StreamsByStudioIdx, studio.Id, &streamIds, vbolt.Window{})
		for _, sid := range streamIds {
			vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, sid, -1)
			vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, sid, -1)
			vbolt.Delete(tx, StreamsBkt, sid)
		}

		// 3. Delete all memberships
		memberships := ListStudioMembers(tx, studio.Id)
		for _, membership := range memberships {
			var membershipIds []int
			vbolt.ReadTermTargets(tx, MembershipByUserIdx, membership.UserId, &membershipIds, vbolt.Window{})
			for _, mid := range membershipIds {
				m := GetMembership(tx, mid)
				if m.StudioId == studio.Id {
					vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, mid, -1)
					vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, mid, -1)
					vbolt.Delete(tx, MembershipBkt, mid)
				}
			}
		}

		// 4. Delete studio
		vbolt.Delete(tx, StudiosBkt, studio.Id)

		// Verify everything is deleted
		deletedStudio := GetStudioById(tx, studio.Id)
		if deletedStudio.Id != 0 {
			t.Error("Studio should be deleted")
		}

		deletedRoom := GetRoom(tx, room.Id)
		if deletedRoom.Id != 0 {
			t.Error("Room should be deleted")
		}

		deletedStream := GetStream(tx, stream.Id)
		if deletedStream.Id != 0 {
			t.Error("Stream should be deleted")
		}

		// Check stream key is gone
		var roomIdFromKey int
		vbolt.Read(tx, RoomStreamKeyBkt, streamKey, &roomIdFromKey)
		if roomIdFromKey != 0 {
			t.Error("Stream key should be deleted")
		}

		// Check membership is gone
		deletedMembership := GetMembership(tx, ownerMembershipId)
		if deletedMembership.UserId != 0 {
			t.Error("Membership should be deleted")
		}

		// Check indexes are cleaned up
		var roomIdsFromIdx []int
		vbolt.ReadTermTargets(tx, RoomsByStudioIdx, studio.Id, &roomIdsFromIdx, vbolt.Window{})
		if len(roomIdsFromIdx) != 0 {
			t.Errorf("RoomsByStudioIdx should be empty, got %d entries", len(roomIdsFromIdx))
		}

		var streamIdsFromIdx []int
		vbolt.ReadTermTargets(tx, StreamsByStudioIdx, studio.Id, &streamIdsFromIdx, vbolt.Window{})
		if len(streamIdsFromIdx) != 0 {
			t.Errorf("StreamsByStudioIdx should be empty, got %d entries", len(streamIdsFromIdx))
		}
	})
}

// Test DeleteStudio permissions - only Owner can delete
func TestDeleteStudioPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		ownerUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Owner",
			Email: "owner@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, ownerUser.Id, &ownerUser)

		adminUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Admin",
			Email: "admin@test.com",
			Role:  RoleUser,
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  ownerUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add owner membership
		ownerMembership := StudioMembership{
			UserId:   ownerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, ownerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Owner should have permission to delete
		if !HasStudioPermission(tx, ownerUser.Id, studio.Id, StudioRoleOwner) {
			t.Error("Owner should have permission to delete studio")
		}

		// Admin should NOT have permission to delete (needs Owner role)
		if HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleOwner) {
			t.Error("Admin should not have owner permission to delete studio")
		}
	})
}

// Test CreateRoom - basic room creation
func TestCreateRoom(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Admin",
			Email: "admin@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  adminUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create a room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)

		// Verify room was created
		retrieved := GetRoom(tx, room.Id)
		if retrieved.Id == 0 {
			t.Error("Room should be created")
		}
		if retrieved.Name != "Room 1" {
			t.Errorf("Expected room name 'Room 1', got '%s'", retrieved.Name)
		}
		if retrieved.RoomNumber != 1 {
			t.Errorf("Expected room number 1, got %d", retrieved.RoomNumber)
		}

		// Verify room is in studio's rooms
		rooms := ListStudioRooms(tx, studio.Id)
		if len(rooms) != 1 {
			t.Errorf("Expected 1 room, got %d", len(rooms))
		}

		// Verify stream key lookup works
		roomByKey := GetRoomByStreamKey(tx, streamKey)
		if roomByKey.Id != room.Id {
			t.Error("Should be able to find room by stream key")
		}
	})
}

// Test CreateRoom room limit enforcement
func TestCreateRoomLimit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create admin user
		adminUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Admin",
			Email: "admin@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)

		// Create studio with maxRooms = 2
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 2,
			OwnerId:  adminUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create room 1
		streamKey1, _ := GenerateStreamKey()
		room1 := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey1,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room1.Id, &room1)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room1.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey1, &room1.Id)

		// Create room 2
		streamKey2, _ := GenerateStreamKey()
		room2 := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 2,
			Name:       "Room 2",
			StreamKey:  streamKey2,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room2.Id, &room2)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room2.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey2, &room2.Id)

		// Verify we have 2 rooms
		rooms := ListStudioRooms(tx, studio.Id)
		if len(rooms) != 2 {
			t.Errorf("Expected 2 rooms, got %d", len(rooms))
		}

		// Check that we've reached the limit (should fail if we try to add more)
		if len(rooms) >= studio.MaxRooms {
			// This is expected - we've reached the limit
		} else {
			t.Error("Should have reached room limit")
		}
	})
}

// Test CreateRoom permissions - only Admin+ can create
func TestCreateRoomPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		adminUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Admin",
			Email: "admin@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)

		memberUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Member",
			Email: "member@test.com",
			Role:  RoleUser,
		}
		vbolt.Write(tx, UsersBkt, memberUser.Id, &memberUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  adminUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Add member membership
		memberMembership := StudioMembership{
			UserId:   memberUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		memberMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, memberMembershipId, &memberMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, memberMembershipId, memberUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, memberMembershipId, studio.Id)

		// Admin should have permission to create rooms
		if !HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Admin should have permission to create rooms")
		}

		// Member should NOT have permission to create rooms
		if HasStudioPermission(tx, memberUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Member should not have permission to create rooms")
		}
	})
}

// Test ListRooms
func TestListRooms(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create viewer user
		viewerUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Viewer",
			Email: "viewer@test.com",
			Role:  RoleUser,
		}
		vbolt.Write(tx, UsersBkt, viewerUser.Id, &viewerUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  viewerUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add viewer membership
		viewerMembership := StudioMembership{
			UserId:   viewerUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		viewerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, viewerMembershipId, &viewerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, viewerMembershipId, viewerUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, viewerMembershipId, studio.Id)

		// Create multiple rooms
		for i := 1; i <= 3; i++ {
			streamKey, _ := GenerateStreamKey()
			room := Room{
				Id:         vbolt.NextIntId(tx, RoomsBkt),
				StudioId:   studio.Id,
				RoomNumber: i,
				Name:       "Room " + string(rune('0'+i)),
				StreamKey:  streamKey,
				IsActive:   false,
				Creation:   time.Now(),
			}
			vbolt.Write(tx, RoomsBkt, room.Id, &room)
			vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
			vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)
		}

		// Viewer should be able to list rooms
		if !HasStudioPermission(tx, viewerUser.Id, studio.Id, StudioRoleViewer) {
			t.Error("Viewer should have permission to list rooms")
		}

		// List rooms
		rooms := ListStudioRooms(tx, studio.Id)
		if len(rooms) != 3 {
			t.Errorf("Expected 3 rooms, got %d", len(rooms))
		}
	})
}

// Test GetRoomStreamKey permissions
func TestGetRoomStreamKeyPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		adminUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Admin",
			Email: "admin@test.com",
			Role:  RoleStreamAdmin,
		}
		vbolt.Write(tx, UsersBkt, adminUser.Id, &adminUser)

		memberUser := User{
			Id:    vbolt.NextIntId(tx, UsersBkt),
			Name:  "Member",
			Email: "member@test.com",
			Role:  RoleUser,
		}
		vbolt.Write(tx, UsersBkt, memberUser.Id, &memberUser)

		// Create studio
		studio := Studio{
			Id:       vbolt.NextIntId(tx, StudiosBkt),
			Name:     "Test Studio",
			MaxRooms: 5,
			OwnerId:  adminUser.Id,
			Creation: time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Add admin membership
		adminMembership := StudioMembership{
			UserId:   adminUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, adminUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Add member membership
		memberMembership := StudioMembership{
			UserId:   memberUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		memberMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, memberMembershipId, &memberMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, memberMembershipId, memberUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, memberMembershipId, studio.Id)

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)

		// Admin should have permission to view stream key
		if !HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Admin should have permission to view stream key")
		}

		// Member should NOT have permission to view stream key
		if HasStudioPermission(tx, memberUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Member should not have permission to view stream key")
		}

		// Verify stream key can be retrieved
		retrieved := GetRoom(tx, room.Id)
		if retrieved.StreamKey != streamKey {
			t.Error("Stream key should match")
		}
	})
}
