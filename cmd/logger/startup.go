package main

import (
	"context"
	"time"

	"github.com/urfave/cli/v3"
)

var timeout = 2 * time.Second

var startup = &cli.Command{
	Name:  "startup",
	Usage: "start the daemon",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := setupenvoy(ctx, cmd); err != nil {
			return err
		}

		if err := setupdb(ctx, cmd); err != nil {
			return err
		}
		results := make(chan envoydata, 2)
		go startDBWriter(ctx, results)

		pollrate := cmd.Int("pollrate")
		if pollrate <= 0 {
			pollrate = 2
		}

		tt := cmd.Int("timeout")
		if tt > 0 {
			timeout = time.Duration(tt) * time.Second
		}

		ticker := time.NewTicker(time.Duration(pollrate) * time.Second)

		for {
			select {
			case <-ctx.Done():
				close(results)
				return nil
			case <-ticker.C:
				// drop any lingering attempts before the next tick
				runtime := time.Duration(pollrate)*time.Second - timeout
				runCtx, cancel := context.WithTimeout(ctx, runtime)
				err := query(runCtx, results)
				cancel()
				if err != nil {
					envlog.Error("query error", "err", err)
				}
			}
		}
		return nil
	},
}
