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

// ON = lit
const state = computed<'on' | 'off' | 'fault' | 'none'>(() => {
  if (!props.signal) return 'none'
  if (props.signal.quality & 0x80) return 'fault'
  return props.signal.value ? 'on' : 'off'
})

const fill = computed(() => ({
  on:    props.config.color_on,
  off:   props.config.color_off,
  fault: props.config.color_fault,
  none:  '#1e293b',
}[state.value]))

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const crossStroke = computed(() =>
  state.value === 'on' ? 'white' : '#64748b'
)

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w, props.h) / 2 - 2)
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Body circle — flat fill, no gradient/glow -->
    <circle :cx="cx" :cy="cy" :r="r" :fill="fill" :stroke="stroke" stroke-width="1.5" style="transition: fill 0.3s ease, stroke 0.3s ease"/>

    <!-- ISA lamp symbol: X cross (two diagonal lines inside circle) -->
    <line :x1="cx - r * 0.55" :y1="cy - r * 0.55" :x2="cx + r * 0.55" :y2="cy + r * 0.55"
          :stroke="crossStroke" stroke-width="1.5" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <line :x1="cx + r * 0.55" :y1="cy - r * 0.55" :x2="cx - r * 0.55" :y2="cy + r * 0.55"
          :stroke="crossStroke" stroke-width="1.5" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- No signal -->
    <text v-if="state === 'none'" :x="cx" :y="cy + 1" text-anchor="middle" dominant-baseline="middle" font-size="7" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 13" text-anchor="middle" font-size="9" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
