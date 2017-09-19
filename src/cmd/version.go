package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"taskc"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of taskc-go",
	Long:  `All software has versions. This is taskc-go's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("taskc-go version %s\n", taskc.Version())
	},
}
