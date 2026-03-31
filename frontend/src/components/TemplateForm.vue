<script setup lang="ts">
import AppInput from './AppInput.vue'
import CodeEditor from './CodeEditor.vue'

defineProps<{
  modelValue: {
    name: string
    subject: string
    html_body: string
    text_body: string
    envelope_sender: string
  }
  minEditorHeight?: string
}>()

defineEmits<{
  'update:modelValue': [value: any]
}>()
</script>

<template>
  <div class="flex flex-col gap-5">
    <div class="grid grid-cols-2 gap-5">
      <AppInput :modelValue="modelValue.name" @update:modelValue="$emit('update:modelValue', { ...modelValue, name: $event })" placeholder="Template name" required />
      <AppInput :modelValue="modelValue.subject" @update:modelValue="$emit('update:modelValue', { ...modelValue, subject: $event })" placeholder="Email subject" required />
    </div>
    <div class="text-xs font-mono text-dim px-1" v-pre>
      Variables: <code class="text-muted">{{.FirstName}}</code> <code class="text-muted">{{.LastName}}</code> <code class="text-muted">{{.Email}}</code> <code class="text-muted">{{.URL}}</code>
    </div>
    <CodeEditor :modelValue="modelValue.html_body" @update:modelValue="$emit('update:modelValue', { ...modelValue, html_body: $event })" language="html" :min-height="minEditorHeight ?? '200px'" />
    <AppInput :modelValue="modelValue.text_body" @update:modelValue="$emit('update:modelValue', { ...modelValue, text_body: $event })" multiline :rows="4" placeholder="Plain text body " />
    <AppInput :modelValue="modelValue.envelope_sender" @update:modelValue="$emit('update:modelValue', { ...modelValue, envelope_sender: $event })" placeholder="Envelope sender / Return-Path " />
  </div>
</template>
