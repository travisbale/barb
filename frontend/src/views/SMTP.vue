<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listSMTPProfiles, createSMTPProfile, updateSMTPProfile, deleteSMTPProfile, type SMTPProfile } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const { confirm } = useConfirm()
const profiles = ref<SMTPProfile[]>([])
const showForm = ref(false)
const editingId = ref<string | null>(null)
const error = ref('')

const emptyForm = { name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' }
const form = ref({ ...emptyForm })

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
  showForm.value = true
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

    <Card v-if="showForm" class="p-7 mb-4">
      <form @submit.prevent="submit" class="grid grid-cols-2 gap-7">
        <AppInput v-model="form.name" placeholder="Profile name (required)" required class="col-span-2" />
        <AppInput v-model="form.host" placeholder="SMTP host (required)" required />
        <AppInput v-model="form.port" placeholder="Port" type="number" />
        <AppInput v-model="form.username" placeholder="Username" />
        <AppInput v-model="form.password" :placeholder="editingId ? 'Password (leave blank to keep)' : 'Password'" type="password" />
        <AppInput v-model="form.from_addr" placeholder="From address (required)" required />
        <AppInput v-model="form.from_name" placeholder="From name" />
        <div class="col-span-2 flex gap-2">
          <AppButton type="submit">{{ editingId ? 'Save' : 'Create' }}</AppButton>
          <AppButton variant="ghost" @click="closeForm">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="profiles.length === 0 && !showForm" message="No SMTP profiles. Add one to enable email delivery." />

    <Card v-else-if="profiles.length > 0" class="overflow-hidden">
      <table class="w-full text-sm font-mono">
        <thead>
          <tr class="border-b border-edge text-dim text-left uppercase tracking-wider">
            <th class="px-4 py-2.5 font-medium">Name</th>
            <th class="px-4 py-2.5 font-medium">Host</th>
            <th class="px-4 py-2.5 font-medium">From</th>
            <th class="px-4 py-2.5 font-medium w-16"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(profile, i) in profiles"
            :key="profile.id"
            :style="{ animationDelay: `${i * 20}ms` }"
            class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover cursor-pointer transition-colors"
            @click="openEdit(profile)"
          >
            <td class="px-4 py-2.5 text-primary">{{ profile.name }}</td>
            <td class="px-4 py-2.5 text-muted">{{ profile.host }}:{{ profile.port }}</td>
            <td class="px-4 py-2.5 text-muted">
              {{ profile.from_name ? `${profile.from_name} <${profile.from_addr}>` : profile.from_addr }}
            </td>
            <td class="px-4 py-2.5 text-right">
              <button @click.stop="remove(profile.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
