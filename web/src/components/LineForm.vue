<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { linesAPI, type Line } from '@/api/client'
import { X } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  line: Partial<Line> | null
}>()
const emit = defineEmits<{
  'update:open': [v: boolean]
  saved: []
}>()

const form = ref<Partial<Line>>({})
const saving = ref(false)
const error = ref('')

const isDnp3 = computed(() => form.value.protocol === 'dnp3')

watch(() => props.open, (v) => {
  if (v) {
    form.value = props.line
      ? { protocol: 'iec104', ...props.line }
      : {
          protocol: 'iec104', port: 2404, common_address: 1,
          k: 12, w: 8, t1_ms: 15000, t2_ms: 10000, t3_ms: 20000,
          gi_interval_s: 900, dnp3_outstation_addr: 1024, dnp3_master_addr: 1,
          enabled: true,
        }
    error.value = ''
  }
})

// Swap port defaults when the driver changes, without clobbering a custom port.
function onProtocolChange() {
  if (isDnp3.value) {
    if (!form.value.port || form.value.port === 2404) form.value.port = 20000
    if (form.value.dnp3_outstation_addr == null) form.value.dnp3_outstation_addr = 1024
    if (form.value.dnp3_master_addr == null) form.value.dnp3_master_addr = 1
  } else if (!form.value.port || form.value.port === 20000) {
    form.value.port = 2404
  }
}

async function save() {
  if (!form.value.name || !form.value.host || !form.value.port) {
    error.value = 'Name, Host and Port are required'
    return
  }
  saving.value = true
  error.value = ''
  try {
    if (form.value.id) {
      await linesAPI.update(form.value.id, form.value)
    } else {
      await linesAPI.create(form.value)
    }
    emit('saved')
    emit('update:open', false)
  } catch (e: any) {
    error.value = e.response?.data?.error ?? e.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200"
      enter-from-class="opacity-0"
      leave-active-class="transition duration-150"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="emit('update:open', false)" />

        <!-- Dialog -->
        <div class="relative w-full max-w-lg card p-6 shadow-2xl space-y-5">
          <div class="flex items-center justify-between">
            <h2 class="text-lg font-semibold text-white">
              {{ form.id ? 'Edit Line' : 'Add Communication Line' }}
            </h2>
            <button @click="emit('update:open', false)" class="btn-ghost p-1">
              <X class="w-4 h-4" />
            </button>
          </div>

          <div v-if="error" class="px-3 py-2 bg-red-900/50 border border-red-700 rounded-md text-red-300 text-sm">
            {{ error }}
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="block text-xs font-medium text-slate-400 mb-1">Protocol</label>
              <select v-model="form.protocol" @change="onProtocolChange" class="input-base">
                <option value="iec104">IEC 60870-5-104 (master)</option>
                <option value="dnp3">DNP3 (master → outstation)</option>
              </select>
            </div>

            <div class="col-span-2">
              <label class="block text-xs font-medium text-slate-400 mb-1">Line Name *</label>
              <input v-model="form.name" class="input-base" placeholder="Substation RTU-01" />
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Host / IP *</label>
              <input v-model="form.host" class="input-base" placeholder="192.168.1.10" />
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Port *</label>
              <input v-model.number="form.port" type="number" class="input-base"
                :placeholder="isDnp3 ? '20000' : '2404'" />
            </div>

            <div v-if="!isDnp3">
              <label class="block text-xs font-medium text-slate-400 mb-1">Common Address (CA)</label>
              <input v-model.number="form.common_address" type="number" class="input-base" placeholder="1" />
            </div>

            <div v-if="isDnp3">
              <label class="block text-xs font-medium text-slate-400 mb-1">Outstation Address</label>
              <input v-model.number="form.dnp3_outstation_addr" type="number" class="input-base" placeholder="1024" />
            </div>

            <div v-if="isDnp3">
              <label class="block text-xs font-medium text-slate-400 mb-1">Master Address</label>
              <input v-model.number="form.dnp3_master_addr" type="number" class="input-base" placeholder="1" />
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">
                {{ isDnp3 ? 'Integrity-Poll Interval (s, 0=off)' : 'GI Interval (s, 0=off)' }}
              </label>
              <input v-model.number="form.gi_interval_s" type="number" class="input-base" />
            </div>

            <fieldset v-if="!isDnp3" class="col-span-2 border border-slate-600 rounded-md p-3 space-y-3">
              <legend class="text-xs text-slate-400 px-1">Link parameters</legend>
              <div class="grid grid-cols-3 gap-3">
                <div>
                  <label class="block text-xs text-slate-400 mb-1">K (window)</label>
                  <input v-model.number="form.k" type="number" class="input-base" />
                </div>
                <div>
                  <label class="block text-xs text-slate-400 mb-1">W (ack after)</label>
                  <input v-model.number="form.w" type="number" class="input-base" />
                </div>
                <div>
                  <label class="block text-xs text-slate-400 mb-1">T1 ms</label>
                  <input v-model.number="form.t1_ms" type="number" class="input-base" />
                </div>
                <div>
                  <label class="block text-xs text-slate-400 mb-1">T2 ms</label>
                  <input v-model.number="form.t2_ms" type="number" class="input-base" />
                </div>
                <div>
                  <label class="block text-xs text-slate-400 mb-1">T3 ms</label>
                  <input v-model.number="form.t3_ms" type="number" class="input-base" />
                </div>
              </div>
            </fieldset>

            <div class="col-span-2 flex items-center gap-2">
              <input id="enabled" v-model="form.enabled" type="checkbox"
                class="w-4 h-4 rounded bg-slate-700 border-slate-500 text-blue-500" />
              <label for="enabled" class="text-sm text-slate-300">Enable line on save</label>
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-2">
            <button class="btn-secondary" @click="emit('update:open', false)">Cancel</button>
            <button class="btn-primary" :disabled="saving" @click="save">
              {{ saving ? 'Saving…' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
