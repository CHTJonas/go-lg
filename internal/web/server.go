package web

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/CHTJonas/go-lg/assets"
	"github.com/CHTJonas/go-lg/internal/logging"
	"github.com/CHTJonas/go-lg/internal/storage"
	"github.com/cbroglie/mustache"
	"github.com/gorilla/mux"
	"go.uber.org/ratelimit"
)

type Server struct {
	r       *mux.Router
	s       *storage.Store
	srv     *http.Server
	version string
	logger  *logging.PrefixedLogger
	rl      ratelimit.Limiter
}

const loggingPrefix string = "http"

func NewServer(store *storage.Store, version string, level logging.Level) *Server {
	s := &Server{
		s:       store,
		version: version,
		logger:  logging.NewPrefixedLogger(loggingPrefix, level),
		rl:      ratelimit.New(5),
	}
	r := mux.NewRouter().StrictSlash(true)
	r.PathPrefix("/static/").Handler(assets.Server())
	r.HandleFunc("/", s.getHomePage)
	r.HandleFunc("/ping", s.getPingForm)
	r.HandleFunc("/ping/action", s.submitPingForm)
	r.HandleFunc("/ping/{uid}", s.getPingResults)
	r.HandleFunc("/traceroute", s.getTracerouteForm)
	r.HandleFunc("/traceroute/action", s.submitTracerouteForm)
	r.HandleFunc("/traceroute/{uid}", s.getTracerouteResults)
	r.HandleFunc("/whois", s.getWHOISForm)
	r.HandleFunc("/whois/action", s.submitWHOISForm)
	r.HandleFunc("/whois/{uid}", s.getWHOISResults)
	r.Use(s.loggingMiddleware)
	r.Use(s.rateLimitingMiddleware)
	s.r = r
	return s
}

func (serv *Server) Start(addr string) error {
	serv.srv = &http.Server{
		Addr:         addr,
		Handler:      serv.r,
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 90,
		ErrorLog:     serv.logger.Logger,
	}
	return serv.srv.ListenAndServe()
}

func (serv *Server) Stop(ctx context.Context) error {
	serv.srv.SetKeepAlivesEnabled(false)
	return serv.srv.Shutdown(ctx)
}

func (serv *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		method := r.Method
		uri := r.RequestURI
		proto := r.Proto
		serv.logger.Infof("%s \"%s %s %s\"", ip, method, uri, proto)
		next.ServeHTTP(w, r)
	})
}

func (serv *Server) rateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serv.rl.Take()
		next.ServeHTTP(w, r)
	})
}

func (serv *Server) getHomePage(w http.ResponseWriter, r *http.Request) {
	partial, _ := assets.ReadFile("home.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Home Page", "version": serv.version}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func (serv *Server) getPingForm(w http.ResponseWriter, r *http.Request) {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Ping Report", "submissionURL": "/ping/action", "placeholder": "Target hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func (serv *Server) submitPingForm(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	protocolVersion := r.URL.Query().Get("protocolVersion")
	var cmd *exec.Cmd
	if protocolVersion == "4" {
		cmd = exec.Command("ping", "-4", "-c", "4", target)
	} else if protocolVersion == "6" {
		cmd = exec.Command("ping", "-6", "-c", "4", target)
	} else {
		cmd = exec.Command("ping", "-c", "4", target)
	}
	stdout, _ := cmd.Output()
	uid, _ := serv.s.TrimWrite("ping", stdout)
	redirect("ping", uid, w, r)
}

func (serv *Server) getPingResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("ping", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Ping Report", "code": string(stdout), "submissionURL": "/ping/action", "placeholder": "Target hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func (serv *Server) getTracerouteForm(w http.ResponseWriter, r *http.Request) {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Traceroute Report", "submissionURL": "/traceroute/action", "placeholder": "Target hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func (serv *Server) submitTracerouteForm(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	protocolVersion := r.URL.Query().Get("protocolVersion")
	var cmd *exec.Cmd
	if protocolVersion == "4" {
		cmd = exec.Command("mtr", "-4", "-c", "4", "-bez", "-w", target)
	} else if protocolVersion == "6" {
		cmd = exec.Command("mtr", "-6", "-c", "4", "-bez", "-w", target)
	} else {
		cmd = exec.Command("mtr", "-c", "4", "-bez", "-w", target)
	}
	stdout, _ := cmd.Output()
	uid, _ := serv.s.TrimWrite("traceroute", stdout)
	redirect("traceroute", uid, w, r)
}

func (serv *Server) getTracerouteResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("traceroute", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "Traceroute Report", "code": string(stdout), "submissionURL": "/traceroute/action", "placeholder": "Target hostname or IP", "checkboxes": "yes"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func (serv *Server) getWHOISForm(w http.ResponseWriter, r *http.Request) {
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "WHOIS Report", "submissionURL": "/whois/action", "placeholder": "Query name"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func (serv *Server) submitWHOISForm(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	cmd := exec.Command("whois", target)
	stdout, _ := cmd.Output()
	uid, _ := serv.s.TrimWrite("whois", stdout)
	redirect("whois", uid, w, r)
}

func (serv *Server) getWHOISResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("whois", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	partial, _ := assets.ReadFile("form.html.mustache")
	layout, _ := assets.ReadFile("layout.html.mustache")
	context := map[string]string{"title": "WHOIS Report", "code": string(stdout), "submissionURL": "/whois/action", "placeholder": "Query name"}
	str, _ := mustache.RenderInLayout(string(partial), string(layout), context)
	fmt.Fprint(w, str)
}

func redirect(base string, uid string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/"+base+"/"+uid, http.StatusTemporaryRedirect)
}
