import { defineStore } from 'pinia'
import { ref } from 'vue'
import { scadaAPI, type ScadaView, type ScadaElement, type ScadaLine } from '@/api/client'
import { generateUUID } from '@/lib/utils'

export const useScadaStore = defineStore('scada', () => {
  const views      = ref<ScadaView[]>([])
  const activeView = ref<ScadaView | null>(null)
  const saving     = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')

  let saveTimer: ReturnType<typeof setTimeout> | null = null

  async function loadViews() {
    views.value = await scadaAPI.list()
  }

  async function loadView(id: number) {
    const v = await scadaAPI.get(id)
    if (v && !v.lines) v.lines = []
    activeView.value = v
  }

  async function createView(name: string) {
    const v = await scadaAPI.create({ name, width: 1400, height: 900, elements: [], lines: [] })
    views.value.push(v)
    return v
  }

  async function deleteView(id: number) {
    await scadaAPI.remove(id)
    views.value = views.value.filter(v => v.id !== id)
    if (activeView.value?.id === id) activeView.value = null
  }

  function scheduleSave() {
    if (!activeView.value) return
    saving.value = 'saving'
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(async () => {
      if (!activeView.value) return
      try {
        await scadaAPI.update(activeView.value.id, {
          name:     activeView.value.name,
          width:    activeView.value.width,
          height:   activeView.value.height,
          elements: activeView.value.elements,
          lines:    activeView.value.lines ?? [],
        })
        saving.value = 'saved'
        setTimeout(() => { saving.value = 'idle' }, 2000)
      } catch {
        saving.value = 'error'
      }
    }, 600)
  }

  // ── Elements ──────────────────────────────────────────────────────────────

  function addElement(el: ScadaElement) {
    activeView.value?.elements.push(el)
    scheduleSave()
  }

  function updateElement(el: ScadaElement) {
    if (!activeView.value) return
    const idx = activeView.value.elements.findIndex(e => e.id === el.id)
    if (idx !== -1) activeView.value.elements[idx] = el
    scheduleSave()
  }

  function removeElement(id: string) {
    if (!activeView.value) return
    activeView.value.elements = activeView.value.elements.filter(e => e.id !== id)
    // Also remove lines attached to this element
    activeView.value.lines = (activeView.value.lines ?? []).filter(
      l => l.from_el !== id && l.to_el !== id,
    )
    scheduleSave()
  }

  function duplicateElement(id: string) {
    if (!activeView.value) return
    const src = activeView.value.elements.find(e => e.id === id)
    if (!src) return
    const clone: ScadaElement = {
      ...src,
      id: generateUUID(),
      x:  src.x + 24,
      y:  src.y + 24,
      config: { ...src.config },
    }
    activeView.value.elements.push(clone)
    scheduleSave()
    return clone.id
  }

  // ── Lines ─────────────────────────────────────────────────────────────────

  function addLine(line: ScadaLine) {
    if (!activeView.value) return
    if (!activeView.value.lines) activeView.value.lines = []
    activeView.value.lines.push(line)
    scheduleSave()
  }

  function updateLine(line: ScadaLine) {
    if (!activeView.value?.lines) return
    const idx = activeView.value.lines.findIndex(l => l.id === line.id)
    if (idx !== -1) activeView.value.lines[idx] = line
    scheduleSave()
  }

  function removeLine(id: string) {
    if (!activeView.value) return
    activeView.value.lines = (activeView.value.lines ?? []).filter(l => l.id !== id)
    scheduleSave()
  }

  return {
    views, activeView, saving,
    loadViews, loadView, createView, deleteView,
    addElement, updateElement, removeElement, duplicateElement,
    addLine, updateLine, removeLine,
    scheduleSave,
  }
})
