package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/cbroglie/mustache"
	"github.com/dgraph-io/badger"
)

var db *badger.DB

func main() {
	bdb, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		panic(err)
	}
	db = bdb
	defer bdb.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		context := map[string]string{"title": "Home Page", "name": "World"}
		str, _ := mustache.RenderFileInLayout("assets/home.html.mustache", "assets/layout.html.mustache", context)
		fmt.Fprint(w, str)
	})

	http.HandleFunc("/ping/", func(w http.ResponseWriter, r *http.Request) {
		var stdout []byte
		uid := strings.TrimPrefix(r.URL.Path, "/ping/")
		if uid != "" {
			stdout = Read("ping", uid)
			context := map[string]string{"title": "Ping Cloudflare", "code": string(stdout)}
			str, _ := mustache.RenderFileInLayout("assets/ping.html.mustache", "assets/layout.html.mustache", context)
			fmt.Fprint(w, str)
		} else {
			cmd := exec.Command("ping", "-c", "4", "1.1.1.1")
			stdout, _ := cmd.Output()
			uid, _ := Write("ping", stdout)
			Redirect(uid, w, r)
		}
	})

	http.HandleFunc("/traceroute/", func(w http.ResponseWriter, r *http.Request) {
		var stdout []byte
		uid := strings.TrimPrefix(r.URL.Path, "/traceroute/")
		if uid != "" {
			stdout = Read("traceroute", uid)
			context := map[string]string{"title": "Traceroute to Cloudflare", "code": string(stdout)}
			str, _ := mustache.RenderFileInLayout("assets/traceroute.html.mustache", "assets/layout.html.mustache", context)
			fmt.Fprint(w, str)
		} else {
			cmd := exec.Command("mtr", "-c", "4", "--report", "1.1.1.1")
			stdout, _ := cmd.Output()
			uid, _ := Write("traceroute", stdout)
			Redirect(uid, w, r)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Redirect(uid string, w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.String()+"/"+uid, http.StatusTemporaryRedirect)
}

func Read(prefix string, uid string) []byte {
	var stdout []byte
	db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefix + uid))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			stdout = append([]byte{}, val...)
			return nil
		})
	})
	return stdout
}

func Write(prefix string, stdout []byte) (string, error) {
	uid := GenerateUID()
	return uid, db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(prefix+uid), stdout)
	})
}

func GenerateUID() string {
	token := make([]byte, 6)
	rand.Read(token)
	return hex.EncodeToString(token)
}
