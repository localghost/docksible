package main

import (
	"github.com/localghost/docksible/cmd"
	"os"
)

func main() {
	if err := cmd.CreateRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
