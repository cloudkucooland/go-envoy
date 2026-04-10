package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudkucooland/go-envoy"
	"github.com/urfave/cli/v3"
)

var envlog *slog.Logger

func main() {
	cmd := &cli.Command{
		Name:      "envoy",
		Version:   "v0.2.0",
		Copyright: "(c) 2026 Scot Bontrager",
		Usage:     "fetch envoy data",
		UsageText: "envoy",

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
				Value:   "/tmp/envoy.jwt",
			},
		},

		Commands: []*cli.Command{
			home,
			now,
			prod,
			// today,
			info,
			inventory,
		},
	}

	envlog = slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.Run(ctx, os.Args); err != nil {
		envlog.Error("error", "error", err.Error())
	}
}

func setupenvoy(ctx context.Context, cmd *cli.Command) (*envoy.Envoy, error) {
	host, err := envoy.Discover(ctx)
	if err != nil {
		return nil, err
	}
	envlog.InfoContext(ctx, "discovered host", "value", host)

	// get cached token
	var token string
	b, err := os.ReadFile(cmd.String("token"))
	switch {
	case err == nil:
		token = string(b)
	case os.IsNotExist(os.ErrNotExist):
		envlog.InfoContext(ctx, "token cache does not exist, refreshing", "value", err)

		token, err = envoy.GetToken(ctx, cmd.String("username"), cmd.String("password"), cmd.String("serial"))
		if err != nil {
			return nil, err
		}

		newfile, err := os.OpenFile(cmd.String("token"), os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}
		defer newfile.Close()

		newfile.WriteString(token)
	default:
		envlog.WarnContext(ctx, "token cache does not exist", "file", cmd.String("token"), "value", err)
		return nil, err
	}

	env := envoy.New(host, string(token))
	return env, nil
}

var now = &cli.Command{
	Name:  "now",
	Usage: "current stats",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
		if err != nil {
			return err
		}

		s, err := e.Production(ctx)
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
		return nil
	},
}

var home = &cli.Command{
	Name:  "home",
	Usage: "system status",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
		if err != nil {
			return err
		}

		s, err := e.Home(ctx)
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
		return nil
	},
}

var prod = &cli.Command{
	Name:  "prod",
	Usage: "system production",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
		if err != nil {
			return err
		}

		s, err := e.Production(ctx)
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
		return nil
	},
}

var inventory = &cli.Command{
	Name:  "inventory",
	Usage: "system inventory",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
		if err != nil {
			return err
		}

		s, err := e.Inventory(ctx)
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
		return nil
	},
}

var info = &cli.Command{
	Name:  "info",
	Usage: "system info",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
		if err != nil {
			return err
		}

		s, err := e.Info(ctx)
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
		return nil
		/* fmt.Println("Serial Number: ", s.Device.Sn)
		fmt.Println("Part Number: ", s.Device.Pn)
		fmt.Println("Software Version: ", s.Device.Software) */
	},
}

/*
var today = &cli.Command{
	Name:  "today",
	Usage: "stats for today",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
        if err != nil {
			return err
		}

		s, err := e.Today(ctx)
		if err != nil {
			panic(err)
		}
		i, _ := json.MarshalIndent(s, "", " ")
		fmt.Printf("%s\n", i)
        return nil
		// fmt.Printf("Production: %2.2fkWh\tConsumption: %2.2fkWh\tNet: %2.2fkWh\n", p/1000, c/1000, net/1000)
    },
} */

/*
	switch command {
	case "now":
		p, c, net, err := e.Now()
		if err != nil {
			panic(err)
		}
		max, err := e.SystemMax()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Production: %2.2fW / %dW\tConsumption: %2.2fW\tNet: %2.2fW\n", p, max, c, 0 - net)
	}
} */
