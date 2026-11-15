package mqtt

import (
	"context"
	"testing"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/cache"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/store"
)

type stubRepo struct {
	store.Repository
	nodes   map[string]*domain.Node
	upserts []domain.TelemetryReading
}

func (r *stubRepo) GetNodeByCode(_ context.Context, code string) (*domain.Node, error) {
	n, ok := r.nodes[code]
	if !ok {
		return nil, store.ErrNotFound
	}
	return n, nil
}

func (r *stubRepo) UpsertReading(_ context.Context, reading domain.TelemetryReading) error {
	r.upserts = append(r.upserts, reading)
	return nil
}

func TestHandleMessage(t *testing.T) {
	repo := &stubRepo{
		nodes: map[string]*domain.Node{
			"TRJ-HUANCHACO-01": {
				ID:     "uuid-1",
				Code:   "TRJ-HUANCHACO-01",
				Status: domain.NodeStatusActive,
			},
			"TRJ-INACTIVE": {
				ID:     "uuid-2",
				Code:   "TRJ-INACTIVE",
				Status: domain.NodeStatusInactive,
			},
		},
	}

	memCache := cache.NewMemoryCache(time.Minute)
	memCache.SetPlaybackTimestamps([]string{"2026-11-13T16:00:00Z"})

	sub := NewSubscriber("tcp://localhost:1883", repo, memCache)

	validJSON := []byte(`{
		"node_code": "TRJ-HUANCHACO-01",
		"recorded_at": "2026-11-13T16:00:00Z",
		"metrics": {
			"pm25": 14.2,
			"pm10": 28.5,
			"co": 1.1,
			"temperature_c": 21.4
		}
	}`)

	if err := sub.HandleMessage(context.Background(), "aire-trujillo/v1/nodes/TRJ-HUANCHACO-01/telemetry", validJSON); err != nil {
		t.Fatalf("unexpected error for valid message: %v", err)
	}
	if len(repo.upserts) != 1 {
		t.Fatalf("expected 1 persisted reading, got %d", len(repo.upserts))
	}
	if repo.upserts[0].Source != domain.SourceMQTT {
		t.Fatalf("expected source mqtt, got %s", repo.upserts[0].Source)
	}
	if _, ok := memCache.GetPlaybackTimestamps(); ok {
		t.Fatalf("expected cache to be invalidated after MQTT ingestion")
	}

	// Inactive node is discarded without persisting
	inactiveJSON := []byte(`{
		"node_code": "TRJ-INACTIVE",
		"recorded_at": "2026-11-13T16:00:00Z",
		"metrics": {"pm25": 10.0}
	}`)
	if err := sub.HandleMessage(context.Background(), "aire-trujillo/v1/nodes/TRJ-INACTIVE/telemetry", inactiveJSON); err != nil {
		t.Fatalf("expected inactive node to be gracefully discarded: %v", err)
	}
	if len(repo.upserts) != 1 {
		t.Fatalf("expected no additional upsert for inactive node")
	}

	// Out-of-range metric is rejected
	invalidRangeJSON := []byte(`{
		"node_code": "TRJ-HUANCHACO-01",
		"recorded_at": "2026-11-13T16:00:00Z",
		"metrics": {"pm25": 5000.0}
	}`)
	if err := sub.HandleMessage(context.Background(), "aire-trujillo/v1/nodes/TRJ-HUANCHACO-01/telemetry", invalidRangeJSON); err == nil {
		t.Fatalf("expected error for out-of-range metric")
	}
}
