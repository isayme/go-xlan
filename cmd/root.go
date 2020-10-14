package cmd

import (
	"fmt"
	"os"

	"github.com/isayme/go-xlan/xlan/util"
	"github.com/spf13/cobra"
)

var versionFlag bool

func init() {
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "xlan version")
}

var rootCmd = &cobra.Command{
	Use: "xlan",
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			util.PrintVersion()
			os.Exit(0)
		}
	},
}

// Execute run root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
