package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-lg",
	Short: "Web-based network looking glass featuring reports with sharable URLs",
}

func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		// No need to log here
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if os.Getenv("JOURNAL_STREAM") != "" {
		log.Default().SetFlags(0)
	}
}
