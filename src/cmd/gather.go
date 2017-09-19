package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "log"
    "net"
    "taskc"
)

func init() {
    RootCmd.AddCommand(gatherCommand)
}

func getConn() net.Conn {
    conn, err := net.Dial("udp", "localhost:8125")
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
        rc := taskc.ReadRC()
        conn := taskc.Connect(rc)
        taskc.Stats(conn, rc["taskd.credentials"])
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
