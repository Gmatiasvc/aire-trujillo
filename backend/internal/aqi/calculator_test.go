package aqi

import (
	"math"
	"testing"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestCalculateSubIndexINCA(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		value    float64
		wantMin  float64
		wantMax  float64
	}{
		{name: "pm25 zero", metric: domain.MetricPM25, value: 0, wantMin: 0, wantMax: 0.001},
		{name: "pm25 buena upper", metric: domain.MetricPM25, value: 25.0, wantMin: 49.99, wantMax: 50.01},
		{name: "pm25 moderada lower", metric: domain.MetricPM25, value: 25.1, wantMin: 50.99, wantMax: 51.01},
		{name: "pm25 moderada midpoint", metric: domain.MetricPM25, value: 37.55, wantMin: 75.0, wantMax: 76.0},
		{name: "pm25 above top clamps to 300", metric: domain.MetricPM25, value: 320.0, wantMin: 299.99, wantMax: 300.01},
		{name: "h2s mala lower", metric: domain.MetricH2S, value: 150.1, wantMin: 100.99, wantMax: 101.01},
		{name: "co moderada upper", metric: domain.MetricCO, value: 30.0, wantMin: 99.99, wantMax: 100.01},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := CalculateSubIndex(tc.metric, tc.value, domain.StandardINCA)
			if !ok {
				t.Fatalf("expected metric %s to be supported in INCA", tc.metric)
			}
			if got < tc.wantMin || got > tc.wantMax {
				t.Fatalf("got %v, want in [%v, %v]", got, tc.wantMin, tc.wantMax)
			}
		})
	}
}

func TestCalculateSubIndexEPAAndConversions(t *testing.T) {
	// PM2.5 updated 2024 EPA breakpoints: 9.0 -> 50, 9.1 -> 51, 35.4 -> 100
	v, ok := CalculateSubIndex(domain.MetricPM25, 9.0, domain.StandardEPA)
	if !ok || math.Abs(v-50.0) > 1e-6 {
		t.Fatalf("EPA pm25=9.0: got (%v, %v), want (50, true)", v, ok)
	}

	// CO conversion: 1 mg/m3 * 0.873 = 0.873 ppm -> inside Good [0.0, 4.4] -> [0, 50]
	// Expected: (50/4.4)*0.873 = 9.92045...
	coVal, ok := CalculateSubIndex(domain.MetricCO, 1.0, domain.StandardEPA)
	if !ok || math.Abs(coVal-9.9204545) > 1e-4 {
		t.Fatalf("EPA co=1.0 mg/m3: got (%v, %v), want ~9.92045", coVal, ok)
	}

	// NO2 conversion: 100 ug/m3 * 0.532 = 53.2 ppb -> rounds to 53 -> 50
	no2Val, ok := CalculateSubIndex(domain.MetricNO2, 100.0, domain.StandardEPA)
	if !ok || math.Abs(no2Val-50.0) > 1e-4 {
		t.Fatalf("EPA no2=100 ug/m3: got (%v, %v), want 50", no2Val, ok)
	}

	// SO2 conversion: 100 ug/m3 * 0.382 = 38.2 ppb -> inside Moderate [36, 75] -> [51, 100]
	so2Val, ok := CalculateSubIndex(domain.MetricSO2, 100.0, domain.StandardEPA)
	wantSO2 := ((100.0 - 51.0) / (75.0 - 36.0))*(38.2-36.0) + 51.0
	if !ok || math.Abs(so2Val-wantSO2) > 1e-4 {
		t.Fatalf("EPA so2=100 ug/m3: got (%v, %v), want %v", so2Val, ok, wantSO2)
	}

	// Above top EPA PM2.5 clamps to 500
	topVal, ok := CalculateSubIndex(domain.MetricPM25, 500.0, domain.StandardEPA)
	if !ok || topVal != 500.0 {
		t.Fatalf("EPA pm25=500: got (%v, %v), want (500, true)", topVal, ok)
	}

	// H2S is not regulated in US EPA table
	if _, ok := CalculateSubIndex(domain.MetricH2S, 50.0, domain.StandardEPA); ok {
		t.Fatalf("expected H2S not to be in EPA table")
	}
}

func TestCalculateCompositeIndex(t *testing.T) {
	t.Run("empty metrics returns ok=false", func(t *testing.T) {
		_, cat, color, _, ok := CalculateCompositeIndex(domain.TelemetryMetrics{}, domain.StandardINCA)
		if ok {
			t.Fatalf("expected ok=false for empty metrics")
		}
		if cat != domain.NoDataLabel || color != domain.NoDataColor {
			t.Fatalf("unexpected fallback: cat=%q color=%q", cat, color)
		}
	})

	t.Run("critical pollutant drives composite INCA index", func(t *testing.T) {
		m := domain.TelemetryMetrics{
			PM25: floatPtr(12.5),  // INCA 25 (Buena)
			PM10: floatPtr(175.0), // INCA 125.5 (Mala)
			CO:   floatPtr(5.0),   // INCA 25 (Buena)
		}
		val, cat, color, dom, ok := CalculateCompositeIndex(m, domain.StandardINCA)
		if !ok {
			t.Fatalf("expected ok=true")
		}
		if dom != domain.MetricPM10 {
			t.Fatalf("expected dominant pollutant pm10, got %s", dom)
		}
		if cat != "Mala" || color != "#F97316" {
			t.Fatalf("got (%s, %s), want (Mala, #F97316)", cat, color)
		}
		if val <= 100 || val > 150 {
			t.Fatalf("unexpected composite value %v", val)
		}
	})

	t.Run("composite EPA index", func(t *testing.T) {
		m := domain.TelemetryMetrics{
			PM25: floatPtr(40.0), // EPA USG (> 100)
			PM10: floatPtr(40.0), // EPA Good (< 50)
		}
		val, cat, color, dom, ok := CalculateCompositeIndex(m, domain.StandardEPA)
		if !ok {
			t.Fatalf("expected ok=true")
		}
		if dom != domain.MetricPM25 {
			t.Fatalf("expected dominant pm25, got %s", dom)
		}
		if cat != "Unhealthy for Sensitive Groups" || color != "#FF7E00" {
			t.Fatalf("got (%s, %s), want (Unhealthy for Sensitive Groups, #FF7E00)", cat, color)
		}
		if val < 101 || val > 150 {
			t.Fatalf("unexpected EPA composite value %v", val)
		}
	})
}

func TestInterpolateLayerColorGradualTransitions(t *testing.T) {
	tests := []struct {
		name     string
		layer    string
		standard string
		value    float64
		want     string
	}{
		{name: "inca low clamps to Buena", layer: domain.LayerAirQuality, standard: domain.StandardINCA, value: 10, want: "#22C55E"},
		{name: "inca boundary 50 blends Buena and Moderada", layer: domain.LayerAirQuality, standard: domain.StandardINCA, value: 50, want: "#86BC33"},
		{name: "inca 75 matches Moderada center", layer: domain.LayerAirQuality, standard: domain.StandardINCA, value: 75, want: "#EAB308"},
		{name: "inca boundary 100 blends Moderada and Mala", layer: domain.LayerAirQuality, standard: domain.StandardINCA, value: 100, want: "#F2930F"},
		{name: "inca high clamps to Peligrosa", layer: domain.LayerAirQuality, standard: domain.StandardINCA, value: 240, want: "#EF4444"},
		{name: "epa low clamps to Good", layer: domain.LayerAirQuality, standard: domain.StandardEPA, value: 15, want: "#00E400"},
		{name: "epa boundary 100 blends Moderate and USG", layer: domain.LayerAirQuality, standard: domain.StandardEPA, value: 100, want: "#FFBF00"},
		{name: "epa high clamps to Hazardous", layer: domain.LayerAirQuality, standard: domain.StandardEPA, value: 450, want: "#7E0023"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := InterpolateLayerColor(tc.layer, tc.standard, tc.value)
			if got != tc.want {
				t.Fatalf("InterpolateLayerColor(%q, %q, %v) = %s, want %s", tc.layer, tc.standard, tc.value, got, tc.want)
			}
		})
	}

	pm25Boundary := InterpolateLayerColor(domain.MetricPM25, domain.StandardINCA, 25)
	if pm25Boundary == "#22C55E" || pm25Boundary == "#EAB308" {
		t.Fatalf("expected pm25=25 to blend smoothly between #22C55E and #EAB308, got %s", pm25Boundary)
	}

	tempMid := InterpolateLayerColor(domain.MetricTemperatureC, domain.StandardINCA, 16.5)
	if tempMid != "#2FA4AA" {
		t.Fatalf("expected temperature_c=16.5 midpoint between #3B82F6 and #22C55E (#2FA4AA), got %s", tempMid)
	}
}
