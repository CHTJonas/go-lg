package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/chtjonas/go-lg/internal/logging"
	"github.com/chtjonas/go-lg/internal/storage"
	"github.com/chtjonas/go-lg/internal/web"
)

var ver string
var path string

const loggingPrefix string = "app"

func init() {
	flag.StringVar(&path, "data-dir", "/var/lib/go-lg", "path to database storage directory")
	flag.Parse()
}

func main() {
	logLevel := logging.INFO
	applicationLogger := logging.NewPrefixedLogger(loggingPrefix, logLevel)
	applicationLogger.Infof("go-lg version %s starting up...", ver)
	defer applicationLogger.Infof("go-lg will now exit...")

	store := storage.NewStore(path, logLevel)
	defer store.Close()

	serv := web.NewServer(store, ver, logLevel)
	go func() {
		if err := serv.Start("127.0.0.1:8080"); err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	applicationLogger.Infof("Shutdown initiated...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	serv.Stop(ctx)
}
