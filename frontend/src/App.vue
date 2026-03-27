<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'

const nav = [
  { to: '/campaigns', label: 'Campaigns' },
  { to: '/targets', label: 'Targets' },
  { to: '/templates', label: 'Templates' },
  { to: '/smtp', label: 'SMTP' },
  { to: '/settings', label: 'Settings' },
]
</script>

<template>
  <div class="flex min-h-screen">
    <nav class="w-56 bg-gray-900 text-gray-300 p-4 flex flex-col gap-1">
      <div class="text-white font-bold text-lg mb-6 px-3">Mirador</div>
      <RouterLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        class="px-3 py-2 rounded-lg text-sm hover:bg-gray-800 hover:text-white transition-colors"
        active-class="bg-gray-800 text-white"
      >
        {{ item.label }}
      </RouterLink>
    </nav>
    <main class="flex-1 p-8">
      <RouterView />
    </main>
  </div>
</template>
