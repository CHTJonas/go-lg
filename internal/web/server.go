package web

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/chtjonas/go-lg/internal/storage"
	"github.com/gorilla/mux"
)

type Server struct {
	r *mux.Router
	s *storage.Store
}

func NewServer(store *storage.Store) *Server {
	s := &Server{
		s: store,
	}
	r := mux.NewRouter()
	r.HandleFunc("/", s.homeHandler)
	r.HandleFunc("/ping", s.performPingHandler)
	r.HandleFunc("/ping/{uid}", s.recallPingHandler)
	r.HandleFunc("/traceroute", s.performTracerouteHandler)
	r.HandleFunc("/traceroute/{uid}", s.recallTracerouteHandler)
	s.r = r
	return s
}

func (serv *Server) Start() {
	http.Handle("/", serv.r)
	http.ListenAndServe(":8080", nil)
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
	uid := strings.TrimPrefix(r.URL.Path, "/ping/")
	stdout := serv.s.Read("ping", uid)
	context := map[string]string{"title": "Ping Cloudflare", "code": string(stdout)}
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
	uid := strings.TrimPrefix(r.URL.Path, "/traceroute/")
	stdout := serv.s.Read("traceroute", uid)
	context := map[string]string{"title": "Traceroute to Cloudflare", "code": string(stdout)}
	str, _ := mustache.RenderFileInLayout("assets/traceroute.html.mustache", "assets/layout.html.mustache", context)
	fmt.Fprint(w, str)
}

func redirect(uid string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.String()+"/"+uid, http.StatusTemporaryRedirect)
}
