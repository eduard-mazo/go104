import { defineStore } from 'pinia'
import { ref } from 'vue'
import { linesAPI, type Line } from '@/api/client'

export const useLinesStore = defineStore('lines', () => {
  const lines = ref<Line[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch() {
    loading.value = true
    error.value = null
    try {
      lines.value = await linesAPI.list()
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  function applyStatus(lineID: number, state: string, rx: number, tx: number) {
    const l = lines.value.find(l => l.id === lineID)
    if (l) {
      l.state = state as Line['state']
      l.rx_count = rx
      l.tx_count = tx
    }
  }

  return { lines, loading, error, fetch, applyStatus }
})
