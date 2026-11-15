CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS nodes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code        VARCHAR(64) NOT NULL UNIQUE,
  name        VARCHAR(120) NOT NULL,
  status      VARCHAR(16) NOT NULL CHECK (status IN ('active', 'inactive')),
  is_virtual  BOOLEAN NOT NULL DEFAULT TRUE,
  location    GEOGRAPHY(Point, 4326) NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nodes_location ON nodes USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes (status);

CREATE TABLE IF NOT EXISTS telemetry_readings (
  id            BIGSERIAL PRIMARY KEY,
  node_id       UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  recorded_at   TIMESTAMPTZ NOT NULL,
  pm25          DOUBLE PRECISION,
  pm10          DOUBLE PRECISION,
  co            DOUBLE PRECISION,
  no2           DOUBLE PRECISION,
  so2           DOUBLE PRECISION,
  h2s           DOUBLE PRECISION,
  co2           DOUBLE PRECISION,
  voc_index     DOUBLE PRECISION,
  noise_dba     DOUBLE PRECISION,
  temperature_c DOUBLE PRECISION,
  humidity_pct  DOUBLE PRECISION,
  pressure_hpa  DOUBLE PRECISION,
  source        VARCHAR(16) NOT NULL CHECK (source IN ('mqtt', 'admin_sim')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (node_id, recorded_at)
);

CREATE INDEX IF NOT EXISTS idx_telemetry_recorded_at ON telemetry_readings (recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_telemetry_node_time ON telemetry_readings (node_id, recorded_at DESC);
