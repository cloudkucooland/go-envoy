package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudkucooland/go-envoy"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/urfave/cli/v3"
)

var (
	client   influxdb2.Client
	writeAPI api.WriteAPI
)

func setupdb(ctx context.Context, cmd *cli.Command) error {
	host := os.Getenv("INFLUX_HOST")
	token := os.Getenv("INFLUX_TOKEN")
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")

	envlog.InfoContext(ctx, "Setting up InfluxDB V2 Client", "host", host, "bucket", bucket)

	client = influxdb2.NewClient(host, token)

	ok, err := client.Health(ctx)
	if err != nil || ok.Status != "pass" {
		return fmt.Errorf("influxdb health check failed: %w", err)
	}

	writeAPI = client.WriteAPI(org, bucket)

	go func() {
		for err := range writeAPI.Errors() {
			envlog.Error("influx async write error", "err", err)
		}
	}()

	return nil
}

func startDBWriter(ctx context.Context, r <-chan envoydata) {
	envlog.Info("DB Writer started (V2 Downgrade)")

	for {
		select {
		case <-ctx.Done():
			envlog.Info("DB Writer shutting down, flushing points...")
			writeAPI.Flush()
			client.Close()
			return
		case v, ok := <-r:
			if !ok {
				return
			}

			processLines(v.P.Consumption, v.DeviceID)
			processLines(v.P.Production, v.DeviceID)
		}
	}
}

func processLines(collections []envoy.Entry, deviceID string) {
	for _, i := range collections {
		if i.Type != "eim" {
			continue
		}
		for j, k := range i.Lines {
			power := k.WNow * 1000
			if power < 0 {
				power = 0
			}

			p := influxdb2.NewPoint("emeter",
				map[string]string{
					"device": deviceID,
					"alias":  fmt.Sprintf("%s-L%d", i.MeasurementType, j+1),
					"phase":  fmt.Sprintf("L%d", j+1),
				},
				map[string]interface{}{
					"slot":      uint64(j),
					"VoltageMV": uint64(k.RmsVoltage * 1000),
					"CurrentMA": uint64(k.RmsCurrent * 1000),
					"PowerMW":   uint64(power),
				},
				time.Now())

			writeAPI.WritePoint(p)
		}
	}
}
