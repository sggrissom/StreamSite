package ingest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// IngestProcess represents a running FFmpeg ingest process
type IngestProcess struct {
	RoomId    int
	Cmd       *exec.Cmd
	StartTime time.Time
	RTSPURL   string
	RTMPOut   string
	cancel    context.CancelFunc
}

// IngestManager manages FFmpeg processes for camera ingests
type IngestManager struct {
	mu        sync.RWMutex
	processes map[int]*IngestProcess // roomId -> process
	ffmpegBin string
}

// NewIngestManager creates a new IngestManager
func NewIngestManager() *IngestManager {
	ffmpegBin := os.Getenv("FFMPEG_BIN")
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}

	return &IngestManager{
		processes: make(map[int]*IngestProcess),
		ffmpegBin: ffmpegBin,
	}
}

// Start starts an FFmpeg ingest process for a room
func (m *IngestManager) Start(ctx context.Context, roomId int, rtspURL, rtmpOut string) error {
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

	// Create context for this process
	procCtx, cancel := context.WithCancel(context.Background())

	// Build FFmpeg command
	// -rtsp_transport tcp: Use TCP for RTSP (more reliable than UDP)
	// -i: Input URL
	// -c:v copy: Copy video codec without re-encoding (fast)
	// -c:a aac: Encode audio to AAC (SRS requirement)
	// -f flv: Output format FLV (RTMP container)
	cmd := exec.CommandContext(procCtx, m.ffmpegBin,
		"-rtsp_transport", "tcp",
		"-i", rtspURL,
		"-c:v", "copy",
		"-c:a", "aac",
		"-f", "flv",
		rtmpOut,
	)

	// Create process struct
	process := &IngestProcess{
		RoomId:    roomId,
		Cmd:       cmd,
		StartTime: time.Now(),
		RTSPURL:   rtspURL,
		RTMPOut:   rtmpOut,
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
	m.processes[roomId] = process

	// Log process start
	log.Printf("[INGEST] FFmpeg ingest started for room %d (PID: %d, RTSP: %s, RTMP: %s)",
		roomId, cmd.Process.Pid, rtspURL, rtmpOut)

	// Start goroutines to read stdout/stderr
	go m.logOutput(roomId, "stdout", stdout)
	go m.logOutput(roomId, "stderr", stderr)

	// Start goroutine to wait for process completion
	go m.waitForProcess(roomId, process)

	return nil
}

// Stop stops an FFmpeg ingest process for a room
func (m *IngestManager) Stop(roomId int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, exists := m.processes[roomId]
	if !exists {
		return errors.New("no ingest running for this room")
	}

	// Cancel the context (sends SIGKILL via CommandContext)
	process.cancel()

	// Try graceful shutdown first with SIGINT
	if process.Cmd.Process != nil {
		// Send SIGINT for graceful shutdown
		if err := process.Cmd.Process.Signal(syscall.SIGINT); err == nil {
			// Wait up to 5 seconds for graceful shutdown
			done := make(chan struct{})
			go func() {
				process.Cmd.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Graceful shutdown succeeded
				log.Printf("[INGEST] FFmpeg ingest stopped gracefully for room %d (PID: %d)",
					roomId, process.Cmd.Process.Pid)
			case <-time.After(5 * time.Second):
				// Graceful shutdown timed out, force kill
				process.Cmd.Process.Kill()
				log.Printf("[INGEST] FFmpeg ingest force killed for room %d (PID: %d)",
					roomId, process.Cmd.Process.Pid)
			}
		} else {
			// SIGINT failed, force kill
			process.Cmd.Process.Kill()
		}
	}

	// Remove from map
	delete(m.processes, roomId)

	return nil
}

// IsRunning checks if an ingest process is running for a room
func (m *IngestManager) IsRunning(roomId int) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.processes[roomId]
	return exists
}

// GetStatus returns the status of an ingest process
func (m *IngestManager) GetStatus(roomId int) (running bool, startTime time.Time, rtspURL string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	process, exists := m.processes[roomId]
	if !exists {
		return false, time.Time{}, ""
	}

	return true, process.StartTime, process.RTSPURL
}

// logOutput reads from an io.Reader and logs each line
func (m *IngestManager) logOutput(roomId int, stream string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[INGEST] FFmpeg %s [room %d]: %s", stream, roomId, line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[INGEST] Error reading FFmpeg %s for room %d: %v", stream, roomId, err)
	}
}

// waitForProcess waits for a process to complete and cleans up
func (m *IngestManager) waitForProcess(roomId int, process *IngestProcess) {
	err := process.Cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if process is still in map (it might have been removed by Stop)
	if _, exists := m.processes[roomId]; !exists {
		return
	}

	// Remove from map
	delete(m.processes, roomId)

	// Log process exit
	duration := time.Since(process.StartTime)
	if err != nil {
		log.Printf("[INGEST] FFmpeg ingest exited with error for room %d (duration: %s): %v",
			roomId, duration, err)
	} else {
		log.Printf("[INGEST] FFmpeg ingest exited normally for room %d (duration: %s)",
			roomId, duration)
	}
}
