<script setup lang="ts">
import Card from './Card.vue'

defineProps<{
  label: string
  value: string
  subtitle: string
}>()
</script>

<template>
  <Card class="p-5">
    <h6>{{ label }}</h6>
    <div class="text-2xl font-mono font-bold text-primary mt-1">{{ value }}</div>
    <div class="text-xs font-mono text-dim mt-2">{{ subtitle }}</div>
  </Card>
</template>
