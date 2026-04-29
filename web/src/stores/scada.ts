import { defineStore } from 'pinia'
import { ref } from 'vue'
import { scadaAPI, type ScadaView, type ScadaElement } from '@/api/client'

export const useScadaStore = defineStore('scada', () => {
  const views      = ref<ScadaView[]>([])
  const activeView = ref<ScadaView | null>(null)
  const saving     = ref<'idle' | 'saving' | 'saved' | 'error'>('idle')

  let saveTimer: ReturnType<typeof setTimeout> | null = null

  async function loadViews() {
    views.value = await scadaAPI.list()
  }

  async function loadView(id: number) {
    activeView.value = await scadaAPI.get(id)
  }

  async function createView(name: string) {
    const v = await scadaAPI.create({ name, width: 1400, height: 900, elements: [] })
    views.value.push(v)
    return v
  }

  async function deleteView(id: number) {
    await scadaAPI.remove(id)
    views.value = views.value.filter(v => v.id !== id)
    if (activeView.value?.id === id) activeView.value = null
  }

  // Debounced autosave — called after any element mutation in the design editor
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
        })
        saving.value = 'saved'
        setTimeout(() => { saving.value = 'idle' }, 2000)
      } catch {
        saving.value = 'error'
      }
    }, 600)
  }

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
    scheduleSave()
  }

  function duplicateElement(id: string) {
    if (!activeView.value) return
    const src = activeView.value.elements.find(e => e.id === id)
    if (!src) return
    const clone: ScadaElement = {
      ...src,
      id: crypto.randomUUID(),
      x:  src.x + 24,
      y:  src.y + 24,
      config: { ...src.config },
    }
    activeView.value.elements.push(clone)
    scheduleSave()
    return clone.id
  }

  return {
    views, activeView, saving,
    loadViews, loadView, createView, deleteView,
    addElement, updateElement, removeElement, duplicateElement,
    scheduleSave,
  }
})
