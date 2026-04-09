package main

import (
	"context"
	"os"
	"time"

	"github.com/cloudkucooland/go-envoy"
	"github.com/urfave/cli/v3"
)

var env *envoy.Envoy

type envoydata struct {
	DeviceID string
	Alias    string
	P        *envoy.Production
}

func setupenvoy(ctx context.Context, cmd *cli.Command) error {
	host, err := envoy.Discover(ctx)
	if err != nil {
		return err
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
			return err
		}

		newfile, err := os.OpenFile(cmd.String("token"), os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer newfile.Close()

		newfile.WriteString(token)
	default:
		envlog.WarnContext(ctx, "token cache does not exist", "file", cmd.String("token"), "value", err)
		return err
	}

	env = envoy.New(host, string(token))
	return nil
}

func query(ctx context.Context, results chan envoydata) error {
	pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	p, err := env.Production(pctx)
	if err != nil {
		return err
	}

	e := envoydata{
		DeviceID: "Envoy",
		Alias:    "Envoy",
		P:        p,
	}

	envlog.Info("Query successful, sending to channel", "points", len(p.Production))
	results <- e
	return nil
}
