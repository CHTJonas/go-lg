package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-lg",
	Short: "Network looking glass featuring reports with sharable URLs",
	Long:  "",
}

func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
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
