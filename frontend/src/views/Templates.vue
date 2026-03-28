<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listTemplates, createTemplate, deleteTemplate, type EmailTemplate } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const templates = ref<EmailTemplate[]>([])
const showCreate = ref(false)
const error = ref('')

const form = ref({ name: '', subject: '', html_body: '', text_body: '' })

async function load() {
  try {
    templates.value = await listTemplates() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function create() {
  try {
    const t = await createTemplate(form.value)
    templates.value.unshift(t)
    form.value = { name: '', subject: '', html_body: '', text_body: '' }
    showCreate.value = false
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
      <AddButton @click="showCreate = true">New Template</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showCreate" class="p-7 mb-4">
      <form @submit.prevent="create" class="flex flex-col gap-7">
        <div class="grid grid-cols-2 gap-5">
          <AppInput v-model="form.name" placeholder="Template name (required)" required />
          <AppInput v-model="form.subject" placeholder="Email subject (required)" required />
        </div>
        <AppInput v-model="form.html_body" multiline :rows="8" placeholder="HTML body" />
        <AppInput v-model="form.text_body" multiline :rows="4" placeholder="Plain text body (optional)" />
        <div class="flex gap-2">
          <AppButton type="submit">Create</AppButton>
          <AppButton variant="ghost" @click="showCreate = false">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="templates.length === 0 && !showCreate" message="No templates. Create one to compose phishing emails." />

    <Card v-else-if="templates.length > 0">
      <div
        v-for="(tmpl, i) in templates"
        :key="tmpl.id"
        :style="{ animationDelay: `${i * 30}ms` }"
        class="animate-in flex items-center justify-between px-4 py-3 border-b border-edge last:border-0 hover:bg-surface-hover transition-colors"
      >
        <div>
          <div class="text-sm font-medium text-primary">{{ tmpl.name }}</div>
          <div class="text-xs text-dim font-mono mt-0.5">Subject: {{ tmpl.subject }}</div>
        </div>
        <button @click="remove(tmpl.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
      </div>
    </Card>
  </div>
</template>
