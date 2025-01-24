package main

import (
	"errors"
	"fmt"
	"os"

	"allenta.com/varnishmon/pkg/application"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	errInvalidArguments = errors.New("invalid arguments")
	manCmd              = &cobra.Command{ //nolint:gochecknoglobals
		Use:   "man [section] [folder]",
		Short: "Generate man pages and write them to the specified location",
		Args: func(cmd *cobra.Command, args []string) error { //nolint:revive
			if len(args) != 2 {
				return fmt.Errorf("requires a section and a destination folder: %w",
					errInvalidArguments)
			}
			if info, err := os.Stat(args[1]); os.IsNotExist(err) || !info.IsDir() {
				return fmt.Errorf("'%s' is an invalid destination folder: %w",
					args[1], errInvalidArguments)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error { //nolint:revive
			return executeMan(args[0], args[1])
		},
	}
)

func init() {
	rootCmd.AddCommand(manCmd)
}

func executeMan(section, folder string) error {
	header := &doc.GenManHeader{
		Title:   "varnishmon",
		Section: section,
	}
	if err := doc.GenManTree(application.RootCmd, header, folder); err != nil {
		return fmt.Errorf("failed to generate man pages: %w", err)
	}
	return nil
}
