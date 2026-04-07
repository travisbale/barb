<script setup lang="ts">
defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'link'
  disabled?: boolean
}>()
</script>

<template>
  <button
    :disabled="disabled"
    :class="[
      'inline-flex items-center font-mono font-medium uppercase transition-all duration-150 disabled:opacity-30 disabled:cursor-not-allowed',
      variant === 'link'
        ? 'gap-1 text-xs tracking-wider text-amber hover:text-amber-dim'
        : 'gap-1.5 px-3 py-1.5 text-sm tracking-wide',
      {
        'bg-amber text-bg hover:bg-amber-dim': variant === 'primary' || !variant,
        'border border-edge text-muted hover:text-amber hover:border-amber/30 bg-transparent': variant === 'secondary',
        'text-muted hover:text-amber bg-transparent': variant === 'ghost',
        'border border-edge text-muted hover:text-danger hover:border-danger/30 bg-transparent': variant === 'danger',
      },
    ]"
  >
    <slot />
  </button>
</template>
