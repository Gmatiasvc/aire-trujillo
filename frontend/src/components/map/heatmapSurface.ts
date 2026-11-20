import type { HeatmapCell, HeatmapResponse } from '../../types/telemetry'

export interface HeatmapGridIndex {
  minLat: number
  maxLat: number
  minLon: number
  maxLon: number
  stepLat: number
  stepLon: number
  rows: number
  cols: number
  cells: Array<HeatmapCell | null>
  fallbackCells: HeatmapCell[] | null
}

export interface RenderedHeatmapSurface {
  image: ImageData | null
  coordinates: [[number, number], [number, number], [number, number], [number, number]] | null
  gridIndex: HeatmapGridIndex
}

const METERS_PER_DEG_LAT = 111320
const MAX_INDEX_CELLS = 500_000
const MAX_RASTER_DIM = 480
const RASTER_PAD_CELLS = 2

export function createEmptyHeatmapImage(): ImageData {
  return new ImageData(new Uint8ClampedArray([156, 163, 175, 0]), 1, 1)
}

function clamp01(value: number): number {
  if (value <= 0) return 0
  if (value >= 1) return 1
  return value
}

function clampByte(value: number): number {
  if (value <= 0) return 0
  if (value >= 255) return 255
  return value
}

function parseHexRgbUnit(hex: string): [number, number, number] {
  const clean = hex.trim().replace(/^#/, '')
  if (clean.length !== 6) {
    return [156 / 255, 163 / 255, 175 / 255]
  }
  const r = Number.parseInt(clean.slice(0, 2), 16)
  const g = Number.parseInt(clean.slice(2, 4), 16)
  const b = Number.parseInt(clean.slice(4, 6), 16)
  if (Number.isNaN(r) || Number.isNaN(g) || Number.isNaN(b)) {
    return [156 / 255, 163 / 255, 175 / 255]
  }
  return [r / 255, g / 255, b / 255]
}

function inferGridStep(values: number[], fallback: number): number {
  const sorted = [...values].sort((a, b) => a - b)
  const diffs: number[] = []
  for (let i = 1; i < sorted.length; i++) {
    const d = sorted[i]! - sorted[i - 1]!
    if (d > 1e-6) {
      diffs.push(d)
    }
  }
  if (diffs.length === 0) return fallback

  let minDiff = diffs[0]!
  for (let i = 1; i < diffs.length; i++) {
    if (diffs[i]! < minDiff) {
      minDiff = diffs[i]!
    }
  }
  const nearMin = diffs.filter((d) => d <= minDiff * 1.5).sort((a, b) => a - b)
  if (nearMin.length === 0) return fallback
  return nearMin[Math.floor(nearMin.length / 2)] ?? fallback
}

function buildGridIndex(
  cells: HeatmapCell[],
  minLat: number,
  maxLat: number,
  minLon: number,
  maxLon: number,
  stepLat: number,
  stepLon: number,
): HeatmapGridIndex {
  const rows = Math.max(1, Math.round((maxLat - minLat) / stepLat) + 1)
  const cols = Math.max(1, Math.round((maxLon - minLon) / stepLon) + 1)

  if (rows * cols > MAX_INDEX_CELLS) {
    return {
      minLat,
      maxLat,
      minLon,
      maxLon,
      stepLat,
      stepLon,
      rows,
      cols,
      cells: [],
      fallbackCells: cells,
    }
  }

  const table = new Array<HeatmapCell | null>(rows * cols).fill(null)
  for (const cell of cells) {
    const r = Math.round((cell.lat - minLat) / stepLat)
    const c = Math.round((cell.lon - minLon) / stepLon)
    if (r >= 0 && r < rows && c >= 0 && c < cols) {
      table[r * cols + c] = cell
    }
  }

  return {
    minLat,
    maxLat,
    minLon,
    maxLon,
    stepLat,
    stepLon,
    rows,
    cols,
    cells: table,
    fallbackCells: null,
  }
}

export function lookupHeatmapCell(
  index: HeatmapGridIndex | null,
  lat: number,
  lon: number,
): HeatmapCell | null {
  if (!index) return null

  const latSlack = index.stepLat * 0.6
  const lonSlack = index.stepLon * 0.6
  if (
    lat < index.minLat - latSlack ||
    lat > index.maxLat + latSlack ||
    lon < index.minLon - lonSlack ||
    lon > index.maxLon + lonSlack
  ) {
    return null
  }

  if (index.fallbackCells) {
    let best: HeatmapCell | null = null
    let bestDistSq = Infinity
    for (const cell of index.fallbackCells) {
      const dLat = Math.abs(cell.lat - lat)
      const dLon = Math.abs(cell.lon - lon)
      if (dLat <= latSlack && dLon <= lonSlack) {
        const nLat = dLat / index.stepLat
        const nLon = dLon / index.stepLon
        const dSq = nLat * nLat + nLon * nLon
        if (dSq < bestDistSq) {
          bestDistSq = dSq
          best = cell
        }
      }
    }
    return best
  }

  const r = Math.round((lat - index.minLat) / index.stepLat)
  const c = Math.round((lon - index.minLon) / index.stepLon)
  if (r < 0 || r >= index.rows || c < 0 || c >= index.cols) {
    return null
  }

  const hit = index.cells[r * index.cols + c] ?? null
  if (!hit) return null
  if (Math.abs(hit.lat - lat) > latSlack || Math.abs(hit.lon - lon) > lonSlack) {
    return null
  }
  return hit
}

function buildGaussianKernel(sigma: number): { weights: Float32Array; radius: number } {
  const radius = Math.max(1, Math.ceil(sigma * 2.2))
  const weights = new Float32Array(radius * 2 + 1)
  const twoSigmaSq = 2 * sigma * sigma
  let sum = 0

  for (let i = -radius; i <= radius; i++) {
    const w = Math.exp(-(i * i) / twoSigmaSq)
    weights[i + radius] = w
    sum += w
  }
  for (let i = 0; i < weights.length; i++) {
    weights[i] = weights[i]! / sum
  }
  return { weights, radius }
}

function upsampleBilinearPremul(
  src: Float32Array,
  srcW: number,
  srcH: number,
  up: number,
): Float32Array {
  if (up <= 1) {
    return src
  }

  const outW = srcW * up
  const outH = srcH * up
  const out = new Float32Array(outW * outH * 4)

  const x0Arr = new Int32Array(outW)
  const x1Arr = new Int32Array(outW)
  const txArr = new Float32Array(outW)
  for (let x = 0; x < outW; x++) {
    const u = Math.min(srcW - 1, Math.max(0, (x + 0.5) / up - 0.5))
    const x0 = Math.floor(u)
    x0Arr[x] = x0
    x1Arr[x] = Math.min(srcW - 1, x0 + 1)
    txArr[x] = u - x0
  }

  for (let y = 0; y < outH; y++) {
    const v = Math.min(srcH - 1, Math.max(0, (y + 0.5) / up - 0.5))
    const y0 = Math.floor(v)
    const y1 = Math.min(srcH - 1, y0 + 1)
    const ty = v - y0
    const row0Base = y0 * srcW
    const row1Base = y1 * srcW
    const dstRowBase = y * outW

    for (let x = 0; x < outW; x++) {
      const x0 = x0Arr[x]!
      const x1 = x1Arr[x]!
      const tx = txArr[x]!
      const itx = 1 - tx
      const ity = 1 - ty

      const w00 = itx * ity
      const w10 = tx * ity
      const w01 = itx * ty
      const w11 = tx * ty

      const i00 = (row0Base + x0) * 4
      const i10 = (row0Base + x1) * 4
      const i01 = (row1Base + x0) * 4
      const i11 = (row1Base + x1) * 4
      const dstIdx = (dstRowBase + x) * 4

      out[dstIdx] = src[i00]! * w00 + src[i10]! * w10 + src[i01]! * w01 + src[i11]! * w11
      out[dstIdx + 1] =
        src[i00 + 1]! * w00 + src[i10 + 1]! * w10 + src[i01 + 1]! * w01 + src[i11 + 1]! * w11
      out[dstIdx + 2] =
        src[i00 + 2]! * w00 + src[i10 + 2]! * w10 + src[i01 + 2]! * w01 + src[i11 + 2]! * w11
      out[dstIdx + 3] =
        src[i00 + 3]! * w00 + src[i10 + 3]! * w10 + src[i01 + 3]! * w01 + src[i11 + 3]! * w11
    }
  }

  return out
}

function blurPremulSeparable(
  buf: Float32Array,
  width: number,
  height: number,
  sigma: number,
): Float32Array {
  const { weights, radius } = buildGaussianKernel(sigma)
  const tmp = new Float32Array(width * height * 4)
  const out = new Float32Array(width * height * 4)

  // Horizontal pass
  for (let y = 0; y < height; y++) {
    const rowBase = y * width
    for (let x = 0; x < width; x++) {
      let r = 0
      let g = 0
      let b = 0
      let a = 0
      for (let k = -radius; k <= radius; k++) {
        const sx = Math.min(width - 1, Math.max(0, x + k))
        const srcIdx = (rowBase + sx) * 4
        const w = weights[k + radius]!
        r += buf[srcIdx]! * w
        g += buf[srcIdx + 1]! * w
        b += buf[srcIdx + 2]! * w
        a += buf[srcIdx + 3]! * w
      }
      const dstIdx = (rowBase + x) * 4
      tmp[dstIdx] = r
      tmp[dstIdx + 1] = g
      tmp[dstIdx + 2] = b
      tmp[dstIdx + 3] = a
    }
  }

  // Vertical pass
  for (let y = 0; y < height; y++) {
    const rowBase = y * width
    for (let x = 0; x < width; x++) {
      let r = 0
      let g = 0
      let b = 0
      let a = 0
      for (let k = -radius; k <= radius; k++) {
        const sy = Math.min(height - 1, Math.max(0, y + k))
        const srcIdx = (sy * width + x) * 4
        const w = weights[k + radius]!
        r += tmp[srcIdx]! * w
        g += tmp[srcIdx + 1]! * w
        b += tmp[srcIdx + 2]! * w
        a += tmp[srcIdx + 3]! * w
      }
      const dstIdx = (rowBase + x) * 4
      out[dstIdx] = r
      out[dstIdx + 1] = g
      out[dstIdx + 2] = b
      out[dstIdx + 3] = a
    }
  }

  return out
}

export function buildHeatmapSurface(
  response: HeatmapResponse | null,
): RenderedHeatmapSurface | null {
  if (!response || response.cells.length === 0) {
    return null
  }

  const allCells = response.cells
  let minLat = allCells[0]!.lat
  let maxLat = allCells[0]!.lat
  let minLon = allCells[0]!.lon
  let maxLon = allCells[0]!.lon

  const lats: number[] = []
  const lons: number[] = []
  for (const cell of allCells) {
    if (cell.lat < minLat) minLat = cell.lat
    if (cell.lat > maxLat) maxLat = cell.lat
    if (cell.lon < minLon) minLon = cell.lon
    if (cell.lon > maxLon) maxLon = cell.lon
    lats.push(cell.lat)
    lons.push(cell.lon)
  }

  const sizeM = response.cell_size_m > 0 ? response.cell_size_m : 50
  const meanLat = (minLat + maxLat) * 0.5
  const cosMeanLat = Math.max(0.01, Math.cos((meanLat * Math.PI) / 180))
  const fallbackStepLat = sizeM / METERS_PER_DEG_LAT
  const fallbackStepLon = sizeM / (METERS_PER_DEG_LAT * cosMeanLat)

  const stepLat = inferGridStep(lats, fallbackStepLat)
  const stepLon = inferGridStep(lons, fallbackStepLon)
  const gridIndex = buildGridIndex(allCells, minLat, maxLat, minLon, maxLon, stepLat, stepLon)

  const activeCells = allCells.filter((cell) => cell.has_data)
  if (activeCells.length === 0) {
    return {
      image: null,
      coordinates: null,
      gridIndex,
    }
  }

  let activeMinLat = activeCells[0]!.lat
  let activeMaxLat = activeCells[0]!.lat
  let activeMinLon = activeCells[0]!.lon
  let activeMaxLon = activeCells[0]!.lon
  for (const cell of activeCells) {
    if (cell.lat < activeMinLat) activeMinLat = cell.lat
    if (cell.lat > activeMaxLat) activeMaxLat = cell.lat
    if (cell.lon < activeMinLon) activeMinLon = cell.lon
    if (cell.lon > activeMaxLon) activeMaxLon = cell.lon
  }

  const rawCols = Math.max(1, Math.round((activeMaxLon - activeMinLon) / stepLon) + 1)
  const rawRows = Math.max(1, Math.round((activeMaxLat - activeMinLat) / stepLat) + 1)
  const ds = Math.max(1, Math.ceil(Math.max(rawCols, rawRows) / MAX_RASTER_DIM))

  const rasterCols = Math.max(1, Math.ceil(rawCols / ds))
  const rasterRows = Math.max(1, Math.ceil(rawRows / ds))
  const rasterStepLat = stepLat * ds
  const rasterStepLon = stepLon * ds

  const srcW = rasterCols + 2 * RASTER_PAD_CELLS
  const srcH = rasterRows + 2 * RASTER_PAD_CELLS
  const srcPremul = new Float32Array(srcW * srcH * 4)
  const srcCounts = new Uint16Array(srcW * srcH)

  for (const cell of activeCells) {
    const rawC = Math.round((cell.lon - activeMinLon) / stepLon)
    const rawR = Math.round((cell.lat - activeMinLat) / stepLat)
    const gx = Math.min(rasterCols - 1, Math.max(0, Math.floor(rawC / ds)))
    const gy = Math.min(rasterRows - 1, Math.max(0, Math.floor(rawR / ds)))

    const px = gx + RASTER_PAD_CELLS
    const py = RASTER_PAD_CELLS + (rasterRows - 1 - gy)
    const texIdx = py * srcW + px
    const rgbaIdx = texIdx * 4

    const [r, g, b] = parseHexRgbUnit(cell.color)
    const alpha = Math.pow(clamp01(cell.confidence), 0.55)

    srcPremul[rgbaIdx] = (srcPremul[rgbaIdx] ?? 0) + r * alpha
    srcPremul[rgbaIdx + 1] = (srcPremul[rgbaIdx + 1] ?? 0) + g * alpha
    srcPremul[rgbaIdx + 2] = (srcPremul[rgbaIdx + 2] ?? 0) + b * alpha
    srcPremul[rgbaIdx + 3] = (srcPremul[rgbaIdx + 3] ?? 0) + alpha
    srcCounts[texIdx] = (srcCounts[texIdx] ?? 0) + 1
  }

  for (let i = 0; i < srcW * srcH; i++) {
    const cnt = srcCounts[i] ?? 0
    if (cnt > 1) {
      const base = i * 4
      srcPremul[base] = srcPremul[base]! / cnt
      srcPremul[base + 1] = srcPremul[base + 1]! / cnt
      srcPremul[base + 2] = srcPremul[base + 2]! / cnt
      srcPremul[base + 3] = srcPremul[base + 3]! / cnt
    }
  }

  const maxSrcDim = Math.max(srcW, srcH)
  const up = maxSrcDim <= 120 ? 4 : maxSrcDim <= 220 ? 3 : maxSrcDim <= 380 ? 2 : 1
  const outW = srcW * up
  const outH = srcH * up

  const upsampled = upsampleBilinearPremul(srcPremul, srcW, srcH, up)
  const blurred = blurPremulSeparable(upsampled, outW, outH, Math.max(0.7, up * 0.55))

  const pixels = new Uint8ClampedArray(outW * outH * 4)
  for (let i = 0; i < outW * outH; i++) {
    const base = i * 4
    const pr = blurred[base]!
    const pg = blurred[base + 1]!
    const pb = blurred[base + 2]!
    const pa = blurred[base + 3]!

    if (pa > 1e-4) {
      pixels[base] = clampByte(Math.round((pr / pa) * 255))
      pixels[base + 1] = clampByte(Math.round((pg / pa) * 255))
      pixels[base + 2] = clampByte(Math.round((pb / pa) * 255))
      pixels[base + 3] = clampByte(Math.round(clamp01(pa) * 255))
    } else {
      // Fill transparent borders with neutral #9CA3AF RGB so GL_LINEAR never introduces dark fringes.
      pixels[base] = 156
      pixels[base + 1] = 163
      pixels[base + 2] = 175
      pixels[base + 3] = 0
    }
  }

  const west = activeMinLon - (RASTER_PAD_CELLS + 0.5) * rasterStepLon
  const east = activeMinLon + (rasterCols - 1 + RASTER_PAD_CELLS + 0.5) * rasterStepLon
  const south = activeMinLat - (RASTER_PAD_CELLS + 0.5) * rasterStepLat
  const north = activeMinLat + (rasterRows - 1 + RASTER_PAD_CELLS + 0.5) * rasterStepLat

  return {
    image: new ImageData(pixels, outW, outH),
    coordinates: [
      [west, north],
      [east, north],
      [east, south],
      [west, south],
    ],
    gridIndex,
  }
}
