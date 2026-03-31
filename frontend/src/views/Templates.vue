<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listTemplates, createTemplate, updateTemplate, previewTemplate, deleteTemplate, type EmailTemplate, type PreviewResult } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import TemplateForm from '../components/TemplateForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import FormCard from '../components/FormCard.vue'
import AddButton from '../components/AddButton.vue'

const { confirm } = useConfirm()
const templates = ref<EmailTemplate[]>([])
const showForm = ref(false)
const editingId = ref<string | null>(null)
const error = ref('')

const emptyForm = { name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' }
const form = ref({ ...emptyForm })

// Preview state.
const previewId = ref<string | null>(null)
const previewData = ref({ first_name: 'Jane', last_name: 'Doe', email: 'jane.doe@example.com', url: 'https://phish.example.com/lure123' })
const previewResult = ref<PreviewResult | null>(null)
const previewing = ref(false)

async function load() {
  try {
    templates.value = await listTemplates() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

function openCreate() {
  editingId.value = null
  form.value = { ...emptyForm }
  showForm.value = true
  closePreview()
}

function openEdit(tmpl: EmailTemplate) {
  editingId.value = tmpl.id
  form.value = { name: tmpl.name, subject: tmpl.subject, html_body: tmpl.html_body, text_body: tmpl.text_body, envelope_sender: tmpl.envelope_sender ?? '' }
  showForm.value = true
  closePreview()
}

function closeForm() {
  showForm.value = false
  editingId.value = null
  form.value = { ...emptyForm }
}

async function submit() {
  try {
    if (editingId.value) {
      const updated = await updateTemplate(editingId.value, form.value)
      const idx = templates.value.findIndex(t => t.id === editingId.value)
      if (idx !== -1) templates.value[idx] = updated
    } else {
      const created = await createTemplate(form.value)
      templates.value.unshift(created)
    }
    closeForm()
  } catch (e: any) {
    error.value = e.message
  }
}

function openPreview(tmpl: EmailTemplate) {
  previewId.value = tmpl.id
  previewResult.value = null
}

function closePreview() {
  previewId.value = null
  previewResult.value = null
}

async function runPreview() {
  if (!previewId.value) return
  previewing.value = true
  error.value = ''
  try {
    previewResult.value = await previewTemplate(previewId.value, previewData.value)
  } catch (e: any) {
    error.value = e.message
  } finally {
    previewing.value = false
  }
}

async function remove(id: string) {
  if (!await confirm('Delete this template?')) return
  try {
    await deleteTemplate(id)
    templates.value = templates.value.filter(t => t.id !== id)
    if (previewId.value === id) closePreview()
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Email Templates" :subtitle="`${templates.length} templates`">
      <AddButton @click="openCreate">New Template</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <FormCard v-if="showForm" @submit="submit">
      <TemplateForm v-model="form" />
      <template #actions>
        <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
      </template>
    </FormCard>

    <!-- Preview panel -->
    <Card v-if="previewId" class="p-7 mb-4">
      <div class="flex flex-col gap-5">
        <div class="text-xs font-mono uppercase tracking-wider text-dim">Preview</div>
        <div class="grid grid-cols-4 gap-4">
          <AppInput v-model="previewData.first_name" placeholder="First name" />
          <AppInput v-model="previewData.last_name" placeholder="Last name" />
          <AppInput v-model="previewData.email" placeholder="Email" />
          <AppInput v-model="previewData.url" placeholder="Lure URL" />
        </div>
        <div class="flex gap-2 justify-end">
          <AppButton variant="ghost" @click="closePreview">Close</AppButton>
          <AppButton :disabled="previewing" @click="runPreview">{{ previewing ? 'Rendering...' : 'Render' }}</AppButton>
        </div>
        <div v-if="previewResult" class="flex flex-col gap-4 mt-2">
          <div>
            <div class="text-xs font-mono text-dim uppercase tracking-wider mb-1">Subject</div>
            <div class="text-sm text-primary font-mono px-3 py-2 bg-bg border border-edge">{{ previewResult.subject }}</div>
          </div>
          <div v-if="previewResult.html_body">
            <div class="text-xs font-mono text-dim uppercase tracking-wider mb-1">HTML</div>
            <iframe
              :srcdoc="previewResult.html_body"
              class="w-full border border-edge bg-white"
              style="min-height: 200px;"
              sandbox=""
            ></iframe>
          </div>
          <div v-if="previewResult.text_body">
            <div class="text-xs font-mono text-dim uppercase tracking-wider mb-1">Plain Text</div>
            <pre class="text-sm text-muted font-mono px-3 py-2 bg-bg border border-edge whitespace-pre-wrap">{{ previewResult.text_body }}</pre>
          </div>
        </div>
      </div>
    </Card>

    <EmptyState v-if="templates.length === 0 && !showForm" message="No templates. Create one to compose phishing emails." />

    <Card v-else-if="templates.length > 0">
      <div
        v-for="(tmpl, i) in templates"
        :key="tmpl.id"
        :style="{ animationDelay: `${i * 30}ms` }"
        class="animate-in flex items-center justify-between px-4 py-3 border-b border-edge last:border-0 hover:bg-surface-hover cursor-pointer transition-colors"
        @click="openEdit(tmpl)"
      >
        <div>
          <div class="text-sm font-medium text-primary">{{ tmpl.name }}</div>
          <div class="text-xs text-dim font-mono mt-0.5">Subject: {{ tmpl.subject }}</div>
        </div>
        <div class="flex items-center gap-2">
          <button @click.stop="openPreview(tmpl)" class="text-xs font-mono text-dim hover:text-amber transition-colors uppercase tracking-wider">Preview</button>
          <button @click.stop="remove(tmpl.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
        </div>
      </div>
    </Card>
  </div>
</template>
