<script setup lang="ts">
defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  disabled?: boolean
}>()
</script>

<template>
  <button
    :disabled="disabled"
    :class="[
      'px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase transition-all duration-150 disabled:opacity-30 disabled:cursor-not-allowed',
      {
        'bg-amber text-bg hover:bg-amber-dim': variant === 'primary' || !variant,
        'border border-edge text-muted hover:text-amber hover:border-amber/30 bg-transparent': variant === 'secondary',
        'text-muted hover:text-amber bg-transparent': variant === 'ghost',
        'text-dim hover:text-danger bg-transparent': variant === 'danger',
      },
    ]"
  >
    <slot />
  </button>
</template>
