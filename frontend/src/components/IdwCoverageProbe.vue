<script setup lang="ts">
import { computed, ref } from 'vue'
import { COVERAGE_CHECKPOINTS, type CoverageCheckpoint } from '../data/demoSnapshot'

const selectedIndex = ref<number>(2)

const activeCheckpoint = computed<CoverageCheckpoint>(() => {
  return COVERAGE_CHECKPOINTS[selectedIndex.value] ?? COVERAGE_CHECKPOINTS[0]!
})

const probePositionX = computed<number>(() => {
  const maxViewDistance = 750
  const minX = 68
  const maxX = 452
  const clamped = Math.min(activeCheckpoint.value.distanceM, maxViewDistance)
  return Math.round(minX + (clamped / maxViewDistance) * (maxX - minX))
})
</script>

<template>
  <div class="probe-layout">
    <div class="probe-visual-pane">
      <div class="pane-strip">
        <span class="mono-label">SONDA ESPACIAL IDW · GRILLA 50 M</span>
        <span class="mono-label tabular">R_FADE 400 M · R_MAX 600 M</span>
      </div>

      <div class="checkpoint-tabs" role="tablist" aria-label="Distancia al nodo más cercano">
        <button
          v-for="(cp, index) in COVERAGE_CHECKPOINTS"
          :key="cp.id"
          type="button"
          role="tab"
          :aria-selected="selectedIndex === index"
          class="cp-tab tabular"
          :class="{ 'is-active': selectedIndex === index }"
          @click="selectedIndex = index"
        >
          <span>{{ cp.distanceM }} m</span>
        </button>
      </div>

      <figure class="cross-section" aria-label="Perfil de decaimiento espacial desde un nodo activo">
        <svg viewBox="0 0 500 184" class="section-svg" role="img" aria-labelledby="idw-svg-title">
          <title id="idw-svg-title">Perfil de distancia IDW desde 0 m hasta más de 600 m</title>
          <defs>
            <linearGradient id="fade-zone-fill" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stop-color="#22C55E" stop-opacity="0.45" />
              <stop offset="100%" stop-color="#9CA3AF" stop-opacity="0.92" />
            </linearGradient>
          </defs>

          <!-- Zone bands scaled across x=68 (0m) to x=452 (750m): 1m = 0.512px -->
          <rect x="68" y="28" width="205" height="96" fill="#22C55E" fill-opacity="0.28" />
          <rect x="273" y="28" width="102" height="96" fill="url(#fade-zone-fill)" />
          <rect x="375" y="28" width="97" height="96" fill="#9CA3AF" fill-opacity="0.88" />

          <!-- 50m cell grid dividers -->
          <g stroke="currentColor" stroke-opacity="0.22" stroke-width="1">
            <line x1="119" y1="28" x2="119" y2="124" />
            <line x1="170" y1="28" x2="170" y2="124" />
            <line x1="222" y1="28" x2="222" y2="124" />
            <line x1="273" y1="28" x2="273" y2="124" />
            <line x1="324" y1="28" x2="324" y2="124" />
            <line x1="375" y1="28" x2="375" y2="124" />
            <line x1="426" y1="28" x2="426" y2="124" />
          </g>

          <!-- Structural boundaries -->
          <line x1="68" y1="22" x2="68" y2="132" stroke="currentColor" stroke-width="2" />
          <line x1="273" y1="22" x2="273" y2="132" stroke="currentColor" stroke-width="1.5" stroke-dasharray="4 3" />
          <line x1="375" y1="22" x2="375" y2="132" stroke="currentColor" stroke-width="2" />
          <rect x="68" y="28" width="404" height="96" fill="none" stroke="currentColor" stroke-width="1.5" />

          <!-- Active station marker at 0m -->
          <circle cx="68" cy="76" r="7" fill="#B83A28" stroke="#F7F1E3" stroke-width="2" />

          <!-- Moving probe cursor -->
          <g class="probe-marker" :style="{ transform: `translateX(${probePositionX}px)` }">
            <line x1="0" y1="14" x2="0" y2="138" stroke="#B83A28" stroke-width="2" />
            <circle cx="0" cy="76" r="8" :fill="activeCheckpoint.hexColor" stroke="#181A1C" stroke-width="2" />
          </g>

          <!-- Zone footer labels -->
          <text x="68" y="154" class="svg-mono">0 m · NODO</text>
          <text x="273" y="154" text-anchor="middle" class="svg-mono">400 m</text>
          <text x="375" y="154" text-anchor="middle" class="svg-mono">600 m</text>
          <text x="444" y="154" text-anchor="middle" class="svg-mono">#9CA3AF</text>
        </svg>
      </figure>
    </div>

    <div class="probe-readout-pane">
      <div class="pane-strip">
        <span class="mono-label">RESPUESTA DEL BACKEND</span>
        <span class="mono-label tabular">GET /api/v1/map/point</span>
      </div>

      <Transition name="slide-swap" mode="out-in">
        <div :key="activeCheckpoint.id" class="readout-card">
          <div class="readout-swatch-row">
            <span
              class="swatch-chip tabular"
              :style="{
                backgroundColor: activeCheckpoint.hexColor,
                color: '#111827',
              }"
            >
              {{ activeCheckpoint.hexColor }}
            </span>
            <span class="distance-pill tabular">d_min = {{ activeCheckpoint.distanceM }} m</span>
          </div>

          <div class="readout-metric tabular">{{ activeCheckpoint.valueDisplay }}</div>
          <p class="readout-category">{{ activeCheckpoint.categoryLabel }}</p>

          <dl class="contract-fields tabular">
            <div>
              <dt>has_data</dt>
              <dd>{{ activeCheckpoint.hasData ? 'true' : 'false' }}</dd>
            </div>
            <div>
              <dt>label</dt>
              <dd>{{ activeCheckpoint.hasData ? 'ECA/INCA' : 'No hay datos' }}</dd>
            </div>
            <div>
              <dt>current.confidence</dt>
              <dd>{{ activeCheckpoint.hasData ? activeCheckpoint.confidence.toFixed(2) : 'null' }}</dd>
            </div>
            <div>
              <dt>current.value</dt>
              <dd>{{ activeCheckpoint.hasData ? '42.0' : 'null' }}</dd>
            </div>
          </dl>

          <p v-if="!activeCheckpoint.hasData" class="contract-null-note tabular">
            PointResponse.current = null (fuera de R_max = 600 m)
          </p>

          <p class="readout-note">{{ activeCheckpoint.note }}</p>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.probe-layout {
  display: grid;
  grid-template-columns: 1.25fr 0.95fr;
  gap: 20px;
}

.probe-visual-pane,
.probe-readout-pane {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  display: flex;
  flex-direction: column;
}

.pane-strip {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border-bottom: 1.5px solid var(--rule-color);
}

.mono-label {
  font-family: var(--font-mono);
  font-size: 11px;
  letter-spacing: 0.06em;
}

.checkpoint-tabs {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  padding: 14px 16px 6px;
}

.cp-tab {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  padding: 9px 6px;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.2s var(--ease-expo),
    color 0.2s var(--ease-expo),
    transform 0.2s var(--ease-expo);
}

.cp-tab > span {
  display: block;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.cp-tab:hover {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  transform: skewX(var(--rake-deg)) translateY(-2px);
}

.cp-tab.is-active {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.cross-section {
  padding: 12px 16px 16px;
  flex: 1;
  display: flex;
  align-items: center;
}

.section-svg {
  width: 100%;
  height: auto;
  color: var(--ink-primary);
}

.probe-marker {
  transition: transform 0.36s var(--ease-expo);
}

.svg-mono {
  fill: var(--ink-secondary);
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
}

.readout-card {
  padding: 18px 20px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.readout-swatch-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.swatch-chip,
.distance-pill {
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  padding: 4px 10px;
  border: 1px solid var(--rule-color);
}

.distance-pill {
  background: var(--surface-ivory);
  color: var(--ink-primary);
}

.readout-metric {
  font-family: var(--font-display);
  font-size: clamp(2rem, 3.2vw, 2.75rem);
  font-weight: 800;
  line-height: 1;
  letter-spacing: -0.02em;
  text-transform: uppercase;
}

.readout-category {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-secondary);
}

.contract-fields {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory);
}

.contract-fields > div {
  padding: 10px 12px;
  border-right: 1px solid var(--rule-soft);
  border-bottom: 1px solid var(--rule-soft);
}

.contract-fields > div:nth-child(2n) {
  border-right: none;
}

.contract-fields > div:nth-last-child(-n + 2) {
  border-bottom: none;
}

.contract-null-note {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
  padding: 6px 10px;
  border: 1px dashed var(--rule-color);
  background: var(--surface-ivory);
}

.contract-fields dt {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
  margin-bottom: 4px;
}

.contract-fields dd {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 13px;
  font-weight: 600;
  color: var(--ink-primary);
}

.readout-note {
  font-size: 13.5px;
  line-height: 1.5;
  color: var(--ink-secondary);
}

@media (max-width: 960px) {
  .probe-layout {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .checkpoint-tabs {
    grid-template-columns: repeat(3, 1fr);
  }

  .contract-fields {
    grid-template-columns: 1fr;
  }

  .contract-fields > div {
    border-right: none;
    border-bottom: 1px solid var(--rule-soft);
  }

  .contract-fields > div:nth-last-child(-n + 2) {
    border-bottom: 1px solid var(--rule-soft);
  }

  .contract-fields > div:last-child {
    border-bottom: none;
  }
}
</style>
