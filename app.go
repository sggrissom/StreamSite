package stream

import (
	"log"
	"net/http"
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

func MakeApplication() *vbeam.Application {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	if cfg.IsRelease {
		vbeam.InitRotatingLogger("stream")
	}

	db := OpenDB(cfg.DBPath)
	var app = vbeam.NewApplication("Stream", db)

	backend.RegisterStreamProxy(app)

	return app
}

func MakeSecureApplication() http.Handler {
	app := MakeApplication()
	return app
}
