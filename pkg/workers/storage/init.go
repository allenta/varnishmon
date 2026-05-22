//nolint:noctx
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/allenta/varnishmon/pkg/config"
	_ "github.com/duckdb/duckdb-go/v2" // Register the DuckDB driver.
)

const (
	SchemaVersion = 2
)

func (stg *Storage) init() {
	// Write lock the db and cache mutexes. Beware of locking order.
	stg.mutex.Lock()
	defer stg.mutex.Unlock()
	stg.cache.mutex.Lock()
	defer stg.cache.mutex.Unlock()

	// Log start of database and cache initialization.
	start := time.Now()
	stg.app.Cfg().Log().Info().
		Str("file", stg.app.Cfg().DBFile()).
		Msg("Initializing database & cache. This may take a while")

	// Initialize the database and cache.
	if stg.db != nil {
		stg.db.Close()
	}
	stg.unsafeOpenDB()
	stg.unsafeConfigureDB()
	stg.unsafeMigrateDBTables()
	stg.unsafeCreateDBTables()
	stg.unsafeInitCache()

	// Fetch some database information, just for logging purposes.
	row := stg.db.QueryRow(`
		SELECT database_size, wal_size, memory_usage, memory_limit
		FROM pragma_database_size()
		LIMIT 1`)
	var databaseSize, walSize, memoryUsage, memoryLimit string
	if err := row.Scan(&databaseSize, &walSize, &memoryUsage, &memoryLimit); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Msg("Failed to query 'database_size'!")
	}
	row = stg.db.QueryRow(`
		SELECT
			current_setting('threads') AS threads,
			current_setting('temp_directory') AS temp_directory,
			current_setting('max_temp_directory_size') AS max_temp_directory_size`)
	var threads, tempDirectory, maxTempDirectorySize string
	if err := row.Scan(&threads, &tempDirectory, &maxTempDirectorySize); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Msg("Failed to query 'current_setting'!")
	}

	// Done!
	stg.app.Cfg().Log().Info().
		Str("file", stg.app.Cfg().DBFile()).
		Str("duration", time.Since(start).String()).
		Str("database_size", databaseSize).
		Str("wal_size", walSize).
		Str("memory_usage", memoryUsage).
		Str("memory_limit", memoryLimit).
		Str("threads", threads).
		Str("temp_directory", tempDirectory).
		Str("max_temp_directory_size", maxTempDirectorySize).
		Msg("Database & cache have been successfully initialized")
}

func (stg *Storage) unsafeOpenDB() {
	// Create a new database instance. A in-memory database is used when an
	// empty string is provided as the database file.
	var err error
	if stg.db, err = sql.Open("duckdb", stg.app.Cfg().DBFile()); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Str("file", stg.app.Cfg().DBFile()).
			Msg("Failed to open database!")
	}
}

func (stg *Storage) unsafeConfigureDB() {
	// Adjust configuration of the database to limit resource usage. See:
	//   - https://duckdb.org/docs/configuration/overview.html.
	//   - https://duckdb.org/2024/07/09/memory-management.html.
	if _, err := stg.db.Exec(fmt.Sprintf(`
		SET memory_limit = '%dMiB';
		SET threads = '%d';
		SET temp_directory = '%s';
		SET max_temp_directory_size = '%dMiB';
		SET preserve_insertion_order = false;`,
		stg.app.Cfg().DBMemoryLimit(),
		stg.app.Cfg().DBThreads(),
		stg.app.Cfg().DBTempDirectory(),
		stg.app.Cfg().DBMaxTempDirectorySize())); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Msg("Failed to set database configuration!")
	}
}

func (stg *Storage) unsafeMigrateDBTables() {
	// Skip when the database is fresh (no 'metadata' table yet); the current
	// schema will be created from scratch by 'unsafeCreateDBTables'.
	var hasMetadata bool
	if err := stg.db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM information_schema.tables
		WHERE table_name = 'metadata'`).Scan(&hasMetadata); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Msg("Failed to check existence of 'metadata' table!")
	}
	if !hasMetadata {
		return
	}

	// Check the current schema version.
	var version int
	if err := stg.db.QueryRow(
		`SELECT schema_version FROM metadata LIMIT 1`).Scan(&version); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Msg("Failed to query 'metadata.schema_version'!")
	}

	// v1 -> v2: drop the FOREIGN KEY constraint on 'metric_values.metric_id'.
	// Recent DuckDB versions enforce FKs strictly on UPDATEs of the referenced
	// row (UPDATE is implemented as DELETE + INSERT internally), which makes
	// it impossible to update metric metadata once samples reference it.
	// Referential integrity is maintained by the application via the metrics
	// cache and the 'metrics_seq' sequence. See:
	//   - https://github.com/duckdb/duckdb/pull/16399.
	//   - https://github.com/duckdb/duckdb-web/pull/4932/changes.
	//   - https://duckdb.org/docs/current/sql/indexes#over-eager-constraint-checking-in-foreign-keys.
	if version < 2 {
		if _, err := stg.db.Exec(`
			CREATE TABLE metric_values_new (
				metric_id INTEGER NOT NULL,
				timestamp TIMESTAMP NOT NULL,
				value UNION(float64 FLOAT8, uint64 UBIGINT) NOT NULL,
				PRIMARY KEY (metric_id, timestamp)
			);
			INSERT INTO metric_values_new SELECT * FROM metric_values;
			DROP TABLE metric_values;
			ALTER TABLE metric_values_new RENAME TO metric_values;
			UPDATE metadata SET schema_version = 2;`); err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to migrate schema from v1 to v2!")
		}
		stg.app.Cfg().Log().Info().
			Msg("Migrated database schema from v1 to v2 (dropped FK on 'metric_values')")
	}
}

func (stg *Storage) unsafeCreateDBTables() {
	// Create the database tables, if they do not exist.
	if _, err := stg.db.Exec(`
		CREATE TABLE IF NOT EXISTS metadata (
			app_version VARCHAR NOT NULL,
			app_revision VARCHAR NOT NULL,
			schema_version INTEGER NOT NULL,
			hostname VARCHAR NOT NULL
		);

		CREATE SEQUENCE IF NOT EXISTS metrics_seq;

		CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY,
			name VARCHAR NOT NULL,
			flag VARCHAR NOT NULL,
			format VARCHAR NOT NULL,
			description VARCHAR NOT NULL,
			class VARCHAR NOT NULL,
			UNIQUE(name)
		);

		CREATE TABLE IF NOT EXISTS metric_values (
			metric_id INTEGER NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			value UNION(float64 FLOAT8, uint64 UBIGINT) NOT NULL,
			PRIMARY KEY (metric_id, timestamp)
		)`); err != nil {
		stg.app.Cfg().Log().Fatal().
			Err(err).
			Msg("Failed to create database tables!")
	}

	// Populate the metadata table, if it is empty.
	{
		hostname, err := os.Hostname()
		if err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to get hostname!")
		}

		row := stg.db.QueryRow(`SELECT COUNT(*) FROM metadata`)
		var count int
		if err := row.Scan(&count); err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to query 'metadata' table!")
		}
		if count == 0 {
			if _, err := stg.db.Exec(`
				INSERT INTO metadata (app_version, app_revision, schema_version, hostname)
				VALUES ($1, $2, $3, $4)`,
				config.Version(), config.Revision(), SchemaVersion,
				hostname); err != nil {
				stg.app.Cfg().Log().Fatal().
					Err(err).
					Msg("Failed to insert into 'metadata' table!")
			}
		}
	}
}

func (stg *Storage) unsafeInitCache() {
	// Initialize the cache of known metrics.
	{
		rows, err := stg.db.Query(`
			SELECT id, name, flag, format, description, class
			FROM metrics`)
		if err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to query 'metrics' table!")
		}
		defer rows.Close()

		stg.cache.metricsByID = make(map[int]*CachedMetric)
		stg.cache.metricsByName = make(map[string]*CachedMetric)
		for rows.Next() {
			var metric CachedMetric
			if err := rows.Scan(
				&metric.ID, &metric.Name, &metric.Flag, &metric.Format,
				&metric.Description, &metric.Class); err != nil {
				stg.app.Cfg().Log().Fatal().
					Err(err).
					Msg("Failed to scan 'metrics' rows!")
			}
			stg.cache.metricsByID[metric.ID] = &metric
			stg.cache.metricsByName[metric.Name] = &metric
		}
		if err := rows.Err(); err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to iterate over 'metrics' rows!")
		}
	}

	// Initialize the cache of earliest and latest timestamps, if some data
	// exists in the database.
	{
		row := stg.db.QueryRow(`
			SELECT
				min(timestamp) AS earliest,
				max(timestamp) AS latest
			FROM metric_values`)
		var earliest, latest *time.Time
		if err := row.Scan(&earliest, &latest); err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to query earliest and latest timestamps in 'metric_values' table!")
		}
		if earliest != nil && latest != nil {
			stg.cache.earliest = *earliest
			stg.cache.latest = *latest
		}
	}

	// Initialize the cached of metadata.
	{
		row := stg.db.QueryRow(`SELECT hostname FROM metadata LIMIT 1`)
		if err := row.Scan(&stg.cache.hostname); err != nil {
			stg.app.Cfg().Log().Fatal().
				Err(err).
				Msg("Failed to query hostname in 'metadata' table!")
		}
	}
}
