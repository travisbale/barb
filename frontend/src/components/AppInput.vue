<script setup lang="ts">
defineProps<{
  modelValue: string
  type?: string
  placeholder?: string
  required?: boolean
  autofocus?: boolean
  multiline?: boolean
  rows?: number
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const baseClass = 'px-3 py-2 bg-bg border border-edge text-sm font-mono text-primary placeholder-dim focus:outline-none focus:border-amber/40 focus:ring-1 focus:ring-amber/20 transition-colors leading-relaxed'
</script>

<template>
  <textarea
    v-if="multiline"
    :value="modelValue"
    @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    :placeholder="placeholder"
    :required="required"
    :autofocus="autofocus"
    :rows="rows ?? 4"
    :class="[baseClass, 'resize-y']"
  ></textarea>
  <input
    v-else
    :value="modelValue"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    :type="type ?? 'text'"
    :placeholder="placeholder"
    :required="required"
    :autofocus="autofocus"
    :class="baseClass"
  />
</template>
