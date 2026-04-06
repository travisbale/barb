<script setup lang="ts">
const message = defineModel<string>({ required: true })
</script>

<template>
  <div v-if="message" class="border border-danger/30 bg-danger-dim px-4 py-2.5 mb-4 flex items-center justify-between">
    <span class="text-danger text-sm font-mono">{{ message }}</span>
    <button @click="message = ''" class="text-danger/50 hover:text-danger transition-colors text-2xl leading-none">&times;</button>
  </div>
</template>
