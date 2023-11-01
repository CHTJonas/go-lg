package assets

import (
	"embed"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

//go:embed *
var content embed.FS

func ReadFile(name string) ([]byte, error) {
	return content.ReadFile(name)
}

func Server() echo.HandlerFunc {
	fs := http.FS(content)
	srv := http.FileServer(fs)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=0")
		srv.ServeHTTP(w, r)
	})
	return echo.WrapHandler(h)
}
