<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useMonitorStore } from '@/stores/monitor'
import {
  Plug, Radio, Activity, Terminal, Zap
} from 'lucide-vue-next'

const monitor = useMonitorStore()
const route = useRoute()

const nav = [
  { to: '/lines',    label: 'Lines',    icon: Plug     },
  { to: '/signals',  label: 'Signals',  icon: Radio    },
  { to: '/monitor',  label: 'Monitor',  icon: Activity },
  { to: '/commands', label: 'Commands', icon: Terminal  },
]

onMounted(() => monitor.connect())
onUnmounted(() => monitor.disconnect())
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-slate-950">
    <!-- Sidebar -->
    <aside class="w-56 flex-shrink-0 flex flex-col bg-slate-900 border-r border-slate-700">
      <!-- Logo -->
      <div class="flex items-center gap-2 px-4 py-5 border-b border-slate-700">
        <div class="flex items-center justify-center w-8 h-8 rounded-lg bg-blue-600">
          <Zap class="w-4 h-4 text-white" />
        </div>
        <div>
          <p class="text-sm font-bold text-white">GO104</p>
          <p class="text-xs text-slate-500">IEC60870-5-104</p>
        </div>
      </div>

      <!-- WebSocket indicator -->
      <div class="flex items-center gap-2 px-4 py-2 text-xs">
        <span :class="monitor.wsConnected
          ? 'bg-green-500 animate-pulse'
          : 'bg-red-500'"
          class="w-2 h-2 rounded-full"
        />
        <span class="text-slate-400">
          {{ monitor.wsConnected ? 'Live' : 'Reconnecting…' }}
        </span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-2 py-3 space-y-1">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
          :class="route.path === item.to
            ? 'bg-blue-600 text-white'
            : 'text-slate-400 hover:bg-slate-800 hover:text-slate-100'"
        >
          <component :is="item.icon" class="w-4 h-4 flex-shrink-0" />
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="px-4 py-3 border-t border-slate-700 text-xs text-slate-500">
        IEC 60870-5-104 Master
      </div>
    </aside>

    <!-- Main -->
    <main class="flex-1 overflow-y-auto">
      <RouterView />
    </main>
  </div>
</template>
