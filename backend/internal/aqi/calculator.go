package aqi

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

type colorStop struct {
	value float64
	hex   string
}

type breakpoint struct {
	CLo float64
	CHi float64
	ILo float64
	IHi float64
}

type pollutantTable struct {
	Precision float64
	Scale     float64
	Brackets  []breakpoint
}

var incaTables = map[string]pollutantTable{
	domain.MetricPM25: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 25.0, 0, 50},
			{25.1, 50.0, 51, 100},
			{50.1, 125.0, 101, 150},
			{125.1, 250.0, 151, 300},
		},
	},
	domain.MetricPM10: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 50.0, 0, 50},
			{50.1, 100.0, 51, 100},
			{100.1, 250.0, 101, 150},
			{250.1, 420.0, 151, 300},
		},
	},
	domain.MetricNO2: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 100.0, 0, 50},
			{100.1, 200.0, 51, 100},
			{200.1, 400.0, 101, 150},
			{400.1, 600.0, 151, 300},
		},
	},
	domain.MetricSO2: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 125.0, 0, 50},
			{125.1, 250.0, 51, 100},
			{250.1, 500.0, 101, 150},
			{500.1, 800.0, 151, 300},
		},
	},
	domain.MetricCO: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 10.0, 0, 50},
			{10.1, 30.0, 51, 100},
			{30.1, 45.0, 101, 150},
			{45.1, 60.0, 151, 300},
		},
	},
	domain.MetricH2S: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 75.0, 0, 50},
			{75.1, 150.0, 51, 100},
			{150.1, 300.0, 101, 150},
			{300.1, 500.0, 151, 300},
		},
	},
}

// EPA conversion factors from canonical mass units at 25 C, 1 atm:
// co: 1 mg/m3 ~= 0.873 ppm
// no2: 1 ug/m3 ~= 0.532 ppb
// so2: 1 ug/m3 ~= 0.382 ppb
var epaTables = map[string]pollutantTable{
	domain.MetricPM25: {
		Precision: 0.1,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0.0, 9.0, 0, 50},
			{9.1, 35.4, 51, 100},
			{35.5, 55.4, 101, 150},
			{55.5, 125.4, 151, 200},
			{125.5, 225.4, 201, 300},
			{225.5, 325.4, 301, 500},
		},
	},
	domain.MetricPM10: {
		Precision: 1.0,
		Scale:     1.0,
		Brackets: []breakpoint{
			{0, 54, 0, 50},
			{55, 154, 51, 100},
			{155, 254, 101, 150},
			{255, 354, 151, 200},
			{355, 424, 201, 300},
			{425, 604, 301, 500},
		},
	},
	domain.MetricCO: {
		Precision: 0.1,
		Scale:     0.873, // mg/m3 -> ppm
		Brackets: []breakpoint{
			{0.0, 4.4, 0, 50},
			{4.5, 9.4, 51, 100},
			{9.5, 12.4, 101, 150},
			{12.5, 15.4, 151, 200},
			{15.5, 30.4, 201, 300},
			{30.5, 50.4, 301, 500},
		},
	},
	domain.MetricNO2: {
		Precision: 1.0,
		Scale:     0.532, // ug/m3 -> ppb
		Brackets: []breakpoint{
			{0, 53, 0, 50},
			{54, 100, 51, 100},
			{101, 360, 101, 150},
			{361, 649, 151, 200},
			{650, 1249, 201, 300},
			{1250, 2049, 301, 500},
		},
	},
	domain.MetricSO2: {
		Precision: 1.0,
		Scale:     0.382, // ug/m3 -> ppb
		Brackets: []breakpoint{
			{0, 35, 0, 50},
			{36, 75, 51, 100},
			{76, 185, 101, 150},
			{186, 304, 151, 200},
			{305, 604, 201, 300},
			{605, 1004, 301, 500},
		},
	},
}

func interpolateBrackets(concentration float64, table pollutantTable) float64 {
	c := concentration * table.Scale
	if math.IsNaN(c) || c < 0 {
		c = 0
	}

	brackets := table.Brackets
	const eps = 1e-9

	if c <= brackets[0].CLo+eps {
		return brackets[0].ILo
	}
	last := brackets[len(brackets)-1]
	if c >= last.CHi-eps {
		return last.IHi
	}

	if idx, ok := findBracket(c, brackets, eps); ok {
		b := brackets[idx]
		return linearSubIndex(c, b)
	}

	// Concentration fell in the discrete precision gap (e.g. between 25.0 and 25.1).
	// Round to the table precision step before matching.
	rounded := math.Round(c/table.Precision) * table.Precision
	if idx, ok := findBracket(rounded, brackets, eps); ok {
		b := brackets[idx]
		return linearSubIndex(rounded, b)
	}

	// Fallback: snap to the nearest upper bracket boundary if still between two brackets.
	for i := 0; i < len(brackets)-1; i++ {
		if c > brackets[i].CHi && c < brackets[i+1].CLo {
			if c-brackets[i].CHi <= brackets[i+1].CLo-c {
				return brackets[i].IHi
			}
			return brackets[i+1].ILo
		}
	}
	return last.IHi
}

func findBracket(c float64, brackets []breakpoint, eps float64) (int, bool) {
	for i, b := range brackets {
		if c >= b.CLo-eps && c <= b.CHi+eps {
			return i, true
		}
	}
	return -1, false
}

func linearSubIndex(c float64, b breakpoint) float64 {
	if c <= b.CLo {
		return b.ILo
	}
	if c >= b.CHi {
		return b.IHi
	}
	return ((b.IHi - b.ILo) / (b.CHi - b.CLo))*(c-b.CLo) + b.ILo
}

// CalculateSubIndex computes the sub-index Ip for a single pollutant under standard ("inca" or "epa").
func CalculateSubIndex(metric string, concentration float64, standard string) (float64, bool) {
	var tables map[string]pollutantTable
	switch standard {
	case domain.StandardEPA:
		tables = epaTables
	default:
		tables = incaTables
	}

	table, ok := tables[metric]
	if !ok {
		return 0, false
	}
	return interpolateBrackets(concentration, table), true
}

// ClassifyIndex returns the category name and base hex color for an air quality index value.
func ClassifyIndex(value float64, standard string) (string, string) {
	if standard == domain.StandardEPA {
		switch {
		case value <= 50:
			return "Good", "#00E400"
		case value <= 100:
			return "Moderate", "#FFFF00"
		case value <= 150:
			return "Unhealthy for Sensitive Groups", "#FF7E00"
		case value <= 200:
			return "Unhealthy", "#FF0000"
		case value <= 300:
			return "Very Unhealthy", "#8F3F97"
		default:
			return "Hazardous", "#7E0023"
		}
	}

	switch {
	case value <= 50:
		return "Buena", "#22C55E"
	case value <= 100:
		return "Moderada", "#EAB308"
	case value <= 150:
		return "Mala", "#F97316"
	default:
		return "Peligrosa", "#EF4444"
	}
}

// CalculateCompositeIndex computes max_p(Ip) over available regulated pollutants in m.
func CalculateCompositeIndex(m domain.TelemetryMetrics, standard string) (value float64, category string, color string, dominantPollutant string, ok bool) {
	candidates := []string{
		domain.MetricPM25,
		domain.MetricPM10,
		domain.MetricCO,
		domain.MetricNO2,
		domain.MetricSO2,
		domain.MetricH2S,
	}

	maxVal := -1.0
	for _, metric := range candidates {
		ptr, exists := m.GetMetric(metric)
		if !exists || ptr == nil {
			continue
		}
		subIdx, supported := CalculateSubIndex(metric, *ptr, standard)
		if !supported {
			continue
		}
		if !ok || subIdx > maxVal {
			maxVal = subIdx
			dominantPollutant = metric
			ok = true
		}
	}

	if !ok {
		return 0, domain.NoDataLabel, domain.NoDataColor, "", false
	}

	rounded := math.Round(maxVal*100) / 100
	category, color = ClassifyIndex(rounded, standard)
	return rounded, category, color, dominantPollutant, true
}

// NodeLayerValue extracts the scalar value for a node given the requested layer and standard.
func NodeLayerValue(m domain.TelemetryMetrics, layer, standard string) (float64, bool) {
	if layer == domain.LayerAirQuality {
		val, _, _, _, ok := CalculateCompositeIndex(m, standard)
		return val, ok
	}
	ptr, exists := m.GetMetric(layer)
	if !exists || ptr == nil {
		return 0, false
	}
	return *ptr, true
}

// ClassifyLayerValue maps a layer value (air_quality index or single metric) to category and base color.
func ClassifyLayerValue(layer, standard string, value float64) (string, string) {
	if layer == domain.LayerAirQuality {
		return ClassifyIndex(value, standard)
	}

	switch layer {
	case domain.MetricPM25:
		return classifyFourTier(value, 25, 50, 100, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricPM10:
		return classifyFourTier(value, 50, 100, 200, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricCO:
		return classifyFourTier(value, 10, 30, 45, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricNO2:
		return classifyFourTier(value, 100, 200, 400, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricSO2:
		return classifyFourTier(value, 125, 250, 500, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricH2S:
		return classifyFourTier(value, 75, 150, 300, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricCO2:
		switch {
		case value <= 1000:
			return "Normal", "#22C55E"
		case value <= 2000:
			return "Elevado", "#EAB308"
		default:
			return "Alto", "#EF4444"
		}
	case domain.MetricVOCIndex:
		return classifyFourTier(value, 100, 200, 400, "Bueno", "Moderado", "Alto", "Muy alto")
	case domain.MetricNoiseDBA:
		switch {
		case value <= 60:
			return "Confortable", "#22C55E"
		case value <= 80:
			return "Moderado", "#EAB308"
		default:
			return "Alto", "#EF4444"
		}
	case domain.MetricTemperatureC:
		switch {
		case value < 16:
			return "Frío", "#3B82F6"
		case value <= 26:
			return "Templado", "#22C55E"
		case value <= 32:
			return "Cálido", "#F97316"
		default:
			return "Caluroso", "#EF4444"
		}
	case domain.MetricHumidityPct:
		switch {
		case value < 30:
			return "Seco", "#EAB308"
		case value <= 75:
			return "Confortable", "#22C55E"
		default:
			return "Húmedo", "#3B82F6"
		}
	case domain.MetricPressureHPA:
		switch {
		case value < 1005:
			return "Baja", "#3B82F6"
		case value <= 1020:
			return "Normal", "#22C55E"
		default:
			return "Alta", "#F97316"
		}
	default:
		return domain.NoDataLabel, domain.NoDataColor
	}
}

func classifyFourTier(v, t1, t2, t3 float64, l1, l2, l3, l4 string) (string, string) {
	switch {
	case v <= t1:
		return l1, "#22C55E"
	case v <= t2:
		return l2, "#EAB308"
	case v <= t3:
		return l3, "#F97316"
	default:
		return l4, "#EF4444"
	}
}

// BlendHexColors linearly interpolates between two #RRGGBB hex colors for t in [0, 1].
func BlendHexColors(hexA, hexB string, t float64) string {
	if t <= 0.0 {
		return strings.ToUpper(hexA)
	}
	if t >= 1.0 {
		return strings.ToUpper(hexB)
	}

	ar, ag, ab, okA := parseHexRGB(hexA)
	br, bg, bb, okB := parseHexRGB(hexB)
	if !okA || !okB {
		return domain.NoDataColor
	}

	r := int(math.Round(float64(ar) + t*float64(br-ar)))
	g := int(math.Round(float64(ag) + t*float64(bg-ag)))
	b := int(math.Round(float64(ab) + t*float64(bb-ab)))

	return fmt.Sprintf("#%02X%02X%02X", clampByte(r), clampByte(g), clampByte(b))
}

func parseHexRGB(hex string) (int, int, int, bool) {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	r, err1 := strconv.ParseInt(s[0:2], 16, 64)
	g, err2 := strconv.ParseInt(s[2:4], 16, 64)
	b, err3 := strconv.ParseInt(s[4:6], 16, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return int(r), int(g), int(b), true
}

func clampByte(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

func interpolateColorStops(value float64, stops []colorStop) string {
	if len(stops) == 0 {
		return domain.NoDataColor
	}
	if value <= stops[0].value {
		return strings.ToUpper(stops[0].hex)
	}
	last := stops[len(stops)-1]
	if value >= last.value {
		return strings.ToUpper(last.hex)
	}

	for i := 0; i < len(stops)-1; i++ {
		a := stops[i]
		b := stops[i+1]
		if value <= b.value {
			span := b.value - a.value
			if span <= 0 {
				return strings.ToUpper(b.hex)
			}
			t := (value - a.value) / span
			return BlendHexColors(a.hex, b.hex, t)
		}
	}
	return strings.ToUpper(last.hex)
}

func fourTierStops(t1, t2, t3 float64) []colorStop {
	return []colorStop{
		{value: t1 * 0.5, hex: "#22C55E"},
		{value: (t1 + t2) * 0.5, hex: "#EAB308"},
		{value: (t2 + t3) * 0.5, hex: "#F97316"},
		{value: t3 * 1.35, hex: "#EF4444"},
	}
}

var (
	incaColorStops = []colorStop{
		{value: 25, hex: "#22C55E"},
		{value: 75, hex: "#EAB308"},
		{value: 125, hex: "#F97316"},
		{value: 200, hex: "#EF4444"},
	}
	epaColorStops = []colorStop{
		{value: 25, hex: "#00E400"},
		{value: 75, hex: "#FFFF00"},
		{value: 125, hex: "#FF7E00"},
		{value: 175, hex: "#FF0000"},
		{value: 250, hex: "#8F3F97"},
		{value: 400, hex: "#7E0023"},
	}
	metricColorStops = map[string][]colorStop{
		domain.MetricPM25:     fourTierStops(25, 50, 100),
		domain.MetricPM10:     fourTierStops(50, 100, 200),
		domain.MetricCO:       fourTierStops(10, 30, 45),
		domain.MetricNO2:      fourTierStops(100, 200, 400),
		domain.MetricSO2:      fourTierStops(125, 250, 500),
		domain.MetricH2S:      fourTierStops(75, 150, 300),
		domain.MetricVOCIndex: fourTierStops(100, 200, 400),
		domain.MetricCO2: {
			{value: 600, hex: "#22C55E"},
			{value: 1500, hex: "#EAB308"},
			{value: 2500, hex: "#EF4444"},
		},
		domain.MetricNoiseDBA: {
			{value: 45, hex: "#22C55E"},
			{value: 70, hex: "#EAB308"},
			{value: 90, hex: "#EF4444"},
		},
		domain.MetricTemperatureC: {
			{value: 12, hex: "#3B82F6"},
			{value: 21, hex: "#22C55E"},
			{value: 29, hex: "#F97316"},
			{value: 36, hex: "#EF4444"},
		},
		domain.MetricHumidityPct: {
			{value: 20, hex: "#EAB308"},
			{value: 52.5, hex: "#22C55E"},
			{value: 85, hex: "#3B82F6"},
		},
		domain.MetricPressureHPA: {
			{value: 998, hex: "#3B82F6"},
			{value: 1012.5, hex: "#22C55E"},
			{value: 1026, hex: "#F97316"},
		},
	}
)

// InterpolateLayerColor returns a continuous #RRGGBB color blended smoothly between band centers.
func InterpolateLayerColor(layer, standard string, value float64) string {
	if layer == domain.LayerAirQuality {
		if standard == domain.StandardEPA {
			return interpolateColorStops(value, epaColorStops)
		}
		return interpolateColorStops(value, incaColorStops)
	}

	stops, ok := metricColorStops[layer]
	if !ok {
		return domain.NoDataColor
	}
	return interpolateColorStops(value, stops)
}
