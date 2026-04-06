<script setup lang="ts">
import { ref } from 'vue'
import { useCrudList } from '../composables/useCrudList'
import { listTemplates, createTemplate, updateTemplate, previewTemplate, deleteTemplate, type EmailTemplate, type PreviewResult } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import DeleteButton from '../components/DeleteButton.vue'
import AppInput from '../components/AppInput.vue'
import TemplateForm from '../components/TemplateForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import FormCard from '../components/FormCard.vue'
import AddButton from '../components/AddButton.vue'

type TemplateForm = { name: string; subject: string; html_body: string; text_body: string; envelope_sender: string }
const emptyForm = (): TemplateForm => ({ name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' })

const { items: templates, showForm, editingId, error, form, openCreate: rawOpenCreate, openEdit: rawOpenEdit, closeForm, submit, remove } = useCrudList<EmailTemplate, TemplateForm>(
  { list: listTemplates, create: createTemplate, update: updateTemplate, remove: deleteTemplate },
  { emptyForm, toForm: (t) => ({ name: t.name, subject: t.subject, html_body: t.html_body, text_body: t.text_body, envelope_sender: t.envelope_sender ?? '' }), confirmMessage: 'Delete this template?' },
)

// Preview state.
const previewId = ref<string | null>(null)
const previewData = ref({ first_name: 'Jane', last_name: 'Doe', email: 'jane.doe@example.com', url: 'https://phish.example.com/lure123' })
const previewResult = ref<PreviewResult | null>(null)
const previewing = ref(false)

function openCreate() {
  rawOpenCreate()
  closePreview()
}

function openEdit(tmpl: EmailTemplate) {
  rawOpenEdit(tmpl)
  closePreview()
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

function removeTemplate(id: string) {
  if (previewId.value === id) closePreview()
  remove(id)
}
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

    <DataTable v-else-if="templates.length > 0" :columns="[{ label: 'Name' }, { label: 'Subject' }, { label: 'Created' }, { label: '', width: 'w-28' }]">
      <DataTableRow
        v-for="(tmpl, i) in templates"
        :key="tmpl.id"
        :index="i"
        clickable
        @click="openEdit(tmpl)"
      >
        <td class="px-4 py-2.5 text-primary">{{ tmpl.name }}</td>
        <td class="px-4 py-2.5 text-muted">{{ tmpl.subject }}</td>
        <td class="px-4 py-2.5 text-dim">{{ new Date(tmpl.created_at).toLocaleDateString() }}</td>
        <td class="px-4 py-2.5 text-right">
          <div class="flex items-center gap-4 justify-end">
            <button @click.stop="openPreview(tmpl)" class="text-xs font-mono text-dim hover:text-amber transition-colors uppercase tracking-wider">Preview</button>
            <DeleteButton @click.stop="removeTemplate(tmpl.id)" />
          </div>
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
