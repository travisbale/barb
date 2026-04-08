<script setup lang="ts">
import { useCrudList } from '../composables/useCrudList'
import { listTemplates, createTemplate, updateTemplate, deleteTemplate, type EmailTemplate } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import DeleteButton from '../components/DeleteButton.vue'
import TemplateForm from '../components/TemplateForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import FormCard from '../components/FormCard.vue'
import AddButton from '../components/AddButton.vue'

type TemplateForm = { name: string; subject: string; html_body: string; text_body: string; envelope_sender: string }
const emptyForm = (): TemplateForm => ({ name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' })

const { items: templates, showForm, editingId, error, form, openCreate, openEdit, closeForm, submit, remove } = useCrudList<EmailTemplate, TemplateForm>(
  { list: listTemplates, create: createTemplate, update: updateTemplate, remove: deleteTemplate },
  { emptyForm, toForm: (t) => ({ name: t.name, subject: t.subject, html_body: t.html_body, text_body: t.text_body, envelope_sender: t.envelope_sender ?? '' }), confirmMessage: 'Delete this template?' },
)

</script>

<template>
  <div>
    <PageHeader title="Email Templates" :subtitle="`${templates.length} templates`">
      <AddButton @click="openCreate">New Template</AddButton>
    </PageHeader>

    <ErrorBanner v-model="error" />

    <FormCard v-if="showForm" @submit="submit">
      <TemplateForm v-model="form" />
      <template #actions>
        <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="templates.length === 0 && !showForm" message="No templates. Create one to compose phishing emails." />

    <DataTable v-else-if="templates.length > 0" :columns="[{ label: 'Name' }, { label: 'Subject' }, { label: 'Created' }, { label: '', width: 'w-12' }]">
      <DataTableRow
        v-for="(tmpl, i) in templates"
        :key="tmpl.id"
        :index="i"
        clickable
        @click="openEdit(tmpl)"
      >
        <td class="text-primary">{{ tmpl.name }}</td>
        <td class="text-muted">{{ tmpl.subject }}</td>
        <td class="text-dim">{{ new Date(tmpl.created_at).toLocaleDateString() }}</td>
        <td class="text-right">
          <DeleteButton @click.stop="remove(tmpl.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
