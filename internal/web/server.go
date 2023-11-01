package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/CHTJonas/go-lg/assets"
	"github.com/CHTJonas/go-lg/internal/storage"
	"github.com/cbroglie/mustache"
	"github.com/labstack/echo/v4"
	"go.uber.org/ratelimit"
)

type Server struct {
	e       *echo.Echo
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
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Internal != nil {
					log.Println(httpErr.Internal)
				}
				msg := strings.Title(httpErr.Message.(string))
				body := fmt.Sprintln(httpErr.Code, msg)
				c.String(httpErr.Code, body)
			} else {
				log.Println(err)
				code := http.StatusInternalServerError
				msg := fmt.Sprintln(code, "Internal Server Error")
				c.String(code, msg)
			}
		}
	}

	// Hide startup banner
	e.HideBanner = true
	e.HidePort = true

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

	e.Server.ReadTimeout = 60 * time.Second
	e.Server.WriteTimeout = 60 * time.Second
	e.Server.IdleTimeout = 90 * time.Second

	s.e = e
	return s
}

func (serv *Server) Start(addr string) error {
	log.Printf("Started Echo v%s listening on %s", echo.Version, addr)
	return serv.e.Start(addr)
}

func (serv *Server) Stop(ctx context.Context) error {
	serv.e.Server.SetKeepAlivesEnabled(false)
	return serv.e.Shutdown(ctx)
}

func (serv *Server) getHomePage(c echo.Context) error {
	partial, _ := assets.ReadFile("home.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Home Page", "version": serv.version}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getPingForm(c echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Ping Report", "submissionURL": "/ping/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitPingForm(c echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	protocolVersion := c.QueryParam("protocolVersion")
	var cmd *exec.Cmd
	if protocolVersion == "4" {
		cmd = exec.Command("ping", "-4", "-c", "4", target)
	} else if protocolVersion == "6" {
		cmd = exec.Command("ping", "-6", "-c", "4", target)
	} else {
		cmd = exec.Command("ping", "-c", "4", target)
	}
	stdout, ok := run(cmd)
	if !ok {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("ping", stdout)
	return redirect("ping", uid, c)
}

func (serv *Server) getPingResults(c echo.Context) error {
	uid := c.Param("uid")
	stdout := serv.s.Read("ping", uid)
	if len(stdout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Ping Report", "code": string(stdout), "submissionURL": "/ping/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getTracerouteForm(c echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Traceroute Report", "submissionURL": "/traceroute/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitTracerouteForm(c echo.Context) error {
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
	stdout, ok := run(cmd)
	if !ok {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("traceroute", stdout)
	return redirect("traceroute", uid, c)
}

func (serv *Server) getTracerouteResults(c echo.Context) error {
	uid := c.Param("uid")
	stdout := serv.s.Read("traceroute", uid)
	if len(stdout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Traceroute Report", "code": string(stdout), "submissionURL": "/traceroute/action", "placeholder": "Hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getWHOISForm(c echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "WHOIS Report", "submissionURL": "/whois/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitWHOISForm(c echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	cmd := exec.Command("whois", target)
	stdout, ok := run(cmd)
	if !ok {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("whois", stdout)
	return redirect("whois", uid, c)
}

func (serv *Server) getWHOISResults(c echo.Context) error {
	uid := c.Param("uid")
	stdout := serv.s.Read("whois", uid)
	if len(stdout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "WHOIS Report", "code": string(stdout), "submissionURL": "/whois/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getHostForm(c echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Host Report", "submissionURL": "/host/action", "placeholder": "Hostname or IP"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitHostForm(c echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	cmd := exec.Command("host", strings.Split(target, " ")...)
	stdout, ok := run(cmd)
	if !ok {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("host", stdout)
	return redirect("host", uid, c)
}

func (serv *Server) getHostResults(c echo.Context) error {
	uid := c.Param("uid")
	stdout := serv.s.Read("host", uid)
	if len(stdout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Host Report", "code": string(stdout), "submissionURL": "/host/action", "placeholder": "Hostname or IP"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getDigForm(c echo.Context) error {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "DIG Report", "submissionURL": "/dig/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) submitDigForm(c echo.Context) error {
	target := c.QueryParam("target")
	target = strings.TrimSpace(target)
	cmd := exec.Command("dig", strings.Split(target, " ")...)
	stdout, ok := run(cmd)
	if !ok {
		return echo.ErrInternalServerError
	}
	uid, _ := serv.s.TrimWrite("dig", stdout)
	return redirect("dig", uid, c)
}

func (serv *Server) getDigResults(c echo.Context) error {
	uid := c.Param("uid")
	stdout := serv.s.Read("dig", uid)
	if len(stdout) == 0 {
		return echo.ErrNotFound
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "DIG Report", "code": string(stdout), "submissionURL": "/dig/action", "placeholder": "Query"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	return c.HTML(http.StatusOK, str)
}

func (serv *Server) getRobotsTXT(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /")
}

func redirect(base string, uid string, c echo.Context) error {
	return c.Redirect(http.StatusTemporaryRedirect, "/"+base+"/"+uid)
}
