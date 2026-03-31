<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
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

const focused = ref(false)
const isActive = () => focused.value || props.modelValue
</script>

<template>
  <div class="relative">
    <label
      v-if="placeholder"
      class="absolute left-3 font-mono transition-all duration-150 pointer-events-none"
      :class="isActive()
        ? ['text-xs -top-2.5 px-1 bg-surface', focused ? 'text-amber/70' : 'text-dim']
        : 'text-sm top-2.5 text-dim'"
    >{{ placeholder }}<span v-if="required" class="text-amber/70 ml-0.5">*</span></label>
    <textarea
      v-if="multiline"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
      :required="required"
      :autofocus="autofocus"
      :rows="rows ?? 4"
      @focus="focused = true"
      @blur="focused = false"
      class="w-full px-3 pt-3 pb-2 bg-surface border border-edge text-sm font-mono text-primary focus:outline-none focus:border-amber/40 focus:ring-1 focus:ring-amber/20 transition-colors leading-relaxed resize-y"
    ></textarea>
    <input
      v-else
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      :type="type ?? 'text'"
      :required="required"
      :autofocus="autofocus"
      @focus="focused = true"
      @blur="focused = false"
      class="w-full px-3 pt-3 pb-2 bg-surface border border-edge text-sm font-mono text-primary focus:outline-none focus:border-amber/40 focus:ring-1 focus:ring-amber/20 transition-colors leading-relaxed"
    />
  </div>
</template>
