<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <div class="flex items-end justify-between mb-8">
    <div>
      <h1 class="text-xl font-mono font-semibold tracking-tight text-primary">{{ title }}</h1>
      <p v-if="subtitle" class="text-sm text-dim font-mono mt-1">{{ subtitle }}</p>
    </div>
    <div class="flex gap-2">
      <slot />
    </div>
  </div>
</template>
