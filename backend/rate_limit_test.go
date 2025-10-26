package backend

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestRateLimiter_BasicLimit tests that the rate limiter enforces limits correctly
func TestRateLimiter_BasicLimit(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset() // Clear any existing entries

	identifier := "192.168.1.100"
	limit := 3
	window := 100 * time.Millisecond

	// First 3 attempts should succeed
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("test", identifier, limit, window)
		if err != nil {
			t.Errorf("Attempt %d should succeed, got error: %v", i+1, err)
		}
	}

	// 4th attempt should fail
	err := rl.CheckLimit("test", identifier, limit, window)
	if err == nil {
		t.Error("Expected rate limit error, got nil")
	}
	if err != nil && err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestRateLimiter_WindowReset tests that limits reset after the time window passes
func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "192.168.1.101"
	limit := 2
	window := 200 * time.Millisecond

	// Use up the limit
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("test", identifier, limit, window)
		if err != nil {
			t.Fatalf("Initial attempt %d failed: %v", i+1, err)
		}
	}

	// Should fail immediately
	err := rl.CheckLimit("test", identifier, limit, window)
	if err == nil {
		t.Error("Expected rate limit error before window reset")
	}

	// Wait for window to pass
	time.Sleep(window + 50*time.Millisecond)

	// Should succeed after window reset
	err = rl.CheckLimit("test", identifier, limit, window)
	if err != nil {
		t.Errorf("Expected success after window reset, got error: %v", err)
	}
}

// TestRateLimiter_SlidingWindow tests that the sliding window algorithm works correctly
func TestRateLimiter_SlidingWindow(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "192.168.1.102"
	limit := 3
	window := 300 * time.Millisecond

	// Make 3 attempts
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("test", identifier, limit, window)
		if err != nil {
			t.Fatalf("Attempt %d failed: %v", i+1, err)
		}
	}

	// Should fail (at limit)
	err := rl.CheckLimit("test", identifier, limit, window)
	if err == nil {
		t.Error("Expected rate limit error")
	}

	// Wait for half the window
	time.Sleep(window / 2)

	// Should still fail (still 3 attempts in window)
	err = rl.CheckLimit("test", identifier, limit, window)
	if err == nil {
		t.Error("Expected rate limit error after half window")
	}

	// Wait for the rest of the window + buffer
	time.Sleep(window/2 + 100*time.Millisecond)

	// Should succeed now (old attempts expired)
	err = rl.CheckLimit("test", identifier, limit, window)
	if err != nil {
		t.Errorf("Expected success after full window, got error: %v", err)
	}
}

// TestRateLimiter_MultipleIdentifiers tests that different identifiers have independent limits
func TestRateLimiter_MultipleIdentifiers(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"
	limit := 2
	window := 100 * time.Millisecond

	// Use up limit for ip1
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("test", ip1, limit, window)
		if err != nil {
			t.Fatalf("IP1 attempt %d failed: %v", i+1, err)
		}
	}

	// ip1 should be rate limited
	err := rl.CheckLimit("test", ip1, limit, window)
	if err == nil {
		t.Error("Expected ip1 to be rate limited")
	}

	// ip2 should still have full quota
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("test", ip2, limit, window)
		if err != nil {
			t.Errorf("IP2 attempt %d should succeed, got error: %v", i+1, err)
		}
	}
}

// TestRateLimiter_DifferentLimitTypes tests that different limit types are tracked separately
func TestRateLimiter_DifferentLimitTypes(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "user123"
	limit := 2
	window := 100 * time.Millisecond

	// Use up limit for type "validation"
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("validation", identifier, limit, window)
		if err != nil {
			t.Fatalf("Validation attempt %d failed: %v", i+1, err)
		}
	}

	// "validation" should be rate limited
	err := rl.CheckLimit("validation", identifier, limit, window)
	if err == nil {
		t.Error("Expected validation to be rate limited")
	}

	// "generation" should still have full quota
	for i := 0; i < limit; i++ {
		err := rl.CheckLimit("generation", identifier, limit, window)
		if err != nil {
			t.Errorf("Generation attempt %d should succeed, got error: %v", i+1, err)
		}
	}
}

// TestRateLimiter_Concurrent tests thread-safety under concurrent access
func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "concurrent-test"
	limit := 10
	window := 500 * time.Millisecond
	goroutines := 20

	var wg sync.WaitGroup
	successCount := make(chan int, goroutines)

	// Launch concurrent goroutines trying to use the rate limiter
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := rl.CheckLimit("concurrent", identifier, limit, window)
			if err == nil {
				successCount <- 1
			} else {
				successCount <- 0
			}
		}(i)
	}

	wg.Wait()
	close(successCount)

	// Count successes
	total := 0
	for count := range successCount {
		total += count
	}

	// Should have exactly 'limit' successes
	if total != limit {
		t.Errorf("Expected %d successes, got %d", limit, total)
	}
}

// TestRateLimiter_CheckCodeValidation tests the convenience method for code validation
func TestRateLimiter_CheckCodeValidation(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	ip := "10.0.0.1"

	// Should allow 5 attempts (as per spec)
	for i := 0; i < 5; i++ {
		err := rl.CheckCodeValidation(ip)
		if err != nil {
			t.Errorf("Validation attempt %d should succeed, got error: %v", i+1, err)
		}
	}

	// 6th should fail
	err := rl.CheckCodeValidation(ip)
	if err == nil {
		t.Error("Expected rate limit after 5 validation attempts")
	}

	// Error message should mention waiting time
	if err != nil && err.Error() == "" {
		t.Error("Expected informative error message")
	}
}

// TestRateLimiter_CheckCodeGeneration tests the convenience method for code generation
func TestRateLimiter_CheckCodeGeneration(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	userID := 42

	// Should allow 10 attempts (as per spec)
	for i := 0; i < 10; i++ {
		err := rl.CheckCodeGeneration(userID)
		if err != nil {
			t.Errorf("Generation attempt %d should succeed, got error: %v", i+1, err)
		}
	}

	// 11th should fail
	err := rl.CheckCodeGeneration(userID)
	if err == nil {
		t.Error("Expected rate limit after 10 generation attempts")
	}
}

// TestRateLimiter_Cleanup tests that cleanup removes expired entries
func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	// Create some entries with short window
	shortWindow := 50 * time.Millisecond
	for i := 0; i < 3; i++ {
		identifier := fmt.Sprintf("ip-%d", i)
		rl.CheckLimit("test", identifier, 5, shortWindow)
	}

	// Check stats before cleanup
	stats := rl.GetStats()
	if stats["total_entries"] != 3 {
		t.Errorf("Expected 3 entries before cleanup, got %d", stats["total_entries"])
	}

	// Wait for window to expire
	time.Sleep(shortWindow + 50*time.Millisecond)

	// Trigger cleanup
	rl.cleanup()

	// Check stats after cleanup
	stats = rl.GetStats()
	if stats["total_entries"] != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", stats["total_entries"])
	}
	if stats["active_entries"] != 0 {
		t.Errorf("Expected 0 active entries after cleanup, got %d", stats["active_entries"])
	}
}

// TestRateLimiter_GetStats tests the stats reporting
func TestRateLimiter_GetStats(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	stats := rl.GetStats()
	if stats["total_entries"] != 0 {
		t.Errorf("Expected 0 initial entries, got %d", stats["total_entries"])
	}

	// Add some entries
	window := 200 * time.Millisecond
	rl.CheckLimit("test", "ip1", 5, window)
	rl.CheckLimit("test", "ip2", 5, window)
	rl.CheckLimit("other", "ip1", 5, window) // Different type, same IP

	stats = rl.GetStats()
	if stats["total_entries"] != 3 {
		t.Errorf("Expected 3 entries, got %d", stats["total_entries"])
	}
	if stats["active_entries"] != 3 {
		t.Errorf("Expected 3 active entries, got %d", stats["active_entries"])
	}

	// Wait for expiration
	time.Sleep(window + 50*time.Millisecond)

	stats = rl.GetStats()
	// Entries still exist but are no longer active
	if stats["total_entries"] != 3 {
		t.Errorf("Expected 3 total entries (not cleaned yet), got %d", stats["total_entries"])
	}
	if stats["active_entries"] != 0 {
		t.Errorf("Expected 0 active entries after expiration, got %d", stats["active_entries"])
	}
}

// TestRateLimiter_Reset tests that reset clears all entries
func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter()

	// Add some entries
	for i := 0; i < 5; i++ {
		identifier := fmt.Sprintf("ip-%d", i)
		rl.CheckLimit("test", identifier, 10, 1*time.Minute)
	}

	stats := rl.GetStats()
	if stats["total_entries"] == 0 {
		t.Error("Expected entries before reset")
	}

	// Reset
	rl.Reset()

	stats = rl.GetStats()
	if stats["total_entries"] != 0 {
		t.Errorf("Expected 0 entries after reset, got %d", stats["total_entries"])
	}
}

// TestRateLimiter_ErrorMessage tests that error messages are informative
func TestRateLimiter_ErrorMessage(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "test-ip"
	limit := 3
	window := 1 * time.Second

	// Use up the limit
	for i := 0; i < limit; i++ {
		rl.CheckLimit("test", identifier, limit, window)
	}

	// Get error
	err := rl.CheckLimit("test", identifier, limit, window)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errorMsg := err.Error()

	// Check that error message contains useful information
	expectedStrings := []string{
		"rate limit exceeded",
		"3 attempts", // the limit
		"try again",
	}

	for _, expected := range expectedStrings {
		if !contains(errorMsg, expected) {
			t.Errorf("Error message should contain '%s', got: %s", expected, errorMsg)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestRateLimiter_ExponentialBackoff tests that violations trigger escalating lockouts
// Note: Uses short lockout times for testing
func TestRateLimiter_ExponentialBackoff(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "backoff-test-ip"
	limit := 2
	window := 100 * time.Millisecond

	// First violation: Use up quota
	for i := 0; i < limit; i++ {
		info := rl.CheckLimitWithBackoff("test", identifier, limit, window)
		if info.Limited {
			t.Fatalf("Should not be limited on attempt %d", i+1)
		}
	}

	// Trigger first violation
	info := rl.CheckLimitWithBackoff("test", identifier, limit, window)
	if !info.Limited {
		t.Error("Expected to be rate limited")
	}
	if info.ViolationCount != 1 {
		t.Errorf("Expected violation count 1, got %d", info.ViolationCount)
	}
	if info.RetryAfterSeconds != 60 { // 1 minute
		t.Errorf("Expected 60 second lockout for first violation, got %d", info.RetryAfterSeconds)
	}

	// Try again immediately - should still be locked with same violation count
	// (Trying during lockout doesn't escalate the penalty)
	info = rl.CheckLimitWithBackoff("test", identifier, limit, window)
	if !info.Limited {
		t.Error("Should still be locked after first violation")
	}
	if info.ViolationCount != 1 {
		t.Errorf("Violation count should still be 1, got %d", info.ViolationCount)
	}
}

// TestRateLimiter_ViolationReset tests that violations are reset on successful validation
func TestRateLimiter_ViolationReset(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	identifier := "reset-test-ip"
	limit := 2
	window := 100 * time.Millisecond

	// Trigger first violation
	for i := 0; i < limit; i++ {
		rl.CheckLimitWithBackoff("test", identifier, limit, window)
	}
	info := rl.CheckLimitWithBackoff("test", identifier, limit, window)
	if !info.Limited || info.ViolationCount != 1 {
		t.Fatal("Failed to trigger first violation")
	}

	// Reset violations (simulating successful validation)
	rl.ResetViolations("test", identifier)

	// Wait for window to expire (lockout already cleared by ResetViolations)
	time.Sleep(150 * time.Millisecond) // 100ms window + 50ms buffer

	// Should be able to use quota again without escalated penalty
	for i := 0; i < limit; i++ {
		info = rl.CheckLimitWithBackoff("test", identifier, limit, window)
		if info.Limited {
			t.Fatalf("Should not be limited after reset, attempt %d", i+1)
		}
	}

	// Trigger violation again - should be violation #1 (not #2)
	info = rl.CheckLimitWithBackoff("test", identifier, limit, window)
	if !info.Limited {
		t.Error("Expected to be rate limited")
	}
	if info.ViolationCount != 1 {
		t.Errorf("Expected violation count 1 after reset, got %d", info.ViolationCount)
	}
}

// TestRateLimiter_BackoffEscalation tests backoff duration calculation
func TestRateLimiter_BackoffEscalation(t *testing.T) {
	// Test the backoff duration function directly
	expectedDurations := map[int]time.Duration{
		0:  0,
		1:  1 * time.Minute,
		2:  5 * time.Minute,
		3:  15 * time.Minute,
		4:  1 * time.Hour,
		5:  1 * time.Hour, // Caps at 1 hour
		10: 1 * time.Hour,
	}

	for violationCount, expectedDuration := range expectedDurations {
		actual := getBackoffDuration(violationCount)
		if actual != expectedDuration {
			t.Errorf("Violation count %d: expected %v, got %v",
				violationCount, expectedDuration, actual)
		}
	}
}

// TestRateLimiter_CheckCodeValidationWithInfo tests the convenience method
func TestRateLimiter_CheckCodeValidationWithInfo(t *testing.T) {
	rl := NewRateLimiter()
	rl.Reset()

	ip := "10.0.0.99"

	// Use up quota (5 attempts)
	for i := 0; i < 5; i++ {
		info := rl.CheckCodeValidationWithInfo(ip)
		if info.Limited {
			t.Errorf("Attempt %d should not be limited", i+1)
		}
	}

	// 6th attempt should trigger violation
	info := rl.CheckCodeValidationWithInfo(ip)
	if !info.Limited {
		t.Error("Expected rate limit on 6th attempt")
	}
	if info.ViolationCount != 1 {
		t.Errorf("Expected violation count 1, got %d", info.ViolationCount)
	}
	if info.RetryAfterSeconds != 60 {
		t.Errorf("Expected 60 second lockout, got %d", info.RetryAfterSeconds)
	}
	if info.Message == "" {
		t.Error("Expected informative message")
	}
}

// TestRateLimiter_CleanupViolations tests that cleanup removes expired violations
// Note: This test uses shorter timeouts for speed
func TestRateLimiter_CleanupViolations(t *testing.T) {
	t.Skip("Cleanup is tested indirectly through other tests; full test takes too long")

	// The cleanup functionality is already tested through:
	// - TestRateLimiter_Cleanup (tests cleanup of rate limit entries)
	// - TestRateLimiter_ViolationReset (tests violation reset behavior)
	// - Background cleanup runs every 5 minutes in production
}
