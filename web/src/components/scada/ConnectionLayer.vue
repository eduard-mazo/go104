<script setup lang="ts">
import { computed } from 'vue'
import type { ScadaLine, ScadaElement } from '@/api/client'
import { scadaRegistry } from '@/scada'

const props = defineProps<{
  lines:          ScadaLine[]
  elements:       ScadaElement[]
  width:          number
  height:         number
  selectedLineId: string | null
  design:         boolean   // true = design mode (clickable)
  // optional preview line during connect
  previewFrom:    { x: number; y: number } | null
  previewTo:      { x: number; y: number } | null
}>()

// Index elements by id for O(1) lookup in resolvePoint.
const elementMap = computed(() => {
  const m = new Map<string, ScadaElement>()
  for (const e of props.elements) m.set(e.id, e)
  return m
})

const emit = defineEmits<{
  selectLine: [id: string]
}>()

// ── Style defaults per line type ──────────────────────────────────────────
const STYLE_DEFAULTS: Record<string, { color: string; dash: string; animated: boolean }> = {
  pipe_water: { color: '#22d3ee', dash: '8 3',  animated: true  },
  pipe_gas:   { color: '#f59e0b', dash: '6 4',  animated: true  },
  wire:       { color: '#94a3b8', dash: 'none', animated: false },
  cable:      { color: '#a78bfa', dash: '4 2',  animated: false },
}

function styleFor(l: ScadaLine) {
  const def = STYLE_DEFAULTS[l.style] ?? STYLE_DEFAULTS.pipe_water
  return {
    color:     l.color || def.color,
    dash:      def.dash,
    animated:  def.animated,
    sw:        l.stroke_width || 2,
  }
}

// ── Orthogonal path (go horizontal midpoint, then vertical) ───────────────
function orthoPath(fx: number, fy: number, tx: number, ty: number): string {
  const mx = (fx + tx) / 2
  return `M ${fx} ${fy} L ${mx} ${fy} L ${mx} ${ty} L ${tx} ${ty}`
}

// Resolve endpoint live: prefer port lookup so connections track moving elements.
// Falls back to saved absolute pt for legacy lines (no port_id) and free endpoints.
function resolvePoint(
  elId:   string | null,
  portId: string | undefined,
  pt:     { x: number; y: number },
): { x: number; y: number } {
  if (elId && portId) {
    const el = elementMap.value.get(elId)
    if (el) {
      const p = scadaRegistry.portPoint(el, portId)
      if (p) return p
    }
  }
  return pt
}

const resolvedLines = computed(() =>
  props.lines.map(l => ({
    ...l,
    fp: resolvePoint(l.from_el, l.from_port, l.from_pt),
    tp: resolvePoint(l.to_el,   l.to_port,   l.to_pt),
  }))
)
</script>

<template>
  <svg
    class="absolute inset-0 pointer-events-none overflow-visible"
    :width="width"
    :height="height"
    :viewBox="`0 0 ${width} ${height}`"
    xmlns="http://www.w3.org/2000/svg"
  >
    <defs>
      <style>
        @keyframes flow-fwd {
          to { stroke-dashoffset: -22; }
        }
        .line-animated {
          animation: flow-fwd 0.9s linear infinite;
        }
      </style>
    </defs>

    <!-- Rendered lines -->
    <g v-for="l in resolvedLines" :key="l.id">
      <!-- Wide transparent hit-area (design mode only) -->
      <path
        v-if="design"
        :d="orthoPath(l.fp.x, l.fp.y, l.tp.x, l.tp.y)"
        fill="none" stroke="transparent" stroke-width="14"
        style="pointer-events: stroke; cursor: pointer"
        @click.stop="emit('selectLine', l.id)"
      />

      <!-- Actual line -->
      <path
        :d="orthoPath(l.fp.x, l.fp.y, l.tp.x, l.tp.y)"
        fill="none"
        :stroke="styleFor(l).color"
        :stroke-width="styleFor(l).sw"
        :stroke-dasharray="styleFor(l).dash === 'none' ? undefined : styleFor(l).dash"
        :class="styleFor(l).animated ? 'line-animated' : ''"
        stroke-linecap="round" stroke-linejoin="round"
      />

      <!-- Selection highlight -->
      <path
        v-if="selectedLineId === l.id"
        :d="orthoPath(l.fp.x, l.fp.y, l.tp.x, l.tp.y)"
        fill="none" stroke="#f59e0b" stroke-width="3"
        stroke-dasharray="5 3" opacity="0.8"
        stroke-linecap="round"
      />

      <!-- Endpoint dots (design, selected) -->
      <template v-if="design && selectedLineId === l.id">
        <circle :cx="l.fp.x" :cy="l.fp.y" r="4"
                fill="#f59e0b" stroke="#0a0e14" stroke-width="1.5"/>
        <circle :cx="l.tp.x" :cy="l.tp.y" r="4"
                fill="#f59e0b" stroke="#0a0e14" stroke-width="1.5"/>
      </template>

      <!-- Line label -->
      <text
        v-if="l.label"
        :x="(l.fp.x + l.tp.x) / 2"
        :y="(l.fp.y + l.tp.y) / 2 - 5"
        text-anchor="middle" font-size="9"
        :fill="styleFor(l).color"
        font-family="monospace" opacity="0.9"
      >{{ l.label }}</text>
    </g>

    <!-- Connect-mode preview line -->
    <path
      v-if="previewFrom && previewTo"
      :d="orthoPath(previewFrom.x, previewFrom.y, previewTo.x, previewTo.y)"
      fill="none" stroke="#3b82f6" stroke-width="2"
      stroke-dasharray="6 3" opacity="0.7"
      stroke-linecap="round"
    />
  </svg>
</template>
