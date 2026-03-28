<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listMiraged, createMiraged, deleteMiraged, testMiraged, type MiragedConnection, type MiragedStatus } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import IconTrash from '../components/IconTrash.vue'
import AddButton from '../components/AddButton.vue'

const connections = ref<MiragedConnection[]>([])
const statuses = ref<Record<string, MiragedStatus>>({})
const showAdd = ref(false)
const error = ref('')

const form = ref({
  name: '', address: '', secret_hostname: '',
  cert_pem: '', key_pem: '', ca_cert_pem: '',
})

async function load() {
  try {
    connections.value = await listMiraged() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function add() {
  try {
    const conn = await createMiraged(form.value)
    connections.value.unshift(conn)
    form.value = { name: '', address: '', secret_hostname: '', cert_pem: '', key_pem: '', ca_cert_pem: '' }
    showAdd.value = false
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(id: string) {
  try {
    await deleteMiraged(id)
    connections.value = connections.value.filter(c => c.id !== id)
    delete statuses.value[id]
  } catch (e: any) {
    error.value = e.message
  }
}

async function test(id: string) {
  try {
    statuses.value[id] = { connected: false, error: 'Testing...' }
    statuses.value[id] = await testMiraged(id)
  } catch (e: any) {
    statuses.value[id] = { connected: false, error: e.message }
  }
}

async function readFile(event: Event, field: 'cert_pem' | 'key_pem' | 'ca_cert_pem') {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  form.value[field] = await file.text()
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Miraged" :subtitle="`${connections.length} connections`">
      <AddButton @click="showAdd = true">Add Connection</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showAdd" class="p-7 mb-4">
      <form @submit.prevent="add" class="flex flex-col gap-7">
        <div class="grid grid-cols-3 gap-5">
          <AppInput v-model="form.name" placeholder="Name (required)" required />
          <AppInput v-model="form.address" placeholder="Address (host:port)" required />
          <AppInput v-model="form.secret_hostname" placeholder="Secret hostname" required />
        </div>

        <div class="grid grid-cols-3 gap-5">
          <div>
            <label class="block text-xs text-dim font-mono mb-1 uppercase tracking-wider">Client Cert</label>
            <input type="file" accept=".crt,.pem" @change="readFile($event, 'cert_pem')"
              class="text-xs text-muted font-mono file:mr-2 file:px-3 file:py-1 file:border file:border-edge file:bg-surface file:text-muted file:font-mono file:text-xs file:cursor-pointer" />
          </div>
          <div>
            <label class="block text-xs text-dim font-mono mb-1 uppercase tracking-wider">Client Key</label>
            <input type="file" accept=".key,.pem" @change="readFile($event, 'key_pem')"
              class="text-xs text-muted font-mono file:mr-2 file:px-3 file:py-1 file:border file:border-edge file:bg-surface file:text-muted file:font-mono file:text-xs file:cursor-pointer" />
          </div>
          <div>
            <label class="block text-xs text-dim font-mono mb-1 uppercase tracking-wider">CA Cert</label>
            <input type="file" accept=".crt,.pem" @change="readFile($event, 'ca_cert_pem')"
              class="text-xs text-muted font-mono file:mr-2 file:px-3 file:py-1 file:border file:border-edge file:bg-surface file:text-muted file:font-mono file:text-xs file:cursor-pointer" />
          </div>
        </div>

        <div class="flex gap-2 pt-1">
          <AppButton type="submit">Add</AppButton>
          <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="connections.length === 0 && !showAdd" message="No miraged connections configured." />

    <Card v-else-if="connections.length > 0" class="overflow-hidden">
      <table class="w-full text-sm font-mono">
        <thead>
          <tr class="border-b border-edge text-dim text-left uppercase tracking-wider">
            <th class="px-4 py-2.5 font-medium">Name</th>
            <th class="px-4 py-2.5 font-medium">Address</th>
            <th class="px-4 py-2.5 font-medium">Status</th>
            <th class="px-4 py-2.5 font-medium w-24"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(conn, i) in connections"
            :key="conn.id"
            :style="{ animationDelay: `${i * 20}ms` }"
            class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
          >
            <td class="px-4 py-2.5 text-primary">{{ conn.name }}</td>
            <td class="px-4 py-2.5 text-muted">{{ conn.address }}</td>
            <td class="px-4 py-2.5">
              <span v-if="!statuses[conn.id]" class="text-dim">—</span>
              <span v-else-if="statuses[conn.id].connected" class="text-teal">
                <span class="inline-block w-1.5 h-1.5 bg-teal rounded-full mr-1.5 align-middle"></span>
                {{ statuses[conn.id].version }}
              </span>
              <span v-else class="text-danger text-xs">{{ statuses[conn.id].error }}</span>
            </td>
            <td class="px-4 py-2.5 text-right flex gap-2 justify-end">
              <AppButton variant="secondary" @click="test(conn.id)">Test</AppButton>
              <button @click="remove(conn.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
