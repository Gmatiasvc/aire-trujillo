<script setup lang="ts">
import { computed } from 'vue'
import AdminMetricChartCard from './AdminMetricChartCard.vue'
import { METRIC_KEYS, METRIC_SPECS, type AdminReadingItem, type MetricKey } from '../../types/telemetry'
import {
  buildVariableSeriesPoints,
  type ParsedMetricInput,
  type PreviewChartMode,
} from '../../utils/adminSeries'

const props = defineProps<{
  savedReadings: AdminReadingItem[]
  parsedDrafts: Record<MetricKey, ParsedMetricInput>
  startTimeIso: string | null
  isLoadingReadings: boolean
  previewMode: PreviewChartMode
}>()

const emit = defineEmits<{
  'update:preview-mode': [mode: PreviewChartMode]
}>()

const MODE_OPTIONS: ReadonlyArray<{ id: PreviewChartMode; label: string; title: string }> = [
  { id: 'combined', label: 'Combinado', title: 'Combinado: Guardado + Borrador' },
  { id: 'draft', label: 'Solo borrador', title: 'Solo valores en borrador antes de guardar' },
  { id: 'saved', label: 'Solo guardado', title: 'Solo lecturas persistidas de las últimas 24 h' },
]

const activeCards = computed(() => {
  return METRIC_KEYS.map((key) => {
    const draftValues = props.parsedDrafts[key]?.values ?? []
    const points = buildVariableSeriesPoints(
      key,
      props.savedReadings,
      draftValues,
      props.startTimeIso,
      props.previewMode,
    )
    return {
      key,
      spec: METRIC_SPECS[key],
      points,
    }
  }).filter((item) => item.points.length > 0)
})

const emptyHint = computed(() => {
  if (props.previewMode === 'draft') {
    return 'Escribe valores separados por comas o elige una plantilla rápida para previsualizar cada variable antes de guardar.'
  }
  if (props.previewMode === 'saved') {
    return 'Esta estación aún no registra lecturas guardadas en las últimas 24 horas.'
  }
  return 'Sin lecturas guardadas ni borradores activos para graficar. Ingresa valores o aplica una plantilla.'
})
</script>

<template>
  <section class="preview-grid-section" aria-label="Gráficos individuales por variable">
    <header class="preview-head">
      <div class="head-meta">
        <h3 class="preview-title">Previsualización por variable</h3>
        <span class="preview-count tabular">{{ activeCards.length }} activas</span>
      </div>

      <div class="mode-tabs" role="group" aria-label="Modo de visualización de series">
        <button
          v-for="opt in MODE_OPTIONS"
          :key="opt.id"
          type="button"
          class="mode-tab"
          :class="{ 'is-active': previewMode === opt.id }"
          :title="opt.title"
          :aria-pressed="previewMode === opt.id"
          @click="emit('update:preview-mode', opt.id)"
        >
          <span>{{ opt.label }}</span>
        </button>
      </div>
    </header>

    <div v-if="isLoadingReadings && activeCards.length === 0" class="loading-box" aria-busy="true">
      Cargando histórico reciente de 24 h…
    </div>

    <div v-else-if="activeCards.length === 0" class="empty-box">
      <p class="empty-text">{{ emptyHint }}</p>
    </div>

    <div v-else class="charts-grid">
      <AdminMetricChartCard
        v-for="card in activeCards"
        :key="card.key"
        :spec="card.spec"
        :points="card.points"
      />
    </div>
  </section>
</template>

<style scoped>
.preview-grid-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  padding-top: 4px;
}

.preview-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--rule-soft);
}

.head-meta {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.preview-title {
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 800;
  text-transform: uppercase;
  color: var(--ink-primary);
}

.preview-count {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
}

.mode-tabs {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
}

.mode-tab {
  padding: 5px 10px;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-secondary);
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.16s var(--ease-expo),
    color 0.16s var(--ease-expo),
    border-color 0.16s var(--ease-expo);
}

.mode-tab > span {
  display: inline-block;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.mode-tab:hover {
  color: var(--ink-primary);
  border-color: var(--accent-vermilion);
}

.mode-tab.is-active {
  background: var(--surface-carbon);
  border-color: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.mode-tab:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.loading-box,
.empty-box {
  padding: 18px 14px;
  border: 1px dashed var(--rule-color);
  background: var(--surface-ivory-2);
  text-align: center;
}

.loading-box,
.empty-text {
  font-size: 12.5px;
  color: var(--ink-secondary);
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.charts-grid > * {
  min-width: 0;
}

@media (max-width: 820px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
}
</style>
