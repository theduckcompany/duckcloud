package config

import (
	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type ConfigKey string

const (
	HostName = "host_name"
	Port     = "port"
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
