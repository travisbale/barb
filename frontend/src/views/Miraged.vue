<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listMiraged, enrollMiraged, updateMiraged, deleteMiraged, testMiraged, type MiragedConnection, type MiragedStatus } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import MiragedForm from '../components/MiragedForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import FormCard from '../components/FormCard.vue'
import DeleteButton from '../components/DeleteButton.vue'
import AddButton from '../components/AddButton.vue'

const { confirm } = useConfirm()
const connections = ref<MiragedConnection[]>([])
const statuses = ref<Record<string, MiragedStatus>>({})
const showAdd = ref(false)
const enrolling = ref(false)
const error = ref('')

const form = ref({ name: '', address: '', secret_hostname: '', token: '' })

// Rename state.
const editingId = ref<string | null>(null)
const editName = ref('')
const saving = ref(false)

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

function startEdit(conn: MiragedConnection) {
  editingId.value = conn.id
  editName.value = conn.name
  error.value = ''
}

function cancelEdit() {
  editingId.value = null
  editName.value = ''
}

async function saveEdit() {
  if (!editingId.value) return
  saving.value = true
  error.value = ''
  try {
    const updated = await updateMiraged(editingId.value, { name: editName.value })
    const idx = connections.value.findIndex(c => c.id === updated.id)
    if (idx !== -1) connections.value[idx] = updated
    editingId.value = null
  } catch (e: any) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

async function remove(id: string) {
  if (!await confirm('Delete this connection?')) return
  try {
    await deleteMiraged(id)
    connections.value = connections.value.filter(c => c.id !== id)
    delete statuses.value[id]
    if (editingId.value === id) cancelEdit()
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

    <ErrorBanner :message="error" />

    <FormCard v-if="showAdd" @submit="add">
      <MiragedForm v-model="form" />
      <template #actions>
        <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        <AppButton type="submit" :disabled="enrolling">{{ enrolling ? 'Enrolling...' : 'Enroll' }}</AppButton>
      </template>
    </FormCard>

    <!-- Rename form -->
    <FormCard v-if="editingId" @submit="saveEdit">
      <AppInput v-model="editName" placeholder="Name" required autofocus />
      <template #actions>
        <AppButton variant="ghost" @click="cancelEdit">Cancel</AppButton>
        <AppButton type="submit" :disabled="saving">{{ saving ? 'Saving...' : 'Save' }}</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="connections.length === 0 && !showAdd" message="No miraged connections configured." />

    <DataTable v-else-if="connections.length > 0" :columns="[{ label: 'Name' }, { label: 'Address' }, { label: 'Status' }, { label: '', width: 'w-12' }]">
      <DataTableRow
        v-for="(conn, i) in connections"
        :key="conn.id"
        :index="i"
        clickable
        @click="startEdit(conn)"
      >
        <td class="px-4 py-2.5 text-primary">{{ conn.name }}</td>
        <td class="px-4 py-2.5 text-muted">{{ conn.address }}</td>
        <td class="px-4 py-2.5">
          <span v-if="!statuses[conn.id]" class="text-dim text-xs">Checking...</span>
          <span v-else-if="statuses[conn.id].connected" class="text-teal">
            <span class="inline-block w-1.5 h-1.5 bg-teal rounded-full mr-1.5 align-middle"></span>
            Connected
          </span>
          <span v-else class="text-danger text-xs">{{ statuses[conn.id].error }}</span>
        </td>
        <td class="px-4 py-2.5 text-right" @click.stop>
          <DeleteButton @click="remove(conn.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
