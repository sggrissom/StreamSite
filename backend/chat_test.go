package backend

import (
	"stream/cfg"
	"testing"
	"time"

	"go.hasen.dev/vbolt"
)

// Test database setup helper
func setupTestChatDB(t *testing.T) *vbolt.DB {
	dbPath := t.TempDir() + "/test_chat.db"
	db := vbolt.Open(dbPath)
	vbolt.InitBuckets(db, &cfg.Info)
	return db
}

// Test ChatMessage packing/unpacking
func TestPackChatMessage(t *testing.T) {
	db := setupTestChatDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test chat message
		original := ChatMessage{
			Id:           1,
			RoomId:       10,
			UserId:       100,
			SessionToken: "test-session-token",
			UserName:     "Test User",
			Text:         "Hello, world!",
			Timestamp:    time.Now().Truncate(time.Second),
		}

		// Write and read back
		vbolt.Write(tx, ChatMessagesBkt, original.Id, &original)
		var retrieved ChatMessage
		vbolt.Read(tx, ChatMessagesBkt, original.Id, &retrieved)

		// Verify all fields match
		if retrieved.Id != original.Id {
			t.Errorf("Id mismatch: got %d, want %d", retrieved.Id, original.Id)
		}
		if retrieved.RoomId != original.RoomId {
			t.Errorf("RoomId mismatch: got %d, want %d", retrieved.RoomId, original.RoomId)
		}
		if retrieved.UserId != original.UserId {
			t.Errorf("UserId mismatch: got %d, want %d", retrieved.UserId, original.UserId)
		}
		if retrieved.SessionToken != original.SessionToken {
			t.Errorf("SessionToken mismatch: got %s, want %s", retrieved.SessionToken, original.SessionToken)
		}
		if retrieved.UserName != original.UserName {
			t.Errorf("UserName mismatch: got %s, want %s", retrieved.UserName, original.UserName)
		}
		if retrieved.Text != original.Text {
			t.Errorf("Text mismatch: got %s, want %s", retrieved.Text, original.Text)
		}
		if !retrieved.Timestamp.Equal(original.Timestamp) {
			t.Errorf("Timestamp mismatch: got %v, want %v", retrieved.Timestamp, original.Timestamp)
		}
	})
}

func TestDeleteChatMessagesForRoom(t *testing.T) {
	db := setupTestChatDB(t)
	defer db.Close()

	// Create a test room and messages
	var testRoom Room
	var messageIds []int

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		testRoom = Room{
			Id:        vbolt.NextIntId(tx, RoomsBkt),
			Name:      "Test Room",
			StreamKey: "test-stream-key",
			Creation:  time.Now(),
		}
		vbolt.Write(tx, RoomsBkt, testRoom.Id, &testRoom)

		// Create some test messages
		for i := 1; i <= 3; i++ {
			msg := ChatMessage{
				Id:        vbolt.NextIntId(tx, ChatMessagesBkt),
				RoomId:    testRoom.Id,
				UserId:    1,
				UserName:  "Test User",
				Text:      "Test message",
				Timestamp: time.Now(),
			}
			vbolt.Write(tx, ChatMessagesBkt, msg.Id, &msg)
			vbolt.SetTargetSingleTerm(tx, ChatByRoomIdx, msg.Id, msg.RoomId)
			messageIds = append(messageIds, msg.Id)
		}

		vbolt.TxCommit(tx)
	})

	// Delete all messages for the room
	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		err := DeleteChatMessagesForRoom(tx, testRoom.Id)
		if err != nil {
			t.Fatalf("DeleteChatMessagesForRoom failed: %v", err)
		}
		vbolt.TxCommit(tx)
	})

	// Verify messages were deleted
	vbolt.WithReadTx(db, func(tx *vbolt.Tx) {
		for _, msgId := range messageIds {
			var msg ChatMessage
			vbolt.Read(tx, ChatMessagesBkt, msgId, &msg)
			if msg.Id != 0 {
				t.Errorf("Message %d was not deleted", msgId)
			}
		}
	})
}
