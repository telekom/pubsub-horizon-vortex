package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "vortex",
	Short: "A tiny service for sending data from Kafka to MongoDB",
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serveCmd)
}
