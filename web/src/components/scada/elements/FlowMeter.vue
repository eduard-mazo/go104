<script setup lang="ts">
import { computed } from 'vue'
import type { GaugeConfig, Datapoint } from '@/api/client'

const props = defineProps<{
  config:   GaugeConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
}>()

const alarm = computed(() =>
  props.config.alarm_high > 0 &&
  props.signal != null &&
  !(props.signal.quality & 0x80) &&
  props.signal.value >= props.config.alarm_high,
)

const fillColor = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '#374151'
  return alarm.value ? props.config.color_alarm : props.config.color_fill
})

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '?'
  return props.signal.value.toFixed(1)
})

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const hw = computed(() => props.w / 2 - 2)   // half-width of diamond
const hh = computed(() => props.h / 2 - 10)  // half-height of diamond
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Pipe stubs -->
    <line :x1="0" :y1="cy" :x2="2" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>
    <line :x1="w - 2" :y1="cy" :x2="w" :y2="cy"
          stroke="#475569" stroke-width="3" stroke-linecap="round"/>

    <!-- Diamond body -->
    <polygon
      :points="`${cx},${cy - hh} ${cx + hw},${cy} ${cx},${cy + hh} ${cx - hw},${cy}`"
      :fill="fillColor" stroke="#475569" stroke-width="1.5"
      style="transition: fill 0.35s ease"/>

    <!-- Flow arrows (→ →) inside diamond -->
    <g fill="none" :stroke="signal && !(signal.quality & 0x80) ? 'white' : '#6b7280'"
       stroke-width="1.5" stroke-linecap="round" opacity="0.7">
      <polyline :points="`${cx - hw*0.35},${cy} ${cx - hw*0.05},${cy}`"/>
      <polyline :points="`${cx - hw*0.05},${cy - hh*0.25} ${cx + hw*0.15},${cy} ${cx - hw*0.05},${cy + hh*0.25}`"/>
      <polyline :points="`${cx + hw*0.15},${cy} ${cx + hw*0.45},${cy}`"/>
    </g>

    <!-- Value text -->
    <text :x="cx" :y="cy + hh + 11"
          text-anchor="middle" font-size="9"
          :fill="fillColor" font-family="monospace" font-weight="600"
          style="transition: fill 0.35s ease">
      {{ displayVal }}<tspan v-if="config.unit" font-size="7" fill="#64748b"> {{ config.unit }}</tspan>
    </text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
