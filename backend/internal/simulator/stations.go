package simulator

import (
	"fmt"
	"math"
	"strings"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/domain"
)

// ReferenceStation represents an optional reference station location in Trujillo / Huanchaco.
type ReferenceStation struct {
	Code string
	Name string
	Lat  float64
	Lon  float64
}

// ReferenceTrujilloStations returns canonical reference coordinates for Trujillo & Huanchaco.
func ReferenceTrujilloStations() []ReferenceStation {
	return []ReferenceStation{
		{Code: "TRJ-01", Name: "Plaza de Armas", Lat: -8.1116, Lon: -79.0288},
		{Code: "TRJ-02", Name: "Óvalo Papal", Lat: -8.1180, Lon: -79.0350},
		{Code: "TRJ-HUANCHACO-01", Name: "Muelle de Huanchaco", Lat: -8.0794, Lon: -79.1208},
	}
}

// ValidateNodeInput validates node creation/update attributes.
func ValidateNodeInput(code, name, status string, lat, lon float64) error {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" || len(trimmedCode) > 64 {
		return fmt.Errorf("code must be between 1 and 64 characters")
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" || len(trimmedName) > 120 {
		return fmt.Errorf("name must be between 1 and 120 characters")
	}
	if status != domain.NodeStatusActive && status != domain.NodeStatusInactive {
		return fmt.Errorf("status must be %q or %q", domain.NodeStatusActive, domain.NodeStatusInactive)
	}
	return ValidateCoordinates(lat, lon)
}

// ValidateCoordinates checks that latitude and longitude are finite WGS84 coordinates.
func ValidateCoordinates(lat, lon float64) error {
	if math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		return fmt.Errorf("latitude %v must be in [-90, 90]", lat)
	}
	if math.IsNaN(lon) || math.IsInf(lon, 0) || lon < -180 || lon > 180 {
		return fmt.Errorf("longitude %v must be in [-180, 180]", lon)
	}
	return nil
}
