<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listSMTPProfiles, createSMTPProfile, updateSMTPProfile, deleteSMTPProfile, type SMTPProfile } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import DeleteButton from '../components/DeleteButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import FormCard from '../components/FormCard.vue'
import AddButton from '../components/AddButton.vue'

const { confirm } = useConfirm()
const profiles = ref<SMTPProfile[]>([])
const showForm = ref(false)
const editingId = ref<string | null>(null)
const error = ref('')

const emptyForm = { name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' }
const form = ref({ ...emptyForm })
const headers = ref<{ key: string; value: string }[]>([])

async function load() {
  try {
    profiles.value = await listSMTPProfiles() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

function openCreate() {
  editingId.value = null
  form.value = { ...emptyForm }
  headers.value = []
  showForm.value = true
}

function openEdit(profile: SMTPProfile) {
  editingId.value = profile.id
  form.value = {
    name: profile.name,
    host: profile.host,
    port: String(profile.port),
    username: profile.username,
    password: '',
    from_addr: profile.from_addr,
    from_name: profile.from_name,
  }
  headers.value = Object.entries(profile.custom_headers ?? {}).map(([key, value]) => ({ key, value }))
  showForm.value = true
}

function addHeader() {
  headers.value.push({ key: '', value: '' })
}

function removeHeader(index: number) {
  headers.value.splice(index, 1)
}

function headersToMap(): Record<string, string> {
  const map: Record<string, string> = {}
  for (const h of headers.value) {
    if (h.key.trim()) map[h.key.trim()] = h.value
  }
  return map
}

function closeForm() {
  showForm.value = false
  editingId.value = null
  form.value = { ...emptyForm }
}

async function submit() {
  try {
    const payload = {
      name: form.value.name,
      host: form.value.host,
      port: parseInt(form.value.port) || 587,
      username: form.value.username,
      password: form.value.password,
      from_addr: form.value.from_addr,
      from_name: form.value.from_name,
      custom_headers: headersToMap(),
    }

    if (editingId.value) {
      const updated = await updateSMTPProfile(editingId.value, payload)
      const idx = profiles.value.findIndex(p => p.id === editingId.value)
      if (idx !== -1) profiles.value[idx] = updated
    } else {
      const created = await createSMTPProfile(payload)
      profiles.value.unshift(created)
    }
    closeForm()
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(id: string) {
  if (!await confirm('Delete this SMTP profile?')) return
  try {
    await deleteSMTPProfile(id)
    profiles.value = profiles.value.filter(p => p.id !== id)
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="SMTP Profiles" :subtitle="`${profiles.length} profiles`">
      <AddButton @click="openCreate">Add Profile</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <FormCard v-if="showForm" @submit="submit">
      <div class="grid grid-cols-2 gap-7">
        <AppInput v-model="form.name" placeholder="Profile name" required class="col-span-2" />
        <AppInput v-model="form.host" placeholder="SMTP host" required />
        <AppInput v-model="form.port" placeholder="Port" type="number" />
        <AppInput v-model="form.username" placeholder="Username" />
        <AppInput v-model="form.password" :placeholder="editingId ? 'Password (leave blank to keep)' : 'Password'" type="password" />
        <AppInput v-model="form.from_addr" placeholder="From address" required />
        <AppInput v-model="form.from_name" placeholder="From name" />
      </div>

      <!-- Custom headers -->
      <div>
        <div class="flex items-center justify-between mb-2">
          <div class="text-xs font-mono text-dim uppercase tracking-wider">Custom Headers</div>
          <button type="button" @click="addHeader" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">+ Add Header</button>
        </div>
        <div v-for="(header, i) in headers" :key="i" class="flex gap-2 mb-2 items-center">
          <AppInput v-model="header.key" placeholder="Header name" class="flex-1" />
          <AppInput v-model="header.value" placeholder="Value" class="flex-1" />
          <DeleteButton @click="removeHeader(i)" />
        </div>
      </div>

      <template #actions>
        <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="profiles.length === 0 && !showForm" message="No SMTP profiles. Add one to enable email delivery." />

    <DataTable v-else-if="profiles.length > 0" :columns="[{ label: 'Name' }, { label: 'Host' }, { label: 'From' }, { label: '', width: 'w-16' }]">
      <DataTableRow
        v-for="(profile, i) in profiles"
        :key="profile.id"
        :index="i"
        clickable
        @click="openEdit(profile)"
      >
        <td class="px-4 py-2.5 text-primary">{{ profile.name }}</td>
        <td class="px-4 py-2.5 text-muted">{{ profile.host }}:{{ profile.port }}</td>
        <td class="px-4 py-2.5 text-muted">
          {{ profile.from_name ? `${profile.from_name} <${profile.from_addr}>` : profile.from_addr }}
        </td>
        <td class="px-4 py-2.5 text-right">
          <DeleteButton @click.stop="remove(profile.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
