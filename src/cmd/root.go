package cmd

import (
	"github.com/spf13/cobra"
	"taskc"
)

var RootCmd = &cobra.Command{
	Use:   "task-client",
	Short: " makes quick connections to taskd",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var Settings taskc.TaskSettings

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&Settings.Server, "server", "", "the taskd server to connect to")
	RootCmd.PersistentFlags().StringVar(&Settings.Certificate, "certificate", "", "the user cert for auth")
	RootCmd.PersistentFlags().StringVar(&Settings.CACert, "cacert", "", "the server's ca cert")
	RootCmd.PersistentFlags().StringVar(&Settings.Key, "key", "", "the user key for auth")
	RootCmd.PersistentFlags().StringVar(&Settings.Creds, "credentials", "", "the user credentials (in group/user/uuid form) for auth (not needed for healthcheck)")
	RootCmd.PersistentFlags().BoolVar(&Settings.NoRC, "norc", false, "Don't attempt to parse taskrc. You must provide auth details with flags.")
}

func Execute() {
	RootCmd.Execute()
}

func initConfig() {
	return
}
