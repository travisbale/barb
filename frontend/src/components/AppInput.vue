<script setup lang="ts">
defineProps<{
  modelValue: string
  type?: string
  placeholder?: string
  required?: boolean
  autofocus?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <input
    :value="modelValue"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    :type="type ?? 'text'"
    :placeholder="placeholder"
    :required="required"
    :autofocus="autofocus"
    class="px-3 py-2 bg-bg border border-edge text-sm font-mono text-gray-200 placeholder-dim focus:outline-none focus:border-amber/40 focus:ring-1 focus:ring-amber/20 transition-colors leading-relaxed"
  />
</template>
