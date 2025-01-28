package main

// This is required to generate mock files, but it's not used anywhere. See:
//   - https://vektra.github.io/mockery/latest/notes/#error-no-go-files-found-in-root-search-path.

import "github.com/allenta/varnishmon/pkg/config"

func main() {
	config.Version()
}
