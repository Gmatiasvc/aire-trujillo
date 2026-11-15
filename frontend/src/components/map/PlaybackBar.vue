<script setup lang="ts">
import { computed } from 'vue'
import UiIcon from '../UiIcon.vue'
import { formatLimaDateTime, formatLimaTime } from '../../data/legends'
import type { PlaybackSpeed } from '../../stores/map'

const props = defineProps<{
  timestamps: string[]
  playbackIndex: number
  isPlaying: boolean
  playbackSpeed: PlaybackSpeed
  isLive: boolean
  displayTimestamp: string | null
  isLoadingMap: boolean
}>()

const emit = defineEmits<{
  'toggle-play': []
  'set-speed': [speed: PlaybackSpeed]
  'scrub-index': [index: number]
  'go-live': []
}>()

const speeds: PlaybackSpeed[] = [1, 2, 4]

const maxIndex = computed<number>(() => Math.max(0, props.timestamps.length - 1))

const progressPct = computed<number>(() => {
  if (maxIndex.value <= 0) return 100
  return Math.round((props.playbackIndex / maxIndex.value) * 1000) / 10
})

const startTimeLabel = computed<string>(() => {
  return formatLimaTime(props.timestamps[0] ?? null)
})

const midTimeLabel = computed<string>(() => {
  if (props.timestamps.length === 0) return '--:--'
  const mid = props.timestamps[Math.floor(props.timestamps.length / 2)]
  return formatLimaTime(mid ?? null)
})

const endTimeLabel = computed<string>(() => {
  if (props.timestamps.length === 0) return '--:--'
  return formatLimaTime(props.timestamps[props.timestamps.length - 1] ?? null)
})

function handleScrub(event: Event): void {
  const input = event.target as HTMLInputElement
  emit('scrub-index', Number(input.value))
}
</script>

<template>
  <section class="playback-bar" aria-label="Reproducción histórica de 24 horas">
    <div class="playback-controls">
      <button
        type="button"
        class="play-btn"
        :disabled="timestamps.length <= 1"
        :aria-label="isPlaying ? 'Pausar reproducción' : 'Reproducir últimas 24 horas'"
        @click="emit('toggle-play')"
      >
        <UiIcon :name="isPlaying ? 'pause' : 'play'" :size="15" />
        <span class="play-btn-text">{{ isPlaying ? 'Pausa' : '24h' }}</span>
      </button>

      <div class="speed-group" role="group" aria-label="Velocidad de reproducción">
        <button
          v-for="spd in speeds"
          :key="spd"
          type="button"
          class="speed-btn tabular"
          :class="{ 'is-active': playbackSpeed === spd }"
          :aria-pressed="playbackSpeed === spd"
          @click="emit('set-speed', spd)"
        >
          {{ spd }}x
        </button>
      </div>
    </div>

    <div class="scrubber-column">
      <div class="scrubber-top">
        <span class="active-ts tabular">
          {{ formatLimaDateTime(displayTimestamp) }}
        </span>
        <span v-if="isLoadingMap" class="loading-tag">Actualizando…</span>
        <span v-else class="step-tag tabular">Paso 5 min · America/Lima</span>
      </div>

      <div class="slider-track-wrap">
        <input
          type="range"
          class="timeline-range"
          :min="0"
          :max="maxIndex"
          :step="1"
          :value="playbackIndex"
          :disabled="timestamps.length <= 1"
          aria-label="Línea de tiempo de 24 horas en intervalos de 5 minutos"
          :style="{ '--progress': `${progressPct}%` }"
          @input="handleScrub"
        />
        <div class="tick-row tabular" aria-hidden="true">
          <span>{{ startTimeLabel }}</span>
          <span>{{ midTimeLabel }}</span>
          <span>{{ endTimeLabel }}</span>
        </div>
      </div>
    </div>

    <button
      type="button"
      class="raked-btn live-btn"
      :class="isLive ? 'is-synced' : 'raked-btn-primary'"
      :aria-pressed="isLive"
      @click="emit('go-live')"
    >
      <span>
        <i class="live-dot" :class="{ 'is-live': isLive }" aria-hidden="true"></i>
        En vivo
      </span>
    </button>
  </section>
</template>

<style scoped>
.playback-bar {
  display: grid;
  grid-template-columns: auto minmax(180px, 1fr) auto;
  align-items: center;
  gap: 14px;
  padding: 10px 14px;
  background: var(--surface-ivory);
  color: var(--ink-primary);
  border: 1.5px solid var(--rule-color);
}

.playback-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.play-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 36px;
  padding: 0 12px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition:
    background-color 0.18s var(--ease-expo),
    border-color 0.18s var(--ease-expo);
}

.play-btn:hover:not(:disabled) {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.play-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.speed-group {
  display: inline-flex;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.speed-btn {
  height: 33px;
  padding: 0 9px;
  border: none;
  border-right: 1px solid var(--rule-soft);
  background: transparent;
  color: var(--ink-secondary);
  font-family: var(--font-mono);
  font-size: 11.5px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background-color 0.16s var(--ease-expo),
    color 0.16s var(--ease-expo);
}

.speed-btn:last-child {
  border-right: none;
}

.speed-btn:hover {
  color: var(--ink-primary);
}

.speed-btn.is-active {
  background: var(--surface-carbon);
  color: var(--accent-amber-fill);
}

.scrubber-column {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.scrubber-top {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
}

.active-ts {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  color: var(--ink-primary);
}

.loading-tag,
.step-tag {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
}

.slider-track-wrap {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.timeline-range {
  appearance: none;
  width: 100%;
  height: 8px;
  margin: 2px 0;
  background: linear-gradient(
    90deg,
    var(--accent-vermilion) 0%,
    var(--accent-vermilion) var(--progress, 100%),
    var(--surface-ivory-2) var(--progress, 100%),
    var(--surface-ivory-2) 100%
  );
  border: 1px solid var(--rule-color);
  cursor: pointer;
}

.timeline-range::-webkit-slider-thumb {
  appearance: none;
  width: 14px;
  height: 18px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  cursor: ew-resize;
}

.timeline-range::-moz-range-thumb {
  width: 14px;
  height: 18px;
  border-radius: 0;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  cursor: ew-resize;
}

.tick-row {
  display: flex;
  justify-content: space-between;
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--ink-secondary);
}

.live-btn {
  padding: 9px 16px;
  font-size: 11.5px;
}

.live-btn.is-synced {
  background: var(--accent-bottle);
  border-color: var(--accent-bottle);
  color: var(--accent-bottle-ink);
}

.live-dot {
  width: 8px;
  height: 8px;
  background: currentColor;
  opacity: 0.45;
}

.live-dot.is-live {
  background: #4ade80;
  opacity: 1;
}

@media (max-width: 780px) {
  .playback-bar {
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 8px 10px;
    padding: 8px 10px;
  }

  .scrubber-column {
    grid-column: 1 / -1;
    grid-row: 1;
  }

  .playback-controls {
    grid-column: 1;
    grid-row: 2;
  }

  .live-btn {
    grid-column: 2;
    grid-row: 2;
    padding: 8px 12px;
  }

  .step-tag {
    display: none;
  }
}
</style>
