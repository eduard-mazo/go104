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

// ON = open (flow allowed), OFF = closed
const fill = computed(() => {
  if (!props.signal) return '#374151'
  if (props.signal.quality & 0x80) return props.config.color_fault
  return props.signal.value ? props.config.color_on : props.config.color_off
})

const isOpen = computed(() => props.signal && !(props.signal.quality & 0x80) && props.signal.value)

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w * 0.35, props.h * 0.42))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Pipe stubs left/right -->
    <line :x1="0" :y1="cy" :x2="cx - r - 1" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>
    <line :x1="cx + r + 1" :y1="cy" :x2="w" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>

    <!-- Ball body -->
    <circle :cx="cx" :cy="cy" :r="r"
            :fill="fill" stroke="#475569" stroke-width="1.5"
            style="transition: fill 0.35s ease"/>

    <!-- Bore hole: horizontal (open) or vertical (closed) -->
    <!-- Open: horizontal bore aligned with pipe -->
    <rect v-if="isOpen"
          :x="cx - r" :y="cy - r*0.28" :width="r*2" :height="r*0.56"
          fill="#0a0e14" rx="3"
          style="transition: all 0.3s ease"/>
    <!-- Closed: vertical bore perpendicular to flow -->
    <rect v-else
          :x="cx - r*0.28" :y="cy - r" :width="r*0.56" :height="r*2"
          fill="#0a0e14" rx="3"
          style="transition: all 0.3s ease"/>

    <!-- Actuator stem (top) -->
    <rect :x="cx - 3" :y="cy - r - 10" width="6" height="10"
          fill="#475569" rx="1"/>
    <rect :x="cx - 7" :y="cy - r - 16" width="14" height="6"
          :fill="fill" stroke="#475569" stroke-width="1" rx="1"
          style="transition: fill 0.35s ease"/>

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
