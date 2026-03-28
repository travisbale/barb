<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listTemplates, createTemplate, updateTemplate, deleteTemplate, type EmailTemplate } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const templates = ref<EmailTemplate[]>([])
const showForm = ref(false)
const editingId = ref<string | null>(null)
const error = ref('')

const emptyForm = { name: '', subject: '', html_body: '', text_body: '' }
const form = ref({ ...emptyForm })

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
}

function openEdit(tmpl: EmailTemplate) {
  editingId.value = tmpl.id
  form.value = { name: tmpl.name, subject: tmpl.subject, html_body: tmpl.html_body, text_body: tmpl.text_body }
  showForm.value = true
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

async function remove(id: string) {
  try {
    await deleteTemplate(id)
    templates.value = templates.value.filter(t => t.id !== id)
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

    <Card v-if="showForm" class="p-7 mb-4">
      <form @submit.prevent="submit" class="flex flex-col gap-7">
        <div class="grid grid-cols-2 gap-5">
          <AppInput v-model="form.name" placeholder="Template name (required)" required />
          <AppInput v-model="form.subject" placeholder="Email subject (required)" required />
        </div>
        <AppInput v-model="form.html_body" multiline :rows="8" placeholder="HTML body" />
        <AppInput v-model="form.text_body" multiline :rows="4" placeholder="Plain text body (optional)" />
        <div class="flex gap-2">
          <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
          <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        </div>
      </form>
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
        <button @click.stop="remove(tmpl.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
      </div>
    </Card>
  </div>
</template>
