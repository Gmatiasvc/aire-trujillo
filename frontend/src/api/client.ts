import {
  createDemoAdminNode,
  deleteDemoAdminNode,
  getDemoAdminNodeReadings,
  getDemoAdminNodes,
  getDemoHeatmap,
  getDemoPlaybackTimestamps,
  getDemoPoint,
  getDemoPublicNodes,
  updateDemoAdminNode,
  upsertDemoNodeReadingsBatch,
} from './mockFallback'
import type {
  AccessResponse,
  AccessVerifyRequest,
  AdminNodesResponse,
  AdminReadingsResponse,
  AirQualityStandard,
  BatchReadingsRequest,
  BatchReadingsResponse,
  CreateNodeRequest,
  ErrorResponse,
  HeatmapResponse,
  LayerKey,
  Node,
  PlaybackTimestampsResponse,
  PointResponse,
  PublicNodesResponse,
  UpdateNodeRequest,
} from '../types/telemetry'

const API_BASE = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '')

export class ApiClientError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiClientError'
    this.status = status
    this.code = code
  }
}

export interface ApiResult<T> {
  data: T
  isOfflineDemo: boolean
}

function isBackendUnreachableStatus(status: number): boolean {
  return status === 502 || status === 503 || status === 504
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE'
  body?: unknown
  allowDemoFallback?: boolean
}

function resolveOfflineOrThrow<T>(
  allowDemoFallback: boolean,
  buildDemoFallback: () => T,
  status: number,
  code: string,
  message: string,
): ApiResult<T> {
  if (import.meta.env.DEV && allowDemoFallback) {
    return {
      data: buildDemoFallback(),
      isOfflineDemo: true,
    }
  }
  throw new ApiClientError(status, code, message)
}

async function requestJsonOrDemo<T>(
  path: string,
  options: RequestOptions,
  buildDemoFallback: () => T,
): Promise<ApiResult<T>> {
  const allowDemoFallback = options.allowDemoFallback ?? true
  const headers: Record<string, string> = {
    Accept: 'application/json',
  }
  const init: RequestInit = {
    method: options.method ?? 'GET',
    headers,
    credentials: 'include',
  }
  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    init.body = JSON.stringify(options.body)
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, init)
  } catch {
    return resolveOfflineOrThrow(
      allowDemoFallback,
      buildDemoFallback,
      0,
      'network_error',
      'No se pudo conectar con el servidor. Intenta nuevamente en unos segundos.',
    )
  }

  if (isBackendUnreachableStatus(response.status)) {
    return resolveOfflineOrThrow(
      allowDemoFallback,
      buildDemoFallback,
      response.status,
      'backend_unavailable',
      'El servidor está iniciando o no responde. Intenta nuevamente en unos segundos.',
    )
  }

  const contentType = response.headers.get('content-type') ?? ''
  const isJson = contentType.includes('application/json')

  if (!response.ok) {
    // Live backend always returns JSON (or explicit 401/403/429). Proxy ECONNREFUSED (500 empty)
    // or static SPA 404/405 HTML indicates local :8080 is absent in dev.
    const isAuthOrRateLimit =
      response.status === 401 || response.status === 403 || response.status === 429
    if (!isJson && !isAuthOrRateLimit) {
      return resolveOfflineOrThrow(
        allowDemoFallback,
        buildDemoFallback,
        response.status,
        'backend_unavailable',
        'El servidor no devolvió una respuesta válida.',
      )
    }

    if (isJson) {
      const body = (await response.json()) as Partial<ErrorResponse>
      throw new ApiClientError(
        response.status,
        body.error ?? 'api_error',
        body.message ?? `Error HTTP ${response.status}`,
      )
    }
    throw new ApiClientError(response.status, 'http_error', `Error HTTP ${response.status}`)
  }

  if (response.status === 204) {
    return {
      data: undefined as T,
      isOfflineDemo: false,
    }
  }

  // If Vite or an SPA fallback returned HTML instead of JSON, treat backend as absent.
  if (!isJson) {
    return resolveOfflineOrThrow(
      allowDemoFallback,
      buildDemoFallback,
      response.status,
      'invalid_content_type',
      'Respuesta inesperada del servidor.',
    )
  }

  const payload = (await response.json()) as T
  return {
    data: payload,
    isOfflineDemo: false,
  }
}

function fetchJsonOrDemo<T>(path: string, buildDemoFallback: () => T): Promise<ApiResult<T>> {
  return requestJsonOrDemo(path, { method: 'GET' }, buildDemoFallback)
}

export function fetchPublicNodes(): Promise<ApiResult<PublicNodesResponse>> {
  return fetchJsonOrDemo('/api/v1/nodes', getDemoPublicNodes)
}

export function fetchHeatmap(
  layer: LayerKey,
  standard: AirQualityStandard,
  at?: string | null,
): Promise<ApiResult<HeatmapResponse>> {
  const params = new URLSearchParams({
    layer,
    standard,
  })
  if (at) {
    params.set('at', at)
  }
  return fetchJsonOrDemo(`/api/v1/map/heatmap?${params.toString()}`, () =>
    getDemoHeatmap(layer, standard, at ?? undefined),
  )
}

export function fetchPointSeries(
  lat: number,
  lon: number,
  layer: LayerKey,
  standard: AirQualityStandard,
  at?: string | null,
): Promise<ApiResult<PointResponse>> {
  const params = new URLSearchParams({
    lat: lat.toFixed(6),
    lon: lon.toFixed(6),
    layer,
    standard,
  })
  if (at) {
    params.set('at', at)
  }
  return fetchJsonOrDemo(`/api/v1/map/point?${params.toString()}`, () =>
    getDemoPoint(lat, lon, layer, standard, at ?? undefined),
  )
}

export function fetchPlaybackTimestamps(): Promise<ApiResult<PlaybackTimestampsResponse>> {
  return fetchJsonOrDemo('/api/v1/map/playback/timestamps', getDemoPlaybackTimestamps)
}

export async function checkAccessSession(): Promise<ApiResult<AccessResponse>> {
  try {
    // When backend :8080 is absent on initial load, stay locked on the gate (ok: false).
    return await requestJsonOrDemo<AccessResponse>(
      '/api/v1/access/session',
      { method: 'GET' },
      () => ({ ok: false }),
    )
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 401) {
      return { data: { ok: false }, isOfflineDemo: false }
    }
    throw err
  }
}

export function verifyAccessPhrase(phrase: string): Promise<ApiResult<AccessResponse>> {
  const payload: AccessVerifyRequest = { phrase }
  return requestJsonOrDemo<AccessResponse>(
    '/api/v1/access/verify',
    { method: 'POST', body: payload, allowDemoFallback: false },
    () => ({ ok: false }),
  )
}

export function logoutAccess(): Promise<ApiResult<AccessResponse>> {
  return requestJsonOrDemo<AccessResponse>(
    '/api/v1/access/logout',
    { method: 'POST' },
    () => ({ ok: true }),
  )
}

export function fetchAdminNodes(): Promise<ApiResult<AdminNodesResponse>> {
  return fetchJsonOrDemo('/api/v1/admin/nodes', getDemoAdminNodes)
}

export function createAdminNode(payload: CreateNodeRequest): Promise<ApiResult<Node>> {
  return requestJsonOrDemo(
    '/api/v1/admin/nodes',
    { method: 'POST', body: payload },
    () => createDemoAdminNode(payload),
  )
}

export function updateAdminNode(
  nodeId: string,
  payload: UpdateNodeRequest,
): Promise<ApiResult<Node>> {
  return requestJsonOrDemo(
    `/api/v1/admin/nodes/${encodeURIComponent(nodeId)}`,
    { method: 'PATCH', body: payload },
    () => {
      const updated = updateDemoAdminNode(nodeId, payload)
      if (!updated) {
        throw new ApiClientError(404, 'not_found', 'Nodo no encontrado')
      }
      return updated
    },
  )
}

export function deleteAdminNode(nodeId: string): Promise<ApiResult<void>> {
  return requestJsonOrDemo<void>(
    `/api/v1/admin/nodes/${encodeURIComponent(nodeId)}`,
    { method: 'DELETE' },
    () => {
      const deleted = deleteDemoAdminNode(nodeId)
      if (!deleted) {
        throw new ApiClientError(404, 'not_found', 'Nodo no encontrado')
      }
    },
  )
}

export function fetchAdminNodeReadings(
  nodeId: string,
  hours = 24,
): Promise<ApiResult<AdminReadingsResponse>> {
  const params = new URLSearchParams({ hours: String(hours) })
  return fetchJsonOrDemo(
    `/api/v1/admin/nodes/${encodeURIComponent(nodeId)}/readings?${params.toString()}`,
    () => getDemoAdminNodeReadings(nodeId, hours),
  )
}

export function batchInsertNodeReadings(
  nodeId: string,
  payload: BatchReadingsRequest,
): Promise<ApiResult<BatchReadingsResponse>> {
  return requestJsonOrDemo(
    `/api/v1/admin/nodes/${encodeURIComponent(nodeId)}/readings/batch`,
    { method: 'POST', body: payload },
    () => upsertDemoNodeReadingsBatch(nodeId, payload),
  )
}
