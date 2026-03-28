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
    class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 focus:border-transparent"
  />
</template>
