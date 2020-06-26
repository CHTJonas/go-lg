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
	r.HandleFunc("/", s.homeHandler)
	r.HandleFunc("/ping", s.performPingHandler)
	r.HandleFunc("/ping/{uid}", s.recallPingHandler)
	r.HandleFunc("/traceroute", s.performTracerouteHandler)
	r.HandleFunc("/traceroute/{uid}", s.recallTracerouteHandler)
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

func (serv *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	context := map[string]string{"title": "Home Page", "name": "World"}
	str, _ := mustache.RenderFileInLayout("assets/home.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func (serv *Server) performPingHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("ping", "-c", "4", "1.1.1.1")
	stdout, _ := cmd.Output()
	uid, _ := serv.s.Write("ping", stdout)
	redirect(uid, w, r)
}

func (serv *Server) recallPingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("ping", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	context := map[string]string{"title": "Ping Report", "code": string(stdout)}
	str, _ := mustache.RenderFileInLayout("assets/ping.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func (serv *Server) performTracerouteHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("mtr", "-c", "4", "--report", "1.1.1.1")
	stdout, _ := cmd.Output()
	uid, _ := serv.s.Write("traceroute", stdout)
	redirect(uid, w, r)
}

func (serv *Server) recallTracerouteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stdout := serv.s.Read("traceroute", vars["uid"])
	if len(stdout) == 0 {
		stdout = []byte("HTTP 404 Report Not Found")
		w.WriteHeader(http.StatusNotFound)
	}
	context := map[string]string{"title": "Traceroute Report", "code": string(stdout)}
	str, _ := mustache.RenderFileInLayout("assets/traceroute.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func redirect(uid string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.String()+"/"+uid, http.StatusTemporaryRedirect)
}
