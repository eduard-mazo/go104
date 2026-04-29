<script setup lang="ts">
import { computed, markRaw, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useScadaStore } from '@/stores/scada'
import { useMonitorStore } from '@/stores/monitor'
import { allSignalsAPI, commandsAPI, type ScadaElement, type Signal, type ValveConfig } from '@/api/client'
import ContextMenu, { type MenuItem } from '@/components/scada/ContextMenu.vue'
import ValveSVG from '@/components/scada/elements/ValveSVG.vue'
import PctBar from '@/components/scada/elements/PctBar.vue'
import { ArrowLeft, Pencil } from 'lucide-vue-next'

const route   = useRoute()
const router  = useRouter()
const store   = useScadaStore()
const monitor = useMonitorStore()

const allSignals = ref<Signal[]>([])
const ctxMenu    = ref<{ x: number; y: number; items: MenuItem[] } | null>(null)
const setpoint   = ref(0)

const viewId = Number(route.params.id)

onMounted(async () => {
  await store.loadView(viewId)
  allSignals.value = await allSignalsAPI.list()
})

const view     = computed(() => store.activeView)
const elements = computed(() => view.value?.elements ?? [])

// signal_id → Signal lookup map
const signalMap = computed(() => {
  const m = new Map<number, Signal>()
  for (const s of allSignals.value) m.set(s.id, s)
  return m
})

const ELEMENT_COMPONENTS: Record<string, any> = {
  valve:   markRaw(ValveSVG),
  pct_bar: markRaw(PctBar),
}
function elementComp(kind: string) { return ELEMENT_COMPONENTS[kind] ?? ValveSVG }

function dpForEl(el: ScadaElement) {
  if (!el.signal_id) return null
  return monitor.datapoints[el.signal_id] ?? null
}

// ── right-click in live mode ──────────────────────────────────────────────
function onElementContextMenu(e: MouseEvent, el: ScadaElement) {
  e.preventDefault()
  if (!el.signal_id) return
  const sig = signalMap.value.get(el.signal_id)
  if (!sig) return
  const dp = monitor.datapoints[el.signal_id]

  const items: MenuItem[] = []

  if (el.kind === 'valve') {
    const cfg = el.config as ValveConfig
    // Digital command — C_SC_NA_1 (45) or C_DC_NA_1 (46)
    const typeId = (sig.type_id === 3 || sig.type_id === 31) ? 46 : 45
    items.push(
      { icon: '▶', label: `Open (value = 1)`, action: () => sendDigital(sig, typeId, 1) },
      { icon: '■', label: `Close (value = 0)`, action: () => sendDigital(sig, typeId, 0) },
    )
    if (cfg.label || sig.name) {
      items.unshift({ icon: '⬤', label: `${cfg.label || sig.name}  (${dp?.value != null ? (dp.value ? 'OPEN' : 'CLOSED') : '—'})`, action: () => {} })
      items.splice(1, 0, { divider: true, label: '', action: () => {} })
    }
  } else if (el.kind === 'pct_bar') {
    // Analog setpoint — C_SE_NC_1 (50)
    const cur = dp?.value ?? 0
    items.push({
      icon: '↕', label: `Set value… (now ${cur.toFixed(1)} ${sig.unit})`,
      action: () => promptSetpoint(sig),
    })
  }

  if (!items.length) return
  ctxMenu.value = { x: e.clientX, y: e.clientY, items }
}

function sendDigital(sig: Signal, typeId: number, value: number) {
  commandsAPI.send({ line_id: sig.line_id, ioa: sig.ioa, type_id: typeId, value, select: false })
}

function promptSetpoint(sig: Signal) {
  const raw = window.prompt(`Setpoint for ${sig.name} (${sig.unit}):`, '0')
  if (raw === null) return
  const v = parseFloat(raw)
  if (isNaN(v)) return
  commandsAPI.send({ line_id: sig.line_id, ioa: sig.ioa, type_id: 50, value: v, select: false })
}
</script>

<template>
  <div class="flex flex-col h-screen overflow-hidden bg-[#060a10]" v-if="view">
    <!-- ── Top toolbar ── -->
    <div class="flex items-center gap-3 px-4 py-2 border-b border-slate-800 bg-[#0d1117] flex-shrink-0">
      <button class="p-1.5 rounded hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
              @click="router.push('/scada')">
        <ArrowLeft class="w-4 h-4" />
      </button>

      <span class="font-mono font-bold text-white text-sm">{{ view.name }}</span>

      <!-- WS indicator -->
      <div class="flex items-center gap-1.5 ml-3">
        <span class="w-1.5 h-1.5 rounded-full"
              :class="monitor.wsConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'" />
        <span class="text-[10px] font-mono" :class="monitor.wsConnected ? 'text-green-400' : 'text-red-400'">
          {{ monitor.wsConnected ? 'Live' : 'Disconnected' }}
        </span>
      </div>

      <span class="ml-auto text-[10px] font-mono text-slate-600 uppercase tracking-widest">Live mode · right-click to command</span>

      <button
        class="flex items-center gap-1.5 px-3 py-1.5 rounded border border-amber-800 text-amber-400
               hover:bg-amber-950/30 text-xs font-mono transition-colors"
        @click="router.push(`/scada/${viewId}/design`)"
      >
        <Pencil class="w-3.5 h-3.5" /> Edit
      </button>
    </div>

    <!-- ── Canvas (scrollable, no drag) ── -->
    <div class="flex-1 overflow-auto p-8 bg-[#060a10]">
      <div
        class="relative border border-slate-800 select-none"
        :style="{
          width:  view.width + 'px',
          height: view.height + 'px',
          background: '#0a0e14',
        }"
        @contextmenu.prevent
      >
        <div
          v-for="el in elements"
          :key="el.id"
          class="absolute"
          :class="el.signal_id ? 'cursor-context-menu' : 'cursor-default'"
          :style="{
            left:      el.x + 'px',
            top:       el.y + 'px',
            width:     el.w + 'px',
            height:    el.h + 'px',
            transform: `rotate(${el.rotation}deg)`,
            transformOrigin: 'center center',
          }"
          @contextmenu.prevent.stop="onElementContextMenu($event, el)"
        >
          <component
            :is="elementComp(el.kind)"
            :config="el.config"
            :signal="dpForEl(el)"
            :selected="false"
            :w="el.w"
            :h="el.h"
          />
        </div>
      </div>
    </div>

    <!-- Context menu -->
    <ContextMenu
      v-if="ctxMenu"
      :x="ctxMenu.x"
      :y="ctxMenu.y"
      :items="ctxMenu.items"
      @close="ctxMenu = null"
    />
  </div>

  <div v-else class="flex items-center justify-center h-screen text-slate-600 font-mono">
    Loading…
  </div>
</template>
