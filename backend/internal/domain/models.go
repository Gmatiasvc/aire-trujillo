package domain

import (
	"fmt"
	"math"
	"time"
)

const (
	LayerAirQuality = "air_quality"

	MetricPM25         = "pm25"
	MetricPM10         = "pm10"
	MetricCO           = "co"
	MetricNO2          = "no2"
	MetricSO2          = "so2"
	MetricH2S          = "h2s"
	MetricCO2          = "co2"
	MetricVOCIndex     = "voc_index"
	MetricNoiseDBA     = "noise_dba"
	MetricTemperatureC = "temperature_c"
	MetricHumidityPct  = "humidity_pct"
	MetricPressureHPA  = "pressure_hpa"

	StandardINCA = "inca"
	StandardEPA  = "epa"

	NoDataColor = "#9CA3AF"
	NoDataLabel = "No hay datos"

	NodeStatusActive   = "active"
	NodeStatusInactive = "inactive"

	SourceMQTT     = "mqtt"
	SourceAdminSim = "admin_sim"
)

// MetricSpec defines the canonical unit and plausible physical range for a measured variable.
type MetricSpec struct {
	Key  string
	Unit string
	Min  float64
	Max  float64
}

var MetricSpecs = map[string]MetricSpec{
	MetricPM25:         {Key: MetricPM25, Unit: "µg/m³", Min: 0, Max: 1000},
	MetricPM10:         {Key: MetricPM10, Unit: "µg/m³", Min: 0, Max: 1000},
	MetricCO:           {Key: MetricCO, Unit: "mg/m³", Min: 0, Max: 100},
	MetricNO2:          {Key: MetricNO2, Unit: "µg/m³", Min: 0, Max: 2000},
	MetricSO2:          {Key: MetricSO2, Unit: "µg/m³", Min: 0, Max: 2000},
	MetricH2S:          {Key: MetricH2S, Unit: "µg/m³", Min: 0, Max: 2000},
	MetricCO2:          {Key: MetricCO2, Unit: "ppm", Min: 400, Max: 10000},
	MetricVOCIndex:     {Key: MetricVOCIndex, Unit: "index", Min: 1, Max: 500},
	MetricNoiseDBA:     {Key: MetricNoiseDBA, Unit: "dBA", Min: 20, Max: 130},
	MetricTemperatureC: {Key: MetricTemperatureC, Unit: "°C", Min: -10, Max: 60},
	MetricHumidityPct:  {Key: MetricHumidityPct, Unit: "%", Min: 0, Max: 100},
	MetricPressureHPA:  {Key: MetricPressureHPA, Unit: "hPa", Min: 800, Max: 1100},
}

// AllMetrics lists all 12 canonical sensor metric keys in deterministic order.
var AllMetrics = []string{
	MetricPM25,
	MetricPM10,
	MetricCO,
	MetricNO2,
	MetricSO2,
	MetricH2S,
	MetricCO2,
	MetricVOCIndex,
	MetricNoiseDBA,
	MetricTemperatureC,
	MetricHumidityPct,
	MetricPressureHPA,
}

// IsValidMetric checks whether metric matches one of the 12 canonical variables.
func IsValidMetric(metric string) bool {
	_, ok := MetricSpecs[metric]
	return ok
}

// IsValidLayer checks whether layer is "air_quality" or one of the 12 variables.
func IsValidLayer(layer string) bool {
	if layer == LayerAirQuality {
		return true
	}
	return IsValidMetric(layer)
}

// IsValidStandard checks whether standard is "inca" or "epa".
func IsValidStandard(standard string) bool {
	return standard == StandardINCA || standard == StandardEPA
}

// UnitForLayer returns the display unit for the requested layer and standard.
func UnitForLayer(layer, standard string) string {
	if layer == LayerAirQuality {
		if standard == StandardEPA {
			return "AQI"
		}
		return "INCA"
	}
	if spec, ok := MetricSpecs[layer]; ok {
		return spec.Unit
	}
	return ""
}

// ValidateMetricValue validates that val is finite and within the plausible sensor range.
func ValidateMetricValue(metric string, val float64) error {
	spec, ok := MetricSpecs[metric]
	if !ok {
		return fmt.Errorf("unknown metric %q", metric)
	}
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return fmt.Errorf("metric %q has non-finite value", metric)
	}
	if val < spec.Min || val > spec.Max {
		return fmt.Errorf("metric %q value %v out of plausible range [%v, %v]", metric, val, spec.Min, spec.Max)
	}
	return nil
}

// ErrorResponse is the standard JSON error payload.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HealthResponse represents the healthcheck response.
type HealthResponse struct {
	Status string `json:"status"`
}

// Node represents a full monitoring station record (used in admin endpoints).
type Node struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Status    string    `json:"status"`
	IsVirtual bool      `json:"is_virtual"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PublicNode represents an active station exposed on public map endpoints.
type PublicNode struct {
	ID     string  `json:"id"`
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Status string  `json:"status"`
}

// AdminNodesResponse wraps admin node listings.
type AdminNodesResponse struct {
	Nodes []Node `json:"nodes"`
}

// PublicNodesResponse wraps public active node listings.
type PublicNodesResponse struct {
	Nodes []PublicNode `json:"nodes"`
}

// CreateNodeRequest is the payload for creating a virtual node.
type CreateNodeRequest struct {
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Status string  `json:"status"`
}

// UpdateNodeRequest is the payload for partially updating a node.
type UpdateNodeRequest struct {
	Name   *string  `json:"name"`
	Lat    *float64 `json:"lat"`
	Lon    *float64 `json:"lon"`
	Status *string  `json:"status"`
}

// TelemetryMetrics holds optional readings for all 12 canonical environmental variables.
type TelemetryMetrics struct {
	PM25         *float64 `json:"pm25,omitempty"`
	PM10         *float64 `json:"pm10,omitempty"`
	CO           *float64 `json:"co,omitempty"`
	NO2          *float64 `json:"no2,omitempty"`
	SO2          *float64 `json:"so2,omitempty"`
	H2S          *float64 `json:"h2s,omitempty"`
	CO2          *float64 `json:"co2,omitempty"`
	VOCIndex     *float64 `json:"voc_index,omitempty"`
	NoiseDBA     *float64 `json:"noise_dba,omitempty"`
	TemperatureC *float64 `json:"temperature_c,omitempty"`
	HumidityPct  *float64 `json:"humidity_pct,omitempty"`
	PressureHPA  *float64 `json:"pressure_hpa,omitempty"`
}

// GetMetric returns the pointer to the requested metric if supported.
func (m TelemetryMetrics) GetMetric(name string) (*float64, bool) {
	switch name {
	case MetricPM25:
		return m.PM25, true
	case MetricPM10:
		return m.PM10, true
	case MetricCO:
		return m.CO, true
	case MetricNO2:
		return m.NO2, true
	case MetricSO2:
		return m.SO2, true
	case MetricH2S:
		return m.H2S, true
	case MetricCO2:
		return m.CO2, true
	case MetricVOCIndex:
		return m.VOCIndex, true
	case MetricNoiseDBA:
		return m.NoiseDBA, true
	case MetricTemperatureC:
		return m.TemperatureC, true
	case MetricHumidityPct:
		return m.HumidityPct, true
	case MetricPressureHPA:
		return m.PressureHPA, true
	default:
		return nil, false
	}
}

// SetMetric assigns a pointer value to the named metric field.
func (m *TelemetryMetrics) SetMetric(name string, val *float64) error {
	switch name {
	case MetricPM25:
		m.PM25 = val
	case MetricPM10:
		m.PM10 = val
	case MetricCO:
		m.CO = val
	case MetricNO2:
		m.NO2 = val
	case MetricSO2:
		m.SO2 = val
	case MetricH2S:
		m.H2S = val
	case MetricCO2:
		m.CO2 = val
	case MetricVOCIndex:
		m.VOCIndex = val
	case MetricNoiseDBA:
		m.NoiseDBA = val
	case MetricTemperatureC:
		m.TemperatureC = val
	case MetricHumidityPct:
		m.HumidityPct = val
	case MetricPressureHPA:
		m.PressureHPA = val
	default:
		return fmt.Errorf("unknown metric %q", name)
	}
	return nil
}

// Validate checks every non-nil metric value against its plausible physical range.
func (m TelemetryMetrics) Validate() error {
	for _, key := range AllMetrics {
		ptr, _ := m.GetMetric(key)
		if ptr != nil {
			if err := ValidateMetricValue(key, *ptr); err != nil {
				return err
			}
		}
	}
	return nil
}

// HasAny reports whether at least one metric field is non-nil.
func (m TelemetryMetrics) HasAny() bool {
	for _, key := range AllMetrics {
		ptr, _ := m.GetMetric(key)
		if ptr != nil {
			return true
		}
	}
	return false
}

// MQTTTelemetryPayload matches the JSON contract on aire-trujillo/v1/nodes/{node_code}/telemetry.
type MQTTTelemetryPayload struct {
	NodeCode   string           `json:"node_code"`
	RecordedAt time.Time        `json:"recorded_at"`
	Metrics    TelemetryMetrics `json:"metrics"`
}

// TelemetryReading represents a persisted reading row for a node.
type TelemetryReading struct {
	ID         int64            `json:"id"`
	NodeID     string           `json:"node_id"`
	RecordedAt time.Time        `json:"recorded_at"`
	Metrics    TelemetryMetrics `json:"metrics"`
	Source     string           `json:"source"`
	CreatedAt  time.Time        `json:"created_at"`
}

// SnapshotNodeReading binds an active node's geographic coordinates to its latest snapshot metrics.
type SnapshotNodeReading struct {
	NodeID     string
	Code       string
	Lat        float64
	Lon        float64
	RecordedAt time.Time
	Metrics    TelemetryMetrics
}

// AdminReadingItem formats a reading for the admin inspection endpoint in America/Lima (-05:00).
type AdminReadingItem struct {
	RecordedAt   string   `json:"recorded_at"`
	PM25         *float64 `json:"pm25,omitempty"`
	PM10         *float64 `json:"pm10,omitempty"`
	CO           *float64 `json:"co,omitempty"`
	NO2          *float64 `json:"no2,omitempty"`
	SO2          *float64 `json:"so2,omitempty"`
	H2S          *float64 `json:"h2s,omitempty"`
	CO2          *float64 `json:"co2,omitempty"`
	VOCIndex     *float64 `json:"voc_index,omitempty"`
	NoiseDBA     *float64 `json:"noise_dba,omitempty"`
	TemperatureC *float64 `json:"temperature_c,omitempty"`
	HumidityPct  *float64 `json:"humidity_pct,omitempty"`
	PressureHPA  *float64 `json:"pressure_hpa,omitempty"`
}

// AdminReadingsResponse wraps readings returned for an admin node inspection.
type AdminReadingsResponse struct {
	NodeID   string             `json:"node_id"`
	Readings []AdminReadingItem `json:"readings"`
}

// Bounds represents the geographic bounding box of an interpolated heatmap grid.
type Bounds struct {
	MinLat float64 `json:"min_lat"`
	MinLon float64 `json:"min_lon"`
	MaxLat float64 `json:"max_lat"`
	MaxLon float64 `json:"max_lon"`
}

// HeatmapCell represents a single 50m grid cell ready for rendering.
type HeatmapCell struct {
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	Value      *float64 `json:"value"`
	Category   string   `json:"category"`
	Color      string   `json:"color"`
	Confidence float64  `json:"confidence"`
	HasData    bool     `json:"has_data"`
}

// HeatmapResponse matches GET /api/v1/map/heatmap.
type HeatmapResponse struct {
	RecordedAt *string       `json:"recorded_at"`
	Layer      string        `json:"layer"`
	Standard   string        `json:"standard"`
	Unit       string        `json:"unit"`
	CellSizeM  float64       `json:"cell_size_m"`
	Bounds     *Bounds       `json:"bounds"`
	Cells      []HeatmapCell `json:"cells"`
}

// PointCurrent represents the current interpolated state at a clicked coordinate.
type PointCurrent struct {
	RecordedAt string  `json:"recorded_at"`
	Value      float64 `json:"value"`
	Category   string  `json:"category"`
	Color      string  `json:"color"`
	Confidence float64 `json:"confidence"`
}

// PointSeriesItem represents a single 5-minute interpolated entry in a 24h point series.
type PointSeriesItem struct {
	RecordedAt string   `json:"recorded_at"`
	Value      *float64 `json:"value"`
	HasData    bool     `json:"has_data"`
}

// PointResponse matches GET /api/v1/map/point.
type PointResponse struct {
	Lat      float64           `json:"lat"`
	Lon      float64           `json:"lon"`
	Layer    string            `json:"layer"`
	Standard string            `json:"standard"`
	Unit     string            `json:"unit"`
	HasData  bool              `json:"has_data"`
	Label    string            `json:"label"`
	Current  *PointCurrent     `json:"current"`
	Series   []PointSeriesItem `json:"series"`
}

// PlaybackTimestampsResponse matches GET /api/v1/map/playback/timestamps.
type PlaybackTimestampsResponse struct {
	Timestamps []string `json:"timestamps"`
}

// AccessVerifyRequest matches POST /api/v1/access/verify.
type AccessVerifyRequest struct {
	Phrase string `json:"phrase"`
}

// AccessResponse matches POST /api/v1/access/verify and GET /api/v1/access/session.
type AccessResponse struct {
	OK bool `json:"ok"`
}

// BatchReadingsRequest matches POST /api/v1/admin/nodes/{id}/readings/batch.
type BatchReadingsRequest struct {
	StartTime string               `json:"start_time"`
	Metrics   map[string][]float64 `json:"metrics"`
}

// BatchReadingsResponse matches POST /api/v1/admin/nodes/{id}/readings/batch response.
type BatchReadingsResponse struct {
	InsertedOrUpdated int    `json:"inserted_or_updated"`
	From              string `json:"from"`
	To                string `json:"to"`
}
