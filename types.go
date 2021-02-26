package envoy

import "encoding/xml"

// from the production endpoint
type production struct {
	Production  []entry `json:"production"`
	Consumption []entry `json:"consumption"`
	Storage     []entry `json:"storage"`
}

type entry struct {
	Type             string  `json:"type"`
	ActiveCount      int     `json:"activeCount"`
	MeasurementType  string  `json:"measurementType"`
	ReadingTime      int     `json:"readingTime"`
	WNow             float64 `json:"wNow,omitempty"`
	WhLifetime       float64 `json:"whLifetime,omitempty"`
	WhNow            float64 `json:"whNow,omitempty"`
	VarhLeadLifetime float64 `json:"varhLeadLifetime,omitempty"`
	VarhLagLifetime  float64 `json:"varhLagLifetime,omitempty"`
	VahLifetime      float64 `json:"vahLifetime,omitempty"`
	RmsCurrent       float64 `json:"rmsCurrent,omitempty"`
	RmsVoltage       float64 `json:"rmsVoltage,omitempty"`
	ReactPwr         float64 `json:"reactPwr,omitempty"`
	ApprntPwr        float64 `json:"apprntPwr,omitempty"`
	PwrFactor        float64 `json:"pwrFactor,omitempty"`
	WhToday          float64 `json:"whToday,omitempty"`
	WhLastSevenDays  float64 `json:"whLastSevenDays,omitempty"`
	VahToday         float64 `json:"vahToday,omitempty"`
	VarhLeadToday    float64 `json:"varhLeadToday,omitempty"`
	VarhLagToday     float64 `json:"varhLagToday,omitempty"`
	State            string  `json:"state,omitempty"`
	Lines            []line  `json:"lines,omitempty"`
}

type line struct {
	WNow             float64 `json:"wNow"`
	WhLifetime       float64 `json:"whLifetime"`
	VarhLeadLifetime float64 `json:"varhLeadLifetime"`
	VarhLagLifetime  float64 `json:"varhLagLifetime"`
	VahLifetime      float64 `json:"vahLifetime"`
	RmsCurrent       float64 `json:"rmsCurrent"`
	RmsVoltage       float64 `json:"rmsVoltage"`
	ReactPwr         float64 `json:"reactPwr"`
	ApprntPwr        float64 `json:"apprntPwr"`
	PwrFactor        float64 `json:"pwrFactor"`
	WhToday          float64 `json:"whToday"`
	WhLastSevenDays  float64 `json:"whLastSevenDays"`
	VahToday         float64 `json:"vahToday"`
	VarhLeadToday    float64 `json:"varhLeadToday"`
	VarhLagToday     float64 `json:"varhLagToday"`
}

// http://192.168.1.223/home.json
type home struct {
	SoftwareBuildEpoch int     `json:"software_build_epoch"`
	IsNonvoy           bool    `json:"is_nonvoy"`
	DbSize             string  `json:"db_size"`
	DbPercentFull      string  `json:"db_percent_full"`
	Timezone           string  `json:"timezone"`
	CurrentDate        string  `json:"current_date"`
	CurrentTime        string  `json:"current_time"`
	Network            homenet `json:"network"`
	Tariff             string  `json:"tariff"`
	Comm               struct {
		Num   int        `json:"num"`
		Level int        `json:"level"`
		Pcu   homenumlev `json:"pcu"`
		Acb   homenumlev `json:"acb"`
		Nsrb  homenumlev `json:"nsrb"`
	} `json:"comm"`
	Alerts       []interface{} `json:"alerts"` // dunno on this one yet
	UpdateStatus string        `json:"update_status"`
}

type homenumlev struct {
	Num   int `json:"num"`
	Level int `json:"level"`
}

type homenet struct {
	WebComm                 bool        `json:"web_comm"`
	EverReportedToEnlighten bool        `json:"ever_reported_to_enlighten"`
	LastEnlightenReportTime int         `json:"last_enlighten_report_time"`
	PrimaryInterface        string      `json:"primary_interface"`
	Interfaces              []homenetif `json:"interfaces"`
}

type homenetif struct {
	SignalStrength    int    `json:"signal_strength"`
	SignalStrengthMax int    `json:"signal_strength_max"`
	Network           bool   `json:"network,omitempty"`
	Type              string `json:"type"`
	Interface         string `json:"interface"`
	Dhcp              bool   `json:"dhcp"`
	IP                string `json:"ip"`
	Carrier           bool   `json:"carrier"`
	Mac               string `json:"mac,omitempty"`
	Supported         bool   `json:"supported,omitempty"`
	Present           bool   `json:"present,omitempty"`
	Configured        bool   `json:"configured,omitempty"`
	Status            string `json:"status,omitempty"`
}

// inventory
type inventory struct {
	Type    string `json:"type"`
	Devices []struct {
		PartNum        string   `json:"part_num"`
		Installed      string   `json:"installed"`
		SerialNum      string   `json:"serial_num"`
		DeviceStatus   []string `json:"device_status"`
		LastRptDate    string   `json:"last_rpt_date"`
		AdminState     int      `json:"admin_state"`
		DevType        int      `json:"dev_type"`
		CreatedDate    string   `json:"created_date"`
		ImgLoadDate    string   `json:"img_load_date"`
		ImgPnumRunning string   `json:"img_pnum_running"`
		Ptpn           string   `json:"ptpn"`
		Chaneid        int      `json:"chaneid"`
		DeviceControl  []struct {
			Gficlearset bool `json:"gficlearset"`
		} `json:"device_control"`
		Producing     bool `json:"producing"`
		Communicating bool `json:"communicating"`
		Provisioned   bool `json:"provisioned"`
		Operating     bool `json:"operating"`
	} `json:"devices"`
}

// for the stream endpoint
type streamEntry struct {
	P  float64 `json:"p"`
	Q  float64 `json:"q"`
	S  float64 `json:"s"`
	V  float64 `json:"v"`
	I  float64 `json:"i"`
	Pf float64 `json:"pf"`
	F  float64 `json:"f"`
}

type streamSet struct {
	A streamEntry `json:"ph-a"`
	B streamEntry `json:"ph-b"`
	C streamEntry `json:"ph-c"`
}

type stream struct {
	Production       streamSet `json:"production"`
	NetConsumption   streamSet `json:"net-consumption"`
	TotalConsumption streamSet `json:"total-consumption"`
}

type EnvoyInfo struct {
	XMLName xml.Name `xml:"envoy_info"`
	Text    string   `xml:",chardata"`
	Time    string   `xml:"time"`
	Device  struct {
		Text     string `xml:",chardata"`
		Sn       string `xml:"sn"`
		Pn       string `xml:"pn"`
		Software string `xml:"software"`
		Euaid    string `xml:"euaid"`
		Seqnum   string `xml:"seqnum"`
		Apiver   string `xml:"apiver"`
		Imeter   string `xml:"imeter"`
	} `xml:"device"`
	Package []struct {
		Text    string `xml:",chardata"`
		Name    string `xml:"name,attr"`
		Pn      string `xml:"pn"`
		Version string `xml:"version"`
		Build   string `xml:"build"`
	} `xml:"package"`
	BuildInfo struct {
		Text         string `xml:",chardata"`
		BuildTimeGmt string `xml:"build_time_gmt"`
		BuildID      string `xml:"build_id"`
	} `xml:"build_info"`
}
