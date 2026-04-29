<script setup lang="ts">
import { computed } from 'vue'
import type { PctBarConfig, Datapoint } from '@/api/client'

const props = defineProps<{
  config:   PctBarConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
}>()

const pct = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return 0
  const range = props.config.max - props.config.min
  if (range === 0) return 0
  return Math.max(0, Math.min(1, (props.signal.value - props.config.min) / range))
})

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

// Vertical layout
const BAR_X  = computed(() => props.w * 0.25)
const BAR_W  = computed(() => props.w * 0.5)
const BAR_Y  = computed(() => 12)
const BAR_H  = computed(() => props.h - 36)
const fillH  = computed(() => BAR_H.value * pct.value)
const fillY  = computed(() => BAR_Y.value + BAR_H.value - fillH.value)

// Tick positions at 25 / 50 / 75 %
const ticks = computed(() => [0.25, 0.5, 0.75].map(t => ({
  y: BAR_Y.value + BAR_H.value * (1 - t),
  label: (t * 100).toFixed(0),
})))
</script>

<template>
  <svg
    :width="w" :height="h"
    :viewBox="`0 0 ${w} ${h}`"
    class="overflow-visible block"
    xmlns="http://www.w3.org/2000/svg"
  >
    <!-- Selection ring -->
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Background bar -->
    <rect :x="BAR_X" :y="BAR_Y" :width="BAR_W" :height="BAR_H"
          :fill="config.color_bg" rx="2"/>

    <!-- Fill bar (animated) -->
    <rect
      :x="BAR_X" :y="fillY" :width="BAR_W" :height="fillH"
      :fill="config.color_fill" rx="2"
      style="transition: y 0.4s ease, height 0.4s ease"
    />

    <!-- Border -->
    <rect :x="BAR_X" :y="BAR_Y" :width="BAR_W" :height="BAR_H"
          fill="none" stroke="#475569" stroke-width="1" rx="2"/>

    <!-- Ticks -->
    <g v-for="t in ticks" :key="t.label">
      <line :x1="BAR_X + BAR_W" :y1="t.y" :x2="BAR_X + BAR_W + 5" :y2="t.y"
            stroke="#475569" stroke-width="1"/>
      <text :x="BAR_X + BAR_W + 7" :y="t.y + 1"
            font-size="7" fill="#64748b" font-family="monospace"
            dominant-baseline="middle">{{ t.label }}%</text>
    </g>

    <!-- Value text -->
    <text :x="w/2" :y="BAR_Y + BAR_H + 11"
          text-anchor="middle" font-size="9"
          :fill="signal && !(signal.quality & 0x80) ? config.color_fill : '#6b7280'"
          font-family="monospace" font-weight="600">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :x="w/2" :y="BAR_Y + BAR_H + 22"
          text-anchor="middle" font-size="8" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
