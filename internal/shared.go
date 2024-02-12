package internal

type NginxConfSvcOp string

const (
	NginxConfSvcOpAdd    NginxConfSvcOp = "add"
	NginxConfSvcOpRemove NginxConfSvcOp = "remove"
)

type NginxConf struct {
	Op      NginxConfSvcOp `json:"op"`
	Host    string         `json:"host"`
	Ip      string         `json:"ip"`
	Port    int            `json:"port"`
	Https   bool           `json:"https"`
	Ssl     bool           `json:"ssl"`
	SslCert string         `json:"ssl_cert"`
	SslKey  string         `json:"ssl_key"`
}
