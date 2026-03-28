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
      'px-4 py-2 text-sm font-medium rounded-lg transition-colors disabled:opacity-50',
      {
        'bg-indigo-500 hover:bg-indigo-600 text-white': variant === 'primary' || !variant,
        'border border-gray-300 hover:bg-gray-50 text-gray-700': variant === 'secondary',
        'text-gray-500 hover:text-gray-700': variant === 'ghost',
        'text-gray-400 hover:text-red-500': variant === 'danger',
      },
    ]"
  >
    <slot />
  </button>
</template>
