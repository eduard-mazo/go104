import { defineStore } from 'pinia'
import { ref } from 'vue'
import { signalsAPI, type Signal } from '@/api/client'

export const useSignalsStore = defineStore('signals', () => {
  const signals = ref<Signal[]>([])
  const currentLineID = ref<number | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch(lineID: number) {
    currentLineID.value = lineID
    loading.value = true
    error.value = null
    try {
      signals.value = await signalsAPI.list(lineID)
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  function clear() {
    signals.value = []
    currentLineID.value = null
  }

  return { signals, currentLineID, loading, error, fetch, clear }
})
