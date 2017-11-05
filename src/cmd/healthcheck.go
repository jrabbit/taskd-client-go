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
        conn, _ := taskc.SimpleConn(Settings)
        conn.Close()
        fmt.Println("OK")
    },
}
