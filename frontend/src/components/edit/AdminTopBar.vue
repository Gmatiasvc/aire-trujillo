<script setup lang="ts">
import { RouterLink } from 'vue-router'
import UiIcon from '../UiIcon.vue'
import { useThemeStore } from '../../stores/theme'

const props = defineProps<{
  totalNodes: number
  activeNodes: number
}>()

const emit = defineEmits<{
  logout: []
}>()

const themeStore = useThemeStore()
</script>

<template>
  <header class="admin-topbar">
    <div class="topbar-brand-group">
      <div class="nav-links-group">
        <RouterLink to="/map" class="nav-chip" aria-label="Ir al mapa de monitoreo">
          <UiIcon name="map" :size="14" />
          <span>Mapa</span>
        </RouterLink>
        <RouterLink to="/" class="nav-chip" aria-label="Volver al inicio">
          <UiIcon name="arrow-left" :size="14" />
          <span class="nav-home-label">Inicio</span>
        </RouterLink>
      </div>

      <div class="brand-divider" aria-hidden="true"></div>

      <div class="brand-title-wrap">
        <h1 class="admin-heading">AIRE TRUJILLO</h1>
        <span class="node-count-pill tabular">
          <i class="status-dot is-active" aria-hidden="true"></i>
          <span>{{ props.activeNodes }}/{{ props.totalNodes }}</span>
          <span class="node-word">estaciones activas</span>
        </span>
      </div>
    </div>

    <div class="topbar-actions-group">
      <button
        type="button"
        class="tool-btn"
        :aria-label="
          themeStore.mode === 'light'
            ? 'Cambiar a modo oscuro cálido'
            : 'Cambiar a modo claro cálido'
        "
        @click="themeStore.toggle"
      >
        <UiIcon :name="themeStore.mode === 'light' ? 'moon' : 'sun'" :size="15" />
        <span class="btn-label">{{ themeStore.mode === 'light' ? 'Oscuro' : 'Claro' }}</span>
      </button>

      <button
        type="button"
        class="raked-logout-btn"
        aria-label="Cerrar sesión de administración"
        @click="emit('logout')"
      >
        <span>
          <UiIcon name="logout" :size="14" />
          <span>Cerrar sesión</span>
        </span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.admin-topbar {
  position: relative;
  z-index: 20;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 16px;
  min-height: 56px;
  padding: 8px 18px;
  background: var(--surface-carbon);
  color: var(--ink-on-carbon);
  border-bottom: 1.5px solid var(--rule-color);
}

.topbar-brand-group,
.topbar-actions-group,
.nav-links-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.nav-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--glass-slate-rule);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  text-decoration: none;
  transition:
    background-color 0.18s var(--ease-expo),
    border-color 0.18s var(--ease-expo),
    color 0.18s var(--ease-expo);
}

.nav-chip:hover {
  background: var(--surface-slate);
  border-color: var(--accent-amber-fill);
}

.brand-divider {
  width: 1px;
  height: 22px;
  background: var(--glass-slate-rule);
}

.brand-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.admin-heading {
  font-family: var(--font-display);
  font-size: 1.38rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  line-height: 1;
  white-space: nowrap;
  color: var(--ink-on-carbon);
}

.node-count-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border: 1px solid var(--glass-slate-rule);
  background: var(--surface-slate);
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
}

.status-dot {
  width: 7px;
  height: 7px;
  background: #22c55e;
}

.tool-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border: 1px solid var(--glass-slate-rule);
  background: transparent;
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  cursor: pointer;
  transition:
    background-color 0.18s var(--ease-expo),
    border-color 0.18s var(--ease-expo);
}

.tool-btn:hover {
  background: var(--surface-slate);
  border-color: var(--accent-amber-fill);
}

.raked-logout-btn {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  border: 1px solid var(--accent-vermilion);
  background: transparent;
  color: var(--ink-on-carbon);
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  cursor: pointer;
  transform: skewX(var(--rake-deg));
  transition:
    background-color 0.18s var(--ease-expo),
    color 0.18s var(--ease-expo);
}

.raked-logout-btn > span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transform: skewX(calc(var(--rake-deg) * -1));
}

.raked-logout-btn:hover {
  background: var(--accent-vermilion);
  color: var(--accent-vermilion-ink);
}

.nav-chip:focus-visible,
.tool-btn:focus-visible,
.raked-logout-btn:focus-visible {
  outline: 2px solid var(--focus-ring);
  outline-offset: 3px;
}

@media (max-width: 900px) {
  .node-word,
  .nav-home-label {
    display: none;
  }
}

@media (max-width: 600px) {
  .admin-topbar {
    padding: 8px 12px;
    gap: 8px;
  }

  .admin-heading {
    font-size: 1.15rem;
  }

  .btn-label {
    display: none;
  }

  .brand-divider {
    display: none;
  }
}
</style>
