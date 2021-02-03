package main

import (
	"fmt"
	"os"

	"github.com/zoumo/make-rules/cmd/make-rules/app"
)

func main() {
	command := app.NewRootCommand()
	if err := command.Execute(); err != nil {
		fmt.Printf("run command error: %v\n", err)
		os.Exit(1)
	}
}
