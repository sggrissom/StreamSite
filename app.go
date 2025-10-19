package stream

import (
	"log"
	"stream/backend"
	"stream/cfg"

	"github.com/joho/godotenv"
	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

var Info vbolt.Info

func OpenDB(dbpath string) *vbolt.DB {
	dbConnection := vbolt.Open(dbpath)
	vbolt.InitBuckets(dbConnection, &cfg.Info)
	return dbConnection
}

func MakeApplicationWithDB() (*vbeam.Application, *vbolt.DB) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize rotating log files for both local and production
	vbeam.InitRotatingLogger("stream")

	// Log application startup
	backend.LogInfo(backend.LogCategorySystem, "Stream application starting", map[string]interface{}{
		"version":   "1.0.0",
		"dbPath":    cfg.DBPath,
		"staticDir": cfg.StaticDir,
	})

	db := OpenDB(cfg.DBPath)

	// Start background jobs
	backend.StartCodeSessionCleanup(db)

	var app = vbeam.NewApplication("Stream", db)

	backend.SetupAuth(app)
	backend.RegisterUserMethods(app)
	backend.RegisterRoleMethods(app)
	backend.RegisterStudioMethods(app)
	backend.RegisterStudioMembershipMethods(app)
	backend.RegisterCodeAccessMethods(app)
	backend.RegisterStreamProxy(app)
	backend.RegisterRoomStreamProxy(app)

	// SRS HTTP callbacks (no auth required - SRS makes these calls)
	vbeam.RegisterProc(app, backend.ValidateStreamKey)
	vbeam.RegisterProc(app, backend.HandleStreamUnpublish)

	return app, db
}
