package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CHTJonas/go-lg/internal/storage"
	"github.com/CHTJonas/go-lg/internal/web"
	"github.com/spf13/cobra"
)

var path string
var addr string

var serverCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run web server",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("go-lg version", version)

		store := storage.NewStore(path)
		defer store.Close()
		serv := web.NewServer(store, version)

		log.Println("Starting server...")
		go func() {
			if err := serv.Start(addr); err != nil && err != http.ErrServerClosed {
				log.Fatalln("Startup error:", err.Error())
			}
		}()
		log.Println("Listening on", addr)

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		signal.Notify(c, syscall.SIGQUIT)
		signal.Notify(c, syscall.SIGTERM)
		<-c
		log.Println("Received shutdown signal!")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log.Println("Waiting for server to exit...")
		if err := serv.Stop(ctx); err != nil {
			log.Fatalln("Shutdown error:", err.Error())
		}
		log.Println("Bye-bye!")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&path, "path", "p", "/var/lib/go-lg", "path to database storage directory")
	serverCmd.Flags().StringVarP(&addr, "bind", "b", "localhost:8080", "address and port to bind to")
}
