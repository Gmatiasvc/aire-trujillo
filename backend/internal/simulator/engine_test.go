package simulator

import (
	"math"
	"testing"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

func TestParseAndValidateBatchSuccess(t *testing.T) {
	req := domain.BatchReadingsRequest{
		StartTime: "2026-11-13T16:00:00-05:00",
		Metrics: map[string][]float64{
			"pm25":          {18.4, 17.9, 16.5},
			"pm10":          {34.0, 33.1, 30.8},
			"temperature_c": {21.5, 21.6, 21.8},
		},
	}

	readings, from, to, err := ParseAndValidateBatch(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(readings) != 3 {
		t.Fatalf("expected 3 consolidated readings, got %d", len(readings))
	}

	if from.Format(time.RFC3339) != "2026-11-13T15:50:00-05:00" {
		t.Fatalf("unexpected from: %s", from.Format(time.RFC3339))
	}
	if to.Format(time.RFC3339) != "2026-11-13T16:00:00-05:00" {
		t.Fatalf("unexpected to: %s", to.Format(time.RFC3339))
	}

	// Readings are sorted ascending: [15:50, 15:55, 16:00]
	// Index 2 of input (16.5) -> 15:50 (-05:00) = 20:50Z
	// Index 0 of input (18.4) -> 16:00 (-05:00) = 21:00Z
	if readings[0].RecordedAt.UTC().Format(time.RFC3339) != "2026-11-13T20:50:00Z" {
		t.Fatalf("unexpected oldest timestamp: %s", readings[0].RecordedAt.UTC().Format(time.RFC3339))
	}
	if *readings[0].Metrics.PM25 != 16.5 {
		t.Fatalf("expected pm25=16.5 at 15:50, got %v", *readings[0].Metrics.PM25)
	}
	if *readings[2].Metrics.PM25 != 18.4 {
		t.Fatalf("expected pm25=18.4 at 16:00, got %v", *readings[2].Metrics.PM25)
	}
	if readings[2].Source != domain.SourceAdminSim {
		t.Fatalf("expected source admin_sim, got %s", readings[2].Source)
	}
}

func TestParseAndValidateBatchRejections(t *testing.T) {
	tests := []struct {
		name string
		req  domain.BatchReadingsRequest
	}{
		{
			name: "non-5-minute aligned start_time",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:03:00-05:00",
				Metrics:   map[string][]float64{"pm25": {10.0}},
			},
		},
		{
			name: "wrong timezone offset",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:00Z",
				Metrics:   map[string][]float64{"pm25": {10.0}},
			},
		},
		{
			name: "non-zero seconds",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:30-05:00",
				Metrics:   map[string][]float64{"pm25": {10.0}},
			},
		},
		{
			name: "out of range metric value",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:00-05:00",
				Metrics:   map[string][]float64{"pm25": {1500.0}},
			},
		},
		{
			name: "NaN metric value",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:00-05:00",
				Metrics:   map[string][]float64{"pm25": {math.NaN()}},
			},
		},
		{
			name: "Inf metric value",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:00-05:00",
				Metrics:   map[string][]float64{"pm25": {math.Inf(1)}},
			},
		},
		{
			name: "unknown metric key",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:00-05:00",
				Metrics:   map[string][]float64{"unknown_var": {10.0}},
			},
		},
		{
			name: "exceeds 288 points",
			req: domain.BatchReadingsRequest{
				StartTime: "2026-11-13T16:00:00-05:00",
				Metrics:   map[string][]float64{"pm25": make([]float64, 289)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, _, err := ParseAndValidateBatch(tc.req); err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}
