package main

import (
	"fmt"
	"os"

	"slack/whisk/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n⚠️  %s\n", err.Error())
		os.Exit(1)
	}
}
