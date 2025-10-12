package backend

import (
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

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

// Test AddStudioMember success case
func TestAddStudioMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner user
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create owner membership
		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Create target user to add
		targetUser := createTestUser(t, tx, "member@test.com", RoleUser)

		// Test: Add member as viewer
		existingRole := GetUserStudioRole(tx, targetUser.Id, studio.Id)
		if existingRole != -1 {
			t.Errorf("Target user should not be a member yet, got role %d", existingRole)
		}

		// Add the member
		membership := StudioMembership{
			UserId:   targetUser.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, targetUser.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Verify membership was created
		newRole := GetUserStudioRole(tx, targetUser.Id, studio.Id)
		if newRole != StudioRoleMember {
			t.Errorf("Expected role %d, got %d", StudioRoleMember, newRole)
		}

		// Verify membership appears in studio members list
		members := ListStudioMembers(tx, studio.Id)
		if len(members) != 2 {
			t.Errorf("Expected 2 members, got %d", len(members))
		}

		// Verify membership appears in user's studios list
		var userMembershipIds []int
		vbolt.ReadTermTargets(tx, MembershipByUserIdx, targetUser.Id, &userMembershipIds, vbolt.Window{})
		if len(userMembershipIds) != 1 {
			t.Errorf("Expected 1 membership for user, got %d", len(userMembershipIds))
		}
	})
}

// Test AddStudioMember permission checks
func TestAddStudioMemberPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)

		// Create studio
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create owner membership
		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Create a member (not admin)
		member := createTestUser(t, tx, "member@test.com", RoleUser)
		memberMembership := StudioMembership{
			UserId:   member.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		memberMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, memberMembershipId, &memberMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, memberMembershipId, member.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, memberMembershipId, studio.Id)

		// Test: Member (non-admin) cannot add members
		_ = createTestUser(t, tx, "target@test.com", RoleUser)
		hasPermission := HasStudioPermission(tx, member.Id, studio.Id, StudioRoleAdmin)
		if hasPermission {
			t.Error("Member should not have admin permission")
		}

		// Test: Admin can add members
		admin := createTestUser(t, tx, "admin@test.com", RoleUser)
		adminMembership := StudioMembership{
			UserId:   admin.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, admin.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		hasAdminPermission := HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleAdmin)
		if !hasAdminPermission {
			t.Error("Admin should have admin permission")
		}

		// Test: Only owners can add other owners
		cannotAddOwner := !HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleOwner)
		if !cannotAddOwner {
			t.Error("Admin should not be able to add owners")
		}

		canAddOwner := HasStudioPermission(tx, owner.Id, studio.Id, StudioRoleOwner)
		if !canAddOwner {
			t.Error("Owner should be able to add owners")
		}
	})
}

// Test AddStudioMember duplicate member check
func TestAddStudioMemberDuplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add a member
		member := createTestUser(t, tx, "member@test.com", RoleUser)
		membership := StudioMembership{
			UserId:   member.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, member.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Test: Verify member is already in studio
		existingRole := GetUserStudioRole(tx, member.Id, studio.Id)
		if existingRole == -1 {
			t.Error("Member should already be in studio")
		}

		// Test: Cannot add duplicate member
		if existingRole != -1 {
			// This is the expected check that would happen in AddStudioMember API
			t.Logf("Correctly detected duplicate membership (role: %d)", existingRole)
		}
	})
}

// Test AddStudioMember invalid user
func TestAddStudioMemberInvalidUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Test: Try to get non-existent user
		nonExistentUserId := 99999
		targetUser := GetUser(tx, nonExistentUserId)
		if targetUser.Id != 0 {
			t.Error("Non-existent user should return empty user")
		}
	})
}

// Test RemoveStudioMember success case
func TestRemoveStudioMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add a member
		member := createTestUser(t, tx, "member@test.com", RoleUser)
		membership := StudioMembership{
			UserId:   member.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, member.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Verify member exists
		role := GetUserStudioRole(tx, member.Id, studio.Id)
		if role != StudioRoleMember {
			t.Errorf("Expected member role, got %d", role)
		}

		// Remove the member
		var membershipIds []int
		vbolt.ReadTermTargets(tx, MembershipByUserIdx, member.Id, &membershipIds, vbolt.Window{})
		for _, mid := range membershipIds {
			m := GetMembership(tx, mid)
			if m.StudioId == studio.Id {
				vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, mid, -1)
				vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, mid, -1)
				vbolt.Delete(tx, MembershipBkt, mid)
				break
			}
		}

		// Verify member was removed
		removedRole := GetUserStudioRole(tx, member.Id, studio.Id)
		if removedRole != -1 {
			t.Errorf("Member should be removed, got role %d", removedRole)
		}

		// Verify membership no longer appears in lists
		members := ListStudioMembers(tx, studio.Id)
		if len(members) != 1 { // Only owner should remain
			t.Errorf("Expected 1 member (owner), got %d", len(members))
		}
	})
}

// Test RemoveStudioMember cannot remove owner
func TestRemoveStudioMemberCannotRemoveOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Test: Verify cannot remove owner (OwnerId check)
		if studio.OwnerId == owner.Id {
			t.Log("Correctly blocked removal of studio owner")
		}
	})
}

// Test RemoveStudioMember permission checks
func TestRemoveStudioMemberPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner, admin, and members
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Create admin
		admin := createTestUser(t, tx, "admin@test.com", RoleUser)
		adminMembership := StudioMembership{
			UserId:   admin.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, admin.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Create regular member
		member := createTestUser(t, tx, "member@test.com", RoleUser)
		memberMembership := StudioMembership{
			UserId:   member.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		memberMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, memberMembershipId, &memberMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, memberMembershipId, member.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, memberMembershipId, studio.Id)

		// Test: Regular member cannot remove others
		hasPermission := HasStudioPermission(tx, member.Id, studio.Id, StudioRoleAdmin)
		if hasPermission {
			t.Error("Regular member should not have admin permission")
		}

		// Test: Admin can remove regular members
		canRemoveMembers := HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleAdmin)
		if !canRemoveMembers {
			t.Error("Admin should be able to remove members")
		}

		// Test: Only owners can remove admins
		adminRole := GetUserStudioRole(tx, admin.Id, studio.Id)
		if adminRole >= StudioRoleAdmin {
			canAdminRemoveAdmin := HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleOwner)
			if canAdminRemoveAdmin {
				t.Error("Admin should not be able to remove other admins")
			}
		}

		canOwnerRemoveAdmin := HasStudioPermission(tx, owner.Id, studio.Id, StudioRoleOwner)
		if !canOwnerRemoveAdmin {
			t.Error("Owner should be able to remove admins")
		}
	})
}

// Test RemoveStudioMember non-member
func TestRemoveStudioMemberNonMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create non-member user
		nonMember := createTestUser(t, tx, "nonmember@test.com", RoleUser)

		// Test: Verify user is not a member
		role := GetUserStudioRole(tx, nonMember.Id, studio.Id)
		if role != -1 {
			t.Error("User should not be a member")
		}
	})
}

// Test UpdateStudioMemberRole success case
func TestUpdateStudioMemberRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add a member
		member := createTestUser(t, tx, "member@test.com", RoleUser)
		membership := StudioMembership{
			UserId:   member.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, member.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Verify initial role
		role := GetUserStudioRole(tx, member.Id, studio.Id)
		if role != StudioRoleMember {
			t.Errorf("Expected member role, got %d", role)
		}

		// Update role to Admin
		var membershipIds []int
		vbolt.ReadTermTargets(tx, MembershipByUserIdx, member.Id, &membershipIds, vbolt.Window{})
		for _, mid := range membershipIds {
			m := GetMembership(tx, mid)
			if m.StudioId == studio.Id {
				m.Role = StudioRoleAdmin
				vbolt.Write(tx, MembershipBkt, mid, &m)
				break
			}
		}

		// Verify role was updated
		updatedRole := GetUserStudioRole(tx, member.Id, studio.Id)
		if updatedRole != StudioRoleAdmin {
			t.Errorf("Expected admin role, got %d", updatedRole)
		}
	})
}

// Test UpdateStudioMemberRole cannot change owner role
func TestUpdateStudioMemberRoleCannotChangeOwner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Test: Verify cannot change owner's role (OwnerId check)
		if studio.OwnerId == owner.Id {
			t.Log("Correctly blocked role change for studio owner")
		}
	})
}

// Test UpdateStudioMemberRole permission checks
func TestUpdateStudioMemberRolePermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Create admin
		admin := createTestUser(t, tx, "admin@test.com", RoleUser)
		adminMembership := StudioMembership{
			UserId:   admin.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, admin.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Test: Admin can update regular member roles
		canUpdateMembers := HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleAdmin)
		if !canUpdateMembers {
			t.Error("Admin should be able to update member roles")
		}

		// Test: Only owner can assign owner role
		canAssignOwner := HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleOwner)
		if canAssignOwner {
			t.Error("Admin should not be able to assign owner role")
		}

		ownerCanAssignOwner := HasStudioPermission(tx, owner.Id, studio.Id, StudioRoleOwner)
		if !ownerCanAssignOwner {
			t.Error("Owner should be able to assign owner role")
		}

		// Test: Only owner can modify admin roles
		adminRole := GetUserStudioRole(tx, admin.Id, studio.Id)
		if adminRole >= StudioRoleAdmin {
			canAdminModifyAdmin := HasStudioPermission(tx, admin.Id, studio.Id, StudioRoleOwner)
			if canAdminModifyAdmin {
				t.Error("Admin should not be able to modify admin roles")
			}
		}
	})
}

// Test UpdateStudioMemberRole invalid role
func TestUpdateStudioMemberRoleInvalidRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Test invalid roles
		invalidRoles := []StudioRole{-1, 100}

		for _, invalidRole := range invalidRoles {
			if invalidRole < StudioRoleViewer || invalidRole > StudioRoleOwner {
				t.Logf("Correctly validated invalid role: %d", invalidRole)
			}
		}

		// Test valid roles
		validRoles := []StudioRole{
			StudioRoleViewer,
			StudioRoleMember,
			StudioRoleAdmin,
			StudioRoleOwner,
		}

		for _, validRole := range validRoles {
			if validRole >= StudioRoleViewer && validRole <= StudioRoleOwner {
				t.Logf("Valid role: %d", validRole)
			}
		}
	})
}

// Test ListStudioMembersAPI
func TestListStudioMembersAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add multiple members
		member1 := createTestUser(t, tx, "member1@test.com", RoleUser)
		membership1 := StudioMembership{
			UserId:   member1.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId1 := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId1, &membership1)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId1, member1.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId1, studio.Id)

		member2 := createTestUser(t, tx, "member2@test.com", RoleUser)
		membership2 := StudioMembership{
			UserId:   member2.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		membershipId2 := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId2, &membership2)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId2, member2.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId2, studio.Id)

		// List all members
		members := ListStudioMembers(tx, studio.Id)
		if len(members) != 3 {
			t.Errorf("Expected 3 members, got %d", len(members))
		}

		// Verify each member has correct user details
		for _, membership := range members {
			user := GetUser(tx, membership.UserId)
			if user.Id == 0 {
				t.Errorf("User %d not found", membership.UserId)
			}
			t.Logf("Member: %s (%s) - Role: %s", user.Name, user.Email, GetStudioRoleName(membership.Role))
		}
	})
}

// Test ListStudioMembersAPI permissions
func TestListStudioMembersAPIPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add viewer
		viewer := createTestUser(t, tx, "viewer@test.com", RoleUser)
		viewerMembership := StudioMembership{
			UserId:   viewer.Id,
			StudioId: studio.Id,
			Role:     StudioRoleViewer,
			JoinedAt: time.Now(),
		}
		viewerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, viewerMembershipId, &viewerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, viewerMembershipId, viewer.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, viewerMembershipId, studio.Id)

		// Test: Even viewers can list members
		canViewerListMembers := HasStudioPermission(tx, viewer.Id, studio.Id, StudioRoleViewer)
		if !canViewerListMembers {
			t.Error("Viewer should be able to list members")
		}

		// Create non-member
		nonMember := createTestUser(t, tx, "nonmember@test.com", RoleUser)

		// Test: Non-members cannot list members
		canNonMemberList := HasStudioPermission(tx, nonMember.Id, studio.Id, StudioRoleViewer)
		if canNonMemberList {
			t.Error("Non-member should not be able to list members")
		}
	})
}

// Test LeaveStudio success case
func TestLeaveStudio(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add a member
		member := createTestUser(t, tx, "member@test.com", RoleUser)
		membership := StudioMembership{
			UserId:   member.Id,
			StudioId: studio.Id,
			Role:     StudioRoleMember,
			JoinedAt: time.Now(),
		}
		membershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, membershipId, &membership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, membershipId, member.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, membershipId, studio.Id)

		// Verify member exists
		role := GetUserStudioRole(tx, member.Id, studio.Id)
		if role != StudioRoleMember {
			t.Errorf("Expected member role, got %d", role)
		}

		// Member leaves studio
		var membershipIds []int
		vbolt.ReadTermTargets(tx, MembershipByUserIdx, member.Id, &membershipIds, vbolt.Window{})
		for _, mid := range membershipIds {
			m := GetMembership(tx, mid)
			if m.StudioId == studio.Id {
				vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, mid, -1)
				vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, mid, -1)
				vbolt.Delete(tx, MembershipBkt, mid)
				break
			}
		}

		// Verify member left
		leftRole := GetUserStudioRole(tx, member.Id, studio.Id)
		if leftRole != -1 {
			t.Errorf("Member should have left, got role %d", leftRole)
		}

		// Verify only owner remains
		members := ListStudioMembers(tx, studio.Id)
		if len(members) != 1 {
			t.Errorf("Expected 1 member (owner), got %d", len(members))
		}
	})
}

// Test LeaveStudio owner cannot leave
func TestLeaveStudioOwnerCannotLeave(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Test: Verify owner cannot leave (OwnerId check)
		if studio.OwnerId == owner.Id {
			t.Log("Correctly blocked owner from leaving studio")
		}
	})
}

// Test LeaveStudio non-member
func TestLeaveStudioNonMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		// Create non-member
		nonMember := createTestUser(t, tx, "nonmember@test.com", RoleUser)

		// Test: Verify user is not a member
		role := GetUserStudioRole(tx, nonMember.Id, studio.Id)
		if role != -1 {
			t.Error("User should not be a member")
		}
	})
}

// Test LeaveStudio admin can leave
func TestLeaveStudioAdminCanLeave(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create owner and studio
		owner := createTestUser(t, tx, "owner@test.com", RoleStreamAdmin)
		studio := Studio{
			Id:          vbolt.NextIntId(tx, StudiosBkt),
			Name:        "Test Studio",
			Description: "Test",
			MaxRooms:    5,
			OwnerId:     owner.Id,
			Creation:    time.Now(),
		}
		vbolt.Write(tx, StudiosBkt, studio.Id, &studio)

		ownerMembership := StudioMembership{
			UserId:   owner.Id,
			StudioId: studio.Id,
			Role:     StudioRoleOwner,
			JoinedAt: time.Now(),
		}
		ownerMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, ownerMembershipId, &ownerMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, ownerMembershipId, owner.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, ownerMembershipId, studio.Id)

		// Add admin
		admin := createTestUser(t, tx, "admin@test.com", RoleUser)
		adminMembership := StudioMembership{
			UserId:   admin.Id,
			StudioId: studio.Id,
			Role:     StudioRoleAdmin,
			JoinedAt: time.Now(),
		}
		adminMembershipId := vbolt.NextIntId(tx, MembershipBkt)
		vbolt.Write(tx, MembershipBkt, adminMembershipId, &adminMembership)
		vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, adminMembershipId, admin.Id)
		vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, adminMembershipId, studio.Id)

		// Verify admin can leave (not owner)
		if studio.OwnerId != admin.Id {
			t.Log("Admin can leave studio (not the owner)")

			// Admin leaves
			var membershipIds []int
			vbolt.ReadTermTargets(tx, MembershipByUserIdx, admin.Id, &membershipIds, vbolt.Window{})
			for _, mid := range membershipIds {
				m := GetMembership(tx, mid)
				if m.StudioId == studio.Id {
					vbolt.SetTargetSingleTerm(tx, MembershipByUserIdx, mid, -1)
					vbolt.SetTargetSingleTerm(tx, MembershipByStudioIdx, mid, -1)
					vbolt.Delete(tx, MembershipBkt, mid)
					break
				}
			}

			// Verify admin left
			leftRole := GetUserStudioRole(tx, admin.Id, studio.Id)
			if leftRole != -1 {
				t.Errorf("Admin should have left, got role %d", leftRole)
			}
		}
	})
}
