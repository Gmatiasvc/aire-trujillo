<script setup lang="ts">
import { RouterLink } from 'vue-router'
import UiIcon from '../UiIcon.vue'
import { INDIVIDUAL_LAYER_OPTIONS, formatLimaDateTime } from '../../data/legends'
import { useThemeStore } from '../../stores/theme'
import {
  LAYER_AIR_QUALITY,
  type AirQualityStandard,
  type LayerKey,
} from '../../types/telemetry'

const props = defineProps<{
  activeLayer: LayerKey
  activeStandard: AirQualityStandard
  activeNodeCount: number
  displayTimestamp: string | null
  isLive: boolean
}>()

const emit = defineEmits<{
  'select-standard': [standard: AirQualityStandard]
  'select-layer': [layer: LayerKey]
  'reset-view': []
}>()

const themeStore = useThemeStore()

function handleLayerSelect(event: Event): void {
  const select = event.target as HTMLSelectElement
  emit('select-layer', select.value as LayerKey)
}
</script>

<template>
  <header class="map-topbar">
    <div class="topbar-brand-group">
      <RouterLink to="/" class="back-link" aria-label="Volver al inicio">
        <UiIcon name="arrow-left" :size="15" />
        <span class="back-label">Inicio</span>
      </RouterLink>
      <div class="brand-divider" aria-hidden="true"></div>
      <div class="brand-title-wrap">
        <h1 class="map-heading">AIRE TRUJILLO</h1>
        <span class="node-count-pill tabular">
          <i class="live-square" :class="{ 'is-live': props.isLive }" aria-hidden="true"></i>
          <span>{{ props.activeNodeCount }}</span>
          <span class="node-word">nodos activos</span>
        </span>
      </div>
    </div>

    <div class="topbar-controls-group">
      <div class="standard-tabs" role="group" aria-label="Estándar de calidad del aire">
        <button
          type="button"
          class="std-tab"
          :class="{
            'is-active':
              props.activeLayer === LAYER_AIR_QUALITY && props.activeStandard === 'inca',
          }"
          :aria-pressed="
            props.activeLayer === LAYER_AIR_QUALITY && props.activeStandard === 'inca'
          "
          @click="emit('select-standard', 'inca')"
        >
          <span><span class="std-long">ECA / </span>INCA<span class="std-long"> (Perú)</span></span>
        </button>
        <button
          type="button"
          class="std-tab"
          :class="{
            'is-active':
              props.activeLayer === LAYER_AIR_QUALITY && props.activeStandard === 'epa',
          }"
          :aria-pressed="
            props.activeLayer === LAYER_AIR_QUALITY && props.activeStandard === 'epa'
          "
          @click="emit('select-standard', 'epa')"
        >
          <span>AQI<span class="std-long"> (EPA)</span></span>
        </button>
      </div>

      <label class="layer-select-wrap">
        <UiIcon name="layers" :size="14" />
        <select
          class="layer-select"
          :value="props.activeLayer"
          aria-label="Seleccionar capa o variable ambiental"
          @change="handleLayerSelect"
        >
          <option :value="LAYER_AIR_QUALITY">
            Índice compuesto ({{ props.activeStandard === 'inca' ? 'ECA / INCA' : 'AQI EPA' }})
          </option>
          <optgroup label="12 variables individuales">
            <option v-for="opt in INDIVIDUAL_LAYER_OPTIONS" :key="opt.key" :value="opt.key">
              {{ opt.label }} ({{ opt.unit }})
            </option>
          </optgroup>
        </select>
        <UiIcon name="chevron-down" :size="14" class="select-caret" />
      </label>
    </div>

    <div class="topbar-meta-group">
      <span class="clock-chip tabular" title="Marca temporal activa en hora local de Perú">
        {{ formatLimaDateTime(props.displayTimestamp) }}
      </span>

      <button
        type="button"
        class="tool-btn"
        title="Centrar vista en Trujillo y Huanchaco"
        aria-label="Centrar vista en Trujillo y Huanchaco"
        @click="emit('reset-view')"
      >
        <UiIcon name="locate" :size="15" />
        <span class="btn-label">Centrar</span>
      </button>

      <button
        type="button"
        class="tool-btn"
        :aria-label="
          themeStore.mode === 'light'
            ? 'Cambiar a modo oscuro cálido'
            : 'Cambiar a modo claro cálido'
        "
        @click="themeStore.toggle"
      >
        <UiIcon :name="themeStore.mode === 'light' ? 'moon' : 'sun'" :size="15" />
        <span class="btn-label">{{ themeStore.mode === 'light' ? 'Oscuro' : 'Claro' }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.map-topbar {
  position: relative;
  z-index: 20;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 16px;
  min-height: 56px;
  padding: 8px 16px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border-bottom: 1.5px solid var(--rule-color);
}

.topbar-brand-group,
.topbar-controls-group,
.topbar-meta-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--glass-slate-rule);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  text-transform: uppercase;
  text-decoration: none;
  transition:
    background-color 0.18s var(--ease-expo),
    border-color 0.18s var(--ease-expo);
}

.back-link:hover {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.brand-divider {
  width: 1px;
  height: 22px;
  background: var(--glass-slate-rule);
}

.brand-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.map-heading {
  font-family: var(--font-display);
  font-size: 1.45rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  line-height: 1;
  white-space: nowrap;
  color: var(--ink-on-carbon);
}

.node-count-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border: 1px solid var(--glass-slate-rule);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

:root[data-theme='light'] .node-count-pill {
  background: var(--surface-slate);
  color: var(--ink-on-carbon);
}

.live-square {
  width: 7px;
  height: 7px;
  background: var(--nodata-bg);
}

.live-square.is-live {
  background: #22c55e;
}

.standard-tabs {
  display: inline-flex;
  gap: 4px;
}

.std-tab {
  padding: 7px 13px;
  border: 1px solid var(--glass-slate-rule);
  background: transparent;
  color: var(--ink-muted-on-carbon);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: 0.04em;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.18s var(--ease-expo),
    color 0.18s var(--ease-expo),
    border-color 0.18s var(--ease-expo);
}

.std-tab > span {
  display: inline-block;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.std-tab:hover {
  color: var(--ink-on-carbon);
  border-color: var(--accent-amber-fill);
}

.std-tab.is-active {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.layer-select-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0 10px;
  height: 33px;
  border: 1px solid var(--glass-slate-rule);
  background: var(--surface-ivory);
  color: var(--ink-primary);
}

.layer-select {
  appearance: none;
  border: none;
  background: transparent;
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  padding-right: 18px;
  height: 100%;
  cursor: pointer;
  min-width: 215px;
}

.layer-select:focus-visible {
  outline: none;
}

.layer-select-wrap:focus-within {
  outline: 2px solid var(--focus-ring);
  outline-offset: 2px;
}

.select-caret {
  position: absolute;
  right: 8px;
  pointer-events: none;
}

.clock-chip {
  padding: 5px 9px;
  border: 1px solid var(--accent-amber);
  background: var(--accent-amber-fill);
  color: #181a1c;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--glass-slate-rule);
  background: transparent;
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition:
    background-color 0.18s var(--ease-expo),
    border-color 0.18s var(--ease-expo);
}

.tool-btn:hover {
  background: var(--surface-slate);
  border-color: var(--accent-amber-fill);
}

@media (max-width: 1100px) {
  .clock-chip {
    display: none;
  }
}

@media (max-width: 780px) {
  .map-topbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    grid-template-rows: auto auto;
    align-items: center;
    padding: 7px 10px;
    gap: 6px 8px;
  }

  .topbar-brand-group {
    grid-column: 1;
    grid-row: 1;
    min-width: 0;
    gap: 7px;
  }

  .topbar-meta-group {
    grid-column: 2;
    grid-row: 1;
    gap: 6px;
  }

  .topbar-controls-group {
    grid-column: 1 / -1;
    grid-row: 2;
    flex-wrap: nowrap;
    gap: 6px;
  }

  .back-link {
    padding: 5px 8px;
  }

  .brand-divider {
    display: none;
  }

  .map-heading {
    font-size: 1.15rem;
  }

  .node-count-pill {
    padding: 2px 6px;
    font-size: 10px;
  }

  .std-tab {
    padding: 5px 9px;
    font-size: 11px;
    white-space: nowrap;
  }

  .layer-select-wrap {
    flex: 1 1 0;
    min-width: 0;
    height: 30px;
    padding: 0 8px;
  }

  .layer-select {
    width: 100%;
    min-width: 0;
    font-size: 11px;
  }

  .tool-btn {
    padding: 5px 8px;
  }

  .btn-label {
    display: none;
  }
}

@media (max-width: 520px) {
  .back-label,
  .node-word,
  .std-long {
    display: none;
  }
}
</style>
