package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/chtjonas/go-lg/internal/logging"
	"github.com/chtjonas/go-lg/internal/storage"
	"github.com/chtjonas/go-lg/internal/web"
)

var ver string

func main() {
	store := storage.NewStore("/tmp/badger", logging.INFO)
	defer store.Close()

	serv := web.NewServer(store, ver)
	go func() {
		if err := serv.Start("127.0.0.1:8080"); err != nil {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	serv.Stop(ctx)
}
