package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	//nolint:godot
	// errMissingTimestamp = errors.New("timestamp field is missing")
	errMissingCounters = errors.New("counters field is missing")
)

type VarnishMetrics struct {
	Version   int                              `json:"version"`
	Timestamp time.Time                        `json:"timestamp"`
	Items     map[string]*VarnishMetricDetails `json:"items"`
}

type VarnishMetricDetails struct {
	Description string `json:"description"`
	Flag        string `json:"flag"`
	Format      string `json:"format"`
	Value       uint64 `json:"value"`
}

func (vm *VarnishMetrics) UnmarshalJSON(data []byte) error {
	// Parse 'varnishstat -1 -j' output.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid 'varnishstat' response: %w", err)
	}

	// Unmarshal the version.
	if version, ok := raw["version"]; ok {
		if err := json.Unmarshal(version, &vm.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	} else {
		vm.Version = 0
	}

	// Unmarshal the timestamp: ideally we'd prefer to parse the timestamp and
	// use it later as the timestamp of the metrics, but the timestamp value
	// does not include the timezone information. That results in the timestamp
	// string being parsed as UTC, which is not correct. For now, we use the
	// current time as the timestamp of the metrics.
	// if timestamp, ok := raw["timestamp"]; ok {
	// 	var ts string
	// 	if err := json.Unmarshal(timestamp, &ts); err != nil {
	// 		return fmt.Errorf("invalid timestamp: %w", err)
	// 	}
	// 	parsedTimestamp, err := time.Parse("2006-01-02T15:04:05", ts)
	// 	if err != nil {
	// 		return fmt.Errorf("invalid timestamp: %w", err)
	// 	}
	// 	vm.Timestamp = parsedTimestamp
	// } else {
	// 	 return fmt.Errorf("invalid timestamp: %w", errMissingTimestamp)
	// }
	vm.Timestamp = time.Now()

	// Unmarshal the metric details.
	vm.Items = make(map[string]*VarnishMetricDetails)
	if vm.Version > 0 {
		// Version > 0: metric details are defined inside the counters object.
		if rawCounters, ok := raw["counters"]; ok {
			if err := json.Unmarshal(rawCounters, &vm.Items); err != nil {
				return fmt.Errorf("invalid counters: %w", err)
			}
		} else {
			return fmt.Errorf("invalid counters: %w", errMissingCounters)
		}
	} else {
		// Version 0: in this version, output is a JSON	object with keys being
		// the metric names and values being objects with the metric details,
		// BUT also an initial key called 'timestamp' with a simple string value
		// indicating the timestamp of the metrics is included. Therefore the
		// timestamp must be filtered out and all metrics have to be parsed.
		for key, value := range raw {
			if key != "timestamp" {
				var details VarnishMetricDetails
				if err := json.Unmarshal(value, &details); err != nil {
					return fmt.Errorf("invalid metric details: %w", err)
				}
				vm.Items[key] = &details
			}
		}
	}

	return nil
}

func (vmd *VarnishMetricDetails) IsCounter() bool {
	return vmd.Flag == "c"
}

func (vmd *VarnishMetricDetails) IsBitmap() bool {
	return vmd.Flag == "b"
}

func (vmd *VarnishMetricDetails) HasDurationFormat() bool {
	return vmd.Format == "d"
}

func ParseVarnishMetrics(input []byte) (*VarnishMetrics, error) {
	var metrics VarnishMetrics
	if err := json.Unmarshal(input, &metrics); err != nil {
		return nil, fmt.Errorf("failed to parse 'varnishstat' output: %w", err)
	}
	return &metrics, nil
}
