<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listPhishlets, createPhishlet, updatePhishlet, deletePhishlet, type Phishlet } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const phishlets = ref<Phishlet[]>([])
const showForm = ref(false)
const editingId = ref<string | null>(null)
const error = ref('')

const emptyForm = { name: '', yaml: '' }
const form = ref({ ...emptyForm })

async function load() {
  try {
    phishlets.value = await listPhishlets() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

function openCreate() {
  editingId.value = null
  form.value = { ...emptyForm }
  showForm.value = true
}

function openEdit(phishlet: Phishlet) {
  editingId.value = phishlet.id
  form.value = { name: phishlet.name, yaml: phishlet.yaml }
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
      const updated = await updatePhishlet(editingId.value, form.value)
      const idx = phishlets.value.findIndex(p => p.id === editingId.value)
      if (idx !== -1) phishlets.value[idx] = updated
    } else {
      const created = await createPhishlet(form.value)
      phishlets.value.unshift(created)
    }
    closeForm()
  } catch (e: any) {
    error.value = e.message
  }
}

async function handleFileUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const text = await file.text()
  form.value.yaml = text

  if (!form.value.name) {
    form.value.name = file.name.replace(/\.(yaml|yml)$/i, '')
  }

  input.value = ''
}

async function remove(id: string) {
  try {
    await deletePhishlet(id)
    phishlets.value = phishlets.value.filter(p => p.id !== id)
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Phishlets" :subtitle="`${phishlets.length} phishlets`">
      <AddButton @click="openCreate">New Phishlet</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showForm" class="p-7 mb-4">
      <form @submit.prevent="submit" class="flex flex-col gap-7">
        <div class="flex items-end gap-5">
          <div class="flex-1">
            <AppInput v-model="form.name" placeholder="Phishlet name (required)" required />
          </div>
          <label class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase border border-edge text-muted hover:text-amber hover:border-amber/30 cursor-pointer transition-all duration-150">
            Upload YAML
            <input type="file" accept=".yaml,.yml" class="hidden" @change="handleFileUpload" />
          </label>
        </div>
        <AppInput v-model="form.yaml" multiline :rows="16" placeholder="Phishlet YAML config (required)" required />
        <div class="flex gap-2">
          <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
          <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="phishlets.length === 0 && !showForm" message="No phishlets. Add one to define phishing site configurations." />

    <Card v-else-if="phishlets.length > 0">
      <div
        v-for="(phishlet, i) in phishlets"
        :key="phishlet.id"
        :style="{ animationDelay: `${i * 30}ms` }"
        class="animate-in flex items-center justify-between px-4 py-3 border-b border-edge last:border-0 hover:bg-surface-hover cursor-pointer transition-colors"
        @click="openEdit(phishlet)"
      >
        <div>
          <div class="text-sm font-medium text-primary">{{ phishlet.name }}</div>
          <div class="text-xs text-dim font-mono mt-0.5">{{ new Date(phishlet.created_at).toLocaleDateString() }}</div>
        </div>
        <button @click.stop="remove(phishlet.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
      </div>
    </Card>
  </div>
</template>
