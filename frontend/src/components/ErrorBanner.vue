<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <div v-if="message" class="bg-red-50 border border-red-200 text-red-600 text-sm rounded-lg px-3 py-2.5 mb-4">
    {{ message }}
  </div>
</template>
