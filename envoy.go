package envoy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Envoy struct {
	client *http.Client
	host   string
	token  string
}

func New(host, token string) *Envoy {
	return &Envoy{
		host:   host,
		token:  strings.TrimSpace(token),
		client: newClient(),
	}
}

func doRequest[T any](ctx context.Context, e *Envoy, path string) (*T, error) {
	url := fmt.Sprintf("https://%s/%s", e.host, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+e.token)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var target T
	if strings.HasSuffix(path, ".xml") {
		err = xml.Unmarshal(body, &target)
	} else {
		err = json.Unmarshal(body, &target)
	}

	if err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}

	return &target, nil
}

func (e *Envoy) Production(ctx context.Context) (*Production, error) {
	return doRequest[Production](ctx, e, "production.json?details=1")
}

func (e *Envoy) Home(ctx context.Context) (*home, error) {
	return doRequest[home](ctx, e, "home.json")
}

func (e *Envoy) Inventory(ctx context.Context) (*[]Inventory, error) {
	return doRequest[[]Inventory](ctx, e, "inventory.json")
}

func (e *Envoy) Info(ctx context.Context) (*EnvoyInfo, error) {
	return doRequest[EnvoyInfo](ctx, e, "info.xml")
}

func (e *Envoy) Inverters(ctx context.Context) (*[]Inverter, error) {
	return doRequest[[]Inverter](ctx, e, "api/v1/production/inverters")
}

func (e *Envoy) Now(ctx context.Context) (tp, tc, net float64, err error) {
	s, err := e.Production(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	for _, v := range s.Production {
		if v.MeasurementType == "production" {
			tp = v.WNow
		}
	}
	for _, v := range s.Consumption {
		switch v.MeasurementType {
		case "total-consumption":
			tc = v.WNow
		case "net-consumption":
			net = v.WNow
		}
	}
	return
}

func newClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:  false,
			MaxIdleConns:       5,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: true,
		},
		Timeout: 30 * time.Second,
	}
}
