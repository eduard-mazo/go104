<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ScadaElement, Signal, ValveConfig, PctBarConfig } from '@/api/client'

const props = defineProps<{
  element:    ScadaElement
  allSignals: Signal[]
}>()
const emit = defineEmits<{
  update: [el: ScadaElement]
  delete: [id: string]
}>()

// Local copy — emit on every change
const local = ref<ScadaElement>(JSON.parse(JSON.stringify(props.element)))

watch(() => props.element.id, () => {
  local.value = JSON.parse(JSON.stringify(props.element))
})

function push() { emit('update', JSON.parse(JSON.stringify(local.value))) }

// Typed config accessors
const valveCfg  = computed(() => local.value.config as ValveConfig)
const pctCfg    = computed(() => local.value.config as PctBarConfig)

const signalLabel = (sig: Signal) =>
  `${sig.line_name} / ${sig.name} (IOA ${sig.ioa})`
</script>

<template>
  <div class="flex flex-col h-full text-xs text-slate-300 overflow-y-auto">
    <div class="px-4 py-3 border-b border-slate-800 font-mono uppercase tracking-widest text-[10px] text-slate-500">
      Properties
    </div>

    <!-- Geometry -->
    <div class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Geometry</p>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">X</label>
          <input v-model.number="local.x" type="number" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Y</label>
          <input v-model.number="local.y" type="number" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">W</label>
          <input v-model.number="local.w" type="number" min="20" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">H</label>
          <input v-model.number="local.h" type="number" min="20" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
      </div>
      <div>
        <label class="block text-slate-500 mb-0.5">Rotation</label>
        <div class="flex gap-1">
          <button v-for="r in [0, 90, 180, 270]" :key="r"
            class="flex-1 py-1 rounded border text-center transition-colors"
            :class="local.rotation === r
              ? 'border-amber-500 text-amber-400 bg-amber-950/30'
              : 'border-slate-700 text-slate-400 hover:border-slate-500'"
            @click="local.rotation = r as any; push()">
            {{ r }}°
          </button>
        </div>
      </div>
    </div>

    <!-- Signal binding -->
    <div class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Signal</p>
      <select
        :value="local.signal_id ?? ''"
        class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1.5 text-slate-200"
        @change="local.signal_id = ($event.target as HTMLSelectElement).value ? Number(($event.target as HTMLSelectElement).value) : null; push()"
      >
        <option value="">— none —</option>
        <option v-for="s in allSignals" :key="s.id" :value="s.id">
          {{ signalLabel(s) }}
        </option>
      </select>
    </div>

    <!-- Valve config -->
    <div v-if="local.kind === 'valve'" class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Valve</p>
      <div>
        <label class="block text-slate-500 mb-0.5">Label</label>
        <input v-model="valveCfg.label" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="FCV-101" />
      </div>
      <div class="grid grid-cols-3 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">ON color</label>
          <div class="flex items-center gap-1">
            <input type="color" v-model="valveCfg.color_on" class="w-7 h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
            <span class="font-mono text-[10px]">{{ valveCfg.color_on }}</span>
          </div>
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">OFF color</label>
          <div class="flex items-center gap-1">
            <input type="color" v-model="valveCfg.color_off" class="w-7 h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
            <span class="font-mono text-[10px]">{{ valveCfg.color_off }}</span>
          </div>
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Fault</label>
          <div class="flex items-center gap-1">
            <input type="color" v-model="valveCfg.color_fault" class="w-7 h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
            <span class="font-mono text-[10px]">{{ valveCfg.color_fault }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- % Bar config -->
    <div v-if="local.kind === 'pct_bar'" class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">% Bar</p>
      <div>
        <label class="block text-slate-500 mb-0.5">Label</label>
        <input v-model="pctCfg.label" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="Flow %" />
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">Min</label>
          <input v-model.number="pctCfg.min" type="number" step="any" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Max</label>
          <input v-model.number="pctCfg.max" type="number" step="any" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
      </div>
      <div>
        <label class="block text-slate-500 mb-0.5">Unit</label>
        <input v-model="pctCfg.unit" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="%, kPa, m³/h…" />
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">Fill color</label>
          <div class="flex items-center gap-1">
            <input type="color" v-model="pctCfg.color_fill" class="w-7 h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
          </div>
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">BG color</label>
          <div class="flex items-center gap-1">
            <input type="color" v-model="pctCfg.color_bg" class="w-7 h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
          </div>
        </div>
      </div>
    </div>

    <!-- Delete -->
    <div class="px-4 py-3 mt-auto">
      <button
        class="w-full py-1.5 rounded border border-red-800 text-red-400 hover:bg-red-950/40 transition-colors text-xs"
        @click="emit('delete', element.id)"
      >
        Delete element
      </button>
    </div>
  </div>
</template>
