<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useMonitorStore } from '@/stores/monitor'
import {
  Plug, Radio, Activity, Terminal, LayoutDashboard,
  Zap, PanelLeftClose, PanelLeftOpen, Menu, X,
} from 'lucide-vue-next'

const monitor = useMonitorStore()
const route   = useRoute()

const collapsed  = ref(false)
const mobileOpen = ref(false)

onMounted(() => {
  collapsed.value = localStorage.getItem('go104:sb-collapsed') === '1'
  monitor.connect()
})
onUnmounted(() => monitor.disconnect())

function toggleSidebar() {
  collapsed.value = !collapsed.value
  localStorage.setItem('go104:sb-collapsed', collapsed.value ? '1' : '0')
}

const nav = [
  { to: '/lines',    label: 'Lines',    icon: Plug,            },
  { to: '/signals',  label: 'Signals',  icon: Radio,           },
  { to: '/monitor',  label: 'Monitor',  icon: Activity,        },
  { to: '/commands', label: 'Commands', icon: Terminal,        },
  { to: '/scada',    label: 'SCADA',    icon: LayoutDashboard, },
]
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background text-foreground">

    <!-- Mobile backdrop -->
    <div
      v-if="mobileOpen"
      class="fixed inset-0 z-40 bg-black/60 md:hidden"
      @click="mobileOpen = false"
    />

    <!-- ── SIDEBAR ──────────────────────────────────────────────────────── -->
    <aside
      class="bg-sidebar text-sidebar-foreground border-r border-sidebar-border
             flex flex-col h-screen overflow-hidden shrink-0
             transition-[width,transform] duration-200 ease-out
             fixed inset-y-0 left-0 z-50 md:static md:z-auto"
      :class="[
        collapsed ? 'w-[60px] sidebar-mini' : 'w-[220px]',
        mobileOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
      ]"
    >
      <!-- Brand -->
      <div class="flex items-center gap-3 px-4 h-[60px] border-b border-sidebar-border shrink-0">
        <div class="grid place-items-center w-8 h-8 rounded-sm shrink-0 select-none font-black text-base
                    bg-[color:var(--epm-citrico)] text-[color:var(--epm-bosque)]">
          G
        </div>
        <div class="sidebar-wide-only leading-none min-w-0 flex-1">
          <div class="text-[16px] font-extrabold tracking-tight text-white truncate">GO104</div>
          <div class="text-[9px] uppercase tracking-[0.22em] text-[color:var(--epm-citrico)] mt-0.5 font-semibold">
            IEC 60870-5-104
          </div>
        </div>
        <!-- Close button (mobile only) -->
        <button
          class="md:hidden grid place-items-center w-7 h-7 rounded-sm
                 hover:bg-sidebar-accent text-sidebar-foreground/60"
          @click="mobileOpen = false"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- WebSocket status -->
      <div class="flex items-center gap-2 px-4 py-2 border-b border-sidebar-border/50 shrink-0">
        <span
          class="status-dot"
          :class="monitor.wsConnected
            ? 'text-[color:var(--epm-citrico)]'
            : 'text-[color:var(--signal-fault)]'"
        />
        <span
          class="sidebar-wide-only text-[11px] font-semibold"
          :class="monitor.wsConnected
            ? 'text-[color:var(--epm-citrico)]'
            : 'text-[color:var(--signal-fault)]'"
        >
          {{ monitor.wsConnected ? 'Live' : 'Offline' }}
        </span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 min-h-0 overflow-y-auto py-3 px-2 space-y-0.5">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          :title="collapsed ? item.label : undefined"
          class="group relative flex items-center gap-3 rounded-sm px-3 py-2.5 text-sm font-semibold
                 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          active-class="bg-sidebar-accent text-sidebar-accent-foreground
                        before:absolute before:left-0 before:top-2 before:bottom-2
                        before:w-0.5 before:rounded-r before:bg-[color:var(--sidebar-primary)]"
          @click="mobileOpen = false"
        >
          <component :is="item.icon" class="w-4 h-4 shrink-0" />
          <span class="sidebar-label truncate">{{ item.label }}</span>
        </RouterLink>
      </nav>

      <!-- Collapse toggle (desktop only) -->
      <div class="border-t border-sidebar-border p-2 shrink-0">
        <button
          class="hidden md:flex w-full items-center gap-3 rounded-sm px-3 py-2 text-sm
                 hover:bg-sidebar-accent transition-colors text-sidebar-foreground/60"
          :title="collapsed ? 'Expand' : 'Collapse'"
          @click="toggleSidebar"
        >
          <component :is="collapsed ? PanelLeftOpen : PanelLeftClose" class="w-4 h-4 shrink-0" />
          <span class="sidebar-label text-[10px] font-bold uppercase tracking-widest">Collapse</span>
        </button>
        <div class="px-3 pt-2 sidebar-wide-only">
          <div class="text-[9px] uppercase tracking-[0.2em] text-[color:var(--epm-citrico)]/60 font-semibold">
            IEC 104 Master
          </div>
        </div>
      </div>
    </aside>

    <!-- ── MAIN ──────────────────────────────────────────────────────────── -->
    <div class="flex flex-col flex-1 min-w-0 h-screen overflow-hidden">

      <!-- Mobile top bar -->
      <header class="md:hidden flex items-center gap-3 px-4 h-14 border-b border-border bg-card shrink-0">
        <button
          class="grid place-items-center w-9 h-9 rounded-sm border border-border
                 hover:bg-muted transition-colors"
          @click="mobileOpen = true"
        >
          <Menu class="w-4 h-4" />
        </button>

        <div class="grid place-items-center w-7 h-7 rounded-sm shrink-0
                    bg-[color:var(--epm-citrico)] text-[color:var(--epm-bosque)] font-black text-sm">
          G
        </div>
        <span class="font-bold text-sm tracking-tight">GO104</span>

        <div class="ml-auto flex items-center gap-1.5">
          <span
            class="status-dot"
            :class="monitor.wsConnected
              ? 'text-[color:var(--epm-citrico)]'
              : 'text-[color:var(--signal-fault)]'"
          />
          <span class="text-xs font-semibold"
            :class="monitor.wsConnected
              ? 'text-[color:var(--epm-citrico)]'
              : 'text-[color:var(--signal-fault)]'"
          >
            {{ monitor.wsConnected ? 'Live' : 'Offline' }}
          </span>
        </div>
      </header>

      <main class="flex-1 min-h-0 overflow-y-auto">
        <RouterView />
      </main>
    </div>

  </div>
</template>

<style scoped>
nav { scrollbar-width: thin; scrollbar-color: rgba(159,207,103,.2) transparent; }
</style>
