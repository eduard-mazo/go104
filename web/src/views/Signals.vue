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
  <div class="p-4 sm:p-6 space-y-6">
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-extrabold flex items-center gap-2 tracking-tight">
          <Radio class="w-5 h-5 text-[color:var(--epm-citrico)]" />
          Signal Mapping
        </h1>
        <p class="text-sm text-muted-foreground mt-0.5">Map IOA addresses to named signals</p>
      </div>
      <button class="btn-primary shrink-0" :disabled="!selectedLine" @click="openAdd">
        <Plus class="w-4 h-4" /> <span class="hidden sm:inline">Map Signal</span>
      </button>
    </div>

    <!-- Line selector -->
    <div class="flex items-center gap-3 flex-wrap">
      <label class="text-sm text-muted-foreground font-semibold shrink-0">Line:</label>
      <select v-model.number="selectedLine" class="input-base w-64">
        <option v-for="l in lines" :key="l.id" :value="l.id">
          {{ l.name }} ({{ l.host }}:{{ l.port }})
        </option>
      </select>
    </div>

    <!-- Signals table -->
    <div class="card overflow-x-auto">
      <div v-if="store.loading" class="px-4 py-10 text-center text-muted-foreground text-sm">Loading…</div>
      <div v-else-if="!store.signals.length" class="px-4 py-10 text-center text-muted-foreground text-sm">
        No signals mapped. Click <b>Map Signal</b> to add.
      </div>
      <table v-else class="table-base min-w-[700px]">
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
          <tr v-for="s in store.signals" :key="s.id" class="data-row">
            <td class="td font-semibold">{{ s.name }}</td>
            <td class="td tabular text-[color:var(--signal-wait)]">{{ s.ioa }}</td>
            <td class="td text-xs font-mono text-muted-foreground">{{ typeIDName(s.type_id) }}</td>
            <td class="td">
              <span :class="s.signal_type === 'digital' ? 'badge-warn' : 'badge-muted'">
                {{ s.signal_type }}
              </span>
            </td>
            <td class="td text-muted-foreground">{{ s.unit || '—' }}</td>
            <td class="td tabular text-muted-foreground text-xs">
              {{ s.scale }} / {{ s.offset }}
            </td>
            <td class="td text-muted-foreground max-w-[200px] truncate">{{ s.description || '—' }}</td>
            <td class="td">
              <div class="flex items-center gap-1">
                <button class="btn-ghost" @click="openEdit(s)">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button class="btn-ghost text-destructive hover:text-destructive" @click="remove(s)">
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
