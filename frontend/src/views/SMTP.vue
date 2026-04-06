<script setup lang="ts">
import { ref } from 'vue'
import { useCrudList } from '../composables/useCrudList'
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

type FormData = { name: string; host: string; port: string; username: string; password: string; from_addr: string; from_name: string }

const emptyForm = (): FormData => ({ name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' })
const headers = ref<{ key: string; value: string }[]>([])

function headersToMap(): Record<string, string> {
  const map: Record<string, string> = {}
  for (const h of headers.value) {
    if (h.key.trim()) map[h.key.trim()] = h.value
  }
  return map
}

function toPayload(form: FormData) {
  return { ...form, port: parseInt(form.port) || 587, custom_headers: headersToMap() }
}

const { items: profiles, showForm, editingId, error, form, openCreate: rawOpenCreate, openEdit: rawOpenEdit, closeForm: rawCloseForm, submit, remove } = useCrudList<SMTPProfile, FormData>(
  {
    list: listSMTPProfiles,
    create: (f) => createSMTPProfile(toPayload(f)),
    update: (id, f) => updateSMTPProfile(id, toPayload(f)),
    remove: deleteSMTPProfile,
  },
  {
    emptyForm,
    toForm: (p) => ({ name: p.name, host: p.host, port: String(p.port), username: p.username, password: '', from_addr: p.from_addr, from_name: p.from_name }),
    confirmMessage: 'Delete this SMTP profile?',
  },
)

function addHeader() { headers.value.push({ key: '', value: '' }) }
function removeHeader(i: number) { headers.value.splice(i, 1) }

function openCreate() {
  headers.value = []
  rawOpenCreate()
}

function openEdit(profile: SMTPProfile) {
  headers.value = Object.entries(profile.custom_headers ?? {}).map(([key, value]) => ({ key, value }))
  rawOpenEdit(profile)
}

function closeForm() {
  headers.value = []
  rawCloseForm()
}
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
