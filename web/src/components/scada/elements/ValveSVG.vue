<script setup lang="ts">
import { computed } from 'vue'
import type { ValveConfig } from '@/api/client'
import type { Datapoint } from '@/api/client'

const props = defineProps<{
  config:   ValveConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
}>()

const fill = computed(() => {
  if (!props.signal) return '#374151'
  if (props.signal.quality & 0x80) return props.config.color_fault
  return props.signal.value ? props.config.color_on : props.config.color_off
})

const strokeColor = computed(() =>
  props.selected ? '#f59e0b' : '#475569'
)
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

    <!-- Valve body: two opposing triangles -->
    <g :style="{ transition: 'fill 0.35s ease' }">
      <!-- left triangle -->
      <polygon
        :points="`0,0 ${w/2},${h/2} 0,${h}`"
        :fill="fill" :stroke="strokeColor" stroke-width="1.5"
        style="transition: fill 0.35s ease"
      />
      <!-- right triangle -->
      <polygon
        :points="`${w},0 ${w/2},${h/2} ${w},${h}`"
        :fill="fill" :stroke="strokeColor" stroke-width="1.5"
        style="transition: fill 0.35s ease"
      />
    </g>

    <!-- Actuator stem -->
    <line :x1="w/2" :y1="0" :x2="w/2" :y2="-10"
          :stroke="strokeColor" stroke-width="2"/>
    <!-- Actuator box -->
    <rect :x="w/2 - 10" :y="-22" width="20" height="12"
          :fill="fill" :stroke="strokeColor" stroke-width="1.5"
          style="transition: fill 0.35s ease" rx="1"/>

    <!-- No-data indicator -->
    <text v-if="!signal" :x="w/2" :y="h/2+1"
          text-anchor="middle" dominant-baseline="middle"
          font-size="8" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label"
          :x="w/2" :y="h + 14"
          text-anchor="middle"
          font-size="9" fill="#94a3b8" font-family="monospace"
          letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
