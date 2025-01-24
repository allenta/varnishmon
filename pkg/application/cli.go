package application

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"allenta.com/varnishmon/pkg/config"
)

var (
	RootCmd = &cobra.Command{ //nolint:gochecknoglobals
		Use:   "varnishmon",
		Short: "varnishmon",
		Long: `varnishmon is a utility inspired by the classic atop tool. It
periodically collects metrics from Varnish Cache / Varnish Enterprise using the
varnishstat utility, stores them in a DuckDB database, and provides a simple
built-in web interface for visualizing the timeseries data.`,
		Version:           config.Version(),
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) { //nolint:revive
			// Providing the root command as a parameter here is required to
			// avoid a circular dependency.
			cfg = boot(cmd.Root())
		},
		Run: func(cmd *cobra.Command, args []string) { //nolint:revive
			NewApplication(cfg).Start()
		},
	}

	cfgFile string         //nolint:gochecknoglobals
	cfg     *config.Config //nolint:gochecknoglobals
)

func Main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Disable 'completion' command.
	RootCmd.CompletionOptions.DisableDefaultCmd = true

	// Customize version template.
	RootCmd.SetVersionTemplate(fmt.Sprintf(
		"varnishmon version {{.Version}} (%s)\n"+
			"Copyright (c) Allenta Consulting S.L.\n", config.Revision()))

	// Global flags. Except for 'config', all other flags are just convenience
	// flags to override configuration settings in a more user-friendly way
	// (i.e., without having to edit the configuration file itself or defining
	// environment variables).
	RootCmd.PersistentFlags().StringVarP(
		&cfgFile, "config", "c", "",
		"set configuration file")
	RootCmd.PersistentFlags().String(
		"logfile", "",
		"set log file (overrides 'global.logfile' setting)")
	RootCmd.PersistentFlags().String(
		"loglevel", "",
		"set log level (overrides 'global.loglevel' setting)")
	RootCmd.PersistentFlags().String(
		"db", "",
		"set DB file (overrides 'db.file' setting)")
	RootCmd.PersistentFlags().Bool(
		"no-scraper", true,
		"disable scraping (overrides 'scraper.enabled' setting)")
	RootCmd.PersistentFlags().String(
		"period", "",
		"set scraping period (overrides 'scraper.period' setting)")
	RootCmd.PersistentFlags().String(
		"varnishstat", "",
		"set location of the 'varnishstat' command (overrides 'scraper.varnishstat' setting)")
	RootCmd.PersistentFlags().Bool(
		"no-api", true,
		"disable API (overrides 'api.enabled' setting)")
	RootCmd.PersistentFlags().String(
		"ip", "",
		"set API listen IP address (overrides 'api.listen-ip' setting)")
	RootCmd.PersistentFlags().String(
		"port", "",
		"set API listen port (overrides 'api.listen-port' setting)")
}

func boot(rootCmd *cobra.Command) *config.Config {
	// Initializations.
	syscall.Umask(0027)

	// Configure logging.
	out := os.Stderr
	log := zerolog.New(out).With().Timestamp().Caller().Logger().
		Level(zerolog.InfoLevel).Output(zerolog.ConsoleWriter{
		Out:        out,
		NoColor:    true,
		TimeFormat: time.RFC3339,
	})

	// Get ready to load the configuration.
	vpr := viper.New()
	vpr.SetConfigType("yml")
	if cfgFile != "" {
		vpr.SetConfigFile(cfgFile)
	} else {
		vpr.SetConfigName("varnishmon.yml")
		vpr.AddConfigPath(".")
		vpr.AddConfigPath("$HOME/.config/varnishmon")
		vpr.AddConfigPath("$HOME/.config")
		vpr.AddConfigPath("/etc/varnish")
	}

	// Enable automatic binding of environment variables.
	vpr.SetEnvPrefix("VARNISHMON")
	vpr.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	vpr.AllowEmptyEnv(true)
	vpr.AutomaticEnv()

	// Bind command line flags to configuration keys. This is not used for
	// boolean flags (see below).
	for key, flag := range map[string]string{
		"global.logfile":      "logfile",
		"global.loglevel":     "loglevel",
		"db.file":             "db",
		"scraper.period":      "period",
		"scraper.varnishstat": "varnishstat",
		"api.listen-ip":       "ip",
		"api.listen-port":     "port",
	} {
		if err := vpr.BindPFlag(key, rootCmd.PersistentFlags().Lookup(flag)); err != nil {
			log.Fatal().
				Err(err).
				Str("key", key).
				Str("flag", flag).
				Msg("Failed to bind command line flag to configuration key!")
		}
	}

	// Load configuration file.
	if err := vpr.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok { //nolint:errorlint
			// It's acceptable if a configuration file is not explicitly
			// provided and a configuration file is not found in the usual
			// locations. All settings have reasonable defaults.
			if cfgFile != "" {
				log.Fatal().Msg("Configuration file not found!")
			}
		} else {
			log.Fatal().Err(err).
				Str("file", vpr.ConfigFileUsed()).
				Msg("Failed to read configuration file!")
		}
	} else {
		log.Info().
			Str("file", vpr.ConfigFileUsed()).
			Msg("Configuration has been successfully loaded")
	}

	// Boolean flags are a bit special because if the way Cobra parses the
	// command line. Example: usually used as '--no-api', that's equivalent to
	// '--no-api=true' (which is different to '--no-api true'), it could also be
	// user a '--no-api=false' (which is different to '--no-api false').
	for key, booleanFlag := range map[string]string{
		"scraper.enabled": "no-scraper",
		"api.enabled":     "no-api",
	} {
		flag := rootCmd.PersistentFlags().Lookup(booleanFlag)
		if flag.Changed {
			value, err := rootCmd.PersistentFlags().GetBool(booleanFlag)
			if err != nil {
				log.Fatal().
					Err(err).
					Str("key", key).
					Str("flag", booleanFlag).
					Msg("Failed to get boolean flag value!")
			}
			vpr.Set(key, !value)
		}
	}

	// Validate & initialize configuration instance using the loaded file,
	// environment variables and command line options.
	return config.NewConfig(config.NewLogger(&log), vpr)
}
