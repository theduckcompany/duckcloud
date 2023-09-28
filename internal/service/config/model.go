package config

import (
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type ConfigKey string

const (
	HostName           ConfigKey = "host.name"
	HostsTrusted                 = "hosts.trusted"
	DevModeEnabled               = "dev_mode.enabled"
	HTTPPort                     = "http.port"
	HTTPAddrs                    = "http.addrs"
	FSDataFolder                 = "fs.data_folder"
	TLSEnabled                   = "tls.enabled"
	SSLCertificatePath           = "tls.ssl.certificate"
	SSLPrivateKeyPath            = "tls.ssl.private_key"
)

type BootstrapCmd struct {
	HostName string
	Port     int
}

func (t BootstrapCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.HostName, v.Required, is.Domain),
		v.Field(&t.Port, v.Required, is.Port),
	)
}
