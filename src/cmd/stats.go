package cmd

import (
    "github.com/spf13/cobra"
    "log"
    "taskc"
)

func init() {
    RootCmd.AddCommand(statsCommand)
}

var statsCommand = &cobra.Command{
    Use:   "stats",
    Short: "Get stats from taskd",
    Long:  `This sends a message of the type "statistics" to taskd, see "gather" for more options`,
    Run: func(cmd *cobra.Command, args []string) {
        rc := taskc.ReadRC()
        conn := taskc.Connect(rc)
        taskc.Stats(conn, rc["taskd.credentials"])
        resp := taskc.Recv(conn)
        conn.Close()
        out := taskc.ParseResponse(resp)
        log.Println(out)
    },
}
