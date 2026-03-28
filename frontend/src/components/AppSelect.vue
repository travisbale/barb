<script setup lang="ts">
defineProps<{
  modelValue: string
  required?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="relative">
    <select
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
      :required="required"
      class="w-full px-3 py-2 pr-9 bg-bg border border-edge text-sm font-mono text-primary appearance-none focus:outline-none focus:border-amber/40 focus:ring-1 focus:ring-amber/20 transition-colors"
    >
      <slot />
    </select>
    <svg
      class="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-muted"
      width="12" height="12" viewBox="0 0 24 24"
      fill="none" stroke="currentColor" stroke-width="2"
    >
      <path d="M6 9l6 6 6-6" />
    </svg>
  </div>
</template>
