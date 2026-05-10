package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/smsufyian/sprig/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.NewRootCmd(ctx).Execute(); err != nil {
		os.Exit(1)
	}
}
