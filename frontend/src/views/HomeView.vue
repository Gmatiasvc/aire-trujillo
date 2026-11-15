<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import IdwCoverageProbe from '../components/IdwCoverageProbe.vue'
import IeeeReportModal from '../components/IeeeReportModal.vue'
import PcbExplorer from '../components/PcbExplorer.vue'
import UiIcon from '../components/UiIcon.vue'
import {
  DEMO_STATION_ROWS,
  formatStandardTitle,
  type DemoStationRow,
} from '../data/demoSnapshot'
import { useThemeStore } from '../stores/theme'
import {
  METRIC_KEYS,
  METRIC_SPECS,
  NO_DATA_COLOR,
  NO_DATA_LABEL,
  type AirQualityStandard,
} from '../types/telemetry'

interface StageNavItem {
  id: string
  code: string
  short: string
  label: string
  summary: string
}

const STAGES: StageNavItem[] = [
  { id: 'etapa-01', code: '01', short: 'MAPA', label: 'Mapa', summary: 'Acceso directo al mapa vivo' },
  { id: 'etapa-02', code: '02', short: 'DEMO', label: 'Pulso', summary: 'INCA vs EPA · 12 variables' },
  { id: 'etapa-03', code: '03', short: 'IDW', label: 'Cobertura', summary: 'Grilla 50 m · límite #9CA3AF' },
  { id: 'etapa-04', code: '04', short: 'PCB', label: 'Placa PCB', summary: 'ESP32-S3 · 11 bloques BOM' },
  { id: 'etapa-05', code: '05', short: 'STACK', label: 'Sistema', summary: 'Fórmulas IDW · Go · PostGIS' },
  { id: 'etapa-06', code: '06', short: 'IEEE', label: 'Informe', summary: 'Ficha IEEE + autor' },
]

const themeStore = useThemeStore()
const activeStageId = ref<string>('etapa-01')
const activeStandard = ref<AirQualityStandard>('inca')
const selectedStationCode = ref<string>('TRJ-01')
const isIeeeModalOpen = ref<boolean>(false)

const selectedStation = computed<DemoStationRow>(() => {
  return (
    DEMO_STATION_ROWS.find((row) => row.code === selectedStationCode.value) ??
    DEMO_STATION_ROWS[0]!
  )
})

function stationIndexValue(row: DemoStationRow): string {
  if (!row.hasData) {
    return NO_DATA_LABEL
  }
  const value = activeStandard.value === 'inca' ? row.incaValue : row.epaValue
  return value !== null ? String(value) : NO_DATA_LABEL
}

function stationCategory(row: DemoStationRow): string {
  if (!row.hasData) {
    return NO_DATA_LABEL
  }
  return activeStandard.value === 'inca' ? row.incaCategory : row.epaCategory
}

function stationColor(row: DemoStationRow): string {
  if (!row.hasData) {
    return NO_DATA_COLOR
  }
  return activeStandard.value === 'inca' ? row.incaColor : row.epaColor
}

function inspectDemoStation(code: string): void {
  selectedStationCode.value = code
  scrollToStage('etapa-02')
}

function scrollToStage(id: string): void {
  activeStageId.value = id
  const target = document.getElementById(id)
  if (target) {
    target.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

let stageObserver: IntersectionObserver | null = null

onMounted(() => {
  stageObserver = new IntersectionObserver(
    (entries) => {
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => b.intersectionRatio - a.intersectionRatio)
      const topEntry = visible[0]
      if (topEntry?.target.id) {
        activeStageId.value = topEntry.target.id
      }
    },
    {
      rootMargin: '-20% 0px -55% 0px',
      threshold: [0.15, 0.4, 0.7],
    },
  )

  for (const stage of STAGES) {
    const el = document.getElementById(stage.id)
    if (el) {
      stageObserver.observe(el)
    }
  }
})

onUnmounted(() => {
  stageObserver?.disconnect()
  stageObserver = null
})
</script>

<template>
  <div class="amber-bed">
    <div class="bed-light-sweep" aria-hidden="true" />

    <div class="rack-shell">
      <!-- Lateral Timetable Progress Rail -->
      <aside class="tick-rail" aria-label="Progresión de lectura técnica">
        <div class="rail-brand-stamp tabular">TRJ · HCH</div>
        <nav class="rail-ticks">
          <button
            v-for="stage in STAGES"
            :key="stage.id"
            type="button"
            class="rail-tick"
            :class="{ 'is-active': activeStageId === stage.id }"
            @click="scrollToStage(stage.id)"
          >
            <span class="tick-code tabular">{{ stage.code }}</span>
            <span class="tick-short">{{ stage.short }}</span>
          </button>
        </nav>
        <div class="rail-tz tabular">UTC-5</div>
      </aside>

      <!-- Main Column -->
      <div class="rack-column">
        <!-- Top Timetable Bar -->
        <header class="top-bar">
          <div class="top-brand">
            <span class="brand-badge">AT</span>
            <span class="brand-name">AIRE TRUJILLO</span>
          </div>

          <nav class="top-nav" aria-label="Secciones de la portada">
            <button
              v-for="stage in STAGES"
              :key="`nav-${stage.id}`"
              type="button"
              class="nav-slide-tab"
              :class="{ 'is-active': activeStageId === stage.id }"
              @click="scrollToStage(stage.id)"
            >
              <span>{{ stage.label }}</span>
            </button>
          </nav>

          <div class="top-actions">
            <button
              type="button"
              class="theme-btn"
              :aria-label="themeStore.mode === 'light' ? 'Activar paleta oscura cálida' : 'Activar paleta clara cálida'"
              @click="themeStore.toggle"
            >
              <UiIcon :name="themeStore.mode === 'light' ? 'moon' : 'sun'" :size="15" />
              <span>{{ themeStore.mode === 'light' ? 'Oscuro' : 'Claro' }}</span>
            </button>

            <RouterLink to="/map" class="raked-btn raked-btn-primary top-map-cta">
              <span>Ir al mapa <UiIcon name="arrow-right" :size="15" /></span>
            </RouterLink>
          </div>
        </header>

        <!-- Mobile stage strip (hidden on desktop) -->
        <nav class="mobile-stage-strip" aria-label="Ir a etapa">
          <button
            v-for="stage in STAGES"
            :key="`mob-${stage.id}`"
            type="button"
            class="mob-stage-pill"
            :class="{ 'is-active': activeStageId === stage.id }"
            @click="scrollToStage(stage.id)"
          >
            <span class="tabular">{{ stage.code }}</span>
            <span>{{ stage.short }}</span>
          </button>
        </nav>

        <main class="slides-stack">
          <!-- STAGE 01: Direct Entry + Link to /map -->
          <section id="etapa-01" class="slide-sheet slide-hero">
            <div class="hero-grid">
              <div class="hero-copy">
                <h1>Monitoreo urbano de Trujillo y Huanchaco en pasos de 5 minutos.</h1>
                <p class="hero-lead">
                  Consulta la calidad del aire interpolada cada 50 metros sobre la ciudad y el litoral.
                  Cuando superas los 600 metros sin estación activa, el mapa declara textualmente «No hay datos».
                </p>

                <div class="hero-actions">
                  <RouterLink to="/map" class="raked-btn raked-btn-primary">
                    <span>Ir al mapa interactivo <UiIcon name="arrow-right" :size="16" /></span>
                  </RouterLink>
                  <button
                    type="button"
                    class="raked-btn raked-btn-secondary"
                    @click="scrollToStage('etapa-04')"
                  >
                    <span>Explorar placa PCB <UiIcon name="arrow-down" :size="15" /></span>
                  </button>
                </div>
              </div>

              <div class="hero-schedule" aria-label="Tabla resumida de estaciones demostrativas">
                <div class="schedule-inner">
                  <div class="schedule-meta-bar tabular">
                    <span class="rack-tag rack-tag-amber">21:15 UTC-5</span>
                    <span class="rack-tag rack-tag-bottle">GRILLA 50 M · INCA</span>
                    <span class="rack-tag rack-tag-nodata">CUT &gt; 600 M</span>
                  </div>

                  <div class="schedule-head tabular">
                    <span>HORA</span>
                    <span>SECTOR</span>
                    <span>INCA</span>
                    <span>ESTADO</span>
                  </div>

                  <button
                    v-for="row in DEMO_STATION_ROWS"
                    :key="`hero-${row.code}`"
                    type="button"
                    class="schedule-row tabular"
                    :class="{ 'is-nodata': !row.hasData }"
                    @click="inspectDemoStation(row.code)"
                  >
                    <span class="row-time">{{ row.recordedAt.replace(' UTC-5', '') }}</span>
                    <span class="row-station">
                      <span class="row-code">{{ row.code }}</span>
                      <span class="row-name">{{ row.name }}</span>
                    </span>
                    <span class="row-val">{{ row.hasData ? row.incaValue : '—' }}</span>
                    <span class="row-state">
                      <span class="state-square" :style="{ backgroundColor: row.incaColor }" />
                      <span>{{ row.incaCategory }}</span>
                    </span>
                  </button>

                  <div class="schedule-foot">
                    <span>Pulso simulado · Toca una fila para inspeccionar</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="depth-preview-strip">
              <button
                v-for="stage in STAGES.slice(1)"
                :key="`peek-${stage.id}`"
                type="button"
                class="peek-cell"
                @click="scrollToStage(stage.id)"
              >
                <span class="peek-top">
                  <span class="peek-code tabular">{{ stage.code }}</span>
                  <span class="peek-title">{{ stage.label }}</span>
                </span>
                <span class="peek-summary">{{ stage.summary }}</span>
              </button>
            </div>
          </section>

          <!-- STAGE 02: Demonstrative Preview + ECA/INCA & EPA AQI Switch -->
          <section id="etapa-02" class="slide-sheet">
            <div class="sheet-body">
              <div class="section-header-row">
                <div class="section-heading-group">
                  <h2>Vista previa de estaciones y estándares</h2>
                  <p class="section-summary">
                    Alterna entre el estándar peruano <strong>ECA / INCA (MINAM)</strong> y <strong>AQI (US EPA)</strong>.
                    Los valores mostrados pertenecen a una serie demostrativa de 5 minutos en zona horaria <code>America/Lima</code>.
                  </p>
                </div>

                <div class="standard-switch" role="group" aria-label="Estándar de calidad del aire">
                  <button
                    type="button"
                    class="raked-btn"
                    :class="activeStandard === 'inca' ? 'raked-btn-primary' : 'raked-btn-secondary'"
                    @click="activeStandard = 'inca'"
                  >
                    <span>ECA / INCA (Perú)</span>
                  </button>
                  <button
                    type="button"
                    class="raked-btn"
                    :class="activeStandard === 'epa' ? 'raked-btn-primary' : 'raked-btn-secondary'"
                    @click="activeStandard = 'epa'"
                  >
                    <span>AQI (US EPA)</span>
                  </button>
                </div>
              </div>

              <div class="demo-split">
                <div class="demo-list" role="listbox" aria-label="Estaciones demostrativas">
                  <button
                    v-for="row in DEMO_STATION_ROWS"
                    :key="`demo-${row.code}`"
                    type="button"
                    role="option"
                    :aria-selected="selectedStation.code === row.code"
                    class="demo-station-btn"
                    :class="{
                      'is-selected': selectedStation.code === row.code,
                      'is-nodata-row': !row.hasData,
                    }"
                    @click="selectedStationCode = row.code"
                  >
                    <span class="code-box tabular">{{ row.code }}</span>
                    <span class="station-meta">
                      <strong>{{ row.name }}</strong>
                      <small>{{ row.sector }} · {{ row.distanceLabel }}</small>
                    </span>
                    <span class="station-reading tabular">
                      <span
                        class="reading-badge"
                        :style="{
                          backgroundColor: stationColor(row),
                          color: '#111827',
                        }"
                      >
                        {{ stationIndexValue(row) }}
                      </span>
                    </span>
                  </button>
                </div>

                <div class="demo-detail-panel">
                  <Transition name="slide-swap" mode="out-in">
                    <div :key="`${selectedStation.code}-${activeStandard}`" class="detail-inner">
                      <div class="detail-top">
                        <div>
                          <h3>{{ selectedStation.name }}</h3>
                          <p class="detail-sub tabular">
                            {{ selectedStation.code }} · {{ selectedStation.sector }} · {{ formatStandardTitle(activeStandard) }}
                          </p>
                        </div>
                        <span
                          class="detail-pill tabular"
                          :style="{
                            backgroundColor: stationColor(selectedStation),
                            color: '#111827',
                          }"
                        >
                          {{ stationCategory(selectedStation) }}
                        </span>
                      </div>

                      <div v-if="selectedStation.hasData" class="metrics-register" role="table" aria-label="12 variables canónicas del nodo">
                        <div
                          v-for="key in METRIC_KEYS"
                          :key="key"
                          class="register-cell"
                          role="cell"
                        >
                          <span class="cell-label">{{ METRIC_SPECS[key].label }}</span>
                          <span class="cell-reading tabular">
                            <strong>{{ selectedStation.metrics[key] ?? '—' }}</strong>
                            <small>{{ METRIC_SPECS[key].unit }}</small>
                          </span>
                        </div>
                      </div>

                      <div v-else class="nodata-callout">
                        <div class="nodata-chip tabular">#9CA3AF · {{ NO_DATA_LABEL }}</div>
                        <p>
                          Este punto supera el radio máximo de cobertura (<code>680 m &gt; 600 m</code>).
                          El backend responde <code>has_data: false</code> y <code>value: null</code>; nunca rellena con <code>0</code>.
                        </p>
                      </div>
                    </div>
                  </Transition>
                </div>
              </div>

              <div class="variables-strip">
                <div class="strip-label tabular">12 VARIABLES CANÓNICAS DEL NODO</div>
                <div class="strip-tags">
                  <span v-for="key in METRIC_KEYS" :key="`var-${key}`" class="var-chip tabular">
                    <strong>{{ key }}</strong>
                    <small>{{ METRIC_SPECS[key].unit }}</small>
                  </span>
                </div>
              </div>
            </div>
          </section>

          <!-- STAGE 03: Honest Spatial Coverage (IDW & #9CA3AF) -->
          <section id="etapa-03" class="slide-sheet">
            <div class="sheet-body">
              <div class="section-heading-group">
                <h2>Cobertura espacial sin extrapolaciones engañosas</h2>
                <p class="section-summary">
                  Una estación urbana no representa toda la provincia. A medida que una celda de
                  <code>50 m × 50 m</code> se aleja de los sensores activos, el backend reduce su confianza
                  linealmente hacia gris <code>#9CA3AF</code> y corta la cobertura en <code>600 m</code>.
                </p>
              </div>

              <IdwCoverageProbe />
            </div>
          </section>

          <!-- STAGE 04: Interactive PCB Hardware Explorer (FE-001) -->
          <section id="etapa-04" class="slide-sheet">
            <div class="sheet-body">
              <div class="section-heading-group">
                <h2>Explorador interactivo de la placa PCB</h2>
                <p class="section-summary">
                  Cada nodo en poste integra panel solar, batería de litio, protección marina
                  (<em>conformal coating</em> + caja <code>IP66</code>) y respaldo local en MicroSD. Pasa el
                  cursor o selecciona cualquier bloque para inspeccionar su función y variables físicas.
                </p>
              </div>

              <PcbExplorer />
            </div>
          </section>

          <!-- STAGE 05: Deep Interpolation Math, Architecture & Implementation Stack -->
          <section id="etapa-05" class="slide-sheet">
            <div class="sheet-body">
              <div class="section-heading-group">
                <h2>Motor IDW, cálculo normativo y arquitectura completa</h2>
                <p class="section-summary">
                  Toda la matemática ambiental y el control de cobertura residen en el backend Go; el cliente
                  web únicamente renderiza celdas, colores y series bajo contrato estricto con <code>02-domain-and-data.md</code>.
                </p>
              </div>

              <div class="formulas-ledger">
                <div class="formula-pane">
                  <h3>Interpolación IDW y decaimiento</h3>
                  <p class="formula-eq tabular">
                    v̂(x) = Σ (w_k · v_k) / Σ w_k , &nbsp; w_k = 1 / d_k² &nbsp; (d_k ≤ 600 m)
                  </p>
                  <p class="formula-desc">
                    Distancia geodésica Haversine en ventana temporal <code>[t − 15 min, t]</code>. Si
                    <code>d_k &lt; 1 m</code> adopta la lectura directa; entre <code>400 m</code> y
                    <code>600 m</code> interpola RGB hacia <code>#9CA3AF</code> según
                    <code>confidence = 1 − (d_min − 400) / 200</code>.
                  </p>
                </div>

                <div class="formula-pane">
                  <h3>Subíndices lineales ECA/INCA y EPA</h3>
                  <p class="formula-eq tabular">
                    I_p = ((I_hi − I_lo) / (C_hi − C_lo)) · (C_p − C_lo) + I_lo
                  </p>
                  <p class="formula-desc">
                    El índice compuesto equivale a <code>max_p I_p</code>. Para US EPA, las concentraciones
                    canónicas en masa a 25 °C / 1 atm se convierten internamente en Go
                    (<code>CO: 0.873 ppm</code>, <code>NO2: 0.532 ppb</code>, <code>SO2: 0.382 ppb</code>).
                  </p>
                </div>
              </div>

              <div class="architecture-ledger">
                <div class="arch-row">
                  <div class="arch-tier">
                    <span class="tier-stamp tier-vermilion tabular">CAPA A</span>
                    <h3>Hardware &amp; Firmware</h3>
                    <span class="tier-role">Adquisición de campo 5–10 min</span>
                  </div>
                  <dl class="arch-specs">
                    <div>
                      <dt>MCU y enlace</dt>
                      <dd>ESP32-S3-WROOM-1-N16R8 con módem celular LTE-M SIMCom SIM7080G.</dd>
                    </div>
                    <div>
                      <dt>Sensores</dt>
                      <dd>SPS30, 4× Winsen ZE03 + 2× TI ADS1115, SCD41, SGP40, BME280 e ICS-43434.</dd>
                    </div>
                    <div>
                      <dt>Energía y campo</dt>
                      <dd>CN3791 solar MPPT, TPS61023 + TPS22919, MicroSD offline, USB-C 16P, IP66.</dd>
                    </div>
                    <div>
                      <dt>Telemetría</dt>
                      <dd>Publicación QoS 1 al tópico <code>aire-trujillo/v1/nodes/{node_code}/telemetry</code>.</dd>
                    </div>
                  </dl>
                </div>

                <div class="arch-row">
                  <div class="arch-tier">
                    <span class="tier-stamp tier-bottle tabular">CAPA B</span>
                    <h3>Backend &amp; Geodatos</h3>
                    <span class="tier-role">Fuente única de cálculo e IDW</span>
                  </div>
                  <dl class="arch-specs">
                    <div>
                      <dt>Servicio HTTP</dt>
                      <dd>Go 1.25+, <code>chi/v5</code>, caché en memoria con <code>sync.RWMutex</code>.</dd>
                    </div>
                    <div>
                      <dt>Persistencia</dt>
                      <dd>PostgreSQL 16 + PostGIS (<code>GEOGRAPHY(Point, 4326)</code>) vía <code>pgx/v5</code>.</dd>
                    </div>
                    <div>
                      <dt>Ingesta MQTT</dt>
                      <dd>Eclipse Mosquitto consumido con <code>paho.mqtt.golang</code> y upsert por lote.</dd>
                    </div>
                    <div>
                      <dt>Seguridad admin</dt>
                      <dd>Passphrase con Argon2id + salt, cookie <code>HttpOnly</code> y rate limiting.</dd>
                    </div>
                  </dl>
                </div>

                <div class="arch-row">
                  <div class="arch-tier">
                    <span class="tier-stamp tier-amber tabular">CAPA C</span>
                    <h3>Frontend &amp; Contrato</h3>
                    <span class="tier-role">Renderizado Zero-Drift</span>
                  </div>
                  <dl class="arch-specs">
                    <div>
                      <dt>Núcleo UI</dt>
                      <dd>Vue 3.5 (<code>&lt;script setup lang="ts"&gt;</code>), Vite 8 y Pinia.</dd>
                    </div>
                    <div>
                      <dt>Tipado estricto</dt>
                      <dd>TypeScript sin <code>any</code>; espejo exacto de <code>models.go</code> en <code>telemetry.ts</code>.</dd>
                    </div>
                    <div>
                      <dt>Visualización prevista</dt>
                      <dd>MapLibre GL JS (WebGL) y gráficos 24h en pasos de 5 min.</dd>
                    </div>
                    <div>
                      <dt>Orquestación</dt>
                      <dd>Despliegue reproducible mediante <code>docker-compose.yml</code>.</dd>
                    </div>
                  </dl>
                </div>
              </div>
            </div>
          </section>

          <!-- STAGE 06: IEEE Technical Report + Author Colophon -->
          <section id="etapa-06" class="slide-sheet slide-close">
            <div class="sheet-body close-grid">
              <div class="close-main">
                <h2>Informe técnico IEEE y trazabilidad del proyecto</h2>
                <p class="section-summary">
                  El manuscrito en formato <code>IEEEtran</code> documenta la arquitectura completa del nodo
                  costero, el motor IDW en Go/PostGIS y los criterios de verificación de
                  <code>03-quality-and-testing.md</code>. Su enlace definitivo se incorporará al cerrar las
                  mediciones reproducibles del informe.
                </p>

                <div class="close-actions">
                  <button
                    type="button"
                    class="raked-btn raked-btn-primary"
                    @click="isIeeeModalOpen = true"
                  >
                    <span><UiIcon name="doc" :size="15" /> Ficha del informe IEEE (Próximamente)</span>
                  </button>

                  <a
                    href="https://gmatiasvc.vercel.app"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="raked-btn raked-btn-secondary"
                  >
                    <span>Autor: gmatiasvc.vercel.app <UiIcon name="external" :size="15" /></span>
                  </a>

                  <RouterLink to="/map" class="raked-btn raked-btn-carbon">
                    <span>Abrir mapa en vivo <UiIcon name="arrow-right" :size="15" /></span>
                  </RouterLink>
                </div>
              </div>

              <aside class="colophon-card">
                <div class="colophon-row">
                  <span>Cobertura inicial</span>
                  <strong>Trujillo &amp; Huanchaco, Perú</strong>
                </div>
                <div class="colophon-row">
                  <span>Zona horaria</span>
                  <strong class="tabular">America/Lima (UTC-5)</strong>
                </div>
                <div class="colophon-row">
                  <span>Estándares</span>
                  <strong>ECA / INCA (MINAM) · US EPA AQI</strong>
                </div>
                <div class="colophon-row">
                  <span>Estado actual</span>
                  <strong>Demo con nodos virtuales</strong>
                </div>
              </aside>
            </div>
          </section>
        </main>
      </div>
    </div>

    <IeeeReportModal :open="isIeeeModalOpen" @close="isIeeeModalOpen = false" />
  </div>
</template>

<style scoped>
.amber-bed {
  position: relative;
  min-height: 100svh;
  background-color: var(--bed-bg);
  background-image:
    radial-gradient(circle at 22% 14%, var(--bed-sheen) 0%, transparent 52%),
    radial-gradient(circle at 82% 72%, rgba(201, 140, 25, 0.24) 0%, transparent 48%);
  padding: 28px clamp(18px, 4vw, 56px) 56px;
  overflow-x: clip;
}

.bed-light-sweep {
  position: fixed;
  inset: -20% -30%;
  pointer-events: none;
  background: linear-gradient(
    112deg,
    transparent 38%,
    var(--bed-sheen) 50%,
    transparent 62%
  );
  transform: translateX(-28%);
  animation: bed-sweep 16s var(--ease-expo) infinite alternate;
  z-index: 0;
}

@keyframes bed-sweep {
  from {
    transform: translateX(-26%) translateY(-4%);
  }
  to {
    transform: translateX(22%) translateY(6%);
  }
}

@keyframes slide-arrive {
  from {
    opacity: 0;
    transform: translateY(14px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.rack-shell {
  position: relative;
  z-index: 1;
  max-width: 1320px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 86px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}

/* Left Vertical Progress Rail with graduated ruler spine */
.tick-rail {
  position: sticky;
  top: 24px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border: 1.5px solid var(--rule-color);
  padding: 16px 12px 16px 10px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 18px;
  min-height: calc(100svh - 48px);
}

.tick-rail::after {
  content: '';
  position: absolute;
  top: 44px;
  bottom: 44px;
  right: 0;
  width: 5px;
  background: repeating-linear-gradient(
    to bottom,
    rgba(231, 176, 67, 0.55) 0 1px,
    transparent 1px 8px
  );
  pointer-events: none;
}

.rail-brand-stamp,
.rail-tz {
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-align: center;
  color: var(--ink-muted-on-carbon);
}

.rail-ticks {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rail-tick {
  border: 1px solid transparent;
  border-top: 1px solid var(--rule-soft);
  background: transparent;
  color: var(--ink-muted-on-carbon);
  padding: 10px 4px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  cursor: pointer;
  font-family: var(--font-mono);
  transition:
    color 0.2s var(--ease-expo),
    background-color 0.2s var(--ease-expo),
    border-color 0.2s var(--ease-expo);
}

.rail-tick:hover {
  color: var(--ink-on-carbon);
  background: rgba(255, 255, 255, 0.05);
}

.rail-tick.is-active {
  background: var(--accent-amber-fill);
  border-color: var(--accent-amber-fill);
  color: #181a1c;
}

.tick-code {
  font-size: 13px;
  font-weight: 600;
}

.tick-short {
  font-size: 9.5px;
  letter-spacing: 0.06em;
}

/* Main Column */
.rack-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

/* Top Bar */
.top-bar {
  position: sticky;
  top: 12px;
  z-index: 20;
  background: var(--surface-ivory);
  border: 1.5px solid var(--rule-color);
  padding: 10px 18px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
}

.top-brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-badge {
  display: inline-grid;
  place-items: center;
  width: 30px;
  height: 30px;
  background: var(--surface-carbon);
  color: var(--accent-amber-fill);
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 16px;
  letter-spacing: 0.04em;
}

.brand-name {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 20px;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.top-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-slide-tab {
  border: 1px solid var(--rule-soft);
  background: transparent;
  color: var(--ink-secondary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  white-space: nowrap;
  padding: 6px 11px;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.2s var(--ease-expo),
    color 0.2s var(--ease-expo),
    border-color 0.2s var(--ease-expo);
}

.nav-slide-tab > span {
  display: block;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.nav-slide-tab:hover {
  border-color: var(--rule-color);
  color: var(--ink-primary);
}

.nav-slide-tab.is-active {
  background: var(--surface-carbon);
  border-color: var(--rule-color);
  color: var(--ink-on-carbon);
}

.top-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.theme-btn {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  padding: 9px 12px;
  cursor: pointer;
  transition:
    background-color 0.2s var(--ease-expo),
    color 0.2s var(--ease-expo);
}

.theme-btn:hover {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.top-map-cta {
  padding: 9px 16px;
  white-space: nowrap;
}

.mobile-stage-strip {
  display: none;
}

/* Progressive Slide Sheets */
.slides-stack {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.slide-sheet {
  scroll-margin-top: 84px;
  background: var(--surface-ivory);
  border: 1.5px solid var(--rule-color);
  transition:
    background-color 0.3s var(--ease-expo),
    border-color 0.3s var(--ease-expo);
}

.sheet-body {
  padding: clamp(24px, 3.2vw, 38px);
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-width: 0;
}

.sheet-body > * {
  min-width: 0;
}

/* Stage 01 Hero */
.hero-grid {
  padding: clamp(26px, 3.4vw, 42px);
  display: grid;
  grid-template-columns: 1.06fr 0.94fr;
  gap: clamp(26px, 3.2vw, 48px);
  align-items: center;
}

.hero-copy {
  animation: slide-arrive 0.45s var(--ease-expo) both;
}

h1 {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: clamp(2.55rem, 4.7vw, 4.15rem);
  line-height: 0.98;
  letter-spacing: -0.02em;
  text-transform: uppercase;
  margin-bottom: 18px;
}

.hero-lead {
  font-size: 17px;
  line-height: 1.55;
  color: var(--ink-secondary);
  max-width: 58ch;
  margin-bottom: 26px;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
}

/* Raked Tinted Glass Slate Timetable Card */
.hero-schedule {
  overflow: hidden;
  border: 1.5px solid var(--rule-color);
  background: var(--glass-slate);
  color: var(--glass-slate-ink);
  backdrop-filter: blur(12px) saturate(118%);
  box-shadow: inset 0 1px 0 var(--glass-slate-border);
  transform: skewX(-2.5deg);
  margin: 0 6px;
  animation: slide-arrive 0.52s var(--ease-expo) 80ms both;
}

.schedule-inner {
  transform: skewX(2.5deg);
}

.schedule-meta-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 10px 18px;
  border-bottom: 1px solid var(--glass-slate-rule);
}

.rack-tag {
  display: inline-block;
  padding: 3px 8px;
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.rack-tag-amber {
  background: var(--accent-amber-fill);
  color: #181a1c;
}

.rack-tag-bottle {
  background: var(--accent-bottle);
  color: var(--accent-bottle-ink);
}

.rack-tag-nodata {
  background: var(--nodata-bg);
  color: var(--nodata-ink);
}

.schedule-head,
.schedule-row {
  display: grid;
  grid-template-columns: 56px 1.25fr 56px 1fr;
  gap: 10px;
  align-items: center;
  padding: 11px 18px;
  font-family: var(--font-mono);
  font-size: 12px;
}

.schedule-head {
  color: var(--glass-slate-muted);
  border-bottom: 1px solid var(--glass-slate-rule);
  font-weight: 600;
  font-size: 10.5px;
  letter-spacing: 0.06em;
}

.schedule-row {
  width: 100%;
  text-align: left;
  border: none;
  border-top: 1px solid var(--glass-slate-rule);
  background: transparent;
  color: var(--glass-slate-ink);
  cursor: pointer;
  transition:
    background-color 0.18s var(--ease-expo),
    transform 0.18s var(--ease-expo);
}

.schedule-row:hover {
  background: var(--glass-slate-row-hover);
  transform: translateX(3px);
}

.schedule-row.is-nodata {
  background: var(--glass-slate-nodata);
}

.row-station {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.row-code {
  font-size: 10.5px;
  font-weight: 600;
  color: var(--accent-amber-fill);
}

.row-name {
  font-family: var(--font-body);
  font-size: 13px;
  font-weight: 600;
}

.row-val {
  font-size: 15px;
  font-weight: 600;
}

.row-state {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 11.5px;
  font-weight: 600;
}

.state-square {
  width: 10px;
  height: 10px;
  flex-shrink: 0;
  border: 1px solid var(--glass-slate-border);
}

.schedule-foot {
  padding: 9px 18px;
  border-top: 1px solid var(--glass-slate-rule);
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--glass-slate-muted);
}

.depth-preview-strip {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  border-top: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  animation: slide-arrive 0.56s var(--ease-expo) 150ms both;
}

.peek-cell {
  border: none;
  border-right: 1px solid var(--rule-soft);
  background: transparent;
  color: var(--ink-primary);
  padding: 13px 14px;
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: 4px;
  cursor: pointer;
  transition:
    background-color 0.2s var(--ease-expo),
    color 0.2s var(--ease-expo);
}

.peek-cell:last-child {
  border-right: none;
}

.peek-cell:hover {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.peek-top {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.peek-code {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-vermilion);
}

:root[data-theme='dark'] .peek-code {
  color: var(--accent-amber);
}

.peek-cell:hover .peek-code {
  color: var(--accent-amber-fill);
}

.peek-title {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.peek-summary {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
  line-height: 1.35;
}

.peek-cell:hover .peek-summary {
  color: var(--ink-muted-on-carbon);
}

/* Shared Section Typography */
.section-header-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: flex-end;
  gap: 18px;
}

.section-heading-group h2 {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: clamp(1.9rem, 3.2vw, 2.75rem);
  line-height: 1.02;
  text-transform: uppercase;
  margin-bottom: 10px;
}

.section-summary {
  max-width: 68ch;
  color: var(--ink-secondary);
  font-size: 15.5px;
  line-height: 1.55;
}

.section-summary code {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 1px 6px;
  background: var(--surface-ivory-2);
  border: 1px solid var(--rule-soft);
}

.standard-switch {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

/* Stage 02 Demo Split */
.demo-split {
  display: grid;
  grid-template-columns: 0.95fr 1.05fr;
  gap: 20px;
}

.demo-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.demo-station-btn {
  display: grid;
  grid-template-columns: 74px 1fr auto;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.2s var(--ease-expo),
    transform 0.2s var(--ease-expo),
    border-color 0.2s var(--ease-expo);
}

.demo-station-btn:hover {
  transform: translateX(3px);
  background: var(--surface-ivory);
}

.demo-station-btn.is-selected {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.code-box {
  display: grid;
  place-items: center;
  padding: 7px 6px;
  background: var(--accent-amber-fill);
  color: #181a1c;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.demo-station-btn.is-nodata-row .code-box {
  background: var(--nodata-bg);
  color: var(--nodata-ink);
}

.station-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.station-meta strong {
  font-size: 14.5px;
}

.station-meta small {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
}

.demo-station-btn.is-selected .station-meta small {
  color: var(--ink-muted-on-carbon);
}

.reading-badge {
  display: inline-block;
  padding: 5px 10px;
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--rule-color);
}

.demo-detail-panel {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  padding: 20px;
  min-height: 258px;
}

.detail-inner {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--rule-soft);
}

.detail-top h3 {
  font-family: var(--font-display);
  font-size: 26px;
  font-weight: 700;
  text-transform: uppercase;
  line-height: 1.05;
}

.detail-sub {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
  margin-top: 4px;
}

.detail-pill {
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  padding: 5px 12px;
  border: 1px solid var(--rule-color);
}

.metrics-register {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border: 1px solid var(--rule-color);
}

.register-cell {
  border-right: 1px solid var(--rule-soft);
  border-bottom: 1px solid var(--rule-soft);
  padding: 9px 10px;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.register-cell:nth-child(3n) {
  border-right: none;
}

.register-cell:nth-last-child(-n + 3) {
  border-bottom: none;
}

.cell-label {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--ink-secondary);
}

.cell-reading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 6px;
}

.cell-reading strong {
  font-family: var(--font-mono);
  font-size: 15.5px;
  font-weight: 600;
}

.cell-reading small {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--ink-secondary);
}

.nodata-callout {
  border: 1px dashed var(--rule-color);
  background: var(--surface-ivory);
  padding: 18px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.nodata-chip {
  background: var(--nodata-bg);
  color: var(--nodata-ink);
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  padding: 5px 12px;
  border: 1px solid var(--rule-color);
}

.nodata-callout p {
  font-size: 14px;
  color: var(--ink-secondary);
  line-height: 1.5;
}

.variables-strip {
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
  padding: 14px 16px;
}

.strip-label {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-secondary);
  margin-bottom: 10px;
}

.strip-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.var-chip {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  padding: 4px 9px;
  border: 1px solid var(--rule-soft);
  background: var(--surface-ivory);
  font-family: var(--font-mono);
  font-size: 11.5px;
}

.var-chip small {
  color: var(--ink-secondary);
  font-size: 10.5px;
}

/* Stage 05 Architecture & Formulas (ruled ledger compartments directly on slide surface) */
.formulas-ledger {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border-top: 1.5px solid var(--rule-color);
  border-bottom: 1.5px solid var(--rule-color);
}

.formula-pane {
  padding: 18px 20px 18px 0;
}

.formula-pane:last-child {
  padding: 18px 0 18px 22px;
  border-left: 1px solid var(--rule-color);
}

.formula-pane h3 {
  font-family: var(--font-display);
  font-size: 23px;
  font-weight: 700;
  text-transform: uppercase;
  margin-bottom: 10px;
}

.formula-eq {
  font-family: var(--font-mono);
  font-size: 12.5px;
  font-weight: 600;
  padding: 10px 12px;
  background: var(--surface-carbon);
  color: var(--accent-amber-fill);
  margin-bottom: 12px;
  overflow-x: auto;
}

.formula-desc {
  font-size: 14px;
  line-height: 1.55;
  color: var(--ink-secondary);
}

.formula-desc code {
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 1px 5px;
  background: var(--surface-ivory-2);
  border: 1px solid var(--rule-soft);
}

.architecture-ledger {
  border-top: 1.5px solid var(--rule-color);
  border-bottom: 1.5px solid var(--rule-color);
}

.arch-row {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  border-bottom: 1px solid var(--rule-soft);
  transition: background-color 0.2s var(--ease-expo);
}

.arch-row:last-child {
  border-bottom: none;
}

.arch-row:hover {
  background: var(--surface-ivory-2);
}

.arch-tier {
  padding: 18px 18px 18px 0;
  border-right: 1px solid var(--rule-color);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.tier-stamp {
  display: inline-block;
  padding: 3px 9px;
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.tier-vermilion {
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.tier-bottle {
  background: var(--accent-bottle);
  color: var(--accent-bottle-ink);
}

.tier-amber {
  background: var(--accent-amber-fill);
  color: #181a1c;
}

.arch-tier h3 {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 700;
  text-transform: uppercase;
  line-height: 1.05;
}

.tier-role {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
}

.arch-specs {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
}

.arch-specs > div {
  padding: 14px 16px;
  border-right: 1px solid var(--rule-soft);
  border-bottom: 1px solid var(--rule-soft);
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.arch-specs > div:nth-child(2n) {
  border-right: none;
}

.arch-specs > div:nth-last-child(-n + 2) {
  border-bottom: none;
}

.arch-specs dt {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-vermilion);
}

:root[data-theme='dark'] .arch-specs dt {
  color: var(--accent-amber);
}

.arch-specs dd {
  margin: 0;
  font-size: 13.5px;
  line-height: 1.48;
  color: var(--ink-secondary);
}

.arch-specs code {
  font-family: var(--font-mono);
  font-size: 11.5px;
  overflow-wrap: anywhere;
}

/* Stage 06 IEEE & Author Close */
.close-grid {
  display: grid;
  grid-template-columns: 1.25fr 0.75fr;
  gap: 28px;
  align-items: center;
}

.close-main h2 {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: clamp(1.9rem, 3.2vw, 2.75rem);
  line-height: 1.02;
  text-transform: uppercase;
  margin-bottom: 12px;
}

.close-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 22px;
}

.colophon-card {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.colophon-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--rule-soft);
  font-size: 13px;
}

.colophon-row:last-child {
  border-bottom: none;
}

.colophon-row span {
  font-family: var(--font-mono);
  font-size: 11.5px;
  color: var(--ink-secondary);
}

.colophon-row strong {
  text-align: right;
}

/* Responsive Breakpoints */
@media (max-width: 1160px) {
  .top-nav {
    display: none;
  }
}

@media (max-width: 960px) {
  .rack-shell {
    grid-template-columns: 1fr;
    gap: 14px;
  }

  .tick-rail {
    display: none;
  }

  .hero-schedule,
  .schedule-inner {
    transform: none;
    margin: 0;
  }

  .mobile-stage-strip {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 6px;
  }

  .mob-stage-pill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px 6px;
    border: 1.5px solid var(--rule-color);
    background: var(--surface-ivory);
    color: var(--ink-primary);
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
  }

  .mob-stage-pill.is-active {
    background: var(--surface-carbon);
    color: var(--accent-amber-fill);
  }

  .hero-grid,
  .demo-split,
  .formulas-ledger,
  .arch-row,
  .arch-specs,
  .close-grid {
    grid-template-columns: 1fr;
  }

  .formula-pane,
  .formula-pane:last-child {
    padding: 16px 0;
    border-left: none;
  }

  .formula-pane:last-child {
    border-top: 1px solid var(--rule-soft);
  }

  .arch-tier {
    padding: 16px 0 12px;
    border-right: none;
    border-bottom: 1px solid var(--rule-soft);
  }

  .arch-specs > div {
    border-right: none;
    padding: 12px 0;
  }

  .arch-specs > div:nth-last-child(-n + 2) {
    border-bottom: 1px solid var(--rule-soft);
  }

  .arch-specs > div:last-child {
    border-bottom: none;
  }

  .depth-preview-strip {
    grid-template-columns: repeat(2, 1fr);
  }

  .peek-cell {
    border-bottom: 1px solid var(--rule-soft);
  }
}

@media (max-width: 580px) {
  .amber-bed {
    padding: 12px 10px 28px;
  }

  .top-bar {
    position: static;
    flex-wrap: wrap;
    padding: 10px 12px;
  }

  .top-map-cta {
    width: 100%;
  }

  .metrics-register {
    grid-template-columns: repeat(2, 1fr);
  }

  .register-cell:nth-child(3n) {
    border-right: 1px solid var(--rule-soft);
  }

  .register-cell:nth-child(2n) {
    border-right: none;
  }

  .register-cell:nth-last-child(-n + 3) {
    border-bottom: 1px solid var(--rule-soft);
  }

  .register-cell:nth-last-child(-n + 2) {
    border-bottom: none;
  }

  .schedule-head,
  .schedule-row {
    grid-template-columns: 46px 1fr 48px 0.95fr;
    font-size: 11px;
    padding: 10px;
  }

  .demo-station-btn {
    grid-template-columns: 62px 1fr auto;
    padding: 10px;
  }

  .depth-preview-strip {
    grid-template-columns: 1fr;
  }
}
</style>
