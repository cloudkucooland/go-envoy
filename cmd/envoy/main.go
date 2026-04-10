package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/cloudkucooland/go-envoy"
	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
    "github.com/fatih/color"
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
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "output raw JSON",
			},
		},

		Commands: []*cli.Command{
			home,
			now,
			prod,
			// today,
			info,
			inventory,
			status,
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
	if !cmd.Root().Bool("json") {
		envlog.InfoContext(ctx, "Discovering Envoy ...")
	}
	host, err := envoy.Discover(ctx)
	if err != nil {
		return nil, err
	}
	if !cmd.Root().Bool("json") {
		envlog.InfoContext(ctx, "discovered host", "value", host)
	}

	// get cached token
	var token string
	b, err := os.ReadFile(cmd.String("token"))
	switch {
	case err == nil:
		token = string(b)
	case errors.Is(err, os.ErrNotExist):
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
			return err
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
			return err
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

		data, err := e.Production(ctx)
		if err != nil {
			return err
		}

		return formatOutput(cmd, data, func() {
			printFriendlyProduction(ctx, data)
		})
	},
}

func printFriendlyProduction(ctx context.Context, data *envoy.Production) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "\nTYPE\tINSTANT (W)\tTODAY (kWh)\tLIFETIME (MWh)")
	fmt.Fprintln(w, "----\t-----------\t-----------\t--------------")

	for _, p := range data.Production {
		if p.Type == "inverters" {
			fmt.Fprintf(w, "Solar (Inverters)\t%.2f W\t-\t%.2f MWh\n",
				p.WNow, float64(p.WhLifetime)/1000000)
		} else {
			fmt.Fprintf(w, "Solar (Meter)\t%.2f W\t%.2f kWh\t%.2f MWh\n",
				p.WNow, p.WhToday/1000, p.WhLifetime/1000000)
		}
	}

	for _, c := range data.Consumption {
		label := "Consumption"
		if c.MeasurementType == "net-consumption" {
			if c.WNow > 0 {
				label = "Net Grid (Import)"
			} else {
				label = "Net Grid (Export)"
			}
			fmt.Fprintf(w, "%s\t%.2f W\t%.2f kWh\t%.2f MWh\n",
				label, c.WNow, c.WhToday/1000, c.WhLifetime/1000000)
		}
	}

	w.Flush()

	for _, s := range data.Storage {
		if s.ActiveCount > 0 {
			fmt.Printf("\n Storage: %s (%d active)\n", s.State, s.ActiveCount)
		}
	}
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
			return err
		}

		return formatOutput(cmd, s, func() {
			printFriendlyInventory(ctx, s)
		})
	},
}

func printFriendlyInventory(ctx context.Context, inventory *[]envoy.Inventory) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	var total, producing, communicating int

	for _, group := range *inventory {
		if len(group.Devices) == 0 {
			continue
		}

		fmt.Fprintf(w, "\n--- %s (%d devices) ---\n", group.Type, len(group.Devices))
		fmt.Fprintln(w, "SERIAL\tSTATE\tPRODUCING\tCOMMUNICATING\tLAST SEEN\tFW VERSION")
		fmt.Fprintln(w, "------\t-----\t---------\t-------------\t---------\t----------")

		for _, d := range group.Devices {
			total++
			if d.Producing {
				producing++
			}
			if d.Communicating {
				communicating++
			}

			lastSeen := "Never"
			if d.LastRptDate != "" {
				ts, err := strconv.ParseInt(d.LastRptDate, 10, 64)
				if err == nil {
					lastSeen = humanize.Time(time.Unix(ts, 0))
				}
			}

			status := "OK"
			if !d.Operating {
				status = "ERR"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				d.SerialNum,
				colorStatus(status),
				colorBool(d.Producing, "Yes", "NO"),
				colorBool(d.Communicating, "Yes", "LOST"),
				lastSeen,
				d.ImgPnumRunning,
			)
		}
		w.Flush()
		fmt.Printf("\n--- Inventory Health Summary ---\n")
		fmt.Printf("Total Devices:  %d\n", total)
		fmt.Printf("Producing:      %d / %d", producing, total)
		if producing < total {
			fmt.Printf(" ⚠️  (%d DOWN)", total-producing)
		}
		fmt.Printf("\nCommunicating:  %d / %d\n", communicating, total)
	}
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
			return err
		}

		return formatOutput(cmd, s, func() {
			fmt.Printf("--- Envoy Information ---\n")
			fmt.Printf("%-18s %s\n", "Serial:", s.Device.Sn)
			fmt.Printf("%-18s %s\n", "Part Number:", s.Device.Pn)
			fmt.Printf("%-18s %s\n", "Software:", s.Device.Software)
		})
	},
}

func formatOutput(cmd *cli.Command, data any, friendly func()) error {
	if cmd.Root().Bool("json") {
		i, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(i))
		return nil
	}

	friendly()
	return nil
}

var status = &cli.Command{
	Name:  "status",
	Usage: "Combined system dashboard",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		e, err := setupenvoy(ctx, cmd)
		if err != nil {
			return err
		}

		var prod *envoy.Production
		var inv *[]envoy.Inventory
		var g errgroup.Group

		g.Go(func() error {
			var err error
			prod, err = e.Production(ctx)
			return err
		})

		g.Go(func() error {
			var err error
			inv, err = e.Inventory(ctx)
			return err
		})

		if err := g.Wait(); err != nil {
			return err
		}

		printFriendlyProduction(ctx, prod)
		printFriendlyInventory(ctx, inv)
		if len(prod.Production) > 0 {
			var invW, meterW float64
			for _, p := range prod.Production {
				if p.Type == "inverters" {
					invW = p.WNow
				}
				if p.Type == "eim" {
					meterW = p.WNow
				}
			}

			if invW > 0 && meterW < 50 {
				fmt.Println("\n🌙 [System Note] Sun has set. Inverter data may be cached/stale.")
			}
		}
		return nil
	},
}

func colorStatus(status string) string {
	if status == "OK" {
		return color.GreenString(status)
	}
	return color.RedString(status)
}

func colorBool(val bool, yesLabel, noLabel string) string {
	if val {
		return color.GreenString(yesLabel)
	}
	return color.RedString(noLabel)
}
