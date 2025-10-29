package backend

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// CameraIngestState represents a running camera ingest process
type CameraIngestState struct {
	RoomId    int
	RTSPURL   string
	StreamKey string
	StartTime time.Time
	Cmd       *exec.Cmd
	cancel    context.CancelFunc
}

// CameraManager manages FFmpeg processes for camera ingests
type CameraManager struct {
	mu        sync.RWMutex
	processes map[int]*CameraIngestState // roomId -> process
	ffmpegBin string
}

// CleanupOrphanedProcesses kills any FFmpeg processes streaming to our SRS instance
// This handles processes that were started by a previous server instance
func CleanupOrphanedProcesses() {
	// Find FFmpeg processes streaming to our local SRS instance
	cmd := exec.Command("sh", "-c", "ps aux | grep ffmpeg | grep 'rtmp://127.0.0.1:1935/live' | grep -v grep | awk '{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		// No processes found or error - either way, nothing to clean up
		return
	}

	pids := strings.TrimSpace(string(output))
	if pids == "" {
		return
	}

	// Kill each orphaned process
	for _, pidStr := range strings.Split(pids, "\n") {
		pidStr = strings.TrimSpace(pidStr)
		if pidStr == "" {
			continue
		}

		pid := 0
		fmt.Sscanf(pidStr, "%d", &pid)
		if pid > 0 {
			if process, err := os.FindProcess(pid); err == nil {
				process.Kill()
				LogInfo(LogCategorySystem, "Cleaned up orphaned FFmpeg process", map[string]interface{}{
					"pid": pid,
				})
			}
		}
	}
}

// NewCameraManager creates a new camera ingest manager
func NewCameraManager() *CameraManager {
	ffmpegBin := os.Getenv("FFMPEG_BIN")
	if ffmpegBin == "" {
		// Try to find ffmpeg in PATH
		if path, err := exec.LookPath("ffmpeg"); err == nil {
			ffmpegBin = path
		} else {
			ffmpegBin = "ffmpeg"
		}
	}

	// Clean up any orphaned FFmpeg processes from previous server instances
	CleanupOrphanedProcesses()

	return &CameraManager{
		processes: make(map[int]*CameraIngestState),
		ffmpegBin: ffmpegBin,
	}
}

// Start starts an FFmpeg ingest process for a camera
func (m *CameraManager) Start(ctx context.Context, roomId int, rtspURL, rtmpOut string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.processes[roomId]; exists {
		return errors.New("ingest already running for this room")
	}

	// Verify FFmpeg binary exists
	_, err := exec.LookPath(m.ffmpegBin)
	if err != nil {
		return fmt.Errorf("FFmpeg binary not found (check FFMPEG_BIN): %w", err)
	}

	// Extract stream key from RTMP URL
	streamKey := extractStreamKeyFromRTMP(rtmpOut)
	if streamKey == "" {
		return errors.New("failed to extract stream key from RTMP URL")
	}

	// Create context for this process
	procCtx, cancel := context.WithCancel(context.Background())

	// Build FFmpeg command with proper RTSP parameters
	// -rtsp_transport tcp: Use TCP for RTSP (prevents UDP packet loss)
	// -timeout: 20s socket I/O timeout (in microseconds)
	// -fflags nobuffer+discardcorrupt: Lower latency, handle corrupt packets
	// -flags low_delay: Reduce RTSP read latency
	// -i: Input URL
	// -c:v copy: Copy video codec without re-encoding (fast)
	// -c:a aac: Transcode audio to AAC
	// -b:a 128k: Audio bitrate 128kbps
	//      Note: Camera configured to use G.711/G.711A (PCMA/PCMU) instead of AAC
	//            G.711 has simple RTP packetization with no complex headers
	//            FFmpeg auto-detects G.711 and transcodes to AAC for RTMP output
	// -f flv: Output format FLV (RTMP container)
	cmd := exec.CommandContext(procCtx, m.ffmpegBin,
		"-rtsp_transport", "tcp",
		"-timeout", "20000000",
		"-fflags", "nobuffer+discardcorrupt",
		"-flags", "low_delay",
		"-i", rtspURL,
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "flv",
		rtmpOut,
	)

	// Create process state
	state := &CameraIngestState{
		RoomId:    roomId,
		RTSPURL:   rtspURL,
		StreamKey: streamKey,
		StartTime: time.Now(),
		Cmd:       cmd,
		cancel:    cancel,
	}

	// Set up stdout/stderr logging
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Store process
	m.processes[roomId] = state

	// Log process start
	LogInfo(LogCategorySystem, "Camera ingest started", map[string]interface{}{
		"roomId":    roomId,
		"pid":       cmd.Process.Pid,
		"rtspURL":   rtspURL,
		"streamKey": streamKey,
	})

	// Start goroutines to read stdout/stderr
	go m.logOutput(roomId, "stdout", stdout)
	go m.logOutput(roomId, "stderr", stderr)

	// Start goroutine to wait for process completion
	go m.waitForProcess(roomId, state)

	return nil
}

// Stop stops an FFmpeg ingest process for a camera
func (m *CameraManager) Stop(roomId int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.processes[roomId]
	if !exists {
		return errors.New("no ingest running for this room")
	}

	// Cancel the context
	state.cancel()

	// Try graceful shutdown first with SIGINT
	if state.Cmd.Process != nil {
		// Send SIGINT for graceful shutdown
		if err := state.Cmd.Process.Signal(syscall.SIGINT); err == nil {
			// Wait up to 5 seconds for graceful shutdown
			done := make(chan struct{})
			go func() {
				state.Cmd.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Graceful shutdown succeeded
			case <-time.After(5 * time.Second):
				// Graceful shutdown timed out, force kill
				state.Cmd.Process.Kill()
				LogInfo(LogCategorySystem, "Camera ingest force killed", map[string]interface{}{
					"roomId": roomId,
					"pid":    state.Cmd.Process.Pid,
				})
			}
		} else {
			// SIGINT failed, force kill
			state.Cmd.Process.Kill()
		}
	}

	// Remove from map
	delete(m.processes, roomId)

	duration := time.Since(state.StartTime)
	LogInfo(LogCategorySystem, "Camera ingest stopped", map[string]interface{}{
		"roomId":   roomId,
		"duration": duration.String(),
	})

	return nil
}

// IsRunning checks if an ingest process is running for a room
func (m *CameraManager) IsRunning(roomId int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.processes[roomId]
	return exists
}

// GetStatus returns the status of an ingest process
func (m *CameraManager) GetStatus(roomId int) (running bool, startTime time.Time, rtspURL string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.processes[roomId]
	if !exists {
		return false, time.Time{}, ""
	}

	return true, state.StartTime, state.RTSPURL
}

// logOutput reads from an io.Reader and logs each line
func (m *CameraManager) logOutput(roomId int, stream string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Log FFmpeg output at info level
		LogInfo(LogCategorySystem, "FFmpeg "+stream, map[string]interface{}{
			"roomId": roomId,
			"output": line,
		})
	}
	if err := scanner.Err(); err != nil {
		LogErrorSimple(LogCategorySystem, "Error reading FFmpeg "+stream, map[string]interface{}{
			"roomId": roomId,
			"stream": stream,
			"error":  err.Error(),
		})
	}
}

// waitForProcess waits for a process to complete and cleans up
func (m *CameraManager) waitForProcess(roomId int, state *CameraIngestState) {
	err := state.Cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if process is still in map (it might have been removed by Stop)
	if _, exists := m.processes[roomId]; !exists {
		return
	}

	// Remove from map
	delete(m.processes, roomId)

	// Log process exit
	duration := time.Since(state.StartTime)
	if err != nil {
		LogErrorSimple(LogCategorySystem, "Camera ingest exited with error", map[string]interface{}{
			"roomId":   roomId,
			"duration": duration.String(),
			"error":    err.Error(),
		})
	} else {
		LogInfo(LogCategorySystem, "Camera ingest exited normally", map[string]interface{}{
			"roomId":   roomId,
			"duration": duration.String(),
		})
	}
}

// extractStreamKeyFromRTMP extracts stream key from RTMP URL
// Example: rtmp://127.0.0.1:1935/live/{streamKey} -> {streamKey}
func extractStreamKeyFromRTMP(rtmpURL string) string {
	// Find the last slash
	lastSlash := -1
	for i := len(rtmpURL) - 1; i >= 0; i-- {
		if rtmpURL[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash == -1 || lastSlash == len(rtmpURL)-1 {
		return ""
	}

	return rtmpURL[lastSlash+1:]
}
