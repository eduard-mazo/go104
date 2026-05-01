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

// ON = running
const fill = computed(() =>
  !props.signal ? '#1e293b'
  : (props.signal.quality & 0x80) ? props.config.color_fault
  : props.signal.value ? props.config.color_on : props.config.color_off
)

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')

const running = computed(() =>
  props.signal && !(props.signal.quality & 0x80) && props.signal.value
)

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
const r  = computed(() => Math.min(props.w, props.h) / 2 - 2)

// Circumference for spinning dashed ring
const ringR  = computed(() => r.value + 3)
const ringC  = computed(() => 2 * Math.PI * ringR.value)
const dashOn = computed(() => ringC.value * 0.2)
const dashOff= computed(() => ringC.value * 0.8)

const fScale = computed(() => Math.min(props.w, props.h) / 60)
const fLabel = computed(() => Math.max(7, Math.min(13, Math.round(9  * props.w / 60))))
const fInner = computed(() => Math.max(6, Math.min(12, Math.round(9  * fScale.value))))
const fTag   = computed(() => Math.max(7, Math.min(16, Math.round(11 * fScale.value))))
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Animated running ring -->
    <circle v-if="running"
            :cx="cx" :cy="cy" :r="ringR"
            fill="none" :stroke="fill" stroke-width="2" opacity="0.45"
            :stroke-dasharray="`${dashOn} ${dashOff}`"
            style="transition: stroke 0.3s ease">
      <animateTransform attributeName="transform" type="rotate"
                        :from="`0 ${cx} ${cy}`" :to="`360 ${cx} ${cy}`"
                        dur="1s" repeatCount="indefinite"/>
    </circle>

    <!-- Pump body circle -->
    <circle :cx="cx" :cy="cy" :r="r" :fill="fill" :stroke="stroke" stroke-width="1.5" style="transition: fill 0.3s ease, stroke 0.3s ease"/>

    <!-- Impeller triangle: apex at center top, base at bottom (centrifugal discharge upward) -->
    <!-- points: center, bottom-left, bottom-right -->
    <polygon :points="`${cx},${cy - r * 0.5} ${cx - r * 0.7},${cy + r * 0.65} ${cx + r * 0.7},${cy + r * 0.65}`"
             fill="none" :stroke="stroke" stroke-width="1.5" stroke-linejoin="round" style="transition: stroke 0.3s ease"/>

    <!-- Center hub dot -->
    <circle :cx="cx" :cy="cy - r * 0.5" :r="r * 0.1" :fill="stroke" style="transition: fill 0.3s ease"/>

    <!-- No signal indicator -->
    <text v-if="!signal" :transform="rotation ? `rotate(${-rotation}, ${cx}, ${cy + 1})` : undefined"
          :x="cx" :y="cy + 1" text-anchor="middle" dominant-baseline="middle" :font-size="Math.max(6, fInner-1)" fill="#6b7280" font-family="monospace">?</text>

    <!-- Label -->
    <text v-if="config.label" :transform="rotation ? `rotate(${-rotation}, ${w/2}, ${h+13})` : undefined"
          :x="w/2" :y="h + 13" text-anchor="middle" :font-size="fLabel" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
