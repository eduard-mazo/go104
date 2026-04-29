<script setup lang="ts">
import { computed, markRaw, onMounted, ref, shallowRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useScadaStore } from '@/stores/scada'
import { allSignalsAPI, type ScadaElement, type Signal } from '@/api/client'
import ContextMenu, { type MenuItem } from '@/components/scada/ContextMenu.vue'
import PropsPanel from '@/components/scada/PropsPanel.vue'
import ValveSVG from '@/components/scada/elements/ValveSVG.vue'
import PctBar from '@/components/scada/elements/PctBar.vue'
import { ArrowLeft, Eye, Save } from 'lucide-vue-next'

const GRID = 12

const route  = useRoute()
const router = useRouter()
const store  = useScadaStore()

const canvasRef  = ref<HTMLElement>()
const scrollRef  = ref<HTMLElement>()
const allSignals = ref<Signal[]>([])
const selectedId = ref<string | null>(null)
const ctxMenu    = ref<{ x: number; y: number; items: MenuItem[] } | null>(null)

const viewId = Number(route.params.id)

onMounted(async () => {
  await store.loadView(viewId)
  allSignals.value = await allSignalsAPI.list()
})

const view     = computed(() => store.activeView)
const elements = computed(() => view.value?.elements ?? [])
const selectedEl = computed(() => elements.value.find(e => e.id === selectedId.value) ?? null)

const savingLabel = computed(() => ({
  idle:   '',
  saving: '● Saving…',
  saved:  '✓ Saved',
  error:  '✕ Error',
}[store.saving]))

const savingColor = computed(() => ({
  idle:   'text-slate-600',
  saving: 'text-amber-400 animate-pulse',
  saved:  'text-green-400',
  error:  'text-red-400',
}[store.saving]))

// ── element component map ─────────────────────────────────────────────────
const ELEMENT_COMPONENTS: Record<string, any> = {
  valve:   markRaw(ValveSVG),
  pct_bar: markRaw(PctBar),
}

function elementComp(kind: string) {
  return ELEMENT_COMPONENTS[kind] ?? ValveSVG
}

// ── default configs ───────────────────────────────────────────────────────
function defaultConfig(kind: string) {
  if (kind === 'valve') return { color_on: '#22c55e', color_off: '#ef4444', color_fault: '#f59e0b', label: '' }
  if (kind === 'pct_bar') return { min: 0, max: 100, color_fill: '#22d3ee', color_bg: '#1e293b', unit: '%', label: '', vertical: true }
  return {}
}

function defaultSize(kind: string) {
  if (kind === 'valve')   return { w: 60, h: 40 }
  if (kind === 'pct_bar') return { w: 56, h: 120 }
  return { w: 60, h: 60 }
}

// ── palette drop ──────────────────────────────────────────────────────────
function addFromPalette(kind: string) {
  const sz = defaultSize(kind)
  const el: ScadaElement = {
    id:        crypto.randomUUID(),
    kind:      kind as any,
    x:         snapGrid(100 + Math.random() * 200),
    y:         snapGrid(80  + Math.random() * 100),
    rotation:  0,
    signal_id: null,
    config:    defaultConfig(kind) as any,
    ...sz,
  }
  store.addElement(el)
  selectedId.value = el.id
}

// ── drag-to-move ──────────────────────────────────────────────────────────
function snapGrid(v: number) { return Math.round(v / GRID) * GRID }

function startDrag(e: MouseEvent, el: ScadaElement) {
  e.preventDefault()
  selectedId.value = el.id
  ctxMenu.value = null

  const canvasRect = canvasRef.value!.getBoundingClientRect()
  const scroll     = scrollRef.value!

  const getPos = (me: MouseEvent) => ({
    x: me.clientX - canvasRect.left  + scroll.scrollLeft,
    y: me.clientY - canvasRect.top   + scroll.scrollTop,
  })

  const start  = getPos(e)
  const ox = start.x - el.x
  const oy = start.y - el.y

  const onMove = (me: MouseEvent) => {
    const p  = getPos(me)
    el.x = snapGrid(Math.max(0, p.x - ox))
    el.y = snapGrid(Math.max(0, p.y - oy))
  }
  const onUp = () => {
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mouseup', onUp)
    store.scheduleSave()
  }
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}

// ── resize (BR corner) ────────────────────────────────────────────────────
function startResize(e: MouseEvent, el: ScadaElement) {
  e.preventDefault()
  e.stopPropagation()
  const startX = e.clientX
  const startY = e.clientY
  const startW = el.w
  const startH = el.h

  const onMove = (me: MouseEvent) => {
    el.w = snapGrid(Math.max(GRID * 2, startW + me.clientX - startX))
    el.h = snapGrid(Math.max(GRID * 2, startH + me.clientY - startY))
  }
  const onUp = () => {
    window.removeEventListener('mousemove', onMove)
    window.removeEventListener('mouseup', onUp)
    store.scheduleSave()
  }
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
}

// ── right-click on element ────────────────────────────────────────────────
function onElementContextMenu(e: MouseEvent, el: ScadaElement) {
  e.preventDefault()
  e.stopPropagation()
  selectedId.value = el.id
  ctxMenu.value = {
    x: e.clientX, y: e.clientY,
    items: [
      { icon: '✎', label: 'Edit Properties', action: () => { selectedId.value = el.id } },
      { icon: '⤭', label: 'Duplicate',       action: () => { const nid = store.duplicateElement(el.id); if (nid) selectedId.value = nid } },
      { divider: true, label: '', action: () => {} },
      { icon: '✕', label: 'Delete',          action: () => { store.removeElement(el.id); selectedId.value = null }, danger: true },
    ],
  }
}

// ── right-click on canvas background ─────────────────────────────────────
function onCanvasContextMenu(e: MouseEvent) {
  e.preventDefault()
  ctxMenu.value = {
    x: e.clientX, y: e.clientY,
    items: [
      { icon: '⊕', label: 'Add Valve',   action: () => addFromPalette('valve')   },
      { icon: '⊕', label: 'Add % Bar',   action: () => addFromPalette('pct_bar') },
    ],
  }
}

function onCanvasMouseDown() {
  ctxMenu.value = null
}

function onUpdateElement(el: ScadaElement) {
  store.updateElement(el)
}

function onDeleteElement(id: string) {
  store.removeElement(id)
  selectedId.value = null
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

      <input
        :value="view.name"
        @change="view.name = ($event.target as HTMLInputElement).value; store.scheduleSave()"
        class="bg-transparent font-mono font-bold text-white text-sm border-none outline-none
               focus:bg-[#0a0e14] focus:px-2 rounded transition-all w-48"
      />

      <span class="text-[10px] font-mono ml-1" :class="savingColor">{{ savingLabel }}</span>

      <span class="ml-auto text-[10px] font-mono text-slate-600 uppercase tracking-widest">Design mode</span>

      <button
        class="flex items-center gap-1.5 px-3 py-1.5 rounded border border-cyan-800 text-cyan-400
               hover:bg-cyan-950/30 text-xs font-mono transition-colors"
        @click="router.push(`/scada/${viewId}/live`)"
      >
        <Eye class="w-3.5 h-3.5" /> Live view
      </button>
    </div>

    <div class="flex flex-1 overflow-hidden">
      <!-- ── Palette sidebar ── -->
      <aside class="w-40 flex-shrink-0 border-r border-slate-800 bg-[#0d1117] flex flex-col py-4 gap-2 px-3">
        <p class="text-[10px] font-mono uppercase tracking-widest text-slate-600 mb-1">Elements</p>

        <button
          class="group flex flex-col items-center gap-2 p-3 rounded border border-slate-800
                 hover:border-amber-700 hover:bg-amber-950/20 transition-colors text-xs text-slate-400
                 hover:text-amber-300 font-mono"
          @click="addFromPalette('valve')"
        >
          <!-- Mini valve preview -->
          <svg width="36" height="24" viewBox="0 0 36 24">
            <polygon points="0,0 18,12 0,24" fill="#374151" stroke="#475569" stroke-width="1"/>
            <polygon points="36,0 18,12 36,24" fill="#374151" stroke="#475569" stroke-width="1"/>
          </svg>
          Valve
        </button>

        <button
          class="group flex flex-col items-center gap-2 p-3 rounded border border-slate-800
                 hover:border-amber-700 hover:bg-amber-950/20 transition-colors text-xs text-slate-400
                 hover:text-amber-300 font-mono"
          @click="addFromPalette('pct_bar')"
        >
          <!-- Mini bar preview -->
          <svg width="20" height="36" viewBox="0 0 20 36">
            <rect x="4" y="2" width="12" height="32" fill="#1e293b" rx="1"/>
            <rect x="4" y="18" width="12" height="16" fill="#22d3ee" rx="1"/>
          </svg>
          % Bar
        </button>
      </aside>

      <!-- ── Canvas area ── -->
      <div ref="scrollRef" class="flex-1 overflow-auto p-8 bg-[#060a10]">
        <div
          ref="canvasRef"
          class="relative border border-slate-800 select-none"
          :style="{
            width:  view.width + 'px',
            height: view.height + 'px',
            background: '#0a0e14',
            backgroundImage: 'linear-gradient(to right,#1a2332 1px,transparent 1px),linear-gradient(to bottom,#1a2332 1px,transparent 1px)',
            backgroundSize: `${GRID}px ${GRID}px`,
          }"
          @mousedown.self="onCanvasMouseDown(); selectedId = null"
          @contextmenu.prevent="onCanvasContextMenu"
        >
          <!-- Elements -->
          <div
            v-for="el in elements"
            :key="el.id"
            class="absolute cursor-move"
            :style="{
              left:      el.x + 'px',
              top:       el.y + 'px',
              width:     el.w + 'px',
              height:    el.h + 'px',
              transform: `rotate(${el.rotation}deg)`,
              transformOrigin: 'center center',
            }"
            @mousedown.prevent="startDrag($event, el)"
            @contextmenu.prevent.stop="onElementContextMenu($event, el)"
          >
            <component
              :is="elementComp(el.kind)"
              :config="el.config"
              :signal="null"
              :selected="selectedId === el.id"
              :w="el.w"
              :h="el.h"
            />

            <!-- Resize handle — bottom-right corner -->
            <div
              v-if="selectedId === el.id"
              class="absolute bottom-0 right-0 w-3 h-3 cursor-se-resize
                     border-r-2 border-b-2 border-amber-500 opacity-80"
              @mousedown.prevent.stop="startResize($event, el)"
            />
          </div>
        </div>
      </div>

      <!-- ── Props panel ── -->
      <aside
        v-if="selectedEl"
        class="w-64 flex-shrink-0 border-l border-slate-800 bg-[#0d1117]"
      >
        <PropsPanel
          :element="selectedEl"
          :all-signals="allSignals"
          @update="onUpdateElement"
          @delete="onDeleteElement"
        />
      </aside>
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
