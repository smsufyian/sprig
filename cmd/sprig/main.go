package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:     "sprig",
		Short:   "Isolated virtual spaces for development, testing, and CI",
		Long:    `sprig creates fully isolated environments for developers, AI agents, and CI pipelines.`,
		Version: "0.1.0",
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
