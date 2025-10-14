package backend

import (
	"crypto/rand"
	"encoding/base64"
	"stream/cfg"
	"time"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// Studio represents an organizational unit that can have multiple rooms
type Studio struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxRooms    int       `json:"maxRooms"` // Configurable limit on number of rooms
	OwnerId     int       `json:"ownerId"`  // User who created/owns the studio
	Creation    time.Time `json:"creation"`
}

// Room represents a streaming endpoint within a studio
type Room struct {
	Id         int       `json:"id"`
	StudioId   int       `json:"studioId"`
	RoomNumber int       `json:"roomNumber"`
	Name       string    `json:"name"`
	StreamKey  string    `json:"streamKey"`
	IsActive   bool      `json:"isActive"` // Currently streaming
	Creation   time.Time `json:"creation"`
}

// Stream represents a streaming session in a room
type Stream struct {
	Id          int       `json:"id"`
	StudioId    int       `json:"studioId"`
	RoomId      int       `json:"roomId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime,omitempty"` // Null if currently live
	CreatedBy   int       `json:"createdBy"`         // User ID who started the stream
}

// Packing functions for vbolt serialization

func PackStudio(self *Studio, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.String(&self.Name, buf)
	vpack.String(&self.Description, buf)
	vpack.Int(&self.MaxRooms, buf)
	vpack.Int(&self.OwnerId, buf)
	vpack.Time(&self.Creation, buf)
}

func PackRoom(self *Room, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.StudioId, buf)
	vpack.Int(&self.RoomNumber, buf)
	vpack.String(&self.Name, buf)
	vpack.String(&self.StreamKey, buf)
	vpack.Bool(&self.IsActive, buf)
	vpack.Time(&self.Creation, buf)
}

func PackStream(self *Stream, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.Id, buf)
	vpack.Int(&self.StudioId, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.String(&self.Title, buf)
	vpack.String(&self.Description, buf)
	vpack.Time(&self.StartTime, buf)
	vpack.Time(&self.EndTime, buf)
	vpack.Int(&self.CreatedBy, buf)
}

// Buckets for entity storage

// Studios bucket: studioId -> Studio
var StudiosBkt = vbolt.Bucket(&cfg.Info, "studios", vpack.FInt, PackStudio)

// Rooms bucket: roomId -> Room
var RoomsBkt = vbolt.Bucket(&cfg.Info, "rooms", vpack.FInt, PackRoom)

// Streams bucket: streamId -> Stream
var StreamsBkt = vbolt.Bucket(&cfg.Info, "streams", vpack.FInt, PackStream)

// Indexes for relationship queries

// RoomsByStudioIdx: studioId (term) -> roomId (target)
// Find all rooms for a given studio
var RoomsByStudioIdx = vbolt.Index(&cfg.Info, "rooms_by_studio", vpack.FInt, vpack.FInt)

// StreamsByStudioIdx: studioId (term) -> streamId (target)
// Find all streams for a studio
var StreamsByStudioIdx = vbolt.Index(&cfg.Info, "streams_by_studio", vpack.FInt, vpack.FInt)

// StreamsByRoomIdx: roomId (term) -> streamId (target)
// Find all streams for a room
var StreamsByRoomIdx = vbolt.Index(&cfg.Info, "streams_by_room", vpack.FInt, vpack.FInt)

// Lookup buckets for unique constraints

// RoomStreamKeyBkt: streamKey (string) -> roomId (int)
// Fast lookup of room by stream key (for authentication)
var RoomStreamKeyBkt = vbolt.Bucket(&cfg.Info, "room_stream_key", vpack.StringZ, vpack.Int)

// Helper functions

// GenerateStreamKey creates a random stream key for a room
func GenerateStreamKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GetStudioById retrieves a studio by ID
func GetStudioById(tx *vbolt.Tx, studioId int) (studio Studio) {
	vbolt.Read(tx, StudiosBkt, studioId, &studio)
	return
}

// GetRoom retrieves a room by ID
func GetRoom(tx *vbolt.Tx, roomId int) (room Room) {
	vbolt.Read(tx, RoomsBkt, roomId, &room)
	return
}

// GetRoomByStreamKey retrieves a room by its stream key
func GetRoomByStreamKey(tx *vbolt.Tx, streamKey string) (room Room) {
	var roomId int
	vbolt.Read(tx, RoomStreamKeyBkt, streamKey, &roomId)
	if roomId > 0 {
		room = GetRoom(tx, roomId)
	}
	return
}

// GetStream retrieves a stream by ID
func GetStream(tx *vbolt.Tx, streamId int) (stream Stream) {
	vbolt.Read(tx, StreamsBkt, streamId, &stream)
	return
}

// ListStudioRooms returns all rooms for a studio
func ListStudioRooms(tx *vbolt.Tx, studioId int) []Room {
	var rooms []Room
	var roomIds []int

	vbolt.ReadTermTargets(tx, RoomsByStudioIdx, studioId, &roomIds, vbolt.Window{})

	for _, roomId := range roomIds {
		room := GetRoom(tx, roomId)
		if room.Id > 0 {
			rooms = append(rooms, room)
		}
	}

	return rooms
}

// API Request/Response types

type CreateStudioRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxRooms    int    `json:"maxRooms"` // Optional, will use default if 0
}

type CreateStudioResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Studio  Studio `json:"studio,omitempty"`
}

type ListMyStudiosRequest struct {
	// Empty for now
}

type StudioWithRole struct {
	Studio
	MyRole     StudioRole `json:"myRole"`
	MyRoleName string     `json:"myRoleName"`
}

type ListMyStudiosResponse struct {
	Studios []StudioWithRole `json:"studios"`
}

type GetStudioRequest struct {
	StudioId int `json:"studioId"`
}

type GetStudioResponse struct {
	Success    bool       `json:"success"`
	Error      string     `json:"error,omitempty"`
	Studio     Studio     `json:"studio,omitempty"`
	MyRole     StudioRole `json:"myRole"`
	MyRoleName string     `json:"myRoleName"`
}

type GetStudioDashboardRequest struct {
	StudioId int `json:"studioId"`
}

type GetStudioDashboardResponse struct {
	Success    bool                `json:"success"`
	Error      string              `json:"error,omitempty"`
	Studio     Studio              `json:"studio,omitempty"`
	MyRole     StudioRole          `json:"myRole"`
	MyRoleName string              `json:"myRoleName"`
	Rooms      []Room              `json:"rooms,omitempty"`
	Members    []MemberWithDetails `json:"members,omitempty"`
}

type UpdateStudioRequest struct {
	StudioId    int    `json:"studioId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxRooms    int    `json:"maxRooms"`
}

type UpdateStudioResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Studio  Studio `json:"studio,omitempty"`
}

type DeleteStudioRequest struct {
	StudioId int `json:"studioId"`
}

type DeleteStudioResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type CreateRoomRequest struct {
	StudioId int    `json:"studioId"`
	Name     string `json:"name"`
}

type CreateRoomResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Room    Room   `json:"room,omitempty"`
}

type ListRoomsRequest struct {
	StudioId int `json:"studioId"`
}

type ListRoomsResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Rooms   []Room `json:"rooms,omitempty"`
}

type GetRoomStreamKeyRequest struct {
	RoomId int `json:"roomId"`
}

type GetRoomStreamKeyResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	StreamKey string `json:"streamKey,omitempty"`
}

type UpdateRoomRequest struct {
	RoomId int    `json:"roomId"`
	Name   string `json:"name"`
}

type UpdateRoomResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Room    Room   `json:"room,omitempty"`
}

type RegenerateStreamKeyRequest struct {
	RoomId int `json:"roomId"`
}

type RegenerateStreamKeyResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	StreamKey string `json:"streamKey,omitempty"`
}

type DeleteRoomRequest struct {
	RoomId int `json:"roomId"`
}

type DeleteRoomResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type GetRoomDetailsRequest struct {
	RoomId int `json:"roomId"`
}

type GetRoomDetailsResponse struct {
	Success    bool       `json:"success"`
	Error      string     `json:"error,omitempty"`
	Room       Room       `json:"room,omitempty"`
	StudioName string     `json:"studioName,omitempty"`
	MyRole     StudioRole `json:"myRole"`
	MyRoleName string     `json:"myRoleName"`
}

type RoomWithStudio struct {
	Room
	StudioName string `json:"studioName"`
}

type ListMyAccessibleRoomsRequest struct {
	// Empty for now
}

type ListMyAccessibleRoomsResponse struct {
	Rooms []RoomWithStudio `json:"rooms"`
}

// API Procedures

func CreateStudio(ctx *vbeam.Context, req CreateStudioRequest) (resp CreateStudioResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Only StreamAdmin or higher can create studios
	if caller.Role < RoleStreamAdmin {
		resp.Success = false
		resp.Error = "Only stream admins can create studios"
		return
	}

	// Validate input
	if req.Name == "" {
		resp.Success = false
		resp.Error = "Studio name is required"
		return
	}

	// Use default max rooms if not specified
	maxRooms := req.MaxRooms
	if maxRooms <= 0 {
		maxRooms = cfg.DefaultMaxRooms
	}

	if maxRooms > 50 {
		resp.Success = false
		resp.Error = "Maximum rooms cannot exceed 50"
		return
	}

	// Create studio
	vbeam.UseWriteTx(ctx)

	studio := Studio{
		Id:          vbolt.NextIntId(ctx.Tx, StudiosBkt),
		Name:        req.Name,
		Description: req.Description,
		MaxRooms:    maxRooms,
		OwnerId:     caller.Id,
		Creation:    time.Now(),
	}

	vbolt.Write(ctx.Tx, StudiosBkt, studio.Id, &studio)

	// Create owner membership for the creator
	membership := StudioMembership{
		UserId:   caller.Id,
		StudioId: studio.Id,
		Role:     StudioRoleOwner,
		JoinedAt: time.Now(),
	}
	membershipId := vbolt.NextIntId(ctx.Tx, MembershipBkt)
	vbolt.Write(ctx.Tx, MembershipBkt, membershipId, &membership)
	vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByUserIdx, membershipId, caller.Id)
	vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByStudioIdx, membershipId, studio.Id)

	vbolt.TxCommit(ctx.Tx)

	// Log studio creation
	LogInfo(LogCategorySystem, "Studio created", map[string]interface{}{
		"studioId":   studio.Id,
		"studioName": studio.Name,
		"ownerId":    caller.Id,
		"ownerEmail": caller.Email,
		"maxRooms":   maxRooms,
	})

	resp.Success = true
	resp.Studio = studio
	return
}

func ListMyStudios(ctx *vbeam.Context, req ListMyStudiosRequest) (resp ListMyStudiosResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		err = authErr
		return
	}

	// Get all studios user is a member of
	studios := ListUserStudios(ctx.Tx, caller.Id)

	// Build response with role information
	resp.Studios = make([]StudioWithRole, 0, len(studios))
	for _, studio := range studios {
		role := GetUserStudioRole(ctx.Tx, caller.Id, studio.Id)
		resp.Studios = append(resp.Studios, StudioWithRole{
			Studio:     studio,
			MyRole:     role,
			MyRoleName: GetStudioRoleName(role),
		})
	}

	return
}

func GetStudio(ctx *vbeam.Context, req GetStudioRequest) (resp GetStudioResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has permission to view this studio
	role := GetUserStudioRole(ctx.Tx, caller.Id, studio.Id)

	// Site admins can view all studios
	if caller.Role != RoleSiteAdmin && role == -1 {
		resp.Success = false
		resp.Error = "You do not have permission to view this studio"
		return
	}

	// Site admins who aren't members get Owner role for display purposes
	if caller.Role == RoleSiteAdmin && role == -1 {
		role = StudioRoleOwner
	}

	resp.Success = true
	resp.Studio = studio
	resp.MyRole = role
	resp.MyRoleName = GetStudioRoleName(role)
	return
}

func GetStudioDashboard(ctx *vbeam.Context, req GetStudioDashboardRequest) (resp GetStudioDashboardResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has permission to view this studio
	role := GetUserStudioRole(ctx.Tx, caller.Id, studio.Id)

	// Site admins can view all studios
	if caller.Role != RoleSiteAdmin && role == -1 {
		resp.Success = false
		resp.Error = "You do not have permission to view this studio"
		return
	}

	// Site admins who aren't members get Owner role for display purposes
	if caller.Role == RoleSiteAdmin && role == -1 {
		role = StudioRoleOwner
	}

	// Get all rooms for this studio
	var roomIds []int
	vbolt.ReadTermTargets(ctx.Tx, RoomsByStudioIdx, studio.Id, &roomIds, vbolt.Window{})

	var rooms []Room
	for _, roomId := range roomIds {
		room := GetRoom(ctx.Tx, roomId)
		if room.Id != 0 {
			rooms = append(rooms, room)
		}
	}

	// If no rooms found, initialize empty slice
	if rooms == nil {
		rooms = []Room{}
	}

	// Get all members for this studio
	memberships := ListStudioMembers(ctx.Tx, studio.Id)

	// Enhance with user details
	members := make([]MemberWithDetails, 0, len(memberships))
	for _, membership := range memberships {
		user := GetUser(ctx.Tx, membership.UserId)
		if user.Id > 0 {
			members = append(members, MemberWithDetails{
				StudioMembership: membership,
				UserName:         user.Name,
				UserEmail:        user.Email,
				RoleName:         GetStudioRoleName(membership.Role),
			})
		}
	}

	resp.Success = true
	resp.Studio = studio
	resp.MyRole = role
	resp.MyRoleName = GetStudioRoleName(role)
	resp.Rooms = rooms
	resp.Members = members
	return
}

func UpdateStudio(ctx *vbeam.Context, req UpdateStudioRequest) (resp UpdateStudioResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can update studios"
		return
	}

	// Validate input
	if req.Name == "" {
		resp.Success = false
		resp.Error = "Studio name is required"
		return
	}

	if req.MaxRooms <= 0 {
		resp.Success = false
		resp.Error = "Maximum rooms must be at least 1"
		return
	}

	if req.MaxRooms > 50 {
		resp.Success = false
		resp.Error = "Maximum rooms cannot exceed 50"
		return
	}

	// Update studio fields
	vbeam.UseWriteTx(ctx)

	studio.Name = req.Name
	studio.Description = req.Description
	studio.MaxRooms = req.MaxRooms

	vbolt.Write(ctx.Tx, StudiosBkt, studio.Id, &studio)

	vbolt.TxCommit(ctx.Tx)

	// Log studio update
	LogInfo(LogCategorySystem, "Studio updated", map[string]interface{}{
		"studioId":   studio.Id,
		"studioName": studio.Name,
		"updatedBy":  caller.Id,
		"userEmail":  caller.Email,
		"maxRooms":   req.MaxRooms,
	})

	resp.Success = true
	resp.Studio = studio
	return
}

func DeleteStudio(ctx *vbeam.Context, req DeleteStudioRequest) (resp DeleteStudioResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Owner permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleOwner) {
		resp.Success = false
		resp.Error = "Only studio owners can delete studios"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Cascade delete all related data

	// 1. Delete all rooms and their stream keys
	rooms := ListStudioRooms(ctx.Tx, studio.Id)
	for _, room := range rooms {
		// Remove stream key lookup
		vbolt.Delete(ctx.Tx, RoomStreamKeyBkt, room.StreamKey)
		// Unindex room from studio
		vbolt.SetTargetSingleTerm(ctx.Tx, RoomsByStudioIdx, room.Id, -1)
		// Delete room
		vbolt.Delete(ctx.Tx, RoomsBkt, room.Id)
	}

	// 2. Delete all streams
	var streamIds []int
	vbolt.ReadTermTargets(ctx.Tx, StreamsByStudioIdx, studio.Id, &streamIds, vbolt.Window{})
	for _, streamId := range streamIds {
		stream := GetStream(ctx.Tx, streamId)
		if stream.Id > 0 {
			// Unindex from both studio and room indexes
			vbolt.SetTargetSingleTerm(ctx.Tx, StreamsByStudioIdx, streamId, -1)
			vbolt.SetTargetSingleTerm(ctx.Tx, StreamsByRoomIdx, streamId, -1)
			// Delete stream
			vbolt.Delete(ctx.Tx, StreamsBkt, streamId)
		}
	}

	// 3. Delete all memberships
	memberships := ListStudioMembers(ctx.Tx, studio.Id)
	for _, membership := range memberships {
		// Get membership ID (we need to find it by iterating user's memberships)
		var membershipIds []int
		vbolt.ReadTermTargets(ctx.Tx, MembershipByUserIdx, membership.UserId, &membershipIds, vbolt.Window{})
		for _, membershipId := range membershipIds {
			m := GetMembership(ctx.Tx, membershipId)
			if m.StudioId == studio.Id {
				// Unindex from both user and studio indexes
				vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByUserIdx, membershipId, -1)
				vbolt.SetTargetSingleTerm(ctx.Tx, MembershipByStudioIdx, membershipId, -1)
				// Delete membership
				vbolt.Delete(ctx.Tx, MembershipBkt, membershipId)
			}
		}
	}

	// 4. Delete the studio itself
	vbolt.Delete(ctx.Tx, StudiosBkt, studio.Id)

	vbolt.TxCommit(ctx.Tx)

	// Log studio deletion
	LogInfo(LogCategorySystem, "Studio deleted", map[string]interface{}{
		"studioId":           studio.Id,
		"studioName":         studio.Name,
		"deletedBy":          caller.Id,
		"userEmail":          caller.Email,
		"roomsDeleted":       len(rooms),
		"streamsDeleted":     len(streamIds),
		"membershipsDeleted": len(memberships),
	})

	resp.Success = true
	return
}

func CreateRoom(ctx *vbeam.Context, req CreateRoomRequest) (resp CreateRoomResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can create rooms"
		return
	}

	// Validate input
	if req.Name == "" {
		resp.Success = false
		resp.Error = "Room name is required"
		return
	}

	// Check current room count
	existingRooms := ListStudioRooms(ctx.Tx, studio.Id)
	if len(existingRooms) >= studio.MaxRooms {
		resp.Success = false
		resp.Error = "Studio has reached maximum room limit"
		return
	}

	// Generate stream key
	streamKey, err := GenerateStreamKey()
	if err != nil {
		resp.Success = false
		resp.Error = "Failed to generate stream key"
		return
	}

	// Calculate next room number
	nextRoomNumber := 1
	for _, room := range existingRooms {
		if room.RoomNumber >= nextRoomNumber {
			nextRoomNumber = room.RoomNumber + 1
		}
	}

	vbeam.UseWriteTx(ctx)

	// Create room
	room := Room{
		Id:         vbolt.NextIntId(ctx.Tx, RoomsBkt),
		StudioId:   studio.Id,
		RoomNumber: nextRoomNumber,
		Name:       req.Name,
		StreamKey:  streamKey,
		IsActive:   false,
		Creation:   time.Now(),
	}

	vbolt.Write(ctx.Tx, RoomsBkt, room.Id, &room)
	vbolt.SetTargetSingleTerm(ctx.Tx, RoomsByStudioIdx, room.Id, studio.Id)
	vbolt.Write(ctx.Tx, RoomStreamKeyBkt, streamKey, &room.Id)

	vbolt.TxCommit(ctx.Tx)

	// Log room creation
	LogInfo(LogCategorySystem, "Room created", map[string]interface{}{
		"roomId":     room.Id,
		"roomName":   room.Name,
		"studioId":   studio.Id,
		"studioName": studio.Name,
		"createdBy":  caller.Id,
		"userEmail":  caller.Email,
	})

	resp.Success = true
	resp.Room = room
	return
}

func ListRooms(ctx *vbeam.Context, req ListRoomsRequest) (resp ListRoomsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, req.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has permission to view this studio (Viewer+)
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleViewer) {
		resp.Success = false
		resp.Error = "You do not have permission to view this studio's rooms"
		return
	}

	// Get all rooms for the studio
	rooms := ListStudioRooms(ctx.Tx, studio.Id)

	resp.Success = true
	resp.Rooms = rooms
	return
}

func GetRoomDetails(ctx *vbeam.Context, req GetRoomDetailsRequest) (resp GetRoomDetailsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		resp.Success = false
		resp.Error = "Room not found"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, room.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has permission to view this studio (Viewer+)
	role := GetUserStudioRole(ctx.Tx, caller.Id, studio.Id)

	// Site admins can view all studios/rooms
	if caller.Role != RoleSiteAdmin && role == -1 {
		resp.Success = false
		resp.Error = "You do not have permission to view this room"
		return
	}

	// Site admins who aren't members get Owner role for display purposes
	if caller.Role == RoleSiteAdmin && role == -1 {
		role = StudioRoleOwner
	}

	resp.Success = true
	resp.Room = room
	resp.StudioName = studio.Name
	resp.MyRole = role
	resp.MyRoleName = GetStudioRoleName(role)
	return
}

func ListMyAccessibleRooms(ctx *vbeam.Context, req ListMyAccessibleRoomsRequest) (resp ListMyAccessibleRoomsResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		err = authErr
		return
	}

	// Get all studios user is a member of
	studios := ListUserStudios(ctx.Tx, caller.Id)

	// Collect all rooms from all studios
	resp.Rooms = make([]RoomWithStudio, 0)
	for _, studio := range studios {
		// Get all rooms for this studio
		rooms := ListStudioRooms(ctx.Tx, studio.Id)

		// Add each room with studio information
		for _, room := range rooms {
			resp.Rooms = append(resp.Rooms, RoomWithStudio{
				Room:       room,
				StudioName: studio.Name,
			})
		}
	}

	return
}

func GetRoomStreamKey(ctx *vbeam.Context, req GetRoomStreamKeyRequest) (resp GetRoomStreamKeyResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		resp.Success = false
		resp.Error = "Room not found"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, room.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Admin+ permission to view stream key
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can view stream keys"
		return
	}

	resp.Success = true
	resp.StreamKey = room.StreamKey
	return
}

func UpdateRoom(ctx *vbeam.Context, req UpdateRoomRequest) (resp UpdateRoomResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		resp.Success = false
		resp.Error = "Room not found"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, room.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can update rooms"
		return
	}

	// Validate input
	if req.Name == "" {
		resp.Success = false
		resp.Error = "Room name is required"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Update room name
	room.Name = req.Name
	vbolt.Write(ctx.Tx, RoomsBkt, room.Id, &room)

	vbolt.TxCommit(ctx.Tx)

	// Log room update
	LogInfo(LogCategorySystem, "Room updated", map[string]interface{}{
		"roomId":     room.Id,
		"roomName":   room.Name,
		"studioId":   studio.Id,
		"studioName": studio.Name,
		"updatedBy":  caller.Id,
		"userEmail":  caller.Email,
	})

	resp.Success = true
	resp.Room = room
	return
}

func RegenerateStreamKey(ctx *vbeam.Context, req RegenerateStreamKeyRequest) (resp RegenerateStreamKeyResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		resp.Success = false
		resp.Error = "Room not found"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, room.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can regenerate stream keys"
		return
	}

	// Generate new stream key
	newStreamKey, err := GenerateStreamKey()
	if err != nil {
		resp.Success = false
		resp.Error = "Failed to generate stream key"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Store old key for logging
	oldStreamKey := room.StreamKey

	// Delete old stream key lookup
	vbolt.Delete(ctx.Tx, RoomStreamKeyBkt, oldStreamKey)

	// Update room with new stream key
	room.StreamKey = newStreamKey
	vbolt.Write(ctx.Tx, RoomsBkt, room.Id, &room)

	// Add new stream key lookup
	vbolt.Write(ctx.Tx, RoomStreamKeyBkt, newStreamKey, &room.Id)

	vbolt.TxCommit(ctx.Tx)

	// Log stream key regeneration (for security audit)
	LogInfo(LogCategorySystem, "Stream key regenerated", map[string]interface{}{
		"roomId":        room.Id,
		"roomName":      room.Name,
		"studioId":      studio.Id,
		"studioName":    studio.Name,
		"regeneratedBy": caller.Id,
		"userEmail":     caller.Email,
		"oldKeyPrefix":  oldStreamKey[:8] + "...", // Only log prefix for security
		"newKeyPrefix":  newStreamKey[:8] + "...",
	})

	resp.Success = true
	resp.StreamKey = newStreamKey
	return
}

func DeleteRoom(ctx *vbeam.Context, req DeleteRoomRequest) (resp DeleteRoomResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		resp.Success = false
		resp.Error = "Authentication required"
		return
	}

	// Get room
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		resp.Success = false
		resp.Error = "Room not found"
		return
	}

	// Get studio
	studio := GetStudioById(ctx.Tx, room.StudioId)
	if studio.Id == 0 {
		resp.Success = false
		resp.Error = "Studio not found"
		return
	}

	// Check if user has Admin+ permission
	if !HasStudioPermission(ctx.Tx, caller.Id, studio.Id, StudioRoleAdmin) {
		resp.Success = false
		resp.Error = "Only studio admins can delete rooms"
		return
	}

	// Check if room is actively streaming
	if room.IsActive {
		resp.Success = false
		resp.Error = "Cannot delete room while it is actively streaming"
		return
	}

	vbeam.UseWriteTx(ctx)

	// Cascade delete all related data

	// 1. Delete all streams for this room
	var streamIds []int
	vbolt.ReadTermTargets(ctx.Tx, StreamsByRoomIdx, room.Id, &streamIds, vbolt.Window{})
	for _, streamId := range streamIds {
		stream := GetStream(ctx.Tx, streamId)
		if stream.Id > 0 {
			// Unindex from both room and studio indexes
			vbolt.SetTargetSingleTerm(ctx.Tx, StreamsByRoomIdx, streamId, -1)
			vbolt.SetTargetSingleTerm(ctx.Tx, StreamsByStudioIdx, streamId, -1)
			// Delete stream
			vbolt.Delete(ctx.Tx, StreamsBkt, streamId)
		}
	}

	// 2. Remove stream key lookup
	vbolt.Delete(ctx.Tx, RoomStreamKeyBkt, room.StreamKey)

	// 3. Unindex room from studio
	vbolt.SetTargetSingleTerm(ctx.Tx, RoomsByStudioIdx, room.Id, -1)

	// 4. Delete the room itself
	vbolt.Delete(ctx.Tx, RoomsBkt, room.Id)

	vbolt.TxCommit(ctx.Tx)

	// Log room deletion
	LogInfo(LogCategorySystem, "Room deleted", map[string]interface{}{
		"roomId":         room.Id,
		"roomName":       room.Name,
		"studioId":       studio.Id,
		"studioName":     studio.Name,
		"deletedBy":      caller.Id,
		"userEmail":      caller.Email,
		"streamsDeleted": len(streamIds),
	})

	resp.Success = true
	return
}

// SRS HTTP Callback Types

type SRSAuthCallback struct {
	ServerId  string `json:"server_id"`
	Action    string `json:"action"`
	ClientId  string `json:"client_id"`
	IP        string `json:"ip"`
	Vhost     string `json:"vhost"`
	App       string `json:"app"`
	TcUrl     string `json:"tcUrl"`
	Stream    string `json:"stream"` // This is the stream key
	Param     string `json:"param"`
	StreamUrl string `json:"stream_url"`
	StreamId  string `json:"stream_id"`
}

type SRSAuthResponse struct {
	Code int `json:"code"` // 0 for success, non-zero for error
}

// ValidateStreamKey handles SRS on_publish callback to authenticate streams
func ValidateStreamKey(ctx *vbeam.Context, req SRSAuthCallback) (resp SRSAuthResponse, err error) {
	// Log the authentication attempt
	LogInfo(LogCategorySystem, "SRS stream authentication", map[string]interface{}{
		"action":    req.Action,
		"stream":    req.Stream,
		"ip":        req.IP,
		"client_id": req.ClientId,
	})

	// The stream key is in the Stream field
	streamKey := req.Stream

	if streamKey == "" {
		LogWarn(LogCategorySystem, "SRS auth failed: empty stream key", map[string]interface{}{
			"ip": req.IP,
		})
		resp.Code = 1 // Reject
		return
	}

	// Look up the room by stream key
	room := GetRoomByStreamKey(ctx.Tx, streamKey)

	if room.Id == 0 {
		LogWarn(LogCategorySystem, "SRS auth failed: invalid stream key", map[string]interface{}{
			"stream_key": streamKey[:8] + "...", // Only log prefix for security
			"ip":         req.IP,
		})
		resp.Code = 1 // Reject
		return
	}

	// Stream key is valid, mark room as active
	vbeam.UseWriteTx(ctx)
	room.IsActive = true
	vbolt.Write(ctx.Tx, RoomsBkt, room.Id, &room)
	vbolt.TxCommit(ctx.Tx)

	// Log successful authentication
	LogInfo(LogCategorySystem, "SRS auth successful, room now live", map[string]interface{}{
		"room_id":    room.Id,
		"room_name":  room.Name,
		"studio_id":  room.StudioId,
		"ip":         req.IP,
		"client_id":  req.ClientId,
		"stream_key": streamKey[:8] + "...",
	})

	resp.Code = 0 // Success
	return
}

// HandleStreamUnpublish handles SRS on_unpublish callback when stream ends
func HandleStreamUnpublish(ctx *vbeam.Context, req SRSAuthCallback) (resp SRSAuthResponse, err error) {
	streamKey := req.Stream

	// Look up the room by stream key
	room := GetRoomByStreamKey(ctx.Tx, streamKey)

	if room.Id > 0 {
		// Mark room as inactive
		vbeam.UseWriteTx(ctx)
		room.IsActive = false
		vbolt.Write(ctx.Tx, RoomsBkt, room.Id, &room)
		vbolt.TxCommit(ctx.Tx)

		LogInfo(LogCategorySystem, "SRS stream ended, room now offline", map[string]interface{}{
			"room_id":    room.Id,
			"room_name":  room.Name,
			"studio_id":  room.StudioId,
			"stream_key": streamKey[:8] + "...",
			"ip":         req.IP,
			"client_id":  req.ClientId,
		})
	} else {
		// Stream key not found, but still return success
		LogInfo(LogCategorySystem, "SRS stream ended (unknown room)", map[string]interface{}{
			"stream_key": streamKey[:8] + "...",
			"ip":         req.IP,
			"client_id":  req.ClientId,
		})
	}

	// Always return success for unpublish callbacks
	resp.Code = 0
	return
}

// RegisterStudioMethods registers studio-related API procedures
func RegisterStudioMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, CreateStudio)
	vbeam.RegisterProc(app, ListMyStudios)
	vbeam.RegisterProc(app, GetStudio)
	vbeam.RegisterProc(app, GetStudioDashboard)
	vbeam.RegisterProc(app, UpdateStudio)
	vbeam.RegisterProc(app, DeleteStudio)
	vbeam.RegisterProc(app, CreateRoom)
	vbeam.RegisterProc(app, ListRooms)
	vbeam.RegisterProc(app, GetRoomDetails)
	vbeam.RegisterProc(app, ListMyAccessibleRooms)
	vbeam.RegisterProc(app, GetRoomStreamKey)
	vbeam.RegisterProc(app, UpdateRoom)
	vbeam.RegisterProc(app, RegenerateStreamKey)
	vbeam.RegisterProc(app, DeleteRoom)
}
