package envoy

import (
    "io"
    "fmt"
    "net/http"
    "encoding/json"
)

type Envoy struct {
    ip string
}

func New(ip string) (*Envoy, error) {
    e := Envoy{
        ip: ip,
    }
    return &e, nil
}

func (e *Envoy) Pull() (*EnvoyData, error) {
    url := fmt.Sprintf("http://%s/production.json?details=1", e.ip)

    resp, err := http.Get(url)
    if err != nil {
	   return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)

    var d EnvoyData
    err = json.Unmarshal(body, &d)
    if err != nil {
	   return nil, err
    }

    return &d, nil
}

func (e *Envoy) Home() (string, error) {
    url := fmt.Sprintf("http://%s/home.json", e.ip)

    resp, err := http.Get(url)
    if err != nil {
	   return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)

    return body.String(), nil
}

// http://192.168.1.223/inventory.json?deleted=1
func (e *Envoy) Inventory() (string, error) {
    url := fmt.Sprintf("http://%s/inventory.json", e.ip)

    resp, err := http.Get(url)
    if err != nil {
	   return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)

    return body.String(), nil
}

// requested, but errors
// http://192.168.1.223/ivp/meters

// frequently "", sometimes { "tunnel_open": false }
// http://192.168.1.223/admin/lib/dba.json

