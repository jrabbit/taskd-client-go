package cmd

import (
    "github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
    Use:   "task-client",
    Short: " makes quick connectinos to taskd",
    Run: func(cmd *cobra.Command, args []string) {
        // Do Stuff Here
    },
}

func init() {
    cobra.OnInitialize(initConfig)
}

func Execute() {
    RootCmd.Execute()
}

func initConfig() {
    return
}
