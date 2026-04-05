package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/CHTJonas/go-lg/assets"
	"github.com/CHTJonas/go-lg/internal/storage"
	"github.com/cbroglie/mustache"
	"github.com/labstack/echo/v5"
	"go.uber.org/ratelimit"
)

type Server struct {
	e       *echo.Echo
	h       *http.Server
	s       *storage.Store
	version string
	rl      ratelimit.Limiter
}

func NewServer(store *storage.Store, version string) *Server {
	s := &Server{
		s:       store,
		version: version,
		rl:      ratelimit.New(5),
	}

	// Echo instance
	e := echo.New()

	// Handle errors as plaintext
	e.HTTPErrorHandler = func(c *echo.Context, err error) {
		if err != nil {
			if he, ok := err.(*echo.HTTPError); ok {
				msg := strings.Title(he.Message)
				body := fmt.Sprintln(he.Code, msg)
				c.String(he.Code, body)
			} else {
				code := http.StatusInternalServerError
				var sc echo.HTTPStatusCoder
				if errors.As(err, &sc) {
					if tmp := sc.StatusCode(); tmp != 0 {
						code = tmp
					}
				}
				var cErr error
				if c.Request().Method == http.MethodHead {
					cErr = c.NoContent(code)
				} else {
					msg := fmt.Sprintln(code, http.StatusText(code))
					cErr = c.String(code, msg)
				}
				if cErr != nil {
					log.Println("Failed to send error to client", cErr)
				}
			}
		}
	}

	// Reverse proxy
	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	// Middleware
	e.Use(requestIDMiddleware())
	e.Use(loggingMiddleware())
	e.Use(recoveryMiddleware())
	e.Use(serverHeaderMiddleware(version))
	e.Use(clientRateLimitingMiddleware())
	e.Use(serverRateLimitingMiddleware())

	// Routes
	e.GET("/static/*", assets.Server())
	e.GET("/", s.getHomePage)
	e.GET("/ping", s.getPingForm)
	e.GET("/ping/action", s.submitPingForm)
	e.GET("/ping/:uid", s.getPingResults)
	e.GET("/traceroute", s.getTracerouteForm)
	e.GET("/traceroute/action", s.submitTracerouteForm)
	e.GET("/traceroute/:uid", s.getTracerouteResults)
	e.GET("/whois", s.getWHOISForm)
	e.GET("/whois/action", s.submitWHOISForm)
	e.GET("/whois/:uid", s.getWHOISResults)
	e.GET("/host", s.getHostForm)
	e.GET("/host/action", s.submitHostForm)
	e.GET("/host/:uid", s.getHostResults)
	e.GET("/dig", s.getDigForm)
	e.GET("/dig/action", s.submitDigForm)
	e.GET("/dig/:uid", s.getDigResults)
	e.GET("/robots.txt", s.getRobotsTXT)

	s.e = e
	return s
}

func (serv *Server) Start(addr string) error {
	log.Printf("Started Echo v%s listening on %s", echo.Version, addr)
	s := &http.Server{
		Addr:         addr,
		Handler:      serv.e,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  90 * time.Second,
	}
	serv.h = s
	return s.ListenAndServe()
}

func (serv *Server) Stop(ctx context.Context) error {
	serv.h.SetKeepAlivesEnabled(false)
	return serv.h.Shutdown(ctx)
}

func (serv *Server) getHomePage(c *echo.Context) error {
	partial, _ := assets.ReadFile("home.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Home Page", "version": serv.version}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getPingForm(c *echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Ping Report", "submissionURL": "/ping/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitPingForm(c *echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	protocolVersion := c.QueryParam("protocolVersion")
	var cmd *exec.Cmd
	if protocolVersion == "4" {
		cmd = exec.Command("ping", "-4", "-B", "-c", "4", "-v", target)
	} else if protocolVersion == "6" {
		cmd = exec.Command("ping", "-6", "-B", "-c", "4", "-v", target)
	} else {
		cmd = exec.Command("ping", "-B", "-c", "4", "-v", target)
	}
	stderrout := run(cmd)
	if len(stderrout) == 0 {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("ping", stderrout)
	return redirect("ping", uid, c)
}

func (serv *Server) getPingResults(c *echo.Context) error {
	uid := c.Param("uid")
	stderrout := serv.s.Read("ping", uid)
	if len(stderrout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Ping Report", "code": string(stderrout), "submissionURL": "/ping/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getTracerouteForm(c *echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Traceroute Report", "submissionURL": "/traceroute/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitTracerouteForm(c *echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	protocolVersion := c.QueryParam("protocolVersion")
	var cmd *exec.Cmd
	if protocolVersion == "4" {
		cmd = exec.Command("mtr", "-4", "-c", "4", "-bez", "-w", target)
	} else if protocolVersion == "6" {
		cmd = exec.Command("mtr", "-6", "-c", "4", "-bez", "-w", target)
	} else {
		cmd = exec.Command("mtr", "-c", "4", "-bez", "-w", target)
	}
	stderrout := run(cmd)
	if len(stderrout) == 0 {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("traceroute", stderrout)
	return redirect("traceroute", uid, c)
}

func (serv *Server) getTracerouteResults(c *echo.Context) error {
	uid := c.Param("uid")
	stderrout := serv.s.Read("traceroute", uid)
	if len(stderrout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Traceroute Report", "code": string(stderrout), "submissionURL": "/traceroute/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getWHOISForm(c *echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "WHOIS Report", "submissionURL": "/whois/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitWHOISForm(c *echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	cmd := exec.Command("whois", target)
	stderrout := run(cmd)
	if len(stderrout) == 0 {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("whois", stderrout)
	return redirect("whois", uid, c)
}

func (serv *Server) getWHOISResults(c *echo.Context) error {
	uid := c.Param("uid")
	stderrout := serv.s.Read("whois", uid)
	if len(stderrout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "WHOIS Report", "code": string(stderrout), "submissionURL": "/whois/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getHostForm(c *echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Host Report", "submissionURL": "/host/action", "placeholder": "Hostname or IP"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitHostForm(c *echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	cmd := exec.Command("host", strings.Split(target, " ")...)
	stderrout := run(cmd)
	if len(stderrout) == 0 {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("host", stderrout)
	return redirect("host", uid, c)
}

func (serv *Server) getHostResults(c *echo.Context) error {
	uid := c.Param("uid")
	stderrout := serv.s.Read("host", uid)
	if len(stderrout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Host Report", "code": string(stderrout), "submissionURL": "/host/action", "placeholder": "Hostname or IP"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getDigForm(c *echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "DIG Report", "submissionURL": "/dig/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitDigForm(c *echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	cmd := exec.Command("dig", strings.Split(target, " ")...)
	stderrout := run(cmd)
	if len(stderrout) == 0 {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("dig", stderrout)
	return redirect("dig", uid, c)
}

func (serv *Server) getDigResults(c *echo.Context) error {
	uid := c.Param("uid")
	stderrout := serv.s.Read("dig", uid)
	if len(stderrout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "DIG Report", "code": string(stderrout), "submissionURL": "/dig/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getRobotsTXT(c *echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /")
}

func redirect(base string, uid string, c *echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, "/"+base+"/"+uid)
}
