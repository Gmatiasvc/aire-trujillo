package spatial

import (
	"math"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/aqi"
	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

const earthRadiusMeters = 6371000.0

// Config holds configurable IDW spatial interpolation parameters.
type Config struct {
	CellSizeM  float64
	Power      float64
	FadeStartM float64
	MaxRadiusM float64
}

// DefaultConfig returns the specification defaults from 02-domain-and-data.md.
func DefaultConfig() Config {
	return Config{
		CellSizeM:  50,
		Power:      2,
		FadeStartM: 400,
		MaxRadiusM: 600,
	}
}

// SamplePoint represents an active node measurement at a geographic location.
type SamplePoint struct {
	Lat   float64
	Lon   float64
	Value float64
}

// HaversineDistanceM computes the great-circle distance in meters between two WGS84 points.
func HaversineDistanceM(lat1, lon1, lat2, lon2 float64) float64 {
	phi1 := lat1 * math.Pi / 180.0
	phi2 := lat2 * math.Pi / 180.0
	dPhi := (lat2 - lat1) * math.Pi / 180.0
	dLambda := (lon2 - lon1) * math.Pi / 180.0

	sinDPhi := math.Sin(dPhi / 2.0)
	sinDLambda := math.Sin(dLambda / 2.0)

	a := sinDPhi*sinDPhi + math.Cos(phi1)*math.Cos(phi2)*sinDLambda*sinDLambda
	if a > 1.0 {
		a = 1.0
	}
	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
	return earthRadiusMeters * c
}

// BlendHexWithGray linearly blends a base #RRGGBB color with #9CA3AF according to confidence in [0, 1].
func BlendHexWithGray(baseHex string, confidence float64) string {
	return aqi.BlendHexColors(domain.NoDataColor, baseHex, confidence)
}

// InterpolatePoint evaluates compactly-supported Shepard IDW, directional interior confidence, and edge fade-to-gray at (lat, lon).
func InterpolatePoint(lat, lon float64, samples []SamplePoint, layer, standard string, cfg Config) domain.HeatmapCell {
	if cfg.MaxRadiusM <= 0 {
		cfg.MaxRadiusM = 600
	}
	if cfg.Power <= 0 {
		cfg.Power = 2
	}

	cosLat := math.Cos(lat * math.Pi / 180.0)

	var (
		hasMatch     bool
		exactMatch   bool
		exactVal     float64
		exactDist    float64
		nearestVal   float64
		weightSum    float64
		valueSum     float64
		confSum      float64
		dirWeightSum float64
		dirX         float64
		dirY         float64
		dMin         = math.Inf(1)
	)

	for _, s := range samples {
		d := HaversineDistanceM(lat, lon, s.Lat, s.Lon)
		if d > cfg.MaxRadiusM {
			continue
		}
		hasMatch = true
		if d < dMin {
			dMin = d
			nearestVal = s.Value
		}

		confSum += computeConfidence(d, cfg.FadeStartM, cfg.MaxRadiusM)

		// O(1)-per-sample directional balance vector: opposes directions cancel between flanked nodes
		// (I -> 1), while outer boundaries point predominantly inward toward one side (I -> 0).
		dy := (s.Lat - lat) * 111320.0
		dx := (s.Lon - lon) * 111320.0 * cosLat
		norm := math.Hypot(dx, dy)
		dirW := 1.0 - d/cfg.MaxRadiusM
		if dirW > 0 && norm > 1e-6 {
			dirWeightSum += dirW
			dirX += dirW * (dx / norm)
			dirY += dirW * (dy / norm)
		}

		if d < 1.0 {
			if !exactMatch || d < exactDist {
				exactMatch = true
				exactDist = d
				exactVal = s.Value
			}
			continue
		}
		if d < cfg.MaxRadiusM {
			taper := (cfg.MaxRadiusM - d) / (cfg.MaxRadiusM * d)
			w := math.Pow(taper, cfg.Power)
			weightSum += w
			valueSum += w * s.Value
		}
	}

	if !hasMatch {
		return domain.HeatmapCell{
			Lat:        roundCoord(lat),
			Lon:        roundCoord(lon),
			Value:      nil,
			Category:   domain.NoDataLabel,
			Color:      domain.NoDataColor,
			Confidence: 0.0,
			HasData:    false,
		}
	}

	var interpVal float64
	if exactMatch {
		interpVal = exactVal
		dMin = exactDist
	} else if weightSum > 0 {
		interpVal = valueSum / weightSum
	} else {
		interpVal = nearestVal
	}

	baseConf := computeConfidence(dMin, cfg.FadeStartM, cfg.MaxRadiusM)
	confidence := baseConf
	if !exactMatch && dirWeightSum > 0 && baseConf < 1.0 && confSum > 0 {
		rx := dirX / dirWeightSum
		ry := dirY / dirWeightSum
		rSq := rx*rx + ry*ry
		if rSq > 1.0 {
			rSq = 1.0
		}
		interiority := 1.0 - rSq
		supportEnvelope := math.Min(1.0, confSum/0.35)
		confidence = baseConf + interiority*supportEnvelope*(1.0-baseConf)
		if confidence > 1.0 {
			confidence = 1.0
		}
	}

	roundedVal := math.Round(interpVal*100) / 100
	roundedConf := math.Round(confidence*10000) / 10000

	cat, _ := aqi.ClassifyLayerValue(layer, standard, roundedVal)
	baseColor := aqi.InterpolateLayerColor(layer, standard, roundedVal)
	finalColor := BlendHexWithGray(baseColor, confidence)

	return domain.HeatmapCell{
		Lat:        roundCoord(lat),
		Lon:        roundCoord(lon),
		Value:      &roundedVal,
		Category:   cat,
		Color:      finalColor,
		Confidence: roundedConf,
		HasData:    true,
	}
}

func computeConfidence(dMin, fadeStartM, maxRadiusM float64) float64 {
	if dMin <= fadeStartM {
		return 1.0
	}
	if dMin >= maxRadiusM {
		return 0.0
	}
	span := maxRadiusM - fadeStartM
	if span <= 0 {
		return 1.0
	}
	c := 1.0 - (dMin-fadeStartM)/span
	if c < 0 {
		return 0
	}
	if c > 1 {
		return 1
	}
	return c
}

// GenerateHeatmapGrid computes the dynamic bounding box and 50m grid cells around active samples.
func GenerateHeatmapGrid(samples []SamplePoint, layer, standard string, cfg Config) (*domain.Bounds, []domain.HeatmapCell) {
	if len(samples) == 0 {
		return nil, []domain.HeatmapCell{}
	}
	if cfg.CellSizeM <= 0 {
		cfg.CellSizeM = 50
	}
	if cfg.MaxRadiusM <= 0 {
		cfg.MaxRadiusM = 600
	}

	minLat, maxLat := samples[0].Lat, samples[0].Lat
	minLon, maxLon := samples[0].Lon, samples[0].Lon
	for _, s := range samples[1:] {
		if s.Lat < minLat {
			minLat = s.Lat
		}
		if s.Lat > maxLat {
			maxLat = s.Lat
		}
		if s.Lon < minLon {
			minLon = s.Lon
		}
		if s.Lon > maxLon {
			maxLon = s.Lon
		}
	}

	meanLat := (minLat + maxLat) / 2.0
	cosLat := math.Cos(meanLat * math.Pi / 180.0)
	if math.Abs(cosLat) < 0.01 {
		cosLat = 0.01
	}

	latMargin := cfg.MaxRadiusM / 111320.0
	lonMargin := cfg.MaxRadiusM / (111320.0 * cosLat)

	bounds := &domain.Bounds{
		MinLat: roundCoord(minLat - latMargin),
		MinLon: roundCoord(minLon - lonMargin),
		MaxLat: roundCoord(maxLat + latMargin),
		MaxLon: roundCoord(maxLon + lonMargin),
	}

	dLat := cfg.CellSizeM / 111320.0
	dLon := cfg.CellSizeM / (111320.0 * cosLat)

	rows := int(math.Ceil((bounds.MaxLat - bounds.MinLat) / dLat))
	cols := int(math.Ceil((bounds.MaxLon - bounds.MinLon) / dLon))
	if rows < 1 {
		rows = 1
	}
	if cols < 1 {
		cols = 1
	}

	// Safety cap for extreme node separation on free-tier memory budgets.
	const maxGridCells = 250000
	if rows*cols > maxGridCells {
		scale := math.Sqrt(float64(rows*cols) / float64(maxGridCells))
		dLat *= scale
		dLon *= scale
		rows = int(math.Ceil((bounds.MaxLat - bounds.MinLat) / dLat))
		cols = int(math.Ceil((bounds.MaxLon - bounds.MinLon) / dLon))
	}

	cells := make([]domain.HeatmapCell, 0, rows*cols)
	for r := 0; r < rows; r++ {
		lat := bounds.MinLat + (float64(r)+0.5)*dLat
		if lat > bounds.MaxLat {
			continue
		}
		for c := 0; c < cols; c++ {
			lon := bounds.MinLon + (float64(c)+0.5)*dLon
			if lon > bounds.MaxLon {
				continue
			}
			cell := InterpolatePoint(lat, lon, samples, layer, standard, cfg)
			cells = append(cells, cell)
		}
	}

	return bounds, cells
}

func roundCoord(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}
