package backend

import (
	"errors"
	"fmt"
	"stream/cfg"
	"strings"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// RegisterChatMethods registers chat-related API procedures
func RegisterChatMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, SendChatMessage)
	vbeam.RegisterProc(app, GetChatHistory)
}

// ChatMessage represents a single chat message in a room
type ChatMessage struct {
	Id           int       `json:"id"`        // Auto-incremented unique ID
	RoomId       int       `json:"roomId"`    // Which room this message belongs to
	UserId       int       `json:"userId"`    // User ID (can be -1 for code sessions or >0 for JWT users)
	SessionToken string    `json:"-"`         // Code session token (not exposed to frontend)
	UserName     string    `json:"userName"`  // Display name for message
	Text         string    `json:"text"`      // Message content (max 500 chars)
	Timestamp    time.Time `json:"timestamp"` // When message was sent
}

// Pack function for serialization
func PackChatMessage(self *ChatMessage, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.Int(&self.UserId, buf)
	vpack.String(&self.SessionToken, buf)
	vpack.String(&self.UserName, buf)
	vpack.String(&self.Text, buf)
	vpack.Time(&self.Timestamp, buf)
}

// Database buckets and indexes

// ChatMessagesBkt stores messages by ID
var ChatMessagesBkt = vbolt.Bucket(&cfg.Info, "chat_messages", vpack.FInt, PackChatMessage)

// ChatByRoomIdx finds all messages for a room (term=roomId, target=messageId)
var ChatByRoomIdx = vbolt.Index(&cfg.Info, "chat_by_room", vpack.FInt, vpack.FInt)

// API Request/Response types

type SendChatMessageRequest struct {
	RoomId int    `json:"roomId"`
	Text   string `json:"text"`
}

type SendChatMessageResponse struct {
	Message ChatMessage `json:"message"`
}

type GetChatHistoryRequest struct {
	RoomId int `json:"roomId"`
	Limit  int `json:"limit"` // Default: 100, max: 500
}

type GetChatHistoryResponse struct {
	Messages []ChatMessage `json:"messages"`
}

// SendChatMessage handles sending a new chat message
// Both JWT users and code sessions can send messages
func SendChatMessage(ctx *vbeam.Context, req SendChatMessageRequest) (resp SendChatMessageResponse, err error) {
	// 1. Get auth user (supports both JWT and code sessions)
	user, err := GetAuthUser(ctx)
	if err != nil {
		return resp, ErrAuthFailure
	}

	// 2. Validate room exists
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("Room not found")
	}

	// 3. Validate message
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return resp, errors.New("Message cannot be empty")
	}
	if len(text) > 500 {
		return resp, errors.New("Message must be 500 characters or less")
	}

	// 4. Rate limiting (1 message per 2 seconds)
	// Use user ID for JWT users and code sessions (code sessions have userId = -1)
	// For code sessions, we could use session token but userId=-1 works for rate limiting
	identifier := fmt.Sprintf("%d-%s", user.Id, ctx.Token)

	if rateLimitErr := globalRateLimiter.CheckChatSend(identifier); rateLimitErr != nil {
		LogDebug(LogCategorySystem, "Rate limit exceeded for chat send", map[string]interface{}{
			"identifier": identifier,
			"roomId":     req.RoomId,
			"error":      rateLimitErr.Error(),
		})
		return resp, errors.New("Please wait before sending another message")
	}

	// 5. Determine display name
	userName := user.Name
	if user.Id == -1 {
		// code session
		userName = "Viewer"
	}

	// 6. Create message
	vbeam.UseWriteTx(ctx)
	msg := ChatMessage{
		Id:        vbolt.NextIntId(ctx.Tx, ChatMessagesBkt),
		RoomId:    req.RoomId,
		UserId:    user.Id,
		UserName:  userName,
		Text:      text,
		Timestamp: time.Now(),
	}

	// 7. Store message in database
	vbolt.Write(ctx.Tx, ChatMessagesBkt, msg.Id, &msg)
	vbolt.SetTargetSingleTerm(ctx.Tx, ChatByRoomIdx, msg.Id, msg.RoomId)
	vbolt.TxCommit(ctx.Tx)

	// 8. Broadcast to all viewers via SSE
	sseManager.BroadcastChatMessage(msg.RoomId, msg)

	LogDebug(LogCategorySystem, "Chat message sent", map[string]interface{}{
		"roomId":     req.RoomId,
		"userId":     user.Id,
		"messageId":  msg.Id,
		"isCodeAuth": user.Id == -1,
	})

	return SendChatMessageResponse{Message: msg}, nil
}

// GetChatHistory retrieves chat message history for a room
// Both JWT users and code sessions can read messages
func GetChatHistory(ctx *vbeam.Context, req GetChatHistoryRequest) (resp GetChatHistoryResponse, err error) {
	// 1. Validate authentication (allows both JWT and code sessions)
	user, err := GetAuthUser(ctx)
	if err != nil {
		return resp, ErrAuthFailure
	}

	// 2. Validate room exists
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("Room not found")
	}

	// 3. Set default limit
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	// 4. Query messages using index
	var messageIds []int
	vbolt.ReadTermTargets(ctx.Tx, ChatByRoomIdx, req.RoomId, &messageIds, vbolt.Window{
		Offset: 0,
		Limit:  limit,
	})

	// 5. Load full message objects
	messages := []ChatMessage{}
	for _, id := range messageIds {
		var msg ChatMessage
		vbolt.Read(ctx.Tx, ChatMessagesBkt, id, &msg)
		if msg.Id != 0 {
			// Don't expose session token to frontend
			msg.SessionToken = ""
			messages = append(messages, msg)
		}
	}

	LogDebug(LogCategorySystem, "Chat history retrieved", map[string]interface{}{
		"roomId":       req.RoomId,
		"userId":       user.Id,
		"messageCount": len(messages),
	})

	return GetChatHistoryResponse{Messages: messages}, nil
}

// DeleteChatMessagesForRoom deletes all chat messages for a room
// Called when a stream ends
func DeleteChatMessagesForRoom(tx *vbolt.Tx, roomId int) error {
	// 1. Get all message IDs for the room
	var messageIds []int
	vbolt.ReadTermTargets(tx, ChatByRoomIdx, roomId, &messageIds, vbolt.Window{})

	// 2. Delete each message
	for _, msgId := range messageIds {
		vbolt.Delete(tx, ChatMessagesBkt, msgId)
	}

	// 3. Remove all index entries for this room
	for _, msgId := range messageIds {
		vbolt.SetTargetSingleTerm(tx, ChatByRoomIdx, msgId, -1)
	}

	LogInfo(LogCategorySystem, "Deleted chat messages for room", map[string]interface{}{
		"roomId":       roomId,
		"messageCount": len(messageIds),
	})

	return nil
}
