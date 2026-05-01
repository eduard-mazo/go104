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

const pct = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return null
  const range = props.config.max - props.config.min
  if (range === 0) return 0
  return Math.max(0, Math.min(1, (props.signal.value - props.config.min) / range))
})

const alarm = computed(() =>
  props.config.alarm_high > 0 &&
  props.signal != null &&
  !(props.signal.quality & 0x80) &&
  props.signal.value >= props.config.alarm_high
)

const fillColor = computed(() =>
  !props.signal || (props.signal.quality & 0x80) ? '#64748b'
  : alarm.value ? props.config.color_alarm : props.config.color_fill
)

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

// Layout constants
const PAD  = 4
const LABEL_H = 16

// Body rectangle (leaves room at bottom for value text)
const bx  = computed(() => PAD)
const by  = computed(() => PAD)
const bw  = computed(() => props.w - PAD * 2)
const bh  = computed(() => props.h - PAD * 2 - LABEL_H)

// Level line Y — computed from pct, top of tank = by, bottom = by+bh
// pct=1 => top, pct=0 => bottom
const levelY = computed(() => {
  if (pct.value === null) return by.value + bh.value  // bottom (no signal)
  return by.value + bh.value * (1 - pct.value)
})

// Alarm level line Y
const alarmY = computed(() => {
  if (!props.config.alarm_high || props.config.alarm_high === 0) return null
  const range = props.config.max - props.config.min
  if (range === 0) return null
  const ap = Math.max(0, Math.min(1, (props.config.alarm_high - props.config.min) / range))
  return by.value + bh.value * (1 - ap)
})

// Nozzle stubs: short horizontal lines at mid-height from sides
const nozzleY = computed(() => by.value + bh.value * 0.65)

// Dome arc: top edge arc
const domeD = computed(() => {
  const x1 = bx.value + 2
  const x2 = bx.value + bw.value - 2
  const y  = by.value
  const mx = bx.value + bw.value / 2
  const my = y - 6
  return `M ${x1} ${y} Q ${mx} ${my} ${x2} ${y}`
})

const fScale = computed(() => Math.min(props.w, props.h) / 60)
const fLabel = computed(() => Math.max(7, Math.min(13, Math.round(9  * props.w / 60))))
const fInner = computed(() => Math.max(6, Math.min(12, Math.round(9  * fScale.value))))
const fTag   = computed(() => Math.max(7, Math.min(16, Math.round(11 * fScale.value))))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Tank body background -->
    <rect :x="bx" :y="by" :width="bw" :height="bh" fill="#0f1923" :stroke="stroke" stroke-width="1.5" rx="2" style="transition: stroke 0.3s ease"/>

    <!-- Dome cap at top -->
    <path :d="domeD" fill="none" :stroke="stroke" stroke-width="1.5" style="transition: stroke 0.3s ease"/>

    <!-- Nozzle stubs (sides at ~65% height) -->
    <line :x1="0" :y1="nozzleY" :x2="bx" :y2="nozzleY" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <line :x1="bx + bw" :y1="nozzleY" :x2="w" :y2="nozzleY" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- Alarm level: dashed horizontal line -->
    <line v-if="alarmY !== null"
          :x1="bx + 2" :y1="alarmY" :x2="bx + bw - 2" :y2="alarmY"
          :stroke="config.color_alarm" stroke-width="1" stroke-dasharray="3 2" opacity="0.8"/>

    <!-- Level line: solid horizontal line at current level -->
    <line v-if="pct !== null"
          :x1="bx + 2" :y1="levelY" :x2="bx + bw - 2" :y2="levelY"
          :stroke="fillColor" stroke-width="2" stroke-linecap="round"
          style="transition: y1 0.5s ease, y2 0.5s ease, stroke 0.3s ease"/>

    <!-- Tank border overlay (above content) -->
    <rect :x="bx" :y="by" :width="bw" :height="bh" fill="none" :stroke="stroke" stroke-width="1.5" rx="2" style="transition: stroke 0.3s ease"/>

    <!-- "LT" tag at top-left inside tank -->
    <text :transform="rotation ? `rotate(${-rotation}, ${bx + 4}, ${by + 9})` : undefined"
          :x="bx + 4" :y="by + 9" :font-size="Math.max(6, fInner-1)" fill="#64748b" font-family="monospace" opacity="0.7">LT</text>

    <!-- Value text at bottom of tank body -->
    <text :transform="rotation ? `rotate(${-rotation}, ${w / 2}, ${by + bh + 12})` : undefined"
          :x="w / 2" :y="by + bh + 12"
          text-anchor="middle" :font-size="fInner"
          :fill="signal && !(signal.quality & 0x80) ? fillColor : '#64748b'"
          font-family="monospace" font-weight="600"
          style="transition: fill 0.3s ease">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :transform="rotation ? `rotate(${-rotation}, ${w/2}, ${h+13})` : undefined"
          :x="w/2" :y="h + 13" text-anchor="middle" :font-size="fLabel" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
