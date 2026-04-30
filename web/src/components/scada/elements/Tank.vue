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

const pct = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return 0
  const range = props.config.max - props.config.min
  if (range === 0) return 0
  return Math.max(0, Math.min(1, (props.signal.value - props.config.min) / range))
})

const alarm = computed(() =>
  props.config.alarm_high > 0 &&
  props.signal != null &&
  !(props.signal.quality & 0x80) &&
  props.signal.value >= props.config.alarm_high,
)

const fillColor = computed(() => alarm.value ? props.config.color_alarm : props.config.color_fill)

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

// Layout
const PAD   = 4
const TX    = computed(() => PAD)
const TY    = computed(() => PAD)
const TW    = computed(() => props.w - PAD * 2)
const BODY  = computed(() => props.h - PAD * 2 - 18) // leave room for label
const fillH = computed(() => BODY.value * pct.value)
const fillY = computed(() => TY.value + BODY.value - fillH.value)

// Alarm level line Y
const alarmY = computed(() => {
  if (!props.config.alarm_high || props.config.alarm_high === 0) return null
  const range = props.config.max - props.config.min
  if (range === 0) return null
  const ap = Math.max(0, Math.min(1, (props.config.alarm_high - props.config.min) / range))
  return TY.value + BODY.value * (1 - ap)
})
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Tank body background -->
    <rect :x="TX" :y="TY" :width="TW" :height="BODY"
          fill="#1e293b" stroke="#475569" stroke-width="1.5" rx="2"/>

    <!-- Fill level (animated) -->
    <rect :x="TX + 1" :y="fillY" :width="TW - 2" :height="fillH"
          :fill="fillColor" rx="1"
          style="transition: y 0.5s ease, height 0.5s ease, fill 0.35s ease"/>

    <!-- Alarm line -->
    <line v-if="alarmY !== null"
          :x1="TX" :y1="alarmY" :x2="TX + TW" :y2="alarmY"
          :stroke="config.color_alarm" stroke-width="1" stroke-dasharray="3 2" opacity="0.8"/>

    <!-- Tank border overlay (above fill) -->
    <rect :x="TX" :y="TY" :width="TW" :height="BODY"
          fill="none" stroke="#475569" stroke-width="1.5" rx="2"/>

    <!-- Dome cap (top) -->
    <path :d="`M ${TX + 2} ${TY + 4} Q ${TX + TW/2} ${TY - 6} ${TX + TW - 2} ${TY + 4}`"
          fill="none" stroke="#475569" stroke-width="1.5"/>

    <!-- Value text -->
    <text :x="w/2" :y="TY + BODY + 12"
          text-anchor="middle" font-size="9"
          :fill="signal && !(signal.quality & 0x80) ? fillColor : '#6b7280'"
          font-family="monospace" font-weight="600"
          style="transition: fill 0.35s ease">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :x="w/2" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
