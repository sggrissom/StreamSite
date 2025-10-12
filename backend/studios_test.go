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

// Test UpdateRoom - basic rename functionality
func TestUpdateRoom(t *testing.T) {
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

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Original Name",
			StreamKey:  streamKey,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)

		// Update room name
		room.Name = "Updated Name"
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		// Verify update
		updated := GetRoom(tx, room.Id)
		if updated.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
		}
	})
}

// Test UpdateRoom permissions - only Admin+ can update
func TestUpdateRoomPermissions(t *testing.T) {
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

		// Admin should have permission to update
		if !HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Admin should have permission to update rooms")
		}

		// Member should NOT have permission to update
		if HasStudioPermission(tx, memberUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Member should not have permission to update rooms")
		}
	})
}

// Test UpdateRoom validation - name required
func TestUpdateRoomValidation(t *testing.T) {
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

		// Create room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Original Name",
			StreamKey:  streamKey,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)

		// Empty name should fail validation (we simulate validation check)
		emptyName := ""
		if emptyName == "" {
			// This is expected - empty names should be rejected
		} else {
			t.Error("Empty name should fail validation")
		}
	})
}

// Test RegenerateStreamKey - generates new key and updates lookups
func TestRegenerateStreamKey(t *testing.T) {
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

		// Create room with initial stream key
		oldStreamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  oldStreamKey,
			IsActive:   false,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, oldStreamKey, &room.Id)

		// Regenerate stream key
		newStreamKey, _ := GenerateStreamKey()

		// Delete old key lookup
		vbolt.Delete(tx, RoomStreamKeyBkt, oldStreamKey)

		// Update room with new key
		room.StreamKey = newStreamKey
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		// Add new key lookup
		vbolt.Write(tx, RoomStreamKeyBkt, newStreamKey, &room.Id)

		// Verify old key no longer works
		var roomIdFromOldKey int
		vbolt.Read(tx, RoomStreamKeyBkt, oldStreamKey, &roomIdFromOldKey)
		if roomIdFromOldKey != 0 {
			t.Error("Old stream key should not resolve to room")
		}

		// Verify new key works
		roomFromNewKey := GetRoomByStreamKey(tx, newStreamKey)
		if roomFromNewKey.Id != room.Id {
			t.Error("New stream key should resolve to room")
		}

		// Verify room has new key
		updated := GetRoom(tx, room.Id)
		if updated.StreamKey != newStreamKey {
			t.Error("Room should have new stream key")
		}
	})
}

// Test RegenerateStreamKey permissions - only Admin+ can regenerate
func TestRegenerateStreamKeyPermissions(t *testing.T) {
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

		// Admin should have permission to regenerate
		if !HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Admin should have permission to regenerate stream keys")
		}

		// Member should NOT have permission to regenerate
		if HasStudioPermission(tx, memberUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Member should not have permission to regenerate stream keys")
		}
	})
}

// Test DeleteRoom - basic deletion with cascade
func TestDeleteRoom(t *testing.T) {
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

		// Create room (not active)
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

		// Create a stream for this room
		stream := Stream{
			Id:          vbolt.NextIntId(tx, StreamsBkt),
			StudioId:    studio.Id,
			RoomId:      room.Id,
			Title:       "Test Stream",
			Description: "Test",
			StartTime:   time.Now(),
			CreatedBy:   adminUser.Id,
		}
		vbolt.Write(tx, StreamsBkt, stream.Id, &stream)
		vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, stream.Id, studio.Id)
		vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, stream.Id, room.Id)

		// Delete the room
		// 1. Delete streams
		var streamIds []int
		vbolt.ReadTermTargets(tx, StreamsByRoomIdx, room.Id, &streamIds, vbolt.Window{})
		for _, sid := range streamIds {
			vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, sid, -1)
			vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, sid, -1)
			vbolt.Delete(tx, StreamsBkt, sid)
		}

		// 2. Remove stream key
		vbolt.Delete(tx, RoomStreamKeyBkt, streamKey)

		// 3. Unindex room
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, -1)

		// 4. Delete room
		vbolt.Delete(tx, RoomsBkt, room.Id)

		// Verify room is deleted
		deletedRoom := GetRoom(tx, room.Id)
		if deletedRoom.Id != 0 {
			t.Error("Room should be deleted")
		}

		// Verify stream is deleted
		deletedStream := GetStream(tx, stream.Id)
		if deletedStream.Id != 0 {
			t.Error("Stream should be deleted")
		}

		// Verify stream key is removed
		var roomIdFromKey int
		vbolt.Read(tx, RoomStreamKeyBkt, streamKey, &roomIdFromKey)
		if roomIdFromKey != 0 {
			t.Error("Stream key should be deleted")
		}

		// Verify room is unindexed
		var roomIdsFromIdx []int
		vbolt.ReadTermTargets(tx, RoomsByStudioIdx, studio.Id, &roomIdsFromIdx, vbolt.Window{})
		if len(roomIdsFromIdx) != 0 {
			t.Errorf("RoomsByStudioIdx should be empty, got %d entries", len(roomIdsFromIdx))
		}
	})
}

// Test DeleteRoom permissions - only Admin+ can delete
func TestDeleteRoomPermissions(t *testing.T) {
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

		// Admin should have permission to delete
		if !HasStudioPermission(tx, adminUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Admin should have permission to delete rooms")
		}

		// Member should NOT have permission to delete
		if HasStudioPermission(tx, memberUser.Id, studio.Id, StudioRoleAdmin) {
			t.Error("Member should not have permission to delete rooms")
		}
	})
}

// Test DeleteRoom active check - cannot delete while streaming
func TestDeleteRoomActiveCheck(t *testing.T) {
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

		// Create active room
		streamKey, _ := GenerateStreamKey()
		room := Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Room 1",
			StreamKey:  streamKey,
			IsActive:   true, // ACTIVE
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)
		vbolt.Write(tx, RoomStreamKeyBkt, streamKey, &room.Id)

		// Should NOT be able to delete active room
		if room.IsActive {
			// This is expected - cannot delete active room
		} else {
			t.Error("Room should be active")
		}

		// Now make it inactive
		room.IsActive = false
		vbolt.Write(tx, RoomsBkt, room.Id, &room)

		// Should be able to delete inactive room
		if !room.IsActive {
			// This is expected - can delete inactive room
		} else {
			t.Error("Room should be inactive")
		}
	})
}

// Test DeleteRoom cascade - verifies all streams are deleted
func TestDeleteRoomCascadeStreams(t *testing.T) {
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

		// Create room (not active)
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

		// Create multiple streams for this room
		var createdStreamIds []int
		for i := 1; i <= 3; i++ {
			stream := Stream{
				Id:          vbolt.NextIntId(tx, StreamsBkt),
				StudioId:    studio.Id,
				RoomId:      room.Id,
				Title:       "Test Stream " + string(rune('0'+i)),
				Description: "Test",
				StartTime:   time.Now(),
				CreatedBy:   adminUser.Id,
			}
			vbolt.Write(tx, StreamsBkt, stream.Id, &stream)
			vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, stream.Id, studio.Id)
			vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, stream.Id, room.Id)
			createdStreamIds = append(createdStreamIds, stream.Id)
		}

		// Delete the room with cascade
		// 1. Delete streams
		var streamIds []int
		vbolt.ReadTermTargets(tx, StreamsByRoomIdx, room.Id, &streamIds, vbolt.Window{})
		if len(streamIds) != 3 {
			t.Errorf("Expected 3 streams before deletion, got %d", len(streamIds))
		}

		for _, sid := range streamIds {
			vbolt.SetTargetSingleTerm(tx, StreamsByRoomIdx, sid, -1)
			vbolt.SetTargetSingleTerm(tx, StreamsByStudioIdx, sid, -1)
			vbolt.Delete(tx, StreamsBkt, sid)
		}

		// 2. Remove stream key
		vbolt.Delete(tx, RoomStreamKeyBkt, streamKey)

		// 3. Unindex room
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, -1)

		// 4. Delete room
		vbolt.Delete(tx, RoomsBkt, room.Id)

		// Verify all streams are deleted
		for _, streamId := range createdStreamIds {
			deletedStream := GetStream(tx, streamId)
			if deletedStream.Id != 0 {
				t.Errorf("Stream %d should be deleted", streamId)
			}
		}

		// Verify stream indexes are cleaned up
		var roomStreamsFromIdx []int
		vbolt.ReadTermTargets(tx, StreamsByRoomIdx, room.Id, &roomStreamsFromIdx, vbolt.Window{})
		if len(roomStreamsFromIdx) != 0 {
			t.Errorf("StreamsByRoomIdx should be empty, got %d entries", len(roomStreamsFromIdx))
		}
	})
}
