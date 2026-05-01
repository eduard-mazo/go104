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

const coilColor = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '#64748b'
  return alarm.value ? props.config.color_alarm : props.config.color_fill
})

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

const cx  = computed(() => props.w / 2)
const cy  = computed(() => props.h / 2 - 4)
// Each coil circle radius — circles must fit within w, leaving some margin
const cr  = computed(() => Math.min(props.w * 0.21, props.h * 0.3, 20))
const gap = 4
// Left/right centers
const lx  = computed(() => cx.value - cr.value - gap / 2)
const rx  = computed(() => cx.value + cr.value + gap / 2)

// Three horizontal lines inside each circle representing winding symbol
function windingLines(centX: number, centY: number, radius: number) {
  const hw = radius * 0.5
  const sp = radius * 0.28
  return [
    { x1: centX - hw, y1: centY - sp, x2: centX + hw, y2: centY - sp },
    { x1: centX - hw, y1: centY,      x2: centX + hw, y2: centY      },
    { x1: centX - hw, y1: centY + sp, x2: centX + hw, y2: centY + sp },
  ]
}

const fScale = computed(() => Math.min(props.w, props.h) / 60)
const fLabel = computed(() => Math.max(7, Math.min(13, Math.round(9  * props.w / 60))))
const fInner = computed(() => Math.max(6, Math.min(12, Math.round(9  * fScale.value))))
const fTag   = computed(() => Math.max(7, Math.min(16, Math.round(11 * fScale.value))))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Primary connection lines (top) -->
    <line :x1="lx" :y1="cy - cr - 1" :x2="lx" :y2="0" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <line :x1="rx" :y1="cy - cr - 1" :x2="rx" :y2="0" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- Left coil circle -->
    <circle :cx="lx" :cy="cy" :r="cr" fill="#0f1923" :stroke="coilColor" stroke-width="1.5" style="transition: stroke 0.3s ease"/>

    <!-- Winding lines — left -->
    <line v-for="(l, i) in windingLines(lx, cy, cr)" :key="`l${i}`"
          :x1="l.x1" :y1="l.y1" :x2="l.x2" :y2="l.y2"
          :stroke="coilColor" stroke-width="1.2" stroke-linecap="round" opacity="0.7"
          style="transition: stroke 0.3s ease"/>

    <!-- Right coil circle -->
    <circle :cx="rx" :cy="cy" :r="cr" fill="#0f1923" :stroke="coilColor" stroke-width="1.5" style="transition: stroke 0.3s ease"/>

    <!-- Winding lines — right -->
    <line v-for="(l, i) in windingLines(rx, cy, cr)" :key="`r${i}`"
          :x1="l.x1" :y1="l.y1" :x2="l.x2" :y2="l.y2"
          :stroke="coilColor" stroke-width="1.2" stroke-linecap="round" opacity="0.7"
          style="transition: stroke 0.3s ease"/>

    <!-- Secondary connection lines (bottom) -->
    <line :x1="lx" :y1="cy + cr + 1" :x2="lx" :y2="h - 16" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <line :x1="rx" :y1="cy + cr + 1" :x2="rx" :y2="h - 16" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- Value text -->
    <text :transform="rotation ? `rotate(${-rotation}, ${cx}, ${h - 4})` : undefined"
          :x="cx" :y="h - 4"
          text-anchor="middle" :font-size="fInner"
          :fill="coilColor" font-family="monospace" font-weight="600"
          style="transition: fill 0.3s ease">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :transform="rotation ? `rotate(${-rotation}, ${w/2}, ${h+13})` : undefined"
          :x="w/2" :y="h + 13" text-anchor="middle" :font-size="fLabel" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
