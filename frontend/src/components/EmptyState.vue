<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <div class="border border-dashed border-edge py-16 text-center">
    <p class="text-base text-dim font-mono">{{ message }}</p>
  </div>
</template>
