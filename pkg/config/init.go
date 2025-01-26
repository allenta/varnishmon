package config

import (
	"math"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/rs/zerolog"
)

func (cfg *Config) init() {
	cfg.initGlobalConfig()
	cfg.initDBConfig()
	cfg.initScraperConfig()
	cfg.initAPIConfig()
}

// ----------------------------------------------------------------------------
// GLOBAL
// ----------------------------------------------------------------------------

func (cfg *Config) initGlobalConfig() {
	cfg.vpr.SetDefault("global.logfile", "")

	cfg.vpr.SetDefault("global.loglevel", "info")
	cfg.checkLoglevel("global.loglevel")

	cfg.vpr.SetDefault("global.log-caller", false)

	cfg.vpr.SetDefault("global.log-json", false)
}

// ----------------------------------------------------------------------------
// DB
// ----------------------------------------------------------------------------

func (cfg *Config) initDBConfig() {
	cfg.vpr.SetDefault("db.file", "")

	cfg.vpr.SetDefault("db.memory-limit", 512)
	cfg.checkInt("db.memory-limit", 1, math.MaxInt32)

	cfg.vpr.SetDefault("db.threads", 1)
	cfg.checkInt("db.threads", 1, math.MaxInt32)

	cfg.vpr.SetDefault("db.temp-directory", cfg.vpr.GetString("db.file")+".tmp")

	cfg.vpr.SetDefault("db.max-temp-directory-size", 128)
	cfg.checkInt("db.max-temp-directory-size", 1, math.MaxInt32)
}

// ----------------------------------------------------------------------------
// SCRAPER
// ----------------------------------------------------------------------------

func (cfg *Config) initScraperConfig() {
	cfg.vpr.SetDefault("scraper.enabled", true)

	if cfg.vpr.GetBool("scraper.enabled") {
		cfg.vpr.SetDefault("scraper.period", 1*time.Minute)
		cfg.checkDuration("scraper.period", 1*time.Second, 24*time.Hour)

		cfg.vpr.SetDefault("scraper.timeout", 5*time.Second)
		cfg.checkDuration("scraper.timeout", 1*time.Second, 10*time.Minute)

		{
			cfg.vpr.SetDefault("scraper.varnishstat", "/usr/bin/varnishstat -1 -j")

			varnishstat := cfg.vpr.GetString("scraper.varnishstat")
			command, err := shellquote.Split(os.ExpandEnv(varnishstat))
			if err != nil {
				cfg.log.Fatal().
					Err(err).
					Str("value", varnishstat).
					Msg("Failed to split 'scraper.varnishstat' command!")
			}
			if len(command) > 0 {
				if info, err := os.Stat(command[0]); os.IsNotExist(err) || info.IsDir() {
					cfg.log.Fatal().
						Err(err).
						Str("value", varnishstat).
						Msgf("'scraper.varnishstat' command not found!")
				}
			} else {
				cfg.log.Fatal().Msg("Empty 'scraper.varnishstat' command!")
			}
			cfg.vpr.Set("scraper.command", command)
		}
	}
}

// ----------------------------------------------------------------------------
// API
// ----------------------------------------------------------------------------

func (cfg *Config) initAPIConfig() {
	cfg.vpr.SetDefault("api.enabled", true)

	if cfg.vpr.GetBool("api.enabled") {
		cfg.vpr.SetDefault("api.workers", runtime.NumCPU())
		cfg.checkInt("api.workers", 1, math.MaxInt32)

		cfg.vpr.SetDefault("api.listen-ip", "127.0.0.1")
		cfg.checkIP("api.listen-ip")

		cfg.vpr.SetDefault("api.listen-port", 6100)
		cfg.checkInt("api.listen-port", 1, 65535)

		cfg.vpr.SetDefault("api.basic-auth.username", "")

		cfg.vpr.SetDefault("api.basic-auth.password", "")

		cfg.vpr.SetDefault("api.tls.certfile", "")
		if cfg.vpr.GetString("api.tls.certfile") != "" {
			cfg.checkFile("api.tls.certfile")
		}

		cfg.vpr.SetDefault("api.tls.keyfile", "")
		if cfg.vpr.GetString("api.tls.keyfile") != "" {
			cfg.checkFile("api.tls.keyfile")
		}

		cfg.vpr.SetDefault("api.backlog", 1024)
		cfg.checkInt("api.backlog", 1, 65536)

		cfg.vpr.SetDefault("api.concurrency", 1024)
		cfg.checkInt("api.concurrency", 1, math.MaxInt32)

		cfg.vpr.SetDefault("api.read-buffer-size", 65536)
		cfg.checkInt("api.read-buffer-size", 1, math.MaxInt32)

		cfg.vpr.SetDefault("api.write-buffer-size", 65536)
		cfg.checkInt("api.write-buffer-size", 1, math.MaxInt32)

		cfg.vpr.SetDefault("api.max-request-body-size", 65536)
		cfg.checkInt("api.max-request-body-size", 1, math.MaxInt32)

		cfg.vpr.SetDefault("api.read-timeout", 1*time.Minute)
		cfg.checkDuration("api.read-timeout", 1*time.Second, 10*time.Minute)

		cfg.vpr.SetDefault("api.write-timeout", 1*time.Minute)
		cfg.checkDuration("api.write-timeout", 1*time.Second, 10*time.Minute)

		cfg.vpr.SetDefault("api.idle-timeout", 2*time.Minute)
		cfg.checkDuration("api.idle-timeout", 1*time.Second, 10*time.Minute)

		cfg.vpr.SetDefault("api.tcp-keepalive", true)

		if cfg.vpr.GetBool("api.tcp-keepalive") {
			cfg.vpr.SetDefault("api.tcp-keepalive-period", 2*time.Minute)
			cfg.checkDuration("api.tcp-keepalive-period", 1*time.Second, 10*time.Minute)
		}
	}
}

// ----------------------------------------------------------------------------
// HELPERS
// ----------------------------------------------------------------------------

func (cfg *Config) checkInt(key string, min, max int) { //nolint:predeclared,revive,unparam
	if value := cfg.vpr.GetInt(key); value < min || value > max {
		cfg.log.Fatal().
			Int("value", value).
			Int("min", min).
			Int("max", max).
			Msgf("'%s' is an invalid integer value", key)
	}
}

func (cfg *Config) checkDuration(key string, min, max time.Duration) { //nolint:predeclared,revive,unparam
	if value := cfg.vpr.GetDuration(key); value < min || value > max {
		cfg.log.Fatal().
			Str("value", value.String()).
			Str("min", min.String()).
			Str("max", max.String()).
			Msgf("'%s' is an invalid duration value", key)
	}
}

func (cfg *Config) checkLoglevel(key string) {
	value := cfg.vpr.GetString(key)
	// Valid options: debug, info, warn, error, fatal, panic, '' (NoLevel).
	if level, err := zerolog.ParseLevel(strings.ToLower(value)); err == nil {
		cfg.vpr.Set(key, level)
	} else {
		cfg.log.Fatal().
			Err(err).
			Str("value", value).
			Msgf("'%s' is an invalid log level value", key)
	}
}

func (cfg *Config) checkIP(key string) {
	value := cfg.vpr.GetString(key)
	if net.ParseIP(value) == nil {
		cfg.log.Fatal().
			Str("value", value).
			Msgf("'%s' is an invalid IP address value", key)
	}
}

func (cfg *Config) checkFile(key string) {
	value := cfg.vpr.GetString(key)
	if info, err := os.Stat(value); os.IsNotExist(err) || info.IsDir() {
		cfg.log.Fatal().
			Err(err).
			Str("value", value).
			Msgf("'%s' is an invalid file value", key)
	}
}
