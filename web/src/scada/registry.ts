import type { ElementDescriptor, ElementGroup, PortDef } from './types'
import type { ScadaElement } from '@/api/client'

class ElementRegistry {
  private _map = new Map<string, ElementDescriptor>()

  register(items: ElementDescriptor | ElementDescriptor[]): this {
    const list = Array.isArray(items) ? items : [items]
    for (const d of list) {
      if (this._map.has(d.kind)) {
        console.warn(`[scada] overriding element kind "${d.kind}"`)
      }
      this._map.set(d.kind, d)
    }
    return this
  }

  get(kind: string): ElementDescriptor | undefined {
    return this._map.get(kind)
  }

  component(kind: string) {
    return this._map.get(kind)?.component
  }

  all(): ElementDescriptor[] {
    return [...this._map.values()]
  }

  byGroup(g: ElementGroup): ElementDescriptor[] {
    return this.all().filter(d => d.group === g)
  }

  groups(): ElementGroup[] {
    const seen = new Set<ElementGroup>()
    for (const d of this._map.values()) seen.add(d.group)
    return [...seen]
  }

  // Convenience: absolute canvas port coords for a placed element
  portsFor(el: ScadaElement): Array<{ id: string; x: number; y: number }> {
    const defs: PortDef[] = this._map.get(el.kind)?.ports ?? []
    return defs.map(d => ({
      id: d.id,
      x:  el.x + d.rx * el.w,
      y:  el.y + d.ry * el.h,
    }))
  }

  // Absolute coords for a specific named port on an element (live tracking).
  portPoint(el: ScadaElement, portId: string): { x: number; y: number } | null {
    const def = this._map.get(el.kind)?.ports.find(p => p.id === portId)
    if (!def) return null
    return { x: el.x + def.rx * el.w, y: el.y + def.ry * el.h }
  }
}

export const scadaRegistry = new ElementRegistry()
export type { ElementRegistry }
