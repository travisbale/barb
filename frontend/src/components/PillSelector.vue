<script setup lang="ts">
const props = defineProps<{
  modelValue: string[]
  options: string[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

function toggle(value: string) {
  const i = props.modelValue.indexOf(value)
  if (i === -1) {
    emit('update:modelValue', [...props.modelValue, value])
  } else {
    emit('update:modelValue', props.modelValue.filter((v) => v !== value))
  }
}
</script>

<template>
  <div class="flex flex-wrap gap-2">
    <button
      v-for="option in options"
      :key="option"
      type="button"
      :aria-pressed="modelValue.includes(option)"
      @click="toggle(option)"
      class="px-3 py-1 text-xs font-mono rounded-full border transition-colors"
      :class="modelValue.includes(option)
        ? 'bg-amber/20 border-amber/40 text-amber'
        : 'bg-surface border-edge text-muted hover:border-muted'"
    >{{ option }}</button>
  </div>
</template>
