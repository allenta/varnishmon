package api

import (
	"github.com/allenta/varnishmon/pkg/config"
)

type Application interface {
	Cfg() *config.Config
}
