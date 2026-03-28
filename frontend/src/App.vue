<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import IconCampaign from './components/IconCampaign.vue'
import IconTarget from './components/IconTarget.vue'
import IconTemplate from './components/IconTemplate.vue'
import IconSMTP from './components/IconSMTP.vue'
import IconSettings from './components/IconSettings.vue'

const nav = [
  { to: '/campaigns', label: 'Campaigns', icon: IconCampaign },
  { to: '/targets', label: 'Targets', icon: IconTarget },
  { to: '/templates', label: 'Templates', icon: IconTemplate },
  { to: '/smtp', label: 'SMTP', icon: IconSMTP },
  { to: '/settings', label: 'Miraged', icon: IconSettings },
]
</script>

<template>
  <div class="flex min-h-screen">
    <!-- Sidebar -->
    <nav class="w-52 bg-surface border-r border-edge flex flex-col">
      <div class="px-5 py-5 border-b border-edge">
        <div class="font-mono text-base font-bold tracking-widest text-amber uppercase">Mirador</div>
        <div class="font-mono text-xs text-dim mt-0.5 tracking-wider">Operations Console</div>
      </div>

      <div class="flex-1 py-3 px-2 flex flex-col gap-0.5">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="group flex items-center gap-3 px-3 py-2.5 text-sm font-mono text-muted hover:text-primary hover:bg-surface-hover transition-colors uppercase tracking-wide"
          active-class="!text-amber bg-amber-glow border-l-2 border-amber !pl-[10px]"
        >
          <component :is="item.icon" class="opacity-50 group-hover:opacity-100 transition-opacity" />
          {{ item.label }}
        </RouterLink>
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
