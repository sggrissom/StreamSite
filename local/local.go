package main

import (
	"fmt"
	"net/http"
	"os"
	"stream"
	"stream/cfg"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbeam/esbuilder"
	"go.hasen.dev/vbeam/local_ui"
)

const Port = 3000
const Domain = "stream.localhost"
const FEDist = ".serve/frontend"

func StartLocalServer() {
	defer vbeam.NiceStackTraceOnPanic()

	vbeam.RunBackServer(cfg.Backport)
	app, db := stream.MakeApplicationWithDB()
	app.Frontend = os.DirFS(FEDist)
	app.StaticData = os.DirFS(cfg.StaticDir)
	vbeam.GenerateTSBindings(app, "frontend/server.ts")

	handler := stream.NewSSEWrapper(app, db)

	var addr = fmt.Sprintf(":%d", Port)
	var appServer = &http.Server{Addr: addr, Handler: handler}
	appServer.ListenAndServe()
}

var FEOpts = esbuilder.FEBuildOptions{
	FERoot: "frontend",
	EntryTS: []string{
		"main.tsx",
	},
	EntryHTML: []string{"index.html"},
	CopyItems: []string{},
	Outdir:    FEDist,
	Define: map[string]string{
		"BROWSER": "true",
		"DEBUG":   "true",
		"VERBOSE": "false",
	},
}

var FEWatchDirs = []string{
	"frontend",
}

func main() {
	os.MkdirAll(".serve", 0755)
	os.MkdirAll(".serve/static", 0755)
	os.MkdirAll(".serve/frontend", 0755)

	var args local_ui.LocalServerArgs
	args.Domain = Domain
	args.Port = Port
	args.FEOpts = FEOpts
	args.FEWatchDirs = FEWatchDirs
	args.StartServer = StartLocalServer

	local_ui.LaunchUI(args)
}
