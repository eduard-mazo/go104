<script setup lang="ts">
import { computed } from 'vue'
import type { ValveConfig, Datapoint } from '@/api/client'

const props = defineProps<{
  config:   ValveConfig
  signal:   Datapoint | null
  selected: boolean
  w:        number
  h:        number
  rotation: 0 | 90 | 180 | 270
}>()

// ON = closed (energized/conducting), OFF = open (tripped)
const state = computed<'on' | 'off' | 'fault' | 'none'>(() => {
  if (!props.signal) return 'none'
  if (props.signal.quality & 0x80) return 'fault'
  return props.signal.value ? 'on' : 'off'
})

const bodyColor = computed(() => ({
  on:    props.config.color_on,
  off:   props.config.color_off,
  fault: props.config.color_fault,
  none:  '#1e293b',
}[state.value]))

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)

// Box bounds
const bx1 = computed(() => props.w * 0.2)
const bx2 = computed(() => props.w * 0.8)
const by1 = computed(() => props.h * 0.1)
const by2 = computed(() => props.h * 0.9)
const bw  = computed(() => bx2.value - bx1.value)
const bh  = computed(() => by2.value - by1.value)

const fScale = computed(() => Math.min(props.w, props.h) / 60)
const fLabel = computed(() => Math.max(7, Math.min(13, Math.round(9  * props.w / 60))))
const fInner = computed(() => Math.max(6, Math.min(12, Math.round(9  * fScale.value))))
const fTag   = computed(() => Math.max(7, Math.min(16, Math.round(11 * fScale.value))))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Left terminal stub -->
    <line :x1="0" :y1="cy" :x2="bx1" :y2="cy" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>
    <!-- Right terminal stub -->
    <line :x1="bx2" :y1="cy" :x2="w" :y2="cy" :stroke="stroke" stroke-width="2" stroke-linecap="round" style="transition: stroke 0.3s ease"/>

    <!-- Body box -->
    <rect :x="bx1" :y="by1" :width="bw" :height="bh" :fill="bodyColor" :stroke="stroke" stroke-width="1.5" rx="2" style="transition: fill 0.3s ease, stroke 0.3s ease"/>

    <!-- ON (closed): horizontal contact line through center -->
    <line v-if="state === 'on'"
          :x1="bx1 + 4" :y1="cy" :x2="bx2 - 4" :y2="cy"
          stroke="white" stroke-width="2" stroke-linecap="round"/>

    <!-- OFF (open): two short stubs + diagonal slash representing open contacts -->
    <template v-else-if="state === 'off'">
      <line :x1="bx1 + 4" :y1="cy" :x2="cx - 4" :y2="cy" stroke="white" stroke-width="2" stroke-linecap="round"/>
      <line :x1="cx + 4" :y1="cy" :x2="bx2 - 4" :y2="cy" stroke="white" stroke-width="2" stroke-linecap="round"/>
      <line :x1="cx - 3" :y1="by1 + 4" :x2="cx + 5" :y2="by2 - 4" stroke="white" stroke-width="1.5" stroke-linecap="round" opacity="0.7"/>
    </template>

    <!-- FAULT: X cross -->
    <template v-else-if="state === 'fault'">
      <line :x1="bx1 + 4" :y1="by1 + 4" :x2="bx2 - 4" :y2="by2 - 4" stroke="white" stroke-width="2" stroke-linecap="round"/>
      <line :x1="bx1 + 4" :y1="by2 - 4" :x2="bx2 - 4" :y2="by1 + 4" stroke="white" stroke-width="2" stroke-linecap="round"/>
    </template>

    <!-- No signal -->
    <text v-if="state === 'none'" :transform="rotation ? `rotate(${-rotation}, ${cx}, ${cy + 1})` : undefined"
          :x="cx" :y="cy + 1" text-anchor="middle" dominant-baseline="middle" :font-size="fInner" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :transform="rotation ? `rotate(${-rotation}, ${w/2}, ${h+13})` : undefined"
          :x="w/2" :y="h + 13" text-anchor="middle" :font-size="fLabel" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
