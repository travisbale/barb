<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import {
  listMiraged, updateMiraged,
  listMiragedNotifications, createMiragedNotification, deleteMiragedNotification, testMiragedNotification,
  listMiragedNotificationEventTypes,
  type MiragedConnection, type MiragedNotificationChannel,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import FormCard from '../components/FormCard.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import AddButton from '../components/AddButton.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import DeleteButton from '../components/DeleteButton.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import PillSelector from '../components/PillSelector.vue'
import RenameForm from '../components/RenameForm.vue'

const route = useRoute()
const { confirm } = useConfirm()
const id = route.params.id as string

const conn = ref<MiragedConnection | null>(null)
const error = ref('')
const saving = ref(false)

// Notification channels — immediate CRUD, no buffering.
const channels = ref<MiragedNotificationChannel[]>([])
const eventTypes = ref<string[]>([])
const showAddChannel = ref(false)
const addingChannel = ref(false)
const testStatus = ref<Record<string, 'sending' | 'sent'>>({})
const channelForm = ref<{ type: 'webhook' | 'slack'; url: string; auth_header: string; filter: string[] }>({
  type: 'slack', url: '', auth_header: '', filter: [],
})

async function load() {
  try {
    const [all, chans, types] = await Promise.all([
      listMiraged(),
      listMiragedNotifications(id),
      listMiragedNotificationEventTypes(id),
    ])
    const found = (all ?? []).find((c) => c.id === id)
    if (!found) {
      error.value = 'Connection not found.'
      return
    }
    conn.value = found
    channels.value = chans ?? []
    eventTypes.value = types ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function saveName(name: string) {
  saving.value = true
  error.value = ''
  try {
    conn.value = await updateMiraged(id, { name })
  } catch (e: any) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

async function addChannel() {
  addingChannel.value = true
  error.value = ''
  try {
    const created = await createMiragedNotification(id, {
      type: channelForm.value.type,
      url: channelForm.value.url,
      auth_header: channelForm.value.auth_header || undefined,
      filter: channelForm.value.filter.length > 0 ? channelForm.value.filter : undefined,
    })
    channels.value.push(created)
    channelForm.value = { type: 'slack', url: '', auth_header: '', filter: [] }
    showAddChannel.value = false
  } catch (e: any) {
    error.value = e.message
  } finally {
    addingChannel.value = false
  }
}

async function removeChannel(channelId: string) {
  if (!await confirm('Delete this notification channel?')) return
  try {
    await deleteMiragedNotification(id, channelId)
    channels.value = channels.value.filter((c) => c.id !== channelId)
  } catch (e: any) {
    error.value = e.message
  }
}

async function testChannel(channelId: string) {
  error.value = ''
  testStatus.value[channelId] = 'sending'
  try {
    await testMiragedNotification(id, channelId)
    testStatus.value[channelId] = 'sent'
    setTimeout(() => {
      delete testStatus.value[channelId]
    }, 2000)
  } catch (e: any) {
    delete testStatus.value[channelId]
    error.value = e.message
  }
}

function testButtonLabel(channelId: string): string {
  switch (testStatus.value[channelId]) {
    case 'sending': return 'Sending...'
    case 'sent': return 'Sent ✓'
    default: return 'Test'
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader
      :title="conn?.name ?? '...'"
      :breadcrumbs="[{ label: 'Miraged', to: '/miraged' }, { label: conn?.name ?? '...' }]"
    />

    <ErrorBanner v-model="error" />

    <RenameForm v-if="conn" :value="conn.name" :saving="saving" @save="saveName" />

    <!-- Notification Channels -->
    <div v-if="conn" class="mt-12">
      <div class="flex items-center gap-2 mb-4">
        <h3 class="text-base text-primary flex-1">Notification Channels</h3>
        <AddButton v-if="!showAddChannel" @click="showAddChannel = true">Add Channel</AddButton>
      </div>

      <FormCard v-if="showAddChannel" @submit="addChannel">
        <div class="grid grid-cols-2 gap-3">
          <AppSelect v-model="channelForm.type" label="Type" required>
            <option value="slack">Slack</option>
            <option value="webhook">Webhook</option>
          </AppSelect>
          <AppInput v-model="channelForm.url" placeholder="Destination URL" required />
        </div>
        <AppInput v-if="channelForm.type === 'webhook'" v-model="channelForm.auth_header" placeholder="Authorization header (optional)" />

        <div class="flex flex-col gap-3">
          <span class="text-xs text-dim font-mono pl-3">Event filter (none selected = all events)</span>
          <PillSelector v-model="channelForm.filter" :options="eventTypes" />
        </div>

        <template #actions>
          <AppButton variant="ghost" type="button" @click="showAddChannel = false">Cancel</AppButton>
          <AppButton type="submit" :disabled="addingChannel">{{ addingChannel ? 'Adding...' : 'Add' }}</AppButton>
        </template>
      </FormCard>

      <EmptyState v-if="channels.length === 0 && !showAddChannel" message="No notification channels configured." />

      <DataTable
        v-else-if="channels.length > 0"
        :columns="[{ label: 'Type' }, { label: 'URL' }, { label: 'Filter' }, { label: '', width: 'w-32' }]"
      >
        <DataTableRow v-for="(ch, i) in channels" :key="ch.id" :index="i">
          <td class="text-primary capitalize">{{ ch.type }}</td>
          <td class="text-muted truncate">{{ ch.url }}</td>
          <td class="text-muted text-xs">{{ ch.filter.length === 0 ? 'all events' : ch.filter.join(', ') }}</td>
          <td class="text-right">
            <AppButton
              variant="ghost"
              type="button"
              :disabled="testStatus[ch.id] !== undefined"
              @click="testChannel(ch.id)"
            >{{ testButtonLabel(ch.id) }}</AppButton>
            <DeleteButton @click="removeChannel(ch.id)" />
          </td>
        </DataTableRow>
      </DataTable>
    </div>
  </div>
</template>
