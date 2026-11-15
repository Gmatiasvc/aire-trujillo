<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import UiIcon from '../UiIcon.vue'
import AdminCoordPickerMap from './AdminCoordPickerMap.vue'
import type { CreateNodeRequest, Node, NodeStatus, UpdateNodeRequest } from '../../types/telemetry'

const props = defineProps<{
  nodes: Node[]
  selectedNode: Node | null
  isLoading: boolean
  isSaving: boolean
  isDeleting: boolean
  storeError: string | null
  nodeNotice: string | null
}>()

const emit = defineEmits<{
  'select-node': [nodeId: string]
  'create-node': [payload: CreateNodeRequest]
  'update-node': [nodeId: string, payload: UpdateNodeRequest]
  'delete-node': [nodeId: string]
  'clear-notice': []
}>()

const searchQuery = ref('')
const isCreateMode = ref(false)
const confirmDeleteId = ref<string | null>(null)

const formCode = ref('')
const formName = ref('')
const formStatus = ref<NodeStatus>('active')
const formLat = ref('-8.111600')
const formLon = ref('-79.028700')
const localError = ref<string | null>(null)

const filteredNodes = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return props.nodes
  return props.nodes.filter(
    (node) => node.code.toLowerCase().includes(q) || node.name.toLowerCase().includes(q),
  )
})

const parsedDraftLat = computed<number | null>(() => {
  const val = Number(formLat.value)
  return formLat.value.trim() !== '' && Number.isFinite(val) ? val : null
})

const parsedDraftLon = computed<number | null>(() => {
  const val = Number(formLon.value)
  return formLon.value.trim() !== '' && Number.isFinite(val) ? val : null
})

function populateFromNode(node: Node): void {
  formCode.value = node.code
  formName.value = node.name
  formStatus.value = node.status
  formLat.value = node.lat.toFixed(6)
  formLon.value = node.lon.toFixed(6)
  localError.value = null
}

watch(
  () => props.selectedNode,
  (node) => {
    confirmDeleteId.value = null
    if (node) {
      isCreateMode.value = false
      populateFromNode(node)
    } else if (!isCreateMode.value && props.nodes.length === 0) {
      startCreateMode()
    }
  },
  { immediate: true },
)

function startCreateMode(): void {
  isCreateMode.value = true
  confirmDeleteId.value = null
  localError.value = null
  emit('clear-notice')
  formCode.value = ''
  formName.value = ''
  formStatus.value = 'active'
  if (props.selectedNode) {
    formLat.value = (props.selectedNode.lat + 0.0025).toFixed(6)
    formLon.value = (props.selectedNode.lon - 0.0025).toFixed(6)
  } else {
    formLat.value = '-8.111600'
    formLon.value = '-79.028700'
  }
}

function cancelCreateMode(): void {
  isCreateMode.value = false
  localError.value = null
  emit('clear-notice')
  if (props.selectedNode) {
    populateFromNode(props.selectedNode)
  }
}

function handlePickMapCoordinate(coord: { lat: number; lon: number }): void {
  formLat.value = coord.lat.toFixed(6)
  formLon.value = coord.lon.toFixed(6)
  localError.value = null
  emit('clear-notice')
}

function handleSelectDirectoryNode(nodeId: string): void {
  isCreateMode.value = false
  confirmDeleteId.value = null
  emit('select-node', nodeId)
}

function validateForm(): { lat: number; lon: number; code: string; name: string } | null {
  const name = formName.value.trim()
  const code = formCode.value.trim()
  if (name.length < 1 || name.length > 120) {
    localError.value = 'El nombre debe tener entre 1 y 120 caracteres.'
    return null
  }
  if (isCreateMode.value && code.length > 64) {
    localError.value = 'El código no puede exceder 64 caracteres.'
    return null
  }
  const lat = Number(formLat.value)
  const lon = Number(formLon.value)
  if (!Number.isFinite(lat) || lat < -90 || lat > 90) {
    localError.value = 'La latitud debe ser numérica entre -90 y 90.'
    return null
  }
  if (!Number.isFinite(lon) || lon < -180 || lon > 180) {
    localError.value = 'La longitud debe ser numérica entre -180 y 180.'
    return null
  }
  localError.value = null
  return { lat, lon, code, name }
}

function handleSubmitForm(): void {
  const valid = validateForm()
  if (!valid) return
  emit('clear-notice')

  if (isCreateMode.value || !props.selectedNode) {
    emit('create-node', {
      code: valid.code,
      name: valid.name,
      lat: valid.lat,
      lon: valid.lon,
      status: formStatus.value,
    })
    return
  }

  emit('update-node', props.selectedNode.id, {
    name: valid.name,
    lat: valid.lat,
    lon: valid.lon,
    status: formStatus.value,
  })
}

function handleConfirmDelete(): void {
  if (!props.selectedNode) return
  const targetId = props.selectedNode.id
  confirmDeleteId.value = null
  emit('delete-node', targetId)
}
</script>

<template>
  <section class="node-panel-card" aria-label="Gestión de estaciones y nodos virtuales">
    <header class="panel-header">
      <div class="panel-title-group">
        <h2 class="panel-title">Estaciones y ubicación</h2>
        <span class="panel-count tabular">{{ nodes.length }} registradas</span>
      </div>

      <button
        v-if="!isCreateMode"
        type="button"
        class="raked-new-btn"
        @click="startCreateMode"
      >
        <span>
          <UiIcon name="plus" :size="13" />
          <span>Nuevo nodo</span>
        </span>
      </button>
      <button
        v-else-if="selectedNode"
        type="button"
        class="secondary-link-btn"
        @click="cancelCreateMode"
      >
        Volver al nodo
      </button>
    </header>

    <div class="directory-block">
      <div class="search-row">
        <input
          v-model="searchQuery"
          type="search"
          class="mono-input search-input"
          placeholder="Filtrar por código o nombre…"
          aria-label="Filtrar estaciones registradas"
        />
      </div>

      <div v-if="isLoading && nodes.length === 0" class="directory-empty" aria-busy="true">
        Cargando directorio de estaciones…
      </div>

      <ul v-else-if="filteredNodes.length > 0" class="directory-list" role="listbox" aria-label="Listado de estaciones">
        <li v-for="node in filteredNodes" :key="node.id">
          <button
            type="button"
            class="directory-item"
            :class="{ 'is-selected': !isCreateMode && selectedNode?.id === node.id }"
            :aria-selected="!isCreateMode && selectedNode?.id === node.id"
            @click="handleSelectDirectoryNode(node.id)"
          >
            <span class="item-main">
              <span
                class="item-status-dot"
                :class="node.status === 'active' ? 'is-active' : 'is-inactive'"
                :title="node.status === 'active' ? 'Activa' : 'Inactiva'"
                aria-hidden="true"
              ></span>
              <span class="item-code tabular">{{ node.code }}</span>
              <span class="item-name">{{ node.name }}</span>
            </span>

            <span class="item-badges">
              <span v-if="node.is_virtual" class="kind-badge is-virtual">Virtual</span>
              <span class="state-badge" :class="node.status === 'active' ? 'is-active' : 'is-inactive'">
                {{ node.status === 'active' ? 'Activa' : 'Inactiva' }}
              </span>
            </span>
          </button>
        </li>
      </ul>

      <div v-else class="directory-empty">
        Sin coincidencias para «{{ searchQuery }}».
      </div>
    </div>

    <div class="form-divider" aria-hidden="true"></div>

    <form class="node-form" @submit.prevent="handleSubmitForm">
      <div class="form-mode-head">
        <h3 class="form-heading">
          {{ isCreateMode || !selectedNode ? 'Crear nodo virtual' : `Editar ${selectedNode.code}` }}
        </h3>
        <span v-if="!isCreateMode && selectedNode" class="form-coord-meta tabular">
          {{ selectedNode.is_virtual ? 'Nodo virtual' : 'Estación base' }}
        </span>
      </div>

      <AdminCoordPickerMap
        :nodes="nodes"
        :selected-node-id="isCreateMode ? null : (selectedNode?.id ?? null)"
        :draft-lat="parsedDraftLat"
        :draft-lon="parsedDraftLon"
        @pick-coordinate="handlePickMapCoordinate"
        @select-node="handleSelectDirectoryNode"
      />

      <div class="fields-grid">
        <label class="field-block">
          <span class="field-label">
            Código
            <em v-if="isCreateMode" class="field-opt">(opcional)</em>
          </span>
          <input
            v-model="formCode"
            type="text"
            class="mono-input"
            :disabled="!isCreateMode && Boolean(selectedNode)"
            :placeholder="isCreateMode ? 'Auto: TRJ-V-XX' : ''"
            maxlength="64"
            aria-label="Código de estación"
          />
        </label>

        <div class="field-block">
          <span class="field-label">Estado operativo</span>
          <div class="status-switch" role="group" aria-label="Estado de la estación">
            <button
              type="button"
              class="status-option"
              :class="{ 'is-chosen': formStatus === 'active' }"
              :aria-pressed="formStatus === 'active'"
              @click="formStatus = 'active'"
            >
              Activa
            </button>
            <button
              type="button"
              class="status-option"
              :class="{ 'is-chosen-inactive': formStatus === 'inactive' }"
              :aria-pressed="formStatus === 'inactive'"
              @click="formStatus = 'inactive'"
            >
              Inactiva
            </button>
          </div>
        </div>

        <label class="field-block is-span-2">
          <span class="field-label">Nombre descriptivo</span>
          <input
            v-model="formName"
            type="text"
            class="body-input"
            required
            maxlength="120"
            placeholder="Ej. Av. Mansiche · Ovalo"
            aria-label="Nombre de la estación"
          />
        </label>

        <label class="field-block">
          <span class="field-label">Latitud</span>
          <input
            v-model="formLat"
            type="number"
            step="0.000001"
            min="-90"
            max="90"
            class="mono-input tabular"
            required
            aria-label="Latitud decimal"
          />
        </label>

        <label class="field-block">
          <span class="field-label">Longitud</span>
          <input
            v-model="formLon"
            type="number"
            step="0.000001"
            min="-180"
            max="180"
            class="mono-input tabular"
            required
            aria-label="Longitud decimal"
          />
        </label>
      </div>

      <p v-if="localError || storeError" class="form-alert is-error" role="alert">
        {{ localError || storeError }}
      </p>
      <p v-else-if="nodeNotice" class="form-alert is-ok" role="status">
        {{ nodeNotice }}
      </p>

      <div class="form-actions">
        <button
          type="submit"
          class="raked-submit-btn"
          :disabled="isSaving || isDeleting"
        >
          <span>
            <UiIcon name="check" :size="14" />
            <span>
              {{
                isSaving
                  ? 'Guardando…'
                  : isCreateMode || !selectedNode
                    ? 'Crear nodo'
                    : 'Guardar cambios'
              }}
            </span>
          </span>
        </button>

        <div v-if="!isCreateMode && selectedNode" class="delete-zone">
          <button
            v-if="confirmDeleteId !== selectedNode.id"
            type="button"
            class="danger-ghost-btn"
            :disabled="isSaving || isDeleting"
            @click="confirmDeleteId = selectedNode.id"
          >
            <UiIcon name="trash" :size="14" />
            <span>Eliminar</span>
          </button>

          <div v-else class="delete-confirm-row">
            <button
              type="button"
              class="danger-solid-btn"
              :disabled="isDeleting"
              @click="handleConfirmDelete"
            >
              {{ isDeleting ? 'Eliminando…' : 'Confirmar baja' }}
            </button>
            <button
              type="button"
              class="secondary-link-btn"
              @click="confirmDeleteId = null"
            >
              Cancelar
            </button>
          </div>
        </div>
      </div>
    </form>
  </section>
</template>

<style scoped>
.node-panel-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-width: 0;
  width: 100%;
  padding: 16px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory);
}

.panel-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--rule-soft);
}

.panel-title-group {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.panel-title {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  color: var(--ink-primary);
}

.panel-count {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--ink-secondary);
}

.raked-new-btn {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.16s var(--ease-expo),
    border-color 0.16s var(--ease-expo);
}

.raked-new-btn > span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.raked-new-btn:hover {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.secondary-link-btn {
  padding: 5px 10px;
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
}

.secondary-link-btn:hover {
  border-color: var(--accent-vermilion);
}

.directory-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
}

.search-input,
.mono-input,
.body-input {
  width: 100%;
  padding: 8px 10px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  font-size: 13px;
}

.mono-input {
  font-family: var(--font-mono);
  font-size: 12.5px;
}

.body-input {
  font-family: var(--font-body);
}

.mono-input:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.mono-input:focus-visible,
.body-input:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.directory-list {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: 224px;
  overflow-x: hidden;
  overflow-y: auto;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.directory-list li + li {
  border-top: 1px solid var(--rule-soft);
}

.directory-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  padding: 8px 10px;
  border: none;
  background: transparent;
  color: var(--ink-primary);
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.14s var(--ease-expo),
    color 0.14s var(--ease-expo);
}

.directory-item:hover {
  background: var(--surface-ivory);
}

.directory-item.is-selected {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.directory-item.is-selected .item-code {
  color: var(--ink-on-carbon);
}

.directory-item.is-selected .item-name {
  color: var(--ink-muted-on-carbon);
}

.directory-item.is-selected .state-badge.is-active {
  border-color: #4ade80;
  color: #4ade80;
}

.directory-item.is-selected .state-badge.is-inactive {
  border-color: var(--glass-slate-rule);
  background: var(--surface-slate);
  color: var(--ink-on-carbon);
}

.item-main {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.item-status-dot {
  width: 8px;
  height: 8px;
  flex-shrink: 0;
}

.item-status-dot.is-active {
  background: #22c55e;
}

.item-status-dot.is-inactive {
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.item-code {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  color: var(--ink-primary);
  flex-shrink: 0;
}

.item-name {
  font-size: 12.5px;
  color: var(--ink-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-badges {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
}

.kind-badge,
.state-badge {
  padding: 2px 6px;
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.kind-badge.is-virtual {
  border: 1px solid var(--accent-amber);
  background: var(--accent-amber-fill);
  color: #181a1c;
}

.state-badge.is-active {
  border: 1px solid #22c55e;
  color: #166534;
}

:root[data-theme='dark'] .state-badge.is-active {
  color: #4ade80;
}

.state-badge.is-inactive {
  border: 1px solid var(--rule-soft);
  background: var(--surface-ivory);
  color: var(--ink-secondary);
}

.directory-empty {
  padding: 14px;
  border: 1px dashed var(--rule-color);
  font-size: 12.5px;
  color: var(--ink-secondary);
  text-align: center;
}

.form-divider {
  height: 1px;
  background: var(--rule-color);
}

.node-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-mode-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}

.form-heading {
  font-family: var(--font-display);
  font-size: 1.05rem;
  font-weight: 800;
  text-transform: uppercase;
  color: var(--ink-primary);
}

.form-coord-meta {
  font-family: var(--font-mono);
  font-size: 10.5px;
  color: var(--ink-secondary);
}

.fields-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.field-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-block.is-span-2 {
  grid-column: span 2;
}

.field-label {
  font-family: var(--font-mono);
  font-size: 10.5px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--ink-secondary);
}

.field-opt {
  font-style: normal;
  font-weight: 400;
  text-transform: none;
}

.status-switch {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
}

.status-option {
  padding: 7px 8px;
  border: none;
  background: transparent;
  color: var(--ink-secondary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
}

.status-option + .status-option {
  border-left: 1px solid var(--rule-color);
}

.status-option.is-chosen {
  background: var(--surface-carbon);
  color: #4ade80;
}

.status-option.is-chosen-inactive {
  background: var(--surface-carbon);
  color: var(--accent-amber-fill);
}

.form-alert {
  padding: 8px 10px;
  border: 1.5px solid var(--rule-color);
  font-family: var(--font-mono);
  font-size: 11.5px;
}

.form-alert.is-error {
  border-color: var(--accent-vermilion);
  background: var(--surface-ivory-2);
  color: var(--accent-vermilion);
}

.form-alert.is-ok {
  border-color: #22c55e;
  background: var(--surface-ivory-2);
  color: #166534;
}

:root[data-theme='dark'] .form-alert.is-ok {
  color: #4ade80;
}

.form-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.raked-submit-btn {
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

.raked-submit-btn > span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.raked-submit-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.danger-ghost-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 7px 11px;
  border: 1px solid var(--accent-vermilion);
  background: transparent;
  color: var(--accent-vermilion);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
}

.danger-ghost-btn:hover {
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.delete-confirm-row {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.danger-solid-btn {
  padding: 6px 10px;
  border: 1px solid var(--accent-vermilion);
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  cursor: pointer;
}

.directory-item:focus-visible,
.raked-new-btn:focus-visible,
.raked-submit-btn:focus-visible,
.status-option:focus-visible,
.danger-ghost-btn:focus-visible,
.danger-solid-btn:focus-visible,
.secondary-link-btn:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

@media (max-width: 560px) {
  .fields-grid {
    grid-template-columns: 1fr;
  }

  .field-block.is-span-2 {
    grid-column: span 1;
  }
}
</style>
