<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'
import { allSignalsAPI, signalHistoryAPI, type Signal } from '@/api/client'
import { X, RefreshCw } from 'lucide-vue-next'

const props = defineProps<{
  signalId: number
}>()
const emit = defineEmits<{ close: [] }>()

// ── State ──────────────────────────────────────────────────────────────────
const allSignals  = ref<Signal[]>([])
const compareId   = ref<number | null>(null)
const range       = ref<'1h' | '6h' | '24h' | '7d'>('1h')
const loading     = ref(false)
const chartEl     = ref<HTMLDivElement>()
let   uplot: uPlot | null = null

const RANGES: Record<string, number> = { '1h': 3600, '6h': 21600, '24h': 86400, '7d': 604800 }

// ── Signal metadata ────────────────────────────────────────────────────────
const primarySignal  = computed(() => allSignals.value.find(s => s.id === props.signalId))
const compareSignal  = computed(() => compareId.value ? allSignals.value.find(s => s.id === compareId.value) : null)
const otherSignals   = computed(() => allSignals.value.filter(s => s.id !== props.signalId))

// ── Fetch & render ──────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  const now  = Date.now() / 1000
  const from = now - RANGES[range.value]

  try {
    const [primary, compare] = await Promise.all([
      signalHistoryAPI.query(props.signalId, from, now),
      compareId.value ? signalHistoryAPI.query(compareId.value, from, now) : Promise.resolve([]),
    ])

    await nextTick()
    renderChart(primary, compare)
  } finally {
    loading.value = false
  }
}

function renderChart(
  primary: { ts: number; value: number; quality: number }[],
  compare: { ts: number; value: number; quality: number }[],
) {
  if (!chartEl.value) return

  // Destroy previous instance
  uplot?.destroy()
  uplot = null

  // Merge timestamps from both series for the shared x-axis
  const allTs = Array.from(
    new Set([...primary.map(p => p.ts), ...compare.map(p => p.ts)])
  ).sort((a, b) => a - b)

  if (allTs.length === 0) {
    // Nothing to draw — show empty state via DOM
    return
  }

  // Build aligned value arrays (null for missing)
  const primaryMap = new Map(primary.map(p => [p.ts, p.value]))
  const compareMap = new Map(compare.map(p => [p.ts, p.value]))

  const primaryVals = allTs.map(t => primaryMap.get(t) ?? null)
  const compareVals = allTs.map(t => compareMap.get(t) ?? null)

  const hasCompare = compare.length > 0
  const priLabel   = primarySignal.value ? `${primarySignal.value.name} (${primarySignal.value.unit || '—'})` : `Signal ${props.signalId}`
  const cmpLabel   = compareSignal.value ? `${compareSignal.value.name} (${compareSignal.value.unit || '—'})` : ''

  const series: uPlot.Series[] = [
    {},  // x-axis placeholder
    {
      label:  priLabel,
      stroke: '#22d3ee',
      width:  2,
      points: { show: false },
    },
  ]
  if (hasCompare) {
    series.push({
      label:  cmpLabel,
      stroke: '#f59e0b',
      width:  2,
      points: { show: false },
      scale:  'cmp',
    })
  }

  const axes: uPlot.Axis[] = [
    {
      stroke:   '#475569',
      ticks:    { stroke: '#1e293b' },
      grid:     { stroke: '#1e293b' },
      values:   (_u, vals) => vals.map(v => v == null ? '' : new Date(v * 1000).toLocaleTimeString()),
    },
    {
      stroke: '#22d3ee',
      ticks:  { stroke: '#1e293b' },
      grid:   { stroke: '#1e293b' },
      size:   60,
    },
  ]
  if (hasCompare) {
    axes.push({
      side:   1,   // right
      scale:  'cmp',
      stroke: '#f59e0b',
      ticks:  { stroke: '#1e293b' },
      grid:   { show: false },
      size:   60,
    })
  }

  const scales: Record<string, uPlot.Scale> = {
    x:   { time: true },
    y:   {},
    cmp: {},
  }

  const w = chartEl.value.clientWidth  || 760
  const h = chartEl.value.clientHeight || 300

  const opts: uPlot.Options = {
    width:  w,
    height: h,
    padding: [8, hasCompare ? 60 : 12, 0, 0],
    cursor: {
      sync: { key: 'scada-history' },
    },
    series,
    axes,
    scales,
    plugins: [],
  }

  const data: uPlot.AlignedData = [
    new Float64Array(allTs) as unknown as number[],
    primaryVals as number[],
    ...(hasCompare ? [compareVals as number[]] : []),
  ]

  uplot = new uPlot(opts, data, chartEl.value)
}

// ── Resize observer ────────────────────────────────────────────────────────
let ro: ResizeObserver | null = null
onMounted(async () => {
  allSignals.value = await allSignalsAPI.list()
  await load()

  ro = new ResizeObserver(() => {
    if (uplot && chartEl.value) {
      uplot.setSize({ width: chartEl.value.clientWidth, height: chartEl.value.clientHeight })
    }
  })
  if (chartEl.value) ro.observe(chartEl.value)
})

onUnmounted(() => {
  ro?.disconnect()
  uplot?.destroy()
  uplot = null
})

// Reload when range or compare signal changes
watch([range, compareId], load)

// ── Keyboard close ──────────────────────────────────────────────────────────
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}
onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>

<template>
  <Teleport to="body">
    <!-- Backdrop -->
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm"
      @mousedown.self="emit('close')"
    >
      <!-- Modal panel -->
      <div class="relative flex flex-col bg-[#0d1117] border border-slate-700 rounded-xl shadow-2xl
                  w-[min(900px,95vw)] h-[min(540px,90vh)] overflow-hidden">

        <!-- ── Header ── -->
        <div class="flex items-center gap-3 px-5 py-3 border-b border-slate-800 flex-shrink-0">
          <span class="font-mono font-bold text-white text-sm truncate">
            Signal History
            <span class="text-cyan-400 ml-1">{{ primarySignal?.name ?? `#${signalId}` }}</span>
          </span>

          <!-- Range buttons -->
          <div class="flex gap-1 ml-4">
            <button
              v-for="r in ['1h','6h','24h','7d']" :key="r"
              class="px-2 py-0.5 rounded text-[10px] font-mono border transition-colors"
              :class="range === r
                ? 'border-cyan-600 text-cyan-400 bg-cyan-950/40'
                : 'border-slate-700 text-slate-400 hover:border-slate-500'"
              @click="range = r as any"
            >{{ r }}</button>
          </div>

          <!-- Refresh -->
          <button
            class="p-1 rounded text-slate-500 hover:text-slate-300 transition-colors ml-1"
            :class="loading ? 'animate-spin' : ''"
            @click="load"
          >
            <RefreshCw class="w-3.5 h-3.5" />
          </button>

          <div class="ml-auto flex items-center gap-3">
            <!-- Compare picker -->
            <div class="flex items-center gap-2">
              <span class="text-[10px] font-mono text-slate-500 uppercase tracking-widest">Compare</span>
              <select
                :value="compareId ?? ''"
                class="bg-[#0a0e14] border border-slate-700 rounded px-2 py-0.5 text-xs text-slate-200
                       font-mono focus:outline-none focus:border-amber-600 max-w-[200px]"
                @change="compareId = ($event.target as HTMLSelectElement).value
                  ? Number(($event.target as HTMLSelectElement).value)
                  : null"
              >
                <option value="">— none —</option>
                <option v-for="s in otherSignals" :key="s.id" :value="s.id">
                  {{ s.line_name }} / {{ s.name }}
                </option>
              </select>
            </div>

            <!-- Close -->
            <button
              class="p-1 rounded text-slate-500 hover:text-red-400 transition-colors"
              @click="emit('close')"
            >
              <X class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- ── Legend strip ── -->
        <div class="flex items-center gap-5 px-5 py-2 border-b border-slate-800/60 flex-shrink-0 text-[11px] font-mono">
          <span class="flex items-center gap-1.5">
            <span class="w-4 h-0.5 bg-cyan-400 inline-block rounded"></span>
            <span class="text-slate-300">{{ primarySignal?.name }}</span>
            <span class="text-slate-600">({{ primarySignal?.unit || '—' }})</span>
          </span>
          <span v-if="compareSignal" class="flex items-center gap-1.5">
            <span class="w-4 h-0.5 bg-amber-400 inline-block rounded"></span>
            <span class="text-slate-300">{{ compareSignal.name }}</span>
            <span class="text-slate-600">({{ compareSignal.unit || '—' }})</span>
          </span>
        </div>

        <!-- ── Chart area ── -->
        <div class="flex-1 relative overflow-hidden p-2">
          <!-- Loading overlay -->
          <div
            v-if="loading"
            class="absolute inset-0 flex items-center justify-center bg-[#0d1117]/80 z-10"
          >
            <span class="text-slate-500 font-mono text-sm animate-pulse">Loading…</span>
          </div>

          <!-- uPlot mount point -->
          <div ref="chartEl" class="w-full h-full" />

          <!-- Empty state -->
          <div
            v-if="!loading && !uplot"
            class="absolute inset-0 flex items-center justify-center text-slate-600 font-mono text-sm"
          >
            No data for this range
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* uPlot canvas fills its container */
:deep(.uplot) {
  width: 100% !important;
  height: 100% !important;
}
:deep(.uplot canvas) {
  display: block;
}
</style>
