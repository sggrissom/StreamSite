//go:build release || !frontend
// +build release !frontend

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"stream"
	"stream/cfg"
)

//go:embed dist
var embedded embed.FS

const Port = 3000

func main() {
	// Create required directories
	os.MkdirAll("data", 0755)
	os.MkdirAll("static", 0755)

	distFS, err := fs.Sub(embedded, "dist")
	if err != nil {
		log.Fatalf("failed to sub‐fs: %v", err)
	}

	// Create the application with frontend assets
	app, db := stream.MakeApplicationWithDB()
	app.Frontend = distFS
	app.StaticData = os.DirFS(cfg.StaticDir)

	// Wrap app with SSE handler to bypass vbeam's ResponseWriter wrapping
	handler := stream.NewSSEWrapper(app, db)

	addr := fmt.Sprintf(":%d", Port)
	log.Printf("listening on %s\n", addr)
	var appServer = &http.Server{Addr: addr, Handler: handler}
	appServer.ListenAndServe()
}
