<script setup lang="ts">
import { onMounted, ref } from 'vue'
import UiIcon from '../components/UiIcon.vue'
import AdminBatchPanel from '../components/edit/AdminBatchPanel.vue'
import AdminNodePanel from '../components/edit/AdminNodePanel.vue'
import AdminTopBar from '../components/edit/AdminTopBar.vue'
import { useAdminStore } from '../stores/admin'
import { useThemeStore } from '../stores/theme'

useThemeStore()
const adminStore = useAdminStore()

const phrase = ref('')
const isSubmittingGate = ref(false)

onMounted(() => {
  void adminStore.bootstrapSession()
})

async function handleGateSubmit(): Promise<void> {
  if (isSubmittingGate.value) return
  const candidate = phrase.value
  if (candidate.trim().length === 0) {
    phrase.value = ''
    return
  }

  isSubmittingGate.value = true
  try {
    await adminStore.login(candidate)
  } finally {
    phrase.value = ''
    isSubmittingGate.value = false
  }
}

async function handleLogout(): Promise<void> {
  phrase.value = ''
  await adminStore.logout()
}
</script>

<template>
  <main v-if="!adminStore.isAuthenticated" class="discreet-gate">
    <form class="discreet-form" @submit.prevent="handleGateSubmit">
      <input
        v-model="phrase"
        type="password"
        autocomplete="off"
        class="discreet-input"
        :disabled="isSubmittingGate"
      />
      <button
        type="submit"
        class="discreet-submit"
        aria-label="Enviar"
        :disabled="isSubmittingGate"
      >
        <UiIcon name="arrow-right" :size="15" />
      </button>
    </form>
  </main>

  <div v-else class="edit-shell">
    <AdminTopBar
      :total-nodes="adminStore.nodes.length"
      :active-nodes="adminStore.activeCount"
      @logout="handleLogout"
    />

    <div v-if="adminStore.isOfflineDemo" class="offline-demo-banner" role="status">
      <div class="banner-copy">
        <UiIcon name="alert" :size="15" />
        <span>Sin conexión al backend :8080 — sesión de muestra local</span>
      </div>
      <button
        type="button"
        class="retry-api-btn"
        @click="adminStore.retryLiveApi"
      >
        <UiIcon name="refresh" :size="13" />
        <span>Reintentar API</span>
      </button>
    </div>

    <main class="workbench-main">
      <AdminNodePanel
        :nodes="adminStore.nodes"
        :selected-node="adminStore.selectedNode"
        :is-loading="adminStore.isLoadingNodes"
        :is-saving="adminStore.isSavingNode"
        :is-deleting="adminStore.isDeletingNode"
        :store-error="adminStore.nodeError"
        :node-notice="adminStore.lastNodeNotice"
        @select-node="adminStore.selectNode"
        @create-node="adminStore.createNode"
        @update-node="adminStore.updateNode"
        @delete-node="adminStore.deleteNode"
        @clear-notice="adminStore.clearNodeFeedback"
      />

      <AdminBatchPanel
        :selected-node="adminStore.selectedNode"
        :saved-readings="adminStore.selectedNodeReadings"
        :is-loading-readings="adminStore.isLoadingReadings"
        :is-saving-batch="adminStore.isSavingBatch"
        :last-batch-result="adminStore.lastBatchResult"
        :store-error="adminStore.batchError"
        @submit-batch="adminStore.submitReadingsBatch"
      />
    </main>
  </div>
</template>

<style scoped>
.discreet-gate {
  min-height: 100svh;
  display: grid;
  place-items: center;
  background: var(--surface-ivory);
  padding: 24px;
}

.discreet-form {
  display: flex;
  gap: 8px;
  width: min(320px, 100%);
}

.discreet-input {
  flex: 1;
  padding: 10px 12px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-ivory-2);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 13px;
}

.discreet-input:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.discreet-submit {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 10px 16px;
  border: 1.5px solid var(--rule-color);
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  cursor: pointer;
  transition:
    background-color 0.16s var(--ease-expo),
    border-color 0.16s var(--ease-expo),
    color 0.16s var(--ease-expo);
}

.discreet-submit:hover:not(:disabled) {
  background: var(--accent-vermilion);
  border-color: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.discreet-submit:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.discreet-submit:disabled {
  opacity: 0.6;
  cursor: wait;
}

.edit-shell {
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  overflow-x: clip;
  background: var(--bed-bg);
  color: var(--ink-primary);
}

.offline-demo-banner {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 16px;
  padding: 9px 18px;
  border-bottom: 1.5px solid var(--accent-amber);
  background: var(--accent-amber-fill);
  color: #181a1c;
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
}

.banner-copy {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.retry-api-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  border: 1.5px solid #181a1c;
  background: var(--surface-ivory);
  color: var(--ink-primary);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  cursor: pointer;
  transition:
    background-color 0.15s var(--ease-expo),
    color 0.15s var(--ease-expo);
}

.retry-api-btn:hover {
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
}

.retry-api-btn:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

.workbench-main {
  width: 100%;
  max-width: 1560px;
  margin: 0 auto;
  padding: 18px;
  display: grid;
  grid-template-columns: minmax(380px, 456px) minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.workbench-main > * {
  min-width: 0;
}

@media (max-width: 1079px) {
  .workbench-main {
    grid-template-columns: 1fr;
    padding: 14px;
    gap: 14px;
  }
}

@media (max-width: 600px) {
  .offline-demo-banner {
    padding: 9px 12px;
  }

  .workbench-main {
    padding: 10px;
    gap: 12px;
  }
}
</style>
