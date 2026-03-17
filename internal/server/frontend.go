package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed all:frontend
var frontendFS embed.FS

// serveFrontend configures the Echo server to serve the embedded frontend.
// It serves static files from the embedded filesystem and falls back to
// index.html for client-side routing (SPA).
func serveFrontend(e *echo.Echo) {
	distFS, err := fs.Sub(frontendFS, "frontend")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(distFS))

	// Serve static assets
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		f, err := distFS.Open(path[1:]) // strip leading /
		if err != nil {
			// File not found — serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})))
}
