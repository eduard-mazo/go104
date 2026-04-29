<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

export interface MenuItem {
  label:    string
  icon?:    string
  danger?:  boolean
  divider?: boolean
  action:   () => void
}

const props = defineProps<{
  x:     number
  y:     number
  items: MenuItem[]
}>()
const emit = defineEmits<{ close: [] }>()

function close() { emit('close') }

function handleClick(item: MenuItem) {
  item.action()
  close()
}

// Close on any outside click or Escape
function onKey(e: KeyboardEvent) { if (e.key === 'Escape') close() }
function onMouseDown(e: MouseEvent) {
  const menu = document.getElementById('ctx-menu')
  if (menu && !menu.contains(e.target as Node)) close()
}
onMounted(() => {
  setTimeout(() => {
    window.addEventListener('mousedown', onMouseDown)
    window.addEventListener('keydown', onKey)
  }, 0)
})
onUnmounted(() => {
  window.removeEventListener('mousedown', onMouseDown)
  window.removeEventListener('keydown', onKey)
})
</script>

<template>
  <Teleport to="body">
    <div
      id="ctx-menu"
      class="fixed z-[999] min-w-[180px] py-1 rounded border border-slate-700
             bg-[#0d1117] shadow-2xl shadow-black/60 text-sm select-none"
      :style="{ left: x + 'px', top: y + 'px' }"
    >
      <template v-for="(item, i) in items" :key="i">
        <div v-if="item.divider" class="my-1 border-t border-slate-800" />
        <button
          v-else
          class="w-full flex items-center gap-2 px-3 py-1.5 text-left transition-colors"
          :class="item.danger
            ? 'text-red-400 hover:bg-red-950/40 hover:text-red-300'
            : 'text-slate-300 hover:bg-slate-800 hover:text-white'"
          @click.stop="handleClick(item)"
        >
          <span v-if="item.icon" class="text-base w-4 text-center opacity-70">{{ item.icon }}</span>
          {{ item.label }}
        </button>
      </template>
    </div>
  </Teleport>
</template>
