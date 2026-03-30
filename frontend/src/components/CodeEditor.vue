<script setup lang="ts">
import { ref, onMounted, watch, shallowRef } from 'vue'
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter } from '@codemirror/view'
import { EditorState } from '@codemirror/state'
import { yaml } from '@codemirror/lang-yaml'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { syntaxHighlighting, defaultHighlightStyle, HighlightStyle } from '@codemirror/language'
import { tags } from '@lezer/highlight'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const container = ref<HTMLDivElement>()
const view = shallowRef<EditorView>()

// Custom highlight style matching the app's color palette.
const miradorHighlight = HighlightStyle.define([
  { tag: tags.keyword, color: 'var(--color-amber)' },
  { tag: tags.atom, color: 'var(--color-teal)' },
  { tag: tags.bool, color: 'var(--color-teal)' },
  { tag: tags.number, color: 'var(--color-teal)' },
  { tag: tags.string, color: '#8fac7e' },
  { tag: tags.comment, color: 'var(--color-text-dim)', fontStyle: 'italic' },
  { tag: tags.propertyName, color: 'var(--color-amber)' },
  { tag: tags.punctuation, color: 'var(--color-text-dim)' },
  { tag: tags.operator, color: 'var(--color-text-muted)' },
  { tag: tags.meta, color: 'var(--color-text-dim)' },
])

const theme = EditorView.theme({
  '&': {
    fontSize: '13px',
    fontFamily: "'JetBrains Mono', monospace",
    backgroundColor: 'var(--color-bg)',
    color: 'var(--color-text)',
    border: '1px solid var(--color-border)',
  },
  '&.cm-focused': {
    outline: 'none',
    borderColor: 'rgba(var(--color-amber), 0.4)',
  },
  '.cm-content': {
    caretColor: 'var(--color-amber)',
    padding: '8px 0',
  },
  '.cm-cursor': {
    borderLeftColor: 'var(--color-amber)',
  },
  '.cm-activeLine': {
    backgroundColor: 'rgba(255, 255, 255, 0.02)',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'rgba(255, 255, 255, 0.02)',
  },
  '.cm-gutters': {
    backgroundColor: 'var(--color-surface)',
    color: 'var(--color-text-dim)',
    border: 'none',
    borderRight: '1px solid var(--color-border)',
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
    color: 'var(--color-text-dim)',
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
      yaml(),
      syntaxHighlighting(miradorHighlight),
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
  <div ref="container" class="code-editor"></div>
</template>

<style scoped>
.code-editor {
  min-height: 300px;
}
.code-editor :deep(.cm-editor) {
  min-height: 300px;
  max-height: 600px;
  overflow: auto;
}
</style>
