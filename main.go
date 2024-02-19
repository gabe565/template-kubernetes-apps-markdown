package main

import (
	"os"

	"github.com/gabe565/template-kubernetes-apps-markdown/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
