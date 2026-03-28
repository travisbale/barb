<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <div v-if="message" class="border border-danger/30 bg-danger-dim px-4 py-2.5 mb-4 flex items-center gap-2">
    <span class="text-danger text-sm">&#x25cf;</span>
    <span class="text-danger text-sm font-mono">{{ message }}</span>
  </div>
</template>
