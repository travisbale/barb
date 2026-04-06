<script setup lang="ts">
defineProps<{
  label: string
  selectable?: boolean
}>()
</script>

<template>
  <div>
    <div class="text-xs text-dim font-mono">{{ label }}</div>
    <div class="text-sm text-primary font-mono break-all" :class="{ 'select-all': selectable !== false }">
      <slot />
    </div>
  </div>
</template>
