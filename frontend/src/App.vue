<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useThemeStore } from './stores/theme'

// Initialize theme store at root so data-theme applies across all routes.
useThemeStore()
</script>

<template>
  <RouterView />
</template>
