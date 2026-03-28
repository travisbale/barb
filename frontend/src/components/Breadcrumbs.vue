<script setup lang="ts">
import { useRouter } from 'vue-router'

defineProps<{
  crumbs: { label: string; to?: string }[]
}>()

const router = useRouter()
</script>

<template>
  <nav class="flex items-center gap-1.5 text-xs font-mono mb-2">
    <template v-for="(crumb, i) in crumbs" :key="i">
      <span v-if="i > 0" class="text-dim select-none">/</span>
      <button
        v-if="crumb.to"
        @click="router.push(crumb.to)"
        class="text-dim hover:text-amber transition-colors"
      >{{ crumb.label }}</button>
      <span v-else class="text-dim">{{ crumb.label }}</span>
    </template>
  </nav>
</template>
