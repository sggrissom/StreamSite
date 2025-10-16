package stream

import (
	"net/http"
	"stream/backend"

	"go.hasen.dev/vbeam"
	"go.hasen.dev/vbolt"
)

type SSEWrapper struct {
	app        *vbeam.Application
	sseHandler http.HandlerFunc
}

func NewSSEWrapper(app *vbeam.Application, db *vbolt.DB) *SSEWrapper {
	return &SSEWrapper{
		app:        app,
		sseHandler: backend.MakeStreamRoomEventsHandler(db),
	}
}

func (w *SSEWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/room/events" {
		w.sseHandler(rw, r)
		return
	}

	w.app.ServeHTTP(rw, r)
}
