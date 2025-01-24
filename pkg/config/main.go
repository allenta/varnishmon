package config

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var (
	version     string
	revision    string //nolint:gochecknoglobals
	environment string //nolint:gochecknoglobals
)

func Version() string {
	return version
}

func Revision() string {
	return revision
}

func IsDevelopment() bool {
	return environment == "development"
}

type Config struct {
	log     *Logger
	vpr     *viper.Viper
	metrics *Metrics
}

func NewConfig(log *Logger, vpr *viper.Viper) *Config {
	cfg := &Config{
		log:     log,
		vpr:     vpr,
		metrics: NewMetrics(),
	}
	cfg.init()
	return cfg
}

func (cfg *Config) ConfigFileUsed() string {
	return cfg.vpr.ConfigFileUsed()
}

// ----------------------------------------------------------------------------
// GLOBAL
// ----------------------------------------------------------------------------

func (cfg *Config) Logfile() string {
	return cfg.vpr.GetString("global.logfile")
}

func (cfg *Config) Loglevel() zerolog.Level {
	return cfg.vpr.Get("global.loglevel").(zerolog.Level)
}

func (cfg *Config) LogCaller() bool {
	return cfg.vpr.GetBool("global.log-caller")
}

func (cfg *Config) LogJSON() bool {
	return cfg.vpr.GetBool("global.log-json")
}

// ----------------------------------------------------------------------------
// DB
// ----------------------------------------------------------------------------

func (cfg *Config) DBFile() string {
	return cfg.vpr.GetString("db.file")
}

func (cfg *Config) DBMemoryLimit() int {
	return cfg.vpr.GetInt("db.memory-limit")
}

func (cfg *Config) DBThreads() int {
	return cfg.vpr.GetInt("db.threads")
}

func (cfg *Config) DBTempDirectory() string {
	return cfg.vpr.GetString("db.temp-directory")
}

func (cfg *Config) DBMaxTempDirectorySize() int {
	return cfg.vpr.GetInt("db.max-temp-directory-size")
}

// ----------------------------------------------------------------------------
// SCRAPER
// ----------------------------------------------------------------------------

func (cfg *Config) ScraperEnabled() bool {
	return cfg.vpr.GetBool("scraper.enabled")
}

func (cfg *Config) ScraperPeriod() time.Duration {
	return cfg.vpr.GetDuration("scraper.period")
}

func (cfg *Config) ScraperTimeout() time.Duration {
	return cfg.vpr.GetDuration("scraper.timeout")
}

func (cfg *Config) ScraperCommand() []string {
	return cfg.vpr.GetStringSlice("scraper.command")
}

// ----------------------------------------------------------------------------
// API
// ----------------------------------------------------------------------------

func (cfg *Config) APIEnabled() bool {
	return cfg.vpr.GetBool("api.enabled")
}

func (cfg *Config) APIWorkers() int {
	return cfg.vpr.GetInt("api.workers")
}

func (cfg *Config) APIListenIP() string {
	return cfg.vpr.GetString("api.listen-ip")
}

func (cfg *Config) APIListenPort() int {
	return cfg.vpr.GetInt("api.listen-port")
}

func (cfg *Config) APIBasicAuthUsername() string {
	return cfg.vpr.GetString("api.basic-auth.username")
}

func (cfg *Config) APIBasicAuthPassword() string {
	return cfg.vpr.GetString("api.basic-auth.password")
}

func (cfg *Config) APITLSCertfile() string {
	return cfg.vpr.GetString("api.tls.certfile")
}

func (cfg *Config) APITLSKeyfile() string {
	return cfg.vpr.GetString("api.tls.keyfile")
}

func (cfg *Config) APIBacklog() int {
	return cfg.vpr.GetInt("api.backlog")
}

func (cfg *Config) APIConcurrency() int {
	return cfg.vpr.GetInt("api.concurrency")
}

func (cfg *Config) APIReadBufferSize() int {
	return cfg.vpr.GetInt("api.read-buffer-size")
}

func (cfg *Config) APIWriteBufferSize() int {
	return cfg.vpr.GetInt("api.write-buffer-size")
}

func (cfg *Config) APIMaxRequestBodySize() int {
	return cfg.vpr.GetInt("api.max-request-body-size")
}

func (cfg *Config) APIReadTimeout() time.Duration {
	return cfg.vpr.GetDuration("api.read-timeout")
}

func (cfg *Config) APIWriteTimeout() time.Duration {
	return cfg.vpr.GetDuration("api.write-timeout")
}

func (cfg *Config) APIIdleTimeout() time.Duration {
	return cfg.vpr.GetDuration("api.idle-timeout")
}

func (cfg *Config) APITCPKeepalive() bool {
	return cfg.vpr.GetBool("api.tcp-keepalive")
}

func (cfg *Config) APITCPKeepalivePeriod() time.Duration {
	return cfg.vpr.GetDuration("api.tcp-keepalive-period")
}
