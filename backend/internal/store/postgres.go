package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

var (
	ErrNotFound  = errors.New("record not found")
	ErrConflict  = errors.New("record conflict")
	ErrInvalidID = errors.New("invalid node id")
)

// Repository defines persistence operations required by API handlers and MQTT ingestion.
type Repository interface {
	Close()
	Ping(ctx context.Context) error
	RunMigrations(ctx context.Context, migrationsDir string) error

	ListActiveNodes(ctx context.Context) ([]domain.PublicNode, error)
	ListAllNodes(ctx context.Context) ([]domain.Node, error)
	GetNodeByCode(ctx context.Context, code string) (*domain.Node, error)
	GetNodeByID(ctx context.Context, id string) (*domain.Node, error)
	CreateNode(ctx context.Context, req domain.CreateNodeRequest, isVirtual bool) (*domain.Node, error)
	UpdateNode(ctx context.Context, id string, req domain.UpdateNodeRequest) (*domain.Node, error)
	DeleteNode(ctx context.Context, id string) error

	UpsertReading(ctx context.Context, reading domain.TelemetryReading) error
	UpsertReadingsBatch(ctx context.Context, nodeID string, readings []domain.TelemetryReading) (int, error)
	GetNodeReadings(ctx context.Context, nodeID string, hours int) ([]domain.TelemetryReading, error)
	GetLatestTimestamp(ctx context.Context) (*time.Time, error)
	GetSnapshotReadings(ctx context.Context, targetTime time.Time) ([]domain.SnapshotNodeReading, error)
	GetPlaybackTimestamps(ctx context.Context) ([]time.Time, error)
	GetActiveNodesReadingsWindow(ctx context.Context, from, to time.Time) ([]domain.SnapshotNodeReading, error)
}

// PostgresStore implements Repository over pgxpool + PostGIS.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore connects to PostgreSQL using conservative free-tier pool limits.
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// Close closes the underlying connection pool.
func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Ping verifies database connectivity.
func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

// RunMigrations executes all .sql files in migrationsDir in lexicographical order.
func (s *PostgresStore) RunMigrations(ctx context.Context, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory %q: %w", migrationsDir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		fullPath := filepath.Join(migrationsDir, name)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", fullPath, err)
		}
		sqlText := strings.TrimSpace(string(content))
		if sqlText == "" {
			continue
		}
		if _, err := s.pool.Exec(ctx, sqlText); err != nil {
			return fmt.Errorf("execute migration %q: %w", name, err)
		}
	}
	return nil
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrConflict
		case "22P02":
			return ErrInvalidID
		}
	}
	return err
}

func (s *PostgresStore) ListActiveNodes(ctx context.Context) ([]domain.PublicNode, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id::text,
			code,
			name,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			status
		FROM nodes
		WHERE status = 'active'
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list active nodes: %w", err)
	}
	defer rows.Close()

	nodes := make([]domain.PublicNode, 0)
	for rows.Next() {
		var n domain.PublicNode
		if err := rows.Scan(&n.ID, &n.Code, &n.Name, &n.Lat, &n.Lon, &n.Status); err != nil {
			return nil, fmt.Errorf("scan active node: %w", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (s *PostgresStore) ListAllNodes(ctx context.Context) ([]domain.Node, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id::text,
			code,
			name,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			status,
			is_virtual,
			created_at,
			updated_at
		FROM nodes
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list all nodes: %w", err)
	}
	defer rows.Close()

	nodes := make([]domain.Node, 0)
	for rows.Next() {
		var n domain.Node
		if err := rows.Scan(&n.ID, &n.Code, &n.Name, &n.Lat, &n.Lon, &n.Status, &n.IsVirtual, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func (s *PostgresStore) GetNodeByCode(ctx context.Context, code string) (*domain.Node, error) {
	var n domain.Node
	err := s.pool.QueryRow(ctx, `
		SELECT
			id::text,
			code,
			name,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			status,
			is_virtual,
			created_at,
			updated_at
		FROM nodes
		WHERE code = $1
	`, code).Scan(&n.ID, &n.Code, &n.Name, &n.Lat, &n.Lon, &n.Status, &n.IsVirtual, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &n, nil
}

func (s *PostgresStore) GetNodeByID(ctx context.Context, id string) (*domain.Node, error) {
	var n domain.Node
	err := s.pool.QueryRow(ctx, `
		SELECT
			id::text,
			code,
			name,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			status,
			is_virtual,
			created_at,
			updated_at
		FROM nodes
		WHERE id = $1
	`, id).Scan(&n.ID, &n.Code, &n.Name, &n.Lat, &n.Lon, &n.Status, &n.IsVirtual, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &n, nil
}

func (s *PostgresStore) CreateNode(ctx context.Context, req domain.CreateNodeRequest, isVirtual bool) (*domain.Node, error) {
	var n domain.Node
	err := s.pool.QueryRow(ctx, `
		INSERT INTO nodes (code, name, status, is_virtual, location)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography)
		RETURNING
			id::text,
			code,
			name,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			status,
			is_virtual,
			created_at,
			updated_at
	`, req.Code, req.Name, req.Status, isVirtual, req.Lon, req.Lat).Scan(
		&n.ID, &n.Code, &n.Name, &n.Lat, &n.Lon, &n.Status, &n.IsVirtual, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &n, nil
}

func (s *PostgresStore) UpdateNode(ctx context.Context, id string, req domain.UpdateNodeRequest) (*domain.Node, error) {
	existing, err := s.GetNodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}
	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}
	lat := existing.Lat
	if req.Lat != nil {
		lat = *req.Lat
	}
	lon := existing.Lon
	if req.Lon != nil {
		lon = *req.Lon
	}

	var updated domain.Node
	err = s.pool.QueryRow(ctx, `
		UPDATE nodes
		SET
			name = $2,
			status = $3,
			location = ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id::text,
			code,
			name,
			ST_Y(location::geometry) AS lat,
			ST_X(location::geometry) AS lon,
			status,
			is_virtual,
			created_at,
			updated_at
	`, id, name, status, lon, lat).Scan(
		&updated.ID, &updated.Code, &updated.Name, &updated.Lat, &updated.Lon,
		&updated.Status, &updated.IsVirtual, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, mapPgError(err)
	}
	return &updated, nil
}

func (s *PostgresStore) DeleteNode(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM nodes WHERE id = $1`, id)
	if err != nil {
		return mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

const upsertReadingSQL = `
INSERT INTO telemetry_readings (
	node_id, recorded_at,
	pm25, pm10, co, no2, so2, h2s,
	co2, voc_index, noise_dba, temperature_c, humidity_pct, pressure_hpa,
	source
) VALUES (
	$1, $2,
	$3, $4, $5, $6, $7, $8,
	$9, $10, $11, $12, $13, $14,
	$15
)
ON CONFLICT (node_id, recorded_at) DO UPDATE SET
	pm25          = COALESCE(EXCLUDED.pm25, telemetry_readings.pm25),
	pm10          = COALESCE(EXCLUDED.pm10, telemetry_readings.pm10),
	co            = COALESCE(EXCLUDED.co, telemetry_readings.co),
	no2           = COALESCE(EXCLUDED.no2, telemetry_readings.no2),
	so2           = COALESCE(EXCLUDED.so2, telemetry_readings.so2),
	h2s           = COALESCE(EXCLUDED.h2s, telemetry_readings.h2s),
	co2           = COALESCE(EXCLUDED.co2, telemetry_readings.co2),
	voc_index     = COALESCE(EXCLUDED.voc_index, telemetry_readings.voc_index),
	noise_dba     = COALESCE(EXCLUDED.noise_dba, telemetry_readings.noise_dba),
	temperature_c = COALESCE(EXCLUDED.temperature_c, telemetry_readings.temperature_c),
	humidity_pct  = COALESCE(EXCLUDED.humidity_pct, telemetry_readings.humidity_pct),
	pressure_hpa  = COALESCE(EXCLUDED.pressure_hpa, telemetry_readings.pressure_hpa),
	source        = EXCLUDED.source
`

func (s *PostgresStore) UpsertReading(ctx context.Context, reading domain.TelemetryReading) error {
	m := reading.Metrics
	_, err := s.pool.Exec(ctx, upsertReadingSQL,
		reading.NodeID, reading.RecordedAt.UTC(),
		m.PM25, m.PM10, m.CO, m.NO2, m.SO2, m.H2S,
		m.CO2, m.VOCIndex, m.NoiseDBA, m.TemperatureC, m.HumidityPct, m.PressureHPA,
		reading.Source,
	)
	return mapPgError(err)
}

func (s *PostgresStore) UpsertReadingsBatch(ctx context.Context, nodeID string, readings []domain.TelemetryReading) (int, error) {
	if len(readings) == 0 {
		return 0, nil
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin batch tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	batch := &pgx.Batch{}
	for _, r := range readings {
		m := r.Metrics
		batch.Queue(upsertReadingSQL,
			nodeID, r.RecordedAt.UTC(),
			m.PM25, m.PM10, m.CO, m.NO2, m.SO2, m.H2S,
			m.CO2, m.VOCIndex, m.NoiseDBA, m.TemperatureC, m.HumidityPct, m.PressureHPA,
			r.Source,
		)
	}

	results := tx.SendBatch(ctx, batch)
	count := 0
	for range readings {
		if _, err := results.Exec(); err != nil {
			_ = results.Close()
			return 0, mapPgError(err)
		}
		count++
	}
	if err := results.Close(); err != nil {
		return 0, fmt.Errorf("close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit batch tx: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) GetNodeReadings(ctx context.Context, nodeID string, hours int) ([]domain.TelemetryReading, error) {
	if hours <= 0 {
		hours = 24
	}
	if _, err := s.GetNodeByID(ctx, nodeID); err != nil {
		return nil, err
	}

	var latest *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(recorded_at)
		FROM telemetry_readings
		WHERE node_id = $1
	`, nodeID).Scan(&latest)
	if err != nil {
		return nil, mapPgError(err)
	}
	if latest == nil {
		return []domain.TelemetryReading{}, nil
	}

	from := latest.Add(-time.Duration(hours) * time.Hour)
	rows, err := s.pool.Query(ctx, `
		SELECT
			id, node_id::text, recorded_at,
			pm25, pm10, co, no2, so2, h2s,
			co2, voc_index, noise_dba, temperature_c, humidity_pct, pressure_hpa,
			source, created_at
		FROM telemetry_readings
		WHERE node_id = $1
		  AND recorded_at >= $2
		  AND recorded_at <= $3
		ORDER BY recorded_at DESC
	`, nodeID, from, *latest)
	if err != nil {
		return nil, fmt.Errorf("query node readings: %w", err)
	}
	defer rows.Close()

	readings := make([]domain.TelemetryReading, 0)
	for rows.Next() {
		var r domain.TelemetryReading
		if err := rows.Scan(
			&r.ID, &r.NodeID, &r.RecordedAt,
			&r.Metrics.PM25, &r.Metrics.PM10, &r.Metrics.CO, &r.Metrics.NO2, &r.Metrics.SO2, &r.Metrics.H2S,
			&r.Metrics.CO2, &r.Metrics.VOCIndex, &r.Metrics.NoiseDBA, &r.Metrics.TemperatureC, &r.Metrics.HumidityPct, &r.Metrics.PressureHPA,
			&r.Source, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan node reading: %w", err)
		}
		readings = append(readings, r)
	}
	return readings, rows.Err()
}

func (s *PostgresStore) GetLatestTimestamp(ctx context.Context) (*time.Time, error) {
	var latest *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(r.recorded_at)
		FROM telemetry_readings r
		JOIN nodes n ON n.id = r.node_id
		WHERE n.status = 'active'
	`).Scan(&latest)
	if err != nil {
		return nil, fmt.Errorf("get latest timestamp: %w", err)
	}
	if latest == nil {
		return nil, nil
	}
	t := latest.UTC()
	return &t, nil
}

func (s *PostgresStore) GetSnapshotReadings(ctx context.Context, targetTime time.Time) ([]domain.SnapshotNodeReading, error) {
	windowStart := targetTime.UTC().Add(-15 * time.Minute)
	windowEnd := targetTime.UTC()

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (n.id)
			n.id::text,
			n.code,
			ST_Y(n.location::geometry) AS lat,
			ST_X(n.location::geometry) AS lon,
			r.recorded_at,
			r.pm25, r.pm10, r.co, r.no2, r.so2, r.h2s,
			r.co2, r.voc_index, r.noise_dba, r.temperature_c, r.humidity_pct, r.pressure_hpa
		FROM nodes n
		JOIN telemetry_readings r ON r.node_id = n.id
		WHERE n.status = 'active'
		  AND r.recorded_at >= $1
		  AND r.recorded_at <= $2
		ORDER BY n.id, r.recorded_at DESC
	`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("get snapshot readings: %w", err)
	}
	defer rows.Close()

	return scanSnapshotRows(rows)
}

func (s *PostgresStore) GetPlaybackTimestamps(ctx context.Context) ([]time.Time, error) {
	latest, err := s.GetLatestTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return []time.Time{}, nil
	}

	from := latest.Add(-24 * time.Hour)
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT
			to_timestamp(((extract(epoch FROM r.recorded_at)::bigint + 299) / 300) * 300) AS bucket
		FROM telemetry_readings r
		JOIN nodes n ON n.id = r.node_id
		WHERE n.status = 'active'
		  AND r.recorded_at >= $1
		  AND r.recorded_at <= $2
		ORDER BY bucket ASC
	`, from, *latest)
	if err != nil {
		return nil, fmt.Errorf("get playback timestamps: %w", err)
	}
	defer rows.Close()

	timestamps := make([]time.Time, 0)
	for rows.Next() {
		var ts time.Time
		if err := rows.Scan(&ts); err != nil {
			return nil, fmt.Errorf("scan playback timestamp: %w", err)
		}
		timestamps = append(timestamps, ts.UTC())
	}
	return timestamps, rows.Err()
}

func (s *PostgresStore) GetActiveNodesReadingsWindow(ctx context.Context, from, to time.Time) ([]domain.SnapshotNodeReading, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			n.id::text,
			n.code,
			ST_Y(n.location::geometry) AS lat,
			ST_X(n.location::geometry) AS lon,
			r.recorded_at,
			r.pm25, r.pm10, r.co, r.no2, r.so2, r.h2s,
			r.co2, r.voc_index, r.noise_dba, r.temperature_c, r.humidity_pct, r.pressure_hpa
		FROM nodes n
		JOIN telemetry_readings r ON r.node_id = n.id
		WHERE n.status = 'active'
		  AND r.recorded_at >= $1
		  AND r.recorded_at <= $2
		ORDER BY r.recorded_at ASC, n.id
	`, from.UTC(), to.UTC())
	if err != nil {
		return nil, fmt.Errorf("get active nodes readings window: %w", err)
	}
	defer rows.Close()

	return scanSnapshotRows(rows)
}

func scanSnapshotRows(rows pgx.Rows) ([]domain.SnapshotNodeReading, error) {
	out := make([]domain.SnapshotNodeReading, 0)
	for rows.Next() {
		var item domain.SnapshotNodeReading
		if err := rows.Scan(
			&item.NodeID,
			&item.Code,
			&item.Lat,
			&item.Lon,
			&item.RecordedAt,
			&item.Metrics.PM25,
			&item.Metrics.PM10,
			&item.Metrics.CO,
			&item.Metrics.NO2,
			&item.Metrics.SO2,
			&item.Metrics.H2S,
			&item.Metrics.CO2,
			&item.Metrics.VOCIndex,
			&item.Metrics.NoiseDBA,
			&item.Metrics.TemperatureC,
			&item.Metrics.HumidityPct,
			&item.Metrics.PressureHPA,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot reading: %w", err)
		}
		item.RecordedAt = item.RecordedAt.UTC()
		out = append(out, item)
	}
	return out, rows.Err()
}
