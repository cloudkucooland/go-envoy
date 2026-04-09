package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

var envlog *slog.Logger

func main() {
	cmd := &cli.Command{
		Name:      "envoylog",
		Version:   "v0.5.0",
		Copyright: "(c) 2026 Scot Bontrager",
		Usage:     "log envoy data to influxdb",
		UsageText: "envoylog",

		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "serial",
				Aliases: []string{"s"},
				Value:   "1234",
			},
			&cli.StringFlag{
				Name:    "username",
				Aliases: []string{"u"},
				Value:   "me@example.com",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Value:   "secret",
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Value:   "/data/envoy.jwt",
			},
			&cli.IntFlag{
				Name:    "pollrate",
				Aliases: []string{"r"},
				Value:   30,
			},
		},

		Commands: []*cli.Command{
			startup,
		},
	}

	envlog = slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.Run(ctx, os.Args); err != nil {
		envlog.Error("error", "error", err.Error())
	}
}
