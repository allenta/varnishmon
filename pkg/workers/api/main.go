package api

import (
	"allenta.com/varnishmon/pkg/config"
)

type Application interface {
	Cfg() *config.Config
}
