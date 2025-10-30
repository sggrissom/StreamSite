package backend

import (
	"errors"

	"go.hasen.dev/vbeam"
)

// RegisterEmoteMethods registers emote-related API procedures
func RegisterEmoteMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, SendEmote)
}

// Allowed emotes for the reaction system
var AllowedEmotes = []string{"❤️", "👏", "🔥", "😂", "😮", "👍"}

// SendEmoteRequest contains the emote data to broadcast
type SendEmoteRequest struct {
	RoomId int    `json:"roomId"` // Room ID where emote is sent
	Emote  string `json:"emote"`  // Emoji character
}

type SendEmoteResponse struct {
}

// SendEmote handles emote reactions from viewers
// Allows anonymous emote sending with rate limiting
func SendEmote(ctx *vbeam.Context, req SendEmoteRequest) (resp SendEmoteResponse, err error) {
	// Validate emote is in allowed list
	if !isValidEmote(req.Emote) {
		return resp, errors.New("Invalid emote. Allowed: ❤️, 👏, 🔥, 😂, 😮, 👍")
	}

	// Validate room exists
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("Room not found")
	}

	// identifier for rate limiting
	identifier := ctx.Token

	// Check rate limit (1 emote per 2 seconds)
	if rateLimitErr := globalRateLimiter.CheckEmoteSend(identifier); rateLimitErr != nil {
		LogDebug(LogCategorySystem, "Rate limit exceeded for emote send", map[string]interface{}{
			"identifier": identifier,
			"roomId":     req.RoomId,
			"error":      rateLimitErr.Error(),
		})
		return resp, errors.New("Please wait before sending another emote")
	}

	// Broadcast emote to all viewers via SSE
	sseManager.BroadcastEmote(req.RoomId, req.Emote)

	LogDebug(LogCategorySystem, "Emote broadcast", map[string]interface{}{
		"roomId": req.RoomId,
		"emote":  req.Emote,
	})

	return resp, nil
}

// isValidEmote checks if the emote is in the allowed list
func isValidEmote(emote string) bool {
	for _, allowed := range AllowedEmotes {
		if emote == allowed {
			return true
		}
	}
	return false
}
