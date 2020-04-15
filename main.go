package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/cbroglie/mustache"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		context := map[string]string{"title": "Home Page", "name": "World"}
		str, _ := mustache.RenderFileInLayout("assets/home.html.mustache", "assets/layout.html.mustache", context)
		fmt.Fprint(w, str)
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("ping", "-c", "4", "1.1.1.1")
		stdout, _ := cmd.Output()
		context := map[string]string{"title": "Ping Cloudflare", "code": string(stdout)}
		str, _ := mustache.RenderFileInLayout("assets/ping.html.mustache", "assets/layout.html.mustache", context)
		fmt.Fprint(w, str)
	})

	http.HandleFunc("/traceroute", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command("mtr", "-c", "4", "--report", "1.1.1.1")
		stdout, _ := cmd.Output()
		context := map[string]string{"title": "Traceroute to Cloudflare", "code": string(stdout)}
		str, _ := mustache.RenderFileInLayout("assets/traceroute.html.mustache", "assets/layout.html.mustache", context)
		fmt.Fprint(w, str)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
