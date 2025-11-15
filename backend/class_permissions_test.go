package backend

import (
	"testing"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

// TestGrantClassPermission tests the GrantClassPermission procedure
func TestGrantClassPermission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var adminUser, regularUser, targetUser User
	var studio Studio
	var room Room
	var schedule ClassSchedule

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		adminUser = createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)
		regularUser = createTestUser(t, tx, "regular@test.com", RoleStreamAdmin)
		targetUser = createTestUser(t, tx, "target@test.com", RoleStreamAdmin)

		// Create studio
		studio = Studio{
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
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create schedule
		schedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room.Id,
			StudioId:        studio.Id,
			Name:            "Test Class",
			Description:     "Test Description",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(25 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			CreatedBy:       adminUser.Id,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, schedule.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule.Id, schedule.StudioId)

		vbolt.TxCommit(tx)
	})

	// Test 1: Grant permission successfully as admin
	t.Run("GrantSuccess", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp GrantClassPermissionResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GrantClassPermissionRequest{
				ScheduleId: schedule.Id,
				UserId:     targetUser.Id,
				Role:       int(StudioRoleMember),
			}
			resp, err = GrantClassPermission(ctx, req)
		})

		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}
		if resp.PermissionId == 0 {
			t.Error("Expected valid permission ID")
		}

		// Verify permission was created
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var perm ClassPermission
			vbolt.Read(tx, ClassPermissionsBkt, resp.PermissionId, &perm)

			if perm.Id == 0 {
				t.Error("Permission should be created")
			}
			if perm.UserId != targetUser.Id {
				t.Errorf("Expected userId %d, got %d", targetUser.Id, perm.UserId)
			}
			if perm.ScheduleId != schedule.Id {
				t.Errorf("Expected scheduleId %d, got %d", schedule.Id, perm.ScheduleId)
			}
			if perm.Role != int(StudioRoleMember) {
				t.Errorf("Expected role %d, got %d", StudioRoleMember, perm.Role)
			}
			if perm.GrantedBy != adminUser.Id {
				t.Errorf("Expected grantedBy %d, got %d", adminUser.Id, perm.GrantedBy)
			}

			// Verify indexes
			var permIds []int
			vbolt.ReadTermTargets(tx, PermsByScheduleIdx, schedule.Id, &permIds, vbolt.Window{})
			if len(permIds) != 1 || permIds[0] != perm.Id {
				t.Errorf("Expected permission in schedule index")
			}

			var userPermIds []int
			vbolt.ReadTermTargets(tx, PermsByUserIdx, targetUser.Id, &userPermIds, vbolt.Window{})
			if len(userPermIds) != 1 || userPermIds[0] != perm.Id {
				t.Errorf("Expected permission in user index")
			}
		})
	})

	// Test 2: Update existing permission (upsert behavior)
	t.Run("UpdateExisting", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Grant permission with viewer role
		var firstResp GrantClassPermissionResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GrantClassPermissionRequest{
				ScheduleId: schedule.Id,
				UserId:     targetUser.Id,
				Role:       int(StudioRoleViewer),
			}
			firstResp, err = GrantClassPermission(ctx, req)
		})

		if err != nil {
			t.Fatalf("First grant failed: %v", err)
		}

		// Update to admin role
		var secondResp GrantClassPermissionResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GrantClassPermissionRequest{
				ScheduleId: schedule.Id,
				UserId:     targetUser.Id,
				Role:       int(StudioRoleAdmin),
			}
			secondResp, err = GrantClassPermission(ctx, req)
		})

		if err != nil {
			t.Fatalf("Second grant failed: %v", err)
		}

		// Should return same permission ID (upsert)
		if firstResp.PermissionId != secondResp.PermissionId {
			t.Errorf("Expected same permission ID, got %d and %d", firstResp.PermissionId, secondResp.PermissionId)
		}

		// Verify role was updated
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var perm ClassPermission
			vbolt.Read(tx, ClassPermissionsBkt, secondResp.PermissionId, &perm)

			if perm.Role != int(StudioRoleAdmin) {
				t.Errorf("Expected role updated to %d, got %d", StudioRoleAdmin, perm.Role)
			}

			// Verify only one permission exists (no duplicate)
			var permIds []int
			vbolt.ReadTermTargets(tx, PermsByScheduleIdx, schedule.Id, &permIds, vbolt.Window{})
			if len(permIds) != 1 {
				t.Errorf("Expected 1 permission, got %d (duplicate created)", len(permIds))
			}
		})
	})

	// Test 3: Non-admin cannot grant permission
	t.Run("NonAdminDenied", func(t *testing.T) {
		token, err := createTestToken(regularUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GrantClassPermissionRequest{
				ScheduleId: schedule.Id,
				UserId:     targetUser.Id,
				Role:       int(StudioRoleMember),
			}
			_, err = GrantClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for non-admin user")
		}
		if err != nil && err.Error() != "admin permission required" {
			t.Errorf("Expected 'admin permission required' error, got: %s", err.Error())
		}
	})

	// Test 4: Invalid schedule ID
	t.Run("InvalidSchedule", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GrantClassPermissionRequest{
				ScheduleId: 99999,
				UserId:     targetUser.Id,
				Role:       int(StudioRoleMember),
			}
			_, err = GrantClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for invalid schedule")
		}
		if err != nil && err.Error() != "schedule not found" {
			t.Errorf("Expected 'schedule not found' error, got: %s", err.Error())
		}
	})

	// Test 5: Invalid role value
	t.Run("InvalidRole", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := GrantClassPermissionRequest{
				ScheduleId: schedule.Id,
				UserId:     targetUser.Id,
				Role:       999, // Invalid role
			}
			_, err = GrantClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for invalid role")
		}
		if err != nil && err.Error() != "invalid role" {
			t.Errorf("Expected 'invalid role' error, got: %s", err.Error())
		}
	})

	// Test 6: Unauthenticated request
	t.Run("Unauthenticated", func(t *testing.T) {
		var err error
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: ""}
			req := GrantClassPermissionRequest{
				ScheduleId: schedule.Id,
				UserId:     targetUser.Id,
				Role:       int(StudioRoleMember),
			}
			_, err = GrantClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for unauthenticated request")
		}
		if err != nil && err.Error() != "authentication required" {
			t.Errorf("Expected 'authentication required' error, got: %s", err.Error())
		}
	})
}

// TestRevokeClassPermission tests the RevokeClassPermission procedure
func TestRevokeClassPermission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var adminUser, regularUser, targetUser User
	var studio Studio
	var room Room
	var schedule ClassSchedule
	var permission ClassPermission

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		adminUser = createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)
		regularUser = createTestUser(t, tx, "regular@test.com", RoleStreamAdmin)
		targetUser = createTestUser(t, tx, "target@test.com", RoleStreamAdmin)

		// Create studio
		studio = Studio{
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
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create schedule
		schedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room.Id,
			StudioId:        studio.Id,
			Name:            "Test Class",
			Description:     "Test Description",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(25 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			CreatedBy:       adminUser.Id,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, schedule.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule.Id, schedule.StudioId)

		// Create permission
		permission = ClassPermission{
			Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
			ScheduleId: schedule.Id,
			UserId:     targetUser.Id,
			Role:       int(StudioRoleMember),
			GrantedBy:  adminUser.Id,
			GrantedAt:  time.Now(),
		}
		vbolt.Write(tx, ClassPermissionsBkt, permission.Id, &permission)
		vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, permission.Id, permission.ScheduleId)
		vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, permission.Id, permission.UserId)

		vbolt.TxCommit(tx)
	})

	// Test 1: Non-admin cannot revoke
	t.Run("NonAdminDenied", func(t *testing.T) {
		token, err := createTestToken(regularUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := RevokeClassPermissionRequest{
				PermissionId: permission.Id,
			}
			_, err = RevokeClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for non-admin user")
		}
		if err != nil && err.Error() != "admin permission required" {
			t.Errorf("Expected 'admin permission required' error, got: %s", err.Error())
		}
	})

	// Test 2: Invalid permission ID
	t.Run("InvalidPermission", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := RevokeClassPermissionRequest{
				PermissionId: 99999,
			}
			_, err = RevokeClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for invalid permission")
		}
		if err != nil && err.Error() != "permission not found" {
			t.Errorf("Expected 'permission not found' error, got: %s", err.Error())
		}
	})

	// Test 3: Unauthenticated request
	t.Run("Unauthenticated", func(t *testing.T) {
		var err error
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: ""}
			req := RevokeClassPermissionRequest{
				PermissionId: permission.Id,
			}
			_, err = RevokeClassPermission(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for unauthenticated request")
		}
		if err != nil && err.Error() != "authentication required" {
			t.Errorf("Expected 'authentication required' error, got: %s", err.Error())
		}
	})

	// Test 4: Revoke permission successfully (run last since it deletes the permission)
	t.Run("RevokeSuccess", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp RevokeClassPermissionResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := RevokeClassPermissionRequest{
				PermissionId: permission.Id,
			}
			resp, err = RevokeClassPermission(ctx, req)
		})

		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}
		if !resp.Success {
			t.Error("Expected Success = true")
		}

		// Verify permission was deleted
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			var perm ClassPermission
			vbolt.Read(tx, ClassPermissionsBkt, permission.Id, &perm)

			if perm.Id != 0 {
				t.Error("Permission should be deleted")
			}

			// Verify indexes were cleared
			var schedulePermIds []int
			vbolt.ReadTermTargets(tx, PermsByScheduleIdx, schedule.Id, &schedulePermIds, vbolt.Window{})
			if len(schedulePermIds) != 0 {
				t.Errorf("Expected 0 permissions in schedule index, got %d", len(schedulePermIds))
			}

			var userPermIds []int
			vbolt.ReadTermTargets(tx, PermsByUserIdx, targetUser.Id, &userPermIds, vbolt.Window{})
			if len(userPermIds) != 0 {
				t.Errorf("Expected 0 permissions in user index, got %d", len(userPermIds))
			}
		})
	})
}

// TestListClassPermissions tests the ListClassPermissions procedure
func TestListClassPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Set up test data
	var adminUser, regularUser, user1, user2, user3 User
	var studio Studio
	var room Room
	var schedule ClassSchedule

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create users
		adminUser = createTestUser(t, tx, "admin@test.com", RoleStreamAdmin)
		regularUser = createTestUser(t, tx, "regular@test.com", RoleStreamAdmin)
		user1 = createTestUser(t, tx, "user1@test.com", RoleStreamAdmin)
		user2 = createTestUser(t, tx, "user2@test.com", RoleStreamAdmin)
		user3 = createTestUser(t, tx, "user3@test.com", RoleStreamAdmin)

		// Create studio
		studio = Studio{
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
		room = Room{
			Id:         vbolt.NextIntId(tx, RoomsBkt),
			StudioId:   studio.Id,
			RoomNumber: 1,
			Name:       "Test Room",
			StreamKey:  streamKey,
			Creation:   time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, room.Id, &room)
		vbolt.SetTargetSingleTerm(tx, RoomsByStudioIdx, room.Id, studio.Id)

		// Create schedule
		schedule = ClassSchedule{
			Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
			RoomId:          room.Id,
			StudioId:        studio.Id,
			Name:            "Test Class",
			Description:     "Test Description",
			IsRecurring:     false,
			StartTime:       time.Now().Add(24 * time.Hour),
			EndTime:         time.Now().Add(25 * time.Hour),
			PreRollMinutes:  5,
			PostRollMinutes: 2,
			AutoStartCamera: true,
			AutoStopCamera:  true,
			CreatedBy:       adminUser.Id,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			IsActive:        true,
		}
		vbolt.Write(tx, ClassSchedulesBkt, schedule.Id, &schedule)
		vbolt.SetTargetSingleTerm(tx, SchedulesByRoomIdx, schedule.Id, schedule.RoomId)
		vbolt.SetTargetSingleTerm(tx, SchedulesByStudioIdx, schedule.Id, schedule.StudioId)

		// Create permissions for three users
		permissions := []ClassPermission{
			{
				Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
				ScheduleId: schedule.Id,
				UserId:     user1.Id,
				Role:       int(StudioRoleViewer),
				GrantedBy:  adminUser.Id,
				GrantedAt:  time.Now(),
			},
			{
				Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
				ScheduleId: schedule.Id,
				UserId:     user2.Id,
				Role:       int(StudioRoleMember),
				GrantedBy:  adminUser.Id,
				GrantedAt:  time.Now(),
			},
			{
				Id:         vbolt.NextIntId(tx, ClassPermissionsBkt),
				ScheduleId: schedule.Id,
				UserId:     user3.Id,
				Role:       int(StudioRoleAdmin),
				GrantedBy:  adminUser.Id,
				GrantedAt:  time.Now(),
			},
		}

		for _, perm := range permissions {
			vbolt.Write(tx, ClassPermissionsBkt, perm.Id, &perm)
			vbolt.SetTargetSingleTerm(tx, PermsByScheduleIdx, perm.Id, perm.ScheduleId)
			vbolt.SetTargetSingleTerm(tx, PermsByUserIdx, perm.Id, perm.UserId)
		}

		vbolt.TxCommit(tx)
	})

	// Test 1: List permissions successfully
	t.Run("ListSuccess", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		var resp ListClassPermissionsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := ListClassPermissionsRequest{
				ScheduleId: schedule.Id,
			}
			resp, err = ListClassPermissions(ctx, req)
		})

		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}

		if len(resp.Permissions) != 3 {
			t.Errorf("Expected 3 permissions, got %d", len(resp.Permissions))
		}

		// Verify user info is included
		for _, permWithUser := range resp.Permissions {
			if permWithUser.UserName == "" {
				t.Error("Expected user name to be populated")
			}
			if permWithUser.UserEmail == "" {
				t.Error("Expected user email to be populated")
			}
			if permWithUser.Permission.Id == 0 {
				t.Error("Expected permission to be populated")
			}
		}

		// Verify specific users
		foundUsers := make(map[int]bool)
		for _, permWithUser := range resp.Permissions {
			foundUsers[permWithUser.Permission.UserId] = true
		}

		if !foundUsers[user1.Id] || !foundUsers[user2.Id] || !foundUsers[user3.Id] {
			t.Error("Expected all three users in permission list")
		}
	})

	// Test 2: Empty list for schedule with no permissions
	t.Run("EmptyList", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Create schedule with no permissions
		var emptySchedule ClassSchedule
		vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
			emptySchedule = ClassSchedule{
				Id:              vbolt.NextIntId(tx, ClassSchedulesBkt),
				RoomId:          room.Id,
				StudioId:        studio.Id,
				Name:            "Empty Class",
				IsRecurring:     false,
				StartTime:       time.Now().Add(24 * time.Hour),
				EndTime:         time.Now().Add(25 * time.Hour),
				PreRollMinutes:  5,
				PostRollMinutes: 2,
				CreatedBy:       adminUser.Id,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
				IsActive:        true,
			}
			vbolt.Write(tx, ClassSchedulesBkt, emptySchedule.Id, &emptySchedule)
			vbolt.TxCommit(tx)
		})

		var resp ListClassPermissionsResponse
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := ListClassPermissionsRequest{
				ScheduleId: emptySchedule.Id,
			}
			resp, err = ListClassPermissions(ctx, req)
		})

		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}

		if len(resp.Permissions) != 0 {
			t.Errorf("Expected 0 permissions, got %d", len(resp.Permissions))
		}
	})

	// Test 3: Non-admin cannot list
	t.Run("NonAdminDenied", func(t *testing.T) {
		token, err := createTestToken(regularUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := ListClassPermissionsRequest{
				ScheduleId: schedule.Id,
			}
			_, err = ListClassPermissions(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for non-admin user")
		}
		if err != nil && err.Error() != "admin permission required" {
			t.Errorf("Expected 'admin permission required' error, got: %s", err.Error())
		}
	})

	// Test 4: Invalid schedule ID
	t.Run("InvalidSchedule", func(t *testing.T) {
		token, err := createTestToken(adminUser.Id)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: token}
			req := ListClassPermissionsRequest{
				ScheduleId: 99999,
			}
			_, err = ListClassPermissions(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for invalid schedule")
		}
		if err != nil && err.Error() != "schedule not found" {
			t.Errorf("Expected 'schedule not found' error, got: %s", err.Error())
		}
	})

	// Test 5: Unauthenticated request
	t.Run("Unauthenticated", func(t *testing.T) {
		var err error
		vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
			ctx := &vbeam.Context{Tx: tx, Token: ""}
			req := ListClassPermissionsRequest{
				ScheduleId: schedule.Id,
			}
			_, err = ListClassPermissions(ctx, req)
		})

		if err == nil {
			t.Error("Expected failure for unauthenticated request")
		}
		if err != nil && err.Error() != "authentication required" {
			t.Errorf("Expected 'authentication required' error, got: %s", err.Error())
		}
	})
}
