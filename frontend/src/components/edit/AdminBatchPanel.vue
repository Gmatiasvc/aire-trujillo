<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import UiIcon from '../UiIcon.vue'
import AdminSeriesPreviewGrid from './AdminSeriesPreviewGrid.vue'
import { formatLimaDateTime } from '../../data/legends'
import {
  METRIC_KEYS,
  METRIC_SPECS,
  type AdminReadingItem,
  type BatchReadingsRequest,
  type BatchReadingsResponse,
  type MetricKey,
  type Node,
} from '../../types/telemetry'
import {
  ADMIN_BATCH_TEMPLATES,
  LIMA_FIVE_MIN_SLOTS,
  buildRfc3339LimaStartTime,
  createEmptyDraftMap,
  getCurrentLimaDateAndSlot,
  parseCommaMetricInput,
  type BatchTemplatePreset,
  type ParsedMetricInput,
  type PreviewChartMode,
} from '../../utils/adminSeries'

const props = defineProps<{
  selectedNode: Node | null
  savedReadings: AdminReadingItem[]
  isLoadingReadings: boolean
  isSavingBatch: boolean
  lastBatchResult: BatchReadingsResponse | null
  storeError: string | null
}>()

const emit = defineEmits<{
  'submit-batch': [nodeId: string, payload: BatchReadingsRequest]
}>()

const initialClock = getCurrentLimaDateAndSlot()
const batchDate = ref(initialClock.dateYmd)
const batchTimeSlot = ref(initialClock.hhmm)
const draftTexts = reactive<Record<MetricKey, string>>(createEmptyDraftMap())
const previewMode = ref<PreviewChartMode>('combined')

const startTimeIso = computed(() => buildRfc3339LimaStartTime(batchDate.value, batchTimeSlot.value))

const parsedDrafts = computed<Record<MetricKey, ParsedMetricInput>>(() => {
  const result = {} as Record<MetricKey, ParsedMetricInput>
  for (const key of METRIC_KEYS) {
    result[key] = parseCommaMetricInput(draftTexts[key], METRIC_SPECS[key])
  }
  return result
})

const validationErrors = computed<string[]>(() => {
  const errs: string[] = []
  if (!startTimeIso.value) {
    errs.push('Selecciona una fecha y hora 24 h válida en múltiplo de 5 min (UTC-5).')
  }
  for (const key of METRIC_KEYS) {
    const err = parsedDrafts.value[key].error
    if (err) {
      errs.push(`${METRIC_SPECS[key].label}: ${err}`)
    }
  }
  return errs
})

const draftSummary = computed(() => {
  let metricCount = 0
  let totalPoints = 0
  for (const key of METRIC_KEYS) {
    const len = parsedDrafts.value[key].values.length
    if (len > 0) {
      metricCount += 1
      totalPoints += len
    }
  }
  return { metricCount, totalPoints }
})

const canSubmit = computed(
  () =>
    Boolean(props.selectedNode) &&
    Boolean(startTimeIso.value) &&
    draftSummary.value.metricCount > 0 &&
    validationErrors.value.length === 0 &&
    !props.isSavingBatch,
)

function handleSetNowLima(): void {
  const now = getCurrentLimaDateAndSlot()
  batchDate.value = now.dateYmd
  batchTimeSlot.value = now.hhmm
}

function applyTemplate(preset: BatchTemplatePreset): void {
  for (const key of METRIC_KEYS) {
    draftTexts[key] = preset.metrics[key] ?? ''
  }
}

function clearDrafts(): void {
  for (const key of METRIC_KEYS) {
    draftTexts[key] = ''
  }
}

function handleSubmitBatch(): void {
  if (!canSubmit.value || !props.selectedNode || !startTimeIso.value) return
  const metricsPayload: Partial<Record<MetricKey, number[]>> = {}
  for (const key of METRIC_KEYS) {
    const values = parsedDrafts.value[key].values
    if (values.length > 0) {
      metricsPayload[key] = values
    }
  }
  emit('submit-batch', props.selectedNode.id, {
    start_time: startTimeIso.value,
    metrics: metricsPayload,
  })
}
</script>

<template>
  <section class="batch-panel-card" aria-label="Carga de series simuladas cada 5 minutos">
    <header class="batch-head">
      <div class="head-title-wrap">
        <h2 class="panel-title">Telemetría por lotes (5 min)</h2>
        <span v-if="selectedNode" class="target-node-chip tabular">
          {{ selectedNode.code }} · {{ selectedNode.name }}
        </span>
        <span v-else class="target-node-chip is-empty">
          Selecciona o crea una estación
        </span>
      </div>
    </header>

    <div class="anchor-toolbar">
      <div class="clock-fields">
        <label class="anchor-field">
          <span class="field-label">Fecha (UTC-5)</span>
          <input
            v-model="batchDate"
            type="date"
            class="mono-input tabular"
            aria-label="Fecha de referencia en America Lima"
          />
        </label>

        <label class="anchor-field">
          <span class="field-label">Hora fin / más reciente (24 h)</span>
          <div class="select-wrap">
            <select
              v-model="batchTimeSlot"
              class="mono-select tabular"
              aria-label="Hora en múltiplos de 5 minutos formato 24 horas"
            >
              <option v-for="slot in LIMA_FIVE_MIN_SLOTS" :key="slot" :value="slot">
                {{ slot }} (UTC-5)
              </option>
            </select>
            <UiIcon name="chevron-down" :size="14" class="select-caret" />
          </div>
        </label>

        <button
          type="button"
          class="now-btn"
          aria-label="Ajustar fecha y hora al instante actual de Trujillo"
          @click="handleSetNowLima"
        >
          <UiIcon name="clock" :size="13" />
          <span>Ahora (Lima)</span>
        </button>
      </div>

      <p class="anchor-rule-copy tabular">
        Índice 0 corresponde a <strong>{{ batchDate }} {{ batchTimeSlot }} UTC-5</strong>; cada coma
        siguiente retrocede 5 min (máx. 288 puntos).
      </p>
    </div>

    <div class="template-bar">
      <span class="template-label">Plantillas rápidas:</span>
      <div class="template-buttons">
        <button
          v-for="preset in ADMIN_BATCH_TEMPLATES"
          :key="preset.id"
          type="button"
          class="preset-btn"
          :title="preset.summary"
          @click="applyTemplate(preset)"
        >
          {{ preset.label }}
        </button>
        <button
          v-if="draftSummary.metricCount > 0"
          type="button"
          class="clear-btn"
          @click="clearDrafts"
        >
          Limpiar campos
        </button>
      </div>
    </div>

    <div class="metrics-input-grid">
      <label
        v-for="key in METRIC_KEYS"
        :key="key"
        class="metric-input-card"
        :class="{
          'has-error': Boolean(parsedDrafts[key].error),
          'has-values': parsedDrafts[key].values.length > 0 && !parsedDrafts[key].error,
        }"
      >
        <div class="metric-card-top">
          <span class="metric-name">
            {{ METRIC_SPECS[key].label }}
            <span class="metric-unit tabular">({{ METRIC_SPECS[key].unit }})</span>
          </span>
          <span v-if="parsedDrafts[key].error" class="pts-badge is-err">Error</span>
          <span v-else-if="parsedDrafts[key].values.length > 0" class="pts-badge is-ok tabular">
            {{ parsedDrafts[key].values.length }} pts
          </span>
          <span v-else class="metric-bounds tabular">
            {{ METRIC_SPECS[key].min }}–{{ METRIC_SPECS[key].max }}
          </span>
        </div>

        <input
          v-model="draftTexts[key]"
          type="text"
          class="metric-csv-input tabular"
          placeholder="Ej. 15.4, 14.8, 14.2"
          :aria-label="`Valores separados por comas para ${METRIC_SPECS[key].label}`"
        />

        <span v-if="parsedDrafts[key].error" class="metric-err-msg">
          {{ parsedDrafts[key].error }}
        </span>
      </label>
    </div>

    <div class="submit-toolbar">
      <div class="submit-meta">
        <span class="draft-counter tabular">
          {{ draftSummary.totalPoints }} puntos listos en {{ draftSummary.metricCount }} variables
        </span>

        <p v-if="validationErrors.length > 0" class="batch-alert is-error" role="alert">
          {{ validationErrors[0] }}
        </p>
        <p v-else-if="storeError" class="batch-alert is-error" role="alert">
          {{ storeError }}
        </p>
        <p v-else-if="lastBatchResult" class="batch-alert is-ok" role="status">
          Guardado: {{ lastBatchResult.inserted_or_updated }} puntos ·
          {{ formatLimaDateTime(lastBatchResult.from) }} → {{ formatLimaDateTime(lastBatchResult.to) }}
        </p>
      </div>

      <button
        type="button"
        class="raked-save-batch-btn"
        :disabled="!canSubmit"
        @click="handleSubmitBatch"
      >
        <span>
          <UiIcon name="check" :size="14" />
          <span>{{ isSavingBatch ? 'Guardando lote…' : 'Guardar lote' }}</span>
        </span>
      </button>
    </div>

    <div class="section-divider" aria-hidden="true"></div>

    <AdminSeriesPreviewGrid
      :saved-readings="savedReadings"
      :parsed-drafts="parsedDrafts"
      :start-time-iso="startTimeIso"
      :is-loading-readings="isLoadingReadings"
      :preview-mode="previewMode"
      @update:preview-mode="previewMode = $event"
    />
  </section>
</template>

<style scoped>
.batch-panel-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-width: 0;
  width: 100%;
  padding: 16px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
}

.batch-head {
  padding-bottom: 10px;
  border-bottom: 1px solid var(--rule-soft);
}

.head-title-wrap {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
}

.panel-title {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--ink-primary);
}

.target-node-chip {
  padding: 3px 9px;
  border: 1px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.target-node-chip.is-empty {
  background: var(--surface-ivory-2);
  color: var(--ink-secondary);
}

.anchor-toolbar {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.clock-fields {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 10px;
}

.anchor-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1 1 160px;
  min-width: 0;
}

.field-label {
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--ink-secondary);
}

.mono-input,
.mono-select {
  width: 100%;
  padding: 7px 10px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
}

.select-wrap {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.mono-select {
  appearance: none;
  padding-right: 28px;
  min-width: 0;
  cursor: pointer;
}

.select-caret {
  position: absolute;
  right: 8px;
  pointer-events: none;
  color: var(--ink-secondary);
}

.now-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 34px;
  padding: 0 11px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition:
    background-color 0.16s var(--ease-expo),
    border-color 0.16s var(--ease-expo);
}

.now-btn:hover {
  border-color: var(--accent-vermilion);
}

.anchor-rule-copy {
  font-size: 11.5px;
  color: var(--ink-secondary);
}

.template-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.template-label {
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--ink-secondary);
}

.template-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.preset-btn,
.clear-btn {
  padding: 5px 9px;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background-color 0.14s var(--ease-expo),
    border-color 0.14s var(--ease-expo),
    color 0.14s var(--ease-expo);
}

.preset-btn:hover {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.clear-btn {
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion);
}

.clear-btn:hover {
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.metrics-input-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 9px;
}

.metric-input-card {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 8px 10px;
  border: 1px solid var(--rule-soft);
  background: var(--surface-ivory-2);
  transition: border-color 0.14s var(--ease-expo);
}

.metric-input-card.has-values {
  border-color: var(--rule-color);
}

.metric-input-card.has-error {
  border-color: var(--accent-vermilion);
}

.metric-card-top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 6px;
}

.metric-name {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  color: var(--ink-primary);
}

.metric-unit,
.metric-bounds {
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 500;
  color: var(--ink-secondary);
}

.pts-badge {
  padding: 1px 5px;
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 700;
}

.pts-badge.is-ok {
  border: 1px solid #22c55e;
  color: #166534;
}

:root[data-theme='dark'] .pts-badge.is-ok {
  border-color: #4ade80;
  color: #4ade80;
}

.pts-badge.is-err {
  border: 1px solid var(--accent-vermilion);
  color: var(--accent-vermilion);
}

.metric-csv-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11.5px;
}

.metric-csv-input:focus-visible,
.mono-input:focus-visible,
.mono-select:focus-visible,
.now-btn:focus-visible,
.preset-btn:focus-visible,
.clear-btn:focus-visible,
.raked-save-batch-btn:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.metric-err-msg {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--accent-vermilion);
}

.submit-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 16px;
  padding: 10px 12px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.submit-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.draft-counter {
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  color: var(--ink-primary);
}

.batch-alert {
  font-family: var(--font-mono);
  font-size: 11px;
}

.batch-alert.is-error {
  color: var(--accent-vermilion);
}

.batch-alert.is-ok {
  color: #166534;
}

:root[data-theme='dark'] .batch-alert.is-ok {
  color: #4ade80;
}

.raked-save-batch-btn {
  display: inline-flex;
  align-items: center;
  padding: 9px 18px;
  border: 1.5px solid var(--accent-vermilion);
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition: opacity 0.16s var(--ease-expo);
}

.raked-save-batch-btn > span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.raked-save-batch-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.section-divider {
  height: 1px;
  background: var(--rule-color);
}

@media (max-width: 960px) {
  .metrics-input-grid {
    grid-template-columns: 1fr;
  }
}
</style>
