package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "task-client",
	Short: " makes quick connections to taskd",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		cmd.Help()
	},
}

var Settings struct {
	Key, Server, Certificate, CACert string
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&Settings.Server, "server", "", "the taskd server to connect to")
	RootCmd.PersistentFlags().StringVar(&Settings.Certificate, "certificate", "", "the user cert for auth")
	RootCmd.PersistentFlags().StringVar(&Settings.CACert, "cacert", "", "the server's ca cert")
	RootCmd.PersistentFlags().StringVar(&Settings.Key, "key", "", "the user key for auth")
}

func Execute() {
	RootCmd.Execute()
}

func initConfig() {
	return
}
