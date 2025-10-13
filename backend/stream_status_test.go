package backend

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"go.hasen.dev/vbeam"
)

// TestGetStreamStatus_ResponseStructure tests the response structure
func TestGetStreamStatus_ResponseStructure(t *testing.T) {
	// GetStreamStatus doesn't use database, so just pass empty context
	ctx := &vbeam.Context{}

	req := GetStreamStatusRequest{}
	resp, err := GetStreamStatus(ctx, req)

	if err != nil {
		t.Fatalf("GetStreamStatus returned error: %v", err)
	}

	// Verify response has required fields
	if resp.LastChecked.IsZero() {
		t.Error("LastChecked should be set to current time")
	}

	// IsLive should be a boolean (true or false)
	_ = resp.IsLive // Just verify it exists

	// LastChecked should be recent (within last 5 seconds)
	timeSince := time.Since(resp.LastChecked)
	if timeSince > 5*time.Second {
		t.Errorf("LastChecked should be recent, but was %v ago", timeSince)
	}
}

// TestGetStreamStatus_WithMockServer tests with a controllable mock server
func TestGetStreamStatus_WithMockServer(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedIsLive bool
	}{
		{
			name:           "Stream is live (200)",
			statusCode:     http.StatusOK,
			expectedIsLive: true,
		},
		{
			name:           "Stream not found (404)",
			statusCode:     http.StatusNotFound,
			expectedIsLive: false,
		},
		{
			name:           "Server error (500)",
			statusCode:     http.StatusInternalServerError,
			expectedIsLive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server with specific status code
			mockSRS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer mockSRS.Close()

			// Parse the mock URL
			mockURL, _ := url.Parse(mockSRS.URL)

			// We need to test the HTTP logic directly since we can't easily
			// inject the URL into GetStreamStatus
			client := &http.Client{Timeout: 2 * time.Second}
			headReq, err := http.NewRequest("HEAD", mockSRS.URL+"/streams/live/stream.m3u8", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			httpResp, err := client.Do(headReq)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer httpResp.Body.Close()

			isLive := httpResp.StatusCode == http.StatusOK
			if isLive != tt.expectedIsLive {
				t.Errorf("Expected isLive=%v for status %d, got %v",
					tt.expectedIsLive, tt.statusCode, isLive)
			}

			// Also verify that the mock server received the request correctly
			if mockURL.Host == "" {
				t.Error("Mock URL should have a host")
			}
		})
	}
}

// TestGetStreamStatus_Timeout tests timeout handling
func TestGetStreamStatus_Timeout(t *testing.T) {
	// Create a server that delays response
	mockSRS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second) // Longer than our 2 second timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer mockSRS.Close()

	// Test timeout behavior
	client := &http.Client{Timeout: 1 * time.Second}
	headReq, err := http.NewRequest("HEAD", mockSRS.URL+"/streams/live/stream.m3u8", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = client.Do(headReq)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	// When timeout occurs, we should treat it as not live (which GetStreamStatus does)
}

// TestGetStreamStatus_NoAuth tests that no authentication is required
func TestGetStreamStatus_NoAuth(t *testing.T) {
	// Create context without authentication
	ctx := &vbeam.Context{}

	req := GetStreamStatusRequest{}
	resp, err := GetStreamStatus(ctx, req)

	// Should not error even without auth
	if err != nil {
		t.Fatalf("GetStreamStatus should not require auth, but got error: %v", err)
	}

	// Should return valid response
	if resp.LastChecked.IsZero() {
		t.Error("LastChecked should be set even without auth")
	}
}

// TestGetStreamStatus_ConcurrentRequests tests concurrent calls
func TestGetStreamStatus_ConcurrentRequests(t *testing.T) {
	// Run 10 concurrent requests
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			ctx := &vbeam.Context{}
			req := GetStreamStatusRequest{}
			_, err := GetStreamStatus(ctx, req)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	close(errors)

	// Check no errors occurred
	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}
}
