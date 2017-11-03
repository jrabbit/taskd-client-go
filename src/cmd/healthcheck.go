package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "taskc"
)

func init() {
    RootCmd.AddCommand(healthcheckCommand)
}

var healthcheckCommand = &cobra.Command{
    Use:   "healthcheck",
    Short: "run a simple up/down healthcheck",
    Long:  `Ping the taskd server`,
    Run: func(cmd *cobra.Command, args []string) {
        rc := taskc.ReadRC()
        conn := taskc.Connect(rc)
        conn.Close()
        fmt.Println(Settings)
    },
}
