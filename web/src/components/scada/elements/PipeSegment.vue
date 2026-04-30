<script setup lang="ts">
import { computed } from 'vue'
import type { PipeConfig } from '@/api/client'

const props = defineProps<{
  config:   PipeConfig
  signal:   null          // PipeSegment has no signal binding
  selected: boolean
  w:        number
  h:        number
}>()

const stroke = computed(() => props.selected ? '#f59e0b' : '#64748b')
const pipeColor = computed(() => props.config.color || '#64748b')

const strokeWidth = computed(() => {
  switch (props.config.style) {
    case 'process':    return 4
    case 'utility':    return 2
    case 'instrument': return 1
    default:           return 2
  }
})

const cx = computed(() => props.w / 2)
const cy = computed(() => props.h / 2)
</script>

<template>
  <svg :width="w" :height="h" :viewBox="`0 0 ${w} ${h}`" class="overflow-visible block" xmlns="http://www.w3.org/2000/svg">
    <rect v-if="selected" x="-3" y="-3" :width="w+6" :height="h+6" fill="none" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="4 2" rx="2"/>

    <!-- Main pipe line -->
    <line x1="0" :y1="cy" :x2="w" :y2="cy"
          :stroke="pipeColor" :stroke-width="strokeWidth" stroke-linecap="round"/>

    <!-- Parallel double line for insulated pipes -->
    <template v-if="config.style === 'process'">
      <line x1="0" :y1="cy - 3" :x2="w" :y2="cy - 3"
            :stroke="pipeColor" stroke-width="1" stroke-linecap="round" opacity="0.35"/>
      <line x1="0" :y1="cy + 3" :x2="w" :y2="cy + 3"
            :stroke="pipeColor" stroke-width="1" stroke-linecap="round" opacity="0.35"/>
    </template>

    <!-- Selection stroke overlay (thin colored border along pipe) -->
    <line v-if="selected" x1="0" :y1="cy" :x2="w" :y2="cy"
          stroke="#f59e0b" :stroke-width="strokeWidth + 2" stroke-linecap="round" opacity="0.3"/>

    <!-- Label centered above pipe -->
    <text v-if="config.label" :x="cx" :y="cy - strokeWidth - 4"
          text-anchor="middle" font-size="9" fill="#94a3b8" font-family="monospace">{{ config.label }}</text>
  </svg>
</template>
