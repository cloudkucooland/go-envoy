package envoy

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	dac "github.com/xinsnake/go-http-digest-auth-client"
	"io"
	"log"
	"net/http"
	// "net"
	"time"
)

var elogger envoylogger = log.Default()

type envoylogger interface {
	Println(...interface{})
	Printf(string, ...interface{})
}

func SetLogger(l envoylogger) {
	elogger = l
}

type Envoy struct {
	client   *http.Client
	host     string
	password string
}

func New(host string) (*Envoy, error) {
	var err error
	e := Envoy{}

	// if no host set, try discovery
	if host == "" {
		host, err = Discover()
	}
	if err != nil {
		elogger.Println(err.Error())
		return &e, err
	}
	if host == "" {
		elogger.Println("auto-discovery failed")
		return &e, err
	}

	e.host = host
	e.client = newClient()

	return &e, nil
}

func (e *Envoy) Host() string {
	return e.host
}

func (e *Envoy) Rediscover() error {
	var err error
	e.host, err = Discover()
	return err
}

func (e *Envoy) Close() {
	e.host = ""
	e.password = ""
	e.client.CloseIdleConnections()
	e.client = nil
}

func (e *Envoy) Production() (*production, error) {
	url := fmt.Sprintf("https://%s/production.json?details=1", e.host)

	resp, err := e.client.Get(url)
	if err != nil {
		elogger.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		elogger.Println(err.Error())
		return nil, err
	}

	var d production
	if err := json.Unmarshal(body, &d); err != nil {
		elogger.Println(err.Error())
		elogger.Println(string(body))
		return nil, err
	}

	return &d, nil
}

func (e *Envoy) Home() (*home, error) {
	url := fmt.Sprintf("https://%s/home.json", e.host)

	resp, err := e.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d home
	if err := json.Unmarshal(body, &d); err != nil {
		elogger.Println(err.Error())
		elogger.Println(string(body))
		return nil, err
	}

	return &d, nil
}

// http://envoy.local/inventory.json?deleted=1
func (e *Envoy) Inventory() (*[]inventory, error) {
	url := fmt.Sprintf("https://%s/inventory.json", e.host)

	resp, err := e.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var d []inventory
	if err = json.Unmarshal(body, &d); err != nil {
		elogger.Println(err.Error())
		elogger.Println(string(body))
		return nil, err
	}

	return &d, nil
}

func (e *Envoy) Info() (*EnvoyInfo, error) {
	url := fmt.Sprintf("https://%s/info.xml", e.host)

	resp, err := e.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var i EnvoyInfo
	if err := xml.Unmarshal(body, &i); err != nil {
		elogger.Println(err.Error())
		elogger.Println(string(body))
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

	url := fmt.Sprintf("https://%s/api/v1/production/inverters", e.host)
	req := dac.NewRequest("envoy", e.password, "GET", url, "")

	resp, err := req.Execute()
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var i []Inverter
	if err = json.Unmarshal(body, &i); err != nil {
		elogger.Println(err.Error())
		elogger.Println(string(body))
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

func newClient() *http.Client {
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{
		ResponseHeaderTimeout: time.Duration(3) * time.Second,
		DisableKeepAlives:     true,
		MaxIdleConns:          5,
		IdleConnTimeout:       time.Duration(10) * time.Second,
		DisableCompression:    true,
		TLSClientConfig:       tlsconfig,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(5) * time.Second,
	}
	return client
}

// requested, but errors
// http://envoy.local/ivp/meters

// frequently "", sometimes { "tunnel_open": false }
// http://envoy.local/admin/lib/dba.json

// slightly different values than /production
// http://envoy.local/api/v1/production
