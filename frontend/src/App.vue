<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const nav = [
  { to: '/campaigns', label: 'Campaigns', icon: '&#x25B6;' },
  { to: '/targets', label: 'Targets', icon: '&#x25C9;' },
  { to: '/templates', label: 'Templates', icon: '&#x25A1;' },
  { to: '/smtp', label: 'SMTP', icon: '&#x25B7;' },
  { to: '/settings', label: 'Settings', icon: '&#x2699;' },
]
</script>

<template>
  <div class="flex min-h-screen">
    <!-- Sidebar -->
    <nav class="w-52 bg-surface border-r border-edge flex flex-col">
      <div class="px-5 py-5 border-b border-edge">
        <div class="font-mono text-sm font-bold tracking-widest text-amber uppercase">Mirador</div>
        <div class="font-mono text-xs text-dim mt-0.5 tracking-wider">Operations Console</div>
      </div>

      <div class="flex-1 py-3 px-2 flex flex-col gap-0.5">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="group flex items-center gap-3 px-3 py-2 text-sm font-mono text-muted hover:text-gray-200 hover:bg-surface-hover transition-colors"
          active-class="!text-amber bg-amber-glow border-l-2 border-amber !pl-[10px]"
        >
          <span class="text-xs opacity-50 group-hover:opacity-100 transition-opacity" v-html="item.icon"></span>
          {{ item.label }}
        </RouterLink>
      </div>

      <div class="px-5 py-4 border-t border-edge">
        <div class="font-mono text-xs text-dim">
          <span class="inline-block w-1.5 h-1.5 bg-teal rounded-full mr-1.5 align-middle"></span>
          Connected
        </div>
      </div>
    </nav>

    <!-- Main content -->
    <main class="flex-1 p-8 overflow-auto">
      <div class="max-w-5xl mx-auto">
        <RouterView />
      </div>
    </main>
  </div>
</template>
