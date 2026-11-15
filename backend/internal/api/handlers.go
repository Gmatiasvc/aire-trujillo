package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/aqi"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/auth"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/cache"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/simulator"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/spatial"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/store"
)

// Handler groups HTTP handlers and dependencies for the REST API.
type Handler struct {
	repo         store.Repository
	cache        *cache.MemoryCache
	auth         *auth.Authenticator
	limiter      *auth.IPRateLimiter
	spatialCfg   spatial.Config
	cookieSecure bool
}

// NewHandler builds an API Handler with all required dependencies.
func NewHandler(
	repo store.Repository,
	memCache *cache.MemoryCache,
	authenticator *auth.Authenticator,
	limiter *auth.IPRateLimiter,
	spatialCfg spatial.Config,
	cookieSecure bool,
) *Handler {
	return &Handler{
		repo:         repo,
		cache:        memCache,
		auth:         authenticator,
		limiter:      limiter,
		spatialCfg:   spatialCfg,
		cookieSecure: cookieSecure,
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, domain.ErrorResponse{
		Error:   code,
		Message: message,
	})
}

func (h *Handler) isSecureCookie(r *http.Request) bool {
	return h.cookieSecure || r.TLS != nil
}

// Health handles GET /health and GET /api/v1/health.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, domain.HealthResponse{Status: "ok"})
}

// ListPublicNodes handles GET /api/v1/nodes returning active monitoring nodes.
func (h *Handler) ListPublicNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.ListActiveNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load nodes")
		return
	}
	if nodes == nil {
		nodes = []domain.PublicNode{}
	}
	writeJSON(w, http.StatusOK, domain.PublicNodesResponse{Nodes: nodes})
}

// GetHeatmap handles GET /api/v1/map/heatmap.
func (h *Handler) GetHeatmap(w http.ResponseWriter, r *http.Request) {
	layer, standard, ok := parseLayerAndStandard(w, r)
	if !ok {
		return
	}

	atRaw := strings.TrimSpace(r.URL.Query().Get("at"))
	var (
		requestedAt *time.Time
		atKey       = "latest"
	)
	if atRaw != "" {
		parsed, err := time.Parse(time.RFC3339, atRaw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid 'at' timestamp, expected RFC3339")
			return
		}
		utc := parsed.UTC()
		requestedAt = &utc
		atKey = utc.Format(time.RFC3339)
	}

	cacheKey := fmt.Sprintf("%s|%s|%s", layer, standard, atKey)
	if h.cache != nil {
		if cached, hit := h.cache.GetHeatmap(cacheKey); hit {
			writeJSON(w, http.StatusOK, cached)
			return
		}
	}

	unit := domain.UnitForLayer(layer, standard)
	var targetTime *time.Time
	if requestedAt != nil {
		targetTime = requestedAt
	} else {
		latest, err := h.repo.GetLatestTimestamp(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to resolve latest snapshot")
			return
		}
		targetTime = latest
	}

	if targetTime == nil {
		resp := &domain.HeatmapResponse{
			RecordedAt: nil,
			Layer:      layer,
			Standard:   standard,
			Unit:       unit,
			CellSizeM:  h.spatialCfg.CellSizeM,
			Bounds:     nil,
			Cells:      []domain.HeatmapCell{},
		}
		if h.cache != nil {
			h.cache.SetHeatmap(cacheKey, resp)
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	snapshots, err := h.repo.GetSnapshotReadings(r.Context(), *targetTime)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query snapshot readings")
		return
	}

	samples := extractLayerSamples(snapshots, layer, standard)
	recStr := targetTime.UTC().Format(time.RFC3339)

	if len(samples) == 0 {
		resp := &domain.HeatmapResponse{
			RecordedAt: &recStr,
			Layer:      layer,
			Standard:   standard,
			Unit:       unit,
			CellSizeM:  h.spatialCfg.CellSizeM,
			Bounds:     nil,
			Cells:      []domain.HeatmapCell{},
		}
		if h.cache != nil {
			h.cache.SetHeatmap(cacheKey, resp)
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	bounds, cells := spatial.GenerateHeatmapGrid(samples, layer, standard, h.spatialCfg)
	resp := &domain.HeatmapResponse{
		RecordedAt: &recStr,
		Layer:      layer,
		Standard:   standard,
		Unit:       unit,
		CellSizeM:  h.spatialCfg.CellSizeM,
		Bounds:     bounds,
		Cells:      cells,
	}
	if h.cache != nil {
		h.cache.SetHeatmap(cacheKey, resp)
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetPoint handles GET /api/v1/map/point.
func (h *Handler) GetPoint(w http.ResponseWriter, r *http.Request) {
	latRaw := strings.TrimSpace(r.URL.Query().Get("lat"))
	lonRaw := strings.TrimSpace(r.URL.Query().Get("lon"))
	if latRaw == "" || lonRaw == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "lat and lon query parameters are required")
		return
	}

	lat, err1 := strconv.ParseFloat(latRaw, 64)
	lon, err2 := strconv.ParseFloat(lonRaw, 64)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "lat and lon must be valid numbers")
		return
	}
	if err := simulator.ValidateCoordinates(lat, lon); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	layer, standard, ok := parseLayerAndStandard(w, r)
	if !ok {
		return
	}
	unit := domain.UnitForLayer(layer, standard)

	noDataResp := domain.PointResponse{
		Lat:      lat,
		Lon:      lon,
		Layer:    layer,
		Standard: standard,
		Unit:     unit,
		HasData:  false,
		Label:    domain.NoDataLabel,
		Current:  nil,
		Series:   []domain.PointSeriesItem{},
	}

	latest, err := h.repo.GetLatestTimestamp(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query latest timestamp")
		return
	}
	if latest == nil {
		writeJSON(w, http.StatusOK, noDataResp)
		return
	}

	currentSnapshots, err := h.repo.GetSnapshotReadings(r.Context(), *latest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query snapshot readings")
		return
	}

	currentSamples := extractLayerSamples(currentSnapshots, layer, standard)
	currentCell := spatial.InterpolatePoint(lat, lon, currentSamples, layer, standard, h.spatialCfg)
	if !currentCell.HasData || currentCell.Value == nil {
		writeJSON(w, http.StatusOK, noDataResp)
		return
	}

	timestamps, err := h.repo.GetPlaybackTimestamps(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query playback timestamps")
		return
	}
	if len(timestamps) == 0 {
		timestamps = []time.Time{*latest}
	}

	windowFrom := timestamps[0].Add(-15 * time.Minute)
	windowTo := timestamps[len(timestamps)-1]
	if latest.After(windowTo) {
		windowTo = *latest
	}

	windowReadings, err := h.repo.GetActiveNodesReadingsWindow(r.Context(), windowFrom, windowTo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to query 24h readings window")
		return
	}

	series := buildInterpolatedPointSeries(lat, lon, layer, standard, h.spatialCfg, timestamps, windowReadings)
	resp := domain.PointResponse{
		Lat:      lat,
		Lon:      lon,
		Layer:    layer,
		Standard: standard,
		Unit:     unit,
		HasData:  true,
		Label:    currentCell.Category,
		Current: &domain.PointCurrent{
			RecordedAt: latest.UTC().Format(time.RFC3339),
			Value:      *currentCell.Value,
			Category:   currentCell.Category,
			Color:      currentCell.Color,
			Confidence: currentCell.Confidence,
		},
		Series: series,
	}
	writeJSON(w, http.StatusOK, resp)
}

func buildInterpolatedPointSeries(
	lat, lon float64,
	layer, standard string,
	cfg spatial.Config,
	timestamps []time.Time,
	readings []domain.SnapshotNodeReading,
) []domain.PointSeriesItem {
	series := make([]domain.PointSeriesItem, 0, len(timestamps))
	latestByNode := make(map[string]domain.SnapshotNodeReading)
	rIdx := 0

	for _, ts := range timestamps {
		tsUTC := ts.UTC()
		for rIdx < len(readings) && !readings[rIdx].RecordedAt.After(tsUTC) {
			r := readings[rIdx]
			latestByNode[r.NodeID] = r
			rIdx++
		}

		samples := make([]spatial.SamplePoint, 0, len(latestByNode))
		for _, r := range latestByNode {
			if tsUTC.Sub(r.RecordedAt) > 15*time.Minute {
				continue
			}
			if val, ok := aqi.NodeLayerValue(r.Metrics, layer, standard); ok {
				samples = append(samples, spatial.SamplePoint{
					Lat:   r.Lat,
					Lon:   r.Lon,
					Value: val,
				})
			}
		}

		cell := spatial.InterpolatePoint(lat, lon, samples, layer, standard, cfg)
		series = append(series, domain.PointSeriesItem{
			RecordedAt: tsUTC.Format(time.RFC3339),
			Value:      cell.Value,
			HasData:    cell.HasData,
		})
	}
	return series
}

// GetPlaybackTimestamps handles GET /api/v1/map/playback/timestamps.
func (h *Handler) GetPlaybackTimestamps(w http.ResponseWriter, r *http.Request) {
	if h.cache != nil {
		if cached, ok := h.cache.GetPlaybackTimestamps(); ok {
			writeJSON(w, http.StatusOK, domain.PlaybackTimestampsResponse{Timestamps: cached})
			return
		}
	}

	tsList, err := h.repo.GetPlaybackTimestamps(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load playback timestamps")
		return
	}

	formatted := make([]string, 0, len(tsList))
	for _, t := range tsList {
		formatted = append(formatted, t.UTC().Format(time.RFC3339))
	}

	if h.cache != nil {
		h.cache.SetPlaybackTimestamps(formatted)
	}
	writeJSON(w, http.StatusOK, domain.PlaybackTimestampsResponse{Timestamps: formatted})
}

// VerifyAccess handles POST /api/v1/access/verify with rate limiting and generic 401 responses.
func (h *Handler) VerifyAccess(w http.ResponseWriter, r *http.Request) {
	clientIP := auth.ExtractClientIP(r)
	if h.limiter != nil && !h.limiter.Allow(clientIP) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req domain.AccessVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusUnauthorized, domain.AccessResponse{OK: false})
		return
	}

	if !h.auth.VerifyPassphrase(req.Phrase) {
		writeJSON(w, http.StatusUnauthorized, domain.AccessResponse{OK: false})
		return
	}

	h.auth.IssueSessionCookie(w, h.isSecureCookie(r))
	writeJSON(w, http.StatusOK, domain.AccessResponse{OK: true})
}

// CheckSession handles GET /api/v1/access/session.
func (h *Handler) CheckSession(w http.ResponseWriter, r *http.Request) {
	if !h.auth.ValidateSessionRequest(r) {
		writeJSON(w, http.StatusUnauthorized, domain.AccessResponse{OK: false})
		return
	}
	writeJSON(w, http.StatusOK, domain.AccessResponse{OK: true})
}

// Logout handles POST /api/v1/access/logout.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.auth.ClearSessionCookie(w, h.isSecureCookie(r))
	writeJSON(w, http.StatusOK, domain.AccessResponse{OK: true})
}

// ListAdminNodes handles GET /api/v1/admin/nodes.
func (h *Handler) ListAdminNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.ListAllNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to load nodes")
		return
	}
	if nodes == nil {
		nodes = []domain.Node{}
	}
	writeJSON(w, http.StatusOK, domain.AdminNodesResponse{Nodes: nodes})
}

// CreateAdminNode handles POST /api/v1/admin/nodes.
func (h *Handler) CreateAdminNode(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON payload")
		return
	}

	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" {
		req.Code = generateVirtualNodeCode()
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	if req.Status == "" {
		req.Status = domain.NodeStatusActive
	}

	if err := simulator.ValidateNodeInput(req.Code, req.Name, req.Status, req.Lat, req.Lon); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	node, err := h.repo.CreateNode(r.Context(), req, true)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusBadRequest, "invalid_request", "node code already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to create node")
		return
	}

	if h.cache != nil {
		h.cache.InvalidateAll()
	}
	writeJSON(w, http.StatusCreated, node)
}

// UpdateAdminNode handles PATCH /api/v1/admin/nodes/{id}.
func (h *Handler) UpdateAdminNode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "node id is required")
		return
	}

	var req domain.UpdateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON payload")
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" || len(trimmed) > 120 {
			writeError(w, http.StatusBadRequest, "invalid_request", "name must be between 1 and 120 characters")
			return
		}
		req.Name = &trimmed
	}
	if req.Status != nil {
		trimmed := strings.TrimSpace(*req.Status)
		if trimmed != domain.NodeStatusActive && trimmed != domain.NodeStatusInactive {
			writeError(w, http.StatusBadRequest, "invalid_request", "status must be active or inactive")
			return
		}
		req.Status = &trimmed
	}
	if req.Lat != nil {
		if math.IsNaN(*req.Lat) || math.IsInf(*req.Lat, 0) || *req.Lat < -90 || *req.Lat > 90 {
			writeError(w, http.StatusBadRequest, "invalid_request", "latitude must be in [-90, 90]")
			return
		}
	}
	if req.Lon != nil {
		if math.IsNaN(*req.Lon) || math.IsInf(*req.Lon, 0) || *req.Lon < -180 || *req.Lon > 180 {
			writeError(w, http.StatusBadRequest, "invalid_request", "longitude must be in [-180, 180]")
			return
		}
	}

	updated, err := h.repo.UpdateNode(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "node not found")
		case errors.Is(err, store.ErrInvalidID):
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid node id")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to update node")
		}
		return
	}

	if h.cache != nil {
		h.cache.InvalidateAll()
	}
	writeJSON(w, http.StatusOK, updated)
}

// DeleteAdminNode handles DELETE /api/v1/admin/nodes/{id}.
func (h *Handler) DeleteAdminNode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "node id is required")
		return
	}

	err := h.repo.DeleteNode(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "node not found")
		case errors.Is(err, store.ErrInvalidID):
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid node id")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to delete node")
		}
		return
	}

	if h.cache != nil {
		h.cache.InvalidateAll()
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetAdminNodeReadings handles GET /api/v1/admin/nodes/{id}/readings?hours=24.
func (h *Handler) GetAdminNodeReadings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "node id is required")
		return
	}

	hours := 24
	if rawHours := strings.TrimSpace(r.URL.Query().Get("hours")); rawHours != "" {
		parsed, err := strconv.Atoi(rawHours)
		if err != nil || parsed < 1 || parsed > 168 {
			writeError(w, http.StatusBadRequest, "invalid_request", "hours must be an integer between 1 and 168")
			return
		}
		hours = parsed
	}

	rows, err := h.repo.GetNodeReadings(r.Context(), id, hours)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "node not found")
		case errors.Is(err, store.ErrInvalidID):
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid node id")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to fetch readings")
		}
		return
	}

	items := make([]domain.AdminReadingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.AdminReadingItem{
			RecordedAt:   row.RecordedAt.In(simulator.LimaLocation).Format(time.RFC3339),
			PM25:         row.Metrics.PM25,
			PM10:         row.Metrics.PM10,
			CO:           row.Metrics.CO,
			NO2:          row.Metrics.NO2,
			SO2:          row.Metrics.SO2,
			H2S:          row.Metrics.H2S,
			CO2:          row.Metrics.CO2,
			VOCIndex:     row.Metrics.VOCIndex,
			NoiseDBA:     row.Metrics.NoiseDBA,
			TemperatureC: row.Metrics.TemperatureC,
			HumidityPct:  row.Metrics.HumidityPct,
			PressureHPA:  row.Metrics.PressureHPA,
		})
	}

	writeJSON(w, http.StatusOK, domain.AdminReadingsResponse{
		NodeID:   id,
		Readings: items,
	})
}

// BatchInsertNodeReadings handles POST /api/v1/admin/nodes/{id}/readings/batch.
func (h *Handler) BatchInsertNodeReadings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "node id is required")
		return
	}

	if _, err := h.repo.GetNodeByID(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			writeError(w, http.StatusNotFound, "not_found", "node not found")
		case errors.Is(err, store.ErrInvalidID):
			writeError(w, http.StatusBadRequest, "invalid_request", "invalid node id")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to verify node")
		}
		return
	}

	var req domain.BatchReadingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "invalid JSON payload")
		return
	}

	readings, from, to, err := simulator.ParseAndValidateBatch(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	count, err := h.repo.UpsertReadingsBatch(r.Context(), id, readings)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to persist simulated readings")
		return
	}

	if h.cache != nil {
		h.cache.InvalidateAll()
	}

	writeJSON(w, http.StatusOK, domain.BatchReadingsResponse{
		InsertedOrUpdated: count,
		From:              from.In(simulator.LimaLocation).Format(time.RFC3339),
		To:                to.In(simulator.LimaLocation).Format(time.RFC3339),
	})
}

func parseLayerAndStandard(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	layer := strings.TrimSpace(r.URL.Query().Get("layer"))
	if layer == "" {
		layer = domain.LayerAirQuality
	}
	if !domain.IsValidLayer(layer) {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("unsupported layer %q", layer))
		return "", "", false
	}

	standard := strings.TrimSpace(r.URL.Query().Get("standard"))
	if standard == "" {
		standard = domain.StandardINCA
	}
	if !domain.IsValidStandard(standard) {
		writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("unsupported standard %q", standard))
		return "", "", false
	}
	return layer, standard, true
}

func extractLayerSamples(snapshots []domain.SnapshotNodeReading, layer, standard string) []spatial.SamplePoint {
	samples := make([]spatial.SamplePoint, 0, len(snapshots))
	for _, s := range snapshots {
		val, ok := aqi.NodeLayerValue(s.Metrics, layer, standard)
		if !ok {
			continue
		}
		samples = append(samples, spatial.SamplePoint{
			Lat:   s.Lat,
			Lon:   s.Lon,
			Value: val,
		})
	}
	return samples
}

func generateVirtualNodeCode() string {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("TRJ-V-%d", time.Now().Unix()%100000)
	}
	return "TRJ-V-" + strings.ToUpper(hex.EncodeToString(b[:]))
}
