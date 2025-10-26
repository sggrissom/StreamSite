package ingest

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewIngestManager(t *testing.T) {
	// Test with default FFmpeg binary
	os.Unsetenv("FFMPEG_BIN")
	manager := NewIngestManager()

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.ffmpegBin != "ffmpeg" {
		t.Errorf("Expected default ffmpeg binary, got %s", manager.ffmpegBin)
	}

	if manager.processes == nil {
		t.Error("Expected initialized processes map")
	}

	// Test with custom FFmpeg binary
	os.Setenv("FFMPEG_BIN", "/usr/bin/ffmpeg")
	defer os.Unsetenv("FFMPEG_BIN")

	manager2 := NewIngestManager()
	if manager2.ffmpegBin != "/usr/bin/ffmpeg" {
		t.Errorf("Expected custom ffmpeg binary, got %s", manager2.ffmpegBin)
	}
}

func TestIsRunning(t *testing.T) {
	manager := NewIngestManager()

	// Test when no process is running
	if manager.IsRunning(1) {
		t.Error("Expected IsRunning to return false for non-existent process")
	}

	// We can't easily test with a real process without FFmpeg,
	// so we'll just verify the basic logic works
}

func TestGetStatus(t *testing.T) {
	manager := NewIngestManager()

	// Test when no process is running
	running, startTime, rtspURL := manager.GetStatus(1)

	if running {
		t.Error("Expected running to be false")
	}

	if !startTime.IsZero() {
		t.Error("Expected zero time for non-existent process")
	}

	if rtspURL != "" {
		t.Error("Expected empty RTSP URL for non-existent process")
	}
}

func TestStartWithInvalidFFmpeg(t *testing.T) {
	// Set invalid FFmpeg binary
	os.Setenv("FFMPEG_BIN", "/nonexistent/ffmpeg")
	defer os.Unsetenv("FFMPEG_BIN")

	manager := NewIngestManager()
	ctx := context.Background()

	err := manager.Start(ctx, 1, "rtsp://example.com/stream", "rtmp://localhost/live/test")

	if err == nil {
		t.Error("Expected error when FFmpeg binary doesn't exist")
	}

	if manager.IsRunning(1) {
		t.Error("Expected no process to be running after failed start")
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	manager := NewIngestManager()
	ctx := context.Background()

	// Manually add a fake process to simulate already running
	manager.mu.Lock()
	manager.processes[1] = &IngestProcess{
		RoomId:    1,
		StartTime: time.Now(),
	}
	manager.mu.Unlock()

	err := manager.Start(ctx, 1, "rtsp://example.com/stream", "rtmp://localhost/live/test")

	if err == nil {
		t.Error("Expected error when trying to start already running process")
	}

	if err.Error() != "ingest already running for this room" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

func TestStopNonExistent(t *testing.T) {
	manager := NewIngestManager()

	err := manager.Stop(1)

	if err == nil {
		t.Error("Expected error when stopping non-existent process")
	}

	if err.Error() != "no ingest running for this room" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Test that concurrent access to manager doesn't cause race conditions
	manager := NewIngestManager()

	done := make(chan bool)

	// Multiple goroutines checking status
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				manager.IsRunning(1)
				manager.GetStatus(1)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Note: Testing actual FFmpeg process start/stop requires FFmpeg to be installed
// and a valid RTSP source. These tests focus on the manager logic itself.
// Integration tests with actual FFmpeg should be done separately.
