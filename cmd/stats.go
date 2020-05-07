package cmd

import (
	"github.com/jrabbit/taskd-client-go/taskc"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	RootCmd.AddCommand(statsCommand)
}

var statsCommand = &cobra.Command{
	Use:   "stats",
	Short: "Get stats from taskd",
	Long:  `This sends a message of the type "statistics" to taskd, see "gather" for more options`,
	Run: func(cmd *cobra.Command, args []string) {
		conn, settings := taskc.SimpleConn(Settings)
		taskc.CheckCreds(settings)
		taskc.Stats(conn, settings.Creds)
		resp := taskc.Recv(conn)
		conn.Close()
		out := taskc.ParseResponse(resp)
		for key, value := range out.RawHeaders {
			log.Println(key, value)
		}
	},
}
