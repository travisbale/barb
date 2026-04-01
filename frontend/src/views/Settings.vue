<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
      <MiragedForm v-model="form" />
      <template #actions>
        <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        <AppButton type="submit" :disabled="enrolling">{{ enrolling ? 'Enrolling...' : 'Enroll' }}</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="connections.length === 0 && !showAdd" message="No miraged connections configured." />

    <DataTable v-else-if="connections.length > 0" :columns="[{ label: 'Name' }, { label: 'Address' }, { label: 'Status' }, { label: '', width: 'w-24' }]">
      <DataTableRow
        v-for="(conn, i) in connections"
        :key="conn.id"
        :index="i"
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
        <td class="px-4 py-2.5 text-right">
          <div class="flex items-center gap-4 justify-end">
            <AppButton variant="secondary" @click="test(conn.id)">Test</AppButton>
            <DeleteButton @click="remove(conn.id)" />
          </div>
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
