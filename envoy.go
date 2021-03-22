package envoy

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	dac "github.com/xinsnake/go-http-digest-auth-client"
	"io"
	"net/http"
)

type Envoy struct {
	host     string
	password string
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

// http://envoy.local/inventory.json?deleted=1
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

func (e *Envoy) Info() (*EnvoyInfo, error) {
	url := fmt.Sprintf("http://%s/info.xml", e.host)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var i EnvoyInfo
	err = xml.Unmarshal(body, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (e *Envoy) Now() (float64, float64, float64, error) {
	s, err := e.Production()
	if err != nil {
		return 0.0, 0.0, 0.0, err
	}
	tp := 0.0
	for _, v := range s.Production {
		if v.MeasurementType == "production" {
			tp = v.WNow
		}
	}
	tc := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "total-consumption" {
			tc = v.WNow
		}
	}
	net := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "net-consumption" {
			net = v.WNow
		}
	}
	return tp, tc, net, nil
}

func (e *Envoy) Today() (float64, float64, float64, error) {
	s, err := e.Production()
	if err != nil {
		return 0.0, 0.0, 0.0, err
	}
	tp := 0.0
	for _, v := range s.Production {
		if v.MeasurementType == "production" {
			tp = v.WhToday
		}
	}
	tc := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "total-consumption" {
			tc = v.WhToday
		}
	}
	tnp := 0.0
	for _, v := range s.Consumption {
		if v.MeasurementType == "net-consumption" {
			tnp = v.WhToday
		}
	}
	return tp, tc, tnp, nil
}

func (e *Envoy) Inverters() (*[]Inverter, error) {
	if e.password == "" {
		e.autoPassword()
	}

	url := fmt.Sprintf("http://%s/api/v1/production/inverters", e.host)
	req := dac.NewRequest("envoy", e.password, "GET", url, "")

	resp, err := req.Execute()
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)

	var i []Inverter
	err = json.Unmarshal(body, &i)
	if err != nil {
		fmt.Println(string(body))
		return nil, err
	}
	return &i, nil
}

func (e *Envoy) SystemMax() (uint64, error) {
	inverters, err := e.Inverters()
	if err != nil {
		return 0, err
	}
	var max uint64
	for _, v := range *inverters {
		max += uint64(v.MaxReportWatts)
	}
	return max, nil
}

func (e *Envoy) Password(p string) {
	e.password = p
}

func (e *Envoy) autoPassword() error {
	info, err := e.Info()
	if err != nil {
		return err
	}

	e.password = info.Device.Sn[6:12]
	return nil
}

// requested, but errors
// http://envoy.local/ivp/meters

// frequently "", sometimes { "tunnel_open": false }
// http://envoy.local/admin/lib/dba.json

// slightly different values than /production
// http://envoy.local/api/v1/production
