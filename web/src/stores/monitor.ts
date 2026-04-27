import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import type { Datapoint, WSMessage } from '@/api/client'
import { useLinesStore } from './lines'
import { monitorAPI } from '@/api/client'

// Stale quality bit sentinel — not a real IEC104 quality byte, just our UI flag
export const QUALITY_STALE = 0x80

export const useMonitorStore = defineStore('monitor', () => {
  const datapoints  = reactive<Record<number, Datapoint>>({})
  const staleIds    = reactive<Set<number>>(new Set())       // signal IDs marked stale
  const giInProgress = reactive<Set<number>>(new Set())      // line IDs with GI running
  const wsConnected = ref(false)

  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    ws = new WebSocket(`${proto}//${location.host}/ws`)

    ws.onopen = () => { wsConnected.value = true }

    ws.onclose = () => {
      wsConnected.value = false
      reconnectTimer = setTimeout(connect, 3000)
    }

    ws.onmessage = (ev: MessageEvent) => {
      const msg: WSMessage = JSON.parse(ev.data)

      if (msg.type === 'datapoint') {
        datapoints[msg.signal_id] = msg
        staleIds.delete(msg.signal_id)

      } else if (msg.type === 'line_status') {
        useLinesStore().applyStatus(msg.line_id, msg.state, msg.rx_count, msg.tx_count)

      } else if (msg.type === 'signal_stale') {
        staleIds.add(msg.signal_id)

      } else if (msg.type === 'gi_start') {
        giInProgress.add(msg.line_id)

      } else if (msg.type === 'gi_complete') {
        giInProgress.delete(msg.line_id)
      }
    }
  }

  function disconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
    ws = null
  }

  async function loadSnapshot(lineID: number) {
    const dps = await monitorAPI.list(lineID)
    for (const dp of dps) {
      datapoints[dp.signal_id] = dp
    }
  }

  return { datapoints, staleIds, giInProgress, wsConnected, connect, disconnect, loadSnapshot }
})
