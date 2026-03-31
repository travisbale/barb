<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { useTheme } from './composables/useTheme'
import { logout } from './api/client'
import IconCampaign from './components/IconCampaign.vue'
import IconTarget from './components/IconTarget.vue'
import IconTemplate from './components/IconTemplate.vue'
import IconPhishlet from './components/IconPhishlet.vue'
import IconSMTP from './components/IconSMTP.vue'
import IconSettings from './components/IconSettings.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const { isDark, toggle } = useTheme()

const showShell = computed(() => !route.meta.public)

const nav = [
  { to: '/campaigns', label: 'Campaigns', icon: IconCampaign },
  { to: '/targets', label: 'Targets', icon: IconTarget },
  { to: '/templates', label: 'Templates', icon: IconTemplate },
  { to: '/phishlets', label: 'Phishlets', icon: IconPhishlet },
  { to: '/smtp', label: 'SMTP', icon: IconSMTP },
  { to: '/settings', label: 'Miraged', icon: IconSettings },
]

async function handleLogout() {
  try {
    await logout()
  } catch {
    // Ignore errors — clear session locally regardless.
  }
  router.push('/login')
}
</script>

<template>
  <div class="flex min-h-screen">
    <!-- Sidebar (hidden on login/change-password) -->
    <nav v-if="showShell" class="w-52 bg-surface border-r border-edge flex flex-col">
      <RouterLink to="/dashboard" class="block px-5 py-5 border-b border-edge hover:bg-surface-hover transition-colors">
        <div class="font-mono text-base font-bold tracking-widest text-amber uppercase">Barb</div>
        <div class="font-mono text-xs text-dim mt-0.5 tracking-wider">Operations Console</div>
      </RouterLink>

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

      <div class="px-3 py-3 border-t border-edge flex flex-col gap-1">
        <button
          @click="toggle"
          class="flex items-center gap-2 px-3 py-2 text-xs font-mono text-dim hover:text-primary transition-colors uppercase tracking-wide w-full"
        >
          <svg v-if="isDark" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
          {{ isDark ? 'Light mode' : 'Dark mode' }}
        </button>
        <button
          @click="handleLogout"
          class="flex items-center gap-2 px-3 py-2 text-xs font-mono text-dim hover:text-danger transition-colors uppercase tracking-wide w-full"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
          Sign out
        </button>
      </div>
    </nav>

    <ConfirmDialog />

    <!-- Main content -->
    <main class="flex-1 p-8 overflow-auto" :class="{ 'p-0': !showShell }">
      <div :class="showShell ? 'max-w-5xl mx-auto' : ''">
        <RouterView />
      </div>
    </main>
  </div>
</template>
