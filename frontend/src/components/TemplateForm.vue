<script setup lang="ts">
import { ref, watch } from 'vue'
import AppInput from './AppInput.vue'
import CodeEditor from './CodeEditor.vue'

const props = defineProps<{
  modelValue: {
    name: string
    subject: string
    html_body: string
    text_body: string
    envelope_sender: string
  }
  minEditorHeight?: string
  previewHtml?: string
  previewing?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: any]
  'preview': []
}>()

const tab = ref<'edit' | 'preview'>('edit')

// Reset to edit tab when preview result clears (e.g. switching templates)
watch(() => props.previewHtml, (val) => {
  if (!val && tab.value === 'preview') tab.value = 'edit'
})
</script>

<template>
  <div class="flex flex-col gap-7">
    <div class="grid grid-cols-2 gap-7">
      <AppInput :modelValue="modelValue.name" @update:modelValue="$emit('update:modelValue', { ...modelValue, name: $event })" placeholder="Template name" required />
      <AppInput :modelValue="modelValue.subject" @update:modelValue="$emit('update:modelValue', { ...modelValue, subject: $event })" placeholder="Email subject" required />
    </div>
    <div>
      <div class="text-xs font-mono text-dim mb-2 pl-3">HTML Body</div>
      <div class="border border-edge overflow-hidden">
      <div class="flex items-center gap-4 px-3 py-3 border-b border-edge bg-surface">
        <button type="button" class="text-xs font-mono uppercase tracking-wider transition-colors" :class="tab === 'edit' ? 'text-primary' : 'text-dim hover:text-muted'" @click="tab = 'edit'">Edit</button>
        <button type="button" class="text-xs font-mono uppercase tracking-wider transition-colors" :class="tab === 'preview' ? 'text-primary' : 'text-dim hover:text-muted'" @click="tab = 'preview'; $emit('preview')">Preview</button>
      </div>
      <div style="min-height: 300px;">
        <CodeEditor v-show="tab === 'edit'" class="borderless" :modelValue="modelValue.html_body" @update:modelValue="$emit('update:modelValue', { ...modelValue, html_body: $event })" language="html" min-height="300px" />
        <template v-if="tab === 'preview'">
          <div v-if="previewing" class="flex items-center justify-center bg-bg text-sm text-dim font-mono h-full" style="min-height: 300px;">Rendering...</div>
          <iframe v-else-if="previewHtml" :srcdoc="previewHtml" class="w-full bg-white h-full" style="min-height: 300px;" sandbox="" />
          <div v-else class="flex items-center justify-center bg-bg text-sm text-dim font-mono h-full" style="min-height: 300px;">Save the template first to preview</div>
        </template>
      </div>
      </div>
    </div>
    <AppInput :modelValue="modelValue.text_body" @update:modelValue="$emit('update:modelValue', { ...modelValue, text_body: $event })" multiline :rows="8" placeholder="Plain text body " />
    <div class="text-xs font-mono text-dim pl-3" v-pre>
      Available variables: <code class="text-muted">{{.FirstName}}</code> <code class="text-muted">{{.LastName}}</code> <code class="text-muted">{{.Email}}</code> <code class="text-muted">{{.URL}}</code>
    </div>
    <AppInput :modelValue="modelValue.envelope_sender" @update:modelValue="$emit('update:modelValue', { ...modelValue, envelope_sender: $event })" placeholder="Envelope sender / Return-Path " />
  </div>
</template>
