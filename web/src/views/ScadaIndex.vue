<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useScadaStore } from '@/stores/scada'
import { LayoutDashboard, Plus, Pencil, Eye, Trash2 } from 'lucide-vue-next'

const store   = useScadaStore()
const router  = useRouter()
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
  <div class="p-8 space-y-8 min-h-full">

    <!-- Header -->
    <div>
      <div class="rule-brand mb-4 w-20" />
      <h1 class="text-2xl font-extrabold flex items-center gap-3 tracking-tight">
        <LayoutDashboard class="w-6 h-6 text-[color:var(--epm-citrico)]" />
        SCADA Views
      </h1>
      <p class="text-muted-foreground mt-1 text-sm">
        Build and monitor P&amp;ID diagrams linked to live IEC 104 signals
      </p>
    </div>

    <!-- Create -->
    <div class="flex items-center gap-3">
      <input
        v-model="newName"
        class="input-base w-72"
        placeholder="New view name…"
        @keydown.enter="create"
      />
      <button
        class="btn btn-primary gap-2"
        :disabled="!newName.trim() || creating"
        @click="create"
      >
        <Plus class="w-4 h-4" /> Create
      </button>
    </div>

    <!-- Cards grid -->
    <div v-if="store.views.length" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="v in store.views"
        :key="v.id"
        class="group relative card p-5 transition-colors
               hover:border-[color:var(--epm-citrico)]/40"
      >
        <div class="flex items-start justify-between gap-2">
          <div class="min-w-0">
            <h2 class="font-bold truncate">{{ v.name }}</h2>
            <p class="text-xs text-muted-foreground mt-0.5 font-mono">
              {{ v.width }} × {{ v.height }}
              <span v-if="v.updated_at">
                · {{ new Date(v.updated_at).toLocaleDateString() }}
              </span>
            </p>
          </div>
          <button
            class="opacity-0 group-hover:opacity-100 p-1.5 rounded text-destructive
                   hover:bg-destructive/10 transition-all"
            @click.stop="remove(v.id, v.name)"
          >
            <Trash2 class="w-4 h-4" />
          </button>
        </div>

        <!-- Divider accent -->
        <div class="rule-brand my-3 opacity-40" />

        <!-- Actions -->
        <div class="flex gap-2">
          <button
            class="flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded
                   border border-border text-muted-foreground text-xs font-semibold
                   hover:border-[color:var(--epm-citrico)] hover:text-[color:var(--epm-citrico)]
                   transition-colors"
            @click="router.push(`/scada/${v.id}/design`)"
          >
            <Pencil class="w-3.5 h-3.5" /> Design
          </button>
          <button
            class="flex-1 flex items-center justify-center gap-1.5 py-1.5 rounded
                   border border-border text-muted-foreground text-xs font-semibold
                   hover:border-[color:var(--signal-wait)] hover:text-[color:var(--signal-wait)]
                   transition-colors"
            @click="router.push(`/scada/${v.id}/live`)"
          >
            <Eye class="w-3.5 h-3.5" /> Live
          </button>
        </div>
      </div>
    </div>

    <div v-else class="text-muted-foreground text-sm font-mono mt-8">
      No views yet. Create one above.
    </div>
  </div>
</template>
