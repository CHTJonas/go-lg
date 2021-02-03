package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/CHTJonas/go-lg/internal/logging"
	"github.com/CHTJonas/go-lg/internal/storage"
	"github.com/CHTJonas/go-lg/internal/web"
	"github.com/spf13/cobra"
)

const loggingPrefix string = "app"

var path string
var addr string
var verbosity int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run web server",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		logLevel := logging.Level(verbosity)
		applicationLogger := logging.NewPrefixedLogger(loggingPrefix, logLevel)
		applicationLogger.Infof("go-lg version %s starting up...", version)
		defer applicationLogger.Infof("go-lg will now exit...")

		store := storage.NewStore(path, logLevel)
		defer store.Close()

		serv := web.NewServer(store, version, logLevel)
		go func() {
			if err := serv.Start(addr); err != nil && err != http.ErrServerClosed {
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
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&path, "path", "p", "/var/lib/go-lg", "path to database storage directory")
	serverCmd.Flags().StringVarP(&addr, "bind", "b", "localhost:8080", "address and port to bind to")
	serverCmd.Flags().IntVarP(&verbosity, "verbosity", "v", 2, "logging verbosity")
}
