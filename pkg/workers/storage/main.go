package storage

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/allenta/varnishmon/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

type Storage struct {
	app Application

	// Locking order: 'stg.mutex' -> 'stg.cache.mutex'.

	// Operations on 'db' are thread-safe, but the mutex is required to safely
	// reopen the database (e.g., when rotating the database file or flushing an
	// in-memory database). Switching the 'db' instance is the only operation
	// requiring a write lock; all other operations on 'db' (both reads and
	// writes) should use a read lock to minimize contention.
	mutex sync.RWMutex
	db    *sql.DB

	// This is a simple cache. It helps to avoid querying the database for
	// metric details and other, mostly static, information. Useful both for
	// performance and to simplify the logic.
	cache struct {
		mutex sync.RWMutex

		// Known metrics, indexed by ID and name.
		metricsByID   map[int]*CachedMetric
		metricsByName map[string]*CachedMetric

		// Hostname, as stored in the 'metadata' table.
		hostname string

		// Earliest and latest timestamps in the 'metric_values' table.
		earliest time.Time
		latest   time.Time
	}
}

type Application interface {
	Cfg() *config.Config
}

type CachedMetric struct {
	ID          int
	Name        string
	Flag        string
	Format      string
	Description string
	Class       string
}
type MetricSample struct {
	Name        string
	Flag        string
	Format      string
	Description string
	Value       interface{}
}

func NewStorage(app Application) *Storage {
	// Create instance.
	stg := &Storage{
		app: app,
	}

	// Initialize database & cache.
	stg.init()

	// Register metrics.
	stg.app.Cfg().Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "storage_db_memory_usage_megabytes",
			Help: "Current amount of memory used by the database.",
		},
		func() float64 {
			return stg.getDuckDBMemoryColumn("memory_usage_bytes")
		},
	))
	stg.app.Cfg().Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "storage_db_temporary_storage_megabytes",
			Help: "Current amount of temporary storage used by the database.",
		},
		func() float64 {
			return stg.getDuckDBMemoryColumn("temporary_storage_bytes")
		},
	))
	stg.app.Cfg().Metrics().Registry.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "storage_db_file_size_megabytes",
			Help: "Current size of the database file.",
		},
		func() float64 {
			return stg.getDuckDBFileSize()
		},
	))

	// Listen to SIGHUP events.
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGHUP)
	go func() {
		// Beware this goroutine finalization logging logic is dummy because
		// termination of this goroutine is not expected. That's intentional,
		// but logging is still useful for troubleshooting in order to identify
		// unexpected terminations.
		defer stg.app.Cfg().Log().Info().
			Msg("Stopped storage event listener")

		stg.app.Cfg().Log().Info().
			Msg("Started storage event listener")

		for {
			sig := <-channel

			stg.app.Cfg().Log().Info().
				Stringer("signal", sig).
				Msg("Got system signal: reopening database")

			stg.init()
		}
	}()

	// Done!
	return stg
}

func (stg *Storage) Earliest() time.Time {
	stg.cache.mutex.RLock()
	defer stg.cache.mutex.RUnlock()
	return stg.cache.earliest
}

func (stg *Storage) Latest() time.Time {
	stg.cache.mutex.RLock()
	defer stg.cache.mutex.RUnlock()
	return stg.cache.latest
}

func (stg *Storage) Hostname() string {
	stg.cache.mutex.RLock()
	defer stg.cache.mutex.RUnlock()
	return stg.cache.hostname
}

func (stg *Storage) Shutdown() error {
	stg.mutex.Lock()
	defer stg.mutex.Unlock()
	stg.cache.mutex.Lock()
	defer stg.cache.mutex.Unlock()

	if err := stg.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	stg.db = nil

	stg.cache.metricsByID = nil
	stg.cache.metricsByName = nil
	stg.cache.hostname = ""
	stg.cache.earliest = time.Time{}
	stg.cache.latest = time.Time{}

	return nil
}

func (stg *Storage) getDuckDBMemoryColumn(name string) float64 {
	stg.mutex.RLock()
	defer stg.mutex.RUnlock()

	// See: https://duckdb.org/docs/sql/meta/duckdb_table_functions.
	//nolint:gosec
	row := stg.db.QueryRow(
		`SELECT sum(` + name + `) / (1024*1024) FROM duckdb_memory()`)
	var result float64
	if err := row.Scan(&result); err != nil {
		stg.app.Cfg().Log().Error().
			Err(err).
			Msg("Failed to query 'duckdb_memory'!")
	}
	return result
}

func (stg *Storage) getDuckDBFileSize() float64 {
	// Using DuckDB's 'database_size' & 'wal_size' columns in 'pragma_database_size()'
	// would be ideal, but the data there is human-readable and not suitable for
	// programmatic use.
	file := stg.app.Cfg().DBFile()
	var result float64
	if file != "" {
		if info, err := os.Stat(file); err == nil {
			result = float64(info.Size()) / (1024 * 1024)
		} else {
			stg.app.Cfg().Log().Error().
				Err(err).
				Str("file", file).
				Msg("Failed to get database file size!")
		}
	}
	return result
}
