package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/chtjonas/go-lg/internal/storage"
)

func main() {
	s := storage.NewStore("/tmp/badger")
	defer s.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		context := map[string]string{"title": "Home Page", "name": "World"}
		str, _ := mustache.RenderFileInLayout("assets/home.html.mustache", "assets/layout.html.mustache", context)
		fmt.Fprint(w, str)
	})

	http.HandleFunc("/ping/", func(w http.ResponseWriter, r *http.Request) {
		var stdout []byte
		uid := strings.TrimPrefix(r.URL.Path, "/ping/")
		if uid != "" {
			stdout = s.Read("ping", uid)
			context := map[string]string{"title": "Ping Cloudflare", "code": string(stdout)}
			str, _ := mustache.RenderFileInLayout("assets/ping.html.mustache", "assets/layout.html.mustache", context)
			fmt.Fprint(w, str)
		} else {
			cmd := exec.Command("ping", "-c", "4", "1.1.1.1")
			stdout, _ := cmd.Output()
			uid, _ := s.Write("ping", stdout)
			Redirect(uid, w, r)
		}
	})

	http.HandleFunc("/traceroute/", func(w http.ResponseWriter, r *http.Request) {
		var stdout []byte
		uid := strings.TrimPrefix(r.URL.Path, "/traceroute/")
		if uid != "" {
			stdout = s.Read("traceroute", uid)
			context := map[string]string{"title": "Traceroute to Cloudflare", "code": string(stdout)}
			str, _ := mustache.RenderFileInLayout("assets/traceroute.html.mustache", "assets/layout.html.mustache", context)
			fmt.Fprint(w, str)
		} else {
			cmd := exec.Command("mtr", "-c", "4", "--report", "1.1.1.1")
			stdout, _ := cmd.Output()
			uid, _ := s.Write("traceroute", stdout)
			Redirect(uid, w, r)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Redirect(uid string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.String()+"/"+uid, http.StatusTemporaryRedirect)
}
