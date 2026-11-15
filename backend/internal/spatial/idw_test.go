package spatial

import (
	"math"
	"testing"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

func offsetEastMeters(lat, lon, meters float64) float64 {
	cosLat := math.Cos(lat * math.Pi / 180.0)
	return lon + meters/(111320.0*cosLat)
}

func TestInterpolatePointExactMatch(t *testing.T) {
	cfg := DefaultConfig()
	samples := []SamplePoint{
		{Lat: -8.1116, Lon: -79.0288, Value: 42.5},
		{Lat: -8.1125, Lon: -79.0295, Value: 120.0},
	}

	// Point within < 1m of first node (approx 0.5m)
	nearLon := offsetEastMeters(-8.1116, -79.0288, 0.5)
	cell := InterpolatePoint(-8.1116, nearLon, samples, domain.LayerAirQuality, domain.StandardINCA, cfg)

	if !cell.HasData || cell.Value == nil {
		t.Fatalf("expected HasData=true and non-nil Value")
	}
	if *cell.Value != 42.5 {
		t.Fatalf("expected exact node value 42.5, got %v", *cell.Value)
	}
	if cell.Confidence != 1.0 {
		t.Fatalf("expected confidence 1.0, got %v", cell.Confidence)
	}
	if cell.Category != "Buena" || cell.Color != "#22C55E" {
		t.Fatalf("got (%s, %s), want (Buena, #22C55E)", cell.Category, cell.Color)
	}
}

func TestInterpolatePointSymmetry(t *testing.T) {
	cfg := DefaultConfig()
	lat := -8.1116
	lon := -79.0288
	samples := []SamplePoint{
		{Lat: lat, Lon: offsetEastMeters(lat, lon, -200), Value: 20.0},
		{Lat: lat, Lon: offsetEastMeters(lat, lon, 200), Value: 80.0},
	}

	cell := InterpolatePoint(lat, lon, samples, domain.LayerAirQuality, domain.StandardINCA, cfg)
	if !cell.HasData || cell.Value == nil {
		t.Fatalf("expected valid interpolated cell")
	}
	if math.Abs(*cell.Value-50.0) > 0.05 {
		t.Fatalf("expected symmetric average ~50.0, got %v", *cell.Value)
	}
	if cell.Confidence != 1.0 {
		t.Fatalf("expected confidence 1.0 at 200m (<= 400m), got %v", cell.Confidence)
	}
}

func TestInterpolatePointFadeToGray(t *testing.T) {
	cfg := DefaultConfig()
	lat := -8.1116
	lon := -79.0288
	samples := []SamplePoint{
		{Lat: lat, Lon: lon, Value: 30.0}, // Buena (#22C55E -> RGB 34, 197, 94)
	}

	// Distance = 500m -> halfway between 400m and 600m -> confidence = 0.5
	queryLon := offsetEastMeters(lat, lon, 500)
	cell := InterpolatePoint(lat, queryLon, samples, domain.LayerAirQuality, domain.StandardINCA, cfg)

	if !cell.HasData || cell.Value == nil {
		t.Fatalf("expected HasData=true at 500m")
	}
	if math.Abs(cell.Confidence-0.5) > 0.01 {
		t.Fatalf("expected confidence ~0.5 at 500m, got %v", cell.Confidence)
	}
	// Blend #22C55E (34,197,94) and #9CA3AF (156,163,175) at 0.5 -> ~(95, 180, 135) -> #5FB487
	if cell.Color == "#22C55E" || cell.Color == domain.NoDataColor {
		t.Fatalf("expected blended intermediate color, got %s", cell.Color)
	}
}

func TestInterpolatePointOutOfCoverageNeverReturnsZero(t *testing.T) {
	cfg := DefaultConfig()
	lat := -8.1116
	lon := -79.0288
	samples := []SamplePoint{
		{Lat: lat, Lon: lon, Value: 40.0},
	}

	// 650m > MaxRadiusM (600m)
	farLon := offsetEastMeters(lat, lon, 650)
	cell := InterpolatePoint(lat, farLon, samples, domain.LayerAirQuality, domain.StandardINCA, cfg)

	if cell.HasData {
		t.Fatalf("expected HasData=false beyond 600m")
	}
	if cell.Value != nil {
		t.Fatalf("expected Value=nil (never 0) beyond 600m, got %v", *cell.Value)
	}
	if cell.Color != domain.NoDataColor {
		t.Fatalf("expected color %s, got %s", domain.NoDataColor, cell.Color)
	}
	if cell.Category != domain.NoDataLabel {
		t.Fatalf("expected category %q, got %q", domain.NoDataLabel, cell.Category)
	}
	if cell.Confidence != 0.0 {
		t.Fatalf("expected confidence 0.0, got %v", cell.Confidence)
	}

	// Empty samples must also return empty grid with nil bounds
	bounds, cells := GenerateHeatmapGrid(nil, domain.LayerAirQuality, domain.StandardINCA, cfg)
	if bounds != nil || len(cells) != 0 {
		t.Fatalf("expected nil bounds and empty cells for no samples")
	}
}

func TestGenerateHeatmapGridIncludesInAndOutOfCoverageCells(t *testing.T) {
	cfg := DefaultConfig()
	samples := []SamplePoint{
		{Lat: -8.1116, Lon: -79.0288, Value: 35.0},
	}

	bounds, cells := GenerateHeatmapGrid(samples, domain.LayerAirQuality, domain.StandardINCA, cfg)
	if bounds == nil {
		t.Fatalf("expected non-nil bounds")
	}
	if len(cells) == 0 {
		t.Fatalf("expected generated grid cells")
	}

	var hasDataCount, noDataCount int
	for _, c := range cells {
		if c.HasData {
			hasDataCount++
		} else {
			noDataCount++
			if c.Value != nil || c.Color != domain.NoDataColor || c.Category != domain.NoDataLabel {
				t.Fatalf("invalid out-of-coverage cell inside grid: %+v", c)
			}
		}
	}
	if hasDataCount == 0 || noDataCount == 0 {
		t.Fatalf("expected both in-coverage (%d) and corner out-of-coverage (%d) cells", hasDataCount, noDataCount)
	}
}
