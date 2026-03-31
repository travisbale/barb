<script setup lang="ts">
defineProps<{
  tabs: string[]
  modelValue: string
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="flex gap-4 mb-4 border-b border-edge">
    <button
      v-for="tab in tabs"
      :key="tab"
      @click="$emit('update:modelValue', tab)"
      class="px-1 py-2 text-sm font-mono uppercase tracking-wider transition-colors"
      :class="modelValue === tab ? 'text-amber border-b-2 border-amber' : 'text-dim hover:text-primary'"
    >{{ tab }}</button>
  </div>
</template>
