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
  none:  '#374151',
}[state.value]))

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w, props.h) / 2 - 3)
const glowId = computed(() => `lamp-glow-${Math.random().toString(36).slice(2, 7)}`)
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <defs>
      <filter :id="glowId" x="-50%" y="-50%" width="200%" height="200%">
        <feGaussianBlur stdDeviation="3" result="blur"/>
        <feMerge>
          <feMergeNode in="blur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Glow ring when ON -->
    <circle v-if="state === 'on'"
            :cx="cx" :cy="cy" :r="r + 4"
            :fill="fill" opacity="0.2"
            style="transition: opacity 0.35s ease"/>

    <!-- Body -->
    <circle :cx="cx" :cy="cy" :r="r"
            :fill="fill" stroke="#475569" stroke-width="1.5"
            :filter="state === 'on' ? `url(#${glowId})` : undefined"
            style="transition: fill 0.35s ease"/>

    <!-- Shine highlight (top-left) -->
    <circle :cx="cx - r*0.25" :cy="cy - r*0.3" :r="r*0.2"
            fill="white" opacity="0.25"/>

    <!-- No signal marker -->
    <text v-if="state === 'none'" :x="cx" :y="cy + 1"
          text-anchor="middle" dominant-baseline="middle"
          font-size="8" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
