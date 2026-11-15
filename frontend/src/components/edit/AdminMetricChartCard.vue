<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
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
  type ChartDataset,
} from 'chart.js'
import { useThemeStore } from '../../stores/theme'
import type { MetricSpec } from '../../types/telemetry'
import type { VariableSeriesPoint } from '../../utils/adminSeries'

Chart.register(LineController, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Filler)

const props = defineProps<{
  spec: MetricSpec
  points: VariableSeriesPoint[]
}>()

const themeStore = useThemeStore()
const canvasEl = ref<HTMLCanvasElement | null>(null)
const wellEl = ref<HTMLElement | null>(null)
let chartInstance: Chart<'line', Array<number | null>, string> | null = null
let resizeObserver: ResizeObserver | null = null

const hasSaved = computed(() => props.points.some((pt) => pt.saved !== null))
const hasDraft = computed(() => props.points.some((pt) => pt.draft !== null))

const latestDisplay = computed(() => {
  for (let i = props.points.length - 1; i >= 0; i -= 1) {
    const pt = props.points[i]
    if (!pt) continue
    if (pt.draft !== null) {
      return { value: pt.draft, kind: 'Borrador', time: pt.label }
    }
    if (pt.saved !== null) {
      return { value: pt.saved, kind: 'Guardado', time: pt.label }
    }
  }
  return null
})

function destroyChart(): void {
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
}

function renderChart(): void {
  const canvas = canvasEl.value
  if (!canvas || props.points.length === 0) {
    destroyChart()
    return
  }

  const isDark = themeStore.mode === 'dark'
  const savedColor = isDark ? '#4ADE80' : '#166534'
  const savedFill = isDark ? 'rgba(74, 222, 128, 0.14)' : 'rgba(22, 101, 52, 0.10)'
  const draftColor = isDark ? '#E6AA3C' : '#B83A28'
  const draftFill = isDark ? 'rgba(230, 170, 60, 0.16)' : 'rgba(184, 58, 40, 0.12)'
  const gridColor = isDark ? 'rgba(107, 86, 58, 0.36)' : 'rgba(199, 182, 152, 0.55)'
  const tickColor = isDark ? '#CBBCA3' : '#433A2E'

  const labels = props.points.map((pt) => pt.label)
  const savedValues = props.points.map((pt) => pt.saved)
  const draftValues = props.points.map((pt) => pt.draft)

  const datasets: Array<ChartDataset<'line', Array<number | null>>> = []

  if (hasSaved.value) {
    datasets.push({
      label: 'Guardado',
      data: savedValues,
      borderColor: savedColor,
      backgroundColor: savedFill,
      borderWidth: 1.8,
      pointRadius: savedValues.filter((v) => v !== null).length <= 18 ? 2.5 : 0,
      pointHoverRadius: 4,
      pointBackgroundColor: savedColor,
      tension: 0.2,
      fill: !hasDraft.value,
      spanGaps: false,
    })
  }

  if (hasDraft.value) {
    datasets.push({
      label: 'Borrador',
      data: draftValues,
      borderColor: draftColor,
      backgroundColor: draftFill,
      borderWidth: 2.2,
      borderDash: [5, 3],
      pointRadius: 3,
      pointHoverRadius: 5,
      pointBackgroundColor: draftColor,
      tension: 0.18,
      fill: true,
      spanGaps: false,
    })
  }

  destroyChart()

  const unit = props.spec.unit
  const config: ChartConfiguration<'line', Array<number | null>, string> = {
    type: 'line',
    data: { labels, datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 150 },
      interaction: {
        mode: 'index',
        intersect: false,
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: isDark ? '#100D0A' : '#191D20',
          titleColor: '#F7F1E3',
          bodyColor: '#F7F1E3',
          borderColor: '#E7B043',
          borderWidth: 1,
          cornerRadius: 0,
          displayColors: true,
          callbacks: {
            title(items) {
              const first = items[0]
              return first ? `${first.label} UTC-5` : ''
            },
            label(item) {
              if (item.raw === null || item.raw === undefined) {
                return ''
              }
              return `${item.dataset.label}: ${Number(item.raw).toFixed(2)} ${unit}`
            },
          },
        },
      },
      scales: {
        x: {
          grid: { color: gridColor },
          ticks: {
            color: tickColor,
            maxTicksLimit: 6,
            font: {
              family: "'Azeret Mono', monospace",
              size: 10,
            },
          },
        },
        y: {
          grid: { color: gridColor },
          ticks: {
            color: tickColor,
            maxTicksLimit: 4,
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

onMounted(() => {
  renderChart()
  if (wellEl.value) {
    resizeObserver = new ResizeObserver(() => {
      chartInstance?.resize()
    })
    resizeObserver.observe(wellEl.value)
  }
})

watch(
  () => [props.points, themeStore.mode],
  () => {
    renderChart()
  },
  { flush: 'post' },
)

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  destroyChart()
})
</script>

<template>
  <article class="metric-chart-card">
    <header class="card-head">
      <div class="head-left">
        <strong class="metric-title">{{ spec.label }}</strong>
        <span class="metric-unit tabular">{{ spec.unit }}</span>
      </div>

      <div class="head-right tabular">
        <span v-if="latestDisplay" class="latest-val">
          {{ latestDisplay.value.toFixed(2) }}
        </span>
        <div class="series-legend">
          <span v-if="hasSaved" class="legend-chip is-saved">Guardado</span>
          <span v-if="hasDraft" class="legend-chip is-draft">Borrador</span>
        </div>
      </div>
    </header>

    <div ref="wellEl" class="canvas-well">
      <canvas ref="canvasEl" :aria-label="`Serie temporal individual de ${spec.label}`"></canvas>
    </div>
  </article>
</template>

<style scoped>
.metric-chart-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  width: 100%;
  overflow: hidden;
  padding: 10px 12px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.head-left {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.metric-title {
  font-family: var(--font-display);
  font-size: 0.98rem;
  font-weight: 800;
  color: var(--ink-primary);
}

.metric-unit {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
}

.head-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.latest-val {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 700;
  color: var(--ink-primary);
}

.series-legend {
  display: inline-flex;
  gap: 4px;
}

.legend-chip {
  padding: 2px 5px;
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 600;
  text-transform: uppercase;
}

.legend-chip.is-saved {
  border: 1px solid #22c55e;
  color: #166534;
}

:root[data-theme='dark'] .legend-chip.is-saved {
  border-color: #4ade80;
  color: #4ade80;
}

.legend-chip.is-draft {
  border: 1px solid var(--accent-vermilion);
  color: var(--accent-vermilion);
}

:root[data-theme='dark'] .legend-chip.is-draft {
  border-color: #e6aa3c;
  color: #e6aa3c;
}

.canvas-well {
  position: relative;
  height: 146px;
  width: 100%;
  min-width: 0;
}
</style>
