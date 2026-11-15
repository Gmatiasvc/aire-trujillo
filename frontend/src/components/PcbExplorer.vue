<script setup lang="ts">
import { computed, ref } from 'vue'
import { PCB_BLOCKS, type PcbBlock, type PcbGroup } from '../data/pcbBlocks'
import { METRIC_SPECS } from '../types/telemetry'

const activeGroup = ref<PcbGroup>('all')
const selectedId = ref<string>('esp32-s3')
const hoveredId = ref<string | null>(null)

const groupFilters: Array<{ id: PcbGroup; label: string }> = [
  { id: 'all', label: 'Todos (11)' },
  { id: 'sensors', label: 'Sensores (7)' },
  { id: 'compute', label: 'MCU & LTE-M (2)' },
  { id: 'power', label: 'Solar & IP66 (2)' },
]

const filteredIds = computed<Set<string>>(() => {
  if (activeGroup.value === 'all') {
    return new Set(PCB_BLOCKS.map((b) => b.id))
  }
  return new Set(PCB_BLOCKS.filter((b) => b.group === activeGroup.value).map((b) => b.id))
})

const inspectedId = computed<string>(() => hoveredId.value ?? selectedId.value)

const inspectedBlock = computed<PcbBlock>(() => {
  return PCB_BLOCKS.find((b) => b.id === inspectedId.value) ?? PCB_BLOCKS[0]!
})

function selectGroup(next: PcbGroup): void {
  activeGroup.value = next
  if (!filteredIds.value.has(selectedId.value)) {
    const firstMatch = PCB_BLOCKS.find((b) => b.group === next)
    if (firstMatch) {
      selectedId.value = firstMatch.id
    }
  }
}

function selectBlock(id: string): void {
  selectedId.value = id
}
</script>

<template>
  <div class="pcb-shell">
    <div class="pcb-toolbar">
      <div class="pcb-filter-row" role="group" aria-label="Filtrar subsistema de hardware">
        <button
          v-for="item in groupFilters"
          :key="item.id"
          type="button"
          class="pcb-filter-btn"
          :class="{ 'is-active': activeGroup === item.id }"
          @click="selectGroup(item.id)"
        >
          <span>{{ item.label }}</span>
        </button>
      </div>
      <span class="pcb-meta tabular">NODO SOLAR IP66 · CONFORMAL COATING · 12 VARIABLES</span>
    </div>

    <div class="pcb-grid">
      <div class="pcb-board-frame">
        <div class="pcb-canvas-scroll">
          <svg
            viewBox="0 0 656 396"
            class="pcb-svg"
            role="group"
            aria-label="Diagrama interactivo de la placa PCB de Aire Trujillo"
          >
          <!-- PCB substrate and silkscreen perimeter -->
          <rect
            x="14"
            y="14"
            width="628"
            height="368"
            rx="10"
            class="pcb-substrate"
          />
          <rect
            x="22"
            y="22"
            width="612"
            height="352"
            rx="6"
            fill="none"
            class="pcb-silk-frame"
          />

          <!-- Mounting holes -->
          <g class="pcb-mount-holes">
            <circle cx="30" cy="30" r="5" />
            <circle cx="626" cy="30" r="5" />
            <circle cx="30" cy="366" r="5" />
            <circle cx="626" cy="366" r="5" />
          </g>

          <!-- Copper routing traces linking subsystems to U1 ESP32-S3 -->
          <g class="pcb-traces">
            <path
              v-for="block in PCB_BLOCKS"
              :key="`trace-${block.id}`"
              :d="`M ${block.pinNodes[0]?.[0] ?? 330} ${block.pinNodes[0]?.[1] ?? 176} L ${block.pinNodes[1]?.[0] ?? 330} ${block.pinNodes[1]?.[1] ?? 176}`"
              class="trace-line"
              :class="{
                'is-lit': inspectedId === block.id,
                'is-dim': !filteredIds.has(block.id),
              }"
            />
            <!-- Internal routing rails -->
            <path d="M 202 294 L 224 294 L 224 196 L 245 196" class="trace-bus" />
            <path d="M 192 79 L 224 79 L 224 146 L 245 146" class="trace-bus" />
            <path d="M 415 90 L 455 90" class="trace-bus" />
            <path d="M 415 212 L 434 212 L 434 335 L 455 335" class="trace-bus" />
          </g>

          <!-- Silkscreen board tag -->
          <text x="34" y="362" class="pcb-silk-tag">AIRE-TRUJILLO REV1 · ESP32-S3 + LTE-M</text>

          <!-- Interactive BOM Component Pads -->
          <g
            v-for="block in PCB_BLOCKS"
            :key="block.id"
            role="button"
            tabindex="0"
            :aria-label="`${block.designator} ${block.name}: ${block.subsystem}`"
            :aria-pressed="selectedId === block.id"
            class="pcb-block"
            :class="{
              'is-selected': selectedId === block.id,
              'is-hovered': hoveredId === block.id,
              'is-dimmed': !filteredIds.has(block.id),
            }"
            @mouseenter="hoveredId = block.id"
            @mouseleave="hoveredId = null"
            @focus="hoveredId = block.id"
            @blur="hoveredId = null"
            @click="selectBlock(block.id)"
            @keydown.enter.prevent="selectBlock(block.id)"
            @keydown.space.prevent="selectBlock(block.id)"
          >
            <rect
              :x="block.x"
              :y="block.y"
              :width="block.w"
              :height="block.h"
              rx="4"
              class="block-body"
            />
            <text :x="block.x + 10" :y="block.y + 18" class="block-designator">
              {{ block.designator }} · {{ block.quantity }}
            </text>
            <text :x="block.x + 10" :y="block.y + 36" class="block-title">
              {{ block.name }}
            </text>
            <text
              v-if="block.h >= 60"
              :x="block.x + 10"
              :y="block.y + 54"
              class="block-sub"
            >
              {{ block.subsystem }}
            </text>
            <text
              v-if="block.h >= 72"
              :x="block.x + 10"
              :y="block.y + block.h - 12"
              class="block-bus"
            >
              {{ block.bus }}
            </text>
          </g>
          </svg>
        </div>

        <p class="pcb-scroll-hint">Desliza horizontalmente la placa para leer toda la serigrafía.</p>

        <div class="pcb-quick-chips" aria-label="Selección rápida de componentes">
          <button
            v-for="block in PCB_BLOCKS"
            :key="`chip-${block.id}`"
            type="button"
            class="quick-chip"
            :class="{
              'is-active': inspectedId === block.id,
              'is-dimmed': !filteredIds.has(block.id),
            }"
            @mouseenter="hoveredId = block.id"
            @mouseleave="hoveredId = null"
            @click="selectBlock(block.id)"
          >
            {{ block.designator }}
          </button>
        </div>
      </div>

      <aside class="pcb-dossier" aria-live="polite">
        <div class="dossier-head">
          <span>INSPECTOR DE BLOQUE BOM</span>
          <span class="tabular">{{ inspectedBlock.designator }}</span>
        </div>

        <Transition name="slide-swap" mode="out-in">
          <div :key="inspectedBlock.id" class="dossier-body">
            <div class="dossier-title-group">
              <h3>{{ inspectedBlock.name }}</h3>
              <span class="subsystem-badge">{{ inspectedBlock.subsystem }}</span>
            </div>

            <p class="dossier-summary">{{ inspectedBlock.functionSummary }}</p>
            <p class="dossier-detail">{{ inspectedBlock.engineeringDetail }}</p>

            <dl class="dossier-specs">
              <div>
                <dt>Cantidad</dt>
                <dd class="tabular">{{ inspectedBlock.quantity }}</dd>
              </div>
              <div>
                <dt>Bus / Señal</dt>
                <dd>{{ inspectedBlock.bus }}</dd>
              </div>
              <div>
                <dt>Conector</dt>
                <dd>{{ inspectedBlock.connector }}</dd>
              </div>
            </dl>

            <div class="dossier-metrics">
              <div class="metrics-heading">Variables canónicas medidas</div>
              <div v-if="inspectedBlock.metrics.length > 0" class="metric-pills">
                <div
                  v-for="key in inspectedBlock.metrics"
                  :key="key"
                  class="metric-chip"
                >
                  <span class="m-key">{{ key }}</span>
                  <span class="m-label">{{ METRIC_SPECS[key].label }}</span>
                  <span class="m-unit tabular">{{ METRIC_SPECS[key].unit }}</span>
                </div>
              </div>
              <p v-else class="no-direct-metrics">
                Subsistema de soporte (procesamiento, enlace LTE-M, potencia solar o respaldo MicroSD para las 12 variables).
              </p>
            </div>
          </div>
        </Transition>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.pcb-shell {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  min-width: 0;
  max-width: 100%;
}

.pcb-toolbar {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 12px 18px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border-bottom: 1.5px solid var(--rule-color);
}

.pcb-filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pcb-filter-btn {
  border: 1px solid var(--rule-color);
  background: transparent;
  color: var(--ink-muted-on-carbon);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  padding: 7px 14px;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.2s var(--ease-expo),
    color 0.2s var(--ease-expo),
    border-color 0.2s var(--ease-expo);
}

.pcb-filter-btn > span {
  display: block;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.pcb-filter-btn:hover {
  color: var(--ink-on-carbon);
  border-color: var(--accent-amber-fill);
}

.pcb-filter-btn.is-active {
  background: var(--accent-amber-fill);
  border-color: var(--accent-amber-fill);
  color: #181a1c;
}

.pcb-meta {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-muted-on-carbon);
  letter-spacing: 0.05em;
}

.pcb-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.35fr) minmax(0, 0.85fr);
  gap: 0;
}

.pcb-grid > * {
  min-width: 0;
}

.pcb-board-frame {
  padding: 18px;
  border-right: 1.5px solid var(--rule-color);
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.pcb-canvas-scroll {
  width: 100%;
  overflow-x: auto;
  border: 1px solid var(--rule-soft);
}

.pcb-svg {
  width: 100%;
  height: auto;
  display: block;
}

.pcb-scroll-hint {
  display: none;
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
}

.pcb-substrate {
  fill: var(--pcb-board);
  stroke: var(--pcb-board-edge);
  stroke-width: 2.5;
}

.pcb-silk-frame {
  stroke: var(--pcb-silk);
  stroke-opacity: 0.32;
  stroke-width: 1;
  stroke-dasharray: 5 4;
}

.pcb-mount-holes circle {
  fill: var(--bed-bg);
  stroke: var(--pcb-copper);
  stroke-width: 2;
}

.trace-bus {
  fill: none;
  stroke: var(--pcb-copper);
  stroke-opacity: 0.28;
  stroke-width: 1.5;
}

.trace-line {
  fill: none;
  stroke: var(--pcb-copper);
  stroke-opacity: 0.35;
  stroke-width: 2;
  transition:
    stroke-opacity 0.2s var(--ease-expo),
    stroke-width 0.2s var(--ease-expo),
    stroke 0.2s var(--ease-expo);
}

.trace-line.is-lit {
  stroke: #fff3d1;
  stroke-opacity: 1;
  stroke-width: 3;
  stroke-dasharray: 9 6;
  animation: pcb-trace-flow 0.9s linear infinite;
}

@keyframes pcb-trace-flow {
  from {
    stroke-dashoffset: 30;
  }
  to {
    stroke-dashoffset: 0;
  }
}

.trace-line.is-dim {
  stroke-opacity: 0.1;
}

.pcb-silk-tag {
  fill: var(--pcb-silk);
  fill-opacity: 0.72;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.06em;
}

.pcb-block {
  cursor: pointer;
  transition: opacity 0.22s var(--ease-expo);
}

.pcb-block.is-dimmed {
  opacity: 0.32;
}

.block-body {
  fill: #151b18;
  stroke: var(--pcb-copper);
  stroke-width: 1.5;
  transition:
    fill 0.2s var(--ease-expo),
    stroke 0.2s var(--ease-expo),
    stroke-width 0.2s var(--ease-expo);
}

.pcb-block:hover .block-body,
.pcb-block.is-hovered .block-body {
  fill: #27322c;
  stroke: #fff3d1;
  stroke-width: 2.2;
}

.pcb-block.is-selected .block-body {
  fill: var(--accent-vermilion);
  stroke: #fff9f2;
  stroke-width: 2.4;
}

.block-designator {
  fill: var(--pcb-copper);
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 600;
}

.pcb-block.is-selected .block-designator {
  fill: #ffe8d6;
}

.block-title {
  fill: #f7f1e3;
  font-family: var(--font-body);
  font-size: 11.5px;
  font-weight: 700;
}

.block-sub {
  fill: #d8ccb8;
  font-family: var(--font-body);
  font-size: 9.8px;
}

.pcb-block.is-selected .block-sub,
.pcb-block.is-selected .block-bus {
  fill: #fff1e6;
}

.block-bus {
  fill: #b9c5bd;
  font-family: var(--font-mono);
  font-size: 9px;
}

.pcb-quick-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.quick-chip {
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  padding: 5px 9px;
  cursor: pointer;
  transition:
    background-color 0.18s var(--ease-expo),
    color 0.18s var(--ease-expo),
    opacity 0.18s var(--ease-expo);
}

.quick-chip.is-dimmed {
  opacity: 0.45;
}

.quick-chip:hover {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.quick-chip.is-active {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.pcb-dossier {
  display: flex;
  flex-direction: column;
  background: var(--surface-ivory);
}

.dossier-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 11px 18px;
  background: var(--surface-ivory-2);
  border-bottom: 1px solid var(--rule-color);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  color: var(--ink-secondary);
}

.dossier-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.dossier-title-group {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
}

.dossier-title-group h3 {
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 700;
  line-height: 1.05;
  text-transform: uppercase;
}

.subsystem-badge {
  display: inline-block;
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  padding: 4px 10px;
  background: var(--accent-bottle);
  color: var(--accent-bottle-ink);
}

.dossier-summary {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.45;
  color: var(--ink-primary);
}

.dossier-detail {
  font-size: 14px;
  line-height: 1.55;
  color: var(--ink-secondary);
}

.dossier-specs {
  margin: 0;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.dossier-specs > div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 9px 12px;
  border-bottom: 1px solid var(--rule-soft);
  font-size: 12.5px;
}

.dossier-specs > div:last-child {
  border-bottom: none;
}

.dossier-specs dt {
  font-family: var(--font-mono);
  color: var(--ink-secondary);
}

.dossier-specs dd {
  margin: 0;
  font-weight: 600;
  text-align: right;
  color: var(--ink-primary);
}

.metrics-heading {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--ink-secondary);
  margin-bottom: 8px;
}

.metric-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.metric-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
  font-size: 12px;
}

.m-key {
  font-family: var(--font-mono);
  font-weight: 600;
  color: var(--accent-vermilion);
}

:root[data-theme='dark'] .m-key {
  color: var(--accent-amber);
}

.m-label {
  font-weight: 600;
}

.m-unit {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
}

.no-direct-metrics {
  font-size: 13px;
  color: var(--ink-secondary);
  line-height: 1.45;
}

@media (max-width: 1024px) {
  .pcb-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .pcb-board-frame {
    border-right: none;
    border-bottom: 1.5px solid var(--rule-color);
    padding: 14px;
  }
}

@media (max-width: 640px) {
  .pcb-toolbar {
    padding: 12px 14px;
  }

  .pcb-svg {
    min-width: 560px;
  }

  .pcb-scroll-hint {
    display: block;
  }

  .dossier-body {
    padding: 16px 14px;
  }
}
</style>
