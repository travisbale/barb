<script setup lang="ts">
import { ref } from 'vue'
import { renderTemplateHTML } from '../api/client'
import AppInput from './AppInput.vue'
import CodeEditor from './CodeEditor.vue'
import HTMLPreview from './HTMLPreview.vue'

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

const tab = ref<'edit' | 'preview'>('edit')
const previewHtml = ref<string>()
const previewing = ref(false)
const lastRenderedBody = ref<string>()

async function preview(htmlBody: string) {
  tab.value = 'preview'
  if (!htmlBody || previewing.value) return
  if (htmlBody === lastRenderedBody.value) return
  previewing.value = true
  try {
    const result = await renderTemplateHTML(htmlBody)
    previewHtml.value = result.html_body
    lastRenderedBody.value = htmlBody
  } catch {
    tab.value = 'edit'
  } finally {
    previewing.value = false
  }
}
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
        <button type="button" class="text-xs font-mono uppercase tracking-wider transition-colors" :class="tab === 'preview' ? 'text-primary' : 'text-dim hover:text-muted'" @click="preview(modelValue.html_body)">Preview</button>
      </div>
      <div style="min-height: 300px;">
        <CodeEditor v-show="tab === 'edit'" class="borderless" :modelValue="modelValue.html_body" @update:modelValue="$emit('update:modelValue', { ...modelValue, html_body: $event })" language="html" min-height="300px" />
        <template v-if="tab === 'preview'">
          <div v-if="previewing" class="flex items-center justify-center bg-bg text-sm text-dim font-mono h-full" style="min-height: 300px;">Rendering...</div>
          <HTMLPreview v-else-if="previewHtml" :srcdoc="previewHtml" />
          <div v-else class="flex items-center justify-center bg-bg text-sm text-dim font-mono h-full" style="min-height: 300px;">No HTML body to preview.</div>
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
