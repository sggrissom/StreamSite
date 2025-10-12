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

// StudioRole defines the authorization level of a user within a studio
type StudioRole int

const (
	StudioRoleViewer StudioRole = 0 // Can watch streams
	StudioRoleMember StudioRole = 1 // Can stream
	StudioRoleAdmin  StudioRole = 2 // Can manage rooms/members
	StudioRoleOwner  StudioRole = 3 // Full control
)

// GetStudioRoleName returns the human-readable name for a studio role
func GetStudioRoleName(role StudioRole) string {
	switch role {
	case StudioRoleViewer:
		return "Viewer"
	case StudioRoleMember:
		return "Member"
	case StudioRoleAdmin:
		return "Admin"
	case StudioRoleOwner:
		return "Owner"
	default:
		return "Unknown"
	}
}

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

// StudioMembership represents a user's membership and role in a studio
type StudioMembership struct {
	UserId   int        `json:"userId"`
	StudioId int        `json:"studioId"`
	Role     StudioRole `json:"role"`
	JoinedAt time.Time  `json:"joinedAt"`
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

func PackStudioMembership(self *StudioMembership, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.UserId, buf)
	vpack.Int(&self.StudioId, buf)
	vpack.Int((*int)(&self.Role), buf)
	vpack.Time(&self.JoinedAt, buf)
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

// Membership bucket: (userId, studioId) composite key -> StudioMembership
var MembershipBkt = vbolt.Bucket(&cfg.Info, "studio_membership", vpack.FInt, PackStudioMembership)

// Indexes for relationship queries

// RoomsByStudioIdx: studioId (term) -> roomId (target)
// Find all rooms for a given studio
var RoomsByStudioIdx = vbolt.Index(&cfg.Info, "rooms_by_studio", vpack.FInt, vpack.FInt)

// MembershipByUserIdx: userId (term) -> membershipId (target)
// Find all studios a user belongs to
var MembershipByUserIdx = vbolt.Index(&cfg.Info, "membership_by_user", vpack.FInt, vpack.FInt)

// MembershipByStudioIdx: studioId (term) -> membershipId (target)
// Find all members of a studio
var MembershipByStudioIdx = vbolt.Index(&cfg.Info, "membership_by_studio", vpack.FInt, vpack.FInt)

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

// GetMembership retrieves a membership by ID
func GetMembership(tx *vbolt.Tx, membershipId int) (membership StudioMembership) {
	vbolt.Read(tx, MembershipBkt, membershipId, &membership)
	return
}

// GetUserStudioRole returns the user's role in a studio, or -1 if not a member
func GetUserStudioRole(tx *vbolt.Tx, userId int, studioId int) StudioRole {
	// Get all memberships for the user
	var membershipIds []int
	vbolt.ReadTermTargets(tx, MembershipByUserIdx, userId, &membershipIds, vbolt.Window{})

	// Find the membership for this studio
	for _, membershipId := range membershipIds {
		membership := GetMembership(tx, membershipId)
		if membership.StudioId == studioId {
			return membership.Role
		}
	}

	// Not a member
	return -1
}

// HasStudioPermission checks if a user has at least the specified role in a studio
func HasStudioPermission(tx *vbolt.Tx, userId int, studioId int, minRole StudioRole) bool {
	// Site admins have full access to all studios
	user := GetUser(tx, userId)
	if user.Role == RoleSiteAdmin {
		return true
	}

	// Check studio-specific role
	userRole := GetUserStudioRole(tx, userId, studioId)
	return userRole >= minRole
}

// ListUserStudios returns all studios the user is a member of
func ListUserStudios(tx *vbolt.Tx, userId int) []Studio {
	var studios []Studio
	var membershipIds []int

	vbolt.ReadTermTargets(tx, MembershipByUserIdx, userId, &membershipIds, vbolt.Window{})

	for _, membershipId := range membershipIds {
		membership := GetMembership(tx, membershipId)
		studio := GetStudioById(tx, membership.StudioId)
		if studio.Id > 0 {
			studios = append(studios, studio)
		}
	}

	return studios
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

// ListStudioMembers returns all memberships for a studio
func ListStudioMembers(tx *vbolt.Tx, studioId int) []StudioMembership {
	var memberships []StudioMembership
	var membershipIds []int

	vbolt.ReadTermTargets(tx, MembershipByStudioIdx, studioId, &membershipIds, vbolt.Window{})

	for _, membershipId := range membershipIds {
		membership := GetMembership(tx, membershipId)
		if membership.UserId > 0 {
			memberships = append(memberships, membership)
		}
	}

	return memberships
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

// RegisterStudioMethods registers studio-related API procedures
func RegisterStudioMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, CreateStudio)
	vbeam.RegisterProc(app, ListMyStudios)
	vbeam.RegisterProc(app, GetStudio)
}
