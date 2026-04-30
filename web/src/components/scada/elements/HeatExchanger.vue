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
  props.signal.value >= props.config.alarm_high
)

const indicatorColor = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '#64748b'
  return alarm.value ? props.config.color_alarm : props.config.color_fill
})

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

// Shell rectangle
const PAD   = 4
const LABEL_H = 14
const sx  = computed(() => PAD)
const sy  = computed(() => PAD)
const sw  = computed(() => props.w - PAD * 2)
const sh  = computed(() => props.h - PAD * 2 - LABEL_H)

// Shell entry/exit stubs at h*0.35
const shellY = computed(() => sy.value + sh.value * 0.35)
// Tube entry/exit stubs at h*0.65
const tubeY  = computed(() => sy.value + sh.value * 0.65)

// Tube bundle S-curve path inside the shell
// Uses a sinusoidal approximation with cubic bezier
const tubePathD = computed(() => {
  const x1 = sx.value + 6
  const x2 = sx.value + sw.value - 6
  const midX = sx.value + sw.value / 2
  const y0 = sy.value + sh.value * 0.28
  const y1 = sy.value + sh.value * 0.5
  const y2 = sy.value + sh.value * 0.72
  return `M ${x1} ${y0} C ${midX * 0.6} ${y0}, ${midX * 0.6} ${y1}, ${midX} ${y1} S ${x2 * 1.1} ${y2} ${x2} ${y2}`
})
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Shell rectangle body -->
    <rect :x="sx" :y="sy" :width="sw" :height="sh" fill="#0f1923" :stroke="stroke" stroke-width="1.5" rx="3" style="transition: stroke 0.3s ease"/>

    <!-- Tube bundle S-curve (representing tube passes) -->
    <path :d="tubePathD" fill="none" :stroke="indicatorColor" stroke-width="1.5" stroke-linecap="round" opacity="0.8" style="transition: stroke 0.3s ease"/>

    <!-- Shell side stubs (at 35% height): left and right -->
    <line :x1="0" :y1="shellY" :x2="sx" :y2="shellY" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <line :x1="sx + sw" :y1="shellY" :x2="w" :y2="shellY" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- Tube side stubs (at 65% height): left and right -->
    <line :x1="0" :y1="tubeY" :x2="sx" :y2="tubeY" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <line :x1="sx + sw" :y1="tubeY" :x2="w" :y2="tubeY" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- HE tag inside top-left -->
    <text :x="sx + 4" :y="sy + 9" font-size="7" fill="#64748b" font-family="monospace" opacity="0.8">HE</text>

    <!-- Value text below shell -->
    <text :x="w / 2" :y="sy + sh + 12"
          text-anchor="middle" font-size="9"
          :fill="indicatorColor" font-family="monospace" font-weight="600"
          style="transition: fill 0.3s ease">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :x="w / 2" :y="h + 13" text-anchor="middle" font-size="9" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
