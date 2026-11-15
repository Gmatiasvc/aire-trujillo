<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import {
  CategoryScale,
  Chart,
  Filler,
  LinearScale,
  LineController,
  LineElement,
  PointElement,
  Tooltip,
  type ChartConfiguration,
} from 'chart.js'
import UiIcon from '../UiIcon.vue'
import { formatLimaDateTime, formatLimaTime } from '../../data/legends'
import { useThemeStore } from '../../stores/theme'
import type { SelectedPointMeta } from '../../stores/map'
import {
  NO_DATA_COLOR,
  NO_DATA_LABEL,
  type PointResponse,
} from '../../types/telemetry'

Chart.register(
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  CategoryScale,
  Tooltip,
  Filler,
)

const props = defineProps<{
  selectedPoint: SelectedPointMeta | null
  pointData: PointResponse | null
  isLoading: boolean
  error: string | null
  layerTitle: string
  activeTimestamp?: string | null
}>()

const emit = defineEmits<{
  close: []
  retry: []
}>()

const themeStore = useThemeStore()
const canvasEl = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart<'line', Array<number | null>, string> | null = null

function getContrastInk(hex: string): string {
  const cleaned = hex.replace('#', '')
  if (cleaned.length !== 6) return '#111827'
  const r = parseInt(cleaned.slice(0, 2), 16)
  const g = parseInt(cleaned.slice(2, 4), 16)
  const b = parseInt(cleaned.slice(4, 6), 16)
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  return luminance > 0.58 ? '#111827' : '#F7F1E3'
}

const currentConfidencePct = computed<number>(() => {
  if (!props.pointData?.current) return 0
  return Math.round(props.pointData.current.confidence * 100)
})

const currentBadgeColor = computed<string>(() => {
  return props.pointData?.current?.color ?? NO_DATA_COLOR
})

const currentBadgeInk = computed<string>(() => {
  return getContrastInk(currentBadgeColor.value)
})

const validSeriesCount = computed<number>(() => {
  if (!props.pointData?.series) return 0
  return props.pointData.series.filter((item) => item.has_data && item.value !== null).length
})

function destroyChart(): void {
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
}

function resolveActiveSeriesIndex(series: PointResponse['series']): number {
  const ts = props.activeTimestamp
  if (!ts || series.length === 0) return -1
  const exact = series.findIndex((pt) => pt.recorded_at === ts)
  if (exact >= 0) return exact
  const formatted = formatLimaDateTime(ts)
  return series.findIndex((pt) => formatLimaDateTime(pt.recorded_at) === formatted)
}

function buildPointStyles(
  series: PointResponse['series'],
  values: Array<number | null>,
  isDark: boolean,
  lineColor: string,
) {
  const activeIdx = resolveActiveSeriesIndex(series)
  const activeStroke = isDark ? '#F7F1E3' : '#181A1C'
  const activeFill = '#B83A28'
  return {
    pointRadius: values.map((v, idx) => (idx === activeIdx && v !== null ? 5 : 0)),
    pointBackgroundColor: values.map((_, idx) => (idx === activeIdx ? activeFill : lineColor)),
    pointBorderColor: values.map((_, idx) => (idx === activeIdx ? activeStroke : lineColor)),
    pointBorderWidth: values.map((_, idx) => (idx === activeIdx ? 2 : 0)),
  }
}

function renderChart(): void {
  destroyChart()
  const canvas = canvasEl.value
  const data = props.pointData
  if (!canvas || !data || !data.has_data || data.series.length === 0) {
    return
  }

  const isDark = themeStore.mode === 'dark'
  const lineColor = isDark ? '#E6AA3C' : '#B83A28'
  const fillColor = isDark ? 'rgba(230, 170, 60, 0.14)' : 'rgba(184, 58, 40, 0.12)'
  const gridColor = isDark ? 'rgba(107, 86, 58, 0.38)' : 'rgba(199, 182, 152, 0.55)'
  const tickColor = isDark ? '#CBBCA3' : '#433A2E'

  const labels = data.series.map((pt) => formatLimaTime(pt.recorded_at))
  const values = data.series.map((pt) => (pt.has_data && pt.value !== null ? pt.value : null))
  const pointStyles = buildPointStyles(data.series, values, isDark, lineColor)

  const config: ChartConfiguration<'line', Array<number | null>, string> = {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: `${props.layerTitle} (${data.unit})`,
          data: values,
          borderColor: lineColor,
          backgroundColor: fillColor,
          borderWidth: 2,
          pointRadius: pointStyles.pointRadius,
          pointBackgroundColor: pointStyles.pointBackgroundColor,
          pointBorderColor: pointStyles.pointBorderColor,
          pointBorderWidth: pointStyles.pointBorderWidth,
          pointHoverRadius: 4,
          pointHoverBackgroundColor: lineColor,
          tension: 0.22,
          fill: true,
          spanGaps: false,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 180,
      },
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          backgroundColor: isDark ? '#100D0A' : '#191D20',
          titleColor: '#F7F1E3',
          bodyColor: '#F7F1E3',
          borderColor: '#E7B043',
          borderWidth: 1,
          cornerRadius: 0,
          displayColors: false,
          callbacks: {
            title(items) {
              const first = items[0]
              return first ? `${first.label} UTC-5` : ''
            },
            label(item) {
              if (item.raw === null || item.raw === undefined) {
                return NO_DATA_LABEL
              }
              return `${Number(item.raw).toFixed(2)} ${data.unit}`
            },
          },
        },
      },
      scales: {
        x: {
          grid: {
            color: gridColor,
          },
          ticks: {
            color: tickColor,
            maxTicksLimit: 7,
            font: {
              family: "'Azeret Mono', monospace",
              size: 10,
            },
          },
        },
        y: {
          grid: {
            color: gridColor,
          },
          ticks: {
            color: tickColor,
            font: {
              family: "'Azeret Mono', monospace",
              size: 10,
            },
          },
        },
      },
    },
  }

  chartInstance = new Chart(canvas, config)
}

watch(
  [() => props.pointData, () => themeStore.mode, canvasEl],
  () => {
    renderChart()
  },
  { flush: 'post' },
)

watch(
  () => props.activeTimestamp,
  () => {
    const data = props.pointData
    if (!chartInstance || !data || !data.has_data || data.series.length === 0) return
    const dataset = chartInstance.data.datasets[0]
    if (!dataset) return
    const isDark = themeStore.mode === 'dark'
    const lineColor = isDark ? '#E6AA3C' : '#B83A28'
    const values = data.series.map((pt) => (pt.has_data && pt.value !== null ? pt.value : null))
    const styles = buildPointStyles(data.series, values, isDark, lineColor)
    dataset.pointRadius = styles.pointRadius
    dataset.pointBackgroundColor = styles.pointBackgroundColor
    dataset.pointBorderColor = styles.pointBorderColor
    dataset.pointBorderWidth = styles.pointBorderWidth
    chartInstance.update('none')
  },
  { flush: 'post' },
)

onBeforeUnmount(() => {
  destroyChart()
})
</script>

<template>
  <Transition name="drawer-slide">
    <aside
      v-if="selectedPoint"
      class="point-drawer"
      aria-label="Inspección puntual de 24 horas"
    >
      <header class="drawer-head">
        <div class="head-meta">
          <div class="head-title-row">
            <span v-if="selectedPoint.nodeCode" class="node-tag tabular">
              {{ selectedPoint.nodeCode }}
            </span>
            <h2 class="drawer-title">
              {{ selectedPoint.nodeName ?? 'Consulta puntual' }}
            </h2>
          </div>
          <p class="coord-line tabular">
            {{ selectedPoint.lat.toFixed(6) }}, {{ selectedPoint.lon.toFixed(6) }}
          </p>
        </div>

        <button
          type="button"
          class="close-btn"
          aria-label="Cerrar inspección puntual"
          @click="emit('close')"
        >
          <UiIcon name="close" :size="16" />
        </button>
      </header>

      <div class="drawer-body">
        <div v-if="isLoading" class="skeleton-stack" aria-busy="true" aria-label="Cargando lectura puntual">
          <div class="skeleton-box is-lg"></div>
          <div class="skeleton-box is-sm"></div>
          <div class="skeleton-box is-chart"></div>
        </div>

        <div v-else-if="error" class="error-box" role="alert">
          <p>{{ error }}</p>
          <button type="button" class="raked-btn raked-btn-secondary retry-btn" @click="emit('retry')">
            <span>
              <UiIcon name="refresh" :size="14" />
              Reintentar
            </span>
          </button>
        </div>

        <template v-else-if="pointData">
          <!-- FE-003 / FE-004: Explicit No hay datos when uncovered -->
          <div v-if="!pointData.has_data || !pointData.current" class="nodata-panel">
            <div class="nodata-banner">
              <span class="nodata-dot" aria-hidden="true"></span>
              <strong>{{ pointData.label || NO_DATA_LABEL }}</strong>
            </div>
            <p class="nodata-copy">
              Coordenada fuera del radio máximo de cobertura IDW (600 m). El sistema declara
              explícitamente ausencia de datos en lugar de extrapolar o mostrar 0.
            </p>
            <div class="empty-chart-well">
              <span class="empty-chart-label">{{ NO_DATA_LABEL }}</span>
              <span class="empty-chart-sub">Sin serie temporal interpolada de 24 horas</span>
            </div>
          </div>

          <div v-else class="covered-panel">
            <div class="readout-well">
              <div class="readout-top">
                <span class="layer-pill">{{ layerTitle }}</span>
                <span
                  class="category-badge"
                  :style="{ backgroundColor: currentBadgeColor, color: currentBadgeInk }"
                >
                  {{ pointData.current.category }}
                </span>
              </div>

              <div class="value-row">
                <strong class="value-number tabular">{{ pointData.current.value.toFixed(1) }}</strong>
                <span class="value-unit">{{ pointData.unit }}</span>
              </div>

              <div class="confidence-wrap">
                <div class="confidence-head tabular">
                  <span>Confianza IDW</span>
                  <span>{{ currentConfidencePct }}%</span>
                </div>
                <div class="confidence-track" aria-hidden="true">
                  <div
                    class="confidence-fill"
                    :style="{
                      width: `${currentConfidencePct}%`,
                      backgroundColor: currentBadgeColor,
                    }"
                  ></div>
                </div>
                <p class="confidence-note tabular">
                  {{ formatLimaDateTime(pointData.current.recorded_at) }} ·
                  <span v-if="currentConfidencePct >= 99" class="nowrap-phrase">
                    Cobertura plena (&le;&nbsp;400&nbsp;m)
                  </span>
                  <span v-else class="nowrap-phrase">
                    Zona de borde (400&ndash;600&nbsp;m)
                  </span>
                </p>
              </div>
            </div>

            <div class="chart-section">
              <div class="chart-head">
                <h3>Evolución · <span class="nowrap-phrase">24&nbsp;h</span></h3>
                <span class="chart-steps tabular">
                  {{ validSeriesCount }} lecturas · <span class="nowrap-phrase">paso 5&nbsp;min</span>
                </span>
              </div>

              <div v-if="validSeriesCount === 0" class="empty-chart-well">
                <span class="empty-chart-label">{{ NO_DATA_LABEL }}</span>
                <span class="empty-chart-sub">Sin histórico disponible en este punto</span>
              </div>
              <div v-else class="chart-canvas-wrap">
                <canvas ref="canvasEl" aria-label="Gráfico temporal de 24 horas en el punto"></canvas>
              </div>
            </div>
          </div>
        </template>
      </div>
    </aside>
  </Transition>
</template>

<style scoped>
.point-drawer {
  position: absolute;
  top: 12px;
  right: 12px;
  bottom: 92px;
  z-index: 15;
  width: min(392px, calc(100vw - 24px));
  display: flex;
  flex-direction: column;
  background: var(--surface-ivory);
  color: var(--ink-primary);
  border: 1.5px solid var(--rule-color);
  overflow: hidden;
}

.drawer-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border-bottom: 1.5px solid var(--rule-color);
}

.head-meta {
  min-width: 0;
}

.head-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.node-tag {
  padding: 2px 6px;
  background: var(--accent-amber-fill);
  color: #181a1c;
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
}

.drawer-title {
  font-family: var(--font-display);
  font-size: 1.35rem;
  font-weight: 700;
  line-height: 1.05;
  text-transform: uppercase;
  color: var(--ink-on-carbon);
}

.coord-line {
  margin-top: 4px;
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-muted-on-carbon);
}

.close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border: 1px solid var(--glass-slate-rule);
  background: transparent;
  color: var(--ink-on-carbon);
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.16s var(--ease-expo);
}

.close-btn:hover {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.skeleton-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.skeleton-box {
  background: var(--surface-ivory-2);
  border: 1px solid var(--rule-soft);
  animation: pulse-well 1.2s ease-in-out infinite;
}

.skeleton-box.is-lg {
  height: 92px;
}

.skeleton-box.is-sm {
  height: 46px;
}

.skeleton-box.is-chart {
  height: 210px;
}

@keyframes pulse-well {
  0%,
  100% {
    opacity: 0.65;
  }
  50% {
    opacity: 1;
  }
}

.error-box {
  padding: 14px;
  border: 1.5px solid var(--accent-vermilion);
  background: var(--surface-ivory-2);
  display: flex;
  flex-direction: column;
  gap: 10px;
  font-size: 13px;
}

.retry-btn {
  align-self: flex-start;
  padding: 8px 14px;
}

.nodata-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.nodata-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: var(--nodata-bg);
  color: var(--nodata-ink);
  border: 1.5px solid var(--rule-color);
  font-family: var(--font-display);
  font-size: 1.55rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.nodata-dot {
  width: 12px;
  height: 12px;
  background: #111827;
}

.nodata-copy {
  font-size: 13px;
  line-height: 1.5;
  color: var(--ink-secondary);
}

.empty-chart-well {
  height: 196px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: 6px;
  background: var(--surface-ivory-2);
  border: 1px dashed var(--rule-color);
}

.empty-chart-label {
  padding: 4px 10px;
  background: var(--nodata-bg);
  color: var(--nodata-ink);
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
}

.empty-chart-sub {
  font-size: 12px;
  color: var(--ink-secondary);
}

.covered-panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.readout-well {
  padding: 12px;
  background: var(--surface-ivory-2);
  border: 1px solid var(--rule-color);
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.readout-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
}

.layer-pill {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-secondary);
  text-transform: uppercase;
}

.category-badge {
  padding: 3px 8px;
  border: 1px solid var(--rule-color);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.value-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.value-number {
  font-family: var(--font-display);
  font-size: 2.75rem;
  font-weight: 800;
  line-height: 0.95;
  color: var(--ink-primary);
}

.value-unit {
  font-family: var(--font-mono);
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink-secondary);
}

.confidence-wrap {
  padding-top: 8px;
  border-top: 1px solid var(--rule-soft);
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.confidence-head {
  display: flex;
  justify-content: space-between;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.confidence-track {
  height: 7px;
  background: var(--surface-ivory);
  border: 1px solid var(--rule-color);
}

.confidence-fill {
  height: 100%;
  transition: width 0.22s var(--ease-expo);
}

.confidence-note {
  font-family: var(--font-mono);
  font-size: 10.5px;
  line-height: 1.45;
  color: var(--ink-secondary);
}

.nowrap-phrase {
  display: inline-block;
  white-space: nowrap;
}

.chart-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.chart-head {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 4px 10px;
}

.chart-head h3 {
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 700;
  text-transform: uppercase;
  white-space: nowrap;
}

.chart-steps {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
}

.chart-canvas-wrap {
  position: relative;
  height: 224px;
  padding: 8px;
  background: var(--surface-ivory-2);
  border: 1px solid var(--rule-color);
}

.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition:
    transform 0.22s var(--ease-expo),
    opacity 0.18s var(--ease-expo);
}

.drawer-slide-enter-from,
.drawer-slide-leave-to {
  opacity: 0;
  transform: translateX(18px);
}

@media (max-width: 900px) {
  .point-drawer {
    top: auto;
    left: 10px;
    right: 10px;
    bottom: 128px;
    width: auto;
    max-height: calc(100% - 254px);
  }

  .drawer-body {
    padding: 10px 12px;
    gap: 8px;
  }

  .readout-well {
    padding: 9px 10px;
    gap: 6px;
  }

  .value-number {
    font-size: 2.1rem;
  }

  .chart-canvas-wrap {
    height: 150px;
    padding: 6px;
  }

  .drawer-slide-enter-from,
  .drawer-slide-leave-to {
    transform: translateY(20px);
  }
}
</style>
