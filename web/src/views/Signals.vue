<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { linesAPI, signalsAPI, typeIDName, type Line, type Signal } from '@/api/client'
import { useSignalsStore } from '@/stores/signals'
import SignalForm from '@/components/SignalForm.vue'
import { Plus, Pencil, Trash2, Radio } from 'lucide-vue-next'

const store = useSignalsStore()
const lines = ref<Line[]>([])
const selectedLine = ref<number | null>(null)
const showForm = ref(false)
const editing = ref<Partial<Signal> | null>(null)

onMounted(async () => {
  lines.value = await linesAPI.list()
  if (lines.value.length) selectedLine.value = lines.value[0].id
})

watch(selectedLine, (v) => { if (v) store.fetch(v) })

function openAdd() { editing.value = null; showForm.value = true }
function openEdit(s: Signal) { editing.value = s; showForm.value = true }

async function remove(s: Signal) {
  if (!confirm(`Delete signal "${s.name}"?`)) return
  await signalsAPI.remove(s.id)
  if (selectedLine.value) store.fetch(selectedLine.value)
}

function onSaved() {
  if (selectedLine.value) store.fetch(selectedLine.value)
}
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-white flex items-center gap-2">
          <Radio class="w-5 h-5 text-blue-400" /> Signal Mapping
        </h1>
        <p class="text-sm text-slate-400 mt-0.5">Map IOA addresses to named signals per communication line</p>
      </div>
      <button class="btn-primary" :disabled="!selectedLine" @click="openAdd">
        <Plus class="w-4 h-4" /> Map Signal
      </button>
    </div>

    <!-- Line selector -->
    <div class="flex items-center gap-3">
      <label class="text-sm text-slate-400">Line:</label>
      <select v-model.number="selectedLine" class="input-base w-64">
        <option v-for="l in lines" :key="l.id" :value="l.id">
          {{ l.name }} ({{ l.host }}:{{ l.port }})
        </option>
      </select>
    </div>

    <!-- Signals table -->
    <div class="card overflow-hidden">
      <div v-if="store.loading" class="px-4 py-10 text-center text-slate-400">Loading…</div>
      <div v-else-if="!store.signals.length" class="px-4 py-10 text-center text-slate-400">
        No signals mapped. Click <b>Map Signal</b> to add.
      </div>
      <table v-else class="table-base">
        <thead>
          <tr>
            <th class="th">Name</th>
            <th class="th">IOA</th>
            <th class="th">TypeID</th>
            <th class="th">Kind</th>
            <th class="th">Unit</th>
            <th class="th">Scale / Offset</th>
            <th class="th">Description</th>
            <th class="th">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in store.signals" :key="s.id" class="hover:bg-slate-750">
            <td class="td font-medium text-white">{{ s.name }}</td>
            <td class="td font-mono text-blue-300">{{ s.ioa }}</td>
            <td class="td text-xs font-mono text-slate-300">{{ typeIDName(s.type_id) }}</td>
            <td class="td">
              <span :class="s.signal_type === 'digital' ? 'badge-yellow' : 'badge-slate'">
                {{ s.signal_type }}
              </span>
            </td>
            <td class="td text-slate-400">{{ s.unit || '—' }}</td>
            <td class="td font-mono text-xs text-slate-400">
              {{ s.scale }} / {{ s.offset }}
            </td>
            <td class="td text-slate-400 max-w-xs truncate">{{ s.description || '—' }}</td>
            <td class="td">
              <div class="flex items-center gap-1">
                <button class="btn-ghost" @click="openEdit(s)">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button class="btn-ghost text-red-400 hover:text-red-300" @click="remove(s)">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <SignalForm
      v-model:open="showForm"
      :line-id="selectedLine ?? 0"
      :signal="editing"
      @saved="onSaved"
    />
  </div>
</template>
