package main

import (
	"github.com/chtjonas/go-lg/internal/storage"
	"github.com/chtjonas/go-lg/internal/web"
)

func main() {
	store := storage.NewStore("/tmp/badger")
	defer store.Close()

	serv := web.NewServer(store)
	serv.Start()
}
