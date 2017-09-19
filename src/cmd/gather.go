package cmd

import (
    "github.com/spf13/cobra"
    "gopkg.in/alexcesaro/statsd.v2"
    "log"
    "taskc"
)

func init() {
    RootCmd.AddCommand(gatherCommand)
}

func getClient() *statsd.Client {
    c, err := statsd.New(statsd.Address("localhost:8125"))
    if err != nil {
        panic(err)
    }
    defer c.Close()
    return c
}

var gatherCommand = &cobra.Command{
    Use:   "gather",
    Short: "Shove stats into statsd",
    Long:  `This sends a message of the type "statistics" to taskd, then parses the headers and passes them to statsd`,
    Run: func(cmd *cobra.Command, args []string) {
        rc := taskc.ReadRC()
        conn := taskc.Connect(rc)
        taskc.Stats(conn, rc["taskd.credentials"])
        resp := taskc.Recv(conn)
        conn.Close()
        out := taskc.ParseResponse(resp)
        client := getClient()
        client.Gauge("taskd_uptime", out.RawHeaders["uptime"])
        log.Println(out.RawHeaders["uptime"])
        client.Flush()
    },
}
