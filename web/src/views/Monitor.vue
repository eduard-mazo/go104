<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { linesAPI, formatValue, typeIDName, type Line } from '@/api/client'
import { useMonitorStore } from '@/stores/monitor'
import { useSignalsStore } from '@/stores/signals'
import { Activity, Loader2 } from 'lucide-vue-next'

const monitor = useMonitorStore()
const sigStore = useSignalsStore()
const lines = ref<Line[]>([])
const selectedLine = ref<number | null>(null)
const filter = ref('')

onMounted(async () => {
  lines.value = await linesAPI.list()
  if (lines.value.length) selectedLine.value = lines.value[0].id
})

watch(selectedLine, async (v) => {
  if (!v) return
  await sigStore.fetch(v)
  await monitor.loadSnapshot(v)
})

const giActive = computed(() =>
  selectedLine.value != null && monitor.giInProgress.has(selectedLine.value)
)

const rows = computed(() => {
  return sigStore.signals
    .filter(s => !filter.value
      || s.name.toLowerCase().includes(filter.value.toLowerCase())
      || String(s.ioa).includes(filter.value))
    .map(s => {
      const dp    = monitor.datapoints[s.id]
      const stale = monitor.staleIds.has(s.id)
      return {
        signal: s,
        dp,
        stale,
        value:   dp ? formatValue(dp) : '—',
        quality: dp?.quality ?? 0x80,
        qualOK:  dp?.quality_ok ?? false,
      }
    })
})

function qualityClass(ok: boolean, q: number, stale: boolean) {
  if (stale)     return 'badge-red'
  if (q & 0x80)  return 'badge-red'
  if (q & 0x40)  return 'badge-yellow'
  if (ok)        return 'badge-green'
  return 'badge-red'
}

function qualityLabel(q: number, stale: boolean) {
  if (stale) return 'STALE'
  const flags: string[] = []
  if (q & 0x80) flags.push('IV')
  if (q & 0x40) flags.push('NT')
  if (q & 0x20) flags.push('SB')
  if (q & 0x10) flags.push('BL')
  return flags.length ? flags.join(' ') : 'OK'
}

function fmtTime(ts: string) {
  if (!ts) return '—'
  const d  = new Date(ts)
  const ms = String(d.getMilliseconds()).padStart(3, '0')
  return d.toLocaleTimeString('en-GB', { hour12: false }) + '.' + ms
}
</script>

<template>
  <div class="p-6 space-y-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-white flex items-center gap-2">
          <Activity class="w-5 h-5 text-blue-400" /> Live Monitor
        </h1>
        <p class="text-sm text-slate-400 mt-0.5">Real-time signal values via WebSocket</p>
      </div>
      <div class="flex items-center gap-3 text-xs">
        <span :class="monitor.wsConnected ? 'text-green-400' : 'text-red-400'">
          {{ monitor.wsConnected ? '● Live' : '○ Disconnected' }}
        </span>
      </div>
    </div>

    <!-- Controls -->
    <div class="flex items-center gap-3 flex-wrap">
      <select v-model.number="selectedLine" class="input-base w-64">
        <option v-for="l in lines" :key="l.id" :value="l.id">
          {{ l.name }} — {{ l.host }}:{{ l.port }}
        </option>
      </select>
      <input v-model="filter" class="input-base w-48" placeholder="Filter by name / IOA…" />
      <span class="text-xs text-slate-500">{{ rows.length }} signals</span>
    </div>

    <!-- GI-in-progress banner -->
    <div v-if="giActive"
      class="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-900/40 border border-blue-700 text-blue-300 text-sm">
      <Loader2 class="w-4 h-4 animate-spin" />
      General Interrogation in progress — refreshing all signal values…
    </div>

    <!-- Monitor table -->
    <div class="card overflow-x-auto">
      <div v-if="sigStore.loading" class="px-4 py-10 text-center text-slate-400">Loading…</div>
      <div v-else-if="!rows.length" class="px-4 py-10 text-center text-slate-400">
        No signals found. Map signals in the Signals view first.
      </div>
      <table v-else class="table-base min-w-[700px]">
        <thead>
          <tr>
            <th class="th">Signal Name</th>
            <th class="th">IOA</th>
            <th class="th">TypeID</th>
            <th class="th">Kind</th>
            <th class="th text-right">Value</th>
            <th class="th text-center">Quality</th>
            <th class="th">Timestamp</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in rows" :key="row.signal.id"
            class="hover:bg-slate-750 transition-colors"
            :class="[
              row.stale ? 'opacity-50' : '',
              row.dp && !row.qualOK && !row.stale ? 'bg-red-950/20' : ''
            ]"
          >
            <td class="td font-medium" :class="row.stale ? 'text-slate-400' : 'text-white'">
              {{ row.signal.name }}
            </td>
            <td class="td font-mono text-blue-300">{{ row.signal.ioa }}</td>
            <td class="td text-xs font-mono text-slate-400">{{ typeIDName(row.signal.type_id) }}</td>
            <td class="td">
              <span :class="row.signal.signal_type === 'digital' ? 'badge-yellow' : 'badge-slate'">
                {{ row.signal.signal_type }}
              </span>
            </td>
            <td class="td text-right font-mono font-semibold"
              :class="row.signal.signal_type === 'digital'
                ? (row.dp?.value ? 'text-green-400' : 'text-slate-400')
                : 'text-cyan-300'"
            >
              {{ row.value }}
            </td>
            <td class="td text-center">
              <span v-if="row.dp || row.stale"
                :class="qualityClass(row.qualOK, row.quality, row.stale)">
                {{ qualityLabel(row.quality, row.stale) }}
              </span>
              <span v-else class="badge-slate">NO DATA</span>
            </td>
            <td class="td font-mono text-xs text-slate-400">
              {{ row.dp ? fmtTime(row.dp.timestamp) : '—' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
