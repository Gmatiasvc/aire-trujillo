import { formatLimaTime } from '../data/legends'
import {
  METRIC_KEYS,
  type AdminReadingItem,
  type MetricKey,
  type MetricSpec,
} from '../types/telemetry'

export const MAX_BATCH_POINTS = 288
export const STEP_MINUTES = 5
export const STEP_MS = STEP_MINUTES * 60 * 1000
const LIMA_OFFSET_MS = 5 * 60 * 60 * 1000

/**
 * All 288 5-minute slots across a 24h day in 24h HH:MM format (00:00 .. 23:55).
 */
export const LIMA_FIVE_MIN_SLOTS: readonly string[] = Array.from({ length: MAX_BATCH_POINTS }, (_, idx) => {
  const totalMinutes = idx * STEP_MINUTES
  const hour = String(Math.floor(totalMinutes / 60)).padStart(2, '0')
  const minute = String(totalMinutes % 60).padStart(2, '0')
  return `${hour}:${minute}`
})

export type PreviewChartMode = 'combined' | 'draft' | 'saved'

export interface ParsedMetricInput {
  values: number[]
  error: string | null
}

export interface VariableSeriesPoint {
  timeMs: number
  iso: string
  label: string
  saved: number | null
  draft: number | null
}

export interface BatchTemplatePreset {
  id: 'typical_trujillo' | 'peak_larco' | 'coastal_huanchaco'
  label: string
  summary: string
  metrics: Partial<Record<MetricKey, string>>
}

/**
 * Preset series ordered newest-first (index 0 at start_time, walking back 5 min per comma step).
 */
export const ADMIN_BATCH_TEMPLATES: readonly BatchTemplatePreset[] = [
  {
    id: 'typical_trujillo',
    label: 'Día típico Trujillo (6 pts)',
    summary: 'Serie urbana moderada · 30 min',
    metrics: {
      pm25: '15.4, 14.8, 14.2, 13.9, 13.5, 13.1',
      pm10: '31.2, 30.4, 29.8, 28.9, 28.1, 27.6',
      co: '1.2, 1.15, 1.1, 1.05, 0.98, 0.95',
      no2: '34.0, 32.8, 31.5, 30.2, 29.4, 28.8',
      co2: '442, 438, 434, 430, 426, 422',
      noise_dba: '62.5, 61.8, 60.9, 60.2, 59.6, 58.9',
      temperature_c: '21.6, 21.5, 21.3, 21.1, 20.9, 20.8',
      humidity_pct: '77.0, 77.5, 78.2, 78.8, 79.4, 80.1',
    },
  },
  {
    id: 'peak_larco',
    label: 'Hora punta Av. Larco (8 pts)',
    summary: 'Congestión vehicular vespertina · 40 min',
    metrics: {
      pm25: '38.6, 36.2, 33.9, 31.4, 28.7, 25.5, 22.8, 20.4',
      pm10: '74.5, 70.1, 65.8, 61.0, 56.2, 51.4, 46.9, 43.2',
      co: '4.8, 4.5, 4.1, 3.7, 3.3, 2.9, 2.5, 2.1',
      no2: '94.0, 88.5, 82.0, 76.4, 70.8, 65.2, 58.6, 52.0',
      so2: '24.5, 23.0, 21.4, 19.8, 18.2, 16.9, 15.6, 14.5',
      co2: '685, 668, 642, 618, 590, 562, 535, 512',
      voc_index: '218, 209, 198, 186, 174, 162, 150, 140',
      noise_dba: '78.4, 77.1, 76.0, 74.8, 73.2, 71.5, 69.8, 68.4',
    },
  },
  {
    id: 'coastal_huanchaco',
    label: 'Brisa costera Huanchaco (6 pts)',
    summary: 'Dispersión marina y humedad alta · 30 min',
    metrics: {
      pm25: '9.2, 9.6, 10.1, 10.5, 11.0, 11.4',
      pm10: '24.0, 24.8, 25.6, 26.2, 27.0, 27.8',
      h2s: '3.2, 3.4, 3.5, 3.8, 4.0, 4.1',
      noise_dba: '52.4, 53.0, 53.6, 54.1, 54.8, 55.2',
      temperature_c: '19.8, 19.9, 20.1, 20.2, 20.4, 20.5',
      humidity_pct: '86.4, 85.9, 85.2, 84.8, 84.1, 83.6',
      pressure_hpa: '1014.2, 1014.0, 1013.9, 1013.8, 1013.6, 1013.5',
    },
  },
]

/**
 * Converts a UTC instant into America/Lima (fixed UTC-05:00) YYYY-MM-DD and floored 5-min HH:MM slot.
 */
export function getCurrentLimaDateAndSlot(now = new Date()): { dateYmd: string; hhmm: string } {
  const limaClock = new Date(now.getTime() - LIMA_OFFSET_MS)
  const year = limaClock.getUTCFullYear()
  const month = String(limaClock.getUTCMonth() + 1).padStart(2, '0')
  const day = String(limaClock.getUTCDate()).padStart(2, '0')
  const hour = String(limaClock.getUTCHours()).padStart(2, '0')
  const minuteFloored = Math.floor(limaClock.getUTCMinutes() / STEP_MINUTES) * STEP_MINUTES
  const minute = String(minuteFloored).padStart(2, '0')
  return {
    dateYmd: `${year}-${month}-${day}`,
    hhmm: `${hour}:${minute}`,
  }
}

/**
 * Formats an exact RFC3339 timestamp with -05:00 offset aligned to a 5-minute multiple.
 */
export function buildRfc3339LimaStartTime(dateYmd: string, hhmm: string): string | null {
  const dateMatch = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateYmd.trim())
  const timeMatch = /^(\d{2}):(\d{2})$/.exec(hhmm.trim())
  if (!dateMatch || !timeMatch) {
    return null
  }
  const year = Number(dateMatch[1])
  const month = Number(dateMatch[2])
  const day = Number(dateMatch[3])
  const hour = Number(timeMatch[1])
  const minute = Number(timeMatch[2])

  if (month < 1 || month > 12 || day < 1 || day > 31 || hour < 0 || hour > 23) {
    return null
  }
  if (minute < 0 || minute > 59 || minute % STEP_MINUTES !== 0) {
    return null
  }

  // Validate real calendar day (reject e.g. 2026-02-30)
  const check = new Date(Date.UTC(year, month - 1, day))
  if (
    check.getUTCFullYear() !== year ||
    check.getUTCMonth() + 1 !== month ||
    check.getUTCDate() !== day
  ) {
    return null
  }

  return `${dateMatch[1]}-${dateMatch[2]}-${dateMatch[3]}T${timeMatch[1]}:${timeMatch[2]}:00-05:00`
}

/**
 * Parses comma-separated metric values adhering to backend batch limits (1..288 finite values inside spec range).
 */
export function parseCommaMetricInput(raw: string, spec: MetricSpec): ParsedMetricInput {
  const trimmed = raw.trim()
  if (trimmed.length === 0) {
    return { values: [], error: null }
  }

  if (trimmed.endsWith(',')) {
    return {
      values: [],
      error: 'Quita la coma final o añade el valor siguiente.',
    }
  }

  const tokens = trimmed.split(',')
  if (tokens.length > MAX_BATCH_POINTS) {
    return {
      values: [],
      error: `Máximo ${MAX_BATCH_POINTS} puntos por lote (24 h cada 5 min).`,
    }
  }

  const values: number[] = []
  for (let i = 0; i < tokens.length; i += 1) {
    const token = tokens[i]?.trim() ?? ''
    if (token.length === 0) {
      return {
        values: [],
        error: `Falta valor numérico en la posición ${i + 1}.`,
      }
    }
    const parsed = Number(token)
    if (!Number.isFinite(parsed)) {
      return {
        values: [],
        error: `Valor inválido en la posición ${i + 1}: «${token}».`,
      }
    }
    if (parsed < spec.min || parsed > spec.max) {
      return {
        values: [],
        error: `Posición ${i + 1} (${parsed}) fuera de [${spec.min}, ${spec.max} ${spec.unit}].`,
      }
    }
    values.push(parsed)
  }

  return { values, error: null }
}

/**
 * Merges saved backend readings and newest-first draft array chronologically (oldest -> newest) for charting.
 */
export function buildVariableSeriesPoints(
  metric: MetricKey,
  savedReadings: readonly AdminReadingItem[],
  draftNewestFirst: readonly number[],
  startTimeIso: string | null,
  mode: PreviewChartMode,
): VariableSeriesPoint[] {
  const byTime = new Map<number, { iso: string; saved: number | null; draft: number | null }>()

  if (mode === 'combined' || mode === 'saved') {
    for (const row of savedReadings) {
      const value = row[metric]
      if (value === undefined || !Number.isFinite(value)) continue
      const ms = Date.parse(row.recorded_at)
      if (Number.isNaN(ms)) continue
      const entry = byTime.get(ms) ?? { iso: row.recorded_at, saved: null, draft: null }
      entry.saved = value
      byTime.set(ms, entry)
    }
  }

  if ((mode === 'combined' || mode === 'draft') && draftNewestFirst.length > 0 && startTimeIso) {
    const startMs = Date.parse(startTimeIso)
    if (!Number.isNaN(startMs)) {
      for (let i = 0; i < draftNewestFirst.length; i += 1) {
        const val = draftNewestFirst[i]
        if (val === undefined || !Number.isFinite(val)) continue
        const ms = startMs - i * STEP_MS
        const iso = new Date(ms).toISOString()
        const entry = byTime.get(ms) ?? { iso, saved: null, draft: null }
        entry.draft = val
        byTime.set(ms, entry)
      }
    }
  }

  return [...byTime.entries()]
    .sort((a, b) => a[0] - b[0])
    .map(([timeMs, item]) => ({
      timeMs,
      iso: item.iso,
      label: formatLimaTime(item.iso),
      saved: item.saved,
      draft: item.draft,
    }))
}

export function createEmptyDraftMap(): Record<MetricKey, string> {
  const record = {} as Record<MetricKey, string>
  for (const key of METRIC_KEYS) {
    record[key] = ''
  }
  return record
}
