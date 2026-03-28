<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listSMTPProfiles, createSMTPProfile, deleteSMTPProfile, type SMTPProfile } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'

const profiles = ref<SMTPProfile[]>([])
const showCreate = ref(false)
const error = ref('')

const form = ref({
  name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '',
})

async function load() {
  try {
    profiles.value = await listSMTPProfiles() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function create() {
  try {
    const profile = await createSMTPProfile({
      name: form.value.name,
      host: form.value.host,
      port: parseInt(form.value.port) || 587,
      username: form.value.username,
      password: form.value.password,
      from_addr: form.value.from_addr,
      from_name: form.value.from_name,
    })
    profiles.value.unshift(profile)
    form.value = { name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' }
    showCreate.value = false
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(id: string) {
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
      <AppButton @click="showCreate = true">+ Add Profile</AppButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showCreate" class="p-4 mb-4">
      <form @submit.prevent="create" class="grid grid-cols-2 gap-3">
        <AppInput v-model="form.name" placeholder="Profile name (required)" required class="col-span-2" />
        <AppInput v-model="form.host" placeholder="SMTP host (required)" required />
        <AppInput v-model="form.port" placeholder="Port" type="number" />
        <AppInput v-model="form.username" placeholder="Username" />
        <AppInput v-model="form.password" placeholder="Password" type="password" />
        <AppInput v-model="form.from_addr" placeholder="From address (required)" required />
        <AppInput v-model="form.from_name" placeholder="From name" />
        <div class="col-span-2 flex gap-2">
          <AppButton type="submit">Create</AppButton>
          <AppButton variant="ghost" @click="showCreate = false">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="profiles.length === 0 && !showCreate" message="No SMTP profiles. Add one to enable email delivery." />

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
            v-for="(p, i) in profiles"
            :key="p.id"
            :style="{ animationDelay: `${i * 20}ms` }"
            class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
          >
            <td class="px-4 py-2.5 text-gray-200">{{ p.name }}</td>
            <td class="px-4 py-2.5 text-muted">{{ p.host }}:{{ p.port }}</td>
            <td class="px-4 py-2.5 text-muted">
              {{ p.from_name ? `${p.from_name} <${p.from_addr}>` : p.from_addr }}
            </td>
            <td class="px-4 py-2.5 text-right">
              <AppButton variant="danger" @click="remove(p.id)">Del</AppButton>
            </td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
