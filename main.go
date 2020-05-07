package main

import (
	"fmt"
	"github.com/jrabbit/taskd-client-go/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
