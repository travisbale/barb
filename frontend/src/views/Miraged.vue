<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { listMiraged, enrollMiraged, deleteMiraged, testMiraged, type MiragedConnection, type MiragedStatus } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import MiragedForm from '../components/MiragedForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import FormCard from '../components/FormCard.vue'
import DeleteButton from '../components/DeleteButton.vue'
import AddButton from '../components/AddButton.vue'

const router = useRouter()
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
    router.push(`/miraged/${conn.id}`)
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

const POLL_INTERVAL = 30_000
let pollTimer: ReturnType<typeof setInterval> | null = null

async function testAll() {
  await Promise.all(connections.value.map(async (conn) => {
    try {
      statuses.value[conn.id] = await testMiraged(conn.id)
    } catch (e: any) {
      statuses.value[conn.id] = { connected: false, error: e.message }
    }
  }))
}

function startPolling() {
  stopPolling()
  if (connections.value.length > 0) {
    testAll()
    pollTimer = setInterval(testAll, POLL_INTERVAL)
  }
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

// Restart polling when the connection list changes.
watch(() => connections.value.length, startPolling)

onMounted(load)
onUnmounted(stopPolling)
</script>

<template>
  <div>
    <PageHeader title="Miraged" :subtitle="`${connections.length} connections`">
      <AddButton @click="showAdd = true">Add Connection</AddButton>
    </PageHeader>

    <ErrorBanner v-model="error" />

    <FormCard v-if="showAdd" @submit="add">
      <MiragedForm v-model="form" />
      <template #actions>
        <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        <AppButton type="submit" :disabled="enrolling">{{ enrolling ? 'Enrolling...' : 'Enroll' }}</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="connections.length === 0 && !showAdd" message="No miraged connections configured." />

    <DataTable v-else-if="connections.length > 0" :columns="[{ label: 'Name' }, { label: 'Address' }, { label: 'Status' }, { label: '', width: 'w-12' }]">
      <DataTableRow
        v-for="(conn, i) in connections"
        :key="conn.id"
        :index="i"
        clickable
        @click="router.push(`/miraged/${conn.id}`)"
      >
        <td class="text-primary">{{ conn.name }}</td>
        <td class="text-muted">{{ conn.address }}</td>
        <td>
          <span v-if="!statuses[conn.id]" class="text-dim text-xs">Checking...</span>
          <span v-else-if="statuses[conn.id].connected" class="text-teal">
            <span class="inline-block w-1.5 h-1.5 bg-teal rounded-full mr-1.5 align-middle"></span>
            Connected
          </span>
          <span v-else class="text-danger text-xs">{{ statuses[conn.id].error }}</span>
        </td>
        <td class="text-right" @click.stop>
          <DeleteButton @click="remove(conn.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
