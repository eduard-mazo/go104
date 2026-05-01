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

const fill   = computed(() => !props.signal ? '#1e293b' : (props.signal.quality & 0x80) ? props.config.color_fault : props.signal.value ? props.config.color_on : props.config.color_off)
const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')
const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)

const fScale = computed(() => Math.min(props.w, props.h) / 60)
const fLabel = computed(() => Math.max(7, Math.min(13, Math.round(9  * props.w / 60))))
const fInner = computed(() => Math.max(6, Math.min(12, Math.round(9  * fScale.value))))
const fTag   = computed(() => Math.max(7, Math.min(16, Math.round(11 * fScale.value))))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>
    <polygon :points="`0,0 ${cx},${cy} 0,${h}`" :fill="fill" :stroke="stroke" stroke-width="1.5" stroke-linejoin="miter" style="transition:fill .3s ease"/>
    <polygon :points="`${w},0 ${cx},${cy} ${w},${h}`" :fill="fill" :stroke="stroke" stroke-width="1.5" stroke-linejoin="miter" style="transition:fill .3s ease"/>
    <circle :cx="cx" :cy="cy" r="2.5" :fill="stroke"/>
    <line :x1="cx" y1="0" :x2="cx" :y2="-10" :stroke="stroke" stroke-width="1.5" stroke-linecap="square"/>
    <line :x1="cx - 9" :y1="-10" :x2="cx + 9" :y2="-10" :stroke="stroke" stroke-width="2" stroke-linecap="round"/>
    <text v-if="!signal" :transform="rotation ? `rotate(${-rotation}, ${cx}, ${cy + 1})` : undefined"
          :x="cx" :y="cy+1" text-anchor="middle" dominant-baseline="middle" :font-size="Math.max(6, fInner-1)" fill="#6b7280" font-family="monospace">?</text>
    <text v-if="config.label" :transform="rotation ? `rotate(${-rotation}, ${w/2}, ${h+13})` : undefined"
          :x="w/2" :y="h+13" text-anchor="middle" :font-size="fLabel" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
