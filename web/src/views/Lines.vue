<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { linesAPI, type Line } from '@/api/client'
import { useLinesStore } from '@/stores/lines'
import LineForm from '@/components/LineForm.vue'
import { Plus, Play, Square, RefreshCw, Pencil, Trash2, Plug } from 'lucide-vue-next'

const store = useLinesStore()
const showForm = ref(false)
const editing = ref<Partial<Line> | null>(null)
const busy = ref<Record<number, string>>({})

onMounted(() => store.fetch())

function openAdd() { editing.value = null; showForm.value = true }
function openEdit(l: Line) { editing.value = l; showForm.value = true }

async function start(l: Line) {
  busy.value[l.id] = 'start'
  try { await linesAPI.start(l.id); await store.fetch() }
  finally { delete busy.value[l.id] }
}
async function stop(l: Line) {
  busy.value[l.id] = 'stop'
  try { await linesAPI.stop(l.id); await store.fetch() }
  finally { delete busy.value[l.id] }
}
async function gi(l: Line) {
  busy.value[l.id] = 'gi'
  try { await linesAPI.gi(l.id) }
  finally { delete busy.value[l.id] }
}
async function remove(l: Line) {
  if (!confirm(`Delete line "${l.name}"?`)) return
  await linesAPI.remove(l.id)
  await store.fetch()
}

function stateBadge(state: Line['state']) {
  switch (state) {
    case 'ACTIVE':     return 'badge-ok'
    case 'CONNECTING': return 'badge-warn'
    case 'STOPPED':    return 'badge-muted'
    default:           return 'badge-fault'
  }
}
</script>

<template>
  <div class="p-4 sm:p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-extrabold flex items-center gap-2 tracking-tight">
          <Plug class="w-5 h-5 text-[color:var(--epm-citrico)]" />
          Communication Lines
        </h1>
        <p class="text-sm text-muted-foreground mt-0.5">IEC-104 / DNP3 master connections</p>
      </div>
      <button class="btn-primary shrink-0" @click="openAdd">
        <Plus class="w-4 h-4" /> <span class="hidden sm:inline">Add Line</span>
      </button>
    </div>

    <!-- Table (scrollable on mobile) -->
    <div class="card overflow-x-auto">
      <div v-if="store.loading" class="px-4 py-10 text-center text-muted-foreground text-sm">Loading…</div>
      <div v-else-if="!store.lines.length" class="px-4 py-10 text-center text-muted-foreground text-sm">
        No lines configured. Click <b>Add Line</b> to get started.
      </div>
      <table v-else class="table-base min-w-[760px]">
        <thead>
          <tr>
            <th class="th">Name</th>
            <th class="th">Protocol</th>
            <th class="th">Host : Port</th>
            <th class="th">CA / Addr</th>
            <th class="th">State</th>
            <th class="th">RX / TX</th>
            <th class="th">GI (s)</th>
            <th class="th">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="l in store.lines" :key="l.id" class="data-row">
            <td class="td font-semibold">{{ l.name }}</td>
            <td class="td">
              <span :class="l.protocol === 'dnp3' ? 'badge-warn' : 'badge-muted'">
                {{ l.protocol === 'dnp3' ? 'DNP3' : 'IEC-104' }}
              </span>
            </td>
            <td class="td font-mono text-sm text-muted-foreground">{{ l.host }}:{{ l.port }}</td>
            <td class="td text-muted-foreground tabular text-xs">
              {{ l.protocol === 'dnp3' ? `OS ${l.dnp3_outstation_addr} / M ${l.dnp3_master_addr}` : l.common_address }}
            </td>
            <td class="td">
              <span :class="stateBadge(l.state)">{{ l.state }}</span>
            </td>
            <td class="td tabular text-muted-foreground">
              {{ l.rx_count }} / {{ l.tx_count }}
            </td>
            <td class="td text-muted-foreground">{{ l.gi_interval_s || '—' }}</td>
            <td class="td">
              <div class="flex items-center gap-1">
                <button class="btn-success"
                  :disabled="!!busy[l.id] || l.state === 'ACTIVE' || l.state === 'CONNECTING'"
                  title="Connect" @click="start(l)">
                  <Play class="w-3.5 h-3.5" />
                </button>
                <button class="btn-danger"
                  :disabled="!!busy[l.id] || l.state === 'DISCONNECTED' || l.state === 'STOPPED'"
                  title="Disconnect" @click="stop(l)">
                  <Square class="w-3.5 h-3.5" />
                </button>
                <button class="btn-secondary"
                  :disabled="busy[l.id] === 'gi' || l.state !== 'ACTIVE'"
                  title="General Interrogation" @click="gi(l)">
                  <RefreshCw class="w-3.5 h-3.5" :class="busy[l.id] === 'gi' && 'animate-spin'" />
                </button>
                <button class="btn-ghost" title="Edit" @click="openEdit(l)">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button class="btn-ghost text-destructive hover:text-destructive"
                  title="Delete" @click="remove(l)">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <LineForm v-model:open="showForm" :line="editing" @saved="store.fetch()" />
  </div>
</template>
