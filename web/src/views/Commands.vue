<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { linesAPI, commandsAPI, CMD_TYPE_IDS, type Line } from '@/api/client'
import { Terminal, Send } from 'lucide-vue-next'

interface CmdHistory {
  time: string
  line: string
  ioa: number
  typeID: string
  value: number
  select: boolean
  result: string
  ok: boolean
}

const lines = ref<Line[]>([])
const history = ref<CmdHistory[]>([])
const sending = ref(false)
const error = ref('')

const form = ref({
  line_id: 0,
  ioa: 0,
  type_id: 45,
  value: 0,
  select: false,
})

onMounted(async () => {
  lines.value = await linesAPI.list()
  if (lines.value.length) form.value.line_id = lines.value[0].id
})

async function send() {
  if (!form.value.line_id || !form.value.ioa) {
    error.value = 'Select a line and enter IOA'
    return
  }
  sending.value = true
  error.value = ''
  const line = lines.value.find(l => l.id === form.value.line_id)
  const typeName = CMD_TYPE_IDS.find(t => t.id === form.value.type_id)?.label ?? String(form.value.type_id)

  try {
    await commandsAPI.send({ ...form.value })
    history.value.unshift({
      time: new Date().toLocaleTimeString(),
      line: line?.name ?? '?',
      ioa: form.value.ioa,
      typeID: typeName,
      value: form.value.value,
      select: form.value.select,
      result: 'Queued',
      ok: true,
    })
  } catch (e: any) {
    const msg = e.response?.data?.error ?? e.message
    error.value = msg
    history.value.unshift({
      time: new Date().toLocaleTimeString(),
      line: line?.name ?? '?',
      ioa: form.value.ioa,
      typeID: typeName,
      value: form.value.value,
      select: form.value.select,
      result: msg,
      ok: false,
    })
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div class="p-6 space-y-6">
    <div>
      <h1 class="text-xl font-bold text-white flex items-center gap-2">
        <Terminal class="w-5 h-5 text-blue-400" /> Command Console
      </h1>
      <p class="text-sm text-slate-400 mt-0.5">Send single/double commands and setpoints to slaves</p>
    </div>

    <!-- Command form -->
    <div class="card p-5 space-y-4 max-w-lg">
      <div v-if="error" class="px-3 py-2 bg-red-900/50 border border-red-700 rounded-md text-red-300 text-sm">
        {{ error }}
      </div>

      <div>
        <label class="block text-xs font-medium text-slate-400 mb-1">Communication Line</label>
        <select v-model.number="form.line_id" class="input-base">
          <option v-for="l in lines" :key="l.id" :value="l.id">
            {{ l.name }} — {{ l.host }}:{{ l.port }}
          </option>
        </select>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-medium text-slate-400 mb-1">IOA</label>
          <input v-model.number="form.ioa" type="number" class="input-base" placeholder="1000" />
        </div>
        <div>
          <label class="block text-xs font-medium text-slate-400 mb-1">Value</label>
          <input v-model.number="form.value" type="number" step="any" class="input-base" placeholder="0 / 1 / float" />
        </div>
      </div>

      <div>
        <label class="block text-xs font-medium text-slate-400 mb-1">Command TypeID</label>
        <select v-model.number="form.type_id" class="input-base">
          <option v-for="t in CMD_TYPE_IDS" :key="t.id" :value="t.id">{{ t.label }}</option>
        </select>
      </div>

      <div class="flex items-center gap-2">
        <input id="sel" v-model="form.select" type="checkbox"
          class="w-4 h-4 rounded bg-slate-700 border-slate-500 text-blue-500" />
        <label for="sel" class="text-sm text-slate-300">Select-Before-Operate (set S/E bit)</label>
      </div>

      <button class="btn-primary w-full justify-center" :disabled="sending" @click="send">
        <Send class="w-4 h-4" />
        {{ sending ? 'Sending…' : 'Send Command' }}
      </button>
    </div>

    <!-- History -->
    <div v-if="history.length" class="space-y-2">
      <h2 class="text-sm font-semibold text-slate-400 uppercase tracking-wider">Command History</h2>
      <div class="card overflow-hidden">
        <table class="table-base">
          <thead>
            <tr>
              <th class="th">Time</th>
              <th class="th">Line</th>
              <th class="th">IOA</th>
              <th class="th">TypeID</th>
              <th class="th">Value</th>
              <th class="th">S/E</th>
              <th class="th">Result</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(h, i) in history" :key="i">
              <td class="td font-mono text-xs text-slate-400">{{ h.time }}</td>
              <td class="td text-slate-300">{{ h.line }}</td>
              <td class="td font-mono text-blue-300">{{ h.ioa }}</td>
              <td class="td text-xs text-slate-400">{{ h.typeID.split(' — ')[0] }}</td>
              <td class="td font-mono">{{ h.value }}</td>
              <td class="td text-center">{{ h.select ? '✓' : '—' }}</td>
              <td class="td">
                <span :class="h.ok ? 'text-green-400' : 'text-red-400'">{{ h.result }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
