package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidFromTo     = errors.New("invalid 'from' & 'to'")
	ErrInvalidAggregator = errors.New("invalid aggregator")
	ErrInvalidMetricType = errors.New("invalid metric type")
	ErrUnknownMetricID   = errors.New("unknown metric ID")
)

func (stg *Storage) GetMetrics(from, to time.Time, step int) (map[string]interface{}, error) {
	// Validate 'from' and 'to' parameters.
	if from.After(to) {
		return nil, ErrInvalidFromTo
	}

	// Lock 'db' instance.
	stg.mutex.RLock()
	defer stg.mutex.RUnlock()

	// Normalize 'from', 'to', and 'step' parameters.
	from, to, step, err := stg.unsafeNormalizeFromToAndStep(from, to, step)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize 'from', 'to', and 'step' parameters: %w", err)
	}

	// Fetch metric IDs with samples in the requested time range.
	rows, err := stg.db.Query(`
		SELECT DISTINCT metric_id
		FROM metric_values
		WHERE timestamp >= $1 AND timestamp < $2`, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query 'metric_values' table: %w", err)
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan 'metric_values' rows: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over 'metric_values' rows: %w", err)
	}

	// Lock 'cache'.
	stg.cache.mutex.RLock()
	defer stg.cache.mutex.RUnlock()

	// Decide metrics to be included in the response.
	metrics := make([]map[string]interface{}, 0, len(stg.cache.metricsByID))
	for _, id := range ids {
		metric := stg.cache.metricsByID[id]
		if metric == nil {
			stg.app.Cfg().Log().Warn().
				Int("id", id).
				Msg("Unknown metric ID in 'metric_values' table")
			continue
		}
		metrics = append(metrics, map[string]interface{}{
			"id":          metric.ID,
			"name":        metric.Name,
			"description": metric.Description,
			"flag":        metric.Flag,
			"format":      metric.Format,
		})
	}

	// Done!
	return map[string]interface{}{
		"from":    from.Unix(),
		"to":      to.Unix(),
		"step":    step,
		"metrics": metrics,
	}, nil
}

func (stg *Storage) GetMetric(
	id int, from, to time.Time, step int,
	aggregator string) (map[string]interface{}, error) {
	// Validate 'from' and 'to' parameters.
	if from.After(to) {
		return nil, ErrInvalidFromTo
	}

	// Validate 'id' parameter.
	stg.cache.mutex.RLock()
	metric := stg.cache.metricsByID[id]
	if metric == nil {
		stg.cache.mutex.RUnlock()
		return nil, ErrUnknownMetricID
	}
	stg.cache.mutex.RUnlock()

	// Validate 'aggregator' parameter. See:
	//   - https://duckdb.org/docs/sql/functions/aggregates.html.
	aggregator = strings.ToLower(aggregator)
	switch metric.Flag {
	case "b":
		switch aggregator {
		case "first", "last", "bit_and", "bit_or", "bit_xor", "count":
		default:
			return nil, ErrInvalidAggregator
		}
	default:
		switch aggregator {
		case "avg", "min", "max", "first", "last", "count":
		default:
			return nil, ErrInvalidAggregator
		}
	}

	// Lock 'db' instance.
	stg.mutex.RLock()
	defer stg.mutex.RUnlock()

	// Normalize 'from', 'to', and 'step' parameters.
	from, to, step, err := stg.unsafeNormalizeFromToAndStep(from, to, step)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize 'from', 'to', and 'step' parameters: %w", err)
	}

	// Prepare query to fetch aggregated samples of the requested metric. Note
	// that some metrics (e.g., gauges) are stored as 'uint64' in the database,
	// but when querying DuckDB, they might be returned as 'float64' (e.g., with
	// the 'avg' aggregator) or 'int64' (e.g., with the 'count' aggregator).
	//nolint:gosec
	query := fmt.Sprintf(`
		SELECT
			time_bucket(INTERVAL '%ds', timestamp) AS timestamp,
			%s(value.%s) AS value
		FROM metric_values
		WHERE
			metric_id=$1 AND
			timestamp >= $2 AND
			timestamp < $3
		GROUP BY time_bucket(INTERVAL '%ds', timestamp)
		ORDER BY timestamp`, step, aggregator, metric.Class, step)

	// Query database.
	rows, err := stg.db.Query(query, id, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query 'metric_values' table: %w", err)
	}
	defer rows.Close()

	// Fetch rows.
	samples := make([][2]interface{}, 0)
	for rows.Next() {
		var timestamp time.Time
		var value interface{}
		if err := rows.Scan(&timestamp, &value); err != nil {
			return nil, fmt.Errorf("failed to scan 'metric_values' rows: %w", err)
		}
		samples = append(samples, [2]interface{}{
			// In the client side, seconds gives more than enough granularity,
			// specially taking into account the minimum 'step' value is '1'
			// (because of  the minimum scraper period).
			timestamp.Unix(),

			// The post-aggregation type returned by DuckDB in general is the
			// right one, but some cases like bitmaps require a special
			// treatment.
			metric.FormatValue(value),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over 'metric_values' rows: %w", err)
	}

	// Done!
	return map[string]interface{}{
		"from":    from.Unix(),
		"to":      to.Unix(),
		"step":    step,
		"samples": samples,
	}, nil
}

func (stg *Storage) PushMetricSamples(timestamp time.Time, samples []*MetricSample) error {
	// This is a write operation on 'db' but a read lock is intentionally used.
	// See the note on the 'Storage' type for more information.
	stg.mutex.RLock()
	defer stg.mutex.RUnlock()

	// Using a transaction is crucial to batch the inserts and avoid performance
	// penalties.
	tx, err := stg.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Build prepared statements for batch inserts, one for each metric class.
	uint64Statement, err := tx.Prepare(`
		INSERT INTO metric_values (metric_id, timestamp, value)
		VALUES (?, ?, union_value(uint64 := ?))`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer uint64Statement.Close()
	float64Statement, err := tx.Prepare(`
		INSERT INTO metric_values (metric_id, timestamp, value)
		VALUES (?, ?, union_value(float64 := ?))`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer float64Statement.Close()

	// Prepare batched inserts into the 'metric_values' table. During this
	// process, insert or update in the 'metrics' table if necessary.
	for _, sample := range samples {
		// Check if the metric is known and identical to the one in the database.
		// Non identical metrics will preserve their internal ID, but the rest
		// of the fields will be updated. Beware of locking order: 'stg.mutex'
		// was locked before 'stg.cache.mutex'.
		var metric *CachedMetric
		stg.cache.mutex.RLock()
		if m := stg.cache.metricsByName[sample.Name]; m != nil {
			if m.Flag == sample.Flag && m.Format == sample.Format &&
				m.Description == sample.Description {
				metric = m
			}
		}
		stg.cache.mutex.RUnlock()

		// Insert / update in the 'metrics' table.
		if metric == nil {
			var class string
			switch sample.Value.(type) {
			case uint64:
				class = "uint64"
			case float64:
				class = "float64"
			default:
				return ErrInvalidMetricType
			}

			var metricID int
			if err := stg.db.QueryRow(`
			INSERT INTO metrics (id, name, flag, format, description, class)
			VALUES (
				COALESCE((SELECT id FROM metrics WHERE name = $1), NEXTVAL('metrics_seq')),
				$1, $2, $3, $4, $5)
			ON CONFLICT(name) DO UPDATE SET
				flag = excluded.flag,
				format = excluded.format,
				description = excluded.description
			RETURNING id`,
				sample.Name, sample.Flag, sample.Format, sample.Description, class).Scan(&metricID); err != nil {
				return fmt.Errorf("failed to insert / update into 'metrics' table: %w", err)
			}

			metric = &CachedMetric{
				ID:          metricID,
				Name:        sample.Name,
				Flag:        sample.Flag,
				Format:      sample.Format,
				Description: sample.Description,
				Class:       class,
			}

			// Beware of locking order: 'stg.mutex' was locked before
			// 'stg.cache.mutex'.
			stg.cache.mutex.Lock()
			stg.cache.metricsByID[metric.ID] = metric
			stg.cache.metricsByName[metric.Name] = metric
			stg.cache.mutex.Unlock()
		}

		// Insert sample in the 'metric_values' table.
		switch metric.Class {
		case "uint64":
			if _, err := uint64Statement.Exec(metric.ID, timestamp, sample.Value); err != nil {
				return fmt.Errorf("failed to insert into 'metric_values' table: %w", err)
			}
		case "float64":
			if _, err := float64Statement.Exec(metric.ID, timestamp, sample.Value); err != nil {
				return fmt.Errorf("failed to insert into 'metric_values' table: %w", err)
			}
		}
	}

	// Commit transaction.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update 'earliest' and 'latest' cache values. Beware of locking order:
	// 'stg.mutex' was locked before 'stg.cache.mutex'.
	stg.cache.mutex.Lock()
	if stg.cache.earliest.IsZero() || timestamp.Before(stg.cache.earliest) {
		stg.cache.earliest = timestamp
	}
	if stg.cache.latest.IsZero() || timestamp.After(stg.cache.latest) {
		stg.cache.latest = timestamp
	}
	stg.cache.mutex.Unlock()

	// Done!
	return nil
}

func (stg *Storage) unsafeNormalizeFromToAndStep(
	from, to time.Time, step int) (time.Time, time.Time, int, error) {
	// Ensure 'step' is at least the scraper period, if enabled. If disabled,
	// 1s is the minimum resolution.
	period := 1
	if stg.app.Cfg().ScraperEnabled() {
		period = int(stg.app.Cfg().ScraperPeriod().Seconds())
	}
	if step < period {
		step = period
	}

	// Adjust 'from' and 'to' to the nearest 'step' boundaries. Let DuckDB do
	// this job to ensure consistency.
	row := stg.db.QueryRow(fmt.Sprintf(`
		SELECT
			time_bucket(INTERVAL '%ds', $1::TIMESTAMP) AS from,
			time_bucket(INTERVAL '%ds', $2::TIMESTAMP) + INTERVAL '%ds' AS to`,
		step, step, step), from.Format(time.RFC3339), to.Format(time.RFC3339))
	if err := row.Scan(&from, &to); err != nil {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("failed to scan row with normalized boundaries: %w", err)
	}

	// Done!
	return from, to, step, nil
}

func (cm *CachedMetric) FormatValue(value interface{}) interface{} {
	if cm.Format == "b" {
		// Once aggregated, bitmaps can be returned as 'uint64' (e.g., 'last')
		// or 'int64' (e.g., 'count'). Hex representation is only useful for
		// the 'uint64' case.
		if v, ok := value.(uint64); ok {
			return strconv.FormatUint(v, 16)
		}
	}
	return value
}
