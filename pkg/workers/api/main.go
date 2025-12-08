package api //nolint:revive,nolintlint

import (
	"github.com/allenta/varnishmon/pkg/config"
)

type Application interface {
	Cfg() *config.Config
}
