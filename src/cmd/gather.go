package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "log"
    "net"
    "taskc"
)

var statsDServer string

func init() {
    RootCmd.AddCommand(gatherCommand)
    gatherCommand.PersistentFlags().StringVar(&statsDServer, "statsd-server", "localhost:8125", "the statsD server to push stats to")
}

func getConn() net.Conn {
    conn, err := net.Dial("udp", statsDServer)
    if err != nil {
        panic(err)
    }
    return conn
}

var gatherCommand = &cobra.Command{
    Use:   "gather",
    Short: "Shove stats into statsd",
    Long:  `This sends a message of the type "statistics" to taskd, then parses the headers and passes them to statsd`,
    Run: func(cmd *cobra.Command, args []string) {
        conn, settings := taskc.SimpleConn(Settings)
        taskc.CheckCreds(settings)
        taskc.Stats(conn, settings.Creds)
        resp := taskc.Recv(conn)
        conn.Close()
        out := taskc.ParseResponse(resp)
        statsConn := getConn()
        for key, value := range out.RawHeaders {
            log.Println(key, value)
            fmt.Fprintf(statsConn, "taskd.%s:%s|g", key, value)
        }
        statsConn.Close()
    },
}
