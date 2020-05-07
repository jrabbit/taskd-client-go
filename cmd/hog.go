package cmd

import (
    "github.com/spf13/cobra"
    "github.com/jrabbit/taskd-client-go/taskc"
    "time"
)

var wait int

func init() {
    hogCmd.PersistentFlags().IntVar(&wait, "wait-time", 6, "the length of time in seconds to do nothing whiel connected to taskd")
    RootCmd.AddCommand(hogCmd)
}

var hogCmd = &cobra.Command{
    Use:   "hog",
    Short: "connect to taskd server and do nothing",
    Long:  `This command like the others in task-client uses your .taskrc file by default.`,
    Run: func(cmd *cobra.Command, args []string) {
        conn, settings := taskc.SimpleConn(Settings)
        taskc.CheckCreds(settings) // because why not
        conn.Handshake()           // try something to mess up taskd
        // Proceed to do nothing
        time.Sleep(time.Duration(wait) * time.Second)
        conn.Close()
    },
}
