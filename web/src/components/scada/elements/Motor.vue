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
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Body circle -->
    <circle :cx="cx" :cy="cy" :r="r"
            :fill="fill" stroke="#475569" stroke-width="1.5"
            style="transition: fill 0.35s ease"/>

    <!-- Inner ring -->
    <circle :cx="cx" :cy="cy" :r="r * 0.72"
            fill="none" stroke="#ffffff22" stroke-width="1"/>

    <!-- M glyph -->
    <text :x="cx" :y="cy + 1"
          text-anchor="middle" dominant-baseline="middle"
          :font-size="r * 0.7" fill="white" font-family="monospace" font-weight="700">M</text>

    <!-- Running indicator: small arc sweep at top right -->
    <path v-if="signal && signal.value && !(signal.quality & 0x80)"
          :d="`M ${cx + r*0.55} ${cy - r*0.55} a ${r*0.3} ${r*0.3} 0 0 1 ${r*0.3} 0`"
          fill="none" stroke="white" stroke-width="1.5" stroke-linecap="round" opacity="0.6"/>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
