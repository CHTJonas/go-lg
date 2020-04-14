package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cbroglie/mustache"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		context := map[string]string{"title": "Home Page", "name": "World"}
		str, _ := mustache.RenderFileInLayout("assets/home.html.mustache", "assets/layout.html.mustache", context)
		fmt.Fprintf(w, str)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
