import type { Component } from 'vue'
import type { ScadaElement } from '@/api/client'

export type ElementGroup =
  | 'fluid'
  | 'gas'
  | 'electrical'
  | 'instrument'
  | 'display'
  | 'custom'

export interface PortDef {
  id: string
  rx: number   // 0–1 relative to element width
  ry: number   // 0–1 relative to element height
}

export interface PaletteDef {
  pw: number
  ph: number
  vb: string            // SVG viewBox
  svg: string           // trusted inner markup (rendered with v-html)
}

export interface ElementDescriptor {
  kind: ScadaElement['kind'] | string
  label: string
  group: ElementGroup
  component: Component                          // pass through markRaw() at definition site
  defaultConfig(): Record<string, unknown>
  defaultSize(): { w: number; h: number }
  ports: PortDef[]
  palette: PaletteDef
}
