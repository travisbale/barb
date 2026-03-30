<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  modelValue: string
  label?: string
  required?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
}>()

const selectEl = ref<HTMLSelectElement>()
const focused = ref(false)
const hasSelection = () => {
  const el = selectEl.value
  if (!el) return !!props.modelValue
  const selected = el.options[el.selectedIndex]
  return selected ? !selected.disabled : !!props.modelValue
}
const isActive = () => focused.value || hasSelection()
</script>

<template>
  <div class="relative">
    <label
      v-if="label"
      class="absolute left-3 font-mono transition-all duration-150 pointer-events-none z-10"
      :class="isActive()
        ? 'text-xs -top-2.5 px-1 bg-surface text-amber/70'
        : 'text-sm top-2.5 text-dim'"
    >{{ label }}</label>
    <select
      ref="selectEl"
      :value="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
      @focus="focused = true"
      @blur="focused = false"
      :required="required"
      class="w-full px-3 pt-3 pb-2 pr-9 bg-surface border border-edge text-sm font-mono text-primary appearance-none focus:outline-none focus:border-amber/40 focus:ring-1 focus:ring-amber/20 transition-colors"
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
