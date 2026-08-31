import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const sidebarOpen = ref(true)
  const legendOpen = ref(true)
  const scanPanelOpen = ref(true)

  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  function toggleLegend() {
    legendOpen.value = !legendOpen.value
  }

  function toggleScanPanel() {
    scanPanelOpen.value = !scanPanelOpen.value
  }

  return {
    sidebarOpen,
    legendOpen,
    scanPanelOpen,
    toggleSidebar,
    toggleLegend,
    toggleScanPanel,
  }
})
