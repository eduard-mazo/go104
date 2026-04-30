<script setup lang="ts">
import { computed } from 'vue'
import type { GaugeConfig, Datapoint } from '@/api/client'

const props = defineProps<{
  config:   GaugeConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
}>()

const alarm = computed(() =>
  props.config.alarm_high > 0 &&
  props.signal != null &&
  !(props.signal.quality & 0x80) &&
  props.signal.value >= props.config.alarm_high,
)

const coilColor = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '#374151'
  return alarm.value ? props.config.color_alarm : props.config.color_fill
})

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2 - 4)
// Each coil circle radius
const cr = computed(() => Math.min(props.w * 0.22, props.h * 0.32))
const gap = 3
// Left coil center, right coil center
const lx = computed(() => cx.value - cr.value - gap / 2)
const rx = computed(() => cx.value + cr.value + gap / 2)

// Coil arcs (5 semicircles stacked to suggest windings)
function coilPath(centX: number, centY: number, radius: number, n: number) {
  const step = (radius * 2) / n
  let d = ''
  for (let i = 0; i < n; i++) {
    const y = centY - radius + i * step + step / 2
    const sx = centX - radius * 0.55
    const ex = centX + radius * 0.55
    d += `M ${sx} ${y} A ${radius*0.55} ${step/2} 0 0 1 ${ex} ${y} `
  }
  return d
}
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Left coil circle -->
    <circle :cx="lx" :cy="cy" :r="cr"
            fill="#0f172a" :stroke="coilColor" stroke-width="1.5"
            style="transition: stroke 0.35s ease"/>

    <!-- Right coil circle -->
    <circle :cx="rx" :cy="cy" :r="cr"
            fill="#0f172a" :stroke="coilColor" stroke-width="1.5"
            style="transition: stroke 0.35s ease"/>

    <!-- Winding arcs — left coil -->
    <path :d="coilPath(lx, cy, cr * 0.7, 5)"
          fill="none" :stroke="coilColor" stroke-width="1.2"
          stroke-linecap="round" opacity="0.7"
          style="transition: stroke 0.35s ease"/>

    <!-- Winding arcs — right coil -->
    <path :d="coilPath(rx, cy, cr * 0.7, 5)"
          fill="none" :stroke="coilColor" stroke-width="1.2"
          stroke-linecap="round" opacity="0.7"
          style="transition: stroke 0.35s ease"/>

    <!-- Center dividing line -->
    <line :x1="cx" :y1="cy - cr * 0.85"
          :x2="cx" :y2="cy + cr * 0.85"
          :stroke="coilColor" stroke-width="1" opacity="0.4"/>

    <!-- Pipe stubs top & bottom (power lines) -->
    <line :x1="lx" :y1="cy - cr - 1" :x2="lx" :y2="0"
          stroke="#475569" stroke-width="2" stroke-linecap="round"/>
    <line :x1="rx" :y1="cy - cr - 1" :x2="rx" :y2="0"
          stroke="#475569" stroke-width="2" stroke-linecap="round"/>
    <line :x1="lx" :y1="cy + cr + 1" :x2="lx" :y2="h - 18"
          stroke="#475569" stroke-width="2" stroke-linecap="round"/>
    <line :x1="rx" :y1="cy + cr + 1" :x2="rx" :y2="h - 18"
          stroke="#475569" stroke-width="2" stroke-linecap="round"/>

    <!-- Value -->
    <text :x="cx" :y="h - 6"
          text-anchor="middle" font-size="9"
          :fill="coilColor" font-family="monospace" font-weight="600"
          style="transition: fill 0.35s ease">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
