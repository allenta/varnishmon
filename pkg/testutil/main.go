package testutil

import (
	"allenta.com/varnishmon/pkg/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewConfig(tl zerolog.TestingLog, cfg ...interface{}) *config.Config {
	tl.Helper()

	var key string
	vpr := viper.New()
	for i, arg := range cfg {
		switch i % 2 {
		case 0:
			key = arg.(string) //nolint:forcetypeassert
		case 1:
			vpr.Set(key, arg)
		}
	}

	if !vpr.IsSet("global.loglevel") {
		vpr.Set("global.loglevel", "info")
	}
	loglevel, err := zerolog.ParseLevel(vpr.GetString("global.loglevel"))
	if err != nil {
		panic(err)
	}

	if !vpr.IsSet("scraper.varnishstat") {
		vpr.Set("scraper.varnishstat", "/dev/null")
	}

	return config.NewConfig(
		config.NewTestLogger(tl, loglevel),
		vpr)
}
