<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useScadaStore } from '@/stores/scada'
import { allSignalsAPI, type ScadaElement, type ScadaLine, type Signal } from '@/api/client'
import { generateUUID } from '@/lib/utils'
import { scadaRegistry, type ElementGroup } from '@/scada'
import ContextMenu, { type MenuItem } from '@/components/scada/ContextMenu.vue'
import ConnectionLayer from '@/components/scada/ConnectionLayer.vue'
import HistoryModal from '@/components/scada/HistoryModal.vue'
import PropsPanel from '@/components/scada/PropsPanel.vue'
import { ArrowLeft, Eye } from 'lucide-vue-next'

const GRID = 12

const route  = useRoute()
const router = useRouter()
const store  = useScadaStore()

const canvasRef     = ref<HTMLElement>()
const scrollRef     = ref<HTMLElement>()
const allSignals    = ref<Signal[]>([])
const selectedIds   = ref<Set<string>>(new Set())
const selectedLineId = ref<string | null>(null)
const ctxMenu       = ref<{ x: number; y: number; items: MenuItem[] } | null>(null)
const historySignal = ref<number | null>(null)

// ── Selection helpers ─────────────────────────────────────────────────────
function isSelected(id: string): boolean {
  return selectedIds.value.has(id)
}
function selectOnly(id: string) {
  selectedIds.value = new Set([id])
}
function toggleSelect(id: string) {
  const s = new Set(selectedIds.value)
  if (s.has(id)) s.delete(id); else s.add(id)
  selectedIds.value = s
}
function clearSelection() {
  if (selectedIds.value.size) selectedIds.value = new Set()
}

// ── Connect mode ──────────────────────────────────────────────────────────
const connectMode   = ref(false)
const pendingFrom   = ref<{ elId: string; portId: string; pt: { x: number; y: number } } | null>(null)
const mouseCanvasPt = ref<{ x: number; y: number } | null>(null)
const connectLineStyle = ref<ScadaLine['style']>('pipe_water')

const viewId = Number(route.params.id)

onMounted(async () => {
  await store.loadView(viewId)
  allSignals.value = await allSignalsAPI.list()
})

const view     = computed(() => store.activeView)
const elements = computed(() => view.value?.elements ?? [])
// Props panel only when exactly one element is selected.
const selectedEl = computed(() => {
  if (selectedIds.value.size !== 1) return null
  const id = [...selectedIds.value][0]
  return elements.value.find(e => e.id === id) ?? null
})

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

// ── registry-driven lookups ───────────────────────────────────────────────
function elementComp(kind: string) {
  return scadaRegistry.component(kind)
}
function defaultConfig(kind: string) {
  return scadaRegistry.get(kind)?.defaultConfig() ?? {}
}
function defaultSize(kind: string) {
  return scadaRegistry.get(kind)?.defaultSize() ?? { w: 60, h: 60 }
}

// ── palette drop ──────────────────────────────────────────────────────────
function addFromPalette(kind: string) {
  const sz = defaultSize(kind)
  const el: ScadaElement = {
    id:        generateUUID(),
    kind:      kind as any,
    x:         snapGrid(100 + Math.random() * 200),
    y:         snapGrid(80  + Math.random() * 100),
    rotation:  0,
    signal_id: null,
    config:    defaultConfig(kind) as any,
    ...sz,
  }
  store.addElement(el)
  selectOnly(el.id)
}

// ── drag-to-move ──────────────────────────────────────────────────────────
function snapGrid(v: number) { return Math.round(v / GRID) * GRID }

function startDrag(e: MouseEvent, el: ScadaElement) {
  e.preventDefault()
  ctxMenu.value = null

  // Shift / Ctrl / Cmd toggles selection — no drag.
  if (e.shiftKey || e.ctrlKey || e.metaKey) {
    toggleSelect(el.id)
    return
  }

  // Click on unselected element → make it the sole selection.
  // Click on already-selected element → keep current set and drag it as a group.
  if (!isSelected(el.id)) selectOnly(el.id)

  const canvasRect = canvasRef.value!.getBoundingClientRect()
  const scroll     = scrollRef.value!
  const getPos = (me: MouseEvent) => ({
    x: me.clientX - canvasRect.left + scroll.scrollLeft,
    y: me.clientY - canvasRect.top  + scroll.scrollTop,
  })

  const start = getPos(e)
  const dragSet = elements.value.filter(x => isSelected(x.id))
  const initial = new Map(dragSet.map(x => [x.id, { x: x.x, y: x.y }]))

  const onMove = (me: MouseEvent) => {
    const p  = getPos(me)
    const dx = p.x - start.x
    const dy = p.y - start.y
    for (const x of dragSet) {
      const sp = initial.get(x.id)!
      x.x = snapGrid(Math.max(0, sp.x + dx))
      x.y = snapGrid(Math.max(0, sp.y + dy))
    }
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
  if (!isSelected(el.id)) selectOnly(el.id)
  const items: MenuItem[] = [
    { icon: '✎', label: 'Edit Properties', action: () => { selectOnly(el.id) } },
    { icon: '⤭', label: 'Duplicate',       action: () => { const nid = store.duplicateElement(el.id); if (nid) selectOnly(nid) } },
  ]
  if (el.signal_id) {
    items.push({ divider: true, label: '', action: () => {} })
    items.push({ icon: '📈', label: 'View history…', action: () => { historySignal.value = el.signal_id } })
  }
  items.push({ divider: true, label: '', action: () => {} })
  const count = selectedIds.value.size
  const delLabel = count > 1 ? `Delete ${count} elements` : 'Delete'
  items.push({ icon: '✕', label: delLabel, action: () => deleteSelectedElements(), danger: true })
  ctxMenu.value = { x: e.clientX, y: e.clientY, items }
}

// ── right-click on canvas background ─────────────────────────────────────
function onCanvasContextMenu(e: MouseEvent) {
  e.preventDefault()
  const items: MenuItem[] = []
  for (const g of paletteGroups.value) {
    if (items.length) items.push({ divider: true, label: '', action: () => {} })
    for (const it of g.items) {
      items.push({ icon: '⊕', label: it.label, action: () => addFromPalette(it.kind) })
    }
  }
  ctxMenu.value = { x: e.clientX, y: e.clientY, items }
}

function onCanvasMouseDown() {
  ctxMenu.value = null
}

function onUpdateElement(el: ScadaElement) {
  store.updateElement(el)
}

function onDeleteElement(id: string) {
  store.removeElement(id)
  const s = new Set(selectedIds.value)
  s.delete(id)
  selectedIds.value = s
}

// Bulk delete every element currently selected.
function deleteSelectedElements() {
  if (!selectedIds.value.size) return
  for (const id of [...selectedIds.value]) store.removeElement(id)
  clearSelection()
}

// ── Port system ───────────────────────────────────────────────────────────
interface Port { id: string; x: number; y: number }

function portsFor(el: ScadaElement): Port[] {
  return scadaRegistry.portsFor(el)
}

// ── Connect mode handlers ─────────────────────────────────────────────────
function toggleConnectMode() {
  connectMode.value = !connectMode.value
  pendingFrom.value = null
  if (connectMode.value) {
    clearSelection()
    selectedLineId.value = null
  }
}

function onPortClick(el: ScadaElement, port: Port) {
  if (!connectMode.value) return
  if (!pendingFrom.value) {
    pendingFrom.value = { elId: el.id, portId: port.id, pt: { x: port.x, y: port.y } }
  } else {
    if (pendingFrom.value.elId === el.id) {
      pendingFrom.value = null
      return
    }
    const line: ScadaLine = {
      id:           generateUUID(),
      from_el:      pendingFrom.value.elId,
      to_el:        el.id,
      from_port:    pendingFrom.value.portId,
      to_port:      port.id,
      from_pt:      pendingFrom.value.pt,
      to_pt:        { x: port.x, y: port.y },
      style:        connectLineStyle.value,
      color:        '',
      stroke_width: 2,
      label:        '',
    }
    store.addLine(line)
    pendingFrom.value = null
  }
}

function onCanvasMouseMove(e: MouseEvent) {
  if (!connectMode.value || !pendingFrom.value) { mouseCanvasPt.value = null; return }
  const rect = canvasRef.value!.getBoundingClientRect()
  const scroll = scrollRef.value!
  mouseCanvasPt.value = {
    x: e.clientX - rect.left  + scroll.scrollLeft,
    y: e.clientY - rect.top   + scroll.scrollTop,
  }
}

function cancelConnect() {
  pendingFrom.value = null
  connectMode.value = false
  mouseCanvasPt.value = null
}

// Keyboard: ignore when focus is in a form field so typing in inputs still works.
function isEditingField(target: EventTarget | null): boolean {
  if (!(target instanceof HTMLElement)) return false
  const tag = target.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable
}

function nudgeSelection(dx: number, dy: number) {
  if (!selectedIds.value.size) return
  for (const el of elements.value) {
    if (!isSelected(el.id)) continue
    el.x = Math.max(0, el.x + dx)
    el.y = Math.max(0, el.y + dy)
  }
  store.scheduleSave()
}

function onKeydown(e: KeyboardEvent) {
  if (isEditingField(e.target)) return

  if (e.key === 'Escape') {
    cancelConnect()
    clearSelection()
    selectedLineId.value = null
    return
  }

  // Delete / Backspace → remove selected elements + selected line.
  if (e.key === 'Delete' || e.key === 'Backspace') {
    let handled = false
    if (selectedIds.value.size) { deleteSelectedElements(); handled = true }
    if (selectedLineId.value)   { store.removeLine(selectedLineId.value); selectedLineId.value = null; handled = true }
    if (handled) e.preventDefault()
    return
  }

  // Arrow keys → nudge. Shift = fine (1px), default = grid step.
  const step = e.shiftKey ? 1 : GRID
  let dx = 0, dy = 0
  if (e.key === 'ArrowLeft')  dx = -step
  if (e.key === 'ArrowRight') dx =  step
  if (e.key === 'ArrowUp')    dy = -step
  if (e.key === 'ArrowDown')  dy =  step
  if (dx || dy) {
    e.preventDefault()
    nudgeSelection(dx, dy)
    return
  }

  // Ctrl/Cmd+A → select all elements.
  if ((e.ctrlKey || e.metaKey) && e.key === 'a') {
    e.preventDefault()
    selectedIds.value = new Set(elements.value.map(x => x.id))
  }
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))

// ── Line right-click ──────────────────────────────────────────────────────
function onLineContextMenu(e: MouseEvent, lineId: string) {
  e.preventDefault()
  e.stopPropagation()
  selectedLineId.value = lineId
  ctxMenu.value = {
    x: e.clientX, y: e.clientY,
    items: [
      { icon: '✕', label: 'Delete connection', action: () => { store.removeLine(lineId); selectedLineId.value = null }, danger: true },
    ],
  }
}

const selectedLine = computed(() =>
  (view.value?.lines ?? []).find(l => l.id === selectedLineId.value) ?? null
)

function onUpdateLine(l: ScadaLine) {
  store.updateLine(l)
}

// ── palette groups (registry-driven) ──────────────────────────────────────
// Display order — groups missing from the registry are simply skipped.
const PALETTE_ORDER: ElementGroup[] = ['fluid', 'gas', 'electrical', 'instrument', 'display', 'custom']
const GROUP_LABELS: Record<ElementGroup, string> = {
  fluid:      'Fluid',
  gas:        'Gas',
  electrical: 'Electrical',
  instrument: 'Instruments',
  display:    'Display',
  custom:     'Custom',
}

const paletteGroups = computed(() =>
  PALETTE_ORDER
    .map(g => ({ group: g, label: GROUP_LABELS[g], items: scadaRegistry.byGroup(g) }))
    .filter(g => g.items.length > 0)
)
</script>

<template>
  <!-- Dark operator canvas — EPM chrome wraps the dark P&ID area -->
  <div class="flex flex-col h-screen overflow-hidden bg-[#060a10]" v-if="view">

    <!-- ── Top toolbar (EPM brand chrome) ── -->
    <div class="flex items-center gap-2 px-3 py-2 border-b border-sidebar-border bg-sidebar flex-shrink-0">
      <button
        class="p-1.5 rounded hover:bg-sidebar-accent text-sidebar-foreground/60 hover:text-sidebar-foreground transition-colors"
        @click="router.push('/scada')">
        <ArrowLeft class="w-4 h-4" />
      </button>

      <div class="w-px h-5 bg-sidebar-border mx-1" />

      <input
        :value="view.name"
        @change="view.name = ($event.target as HTMLInputElement).value; store.scheduleSave()"
        class="bg-transparent font-mono font-bold text-sidebar-foreground text-sm
               border-none outline-none focus:bg-sidebar-accent focus:px-2 rounded
               transition-all w-44"
      />

      <span class="text-[10px] font-mono ml-1 min-w-[60px]" :class="{
        'text-[color:var(--epm-citrico)] animate-pulse': store.saving === 'saving',
        'text-[color:var(--signal-ok)]':                 store.saving === 'saved',
        'text-[color:var(--signal-fault)]':              store.saving === 'error',
        'opacity-0':                                     store.saving === 'idle',
      }">{{ savingLabel }}</span>

      <span class="ml-auto text-[9px] font-mono text-sidebar-foreground/30 uppercase tracking-widest hidden sm:block">
        Design
      </span>

      <!-- Connect style selector (only in connect mode) -->
      <div v-if="connectMode" class="flex items-center gap-1">
        <span class="text-[9px] font-mono text-[color:var(--epm-citrico)] opacity-70">pipe:</span>
        <button v-for="s in ['pipe_water','pipe_gas','wire','cable']" :key="s"
          class="px-1.5 py-0.5 rounded text-[9px] font-mono border transition-colors"
          :class="connectLineStyle === s
            ? 'border-[color:var(--epm-citrico)] text-[color:var(--epm-citrico)] bg-[color:var(--epm-citrico)]/10'
            : 'border-sidebar-border text-sidebar-foreground/40 hover:border-sidebar-foreground/40'"
          @click="connectLineStyle = s as any">
          {{ s.replace('_',' ') }}
        </button>
        <span class="text-[9px] text-sidebar-foreground/30 ml-1">ESC</span>
      </div>

      <!-- Connect toggle -->
      <button
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded border text-xs font-mono transition-colors"
        :class="connectMode
          ? 'border-[color:var(--epm-citrico)] text-[color:var(--epm-citrico)] bg-[color:var(--epm-citrico)]/10'
          : 'border-sidebar-border text-sidebar-foreground/50 hover:border-sidebar-foreground/50 hover:text-sidebar-foreground'"
        @click="toggleConnectMode"
      >
        <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
          <circle cx="2" cy="6" r="1.5" :fill="connectMode ? '#9FCF67' : '#6b7280'"/>
          <circle cx="10" cy="6" r="1.5" :fill="connectMode ? '#9FCF67' : '#6b7280'"/>
          <line x1="3.5" y1="6" x2="8.5" y2="6" :stroke="connectMode ? '#9FCF67' : '#6b7280'" stroke-width="1.5" stroke-dasharray="2 1"/>
        </svg>
        Connect
      </button>

      <button
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded border
               border-[color:var(--signal-wait)]/50 text-[color:var(--signal-wait)]
               hover:bg-[color:var(--signal-wait)]/10 text-xs font-mono transition-colors"
        @click="router.push(`/scada/${viewId}/live`)"
      >
        <Eye class="w-3.5 h-3.5" /> Live
      </button>
    </div>

    <div class="flex flex-1 overflow-hidden">
      <!-- ── Palette sidebar (EPM dark panel) ── -->
      <aside class="w-36 flex-shrink-0 border-r border-[#1e2d24] bg-[#0d1a11] flex flex-col py-3 gap-0.5 px-2 overflow-y-auto">
        <template v-for="(grp, gi) in paletteGroups" :key="grp.group">
          <p class="text-[8px] font-mono uppercase tracking-widest text-[color:var(--epm-citrico)]/50 px-1 font-bold"
             :class="gi === 0 ? 'mt-1 mb-1' : 'mt-3 mb-1'">{{ grp.label }}</p>
          <button v-for="d in grp.items" :key="d.kind"
            class="flex items-center gap-2 px-2 py-1.5 rounded border border-transparent
                   hover:border-[color:var(--epm-citrico)]/30 hover:bg-[color:var(--epm-citrico)]/5
                   transition-colors text-[11px] text-[#6b8a70] hover:text-[color:var(--epm-citrico)]
                   font-mono text-left"
            @click="addFromPalette(d.kind)">
            <svg :width="d.palette.pw" :height="d.palette.ph" :viewBox="d.palette.vb"
                 class="flex-shrink-0" v-html="d.palette.svg"/>
            {{ d.label }}
          </button>
        </template>
      </aside>

      <!-- ── Canvas area ── -->
      <div ref="scrollRef" class="flex-1 overflow-auto p-8 bg-[#060a10]">
        <div
          ref="canvasRef"
          class="relative select-none"
          :class="connectMode ? 'cursor-crosshair' : ''"
          :style="{
            width:  view.width + 'px',
            height: view.height + 'px',
            background: '#0a0e14',
            border: '1px solid rgba(159,207,103,0.08)',
            backgroundImage: 'linear-gradient(to right,#142018 1px,transparent 1px),linear-gradient(to bottom,#142018 1px,transparent 1px)',
            backgroundSize: `${GRID}px ${GRID}px`,
          }"
          @mousedown.self="onCanvasMouseDown(); clearSelection(); selectedLineId = null"
          @mousemove="onCanvasMouseMove"
          @contextmenu.prevent="connectMode ? null : onCanvasContextMenu($event)"
        >
          <!-- Connection layer (behind elements) -->
          <ConnectionLayer
            :lines="view.lines ?? []"
            :elements="elements"
            :width="view.width"
            :height="view.height"
            :selected-line-id="selectedLineId"
            :design="true"
            :preview-from="pendingFrom?.pt ?? null"
            :preview-to="mouseCanvasPt"
            @select-line="(id) => { selectedLineId = id; clearSelection() }"
          />

          <!-- Elements -->
          <div
            v-for="el in elements"
            :key="el.id"
            class="absolute"
            :class="connectMode ? 'cursor-crosshair' : 'cursor-move'"
            :style="{
              left:      el.x + 'px',
              top:       el.y + 'px',
              width:     el.w + 'px',
              height:    el.h + 'px',
              transform: `rotate(${el.rotation}deg)`,
              transformOrigin: 'center center',
            }"
            @mousedown.prevent="connectMode ? null : startDrag($event, el)"
            @contextmenu.prevent.stop="connectMode ? null : onElementContextMenu($event, el)"
          >
            <component
              :is="elementComp(el.kind)"
              :config="el.config"
              :signal="null"
              :selected="!connectMode && isSelected(el.id)"
              :w="el.w"
              :h="el.h"
              :rotation="el.rotation"
            />

            <!-- Resize handle -->
            <div
              v-if="!connectMode && selectedIds.size === 1 && isSelected(el.id)"
              class="absolute bottom-0 right-0 w-3 h-3 cursor-se-resize
                     border-r-2 border-b-2 border-[color:var(--epm-citrico)] opacity-80"
              @mousedown.prevent.stop="startResize($event, el)"
            />

            <!-- Port handles (connect mode) -->
            <template v-if="connectMode">
              <div
                v-for="port in portsFor(el)"
                :key="port.id"
                class="absolute z-20 w-3 h-3 rounded-full border-2 cursor-pointer transition-all hover:scale-150"
                :class="pendingFrom?.elId === el.id
                  ? 'bg-[color:var(--epm-citrico)] border-[color:var(--epm-citrico-soft)]'
                  : 'bg-[color:var(--signal-wait)] border-white/40'"
                :style="{
                  left: (port.x - el.x - 6) + 'px',
                  top:  (port.y - el.y - 6) + 'px',
                }"
                @mousedown.stop.prevent="onPortClick(el, port)"
              />
            </template>
          </div>
        </div>
      </div>

      <!-- ── Right panel ── -->
      <aside
        v-if="selectedEl || selectedLine"
        class="w-64 flex-shrink-0 border-l border-[#1e2d24] bg-[#0d1a11]"
      >
        <!-- Element props -->
        <PropsPanel
          v-if="selectedEl"
          :element="selectedEl"
          :all-signals="allSignals"
          @update="onUpdateElement"
          @delete="onDeleteElement"
        />

        <!-- Line props -->
        <div v-else-if="selectedLine" class="flex flex-col h-full text-xs overflow-y-auto">
          <div class="px-4 py-3 border-b border-[#1e2d24] font-mono uppercase tracking-widest text-[9px]
                      text-[color:var(--epm-citrico)]/60 font-bold">
            Connection
          </div>
          <div class="px-4 py-3 space-y-3 border-b border-[#1e2d24]">
            <div>
              <label class="block text-[#6b8a70] mb-1 text-[10px] uppercase tracking-wider font-semibold">Label</label>
              <input :value="selectedLine.label"
                @input="onUpdateLine({ ...selectedLine, label: ($event.target as HTMLInputElement).value })"
                class="w-full bg-[#060a10] border border-[#1e2d24] rounded px-2 py-1 text-[#c8ddb0] font-mono
                       focus:outline-none focus:border-[color:var(--epm-citrico)]/50"
                placeholder="optional tag" />
            </div>
            <div>
              <label class="block text-[#6b8a70] mb-1 text-[10px] uppercase tracking-wider font-semibold">Style</label>
              <div class="grid grid-cols-2 gap-1">
                <button v-for="s in ['pipe_water','pipe_gas','wire','cable']" :key="s"
                  class="py-1 rounded border text-[9px] font-mono transition-colors"
                  :class="selectedLine.style === s
                    ? 'border-[color:var(--epm-citrico)]/60 text-[color:var(--epm-citrico)] bg-[color:var(--epm-citrico)]/10'
                    : 'border-[#1e2d24] text-[#6b8a70] hover:border-[#2e4030]'"
                  @click="onUpdateLine({ ...selectedLine, style: s as ScadaLine['style'] })">
                  {{ s.replace('_',' ') }}
                </button>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="block text-[#6b8a70] mb-1 text-[10px] uppercase tracking-wider font-semibold">Color</label>
                <input type="color" :value="selectedLine.color || '#9FCF67'"
                  @input="onUpdateLine({ ...selectedLine, color: ($event.target as HTMLInputElement).value })"
                  class="w-full h-7 cursor-pointer rounded border border-[#1e2d24] bg-transparent" />
              </div>
              <div>
                <label class="block text-[#6b8a70] mb-1 text-[10px] uppercase tracking-wider font-semibold">Width</label>
                <input type="number" min="1" max="8" :value="selectedLine.stroke_width"
                  @change="onUpdateLine({ ...selectedLine, stroke_width: Number(($event.target as HTMLInputElement).value) })"
                  class="w-full bg-[#060a10] border border-[#1e2d24] rounded px-2 py-1 text-[#c8ddb0] font-mono
                         focus:outline-none focus:border-[color:var(--epm-citrico)]/50" />
              </div>
            </div>
          </div>
          <div class="px-4 py-3 mt-auto">
            <button
              class="btn btn-danger w-full text-xs justify-center"
              @click="store.removeLine(selectedLine.id); selectedLineId = null">
              Delete connection
            </button>
          </div>
        </div>
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

    <!-- History modal -->
    <HistoryModal
      v-if="historySignal !== null"
      :signal-id="historySignal"
      @close="historySignal = null"
    />
  </div>

  <div v-else class="flex items-center justify-center h-screen text-muted-foreground font-mono text-sm">
    Loading…
  </div>
</template>
