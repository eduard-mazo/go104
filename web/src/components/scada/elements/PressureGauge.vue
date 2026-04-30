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

// Needle angle: -135° (min) to +135° (max), 0° = pointing right
const ARC_START = -135
const ARC_RANGE = 270

const pct = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return 0
  const range = props.config.max - props.config.min
  if (range === 0) return 0
  return Math.max(0, Math.min(1, (props.signal.value - props.config.min) / range))
})

const needleAngleDeg = computed(() => ARC_START + pct.value * ARC_RANGE)
const needleAngleRad = computed(() => needleAngleDeg.value * Math.PI / 180)

const alarm = computed(() =>
  props.config.alarm_high > 0 &&
  props.signal != null &&
  !(props.signal.quality & 0x80) &&
  props.signal.value >= props.config.alarm_high,
)

const needleColor = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '#374151'
  return alarm.value ? props.config.color_alarm : props.config.color_fill
})

const displayVal = computed(() => {
  if (!props.signal || props.signal.quality & 0x80) return '—'
  return props.signal.value.toFixed(1) + (props.config.unit ? ' ' + props.config.unit : '')
})

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h * 0.52)
const r  = computed(() => Math.min(props.w, props.h * 0.9) / 2 - 3)

// Arc path for background sweep (270° arc)
function describeArc(cx: number, cy: number, r: number, startDeg: number, endDeg: number) {
  const s  = startDeg * Math.PI / 180
  const e  = endDeg   * Math.PI / 180
  const x1 = cx + r * Math.cos(s)
  const y1 = cy + r * Math.sin(s)
  const x2 = cx + r * Math.cos(e)
  const y2 = cy + r * Math.sin(e)
  return `M ${x1} ${y1} A ${r} ${r} 0 1 1 ${x2} ${y2}`
}

// Alarm arc start angle
const alarmStartDeg = computed(() => {
  if (!props.config.alarm_high) return 135
  const range = props.config.max - props.config.min
  if (range === 0) return 135
  const ap = Math.max(0, Math.min(1, (props.config.alarm_high - props.config.min) / range))
  return ARC_START + ap * ARC_RANGE
})
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`"
       class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">

    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6"
          fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Gauge body -->
    <circle :cx="cx" :cy="cy" :r="r"
            fill="#0f172a" stroke="#334155" stroke-width="1.5"/>

    <!-- Normal arc track -->
    <path :d="describeArc(cx, cy, r * 0.75, -135, alarmStartDeg)"
          fill="none" :stroke="config.color_fill" stroke-width="4"
          stroke-linecap="round" opacity="0.25"/>

    <!-- Alarm arc track -->
    <path v-if="config.alarm_high > 0"
          :d="describeArc(cx, cy, r * 0.75, alarmStartDeg, 135)"
          fill="none" :stroke="config.color_alarm" stroke-width="4"
          stroke-linecap="round" opacity="0.25"/>

    <!-- Active arc (fill to needle position) -->
    <path :d="describeArc(cx, cy, r * 0.75, -135, needleAngleDeg)"
          fill="none" :stroke="needleColor" stroke-width="4"
          stroke-linecap="round" opacity="0.8"
          style="transition: stroke 0.35s ease"/>

    <!-- Tick marks at 0%, 25%, 50%, 75%, 100% -->
    <g v-for="t in [0, 0.25, 0.5, 0.75, 1]" :key="t">
      <line
        :x1="cx + r*0.82 * Math.cos((ARC_START + t * ARC_RANGE) * Math.PI / 180)"
        :y1="cy + r*0.82 * Math.sin((ARC_START + t * ARC_RANGE) * Math.PI / 180)"
        :x2="cx + r*0.95 * Math.cos((ARC_START + t * ARC_RANGE) * Math.PI / 180)"
        :y2="cy + r*0.95 * Math.sin((ARC_START + t * ARC_RANGE) * Math.PI / 180)"
        stroke="#475569" stroke-width="1.5" stroke-linecap="round"/>
    </g>

    <!-- Needle -->
    <line
      :x1="cx - r*0.18 * Math.cos(needleAngleRad)"
      :y1="cy - r*0.18 * Math.sin(needleAngleRad)"
      :x2="cx + r*0.7  * Math.cos(needleAngleRad)"
      :y2="cy + r*0.7  * Math.sin(needleAngleRad)"
      :stroke="needleColor" stroke-width="2" stroke-linecap="round"
      style="transition: x2 0.5s ease, y2 0.5s ease, stroke 0.35s ease"/>

    <!-- Center pivot -->
    <circle :cx="cx" :cy="cy" :r="r * 0.09"
            fill="#334155" stroke="#475569" stroke-width="1"/>

    <!-- Value readout -->
    <text :x="cx" :y="cy + r*0.38"
          text-anchor="middle" font-size="9"
          :fill="needleColor" font-family="monospace" font-weight="600"
          style="transition: fill 0.35s ease">{{ displayVal }}</text>

    <!-- Label -->
    <text v-if="config.label" :x="cx" :y="h + 14"
          text-anchor="middle" font-size="9" fill="#94a3b8"
          font-family="monospace" letter-spacing="0.5">{{ config.label }}</text>
  </svg>
</template>
