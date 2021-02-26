package envoy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Envoy struct {
	host string
}

func New(host string) (*Envoy, error) {
	e := Envoy{
		host: host,
	}
	return &e, nil
}

func (e *Envoy) Production() (*production, error) {
	url := fmt.Sprintf("http://%s/production.json?details=1", e.host)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var d production
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (e *Envoy) Home() (*home, error) {
	url := fmt.Sprintf("http://%s/home.json", e.host)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var d home
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// http://192.168.1.223/inventory.json?deleted=1
func (e *Envoy) Inventory() (*[]inventory, error) {
	url := fmt.Sprintf("http://%s/inventory.json", e.host)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var d []inventory
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (e *Envoy) Now() (float64, error) {
	s, err := e.Production()
	if err != nil {
		return 0, err
	}
	totprod := 0.0
	for _, v := range s.Production {
		lineprod := 0.0
		for _, lv := range v.Lines {
			lineprod += lv.WNow
		}
		totprod += lineprod
	}
	return totprod, nil
}

func (e *Envoy) Today() (float64, error) {
	s, err := e.Production()
	if err != nil {
		return 0, err
	}
	totprod := 0.0
	for _, v := range s.Production {
		lineprod := 0.0
		for _, lv := range v.Lines {
			lineprod += lv.WhToday
		}
		totprod += lineprod
	}
	return totprod, nil
}

// requested, but errors
// http://192.168.1.223/ivp/meters

// frequently "", sometimes { "tunnel_open": false }
// http://192.168.1.223/admin/lib/dba.json
