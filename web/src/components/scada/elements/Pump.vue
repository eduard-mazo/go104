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

const running = computed(() =>
  props.signal && !(props.signal.quality & 0x80) && props.signal.value,
)

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w, props.h) / 2 - 2)
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Inlet pipe (left) -->
    <line :x1="0" :y1="cy" :x2="cx - r" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>
    <!-- Outlet pipe (top — centrifugal) -->
    <line :x1="cx" :y1="0" :x2="cx" :y2="cy - r"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>

    <!-- Pump body circle -->
    <circle :cx="cx" :cy="cy" :r="r"
            :fill="fill" stroke="#475569" stroke-width="1.5"
            style="transition: fill 0.35s ease"/>

    <!-- Impeller blades (4 lines from center) -->
    <g :style="{ transformOrigin: `${cx}px ${cy}px`, transition: 'transform 0.5s ease' }">
      <line v-for="angle in [0, 90, 180, 270]" :key="angle"
            :x1="cx" :y1="cy"
            :x2="cx + r*0.65 * Math.cos((angle - 30) * Math.PI / 180)"
            :y2="cy + r*0.65 * Math.sin((angle - 30) * Math.PI / 180)"
            stroke="white" stroke-width="2" stroke-linecap="round" opacity="0.7"/>
    </g>

    <!-- Center hub -->
    <circle :cx="cx" :cy="cy" :r="r * 0.18"
            fill="#0a0e14" stroke="#ffffff44" stroke-width="1"/>

    <!-- Running indicator dot -->
    <circle v-if="running"
            :cx="cx + r * 0.55" :cy="cy - r * 0.55" r="3"
            fill="white" opacity="0.8"/>

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
