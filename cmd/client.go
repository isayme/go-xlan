package cmd

import (
	"github.com/isayme/go-xlan/cmd/client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Run client of xlan",
	Run: func(cmd *cobra.Command, args []string) {
		client.Run()
	},
}
