<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, shallowRef } from 'vue'
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { yaml } from '@codemirror/lang-yaml'
import { html } from '@codemirror/lang-html'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { syntaxHighlighting, defaultHighlightStyle, HighlightStyle } from '@codemirror/language'
import { tags } from '@lezer/highlight'

const props = defineProps<{
  modelValue: string
  language?: 'yaml' | 'html'
  placeholder?: string
  minHeight?: string
  label?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const container = ref<HTMLDivElement>()
const view = shallowRef<EditorView>()

// Custom highlight style matching the app's color palette.
const barbHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: 'rgb(var(--color-amber))' },
  { tag: tags.atom, color: 'rgb(var(--color-teal))' },
  { tag: tags.bool, color: 'rgb(var(--color-teal))' },
  { tag: tags.number, color: 'rgb(var(--color-teal))' },
  { tag: tags.string, color: '#8fac7e' },
  { tag: tags.comment, color: 'rgb(var(--color-text-dim))', fontStyle: 'italic' },
  { tag: tags.propertyName, color: 'rgb(var(--color-amber))' },
  { tag: tags.punctuation, color: 'rgb(var(--color-text-dim))' },
  { tag: tags.operator, color: 'rgb(var(--color-text-muted))' },
  { tag: tags.meta, color: 'rgb(var(--color-text-dim))' },
])

const theme = EditorView.theme({
  '&': {
    fontSize: '13px',
    fontFamily: "'JetBrains Mono', monospace",
    backgroundColor: 'rgb(var(--color-bg))',
    color: 'rgb(var(--color-text))',
    border: '1px solid rgb(var(--color-border))',
  },
  '&.cm-focused': {
    outline: 'none',
    borderColor: 'rgb(var(--color-amber) / 0.4)',
  },
  '.cm-content': {
    caretColor: 'rgb(var(--color-amber))',
    padding: '8px 0',
  },
  '.cm-cursor': {
    borderLeftColor: 'rgb(var(--color-amber))',
  },
  '.cm-activeLine': {
    backgroundColor: 'rgba(255, 255, 255, 0.02)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'rgba(255, 255, 255, 0.02)',
  },
  '.cm-gutters': {
    backgroundColor: 'rgb(var(--color-bg))',
    color: 'rgb(var(--color-text-dim) / 0.5)',
    border: 'none',
    borderRight: '1px solid rgb(var(--color-border))',
  },
  '.cm-lineNumbers .cm-gutterElement': {
    padding: '0 8px 0 12px',
    minWidth: '3ch',
  },
  '.cm-selectionBackground': {
    backgroundColor: 'rgba(176, 136, 37, 0.2) !important',
  },
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: 'rgba(176, 136, 37, 0.3) !important',
  },
  '.cm-placeholder': {
    color: 'rgb(var(--color-text-dim))',
    fontStyle: 'italic',
  },
}, { dark: true })

function createState(doc: string): EditorState {
  return EditorState.create({
    doc,
    extensions: [
      lineNumbers(),
      highlightActiveLine(),
      highlightActiveLineGutter(),
      history(),
      props.language === 'html' ? html() : yaml(),
      syntaxHighlighting(barbHighlight),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      keymap.of([...defaultKeymap, ...historyKeymap]),
      theme,
      EditorView.updateListener.of((update) => {
        if (update.docChanged) {
          emit('update:modelValue', update.state.doc.toString())
        }
      }),
      props.placeholder ? EditorView.contentAttributes.of({ 'aria-placeholder': props.placeholder }) : [],
    ],
  })
}

onMounted(() => {
  if (!container.value) return
  view.value = new EditorView({
    state: createState(props.modelValue),
    parent: container.value,
  })
})

onUnmounted(() => {
  view.value?.destroy()
})

// Sync external changes (e.g. file upload) into the editor.
watch(() => props.modelValue, (newVal) => {
  if (!view.value) return
  const currentDoc = view.value.state.doc.toString()
  if (newVal !== currentDoc) {
    view.value.dispatch({
      changes: { from: 0, to: currentDoc.length, insert: newVal },
    })
  }
})
</script>

<template>
  <div>
    <div v-if="props.label" class="text-xs font-mono text-dim mb-2">{{ props.label }}</div>
    <div ref="container" class="code-editor" :style="{ minHeight: props.minHeight ?? '300px' }"></div>
  </div>
</template>

<style scoped>
.code-editor :deep(.cm-editor) {
  min-height: inherit;
  max-height: 600px;
  overflow: auto;
}
</style>

<style>
.borderless .cm-editor {
  border: none;
}
</style>
