package helpers

import (
	"encoding/json"
	"fmt"
	"time"
)

type VarnishMetrics struct {
	Timestamp time.Time `json:"timestamp"`
	Items     map[string]*VarnishMetricDetails
}

type VarnishMetricDetails struct {
	Description string `json:"description"`
	Flag        string `json:"flag"`
	Format      string `json:"format"`
	Value       uint64 `json:"value"`
}

func (vm *VarnishMetrics) UnmarshalJSON(data []byte) error {
	// Initialize the 'Items' map to avoid nil pointer dereference.
	vm.Items = make(map[string]*VarnishMetricDetails)

	// Parse 'varnishstat -1 -j' output. It is a JSON object with keys being the
	// metric names and values being objects with the metric details, BUT also
	// an initial key called 'timestamp' with a simple string value indicating
	// the timestamp of the metrics is included. This little detail complicates
	// the unmarshaling a little bit and therefore the need for this function.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("invalid 'varnishstat' response: %w", err)
	}

	// Unmarshal the timestamp and the metric details.
	for key, value := range raw {
		if key == "timestamp" {
			// Ideally we'd prefer to parse the timestamp and use it later as
			// the timestamp of the metrics, but the timestamp value does not
			// include the timezone information. That results in the timestamp
			// string being parsed as UTC, which is not correct. For now,
			// we use the current time as the timestamp of the metrics.
			// var ts string
			// if err := json.Unmarshal(value, &ts); err != nil {
			// 	return fmt.Errorf("invalid timestamp: %w", err)
			// }
			// parsedTimestamp, err := time.Parse("2006-01-02T15:04:05", ts)
			// if err != nil {
			// 	return fmt.Errorf("invalid timestamp: %w", err)
			// }
			vm.Timestamp = time.Now()
		} else {
			var details VarnishMetricDetails
			if err := json.Unmarshal(value, &details); err != nil {
				return fmt.Errorf("invalid metric details: %w", err)
			}
			vm.Items[key] = &details
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
	var stats VarnishMetrics
	err := json.Unmarshal(input, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'varnishstat' output: %w", err)
	}
	return &stats, nil
}
