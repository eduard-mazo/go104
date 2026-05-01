<script setup lang="ts">
import { computed } from 'vue'
import type { GaugeConfig, Datapoint } from '@/api/client'

const props = defineProps<{
  config:   GaugeConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
  rotation: 0 | 90 | 180 | 270
}>()

const alarm = computed(() =>
  props.config.alarm_high > 0 &&
  props.signal != null &&
  !(props.signal.quality & 0x80) &&
  props.signal.value >= props.config.alarm_high
)

const indicatorColor = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '#64748b'
  return alarm.value ? props.config.color_alarm : props.config.color_fill
})

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1)
})

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w, props.h) * 0.4)

const fScale = computed(() => Math.min(props.w, props.h) / 60)
const fLabel = computed(() => Math.max(7, Math.min(13, Math.round(9  * props.w / 60))))
const fInner = computed(() => Math.max(6, Math.min(12, Math.round(9  * fScale.value))))
const fTag   = computed(() => Math.max(7, Math.min(16, Math.round(11 * fScale.value))))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- ISA instrument circle -->
    <circle :cx="cx" :cy="cy" :r="r" fill="#0f1923" :stroke="stroke" stroke-width="1.5" style="transition: stroke 0.3s ease"/>

    <!-- DCS horizontal line through circle center -->
    <line :x1="cx - r" :y1="cy" :x2="cx + r" :y2="cy" :stroke="stroke" stroke-width="1" opacity="0.6" style="transition: stroke 0.3s ease"/>

    <!-- "PT" tag in upper half -->
    <text :transform="rotation ? `rotate(${-rotation}, ${cx}, ${cy - r * 0.2})` : undefined"
          :x="cx" :y="cy - r * 0.2"
          text-anchor="middle" dominant-baseline="middle"
          :font-size="fTag" :fill="indicatorColor" font-family="monospace" font-weight="700"
          style="transition: fill 0.3s ease">PT</text>

    <!-- Value in lower half -->
    <text :transform="rotation ? `rotate(${-rotation}, ${cx}, ${cy + r * 0.38})` : undefined"
          :x="cx" :y="cy + r * 0.38"
          text-anchor="middle" dominant-baseline="middle"
          :font-size="fInner" :fill="indicatorColor" font-family="monospace" font-weight="600"
          style="transition: fill 0.3s ease">{{ displayVal }}</text>

    <!-- Unit below circle -->
    <text v-if="config.unit" :transform="rotation ? `rotate(${-rotation}, ${cx}, ${cy + r + 9})` : undefined"
          :x="cx" :y="cy + r + 9"
          text-anchor="middle" :font-size="Math.max(6, fInner-1)" fill="#64748b" font-family="monospace">{{ config.unit }}</text>

    <!-- Impulse line / process tap stub below circle -->
    <line :x1="cx" :y1="cy + r" :x2="cx" :y2="cy + r + 10"
          :stroke="stroke" stroke-width="2" stroke-linecap="round" opacity="0.6" style="transition: stroke 0.3s ease"/>

    <!-- Label -->
    <text v-if="config.label" :transform="rotation ? `rotate(${-rotation}, ${w/2}, ${h+13})` : undefined"
          :x="w/2" :y="h + 13" text-anchor="middle" :font-size="fLabel" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
