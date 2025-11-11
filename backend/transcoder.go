package backend

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.hasen.dev/vbeam"
)

// Global transcoder manager instance
var transcoderManager *TranscoderManager

// Maximum number of concurrent transcoding sessions (configurable via MAX_CONCURRENT_TRANSCODERS env var)
var MaxConcurrentTranscoders = 10

func init() {
	// Allow override via environment variable
	if envVal := os.Getenv("MAX_CONCURRENT_TRANSCODERS"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil && val > 0 && val <= 100 {
			MaxConcurrentTranscoders = val
			log.Printf("[Transcoder] Using MAX_CONCURRENT_TRANSCODERS=%d from environment", val)
		} else {
			log.Printf("[Transcoder] Warning: Invalid MAX_CONCURRENT_TRANSCODERS value '%s', using default %d",
				envVal, MaxConcurrentTranscoders)
		}
	}
}

// InitTranscoder initializes the global transcoder manager
func InitTranscoder(cfg TranscoderConfig) {
	transcoderManager = NewTranscoderManager(cfg)
	log.Printf("[Transcoder] Initialized with HLS dir: %s, RTMP base: %s, max transcoders: %d",
		cfg.HLSBaseDir, cfg.SRSRTMPBase, MaxConcurrentTranscoders)

	// Clean up any orphaned resources from previous runs
	transcoderManager.CleanupOrphanedProcesses()
	transcoderManager.CleanupOrphanedDirectories()
}

// TranscoderConfig holds configuration for the transcoding system
type TranscoderConfig struct {
	// Base RTMP URL for SRS (e.g., "rtmp://localhost:1935/live")
	SRSRTMPBase string
	// Root directory for HLS output (e.g., "./hls" or "/var/www/hls")
	HLSBaseDir string
}

// Transcoder manages a single FFmpeg transcoding process for ABR HLS
type Transcoder struct {
	roomID    string
	streamKey string
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	outDir    string
	inputRTMP string
	startedAt time.Time
}

// validateStreamKey checks if a stream key is safe for use in shell commands
func validateStreamKey(key string) error {
	if key == "" {
		return fmt.Errorf("stream key cannot be empty")
	}

	// Stream keys should be alphanumeric with hyphens/underscores only
	// Reject any shell metacharacters or path separators
	dangerousChars := []string{
		";", "&", "|", "`", "$", "(", ")", "{", "}", "[", "]",
		"<", ">", "'", "\"", "\\", "/", " ", "\n", "\r", "\t",
		"*", "?", "!", "#", "~", "^",
	}

	for _, char := range dangerousChars {
		if strings.Contains(key, char) {
			return fmt.Errorf("stream key contains invalid character: %s", char)
		}
	}

	return nil
}

// NewTranscoder creates a new transcoder instance for a room
func NewTranscoder(roomID, streamKey string, cfg TranscoderConfig) *Transcoder {
	return &Transcoder{
		roomID:    roomID,
		streamKey: streamKey,
		outDir:    filepath.Join(cfg.HLSBaseDir, roomID),
		inputRTMP: fmt.Sprintf("%s/%s", cfg.SRSRTMPBase, streamKey),
	}
}

// StartWithRetry attempts to start FFmpeg with exponential backoff retries
// This handles race conditions where FFmpeg starts before SRS has buffered data
func (t *Transcoder) StartWithRetry(maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("[Transcoder] Retry attempt %d/%d for room=%s after %v",
				attempt, maxRetries, t.roomID, delay)
			time.Sleep(delay)
		}

		err := t.Start()
		if err == nil {
			if attempt > 0 {
				log.Printf("[Transcoder] Successfully started on retry %d for room=%s",
					attempt, t.roomID)
			}
			return nil
		}

		lastErr = err
		log.Printf("[Transcoder] Start attempt %d failed for room=%s: %v",
			attempt+1, t.roomID, err)

		// If it's not a connection error, don't retry
		if !strings.Contains(err.Error(), "Connection refused") &&
			!strings.Contains(err.Error(), "No such file") {
			break
		}
	}

	return fmt.Errorf("failed to start after %d attempts: %w", maxRetries+1, lastErr)
}

// Start begins the FFmpeg transcoding process
func (t *Transcoder) Start() error {
	// Create output directories
	if err := os.MkdirAll(t.outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create HLS directory: %w", err)
	}

	// Build FFmpeg arguments for ABR HLS with multiple quality variants
	args := []string{
		"-hide_banner", "-loglevel", "info",
		"-i", t.inputRTMP,

		// Map input streams: video 3x, audio 3x (once for each variant)
		"-map", "0:v:0", "-map", "0:v:0", "-map", "0:v:0", "-map", "0:a:0", "-map", "0:a:0", "-map", "0:a:0",

		// Base encoders
		"-c:v", "libx264", "-preset", "veryfast",
		"-c:a", "aac", "-ar", "48000", "-ac", "2", "-b:a", "128k",

		// GOP alignment for smooth quality switching (60 frames = 2s at 30fps)
		"-x264-params", "keyint=60:min-keyint=60:scenecut=0:nal-hrd=cbr:force-cfr=1",

		// Variant 0: 1080p @ 5 Mbps
		"-filter:v:0", "scale=w=1920:h=1080:flags=bicubic",
		"-b:v:0", "5000k", "-maxrate:v:0", "5500k", "-bufsize:v:0", "10000k",

		// Variant 1: 720p @ 2.5 Mbps
		"-filter:v:1", "scale=w=1280:h=720:flags=bicubic",
		"-b:v:1", "2500k", "-maxrate:v:1", "2700k", "-bufsize:v:1", "5000k",

		// Variant 2: 480p @ 1.2 Mbps
		"-filter:v:2", "scale=w=854:h=480:flags=bicubic",
		"-b:v:2", "1200k", "-maxrate:v:2", "1300k", "-bufsize:v:2", "2400k",

		// Map variants: v:0 with a:0 (1080p), v:1 with a:1 (720p), v:2 with a:2 (480p)
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2",

		// HLS settings
		"-hls_time", "2",
		"-hls_list_size", "5",
		"-hls_flags", "independent_segments+delete_segments+program_date_time",

		// %v is the variant index (0 for 1080p, 1 for 720p, 2 for 480p)
		"-hls_segment_filename", filepath.Join(t.outDir, "%v", "seg_%06d.ts"),
		"-master_pl_name", "master.m3u8",

		// Output per-variant playlists
		"-f", "hls", filepath.Join(t.outDir, "%v", "stream.m3u8"),
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	// Create command
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// Log FFmpeg output to stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	t.cmd = cmd
	t.startedAt = time.Now()

	// Start process
	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	log.Printf("[Transcoder] Started for room=%s key=%s pid=%d input=%s output=%s",
		t.roomID, t.streamKey, cmd.Process.Pid, t.inputRTMP, t.outDir)

	// Monitor process in background
	go func() {
		err := cmd.Wait()
		if err != nil && ctx.Err() == nil {
			// Process died unexpectedly (not from cancellation)
			duration := time.Since(t.startedAt)
			log.Printf("[Transcoder] FFmpeg died unexpectedly for room=%s after %v: %v",
				t.roomID, duration, err)

			// Notify viewers via SSE
			if roomID, parseErr := strconv.Atoi(t.roomID); parseErr == nil {
				errorMsg := fmt.Sprintf("Transcoding failed after %v", duration.Round(time.Second))
				sseManager.BroadcastTranscoderError(roomID, errorMsg)
			}
		}
	}()

	return nil
}

// Stop terminates the FFmpeg process gracefully
func (t *Transcoder) Stop() {
	if t == nil || t.cancel == nil {
		return
	}

	duration := time.Since(t.startedAt)
	log.Printf("[Transcoder] Stopping for room=%s (ran for %v)", t.roomID, duration)

	// Cancel context (sends SIGKILL to FFmpeg)
	t.cancel()

	// Give FFmpeg time to flush playlists
	time.Sleep(500 * time.Millisecond)

	// Clean up HLS directory to free space and prevent stale files
	if err := os.RemoveAll(t.outDir); err != nil {
		log.Printf("[Transcoder] Warning: Failed to clean HLS directory for room %s: %v", t.roomID, err)
	} else {
		log.Printf("[Transcoder] Cleaned HLS directory for room %s", t.roomID)
	}
}

// IsRunning returns true if the FFmpeg process is still active
func (t *Transcoder) IsRunning() bool {
	if t == nil || t.cmd == nil || t.cmd.Process == nil {
		return false
	}
	// ProcessState is nil while process is running
	return t.cmd.ProcessState == nil
}

// TranscoderManager manages multiple transcoder instances
type TranscoderManager struct {
	transcoders map[string]*Transcoder
	config      TranscoderConfig
	mu          sync.RWMutex
}

// NewTranscoderManager creates a new transcoder manager
func NewTranscoderManager(cfg TranscoderConfig) *TranscoderManager {
	// Ensure HLS base directory exists
	if err := os.MkdirAll(cfg.HLSBaseDir, 0o755); err != nil {
		log.Printf("[Transcoder] Warning: Failed to create HLS base directory %s: %v",
			cfg.HLSBaseDir, err)
	}

	return &TranscoderManager{
		transcoders: make(map[string]*Transcoder),
		config:      cfg,
	}
}

// Start creates and starts a new transcoder for a room
func (m *TranscoderManager) Start(roomID, streamKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check resource limits
	if len(m.transcoders) >= MaxConcurrentTranscoders {
		return fmt.Errorf("transcoder limit reached: %d active (max %d)",
			len(m.transcoders), MaxConcurrentTranscoders)
	}

	// Validate room ID is safe for filesystem - prevent path traversal
	roomID = filepath.Clean(roomID)
	if roomID == "." || roomID == ".." || strings.Contains(roomID, "..") ||
		filepath.IsAbs(roomID) || strings.ContainsAny(roomID, "/\\") {
		return fmt.Errorf("invalid room ID: %s", roomID)
	}

	// Validate stream key is safe for shell commands
	if err := validateStreamKey(streamKey); err != nil {
		return fmt.Errorf("invalid stream key: %w", err)
	}

	// Don't start duplicate transcoder
	if _, exists := m.transcoders[roomID]; exists {
		log.Printf("[Transcoder] Transcoder already running for room %s", roomID)
		return nil // Not an error, just ignore
	}

	// Clean up old HLS files from previous stream sessions
	// This prevents stale playlists from causing 404 errors
	hlsDir := filepath.Join(m.config.HLSBaseDir, roomID)
	if err := os.RemoveAll(hlsDir); err != nil && !os.IsNotExist(err) {
		log.Printf("[Transcoder] Warning: Failed to clean old HLS directory for room %s: %v", roomID, err)
		// Continue anyway - not fatal
	}

	// Create and start transcoder with retry logic
	tc := NewTranscoder(roomID, streamKey, m.config)
	if err := tc.StartWithRetry(3); err != nil {
		return fmt.Errorf("failed to start transcoder for room %s: %w", roomID, err)
	}

	m.transcoders[roomID] = tc
	log.Printf("[Transcoder] Active transcoders: %d", len(m.transcoders))

	return nil
}

// Stop terminates the transcoder for a room
func (m *TranscoderManager) Stop(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tc, exists := m.transcoders[roomID]; exists {
		tc.Stop()
		delete(m.transcoders, roomID)
		log.Printf("[Transcoder] Active transcoders: %d", len(m.transcoders))
	}
}

// GetActiveCount returns the number of active transcoders
func (m *TranscoderManager) GetActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.transcoders)
}

// IsRunning returns true if a transcoder is active for the given room
func (m *TranscoderManager) IsRunning(roomID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if tc, exists := m.transcoders[roomID]; exists {
		return tc.IsRunning()
	}
	return false
}

// CleanupOrphanedProcesses kills any leftover FFmpeg processes from previous runs
func (m *TranscoderManager) CleanupOrphanedProcesses() {
	// Use pkill to find and kill FFmpeg processes that match our HLS output pattern
	// This is safer than killing all ffmpeg processes
	pattern := fmt.Sprintf("ffmpeg.*%s", m.config.HLSBaseDir)
	cmd := exec.Command("pkill", "-f", pattern)

	if err := cmd.Run(); err != nil {
		// Exit code 1 means no processes found, which is fine
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			log.Printf("[Transcoder] No orphaned FFmpeg processes found")
			return
		}
		log.Printf("[Transcoder] Warning: Failed to cleanup orphaned FFmpeg processes: %v", err)
		return
	}

	log.Printf("[Transcoder] Cleaned up orphaned FFmpeg processes")
}

// CleanupOrphanedDirectories removes leftover HLS directories from previous runs
func (m *TranscoderManager) CleanupOrphanedDirectories() {
	entries, err := os.ReadDir(m.config.HLSBaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return // Directory doesn't exist yet, nothing to clean
		}
		log.Printf("[Transcoder] Warning: Failed to read HLS directory for cleanup: %v", err)
		return
	}

	cleaned := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := filepath.Join(m.config.HLSBaseDir, entry.Name())
		if err := os.RemoveAll(dirPath); err != nil {
			log.Printf("[Transcoder] Warning: Failed to remove orphaned directory %s: %v",
				dirPath, err)
		} else {
			cleaned++
		}
	}

	if cleaned > 0 {
		log.Printf("[Transcoder] Cleaned up %d orphaned HLS directories from previous run", cleaned)
	}
}

// TranscoderStatus represents the status of a single transcoder
type TranscoderStatus struct {
	RoomID    string    `json:"roomId"`
	StreamKey string    `json:"streamKey"`
	Running   bool      `json:"running"`
	StartedAt time.Time `json:"startedAt"`
	Duration  string    `json:"duration"`
}

// TranscoderHealthResponse contains health check information
type TranscoderHealthResponse struct {
	Active      int                `json:"active"`
	MaxCapacity int                `json:"maxCapacity"`
	Healthy     bool               `json:"healthy"`
	Transcoders []TranscoderStatus `json:"transcoders"`
}

// GetHealth returns health status of all transcoders
func (m *TranscoderManager) GetHealth() TranscoderHealthResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]TranscoderStatus, 0, len(m.transcoders))
	for roomID, tc := range m.transcoders {
		statuses = append(statuses, TranscoderStatus{
			RoomID:    roomID,
			StreamKey: tc.streamKey,
			Running:   tc.IsRunning(),
			StartedAt: tc.startedAt,
			Duration:  time.Since(tc.startedAt).Round(time.Second).String(),
		})
	}

	return TranscoderHealthResponse{
		Active:      len(m.transcoders),
		MaxCapacity: MaxConcurrentTranscoders,
		Healthy:     len(m.transcoders) < MaxConcurrentTranscoders,
		Transcoders: statuses,
	}
}

// GetTranscoderHealth API procedure
type GetTranscoderHealthRequest struct{}

func GetTranscoderHealth(ctx *vbeam.Context, req GetTranscoderHealthRequest) (TranscoderHealthResponse, error) {
	if transcoderManager == nil {
		return TranscoderHealthResponse{
			Active:      0,
			MaxCapacity: MaxConcurrentTranscoders,
			Healthy:     true,
			Transcoders: []TranscoderStatus{},
		}, nil
	}

	return transcoderManager.GetHealth(), nil
}
