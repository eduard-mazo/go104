<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ScadaElement, Signal, ValveConfig, PctBarConfig, GaugeConfig, PipeConfig } from '@/api/client'

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
const gaugeCfg  = computed(() => local.value.config as GaugeConfig)
const pipeCfg   = computed(() => local.value.config as PipeConfig)

const DIGITAL_KINDS = ['valve','circuit_breaker','motor','indicator_lamp','ball_valve','pump','compressor','control_valve','check_valve']
const GAUGE_KINDS   = ['transformer','tank','flow_meter','pressure_gauge','heat_exchanger','temp_tx']
const isDigital     = computed(() => DIGITAL_KINDS.includes(local.value.kind))
const isGauge       = computed(() => GAUGE_KINDS.includes(local.value.kind))
const isPipe        = computed(() => local.value.kind === 'pipe_segment')

const kindLabel = computed(() => ({
  valve:           'Gate Valve',
  pct_bar:         '% Bar',
  circuit_breaker: 'Circuit Breaker',
  motor:           'Motor',
  transformer:     'Transformer',
  indicator_lamp:  'Lamp',
  ball_valve:      'Ball Valve',
  pump:            'Pump',
  tank:            'Tank',
  flow_meter:      'Flow Meter',
  pressure_gauge:  'Pressure Gauge',
  compressor:      'Compressor',
  control_valve:   'Control Valve',
  check_valve:     'Check Valve',
  heat_exchanger:  'Heat Exchanger',
  temp_tx:         'Temp Transmitter',
  pipe_segment:    'Pipe Segment',
}[local.value.kind] ?? local.value.kind))

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

    <!-- Kind header -->
    <div class="px-4 py-2 border-b border-slate-800/60">
      <span class="text-[10px] font-mono uppercase tracking-widest text-amber-600">{{ kindLabel }}</span>
    </div>

    <!-- Digital element config (valve, breaker, motor, lamp, ball valve, pump, compressor) -->
    <div v-if="isDigital" class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Appearance</p>
      <div>
        <label class="block text-slate-500 mb-0.5">Label</label>
        <input v-model="valveCfg.label" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="Tag / name" />
      </div>
      <div class="grid grid-cols-3 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">ON</label>
          <input type="color" v-model="valveCfg.color_on" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">OFF</label>
          <input type="color" v-model="valveCfg.color_off" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Fault</label>
          <input type="color" v-model="valveCfg.color_fault" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
      </div>
    </div>

    <!-- % Bar config -->
    <div v-if="local.kind === 'pct_bar'" class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Appearance</p>
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
          <label class="block text-slate-500 mb-0.5">Fill</label>
          <input type="color" v-model="pctCfg.color_fill" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">BG</label>
          <input type="color" v-model="pctCfg.color_bg" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
      </div>
    </div>

    <!-- Gauge config (tank, flow meter, pressure gauge, transformer) -->
    <div v-if="isGauge" class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Appearance</p>
      <div>
        <label class="block text-slate-500 mb-0.5">Label</label>
        <input v-model="gaugeCfg.label" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="PT-101" />
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">Min</label>
          <input v-model.number="gaugeCfg.min" type="number" step="any" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Max</label>
          <input v-model.number="gaugeCfg.max" type="number" step="any" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">Unit</label>
          <input v-model="gaugeCfg.unit" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="bar, m³…" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Alarm ≥</label>
          <input v-model.number="gaugeCfg.alarm_high" type="number" step="any" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200 font-mono" @change="push" placeholder="0 = off" />
        </div>
      </div>
      <div class="grid grid-cols-2 gap-2">
        <div>
          <label class="block text-slate-500 mb-0.5">Normal</label>
          <input type="color" v-model="gaugeCfg.color_fill" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
        <div>
          <label class="block text-slate-500 mb-0.5">Alarm</label>
          <input type="color" v-model="gaugeCfg.color_alarm" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
        </div>
      </div>
    </div>

    <!-- Pipe segment config -->
    <div v-if="isPipe" class="px-4 py-3 space-y-2 border-b border-slate-800">
      <p class="text-[10px] uppercase tracking-widest text-slate-600 mb-1">Appearance</p>
      <div>
        <label class="block text-slate-500 mb-0.5">Label</label>
        <input v-model="pipeCfg.label" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1 text-slate-200" @input="push" placeholder="Feed line" />
      </div>
      <div>
        <label class="block text-slate-500 mb-0.5">Style</label>
        <select v-model="pipeCfg.style" class="w-full bg-[#0a0e14] border border-slate-700 rounded px-2 py-1.5 text-slate-200" @change="push">
          <option value="process">Process (thick)</option>
          <option value="utility">Utility (medium)</option>
          <option value="instrument">Instrument (thin)</option>
        </select>
      </div>
      <div>
        <label class="block text-slate-500 mb-0.5">Color</label>
        <input type="color" v-model="pipeCfg.color" class="w-full h-7 cursor-pointer rounded border border-slate-700 bg-transparent" @input="push" />
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
