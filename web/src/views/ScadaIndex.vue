<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useScadaStore } from '@/stores/scada'
import { LayoutDashboard, Plus, Pencil, Eye, Trash2 } from 'lucide-vue-next'

const store  = useScadaStore()
const router = useRouter()
const newName = ref('')
const creating = ref(false)

onMounted(() => store.loadViews())

async function create() {
  if (!newName.value.trim()) return
  creating.value = true
  const v = await store.createView(newName.value.trim())
  newName.value = ''
  creating.value = false
  router.push(`/scada/${v.id}/design`)
}

async function remove(id: number, name: string) {
  if (!confirm(`Delete view "${name}"?`)) return
  await store.deleteView(id)
}
</script>

<template>
  <div class="p-8 space-y-8 min-h-screen bg-[#060a10]">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold text-white flex items-center gap-3 font-mono tracking-tight">
        <LayoutDashboard class="w-6 h-6 text-amber-500" />
        SCADA Views
      </h1>
      <p class="text-slate-500 mt-1 text-sm">Build and monitor process diagrams linked to live IEC104 signals</p>
    </div>

    <!-- Create new -->
    <div class="flex items-center gap-3">
      <input
        v-model="newName"
        class="bg-[#0d1117] border border-slate-700 rounded px-3 py-2 text-white
               placeholder:text-slate-600 focus:outline-none focus:border-amber-600 w-72 text-sm"
        placeholder="New view name…"
        @keydown.enter="create"
      />
      <button
        class="flex items-center gap-2 px-4 py-2 rounded bg-amber-600 hover:bg-amber-500
               text-black font-semibold text-sm transition-colors disabled:opacity-40"
        :disabled="!newName.trim() || creating"
        @click="create"
      >
        <Plus class="w-4 h-4" /> Create
      </button>
    </div>

    <!-- View cards -->
    <div v-if="store.views.length" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="v in store.views" :key="v.id"
        class="group relative bg-[#0d1117] border border-slate-800 rounded-lg p-5
               hover:border-slate-600 transition-colors"
      >
        <!-- Card header -->
        <div class="flex items-start justify-between gap-2">
          <div>
            <h2 class="font-bold text-white font-mono">{{ v.name }}</h2>
            <p class="text-xs text-slate-600 mt-0.5 font-mono">
              {{ v.width }} × {{ v.height }}px
              <span v-if="v.updated_at"> · {{ new Date(v.updated_at).toLocaleDateString() }}</span>
            </p>
          </div>
          <button
            class="opacity-0 group-hover:opacity-100 p-1 rounded text-red-500 hover:text-red-400 transition-all"
            @click.stop="remove(v.id, v.name)"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>

        <!-- Actions -->
        <div class="flex gap-2 mt-4">
          <button
            class="flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded
                   border border-slate-700 text-slate-300 hover:border-amber-600 hover:text-amber-400
                   transition-colors text-xs font-mono"
            @click="router.push(`/scada/${v.id}/design`)"
          >
            <Pencil class="w-3.5 h-3.5" /> Design
          </button>
          <button
            class="flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded
                   border border-slate-700 text-slate-300 hover:border-cyan-600 hover:text-cyan-400
                   transition-colors text-xs font-mono"
            @click="router.push(`/scada/${v.id}/live`)"
          >
            <Eye class="w-3.5 h-3.5" /> Live
          </button>
        </div>
      </div>
    </div>

    <div v-else class="text-slate-600 text-sm font-mono mt-8">
      No views yet. Create one above.
    </div>
  </div>
</template>
