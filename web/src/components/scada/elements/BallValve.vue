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

// ON = open (flow), OFF = closed
const fill = computed(() =>
  !props.signal ? '#1e293b'
  : (props.signal.quality & 0x80) ? props.config.color_fault
  : props.signal.value ? props.config.color_on : props.config.color_off
)

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const isOpen = computed(() =>
  props.signal && !(props.signal.quality & 0x80) && props.signal.value
)

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w * 0.38, props.h * 0.38))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Ball body -->
    <circle :cx="cx" :cy="cy" :r="r" :fill="fill" :stroke="stroke" stroke-width="1.5" style="transition: fill 0.3s ease, stroke 0.3s ease"/>

    <!-- Bore line: horizontal = open, vertical = closed -->
    <!-- Open bore (horizontal through ball) -->
    <line v-if="isOpen"
          :x1="cx - r * 0.75" :y1="cy" :x2="cx + r * 0.75" :y2="cy"
          stroke="#0a0e14" stroke-width="6" stroke-linecap="round"/>
    <!-- Closed bore (vertical, perpendicular to flow) -->
    <line v-else
          :x1="cx" :y1="cy - r * 0.75" :x2="cx" :y2="cy + r * 0.75"
          stroke="#0a0e14" stroke-width="6" stroke-linecap="round"/>

    <!-- T-handle actuator stem -->
    <line :x1="cx" :y1="cy - r" :x2="cx" :y2="cy - r - 9" :stroke="stroke" stroke-width="1.5" stroke-linecap="square" style="transition: stroke 0.3s ease"/>
    <!-- T-bar -->
    <line :x1="cx - 9" :y1="cy - r - 9" :x2="cx + 9" :y2="cy - r - 9" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- No signal indicator -->
    <text v-if="!signal" :x="cx" :y="cy + 1" text-anchor="middle" dominant-baseline="middle" font-size="7" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 13" text-anchor="middle" font-size="9" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
