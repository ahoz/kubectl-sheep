package main

import (
	"os"

	"github.com/ahoz/kubectl-sheep/internal/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
