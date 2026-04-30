<script setup lang="ts">
import { computed, markRaw, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useScadaStore } from '@/stores/scada'
import { allSignalsAPI, type ScadaElement, type ScadaLine, type Signal } from '@/api/client'
import { generateUUID } from '@/lib/utils'
import ContextMenu, { type MenuItem } from '@/components/scada/ContextMenu.vue'
import ConnectionLayer from '@/components/scada/ConnectionLayer.vue'
import HistoryModal from '@/components/scada/HistoryModal.vue'
import PropsPanel from '@/components/scada/PropsPanel.vue'
import ValveSVG        from '@/components/scada/elements/ValveSVG.vue'
import PctBar          from '@/components/scada/elements/PctBar.vue'
import CircuitBreaker  from '@/components/scada/elements/CircuitBreaker.vue'
import Motor           from '@/components/scada/elements/Motor.vue'
import Transformer     from '@/components/scada/elements/Transformer.vue'
import IndicatorLamp   from '@/components/scada/elements/IndicatorLamp.vue'
import BallValve       from '@/components/scada/elements/BallValve.vue'
import Pump            from '@/components/scada/elements/Pump.vue'
import Tank            from '@/components/scada/elements/Tank.vue'
import FlowMeter       from '@/components/scada/elements/FlowMeter.vue'
import PressureGauge   from '@/components/scada/elements/PressureGauge.vue'
import Compressor      from '@/components/scada/elements/Compressor.vue'
import ControlValve   from '@/components/scada/elements/ControlValve.vue'
import CheckValve     from '@/components/scada/elements/CheckValve.vue'
import HeatExchanger  from '@/components/scada/elements/HeatExchanger.vue'
import TempTransmitter from '@/components/scada/elements/TempTransmitter.vue'
import PipeSegment    from '@/components/scada/elements/PipeSegment.vue'
import { ArrowLeft, Eye } from 'lucide-vue-next'

const GRID = 12

const route  = useRoute()
const router = useRouter()
const store  = useScadaStore()

const canvasRef     = ref<HTMLElement>()
const scrollRef     = ref<HTMLElement>()
const allSignals    = ref<Signal[]>([])
const selectedId    = ref<string | null>(null)
const selectedLineId = ref<string | null>(null)
const ctxMenu       = ref<{ x: number; y: number; items: MenuItem[] } | null>(null)
const historySignal = ref<number | null>(null)

// ── Connect mode ──────────────────────────────────────────────────────────
const connectMode   = ref(false)
const pendingFrom   = ref<{ elId: string; pt: { x: number; y: number } } | null>(null)
const mouseCanvasPt = ref<{ x: number; y: number } | null>(null)
const connectLineStyle = ref<ScadaLine['style']>('pipe_water')

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
  valve:           markRaw(ValveSVG),
  pct_bar:         markRaw(PctBar),
  circuit_breaker: markRaw(CircuitBreaker),
  motor:           markRaw(Motor),
  transformer:     markRaw(Transformer),
  indicator_lamp:  markRaw(IndicatorLamp),
  ball_valve:      markRaw(BallValve),
  pump:            markRaw(Pump),
  tank:            markRaw(Tank),
  flow_meter:      markRaw(FlowMeter),
  pressure_gauge:  markRaw(PressureGauge),
  compressor:      markRaw(Compressor),
  control_valve:   markRaw(ControlValve),
  check_valve:     markRaw(CheckValve),
  heat_exchanger:  markRaw(HeatExchanger),
  temp_tx:         markRaw(TempTransmitter),
  pipe_segment:    markRaw(PipeSegment),
}

function elementComp(kind: string) {
  return ELEMENT_COMPONENTS[kind] ?? ValveSVG
}

const DIGITAL_DEFAULT = { color_on: '#22c55e', color_off: '#ef4444', color_fault: '#f59e0b', label: '' }
const GAUGE_DEFAULT   = { min: 0, max: 100, color_fill: '#22d3ee', color_alarm: '#ef4444', alarm_high: 0, unit: '', label: '' }

// ── default configs ───────────────────────────────────────────────────────
function defaultConfig(kind: string) {
  switch (kind) {
    case 'valve':           return { ...DIGITAL_DEFAULT }
    case 'pct_bar':         return { min: 0, max: 100, color_fill: '#22d3ee', color_bg: '#1e293b', unit: '%', label: '', vertical: true }
    case 'circuit_breaker': return { ...DIGITAL_DEFAULT, color_on: '#22c55e', color_off: '#64748b', label: '' }
    case 'motor':           return { ...DIGITAL_DEFAULT }
    case 'indicator_lamp':  return { ...DIGITAL_DEFAULT, color_off: '#1e293b', label: '' }
    case 'ball_valve':      return { ...DIGITAL_DEFAULT }
    case 'pump':            return { ...DIGITAL_DEFAULT }
    case 'compressor':      return { ...DIGITAL_DEFAULT }
    case 'control_valve':   return { ...DIGITAL_DEFAULT }
    case 'check_valve':     return { ...DIGITAL_DEFAULT, color_on: '#22d3ee', color_off: '#1e293b', label: '' }
    case 'transformer':     return { ...GAUGE_DEFAULT, unit: 'kVA' }
    case 'tank':            return { ...GAUGE_DEFAULT, unit: 'm³', alarm_high: 90 }
    case 'flow_meter':      return { ...GAUGE_DEFAULT, unit: 'm³/h' }
    case 'pressure_gauge':  return { ...GAUGE_DEFAULT, unit: 'bar', alarm_high: 80 }
    case 'heat_exchanger':  return { ...GAUGE_DEFAULT, unit: '°C', alarm_high: 80 }
    case 'temp_tx':         return { ...GAUGE_DEFAULT, unit: '°C', alarm_high: 80 }
    case 'pipe_segment':    return { style: 'process', color: '#64748b', label: '' }
    default: return {}
  }
}

function defaultSize(kind: string) {
  switch (kind) {
    case 'pct_bar':        return { w: 56, h: 120 }
    case 'tank':           return { w: 64, h: 140 }
    case 'transformer':    return { w: 80, h: 80 }
    case 'pressure_gauge': return { w: 72, h: 72 }
    case 'flow_meter':     return { w: 72, h: 56 }
    case 'circuit_breaker':return { w: 72, h: 40 }
    case 'control_valve':  return { w: 48, h: 48 }
    case 'check_valve':    return { w: 40, h: 32 }
    case 'heat_exchanger': return { w: 100, h: 60 }
    case 'temp_tx':        return { w: 56, h: 56 }
    case 'pipe_segment':   return { w: 120, h: 16 }
    default:               return { w: 60, h: 60 }
  }
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
  const items: MenuItem[] = [
    { icon: '✎', label: 'Edit Properties', action: () => { selectedId.value = el.id } },
    { icon: '⤭', label: 'Duplicate',       action: () => { const nid = store.duplicateElement(el.id); if (nid) selectedId.value = nid } },
  ]
  if (el.signal_id) {
    items.push({ divider: true, label: '', action: () => {} })
    items.push({ icon: '📈', label: 'View history…', action: () => { historySignal.value = el.signal_id } })
  }
  items.push({ divider: true, label: '', action: () => {} })
  items.push({ icon: '✕', label: 'Delete', action: () => { store.removeElement(el.id); selectedId.value = null }, danger: true })
  ctxMenu.value = { x: e.clientX, y: e.clientY, items }
}

// ── right-click on canvas background ─────────────────────────────────────
function onCanvasContextMenu(e: MouseEvent) {
  e.preventDefault()
  ctxMenu.value = {
    x: e.clientX, y: e.clientY,
    items: [
      { icon: '⊕', label: 'Gate valve',    action: () => addFromPalette('valve')           },
      { icon: '⊕', label: 'Ball valve',    action: () => addFromPalette('ball_valve')      },
      { icon: '⊕', label: 'Pump',          action: () => addFromPalette('pump')            },
      { icon: '⊕', label: 'Tank',          action: () => addFromPalette('tank')            },
      { icon: '⊕', label: '% Bar',         action: () => addFromPalette('pct_bar')         },
      { divider: true, label: '', action: () => {} },
      { icon: '⊕', label: 'Compressor',    action: () => addFromPalette('compressor')      },
      { divider: true, label: '', action: () => {} },
      { icon: '⊕', label: 'Breaker',       action: () => addFromPalette('circuit_breaker') },
      { icon: '⊕', label: 'Motor',         action: () => addFromPalette('motor')           },
      { icon: '⊕', label: 'Transformer',   action: () => addFromPalette('transformer')     },
      { icon: '⊕', label: 'Lamp',          action: () => addFromPalette('indicator_lamp')  },
      { divider: true, label: '', action: () => {} },
      { icon: '⊕', label: 'Pressure gauge',action: () => addFromPalette('pressure_gauge')  },
      { icon: '⊕', label: 'Flow meter',    action: () => addFromPalette('flow_meter')      },
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

// ── Port system ───────────────────────────────────────────────────────────
// Returns absolute canvas port positions for an element
interface Port { id: string; x: number; y: number }
const PORT_DEFS: Record<string, { id: string; rx: number; ry: number }[]> = {
  valve:           [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  ball_valve:      [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  circuit_breaker: [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  flow_meter:      [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  compressor:      [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  control_valve:   [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  check_valve:     [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  heat_exchanger:  [{ id:'sl', rx:0, ry:0.35 }, { id:'sr', rx:1, ry:0.35 }, { id:'tl', rx:0, ry:0.65 }, { id:'tr', rx:1, ry:0.65 }],
  temp_tx:         [{ id:'b', rx:0.5, ry:1 }],
  pipe_segment:    [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  motor:           [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }, { id:'t', rx:0.5, ry:0 }],
  pump:            [{ id:'in', rx:0, ry:0.5 }, { id:'out', rx:0.5, ry:0 }],
  tank:            [{ id:'t', rx:0.5, ry:0 }, { id:'b', rx:0.5, ry:1 }, { id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }],
  pct_bar:         [{ id:'t', rx:0.5, ry:0 }, { id:'b', rx:0.5, ry:1 }],
  transformer:     [{ id:'tl', rx:0.3, ry:0 }, { id:'tr', rx:0.7, ry:0 }, { id:'bl', rx:0.3, ry:1 }, { id:'br', rx:0.7, ry:1 }],
  indicator_lamp:  [{ id:'b', rx:0.5, ry:1 }],
  pressure_gauge:  [{ id:'b', rx:0.5, ry:1 }],
}

function portsFor(el: ScadaElement): Port[] {
  const defs = PORT_DEFS[el.kind] ?? [{ id:'l', rx:0, ry:0.5 }, { id:'r', rx:1, ry:0.5 }]
  return defs.map(d => ({
    id: d.id,
    x:  el.x + d.rx * el.w,
    y:  el.y + d.ry * el.h,
  }))
}

// ── Connect mode handlers ─────────────────────────────────────────────────
function toggleConnectMode() {
  connectMode.value = !connectMode.value
  pendingFrom.value = null
  if (connectMode.value) {
    selectedId.value = null
    selectedLineId.value = null
  }
}

function onPortClick(el: ScadaElement, port: Port) {
  if (!connectMode.value) return
  if (!pendingFrom.value) {
    // Start connection
    pendingFrom.value = { elId: el.id, pt: { x: port.x, y: port.y } }
  } else {
    // Complete connection — don't connect element to itself
    if (pendingFrom.value.elId === el.id) {
      pendingFrom.value = null
      return
    }
    const line: ScadaLine = {
      id:           generateUUID(),
      from_el:      pendingFrom.value.elId,
      to_el:        el.id,
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

// ESC cancels connect mode
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') cancelConnect()
  if (e.key === 'Delete' && selectedLineId.value) {
    store.removeLine(selectedLineId.value)
    selectedLineId.value = null
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

// ── palette definitions (mini SVG previews) ───────────────────────────────
interface PaletteItem { kind: string; label: string; pw: number; ph: number; vb: string; svg: string }

const PALETTE_FLUID: PaletteItem[] = [
  { kind: 'valve',         label: 'Gate valve',    pw: 30, ph: 20, vb: '0 0 30 20',
    svg: '<polygon points="0,0 15,10 0,20" fill="#374151" stroke="#475569" stroke-width="1"/><polygon points="30,0 15,10 30,20" fill="#374151" stroke="#475569" stroke-width="1"/>' },
  { kind: 'ball_valve',    label: 'Ball valve',    pw: 28, ph: 22, vb: '0 0 28 22',
    svg: '<circle cx="14" cy="11" r="9" fill="#374151" stroke="#475569" stroke-width="1"/><line x1="5" y1="11" x2="23" y2="11" stroke="#0a0e14" stroke-width="5" stroke-linecap="round"/>' },
  { kind: 'control_valve', label: 'Control valve', pw: 30, ph: 20, vb: '0 0 30 20',
    svg: '<polygon points="0,0 15,10 0,20" fill="#374151" stroke="#475569" stroke-width="1"/><polygon points="30,0 15,10 30,20" fill="#374151" stroke="#475569" stroke-width="1"/><circle cx="15" cy="10" r="2" fill="#475569"/><line x1="15" y1="0" x2="15" y2="-4" stroke="#475569" stroke-width="1.5"/><circle cx="15" cy="-10" r="5" fill="none" stroke="#475569" stroke-width="1"/>' },
  { kind: 'check_valve',   label: 'Check valve',   pw: 30, ph: 20, vb: '0 0 30 20',
    svg: '<polygon points="0,0 29,10 0,20" fill="#374151" stroke="#475569" stroke-width="1"/><line x1="29" y1="0" x2="29" y2="20" stroke="#475569" stroke-width="2.5"/>' },
  { kind: 'pump',          label: 'Pump',          pw: 24, ph: 24, vb: '0 0 24 24',
    svg: '<circle cx="12" cy="12" r="10" fill="#374151" stroke="#475569" stroke-width="1"/><polygon points="12,7 19,16.5 5,16.5" fill="none" stroke="white" stroke-width="1.2" stroke-linejoin="round" opacity="0.7"/>' },
  { kind: 'tank',          label: 'Tank',          pw: 18, ph: 32, vb: '0 0 18 32',
    svg: '<rect x="2" y="2" width="14" height="28" fill="#0f1923" rx="1" stroke="#475569" stroke-width="1"/><line x1="3" y1="20" x2="15" y2="20" stroke="#22d3ee" stroke-width="1.5"/>' },
  { kind: 'pct_bar',       label: '% Bar',         pw: 16, ph: 32, vb: '0 0 16 32',
    svg: '<rect x="3" y="2" width="10" height="28" fill="#1e293b" rx="1" stroke="#475569" stroke-width="1"/><rect x="3" y="16" width="10" height="14" fill="#22d3ee" rx="1"/>' },
  { kind: 'pipe_segment',  label: 'Pipe',          pw: 36, ph: 10, vb: '0 0 36 10',
    svg: '<line x1="0" y1="5" x2="36" y2="5" stroke="#64748b" stroke-width="4" stroke-linecap="round"/>' },
]

const PALETTE_GAS: PaletteItem[] = [
  { kind: 'compressor', label: 'Compressor', pw: 24, ph: 24, vb: '0 0 24 24',
    svg: '<circle cx="12" cy="12" r="10" fill="#374151" stroke="#475569" stroke-width="1"/><path d="M6 9 L10 12 L6 15 M14 9 L18 12 L14 15" fill="none" stroke="white" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round" opacity="0.8"/>' },
]

const PALETTE_ELEC: PaletteItem[] = [
  { kind: 'circuit_breaker', label: 'Breaker',     pw: 32, ph: 18, vb: '0 0 32 18',
    svg: '<line x1="0" y1="9" x2="9" y2="9" stroke="#475569" stroke-width="2"/><rect x="9" y="3" width="14" height="12" fill="#374151" stroke="#475569" stroke-width="1" rx="1"/><line x1="11" y1="9" x2="21" y2="9" stroke="white" stroke-width="1.5"/><line x1="23" y1="9" x2="32" y2="9" stroke="#475569" stroke-width="2"/>' },
  { kind: 'motor',           label: 'Motor',       pw: 24, ph: 24, vb: '0 0 24 24',
    svg: '<circle cx="12" cy="12" r="10" fill="#374151" stroke="#475569" stroke-width="1"/><text x="12" y="13" text-anchor="middle" dominant-baseline="middle" font-size="9" fill="white" font-family="monospace" font-weight="700">M</text>' },
  { kind: 'transformer',     label: 'Transformer', pw: 32, ph: 24, vb: '0 0 32 24',
    svg: '<circle cx="10" cy="12" r="8" fill="#0f172a" stroke="#475569" stroke-width="1"/><circle cx="22" cy="12" r="8" fill="#0f172a" stroke="#475569" stroke-width="1"/>' },
  { kind: 'indicator_lamp',  label: 'Lamp',        pw: 22, ph: 22, vb: '0 0 22 22',
    svg: '<circle cx="11" cy="11" r="9" fill="#22c55e" stroke="#475569" stroke-width="1" opacity="0.7"/><line x1="5" y1="5" x2="17" y2="17" stroke="white" stroke-width="1.5" stroke-linecap="round" opacity="0.7"/><line x1="17" y1="5" x2="5" y2="17" stroke="white" stroke-width="1.5" stroke-linecap="round" opacity="0.7"/>' },
]

const PALETTE_INST: PaletteItem[] = [
  { kind: 'pressure_gauge', label: 'Pressure TX',  pw: 24, ph: 24, vb: '0 0 24 24',
    svg: '<circle cx="12" cy="12" r="10" fill="#0f1923" stroke="#475569" stroke-width="1"/><line x1="2" y1="12" x2="22" y2="12" stroke="#475569" stroke-width="0.8" opacity="0.5"/><text x="12" y="10" text-anchor="middle" dominant-baseline="middle" font-size="5.5" fill="#22d3ee" font-family="monospace" font-weight="700">PT</text><text x="12" y="16" text-anchor="middle" dominant-baseline="middle" font-size="5" fill="#22d3ee" font-family="monospace">—</text>' },
  { kind: 'flow_meter',     label: 'Flow TX',      pw: 24, ph: 24, vb: '0 0 24 24',
    svg: '<circle cx="12" cy="12" r="10" fill="#0f1923" stroke="#475569" stroke-width="1"/><line x1="2" y1="12" x2="22" y2="12" stroke="#475569" stroke-width="0.8" opacity="0.5"/><text x="12" y="10" text-anchor="middle" dominant-baseline="middle" font-size="5.5" fill="#22d3ee" font-family="monospace" font-weight="700">FT</text><text x="12" y="16" text-anchor="middle" dominant-baseline="middle" font-size="5" fill="#22d3ee" font-family="monospace">—</text>' },
  { kind: 'temp_tx',        label: 'Temp TX',      pw: 24, ph: 24, vb: '0 0 24 24',
    svg: '<circle cx="12" cy="12" r="10" fill="#0f1923" stroke="#475569" stroke-width="1"/><line x1="2" y1="12" x2="22" y2="12" stroke="#475569" stroke-width="0.8" opacity="0.5"/><text x="12" y="10" text-anchor="middle" dominant-baseline="middle" font-size="5.5" fill="#22d3ee" font-family="monospace" font-weight="700">TT</text><text x="12" y="16" text-anchor="middle" dominant-baseline="middle" font-size="5" fill="#22d3ee" font-family="monospace">—</text><rect x="10" y="22" width="4" height="6" fill="#64748b" rx="1"/>' },
  { kind: 'heat_exchanger', label: 'Heat exch.',   pw: 36, ph: 22, vb: '0 0 36 22',
    svg: '<rect x="2" y="2" width="32" height="18" fill="#0f1923" stroke="#475569" stroke-width="1" rx="2"/><path d="M4 7 C 12 7 12 15 18 15 S 24 7 32 7" fill="none" stroke="#22d3ee" stroke-width="1.2" stroke-linecap="round"/>' },
]
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

      <!-- Connect mode style selector -->
      <div v-if="connectMode" class="flex items-center gap-1">
        <span class="text-[10px] font-mono text-blue-400">Connect:</span>
        <button v-for="s in ['pipe_water','pipe_gas','wire','cable']" :key="s"
          class="px-1.5 py-0.5 rounded text-[9px] font-mono border transition-colors"
          :class="connectLineStyle === s
            ? 'border-blue-500 text-blue-300 bg-blue-950/40'
            : 'border-slate-700 text-slate-500 hover:border-slate-500'"
          @click="connectLineStyle = s as any">
          {{ s.replace('_',' ') }}
        </button>
        <span class="text-[9px] text-slate-500 ml-1">ESC to exit</span>
      </div>

      <!-- Connect toggle button -->
      <button
        class="flex items-center gap-1.5 px-3 py-1.5 rounded border text-xs font-mono transition-colors"
        :class="connectMode
          ? 'border-blue-500 text-blue-300 bg-blue-950/30'
          : 'border-slate-700 text-slate-400 hover:border-slate-500 hover:text-white'"
        @click="toggleConnectMode"
      >
        <svg width="12" height="12" viewBox="0 0 12 12" fill="none">
          <circle cx="2" cy="6" r="1.5" :fill="connectMode ? '#93c5fd' : '#64748b'"/>
          <circle cx="10" cy="6" r="1.5" :fill="connectMode ? '#93c5fd' : '#64748b'"/>
          <line x1="3.5" y1="6" x2="8.5" y2="6" :stroke="connectMode ? '#93c5fd' : '#64748b'" stroke-width="1.5" stroke-dasharray="2 1"/>
        </svg>
        Connect
      </button>

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
      <aside class="w-36 flex-shrink-0 border-r border-slate-800 bg-[#0d1117] flex flex-col py-3 gap-1 px-2 overflow-y-auto">

        <!-- Fluid group -->
        <p class="text-[9px] font-mono uppercase tracking-widest text-slate-600 mt-1 mb-0.5 px-1">Fluid</p>
        <button v-for="item in PALETTE_FLUID" :key="item.kind"
          class="flex items-center gap-2 px-2 py-1.5 rounded border border-slate-800
                 hover:border-amber-700 hover:bg-amber-950/20 transition-colors text-[11px] text-slate-400
                 hover:text-amber-300 font-mono text-left"
          @click="addFromPalette(item.kind)">
          <svg :width="item.pw" :height="item.ph" :viewBox="item.vb" class="flex-shrink-0" v-html="item.svg"/>
          {{ item.label }}
        </button>

        <!-- Gas group -->
        <p class="text-[9px] font-mono uppercase tracking-widest text-slate-600 mt-2 mb-0.5 px-1">Gas</p>
        <button v-for="item in PALETTE_GAS" :key="item.kind"
          class="flex items-center gap-2 px-2 py-1.5 rounded border border-slate-800
                 hover:border-amber-700 hover:bg-amber-950/20 transition-colors text-[11px] text-slate-400
                 hover:text-amber-300 font-mono text-left"
          @click="addFromPalette(item.kind)">
          <svg :width="item.pw" :height="item.ph" :viewBox="item.vb" class="flex-shrink-0" v-html="item.svg"/>
          {{ item.label }}
        </button>

        <!-- Electrical group -->
        <p class="text-[9px] font-mono uppercase tracking-widest text-slate-600 mt-2 mb-0.5 px-1">Electrical</p>
        <button v-for="item in PALETTE_ELEC" :key="item.kind"
          class="flex items-center gap-2 px-2 py-1.5 rounded border border-slate-800
                 hover:border-amber-700 hover:bg-amber-950/20 transition-colors text-[11px] text-slate-400
                 hover:text-amber-300 font-mono text-left"
          @click="addFromPalette(item.kind)">
          <svg :width="item.pw" :height="item.ph" :viewBox="item.vb" class="flex-shrink-0" v-html="item.svg"/>
          {{ item.label }}
        </button>

        <!-- Instruments group -->
        <p class="text-[9px] font-mono uppercase tracking-widest text-slate-600 mt-2 mb-0.5 px-1">Instruments</p>
        <button v-for="item in PALETTE_INST" :key="item.kind"
          class="flex items-center gap-2 px-2 py-1.5 rounded border border-slate-800
                 hover:border-amber-700 hover:bg-amber-950/20 transition-colors text-[11px] text-slate-400
                 hover:text-amber-300 font-mono text-left"
          @click="addFromPalette(item.kind)">
          <svg :width="item.pw" :height="item.ph" :viewBox="item.vb" class="flex-shrink-0" v-html="item.svg"/>
          {{ item.label }}
        </button>
      </aside>

      <!-- ── Canvas area ── -->
      <div ref="scrollRef" class="flex-1 overflow-auto p-8 bg-[#060a10]">
        <div
          ref="canvasRef"
          class="relative border border-slate-800 select-none"
          :class="connectMode ? 'cursor-crosshair' : ''"
          :style="{
            width:  view.width + 'px',
            height: view.height + 'px',
            background: '#0a0e14',
            backgroundImage: 'linear-gradient(to right,#1a2332 1px,transparent 1px),linear-gradient(to bottom,#1a2332 1px,transparent 1px)',
            backgroundSize: `${GRID}px ${GRID}px`,
          }"
          @mousedown.self="onCanvasMouseDown(); selectedId = null; selectedLineId = null"
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
            @select-line="(id) => { selectedLineId = id; selectedId = null }"
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
              :selected="!connectMode && selectedId === el.id"
              :w="el.w"
              :h="el.h"
            />

            <!-- Resize handle -->
            <div
              v-if="!connectMode && selectedId === el.id"
              class="absolute bottom-0 right-0 w-3 h-3 cursor-se-resize
                     border-r-2 border-b-2 border-amber-500 opacity-80"
              @mousedown.prevent.stop="startResize($event, el)"
            />

            <!-- Port handles (connect mode) -->
            <template v-if="connectMode">
              <div
                v-for="port in portsFor(el)"
                :key="port.id"
                class="absolute z-20 w-3 h-3 rounded-full border-2 cursor-pointer
                       transition-all hover:scale-150"
                :class="pendingFrom?.elId === el.id
                  ? 'bg-amber-400 border-amber-200'
                  : 'bg-blue-500 border-blue-200'"
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

      <!-- ── Right panel: element props or line props ── -->
      <aside
        v-if="selectedEl || selectedLine"
        class="w-64 flex-shrink-0 border-l border-slate-800 bg-[#0d1117]"
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
        <div v-else-if="selectedLine" class="flex flex-col h-full text-xs text-slate-300 overflow-y-auto">
          <div class="px-4 py-3 border-b border-slate-800 font-mono uppercase tracking-widest text-[10px] text-slate-500">
            Connection
          </div>
          <div class="px-4 py-3 space-y-3 border-b border-slate-800">
            <div>
              <label class="block text-slate-500 mb-0.5">Label</label>
              <input :value="selectedLine.label"
                @input="onUpdateLine({ ...selectedLine, label: ($event.target as HTMLInputElement).value })"
                class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono"
                placeholder="optional tag" />
            </div>
            <div>
              <label class="block text-slate-500 mb-0.5">Style</label>
              <div class="grid grid-cols-2 gap-1">
                <button v-for="s in ['pipe_water','pipe_gas','wire','cable']" :key="s"
                  class="py-1 rounded border text-[10px] font-mono transition-colors"
                  :class="selectedLine.style === s
                    ? 'border-cyan-600 text-cyan-300 bg-cyan-950/30'
                    : 'border-slate-700 text-slate-500 hover:border-slate-500'"
                  @click="onUpdateLine({ ...selectedLine, style: s as ScadaLine['style'] })">
                  {{ s.replace('_',' ') }}
                </button>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="block text-slate-500 mb-0.5">Color</label>
                <input type="color" :value="selectedLine.color || '#22d3ee'"
                  @input="onUpdateLine({ ...selectedLine, color: ($event.target as HTMLInputElement).value })"
                  class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" />
              </div>
              <div>
                <label class="block text-slate-500 mb-0.5">Width</label>
                <input type="number" min="1" max="8" :value="selectedLine.stroke_width"
                  @change="onUpdateLine({ ...selectedLine, stroke_width: Number(($event.target as HTMLInputElement).value) })"
                  class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" />
              </div>
            </div>
          </div>
          <div class="px-4 py-3 mt-auto">
            <button
              class="w-full py-1.5 rounded border border-red-800 text-red-400 hover:bg-red-950/40 transition-colors text-xs"
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

  <div v-else class="flex items-center justify-center h-screen text-slate-600 font-mono">
    Loading…
  </div>
</template>
