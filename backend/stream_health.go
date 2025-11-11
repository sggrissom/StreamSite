package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CheckHlsAvailability verifies that HLS segments are available for the given room.
// It checks the filesystem where FFmpeg transcoder writes HLS files.
// Returns true if HLS is ready for playback, false otherwise.
func CheckHlsAvailability(hlsBaseDir string, roomId int64) bool {
	// Construct the filesystem path for the master playlist
	// FFmpeg transcoder writes to: {hlsBaseDir}/{roomId}/master.m3u8
	roomIdStr := fmt.Sprintf("%d", roomId)
	masterPath := filepath.Join(hlsBaseDir, roomIdStr, "master.m3u8")

	// Check if master.m3u8 exists
	info, err := os.Stat(masterPath)
	if err != nil {
		// File doesn't exist yet or can't be accessed
		return false
	}

	// Verify it's a regular file and not empty
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return false
	}

	// Read and validate the playlist content
	content, err := os.ReadFile(masterPath)
	if err != nil {
		return false
	}

	contentStr := string(content)

	// Basic validation: check for HLS playlist header
	if !strings.Contains(contentStr, "#EXTM3U") {
		// Not a valid m3u8 file
		return false
	}

	// Check for variant playlists (adaptive bitrate streams)
	// Format: #EXT-X-STREAM-INF:BANDWIDTH=...
	if !strings.Contains(contentStr, "#EXT-X-STREAM-INF") {
		// Master playlist exists but has no variants yet
		return false
	}

	// Verify that multiple variant directories exist with segments
	// FFmpeg creates directories: 0/ (1080p), 1/ (720p), 2/ (480p)
	// We require at least 2 variants to be ready to avoid race conditions
	// where clients try to load variants that don't exist yet
	readyVariants := 0
	for i := 0; i < 3; i++ {
		variantDir := filepath.Join(hlsBaseDir, roomIdStr, fmt.Sprintf("%d", i))
		if dirInfo, err := os.Stat(variantDir); err == nil && dirInfo.IsDir() {
			// Check if variant has a stream.m3u8 file
			variantPlaylist := filepath.Join(variantDir, "stream.m3u8")
			if info, err := os.Stat(variantPlaylist); err == nil {
				// Also verify the playlist is not empty (file is fully written)
				if info.Size() > 0 {
					readyVariants++
				}
			}
		}
	}

	// Require at least 2 out of 3 variants to be ready
	// This ensures clients won't get 404s when HLS.js tries to load a variant
	return readyVariants >= 2
}

// PollForHlsAvailability polls for HLS availability up to maxAttempts times
// with a delay between attempts. Returns true if HLS becomes available, false on timeout.
func PollForHlsAvailability(hlsBaseDir string, roomId int64, maxAttempts int, delayBetweenAttempts time.Duration) bool {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if CheckHlsAvailability(hlsBaseDir, roomId) {
			return true
		}
		if attempt < maxAttempts-1 {
			time.Sleep(delayBetweenAttempts)
		}
	}
	return false
}
