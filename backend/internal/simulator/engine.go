package simulator

import (
	"fmt"
	"sort"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

const (
	// MaxBatchPoints is the maximum number of 5-minute points allowed per batch (24h * 12).
	MaxBatchPoints = 288
	limaOffsetSec  = -5 * 3600
)

// LimaLocation represents Peru's fixed UTC-5 timezone (America/Lima).
var LimaLocation = time.FixedZone("America/Lima", limaOffsetSec)

// ParseAndValidateBatch validates a simulated 5-minute series payload and expands it into readings.
// Index 0 corresponds to start_time (most recent) and index i shifts backwards by i * 5 minutes.
func ParseAndValidateBatch(req domain.BatchReadingsRequest) ([]domain.TelemetryReading, time.Time, time.Time, error) {
	if req.StartTime == "" {
		return nil, time.Time{}, time.Time{}, fmt.Errorf("start_time is required")
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, time.Time{}, time.Time{}, fmt.Errorf("invalid start_time (expected RFC3339 with -05:00): %w", err)
	}

	_, offset := startTime.Zone()
	if offset != limaOffsetSec {
		return nil, time.Time{}, time.Time{}, fmt.Errorf("start_time must use America/Lima offset (-05:00)")
	}
	if startTime.Second() != 0 || startTime.Nanosecond() != 0 {
		return nil, time.Time{}, time.Time{}, fmt.Errorf("start_time must have 00 seconds")
	}
	if startTime.Minute()%5 != 0 {
		return nil, time.Time{}, time.Time{}, fmt.Errorf("start_time minute must be an exact multiple of 5")
	}

	if len(req.Metrics) == 0 {
		return nil, time.Time{}, time.Time{}, fmt.Errorf("metrics cannot be empty")
	}

	startLima := startTime.In(LimaLocation)
	byTimestamp := make(map[int64]*domain.TelemetryReading)
	maxLen := 0

	for metricName, values := range req.Metrics {
		if !domain.IsValidMetric(metricName) {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("unsupported metric %q", metricName)
		}
		if len(values) == 0 {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("metric %q array cannot be empty", metricName)
		}
		if len(values) > MaxBatchPoints {
			return nil, time.Time{}, time.Time{}, fmt.Errorf("metric %q exceeds maximum %d points per batch", metricName, MaxBatchPoints)
		}
		if len(values) > maxLen {
			maxLen = len(values)
		}

		for i, v := range values {
			if err := domain.ValidateMetricValue(metricName, v); err != nil {
				return nil, time.Time{}, time.Time{}, fmt.Errorf("index %d: %w", i, err)
			}

			ts := startLima.Add(-time.Duration(i) * 5 * time.Minute)
			unixKey := ts.Unix()
			row, exists := byTimestamp[unixKey]
			if !exists {
				row = &domain.TelemetryReading{
					RecordedAt: ts.UTC(),
					Source:     domain.SourceAdminSim,
				}
				byTimestamp[unixKey] = row
			}
			valCopy := v
			if err := row.Metrics.SetMetric(metricName, &valCopy); err != nil {
				return nil, time.Time{}, time.Time{}, err
			}
		}
	}

	keys := make([]int64, 0, len(byTimestamp))
	for k := range byTimestamp {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	readings := make([]domain.TelemetryReading, 0, len(keys))
	for _, k := range keys {
		readings = append(readings, *byTimestamp[k])
	}

	fromTime := startLima.Add(-time.Duration(maxLen-1) * 5 * time.Minute)
	toTime := startLima
	return readings, fromTime, toTime, nil
}
