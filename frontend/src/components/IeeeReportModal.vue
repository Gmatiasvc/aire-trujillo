<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import UiIcon from './UiIcon.vue'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const closeBtnRef = ref<HTMLButtonElement | null>(null)

function onKeyDown(event: KeyboardEvent): void {
  if (event.key === 'Escape' && props.open) {
    emit('close')
  }
}

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      window.setTimeout(() => {
        closeBtnRef.value?.focus()
      }, 40)
    }
  },
)

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-sheet">
      <div
        v-if="open"
        class="modal-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="ieee-modal-title"
        @click.self="emit('close')"
      >
        <div class="modal-card">
          <header class="modal-head">
            <span class="modal-code">DOC · IEEE-PAPER · PRÓXIMAMENTE</span>
            <button
              ref="closeBtnRef"
              type="button"
              class="modal-close"
              aria-label="Cerrar ficha técnica del informe IEEE"
              @click="emit('close')"
            >
              <span>Cerrar</span>
              <UiIcon name="close" :size="14" />
            </button>
          </header>

          <div class="modal-body">
            <h2 id="ieee-modal-title">Ficha Técnica del Informe IEEE</h2>
            <p class="modal-lead">
              El artículo técnico completo en formato <code>IEEEtran</code> está en etapa de
              preparación en <code>docs/ieee-paper/</code>. El enlace público se habilitará en esta
              sección al cerrar la compilación final de métricas reales.
            </p>

            <div class="modal-sections">
              <div class="modal-row">
                <h3>Alcance metodológico</h3>
                <p>
                  Arquitectura de nodo solar costero con protección IP66, cálculo de índices ECA/INCA
                  y US EPA en backend Go e interpolación IDW (grilla 50 m, decaimiento 400–600 m).
                </p>
              </div>
              <div class="modal-row">
                <h3>Evidencia trazable requerida</h3>
                <p>
                  Latencia de interpolación IDW sobre grilla de 50 m, tamaño de payload JSON, consumo
                  de memoria en servicio Go, cobertura real de pruebas (<code>go test -cover ./...</code>)
                  y validación ERC/DRC del esquemático y PCB.
                </p>
              </div>
              <div class="modal-row">
                <h3>Integridad de datos</h3>
                <p>
                  Este portal evita publicar cifras preliminares o no verificadas: los parámetros aquí
                  mostrados en modo demostración están marcados como simulación hasta enlazar la
                  versión definitiva del informe.
                </p>
              </div>
            </div>

            <div class="modal-footer">
              <a
                href="https://gmatiasvc.vercel.app"
                target="_blank"
                rel="noopener noreferrer"
                class="raked-btn raked-btn-secondary"
              >
                <span>Sitio del autor <UiIcon name="external" :size="14" /></span>
              </a>
              <button type="button" class="raked-btn raked-btn-primary" @click="emit('close')">
                <span>Entendido</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 60;
  background: var(--modal-backdrop);
  display: grid;
  place-items: center;
  padding: 20px;
}

.modal-card {
  width: min(680px, 100%);
  max-height: 90svh;
  overflow-y: auto;
  background: var(--surface-ivory);
  border: 2px solid var(--rule-color);
  color: var(--ink-primary);
}

.modal-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border-bottom: 1.5px solid var(--rule-color);
}

.modal-code {
  font-family: var(--font-mono);
  font-size: 11.5px;
  letter-spacing: 0.06em;
}

.modal-close {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--accent-amber-fill);
  background: transparent;
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 5px 12px;
  cursor: pointer;
  transition:
    background-color 0.2s var(--ease-expo),
    color 0.2s var(--ease-expo);
}

.modal-close:hover {
  background: var(--accent-amber-fill);
  color: #181a1c;
}

.modal-body {
  padding: 24px 26px 28px;
}

h2 {
  font-family: var(--font-display);
  font-size: clamp(1.9rem, 3.2vw, 2.5rem);
  font-weight: 800;
  text-transform: uppercase;
  line-height: 1.05;
  margin-bottom: 12px;
}

.modal-lead {
  font-size: 15px;
  line-height: 1.55;
  color: var(--ink-secondary);
  margin-bottom: 20px;
}

.modal-lead code {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 2px 6px;
  background: var(--surface-ivory-2);
  border: 1px solid var(--rule-soft);
}

.modal-sections {
  border: 1px solid var(--rule-color);
  background: var(--surface-ivory-2);
  margin-bottom: 24px;
}

.modal-row {
  padding: 14px 16px;
  border-bottom: 1px solid var(--rule-soft);
}

.modal-row:last-child {
  border-bottom: none;
}

.modal-row h3 {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  color: var(--accent-vermilion);
  margin-bottom: 6px;
}

:root[data-theme='dark'] .modal-row h3 {
  color: var(--accent-amber);
}

.modal-row p {
  font-size: 14px;
  line-height: 1.5;
  color: var(--ink-primary);
}

.modal-row code {
  font-family: var(--font-mono);
  font-size: 12.5px;
}

.modal-footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
}
</style>
