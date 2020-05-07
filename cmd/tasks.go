package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/jrabbit/taskd-client-go/taskc"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

func init() {
	RootCmd.AddCommand(tasksCmd)
}

func getJSONTasks(resp []byte) []string {
	buff := bytes.NewBuffer(resp)
	scanner := bufio.NewScanner(buff)
	var jsonTasks []string
	for scanner.Scan() {
		text := scanner.Text()
		if len(text) < 1 {
			continue
		} else if strings.Split(text, "")[0] == "{" {
			jsonTasks = append(jsonTasks, scanner.Text())
		}
	}
	return jsonTasks
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "grab json tasks for your user from taskd server",
	Long: `This subcommand downloads all your taskwarrior-json tasks and prints them to the console.
This command like the others in task-client uses your .taskrc file by default.`,
	Run: func(cmd *cobra.Command, args []string) {
		conn, settings := taskc.SimpleConn(Settings)
		taskc.CheckCreds(settings)
		taskc.Pull(conn, settings.Creds)
		resp := taskc.Recv(conn)
		log.Printf("%s", resp)
		out := getJSONTasks(resp)
		for _, item := range out {
			fmt.Println(item)
		}
	},
}
