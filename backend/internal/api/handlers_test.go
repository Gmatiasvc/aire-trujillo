package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/auth"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/cache"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/spatial"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/store"
)

type memRepo struct {
	mu       sync.Mutex
	nodes    map[string]domain.Node
	readings map[string][]domain.TelemetryReading
	nextSeq  int
}

func newMemRepo() *memRepo {
	return &memRepo{
		nodes:    make(map[string]domain.Node),
		readings: make(map[string][]domain.TelemetryReading),
	}
}

func (m *memRepo) Close()                        {}
func (m *memRepo) Ping(_ context.Context) error  { return nil }
func (m *memRepo) RunMigrations(_ context.Context, _ string) error { return nil }

func (m *memRepo) ListActiveNodes(_ context.Context) ([]domain.PublicNode, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.PublicNode, 0)
	for _, n := range m.nodes {
		if n.Status == domain.NodeStatusActive {
			out = append(out, domain.PublicNode{
				ID:     n.ID,
				Code:   n.Code,
				Name:   n.Name,
				Lat:    n.Lat,
				Lon:    n.Lon,
				Status: n.Status,
			})
		}
	}
	return out, nil
}

func (m *memRepo) ListAllNodes(_ context.Context) ([]domain.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]domain.Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		out = append(out, n)
	}
	return out, nil
}

func (m *memRepo) GetNodeByCode(_ context.Context, code string) (*domain.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, n := range m.nodes {
		if n.Code == code {
			copyN := n
			return &copyN, nil
		}
	}
	return nil, store.ErrNotFound
}

func (m *memRepo) GetNodeByID(_ context.Context, id string) (*domain.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	copyN := n
	return &copyN, nil
}

func (m *memRepo) CreateNode(_ context.Context, req domain.CreateNodeRequest, isVirtual bool) (*domain.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.nodes {
		if existing.Code == req.Code {
			return nil, store.ErrConflict
		}
	}
	m.nextSeq++
	id := "550e8400-e29b-41d4-a716-44665544000" + string(rune('0'+m.nextSeq))
	now := time.Now().UTC()
	n := domain.Node{
		ID:        id,
		Code:      req.Code,
		Name:      req.Name,
		Lat:       req.Lat,
		Lon:       req.Lon,
		Status:    req.Status,
		IsVirtual: isVirtual,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.nodes[id] = n
	return &n, nil
}

func (m *memRepo) UpdateNode(_ context.Context, id string, req domain.UpdateNodeRequest) (*domain.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	if req.Name != nil {
		n.Name = *req.Name
	}
	if req.Status != nil {
		n.Status = *req.Status
	}
	if req.Lat != nil {
		n.Lat = *req.Lat
	}
	if req.Lon != nil {
		n.Lon = *req.Lon
	}
	n.UpdatedAt = time.Now().UTC()
	m.nodes[id] = n
	return &n, nil
}

func (m *memRepo) DeleteNode(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.nodes[id]; !ok {
		return store.ErrNotFound
	}
	delete(m.nodes, id)
	delete(m.readings, id)
	return nil
}

func (m *memRepo) UpsertReading(_ context.Context, r domain.TelemetryReading) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readings[r.NodeID] = append(m.readings[r.NodeID], r)
	return nil
}

func (m *memRepo) UpsertReadingsBatch(_ context.Context, nodeID string, readings []domain.TelemetryReading) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range readings {
		r.NodeID = nodeID
		m.readings[nodeID] = append(m.readings[nodeID], r)
	}
	return len(readings), nil
}

func (m *memRepo) GetNodeReadings(_ context.Context, nodeID string, _ int) ([]domain.TelemetryReading, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.nodes[nodeID]; !ok {
		return nil, store.ErrNotFound
	}
	rows := append([]domain.TelemetryReading(nil), m.readings[nodeID]...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].RecordedAt.After(rows[j].RecordedAt) })
	return rows, nil
}

func (m *memRepo) GetLatestTimestamp(_ context.Context) (*time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latest *time.Time
	for id, n := range m.nodes {
		if n.Status != domain.NodeStatusActive {
			continue
		}
		for _, r := range m.readings[id] {
			ts := r.RecordedAt.UTC()
			if latest == nil || ts.After(*latest) {
				cp := ts
				latest = &cp
			}
		}
	}
	return latest, nil
}

func (m *memRepo) GetSnapshotReadings(_ context.Context, targetTime time.Time) ([]domain.SnapshotNodeReading, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.SnapshotNodeReading
	start := targetTime.UTC().Add(-15 * time.Minute)
	end := targetTime.UTC()

	for id, n := range m.nodes {
		if n.Status != domain.NodeStatusActive {
			continue
		}
		var best *domain.TelemetryReading
		for i := range m.readings[id] {
			r := &m.readings[id][i]
			ts := r.RecordedAt.UTC()
			if ts.Before(start) || ts.After(end) {
				continue
			}
			if best == nil || ts.After(best.RecordedAt) {
				best = r
			}
		}
		if best != nil {
			out = append(out, domain.SnapshotNodeReading{
				NodeID:     n.ID,
				Code:       n.Code,
				Lat:        n.Lat,
				Lon:        n.Lon,
				RecordedAt: best.RecordedAt.UTC(),
				Metrics:    best.Metrics,
			})
		}
	}
	return out, nil
}

func (m *memRepo) GetPlaybackTimestamps(_ context.Context) ([]time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	seen := make(map[int64]time.Time)
	for id, n := range m.nodes {
		if n.Status != domain.NodeStatusActive {
			continue
		}
		for _, r := range m.readings[id] {
			ts := r.RecordedAt.UTC()
			seen[ts.Unix()] = ts
		}
	}
	keys := make([]int64, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	out := make([]time.Time, 0, len(keys))
	for _, k := range keys {
		out = append(out, seen[k])
	}
	return out, nil
}

func (m *memRepo) GetActiveNodesReadingsWindow(_ context.Context, from, to time.Time) ([]domain.SnapshotNodeReading, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []domain.SnapshotNodeReading
	for id, n := range m.nodes {
		if n.Status != domain.NodeStatusActive {
			continue
		}
		for _, r := range m.readings[id] {
			ts := r.RecordedAt.UTC()
			if ts.Before(from.UTC()) || ts.After(to.UTC()) {
				continue
			}
			out = append(out, domain.SnapshotNodeReading{
				NodeID:     n.ID,
				Code:       n.Code,
				Lat:        n.Lat,
				Lon:        n.Lon,
				RecordedAt: ts,
				Metrics:    r.Metrics,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RecordedAt.Before(out[j].RecordedAt) })
	return out, nil
}

func setupTestRouter(t *testing.T) (http.Handler, *memRepo) {
	t.Helper()
	phrase := "test-secret-phrase"
	salt := "616972655f7472756a696c6c6f5f73616c74"
	hash, err := auth.DeriveHashHex(phrase, salt)
	if err != nil {
		t.Fatalf("DeriveHashHex error: %v", err)
	}
	authenticator, err := auth.NewAuthenticator(hash, salt, "test-hmac-secret")
	if err != nil {
		t.Fatalf("NewAuthenticator error: %v", err)
	}

	repo := newMemRepo()
	memCache := cache.NewMemoryCache(time.Minute)
	limiter := auth.NewIPRateLimiter(10, time.Minute)
	h := NewHandler(repo, memCache, authenticator, limiter, spatial.DefaultConfig(), false)
	router := NewRouter(h, []string{"http://localhost:5173"})
	return router, repo
}

func TestEndToEndFlowThroughAPI(t *testing.T) {
	router, _ := setupTestRouter(t)

	// 1. Health endpoint
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on health, got %d", rec.Code)
	}

	// 2. Empty heatmap returns bounds=null and cells=[]
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/map/heatmap", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on empty heatmap, got %d", rec.Code)
	}
	var emptyHM domain.HeatmapResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &emptyHM); err != nil {
		t.Fatalf("unmarshal heatmap: %v", err)
	}
	if emptyHM.Bounds != nil || len(emptyHM.Cells) != 0 {
		t.Fatalf("expected nil bounds and empty cells, got %+v", emptyHM)
	}

	// 3. Unauthorized admin access returns 401
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/nodes", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on unauth admin, got %d", rec.Code)
	}

	// 4. Verify passphrase issues session cookie
	verifyBody := bytes.NewBufferString(`{"phrase":"test-secret-phrase"}`)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/access/verify", verifyBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on access verify, got %d", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	sessionCookie := cookies[0]

	// 5. Create node as admin
	createBody := bytes.NewBufferString(`{"code":"TRJ-01","name":"Plaza de Armas","lat":-8.1116,"lon":-79.0288,"status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/nodes", createBody)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create node, got %d: %s", rec.Code, rec.Body.String())
	}
	var created domain.Node
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal created node: %v", err)
	}

	// 6. Upload simulated 5-min batch
	batchBody := bytes.NewBufferString(`{
		"start_time":"2026-11-13T16:00:00-05:00",
		"metrics":{"pm25":[18.4,17.9,16.5],"pm10":[34.0,33.1,30.8]}
	}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/nodes/"+created.ID+"/readings/batch", batchBody)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on batch upload, got %d: %s", rec.Code, rec.Body.String())
	}

	// 7. Query point inside coverage
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/map/point?lat=-8.1116&lon=-79.0288&layer=air_quality&standard=inca", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on point query, got %d", rec.Code)
	}
	var ptIn domain.PointResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &ptIn); err != nil {
		t.Fatalf("unmarshal point: %v", err)
	}
	if !ptIn.HasData || ptIn.Current == nil || len(ptIn.Series) != 3 {
		t.Fatalf("expected valid point data with 3 series points, got %+v", ptIn)
	}

	// 8. Query point far outside coverage returns No hay datos (never 0)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/map/point?lat=-8.1500&lon=-79.1300", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on out-of-coverage point, got %d", rec.Code)
	}
	var ptOut domain.PointResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &ptOut); err != nil {
		t.Fatalf("unmarshal out-of-coverage point: %v", err)
	}
	if ptOut.HasData || ptOut.Label != domain.NoDataLabel || ptOut.Current != nil || len(ptOut.Series) != 0 {
		t.Fatalf("expected explicit No hay datos state, got %+v", ptOut)
	}
}
