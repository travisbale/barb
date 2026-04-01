<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listPhishlets, createPhishlet, updatePhishlet, deletePhishlet, type Phishlet } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import DeleteButton from '../components/DeleteButton.vue'
import CodeEditor from '../components/CodeEditor.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import FormCard from '../components/FormCard.vue'
import AddButton from '../components/AddButton.vue'

const { confirm } = useConfirm()
const phishlets = ref<Phishlet[]>([])
const showForm = ref(false)
const editingId = ref<string | null>(null)
const error = ref('')

const yaml = ref('')

async function load() {
  try {
    phishlets.value = await listPhishlets() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

function openCreate() {
  editingId.value = null
  yaml.value = ''
  showForm.value = true
}

function openEdit(phishlet: Phishlet) {
  editingId.value = phishlet.id
  yaml.value = phishlet.yaml
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingId.value = null
  yaml.value = ''
}

async function submit() {
  try {
    if (editingId.value) {
      const updated = await updatePhishlet(editingId.value, yaml.value)
      const idx = phishlets.value.findIndex(p => p.id === editingId.value)
      if (idx !== -1) phishlets.value[idx] = updated
    } else {
      const created = await createPhishlet(yaml.value)
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
  yaml.value = await file.text()
  input.value = ''
}

async function remove(id: string) {
  if (!await confirm('Delete this phishlet?')) return
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

    <FormCard v-if="showForm" @submit="submit">
      <CodeEditor v-model="yaml" label="YAML" />
      <template #actions>
        <label class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase border border-edge text-muted hover:text-amber hover:border-amber/30 cursor-pointer transition-all duration-150">
          Upload
          <input type="file" accept=".yaml,.yml" class="hidden" @change="handleFileUpload" />
        </label>
        <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="phishlets.length === 0 && !showForm" message="No phishlets. Add one to define phishing site configurations." />

    <DataTable v-else-if="phishlets.length > 0" :columns="[{ label: 'Name' }, { label: 'Created' }, { label: '', width: 'w-16' }]">
      <DataTableRow
        v-for="(phishlet, i) in phishlets"
        :key="phishlet.id"
        :index="i"
        clickable
        @click="openEdit(phishlet)"
      >
        <td class="px-4 py-2.5 text-primary">{{ phishlet.name }}</td>
        <td class="px-4 py-2.5 text-dim">{{ new Date(phishlet.created_at).toLocaleDateString() }}</td>
        <td class="px-4 py-2.5 text-right">
          <DeleteButton @click.stop="remove(phishlet.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
