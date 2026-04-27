<script setup lang="ts">
import { ref, watch } from 'vue'
import { signalsAPI, TYPE_IDS, type Signal } from '@/api/client'
import { X } from 'lucide-vue-next'

const props = defineProps<{
  open: boolean
  lineId: number
  signal: Partial<Signal> | null
}>()
const emit = defineEmits<{
  'update:open': [v: boolean]
  saved: []
}>()

const form = ref<Partial<Signal>>({})
const saving = ref(false)
const error = ref('')

watch(() => props.open, (v) => {
  if (v) {
    form.value = props.signal
      ? { ...props.signal }
      : { line_id: props.lineId, type_id: 13, signal_type: 'analog', scale: 1, offset: 0, unit: '', description: '' }
    error.value = ''
  }
})

function onTypeChange() {
  const t = TYPE_IDS.find(t => t.id === form.value.type_id)
  if (t) form.value.signal_type = t.kind as 'digital' | 'analog'
}

async function save() {
  if (!form.value.name || !form.value.ioa) {
    error.value = 'Name and IOA are required'
    return
  }
  saving.value = true
  error.value = ''
  try {
    if (form.value.id) {
      await signalsAPI.update(form.value.id, form.value)
    } else {
      await signalsAPI.create({ ...form.value, line_id: props.lineId })
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
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="emit('update:open', false)" />

        <div class="relative w-full max-w-lg card p-6 shadow-2xl space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-lg font-semibold text-white">
              {{ form.id ? 'Edit Signal' : 'Map Signal' }}
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
              <label class="block text-xs font-medium text-slate-400 mb-1">Signal Name *</label>
              <input v-model="form.name" class="input-base" placeholder="Voltage Bus 1" />
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">IOA (1–16777215) *</label>
              <input v-model.number="form.ioa" type="number" min="1" max="16777215" class="input-base" />
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">TypeID *</label>
              <select v-model.number="form.type_id" @change="onTypeChange" class="input-base">
                <option v-for="t in TYPE_IDS" :key="t.id" :value="t.id">
                  {{ t.label }}
                </option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Signal Type</label>
              <select v-model="form.signal_type" class="input-base">
                <option value="analog">Analog</option>
                <option value="digital">Digital</option>
              </select>
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Unit</label>
              <input v-model="form.unit" class="input-base" placeholder="kV, MW, A…" />
            </div>

            <div v-if="form.signal_type === 'analog'">
              <label class="block text-xs font-medium text-slate-400 mb-1">Scale factor</label>
              <input v-model.number="form.scale" type="number" step="any" class="input-base" />
            </div>

            <div v-if="form.signal_type === 'analog'">
              <label class="block text-xs font-medium text-slate-400 mb-1">Offset</label>
              <input v-model.number="form.offset" type="number" step="any" class="input-base" />
            </div>

            <div class="col-span-2">
              <label class="block text-xs font-medium text-slate-400 mb-1">Description</label>
              <input v-model="form.description" class="input-base" placeholder="Optional description" />
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
