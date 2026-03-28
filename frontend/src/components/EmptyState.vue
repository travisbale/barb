<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <div class="bg-white rounded-lg border border-gray-200 p-8 text-center text-gray-500">
    {{ message }}
  </div>
</template>
