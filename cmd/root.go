package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	ipAddress   string
	port        int
	databaseUrl string
	hmacSecret  string
)

var rootCmd = &cobra.Command{
	Use:   "mothership-server",
	Short: "Log your IoT data",
	Long:  `The purpose of this application is to provide a gRPC service for storing your internet of things time-series data.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do nothing...
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
