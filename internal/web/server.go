package web

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/cbroglie/mustache"
	"github.com/chtjonas/go-lg/internal/storage"
	"github.com/gorilla/mux"
)

type Server struct {
	r   *mux.Router
	s   *storage.Store
	srv *http.Server
}

func NewServer(store *storage.Store) *Server {
	s := &Server{
		s: store,
	}
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", s.getHomePage)
	r.HandleFunc("/ping", s.getPingForm)
	r.HandleFunc("/ping/action", s.submitPingForm)
	r.HandleFunc("/ping/{uid}", s.getPingResults)
	r.HandleFunc("/traceroute", s.getTracerouteForm)
	r.HandleFunc("/traceroute/action", s.submitTracerouteForm)
	r.HandleFunc("/traceroute/{uid}", s.getTracerouteResults)
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
	}
	return serv.srv.ListenAndServe()
}

func (serv *Server) Stop(ctx context.Context) error {
	return serv.srv.Shutdown(ctx)
}

func (serv *Server) getHomePage(w http.ResponseWriter, r *http.Request) {
	context := map[string]string{"title": "Home Page", "name": "World"}
	str, _ := mustache.RenderFileInLayout("assets/home.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func (serv *Server) getPingForm(w http.ResponseWriter, r *http.Request) {
	context := map[string]string{"title": "Ping Report", "submissionURL": "/ping/action"}
	str, _ := mustache.RenderFileInLayout("assets/form.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func (serv *Server) submitPingForm(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	cmd := exec.Command("ping", "-c", "4", target)
	stdout, _ := cmd.Output()
	uid, _ := serv.s.Write("ping", stdout)
	redirect("ping", uid, w, r)
}

func (serv *Server) getPingResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("ping", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	context := map[string]string{"title": "Ping Report", "code": string(stdout), "submissionURL": "/ping/action"}
	str, _ := mustache.RenderFileInLayout("assets/form.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func (serv *Server) getTracerouteForm(w http.ResponseWriter, r *http.Request) {
	context := map[string]string{"title": "Traceroute Report", "submissionURL": "/traceroute/action"}
	str, _ := mustache.RenderFileInLayout("assets/form.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func (serv *Server) submitTracerouteForm(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	cmd := exec.Command("mtr", "-c", "4", "--report-wide", target)
	stdout, _ := cmd.Output()
	uid, _ := serv.s.Write("traceroute", stdout)
	redirect("traceroute", uid, w, r)
}

func (serv *Server) getTracerouteResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("traceroute", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	context := map[string]string{"title": "Traceroute Report", "code": string(stdout), "submissionURL": "/traceroute/action"}
	str, _ := mustache.RenderFileInLayout("assets/form.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func redirect(base string, uid string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/"+base+"/"+uid, http.StatusTemporaryRedirect)
}
