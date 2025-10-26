package backend

import (
	"testing"

	"go.hasen.dev/vbolt"
)

// Test CameraConfig packing/unpacking
func TestPackCameraConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create a test camera config
		original := CameraConfig{
			RoomId:  42,
			RTSPURL: "rtsp://192.168.1.100:554/stream1",
		}

		// Write and read back
		vbolt.Write(tx, CameraConfigBkt, original.RoomId, &original)
		var retrieved CameraConfig
		vbolt.Read(tx, CameraConfigBkt, original.RoomId, &retrieved)

		// Verify all fields match
		if retrieved.RoomId != original.RoomId {
			t.Errorf("RoomId mismatch: got %d, want %d", retrieved.RoomId, original.RoomId)
		}
		if retrieved.RTSPURL != original.RTSPURL {
			t.Errorf("RTSPURL mismatch: got %s, want %s", retrieved.RTSPURL, original.RTSPURL)
		}
	})
}

// Test GetCameraConfig helper
func TestGetCameraConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Test getting non-existent config
		config := GetCameraConfig(tx, 999)
		if config.RoomId != 0 {
			t.Error("Expected zero value for non-existent config")
		}

		// Write a config
		SetCameraConfigData(tx, 1, "rtsp://example.com/stream")

		// Read it back
		config = GetCameraConfig(tx, 1)
		if config.RoomId != 1 {
			t.Errorf("Expected RoomId 1, got %d", config.RoomId)
		}
		if config.RTSPURL != "rtsp://example.com/stream" {
			t.Errorf("Expected specific RTSP URL, got %s", config.RTSPURL)
		}
	})
}

// Test SetCameraConfigData helper
func TestSetCameraConfigData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Set config
		SetCameraConfigData(tx, 5, "rtsp://camera.local/stream")

		// Verify it was written
		config := GetCameraConfig(tx, 5)
		if config.RoomId != 5 {
			t.Errorf("Expected RoomId 5, got %d", config.RoomId)
		}
		if config.RTSPURL != "rtsp://camera.local/stream" {
			t.Errorf("Expected specific RTSP URL, got %s", config.RTSPURL)
		}

		// Update existing config
		SetCameraConfigData(tx, 5, "rtsp://camera.local/updated")
		config = GetCameraConfig(tx, 5)
		if config.RTSPURL != "rtsp://camera.local/updated" {
			t.Errorf("Expected updated RTSP URL, got %s", config.RTSPURL)
		}
	})
}

// Test DeleteCameraConfigData helper
func TestDeleteCameraConfigData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Set and then delete
		SetCameraConfigData(tx, 10, "rtsp://test.com/stream")
		config := GetCameraConfig(tx, 10)
		if config.RoomId != 10 {
			t.Error("Expected config to exist after Set")
		}

		DeleteCameraConfigData(tx, 10)
		config = GetCameraConfig(tx, 10)
		if config.RoomId != 0 {
			t.Error("Expected config to be deleted")
		}

		// Delete non-existent config (should not error)
		DeleteCameraConfigData(tx, 999)
	})
}

// Test multiple camera configs
func TestMultipleCameraConfigs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		// Create multiple configs
		SetCameraConfigData(tx, 1, "rtsp://camera1.local/stream")
		SetCameraConfigData(tx, 2, "rtsp://camera2.local/stream")
		SetCameraConfigData(tx, 3, "rtsp://camera3.local/stream")

		// Verify each config
		config1 := GetCameraConfig(tx, 1)
		if config1.RTSPURL != "rtsp://camera1.local/stream" {
			t.Errorf("Config 1 mismatch: got %s", config1.RTSPURL)
		}

		config2 := GetCameraConfig(tx, 2)
		if config2.RTSPURL != "rtsp://camera2.local/stream" {
			t.Errorf("Config 2 mismatch: got %s", config2.RTSPURL)
		}

		config3 := GetCameraConfig(tx, 3)
		if config3.RTSPURL != "rtsp://camera3.local/stream" {
			t.Errorf("Config 3 mismatch: got %s", config3.RTSPURL)
		}

		// Delete one and verify others still exist
		DeleteCameraConfigData(tx, 2)

		config1 = GetCameraConfig(tx, 1)
		if config1.RoomId == 0 {
			t.Error("Config 1 should still exist")
		}

		config2 = GetCameraConfig(tx, 2)
		if config2.RoomId != 0 {
			t.Error("Config 2 should be deleted")
		}

		config3 = GetCameraConfig(tx, 3)
		if config3.RoomId == 0 {
			t.Error("Config 3 should still exist")
		}
	})
}

// Test UpdateCameraConfig (overwrite existing)
func TestUpdateCameraConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	vbolt.WithWriteTx(db, func(tx *vbolt.Tx) {
		roomId := 100

		// Set initial config
		SetCameraConfigData(tx, roomId, "rtsp://old-camera.local/stream")
		config := GetCameraConfig(tx, roomId)
		if config.RTSPURL != "rtsp://old-camera.local/stream" {
			t.Error("Expected initial config to be set")
		}

		// Update config
		SetCameraConfigData(tx, roomId, "rtsp://new-camera.local/stream")
		config = GetCameraConfig(tx, roomId)
		if config.RTSPURL != "rtsp://new-camera.local/stream" {
			t.Error("Expected config to be updated")
		}

		// Update again
		SetCameraConfigData(tx, roomId, "rtsp://newer-camera.local/stream")
		config = GetCameraConfig(tx, roomId)
		if config.RTSPURL != "rtsp://newer-camera.local/stream" {
			t.Error("Expected config to be updated again")
		}
	})
}
