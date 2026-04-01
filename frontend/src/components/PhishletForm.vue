<script setup lang="ts">
import AppButton from './AppButton.vue'
import CodeEditor from './CodeEditor.vue'

defineProps<{
  modelValue: string
  loading?: boolean
  submitLabel?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'submit': []
  'cancel': []
}>()

function handleFileUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  file.text().then(text => emit('update:modelValue', text))
  input.value = ''
}
</script>

<template>
  <div class="flex flex-col gap-7">
    <CodeEditor :modelValue="modelValue" @update:modelValue="$emit('update:modelValue', $event)" label="Phishlet YAML" />
    <div class="flex justify-between">
      <label class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase border border-edge text-muted hover:text-amber hover:border-amber/30 cursor-pointer transition-all duration-150">
        Upload
        <input type="file" accept=".yaml,.yml" class="hidden" @change="handleFileUpload" />
      </label>
      <div class="flex gap-2">
        <AppButton variant="ghost" @click="$emit('cancel')">Cancel</AppButton>
        <AppButton :disabled="loading" @click="$emit('submit')">{{ submitLabel ?? 'Create' }}</AppButton>
      </div>
    </div>
  </div>
</template>
