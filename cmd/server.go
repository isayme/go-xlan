package cmd

import (
	"github.com/isayme/go-xlan/cmd/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run server of xlan",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run()
	},
}
