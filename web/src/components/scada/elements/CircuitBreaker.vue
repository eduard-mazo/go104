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

// ON = closed (energized), OFF = open
const state = computed<'on' | 'off' | 'fault' | 'none'>(() => {
  if (!props.signal) return 'none'
  if (props.signal.quality & 0x80) return 'fault'
  return props.signal.value ? 'on' : 'off'
})

const bodyColor = computed(() => ({
  on:    props.config.color_on,
  off:   props.config.color_off,
  fault: props.config.color_fault,
  none:  '#374151',
}[state.value]))

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const bw = computed(() => Math.min(props.w * 0.4, 28))
const bh = computed(() => Math.min(props.h * 0.55, 28))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Left pipe stub -->
    <line :x1="0" :y1="cy" :x2="cx - bw/2 - 1" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>
    <!-- Right pipe stub -->
    <line :x1="cx + bw/2 + 1" :y1="cy" :x2="w" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>

    <!-- Body rectangle -->
    <rect :x="cx - bw/2" :y="cy - bh/2" :width="bw" :height="bh"
          :fill="bodyColor" stroke="#475569" stroke-width="1.5" rx="2"
          style="transition: fill 0.35s ease"/>

    <!-- Closed symbol: horizontal line through box -->
    <line v-if="state === 'on'"
          :x1="cx - bw/2 + 3" :y1="cy" :x2="cx + bw/2 - 3" :y2="cy"
          stroke="white" stroke-width="2.5" stroke-linecap="round"/>

    <!-- Open symbol: diagonal slash -->
    <line v-else-if="state === 'off'"
          :x1="cx - bw/2 + 4" :y1="cy + bh/2 - 4" :x2="cx + bw/2 - 4" :y2="cy - bh/2 + 4"
          stroke="white" stroke-width="2" stroke-linecap="round"/>

    <!-- Fault: X -->
    <template v-else-if="state === 'fault'">
      <line :x1="cx - bw/2 + 4" :y1="cy - bh/2 + 4" :x2="cx + bw/2 - 4" :y2="cy + bh/2 - 4"
            stroke="white" stroke-width="2" stroke-linecap="round"/>
      <line :x1="cx - bw/2 + 4" :y1="cy + bh/2 - 4" :x2="cx + bw/2 - 4" :y2="cy - bh/2 + 4"
            stroke="white" stroke-width="2" stroke-linecap="round"/>
    </template>

    <!-- No signal -->
    <text v-if="state === 'none'" :x="cx" :y="cy + 1"
          text-anchor="middle" dominant-baseline="middle"
          font-size="8" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
