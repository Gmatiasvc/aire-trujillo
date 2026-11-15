package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gmatiasvc/air-quality-trujillo/backend/internal/spatial"
)

// Config holds runtime settings loaded from environment variables (and optional .env file).
type Config struct {
	Port                string
	DatabaseURL         string
	MQTTBrokerURL       string
	MQTTEnabled         bool
	AdminPassphraseSalt string
	AdminPassphraseHash string
	SessionSecret       string
	CookieSecure        bool
	CORSAllowedOrigins  []string
	MigrationsDir       string
	Spatial             spatial.Config
}

// Load reads environment variables (with local debugging defaults) and optional .env overrides.
func Load() Config {
	loadDotEnvIfPresent(".env")

	cfg := Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://air_admin:air_admin@localhost:5432/air_quality_db?sslmode=disable"),
		MQTTBrokerURL:       getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
		MQTTEnabled:         getEnvBool("MQTT_ENABLED", true),
		AdminPassphraseSalt: getEnv("ADMIN_PASSPHRASE_SALT", "616972655f7472756a696c6c6f5f73616c74"),
		AdminPassphraseHash: getEnv("ADMIN_PASSPHRASE_HASH", "11bbd278fb82bb25da2f6d17094d1cca4ed67844fc4099d23413c6b632ea8f2b"),
		SessionSecret:       getEnv("SESSION_SECRET", "aire-trujillo-dev-session-secret-change-in-prod"),
		CookieSecure:        getEnvBool("COOKIE_SECURE", false),
		CORSAllowedOrigins:  parseCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173,https://aire-trujillo.vercel.app,https://aire-trujillo-*.vercel.app")),
		MigrationsDir:       resolveMigrationsDir(getEnv("MIGRATIONS_DIR", "migrations")),
		Spatial: spatial.Config{
			CellSizeM:  getEnvFloat("INTERP_CELL_SIZE_M", 50),
			Power:      getEnvFloat("INTERP_IDW_POWER", 2),
			FadeStartM: getEnvFloat("INTERP_FADE_START_M", 400),
			MaxRadiusM: getEnvFloat("INTERP_MAX_RADIUS_M", 600),
		},
	}
	return cfg
}

func resolveMigrationsDir(preferred string) string {
	if info, err := os.Stat(preferred); err == nil && info.IsDir() {
		return preferred
	}
	alt := filepath.Join("backend", preferred)
	if info, err := os.Stat(alt); err == nil && info.IsDir() {
		return alt
	}
	return preferred
}

func loadDotEnvIfPresent(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return []string{"http://localhost:5173"}
	}
	return out
}
