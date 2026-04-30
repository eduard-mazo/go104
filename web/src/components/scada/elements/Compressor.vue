<script setup lang="ts">
import { computed } from 'vue'
import type { ValveConfig, Datapoint } from '@/api/client'

const props = defineProps<{
  config:   ValveConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
}>()

// ON = running
const fill = computed(() => {
  if (!props.signal) return '#374151'
  if (props.signal.quality & 0x80) return props.config.color_fault
  return props.signal.value ? props.config.color_on : props.config.color_off
})

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w, props.h) / 2 - 2)

// Compression wave path — 3 arcs converging left-to-right
function wavePath(scale: number, cx: number, cy: number, r: number) {
  const y0 = cy - r * scale
  const y1 = cy
  const h  = r * scale
  return `M ${cx - r*0.55} ${y0} q ${r*0.18} ${h} 0 ${h*2} M ${cx - r*0.18} ${y0} q ${r*0.18} ${h} 0 ${h*2} M ${cx + r*0.18} ${y0} q ${r*0.18} ${h} 0 ${h*2}`
}
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Inlet/outlet pipe stubs (left, right) -->
    <line :x1="0" :y1="cy" :x2="cx - r" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>
    <line :x1="cx + r" :y1="cy" :x2="w" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>

    <!-- Body -->
    <circle :cx="cx" :cy="cy" :r="r"
            :fill="fill" stroke="#475569" stroke-width="1.5"
            style="transition: fill 0.35s ease"/>

    <!-- Compression wave symbol -->
    <path :d="wavePath(0.45, cx, cy, r)"
          fill="none" stroke="white" stroke-width="1.5"
          stroke-linecap="round" opacity="0.75"/>

    <!-- C label -->
    <text :x="cx + r*0.55" :y="cy + 1"
          text-anchor="middle" dominant-baseline="middle"
          :font-size="r * 0.42" fill="white" font-family="monospace" font-weight="700"
          opacity="0.5">C</text>

    <!-- No signal -->
    <text v-if="!signal" :x="cx" :y="cy + 1"
          text-anchor="middle" dominant-baseline="middle"
          font-size="8" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
