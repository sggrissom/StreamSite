package backend

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"go.hasen.dev/vbeam"
)

func RegisterStreamProxy(app *vbeam.Application) {
	// --- HLS proxy: /streams/* -> SRS on localhost:8080 ---
	srsURL, _ := url.Parse("http://127.0.0.1:8080")
	srsProxy := httputil.NewSingleHostReverseProxy(srsURL)

	// Preserve incoming path (/streams/...) and set helpful headers
	origDirector := srsProxy.Director
	srsProxy.Director = func(r *http.Request) {
		origDirector(r) // sets scheme/host to 127.0.0.1:8080, keeps r.URL.Path intact
		// pass through original Host if you like:
		// r.Host = "127.0.0.1:8080"
	}
	srsProxy.ModifyResponse = func(res *http.Response) error {
		p := res.Request.URL.Path
		if strings.HasSuffix(p, ".m3u8") {
			res.Header.Set("Content-Type", "application/vnd.apple.mpegurl")
			// playlists change constantly; keep them fresh
			res.Header.Set("Cache-Control", "no-store, must-revalidate")
			res.Header.Set("Pragma", "no-cache")
			res.Header.Set("Expires", "0")
		} else if strings.HasSuffix(p, ".ts") || strings.HasSuffix(p, ".m4s") {
			// short cache for segments
			if res.Header.Get("Content-Type") == "" {
				res.Header.Set("Content-Type", "video/mp2t")
			}
			res.Header.Set("Cache-Control", "public, max-age=60")
		}
		return nil
	}

	// Register the proxy handler for /streams/* path
	app.HandleFunc("/streams/", srsProxy.ServeHTTP)
}

// Stream status API types

type GetStreamStatusRequest struct {
	// Empty for now - checks the default stream
}

type GetStreamStatusResponse struct {
	IsLive      bool      `json:"isLive"`
	LastChecked time.Time `json:"lastChecked"`
}

// GetStreamStatus checks if the stream is currently live by probing SRS
func GetStreamStatus(ctx *vbeam.Context, req GetStreamStatusRequest) (resp GetStreamStatusResponse, err error) {
	resp.LastChecked = time.Now()
	resp.IsLive = false

	// Make HTTP HEAD request to SRS to check if stream exists
	srsURL := "http://127.0.0.1:8080/streams/live/stream.m3u8"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// Use HEAD request to avoid downloading the playlist
	headReq, err := http.NewRequest("HEAD", srsURL, nil)
	if err != nil {
		// Error creating request, stream is not live
		err = nil // Don't propagate error, just return isLive=false
		return
	}

	// Make the request
	httpResp, err := client.Do(headReq)
	if err != nil {
		// Error reaching SRS or timeout, stream is not live
		err = nil // Don't propagate error, just return isLive=false
		return
	}
	defer httpResp.Body.Close()

	// Check if we got a successful response
	if httpResp.StatusCode == http.StatusOK {
		resp.IsLive = true
	}

	return
}
