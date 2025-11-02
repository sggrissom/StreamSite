package backend

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateStreamKey(t *testing.T) {
	tests := []struct {
		name      string
		streamKey string
		shouldErr bool
	}{
		{"Valid alphanumeric", "abc123XYZ", false},
		{"Valid with hyphen", "stream-key-123", false},
		{"Valid with underscore", "stream_key_123", false},
		{"Empty key", "", true},
		{"Shell injection semicolon", "key;rm -rf", true},
		{"Shell injection pipe", "key|cat", true},
		{"Shell injection backtick", "key`whoami`", true},
		{"Path traversal", "../../../etc/passwd", true},
		{"Space character", "stream key", true},
		{"Quote character", "stream'key", true},
		{"Dollar sign", "stream$key", true},
		{"Newline", "stream\nkey", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStreamKey(tt.streamKey)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for stream key %q, but got nil", tt.streamKey)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error for stream key %q, but got: %v", tt.streamKey, err)
			}
		})
	}
}

func TestNewTranscoderManager(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	if manager == nil {
		t.Fatal("NewTranscoderManager returned nil")
	}

	if manager.config.HLSBaseDir != tmpDir {
		t.Errorf("Expected HLS base dir %s, got %s", tmpDir, manager.config.HLSBaseDir)
	}

	if manager.transcoders == nil {
		t.Error("Transcoders map not initialized")
	}

	// Verify directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Errorf("HLS base directory was not created: %s", tmpDir)
	}
}

func TestTranscoderManagerResourceLimits(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore original value
	originalLimit := MaxConcurrentTranscoders
	MaxConcurrentTranscoders = 2
	defer func() { MaxConcurrentTranscoders = originalLimit }()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	// Manually add mock transcoders to test limit without starting FFmpeg
	manager.transcoders["room1"] = &Transcoder{roomID: "room1"}
	manager.transcoders["room2"] = &Transcoder{roomID: "room2"}

	// Now try to start another - should hit the limit
	err := manager.Start("room3", "valid-key")
	if err == nil {
		t.Error("Expected error when exceeding max concurrent transcoders")
	}

	if !strings.Contains(err.Error(), "transcoder limit reached") {
		t.Errorf("Expected 'transcoder limit reached' error, got: %v", err)
	}
}

func TestTranscoderManagerRoomIDValidation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	tests := []struct {
		name      string
		roomID    string
		shouldErr bool
		errMsg    string
	}{
		{"Path traversal dots", "..", true, "invalid room ID"},
		{"Path traversal relative", "../etc", true, "invalid room ID"},
		{"Absolute path", "/etc/passwd", true, "invalid room ID"},
		{"Forward slash", "room/123", true, "invalid room ID"},
		{"Backslash", "room\\123", true, "invalid room ID"},
		{"Current directory", ".", true, "invalid room ID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Start(tt.roomID, "valid-key")

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for room ID %q, but got nil", tt.roomID)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q for room ID %q, got: %v", tt.errMsg, tt.roomID, err)
				}
			}
		})
	}
}

func TestTranscoderManagerStreamKeyValidation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	tests := []struct {
		name      string
		streamKey string
		shouldErr bool
		errMsg    string
	}{
		{"Empty key", "", true, "invalid stream key"},
		{"Shell metachar semicolon", "key;cmd", true, "invalid stream key"},
		{"Shell metachar pipe", "key|cmd", true, "invalid stream key"},
		{"Shell metachar backtick", "key`cmd`", true, "invalid stream key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Start("test-room", tt.streamKey)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error for stream key %q, but got nil", tt.streamKey)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q for stream key %q, got: %v", tt.errMsg, tt.streamKey, err)
				}
			}
		})
	}
}

func TestTranscoderManagerDuplicatePrevention(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	// Manually add a transcoder to the map (simulating running transcoder)
	tc := &Transcoder{
		roomID:    "test-room",
		streamKey: "test-key",
	}
	manager.transcoders["test-room"] = tc

	// Try to start another transcoder for the same room
	err := manager.Start("test-room", "different-key")

	// Should not error (returns nil for duplicates)
	if err != nil {
		t.Errorf("Expected nil error for duplicate start, got: %v", err)
	}

	// Verify only one transcoder exists
	if len(manager.transcoders) != 1 {
		t.Errorf("Expected 1 transcoder, got %d", len(manager.transcoders))
	}
}

func TestTranscoderManagerStopCleansUp(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	// Manually add a transcoder
	tc := &Transcoder{
		roomID:    "test-room",
		streamKey: "test-key",
		cancel:    func() {}, // Mock cancel function
	}
	manager.transcoders["test-room"] = tc

	// Stop the transcoder
	manager.Stop("test-room")

	// Verify it was removed
	if len(manager.transcoders) != 0 {
		t.Errorf("Expected 0 transcoders after stop, got %d", len(manager.transcoders))
	}

	// Verify we can query it
	if manager.IsRunning("test-room") {
		t.Error("Transcoder should not be running after stop")
	}
}

func TestTranscoderManagerGetActiveCount(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	// Initially should be 0
	if count := manager.GetActiveCount(); count != 0 {
		t.Errorf("Expected 0 active transcoders, got %d", count)
	}

	// Add some transcoders
	manager.transcoders["room1"] = &Transcoder{roomID: "room1"}
	manager.transcoders["room2"] = &Transcoder{roomID: "room2"}

	if count := manager.GetActiveCount(); count != 2 {
		t.Errorf("Expected 2 active transcoders, got %d", count)
	}
}

func TestCleanupOrphanedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some orphaned directories
	os.MkdirAll(filepath.Join(tmpDir, "room1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "room2"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "room3", "0"), 0755)

	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	manager := NewTranscoderManager(cfg)

	// Clean up
	manager.CleanupOrphanedDirectories()

	// Verify directories were removed
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp dir: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 directories after cleanup, got %d", len(entries))
	}
}

func TestGetTranscoderHealthProcedure(t *testing.T) {
	// Save original manager
	origManager := transcoderManager
	defer func() { transcoderManager = origManager }()

	// Test with nil manager
	transcoderManager = nil
	resp, err := GetTranscoderHealth(nil, GetTranscoderHealthRequest{})

	if err != nil {
		t.Errorf("Expected no error with nil manager, got: %v", err)
	}

	if resp.Active != 0 {
		t.Errorf("Expected 0 active with nil manager, got %d", resp.Active)
	}

	if !resp.Healthy {
		t.Error("Expected healthy=true with nil manager")
	}

	// Test with actual manager
	tmpDir := t.TempDir()
	cfg := TranscoderConfig{
		HLSBaseDir:  tmpDir,
		SRSRTMPBase: "rtmp://localhost:1935/live",
	}

	transcoderManager = NewTranscoderManager(cfg)

	// Add a mock transcoder
	startTime := time.Now().Add(-5 * time.Minute)
	transcoderManager.transcoders["test-room"] = &Transcoder{
		roomID:    "test-room",
		streamKey: "test-key",
		startedAt: startTime,
	}

	resp, err = GetTranscoderHealth(nil, GetTranscoderHealthRequest{})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if resp.Active != 1 {
		t.Errorf("Expected 1 active transcoder, got %d", resp.Active)
	}

	if resp.MaxCapacity != MaxConcurrentTranscoders {
		t.Errorf("Expected max capacity %d, got %d", MaxConcurrentTranscoders, resp.MaxCapacity)
	}

	if !resp.Healthy {
		t.Error("Expected healthy=true with 1 transcoder")
	}

	if len(resp.Transcoders) != 1 {
		t.Errorf("Expected 1 transcoder in list, got %d", len(resp.Transcoders))
	}

	if len(resp.Transcoders) > 0 {
		tc := resp.Transcoders[0]
		if tc.RoomID != "test-room" {
			t.Errorf("Expected room ID 'test-room', got %s", tc.RoomID)
		}
		if tc.StreamKey != "test-key" {
			t.Errorf("Expected stream key 'test-key', got %s", tc.StreamKey)
		}
	}
}

func TestMaxConcurrentTranscodersEnvVar(t *testing.T) {
	// This test verifies the init() function reads the environment variable
	// Note: We can't easily test this without reloading the package
	// This is more of a documentation test

	// The init() function should have already run
	// We can verify the default value is set
	if MaxConcurrentTranscoders <= 0 || MaxConcurrentTranscoders > 100 {
		t.Errorf("MaxConcurrentTranscoders has invalid value: %d", MaxConcurrentTranscoders)
	}
}
