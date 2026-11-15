import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

export type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'aire-trujillo-theme'

function detectInitialTheme(): ThemeMode {
  if (typeof window === 'undefined') {
    return 'light'
  }
  const saved = window.localStorage.getItem(STORAGE_KEY)
  if (saved === 'light' || saved === 'dark') {
    return saved
  }
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>(detectInitialTheme())

  const applyTheme = (next: ThemeMode): void => {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', next)
    }
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, next)
    }
  }

  applyTheme(mode.value)

  watch(mode, (next) => {
    applyTheme(next)
  })

  function toggle(): void {
    mode.value = mode.value === 'light' ? 'dark' : 'light'
  }

  function setMode(next: ThemeMode): void {
    mode.value = next
  }

  return {
    mode,
    toggle,
    setMode,
  }
})
