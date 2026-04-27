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
    case 'ACTIVE':       return 'badge-green'
    case 'CONNECTING':   return 'badge-yellow'
    case 'STOPPED':      return 'badge-slate'
    default:             return 'badge-red'
  }
}
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-white flex items-center gap-2">
          <Plug class="w-5 h-5 text-blue-400" /> Communication Lines
        </h1>
        <p class="text-sm text-slate-400 mt-0.5">Manage IEC104 master connections (each line is independent)</p>
      </div>
      <button class="btn-primary" @click="openAdd">
        <Plus class="w-4 h-4" /> Add Line
      </button>
    </div>

    <!-- Table -->
    <div class="card overflow-hidden">
      <div v-if="store.loading" class="px-4 py-10 text-center text-slate-400">Loading…</div>
      <div v-else-if="!store.lines.length" class="px-4 py-10 text-center text-slate-400">
        No lines configured. Click <b>Add Line</b> to get started.
      </div>
      <table v-else class="table-base">
        <thead>
          <tr>
            <th class="th">Name</th>
            <th class="th">Host : Port</th>
            <th class="th">CA</th>
            <th class="th">State</th>
            <th class="th">RX / TX</th>
            <th class="th">GI (s)</th>
            <th class="th">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="l in store.lines" :key="l.id" class="hover:bg-slate-750">
            <td class="td font-medium text-white">{{ l.name }}</td>
            <td class="td font-mono text-slate-300">{{ l.host }}:{{ l.port }}</td>
            <td class="td text-slate-300">{{ l.common_address }}</td>
            <td class="td">
              <span :class="stateBadge(l.state)">{{ l.state }}</span>
            </td>
            <td class="td font-mono text-xs text-slate-400">
              {{ l.rx_count }} / {{ l.tx_count }}
            </td>
            <td class="td text-slate-400">{{ l.gi_interval_s || '—' }}</td>
            <td class="td">
              <div class="flex items-center gap-1">
                <button class="btn-success" :disabled="!!busy[l.id] || l.state === 'ACTIVE' || l.state === 'CONNECTING'" @click="start(l)" title="Connect">
                  <Play class="w-3.5 h-3.5" />
                </button>
                <button class="btn-danger" :disabled="!!busy[l.id] || l.state === 'DISCONNECTED' || l.state === 'STOPPED'" @click="stop(l)" title="Disconnect">
                  <Square class="w-3.5 h-3.5" />
                </button>
                <button class="btn-secondary" :disabled="busy[l.id] === 'gi' || l.state !== 'ACTIVE'" @click="gi(l)" title="General Interrogation">
                  <RefreshCw class="w-3.5 h-3.5" :class="busy[l.id] === 'gi' && 'animate-spin'" />
                </button>
                <button class="btn-ghost" @click="openEdit(l)" title="Edit">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button class="btn-ghost text-red-400 hover:text-red-300" @click="remove(l)" title="Delete">
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
