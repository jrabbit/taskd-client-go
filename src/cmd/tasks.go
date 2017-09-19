package cmd

import (
    "bufio"
    "bytes"
    // "github.com/davecgh/go-spew/spew"
    "fmt"
    "github.com/spf13/cobra"
    "strings"
    "taskc"
)

func init() {
    RootCmd.AddCommand(tasksCmd)
}

func getJSONTasks(resp []byte) []string {
    buff := bytes.NewBuffer(resp)
    scanner := bufio.NewScanner(buff)
    var jsonTasks []string
    headers := make(map[string]string)
    for scanner.Scan() {
        text := scanner.Text()
        if len(text) < 1 {
            continue
        } else if strings.Split(text, "")[0] == "{" {
            jsonTasks = append(jsonTasks, scanner.Text())
        } else if strings.Contains(text, ":") {
            xyz := strings.Split(text, ":")
            headers[xyz[0]] = strings.TrimLeft(xyz[1], " ")
        } else if len(text) == 36 {
        }
    }
    return jsonTasks
}

var tasksCmd = &cobra.Command{
    Use:   "tasks",
    Short: "grab json tasks for your user from taskd server",
    Long:  `This subcommand downloads all your taskwarrior-json tasks and prints them to the console`,
    Run: func(cmd *cobra.Command, args []string) {
        rc := taskc.ReadRC()
        conn := taskc.Connect(rc)
        taskc.Pull(conn, rc["taskd.credentials"])
        resp := taskc.Recv(conn)
        // out := taskc.ParseResponse(resp)
        out := getJSONTasks(resp)
        // spew.Dump(out)
        for _, item := range out {
            fmt.Println(item)
        }
    },
}
