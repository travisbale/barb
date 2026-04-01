<script setup lang="ts">
const props = defineProps<{
  clickable?: boolean
  index?: number
}>()
</script>

<template>
  <tr
    :style="props.index != null ? { animationDelay: `${props.index * 20}ms` } : undefined"
    class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
    :class="{ 'cursor-pointer': clickable }"
  >
    <slot />
  </tr>
</template>
