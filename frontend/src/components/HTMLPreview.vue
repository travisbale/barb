<script setup lang="ts">
defineProps<{
  srcdoc: string
}>()
</script>

<template>
  <iframe :srcdoc="srcdoc" class="w-full bg-white" style="min-height: 300px;" sandbox="" />
</template>
