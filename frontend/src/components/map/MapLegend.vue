<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import UiIcon from '../UiIcon.vue'
import type { LegendSpec } from '../../data/legends'

const props = defineProps<{
  legend: LegendSpec
}>()

const scaleGradient = computed<string>(() => {
  const colors = props.legend.bands.map((band) => band.color)
  if (colors.length === 0) return 'none'
  if (colors.length === 1) return colors[0]!
  return `linear-gradient(90deg, ${colors.join(', ')})`
})

const isCollapsed = ref<boolean>(typeof window !== 'undefined' ? window.innerWidth <= 900 : false)
let mobileMediaQuery: MediaQueryList | null = null

function handleViewportChange(ev: MediaQueryListEvent): void {
  isCollapsed.value = ev.matches
}

onMounted(() => {
  if (typeof window === 'undefined') return
  mobileMediaQuery = window.matchMedia('(max-width: 900px)')
  isCollapsed.value = mobileMediaQuery.matches
  mobileMediaQuery.addEventListener('change', handleViewportChange)
})

onBeforeUnmount(() => {
  mobileMediaQuery?.removeEventListener('change', handleViewportChange)
})
</script>

<template>
  <aside class="legend-panel" aria-label="Leyenda del mapa">
    <button
      type="button"
      class="legend-header"
      :aria-expanded="!isCollapsed"
      @click="isCollapsed = !isCollapsed"
    >
      <span class="legend-header-text">
        <strong>{{ legend.title }}</strong>
        <small class="tabular">{{ legend.unit }}</small>
      </span>
      <UiIcon
        name="chevron-down"
        :size="15"
        class="chevron-icon"
        :class="{ 'is-collapsed': isCollapsed }"
      />
    </button>

    <div v-if="!isCollapsed" class="legend-body">
      <p class="legend-subtitle">{{ legend.subtitle }}</p>

      <div
        class="scale-ramp"
        :style="{ background: scaleGradient }"
        aria-hidden="true"
      ></div>

      <ul class="legend-bands">
        <li v-for="band in legend.bands" :key="band.label" class="legend-row">
          <span class="swatch" :style="{ backgroundColor: band.color }" aria-hidden="true"></span>
          <span class="band-label">{{ band.label }}</span>
          <span class="band-range tabular">{{ band.range }}</span>
        </li>
      </ul>

      <div class="fade-block">
        <div class="fade-labels tabular">
          <span>Pleno / Entre nodos</span>
          <span>Borde ext. 400–600 m</span>
        </div>
        <div class="fade-ramp" aria-hidden="true"></div>
      </div>

      <div class="nodata-row">
        <span
          class="swatch swatch-nodata"
          :style="{ backgroundColor: legend.noDataColor }"
          aria-hidden="true"
        ></span>
        <span class="nodata-text">
          <strong>No hay datos</strong>
          <small>Sin cobertura</small>
        </span>
        <span class="nodata-radius tabular">&gt; 600 m</span>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.legend-panel {
  width: 268px;
  max-width: calc(100vw - 24px);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  border: 1.5px solid var(--rule-color);
}

.legend-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 12px;
  border: none;
  border-bottom: 1px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  cursor: pointer;
  text-align: left;
}

.legend-header-text {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  flex: 1;
}

.legend-header-text strong {
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.legend-header-text small {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--accent-amber-fill);
}

.chevron-icon {
  transition: transform 0.2s var(--ease-expo);
}

.chevron-icon.is-collapsed {
  transform: rotate(180deg);
}

.legend-body {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.legend-subtitle {
  font-size: 11.5px;
  line-height: 1.35;
  color: var(--ink-secondary);
}

.scale-ramp {
  height: 8px;
  border: 1px solid var(--rule-color);
}

.legend-bands {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.legend-row {
  display: grid;
  grid-template-columns: 14px 1fr auto;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.swatch {
  width: 14px;
  height: 14px;
  border: 1px solid var(--rule-color);
}

.band-label {
  font-weight: 600;
  color: var(--ink-primary);
}

.band-range {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
}

.fade-block {
  padding-top: 8px;
  border-top: 1px solid var(--rule-soft);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.fade-labels {
  display: flex;
  justify-content: space-between;
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--ink-secondary);
}

.fade-ramp {
  height: 8px;
  border: 1px solid var(--rule-color);
  background: linear-gradient(90deg, #22c55e 0%, #22c55e 45%, #9ca3af 100%);
}

.nodata-row {
  display: grid;
  grid-template-columns: 14px 1fr auto;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: var(--nodata-bg);
  color: var(--nodata-ink);
  border: 1px solid var(--rule-color);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.nodata-text {
  display: flex;
  flex-direction: column;
  line-height: 1.25;
}

.nodata-text strong {
  white-space: nowrap;
}

.nodata-text small {
  font-size: 10px;
  font-weight: 500;
}

.swatch-nodata {
  border-color: #111827;
}

.nodata-radius {
  font-size: 10px;
  white-space: nowrap;
}

@media (max-width: 900px) {
  .legend-panel {
    width: 236px;
  }

  .legend-body {
    padding: 8px 10px;
    gap: 7px;
    max-height: 40svh;
    overflow-y: auto;
  }
}
</style>
