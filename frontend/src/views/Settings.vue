<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listMiraged, enrollMiraged, deleteMiraged, testMiraged, type MiragedConnection, type MiragedStatus } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import FormCard from '../components/FormCard.vue'
import IconTrash from '../components/IconTrash.vue'
import AddButton from '../components/AddButton.vue'

const { confirm } = useConfirm()
const connections = ref<MiragedConnection[]>([])
const statuses = ref<Record<string, MiragedStatus>>({})
const showAdd = ref(false)
const enrolling = ref(false)
const error = ref('')

const form = ref({ name: '', address: '', secret_hostname: '', token: '' })

async function load() {
  try {
    connections.value = await listMiraged() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function add() {
  enrolling.value = true
  error.value = ''
  try {
    const conn = await enrollMiraged(form.value)
    connections.value.unshift(conn)
    form.value = { name: '', address: '', secret_hostname: '', token: '' }
    showAdd.value = false
  } catch (e: any) {
    error.value = e.message
  } finally {
    enrolling.value = false
  }
}

async function remove(id: string) {
  if (!await confirm('Delete this connection?')) return
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

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Miraged" :subtitle="`${connections.length} connections`">
      <AddButton @click="showAdd = true">Add Connection</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <FormCard v-if="showAdd" @submit="add">
      <div class="grid grid-cols-2 gap-5">
        <AppInput v-model="form.name" placeholder="Name (required)" required />
        <AppInput v-model="form.address" placeholder="Address (host:port)" required />
      </div>
      <div class="grid grid-cols-2 gap-5">
        <AppInput v-model="form.secret_hostname" placeholder="Secret hostname" required />
        <AppInput v-model="form.token" placeholder="Invite token" required />
      </div>
      <template #actions>
        <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        <AppButton type="submit" :disabled="enrolling">{{ enrolling ? 'Enrolling...' : 'Enroll' }}</AppButton>
      </template>
    </FormCard>

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
