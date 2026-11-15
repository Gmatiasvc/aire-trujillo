import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  ApiClientError,
  batchInsertNodeReadings,
  checkAccessSession,
  createAdminNode,
  deleteAdminNode,
  fetchAdminNodeReadings,
  fetchAdminNodes,
  logoutAccess,
  updateAdminNode,
  verifyAccessPhrase,
} from '../api/client'
import type {
  AdminReadingItem,
  BatchReadingsRequest,
  BatchReadingsResponse,
  CreateNodeRequest,
  Node,
  UpdateNodeRequest,
} from '../types/telemetry'

function errorMessageOf(err: unknown, fallback: string): string {
  if (err instanceof ApiClientError && err.message.trim().length > 0) {
    return err.message
  }
  if (err instanceof Error && err.message.trim().length > 0) {
    return err.message
  }
  return fallback
}

export const useAdminStore = defineStore('admin', () => {
  const sessionChecked = ref(false)
  const isAuthenticated = ref(false)
  const isOfflineDemo = ref(false)

  const nodes = ref<Node[]>([])
  const selectedNodeId = ref<string | null>(null)
  const selectedNodeReadings = ref<AdminReadingItem[]>([])

  const isLoadingNodes = ref(false)
  const isLoadingReadings = ref(false)
  const isSavingNode = ref(false)
  const isDeletingNode = ref(false)
  const isSavingBatch = ref(false)

  const nodeError = ref<string | null>(null)
  const batchError = ref<string | null>(null)
  const lastNodeNotice = ref<string | null>(null)
  const lastBatchResult = ref<BatchReadingsResponse | null>(null)

  function clearNodeFeedback(): void {
    nodeError.value = null
    lastNodeNotice.value = null
  }

  let readingsRequestSeq = 0

  const selectedNode = computed(
    () => nodes.value.find((item) => item.id === selectedNodeId.value) ?? null,
  )

  const activeCount = computed(
    () => nodes.value.filter((item) => item.status === 'active').length,
  )

  async function loadReadings(nodeId: string): Promise<void> {
    const seq = ++readingsRequestSeq
    isLoadingReadings.value = true
    try {
      const res = await fetchAdminNodeReadings(nodeId, 24)
      if (seq !== readingsRequestSeq) return
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
      selectedNodeReadings.value = res.data.readings
    } catch (err) {
      if (seq !== readingsRequestSeq) return
      if (err instanceof ApiClientError && err.status === 401) {
        isAuthenticated.value = false
        return
      }
      batchError.value = errorMessageOf(err, 'No se pudieron cargar las lecturas de la estación.')
    } finally {
      if (seq === readingsRequestSeq) {
        isLoadingReadings.value = false
      }
    }
  }

  async function loadNodes(preserveSelection = true): Promise<void> {
    isLoadingNodes.value = true
    nodeError.value = null
    batchError.value = null
    try {
      const res = await fetchAdminNodes()
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
      nodes.value = res.data.nodes

      const stillExists =
        preserveSelection &&
        selectedNodeId.value !== null &&
        res.data.nodes.some((item) => item.id === selectedNodeId.value)

      const nextId = stillExists ? selectedNodeId.value : (res.data.nodes[0]?.id ?? null)
      selectedNodeId.value = nextId

      if (nextId) {
        await loadReadings(nextId)
      } else {
        selectedNodeReadings.value = []
      }
    } catch (err) {
      if (err instanceof ApiClientError && err.status === 401) {
        isAuthenticated.value = false
        return
      }
      nodeError.value = errorMessageOf(err, 'No se pudo cargar el directorio de estaciones.')
    } finally {
      isLoadingNodes.value = false
    }
  }

  async function bootstrapSession(): Promise<void> {
    try {
      const res = await checkAccessSession()
      if (res.data.ok) {
        isAuthenticated.value = true
        isOfflineDemo.value = res.isOfflineDemo
        await loadNodes(false)
      } else {
        isAuthenticated.value = false
        isOfflineDemo.value = false
      }
    } catch {
      isAuthenticated.value = false
      isOfflineDemo.value = false
    } finally {
      sessionChecked.value = true
    }
  }

  async function login(phrase: string): Promise<boolean> {
    try {
      const res = await verifyAccessPhrase(phrase)
      if (!res.data.ok) {
        return false
      }
      isAuthenticated.value = true
      isOfflineDemo.value = res.isOfflineDemo
      nodeError.value = null
      batchError.value = null
      lastNodeNotice.value = null
      await loadNodes(false)
      return true
    } catch {
      // Keep gate behavior strictly silent on bad credentials or rate limit
      return false
    }
  }

  async function logout(): Promise<void> {
    try {
      await logoutAccess()
    } catch {
      // Clear local admin state even if backend session already expired or is offline
    }
    isAuthenticated.value = false
    isOfflineDemo.value = false
    nodes.value = []
    selectedNodeId.value = null
    selectedNodeReadings.value = []
    lastBatchResult.value = null
    nodeError.value = null
    batchError.value = null
    lastNodeNotice.value = null
  }

  async function retryLiveApi(): Promise<void> {
    nodeError.value = null
    batchError.value = null
    lastNodeNotice.value = null
    try {
      const sessionRes = await checkAccessSession()
      if (sessionRes.isOfflineDemo) {
        isOfflineDemo.value = true
        return
      }
      if (!sessionRes.data.ok) {
        isAuthenticated.value = false
        isOfflineDemo.value = false
        return
      }
      isOfflineDemo.value = false
      await loadNodes(true)
    } catch {
      nodeError.value = 'No se pudo restablecer la sesión con :8080.'
    }
  }

  async function selectNode(nodeId: string | null): Promise<void> {
    selectedNodeId.value = nodeId
    lastBatchResult.value = null
    nodeError.value = null
    batchError.value = null
    lastNodeNotice.value = null
    if (nodeId) {
      await loadReadings(nodeId)
    } else {
      selectedNodeReadings.value = []
    }
  }

  async function createNode(payload: CreateNodeRequest): Promise<Node | null> {
    isSavingNode.value = true
    nodeError.value = null
    lastNodeNotice.value = null
    try {
      const res = await createAdminNode(payload)
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
      nodes.value = [...nodes.value, res.data]
      selectedNodeId.value = res.data.id
      selectedNodeReadings.value = []
      lastBatchResult.value = null
      lastNodeNotice.value = `Estación ${res.data.code} creada en directorio.`
      await loadReadings(res.data.id)
      return res.data
    } catch (err) {
      if (err instanceof ApiClientError && err.status === 401) {
        isAuthenticated.value = false
        return null
      }
      nodeError.value = errorMessageOf(err, 'No se pudo crear la estación.')
      return null
    } finally {
      isSavingNode.value = false
    }
  }

  async function updateNode(nodeId: string, payload: UpdateNodeRequest): Promise<Node | null> {
    isSavingNode.value = true
    nodeError.value = null
    lastNodeNotice.value = null
    try {
      const res = await updateAdminNode(nodeId, payload)
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
      nodes.value = nodes.value.map((item) => (item.id === nodeId ? res.data : item))
      lastNodeNotice.value = `Estación ${res.data.code} actualizada.`
      return res.data
    } catch (err) {
      if (err instanceof ApiClientError && err.status === 401) {
        isAuthenticated.value = false
        return null
      }
      nodeError.value = errorMessageOf(err, 'No se pudo actualizar la estación.')
      return null
    } finally {
      isSavingNode.value = false
    }
  }

  async function deleteNode(nodeId: string): Promise<boolean> {
    isDeletingNode.value = true
    nodeError.value = null
    lastNodeNotice.value = null
    try {
      const res = await deleteAdminNode(nodeId)
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
      const remaining = nodes.value.filter((item) => item.id !== nodeId)
      nodes.value = remaining
      const nextId = remaining[0]?.id ?? null
      selectedNodeId.value = nextId
      lastBatchResult.value = null
      lastNodeNotice.value = 'Estación eliminada.'
      if (nextId) {
        await loadReadings(nextId)
      } else {
        selectedNodeReadings.value = []
      }
      return true
    } catch (err) {
      if (err instanceof ApiClientError && err.status === 401) {
        isAuthenticated.value = false
        return false
      }
      nodeError.value = errorMessageOf(err, 'No se pudo eliminar la estación.')
      return false
    } finally {
      isDeletingNode.value = false
    }
  }

  async function submitReadingsBatch(
    nodeId: string,
    payload: BatchReadingsRequest,
  ): Promise<BatchReadingsResponse | null> {
    isSavingBatch.value = true
    batchError.value = null
    lastBatchResult.value = null
    try {
      const res = await batchInsertNodeReadings(nodeId, payload)
      if (res.isOfflineDemo) {
        isOfflineDemo.value = true
      }
      lastBatchResult.value = res.data
      await loadReadings(nodeId)
      return res.data
    } catch (err) {
      if (err instanceof ApiClientError && err.status === 401) {
        isAuthenticated.value = false
        return null
      }
      batchError.value = errorMessageOf(err, 'No se pudo guardar el lote de lecturas.')
      return null
    } finally {
      isSavingBatch.value = false
    }
  }

  return {
    sessionChecked,
    isAuthenticated,
    isOfflineDemo,
    nodes,
    selectedNodeId,
    selectedNodeReadings,
    selectedNode,
    activeCount,
    isLoadingNodes,
    isLoadingReadings,
    isSavingNode,
    isDeletingNode,
    isSavingBatch,
    nodeError,
    batchError,
    lastNodeNotice,
    lastBatchResult,
    clearNodeFeedback,
    bootstrapSession,
    login,
    logout,
    retryLiveApi,
    loadNodes,
    selectNode,
    loadReadings,
    createNode,
    updateNode,
    deleteNode,
    submitReadingsBatch,
  }
})
