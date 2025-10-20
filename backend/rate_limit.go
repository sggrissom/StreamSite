package backend

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a thread-safe in-memory rate limiter using sliding window algorithm.
// It tracks attempts by key (e.g., IP address, user ID) and enforces configurable limits.
// Also implements exponential backoff for repeated violations.
type RateLimiter struct {
	mu         sync.RWMutex
	entries    map[string]*rateLimitEntry
	violations map[string]*violationTracker
}

// rateLimitEntry tracks attempts for a specific key within a time window
type rateLimitEntry struct {
	attempts []time.Time
	limit    int
	window   time.Duration
}

// violationTracker tracks consecutive rate limit violations for exponential backoff
type violationTracker struct {
	count       int       // Number of consecutive violations
	lockedUntil time.Time // When the current lockout expires
}

// RateLimitInfo contains structured information about rate limit status
type RateLimitInfo struct {
	Limited           bool
	RetryAfterSeconds int
	ViolationCount    int
	Message           string
}

// NewRateLimiter creates a new rate limiter and starts a background cleanup goroutine
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		entries:    make(map[string]*rateLimitEntry),
		violations: make(map[string]*violationTracker),
	}

	// Start background cleanup every 5 minutes
	go rl.cleanupLoop()

	return rl
}

// CheckLimit checks if an action is allowed for the given key.
// It returns an error if the rate limit is exceeded.
// Parameters:
//   - limitType: Type of limit (e.g., "code_validation", "code_generation")
//   - identifier: Unique identifier (e.g., IP address, user ID)
//   - limit: Maximum number of attempts allowed
//   - window: Time window for the limit (e.g., 1 minute, 1 hour)
func (rl *RateLimiter) CheckLimit(limitType, identifier string, limit int, window time.Duration) error {
	key := fmt.Sprintf("%s:%s", limitType, identifier)
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.entries[key]
	if !exists {
		// First attempt - create new entry
		rl.entries[key] = &rateLimitEntry{
			attempts: []time.Time{now},
			limit:    limit,
			window:   window,
		}
		return nil
	}

	// Remove attempts outside the time window
	cutoff := now.Add(-window)
	validAttempts := make([]time.Time, 0, len(entry.attempts))
	for _, t := range entry.attempts {
		if t.After(cutoff) {
			validAttempts = append(validAttempts, t)
		}
	}

	// Check if limit is exceeded
	if len(validAttempts) >= limit {
		// Calculate when the oldest attempt will expire
		oldestAttempt := validAttempts[0]
		resetTime := oldestAttempt.Add(window)
		waitSeconds := int(resetTime.Sub(now).Seconds()) + 1

		return fmt.Errorf("rate limit exceeded: %d attempts in %v (max %d), try again in %d seconds",
			len(validAttempts), window, limit, waitSeconds)
	}

	// Add new attempt
	validAttempts = append(validAttempts, now)
	entry.attempts = validAttempts

	return nil
}

// getBackoffDuration returns the lockout duration based on violation count
// Escalates: 1min → 5min → 15min → 1hour
func getBackoffDuration(violationCount int) time.Duration {
	switch violationCount {
	case 0:
		return 0
	case 1:
		return 1 * time.Minute
	case 2:
		return 5 * time.Minute
	case 3:
		return 15 * time.Minute
	default:
		return 1 * time.Hour // 4+ violations
	}
}

// CheckLimitWithBackoff checks if an action is allowed and applies exponential backoff on violations.
// Returns structured rate limit information instead of just an error.
func (rl *RateLimiter) CheckLimitWithBackoff(limitType, identifier string, limit int, window time.Duration) *RateLimitInfo {
	key := fmt.Sprintf("%s:%s", limitType, identifier)
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if currently locked out due to previous violations
	violation, hasViolation := rl.violations[key]
	if hasViolation && now.Before(violation.lockedUntil) {
		waitSeconds := int(violation.lockedUntil.Sub(now).Seconds()) + 1
		return &RateLimitInfo{
			Limited:           true,
			RetryAfterSeconds: waitSeconds,
			ViolationCount:    violation.count,
			Message: fmt.Sprintf("Too many failed attempts. Locked out for %d seconds (violation #%d)",
				waitSeconds, violation.count),
		}
	}

	// Check normal rate limit
	entry, exists := rl.entries[key]
	if !exists {
		// First attempt - create new entry
		rl.entries[key] = &rateLimitEntry{
			attempts: []time.Time{now},
			limit:    limit,
			window:   window,
		}
		return &RateLimitInfo{Limited: false}
	}

	// Remove attempts outside the time window
	cutoff := now.Add(-window)
	validAttempts := make([]time.Time, 0, len(entry.attempts))
	for _, t := range entry.attempts {
		if t.After(cutoff) {
			validAttempts = append(validAttempts, t)
		}
	}

	// Check if limit is exceeded
	if len(validAttempts) >= limit {
		// Rate limit exceeded - apply exponential backoff
		if !hasViolation {
			violation = &violationTracker{count: 0}
			rl.violations[key] = violation
		}

		// Increment violation count and apply backoff
		violation.count++
		backoffDuration := getBackoffDuration(violation.count)
		violation.lockedUntil = now.Add(backoffDuration)

		waitSeconds := int(backoffDuration.Seconds())

		LogWarn(LogCategorySystem, "Rate limit violation with backoff", map[string]interface{}{
			"key":            key,
			"violationCount": violation.count,
			"lockoutSeconds": waitSeconds,
		})

		return &RateLimitInfo{
			Limited:           true,
			RetryAfterSeconds: waitSeconds,
			ViolationCount:    violation.count,
			Message: fmt.Sprintf("Rate limit exceeded. Try again in %d seconds (attempt %d of escalating lockout)",
				waitSeconds, violation.count),
		}
	}

	// Add new attempt
	validAttempts = append(validAttempts, now)
	entry.attempts = validAttempts

	return &RateLimitInfo{Limited: false}
}

// ResetViolations clears violation tracking for a key (call after successful validation)
func (rl *RateLimiter) ResetViolations(limitType, identifier string) {
	key := fmt.Sprintf("%s:%s", limitType, identifier)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.violations, key)
}

// CheckCodeValidation checks rate limit for access code validation attempts
// Limit: 5 attempts per IP per minute, with exponential backoff on violations
func (rl *RateLimiter) CheckCodeValidation(ipAddress string) error {
	info := rl.CheckLimitWithBackoff("code_validation", ipAddress, 5, 1*time.Minute)
	if info.Limited {
		return fmt.Errorf("%s", info.Message)
	}
	return nil
}

// CheckCodeValidationWithInfo returns structured rate limit information
func (rl *RateLimiter) CheckCodeValidationWithInfo(ipAddress string) *RateLimitInfo {
	return rl.CheckLimitWithBackoff("code_validation", ipAddress, 5, 1*time.Minute)
}

// CheckCodeGeneration checks rate limit for access code generation
// Limit: 10 generations per user per hour
func (rl *RateLimiter) CheckCodeGeneration(userID int) error {
	return rl.CheckLimit("code_generation", fmt.Sprintf("%d", userID), 10, 1*time.Hour)
}

// cleanupLoop runs periodically to remove expired entries and free memory
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes entries that have no valid attempts within their time window
// and expired violation trackers
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	keysToDelete := make([]string, 0)
	violationsToDelete := make([]string, 0)

	// Clean up old rate limit entries
	for key, entry := range rl.entries {
		cutoff := now.Add(-entry.window)
		hasValidAttempts := false

		for _, t := range entry.attempts {
			if t.After(cutoff) {
				hasValidAttempts = true
				break
			}
		}

		if !hasValidAttempts {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(rl.entries, key)
	}

	// Clean up expired violations (lockout expired)
	for key, violation := range rl.violations {
		if now.After(violation.lockedUntil) {
			violationsToDelete = append(violationsToDelete, key)
		}
	}

	for _, key := range violationsToDelete {
		delete(rl.violations, key)
	}

	if len(keysToDelete) > 0 || len(violationsToDelete) > 0 {
		LogDebug(LogCategorySystem, "Rate limiter cleanup", map[string]interface{}{
			"removedEntries":    len(keysToDelete),
			"removedViolations": len(violationsToDelete),
		})
	}
}

// GetStats returns statistics about current rate limiter state (for testing/monitoring)
func (rl *RateLimiter) GetStats() map[string]int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := map[string]int{
		"total_entries": len(rl.entries),
	}

	now := time.Now()
	activeEntries := 0

	for _, entry := range rl.entries {
		cutoff := now.Add(-entry.window)
		for _, t := range entry.attempts {
			if t.After(cutoff) {
				activeEntries++
				break
			}
		}
	}

	stats["active_entries"] = activeEntries

	return stats
}

// Reset clears all rate limit entries and violations (for testing only)
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.entries = make(map[string]*rateLimitEntry)
	rl.violations = make(map[string]*violationTracker)
}

// Global rate limiter instance
var globalRateLimiter = NewRateLimiter()
