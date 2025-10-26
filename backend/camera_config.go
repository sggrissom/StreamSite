package backend

import (
	"errors"
	"stream/cfg"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
	"go.hasen.dev/vpack"
)

// CameraConfig represents RTSP camera configuration for a room
type CameraConfig struct {
	RoomId  int    `json:"roomId"`
	RTSPURL string `json:"rtspUrl"`
}

// PackCameraConfig serializes CameraConfig for vbolt storage
func PackCameraConfig(self *CameraConfig, buf *vpack.Buffer) {
	vpack.Version(1, buf)
	vpack.Int(&self.RoomId, buf)
	vpack.String(&self.RTSPURL, buf)
}

// CameraConfigBkt stores camera configurations: roomId -> CameraConfig
var CameraConfigBkt = vbolt.Bucket(&cfg.Info, "camera_config", vpack.FInt, PackCameraConfig)

// Helper functions

// GetCameraConfig retrieves camera configuration for a room
func GetCameraConfig(tx *vbolt.Tx, roomId int) (config CameraConfig) {
	vbolt.Read(tx, CameraConfigBkt, roomId, &config)
	return
}

// SetCameraConfigData sets or updates camera configuration for a room
func SetCameraConfigData(tx *vbolt.Tx, roomId int, rtspURL string) {
	config := CameraConfig{
		RoomId:  roomId,
		RTSPURL: rtspURL,
	}
	vbolt.Write(tx, CameraConfigBkt, roomId, &config)
}

// DeleteCameraConfigData removes camera configuration for a room
func DeleteCameraConfigData(tx *vbolt.Tx, roomId int) {
	vbolt.Delete(tx, CameraConfigBkt, roomId)
}

// API Request/Response types

type SetCameraConfigRequest struct {
	RoomId  int    `json:"roomId"`
	RTSPURL string `json:"rtspUrl"`
}

type SetCameraConfigResponse struct {
}

type GetCameraConfigRequest struct {
	RoomId int `json:"roomId"`
}

type GetCameraConfigResponse struct {
	RoomId  int     `json:"roomId"`
	RTSPURL *string `json:"rtspUrl"` // Pointer to allow null if not configured
}

type DeleteCameraConfigRequest struct {
	RoomId int `json:"roomId"`
}

type DeleteCameraConfigResponse struct {
}

// API Procedures

// SetCameraConfig sets the RTSP camera configuration for a room
func SetCameraConfig(ctx *vbeam.Context, req SetCameraConfigRequest) (resp SetCameraConfigResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Validate input
	if req.RoomId <= 0 {
		return resp, errors.New("invalid room ID")
	}

	if req.RTSPURL == "" {
		return resp, errors.New("RTSP URL is required")
	}

	// Get room to verify it exists and get studio ID
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("room not found")
	}

	// Check permissions - require Admin or higher on studio
	if !HasStudioPermission(ctx.Tx, caller.Id, room.StudioId, StudioRoleAdmin) {
		return resp, errors.New("only studio admins can configure cameras")
	}

	// Set camera configuration
	vbeam.UseWriteTx(ctx)
	SetCameraConfigData(ctx.Tx, req.RoomId, req.RTSPURL)
	vbolt.TxCommit(ctx.Tx)

	// Log configuration
	LogInfo(LogCategorySystem, "Camera config set", map[string]interface{}{
		"roomId":   req.RoomId,
		"studioId": room.StudioId,
		"userId":   caller.Id,
		"rtspURL":  req.RTSPURL,
	})

	return
}

// GetCameraConfig retrieves the camera configuration for a room
func GetCameraConfigProc(ctx *vbeam.Context, req GetCameraConfigRequest) (resp GetCameraConfigResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Validate input
	if req.RoomId <= 0 {
		return resp, errors.New("invalid room ID")
	}

	// Get room to verify it exists and get studio ID
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("room not found")
	}

	// Check permissions - require Viewer or higher on studio
	if !HasStudioPermission(ctx.Tx, caller.Id, room.StudioId, StudioRoleViewer) {
		return resp, errors.New("permission denied")
	}

	// Get camera configuration
	config := GetCameraConfig(ctx.Tx, req.RoomId)

	resp.RoomId = req.RoomId
	if config.RoomId > 0 {
		resp.RTSPURL = &config.RTSPURL
	}

	return
}

// DeleteCameraConfig removes the camera configuration for a room
func DeleteCameraConfigProc(ctx *vbeam.Context, req DeleteCameraConfigRequest) (resp DeleteCameraConfigResponse, err error) {
	// Check authentication
	caller, authErr := GetAuthUser(ctx)
	if authErr != nil {
		return resp, errors.New("authentication required")
	}

	// Validate input
	if req.RoomId <= 0 {
		return resp, errors.New("invalid room ID")
	}

	// Get room to verify it exists and get studio ID
	room := GetRoom(ctx.Tx, req.RoomId)
	if room.Id == 0 {
		return resp, errors.New("room not found")
	}

	// Check permissions - require Admin or higher on studio
	if !HasStudioPermission(ctx.Tx, caller.Id, room.StudioId, StudioRoleAdmin) {
		return resp, errors.New("only studio admins can delete camera configurations")
	}

	// Delete camera configuration
	vbeam.UseWriteTx(ctx)
	DeleteCameraConfigData(ctx.Tx, req.RoomId)
	vbolt.TxCommit(ctx.Tx)

	// Log deletion
	LogInfo(LogCategorySystem, "Camera config deleted", map[string]interface{}{
		"roomId":   req.RoomId,
		"studioId": room.StudioId,
		"userId":   caller.Id,
	})

	return
}

// RegisterCameraConfigMethods registers camera config API procedures
func RegisterCameraConfigMethods(app *vbeam.Application) {
	vbeam.RegisterProc(app, SetCameraConfig)
	vbeam.RegisterProc(app, GetCameraConfigProc)
	vbeam.RegisterProc(app, DeleteCameraConfigProc)
}
