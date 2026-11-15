package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

func f64(v float64) *float64 {
	return &v
}

func TestPostgresStoreIntegration(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://air_admin:air_admin@localhost:5432/air_quality_db?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	st, err := NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping Postgres integration test (database unreachable): %v", err)
	}
	defer st.Close()

	if err := st.RunMigrations(ctx, "../../migrations"); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Ensure clean starting state for TEST-INT-NODE if a previous run aborted.
	if existing, err := st.GetNodeByCode(ctx, "TEST-INT-NODE"); err == nil && existing != nil {
		_ = st.DeleteNode(ctx, existing.ID)
	}

	created, err := st.CreateNode(ctx, domain.CreateNodeRequest{
		Code:   "TEST-INT-NODE",
		Name:   "Test Integration Plaza",
		Lat:    -8.1116,
		Lon:    -79.0288,
		Status: domain.NodeStatusActive,
	}, true)
	if err != nil {
		t.Fatalf("CreateNode failed: %v", err)
	}
	defer func() {
		_ = st.DeleteNode(context.Background(), created.ID)
	}()

	byCode, err := st.GetNodeByCode(ctx, "TEST-INT-NODE")
	if err != nil || byCode.ID != created.ID {
		t.Fatalf("GetNodeByCode mismatch: got (%+v, %v)", byCode, err)
	}

	byID, err := st.GetNodeByID(ctx, created.ID)
	if err != nil || byID.Code != "TEST-INT-NODE" {
		t.Fatalf("GetNodeByID mismatch: got (%+v, %v)", byID, err)
	}

	activeNodes, err := st.ListActiveNodes(ctx)
	if err != nil {
		t.Fatalf("ListActiveNodes failed: %v", err)
	}
	foundActive := false
	for _, n := range activeNodes {
		if n.ID == created.ID {
			foundActive = true
			break
		}
	}
	if !foundActive {
		t.Fatalf("expected created node in ListActiveNodes")
	}

	newName := "Test Integration Updated"
	updated, err := st.UpdateNode(ctx, created.ID, domain.UpdateNodeRequest{
		Name: &newName,
	})
	if err != nil || updated.Name != newName {
		t.Fatalf("UpdateNode mismatch: got (%+v, %v)", updated, err)
	}

	t1 := time.Date(2026, 11, 13, 21, 0, 0, 0, time.UTC)
	t0 := t1.Add(-5 * time.Minute)

	// First batch: pm25 + pm10
	n1, err := st.UpsertReadingsBatch(ctx, created.ID, []domain.TelemetryReading{
		{
			RecordedAt: t0,
			Source:     domain.SourceAdminSim,
			Metrics:    domain.TelemetryMetrics{PM25: f64(16.5), PM10: f64(30.8)},
		},
		{
			RecordedAt: t1,
			Source:     domain.SourceAdminSim,
			Metrics:    domain.TelemetryMetrics{PM25: f64(18.4), PM10: f64(34.0)},
		},
	})
	if err != nil || n1 != 2 {
		t.Fatalf("first UpsertReadingsBatch failed: n=%d err=%v", n1, err)
	}

	// Second batch on same t1 with only temperature_c -> COALESCE must keep pm25 and pm10 intact
	n2, err := st.UpsertReadingsBatch(ctx, created.ID, []domain.TelemetryReading{
		{
			RecordedAt: t1,
			Source:     domain.SourceAdminSim,
			Metrics:    domain.TelemetryMetrics{TemperatureC: f64(21.5)},
		},
	})
	if err != nil || n2 != 1 {
		t.Fatalf("second UpsertReadingsBatch failed: n=%d err=%v", n2, err)
	}

	readings, err := st.GetNodeReadings(ctx, created.ID, 24)
	if err != nil || len(readings) != 2 {
		t.Fatalf("GetNodeReadings mismatch: len=%d err=%v", len(readings), err)
	}
	// Descending order: index 0 is t1
	rLatest := readings[0]
	if rLatest.Metrics.PM25 == nil || *rLatest.Metrics.PM25 != 18.4 ||
		rLatest.Metrics.PM10 == nil || *rLatest.Metrics.PM10 != 34.0 ||
		rLatest.Metrics.TemperatureC == nil || *rLatest.Metrics.TemperatureC != 21.5 {
		t.Fatalf("COALESCE upsert did not preserve partial columns: %+v", rLatest.Metrics)
	}

	latestTS, err := st.GetLatestTimestamp(ctx)
	if err != nil || latestTS == nil {
		t.Fatalf("GetLatestTimestamp failed: ts=%v err=%v", latestTS, err)
	}

	snaps, err := st.GetSnapshotReadings(ctx, t1)
	if err != nil {
		t.Fatalf("GetSnapshotReadings failed: %v", err)
	}
	foundSnap := false
	for _, s := range snaps {
		if s.NodeID == created.ID {
			foundSnap = true
			if s.Metrics.PM25 == nil || *s.Metrics.PM25 != 18.4 {
				t.Fatalf("unexpected snapshot pm25: %+v", s.Metrics)
			}
		}
	}
	if !foundSnap {
		t.Fatalf("expected created node in snapshot readings")
	}

	playback, err := st.GetPlaybackTimestamps(ctx)
	if err != nil || len(playback) < 2 {
		t.Fatalf("GetPlaybackTimestamps mismatch: len=%d err=%v", len(playback), err)
	}

	windowRows, err := st.GetActiveNodesReadingsWindow(ctx, t0, t1)
	if err != nil || len(windowRows) < 2 {
		t.Fatalf("GetActiveNodesReadingsWindow mismatch: len=%d err=%v", len(windowRows), err)
	}

	if err := st.DeleteNode(ctx, created.ID); err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}
	if _, err := st.GetNodeByID(ctx, created.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
