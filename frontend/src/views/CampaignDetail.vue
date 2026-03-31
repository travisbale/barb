<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  getCampaign, startCampaign, cancelCampaign, sendTestEmail, listCampaignResults,
  getMiragedSession, exportMiragedSessionCookies,
  type Campaign, type CampaignResult, type MiragedSession,
} from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import PageHeader from '../components/PageHeader.vue'

const route = useRoute()
const id = route.params.id as string

const campaign = ref<Campaign | null>(null)
const results = ref<CampaignResult[]>([])
const error = ref('')
const starting = ref(false)
const cancelling = ref(false)
const showTestEmail = ref(false)
const testEmailAddress = ref('')
const testEmailSending = ref(false)
const testEmailStatus = ref('')

// Session detail state.
const selectedSession = ref<MiragedSession | null>(null)
const sessionLoading = ref(false)
const sessionError = ref('')

const isDraft = computed(() => campaign.value?.status === 'draft')
const isActive = computed(() => campaign.value?.status === 'active')

async function load() {
  try {
    campaign.value = await getCampaign(id)
    results.value = await listCampaignResults(id) ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function start() {
  starting.value = true
  error.value = ''
  try {
    await startCampaign(id)
    await load()
    startPolling()
  } catch (e: any) {
    error.value = e.message
  } finally {
    starting.value = false
  }
}

async function cancel() {
  if (!confirm('Cancel this campaign? Sending will stop immediately.')) return
  cancelling.value = true
  error.value = ''
  try {
    await cancelCampaign(id)
    await load()
    stopPolling()
  } catch (e: any) {
    error.value = e.message
  } finally {
    cancelling.value = false
  }
}

async function sendTest() {
  testEmailSending.value = true
  testEmailStatus.value = ''
  try {
    await sendTestEmail(id, testEmailAddress.value)
    testEmailStatus.value = 'Test email sent'
    showTestEmail.value = false
  } catch (e: any) {
    testEmailStatus.value = e.message
  } finally {
    testEmailSending.value = false
  }
}

// Auto-refresh while the campaign is active.
let pollInterval: ReturnType<typeof setInterval> | null = null

function startPolling() {
  stopPolling()
  pollInterval = setInterval(async () => {
    await load()
    if (!isActive.value) stopPolling()
  }, 2000)
}

function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

const statusColor: Record<string, string> = {
  pending: 'text-dim',
  sent: 'text-muted',
  failed: 'text-danger',
  clicked: 'text-amber',
  captured: 'text-teal',
}

const sentCount = computed(() => results.value.filter(result => result.status !== 'pending').length)
const totalCount = computed(() => results.value.length)

async function viewSession(result: CampaignResult) {
  if (!result.session_id || !campaign.value?.miraged_id) return
  stopPolling()
  selectedSession.value = null
  sessionError.value = ''
  sessionLoading.value = true
  try {
    selectedSession.value = await getMiragedSession(campaign.value.miraged_id, result.session_id)
  } catch (e: any) {
    sessionError.value = e.message
  } finally {
    sessionLoading.value = false
  }
}

function closeSession() {
  selectedSession.value = null
  sessionError.value = ''
  if (isActive.value) startPolling()
}

function downloadCookies() {
  if (!selectedSession.value || !campaign.value?.miraged_id) return
  const url = exportMiragedSessionCookies(campaign.value.miraged_id, selectedSession.value.id)
  window.open(url, '_blank')
}

function exportCSV() {
  if (results.value.length === 0) return

  const headers = ['email', 'status', 'sent_at', 'clicked_at', 'captured_at', 'session_id']
  const rows = results.value.map(result => [
    result.email,
    result.status,
    result.sent_at ?? '',
    result.clicked_at ?? '',
    result.captured_at ?? '',
    result.session_id,
  ].map(field => `"${String(field).replace(/"/g, '""')}"`).join(','))

  const csv = [headers.join(','), ...rows].join('\n')
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)

  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `${campaign.value?.name ?? 'campaign'}-results.csv`
  anchor.click()

  URL.revokeObjectURL(url)
}

onMounted(async () => {
  await load()
  if (isActive.value) startPolling()
})

onUnmounted(stopPolling)
</script>

<template>
  <div>
    <PageHeader
      :title="campaign?.name ?? '...'"
      :breadcrumbs="[{ label: 'Campaigns', to: '/campaigns' }, { label: campaign?.name ?? '...' }]"
    >
      <template #subtitle>
        <div class="flex flex-col gap-1 mt-1">
          <div class="flex items-center gap-3">
            <span class="text-xs font-mono uppercase tracking-wider" :class="{
              'text-dim': campaign?.status === 'draft',
              'text-teal': campaign?.status === 'active',
              'text-amber': campaign?.status === 'paused',
              'text-muted': campaign?.status === 'completed',
              'text-danger': campaign?.status === 'cancelled',
            }">{{ campaign?.status }}</span>
            <span class="text-xs text-dim font-mono">{{ sentCount }}/{{ totalCount }} sent</span>
          </div>
          <span v-if="campaign?.lure_url" class="text-xs text-dim font-mono select-all">{{ campaign.lure_url }}</span>
        </div>
      </template>
      <AppButton v-if="results.length > 0" variant="secondary" @click="exportCSV">Export CSV</AppButton>
      <AppButton v-if="isDraft" variant="secondary" @click="showTestEmail = !showTestEmail">Send Test</AppButton>
      <AppButton v-if="isDraft" :disabled="starting" @click="start">
        {{ starting ? 'Starting...' : 'Start Campaign' }}
      </AppButton>
      <AppButton v-if="isActive" variant="danger" :disabled="cancelling" @click="cancel">
        {{ cancelling ? 'Cancelling...' : 'Cancel Campaign' }}
      </AppButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <!-- Test email -->
    <Card v-if="showTestEmail" class="p-5 mb-4">
      <form @submit.prevent="sendTest" class="flex items-end gap-3">
        <AppInput v-model="testEmailAddress" type="email" placeholder="Recipient email" required class="flex-1" />
        <AppButton variant="ghost" @click="showTestEmail = false">Cancel</AppButton>
        <AppButton type="submit" :disabled="testEmailSending || !testEmailAddress">
          {{ testEmailSending ? 'Sending...' : 'Send' }}
        </AppButton>
      </form>
      <div v-if="testEmailStatus" class="text-xs font-mono mt-2" :class="testEmailStatus.startsWith('Test email') ? 'text-teal' : 'text-danger'">
        {{ testEmailStatus }}
      </div>
    </Card>

    <!-- Session detail panel -->
    <Card v-if="selectedSession || sessionLoading || sessionError" class="p-7 mb-4">
      <div class="flex items-center justify-between mb-5">
        <div class="text-xs font-mono text-dim uppercase tracking-wider">Session Details</div>
        <button @click="closeSession" class="text-xs font-mono text-dim hover:text-primary transition-colors uppercase tracking-wider">Close</button>
      </div>

      <div v-if="sessionLoading" class="text-sm font-mono text-dim">Loading session...</div>
      <ErrorBanner v-else-if="sessionError" :message="sessionError" />

      <div v-else-if="selectedSession" class="flex flex-col gap-5">
        <!-- Credentials -->
        <div v-if="selectedSession.username || selectedSession.password">
          <div class="text-xs font-mono text-dim uppercase tracking-wider mb-2">Credentials</div>
          <div class="grid grid-cols-2 gap-3">
            <div v-if="selectedSession.username" class="px-3 py-2 bg-bg border border-edge">
              <div class="text-xs text-dim font-mono">Username</div>
              <div class="text-sm text-teal font-mono select-all">{{ selectedSession.username }}</div>
            </div>
            <div v-if="selectedSession.password" class="px-3 py-2 bg-bg border border-edge">
              <div class="text-xs text-dim font-mono">Password</div>
              <div class="text-sm text-teal font-mono select-all">{{ selectedSession.password }}</div>
            </div>
          </div>
        </div>

        <!-- Custom fields -->
        <div v-if="selectedSession.custom && Object.keys(selectedSession.custom).length > 0">
          <div class="text-xs font-mono text-dim uppercase tracking-wider mb-2">Custom Fields</div>
          <div class="flex flex-col gap-1">
            <div v-for="(value, key) in selectedSession.custom" :key="key" class="px-3 py-2 bg-bg border border-edge flex justify-between">
              <span class="text-xs text-dim font-mono">{{ key }}</span>
              <span class="text-sm text-primary font-mono select-all">{{ value }}</span>
            </div>
          </div>
        </div>

        <!-- Cookies -->
        <div v-if="selectedSession.cookie_tokens && Object.keys(selectedSession.cookie_tokens).length > 0">
          <div class="flex items-center justify-between mb-2">
            <div class="text-xs font-mono text-dim uppercase tracking-wider">Cookies</div>
            <AppButton variant="secondary" @click="downloadCookies">Export Cookies</AppButton>
          </div>
          <div v-for="(cookies, domain) in selectedSession.cookie_tokens" :key="domain" class="mb-3">
            <div class="text-xs text-muted font-mono mb-1">{{ domain }}</div>
            <div class="flex flex-col gap-1">
              <div v-for="(value, name) in cookies" :key="name" class="px-3 py-1.5 bg-bg border border-edge flex justify-between gap-4">
                <span class="text-xs text-dim font-mono shrink-0">{{ name }}</span>
                <span class="text-xs text-primary font-mono select-all truncate">{{ value }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Metadata -->
        <div class="text-xs font-mono text-dim flex flex-wrap gap-4">
          <span v-if="selectedSession.remote_addr">IP: {{ selectedSession.remote_addr }}</span>
          <span v-if="selectedSession.phishlet">Phishlet: {{ selectedSession.phishlet }}</span>
          <span v-if="selectedSession.started_at">Started: {{ new Date(selectedSession.started_at).toLocaleString() }}</span>
        </div>

        <div v-if="selectedSession.user_agent" class="text-xs font-mono text-dim break-all">
          {{ selectedSession.user_agent }}
        </div>
      </div>
    </Card>

    <EmptyState v-if="results.length === 0" message="No results yet." />

    <Card v-else class="overflow-hidden">
      <table class="w-full text-sm font-mono">
        <thead>
          <tr class="border-b border-edge text-dim text-left uppercase tracking-wider">
            <th class="px-4 py-2.5 font-medium">Email</th>
            <th class="px-4 py-2.5 font-medium">Status</th>
            <th class="px-4 py-2.5 font-medium">Sent</th>
            <th class="px-4 py-2.5 font-medium">Clicked</th>
            <th class="px-4 py-2.5 font-medium">Captured</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(result, i) in results"
            :key="result.id"
            :style="{ animationDelay: `${i * 20}ms` }"
            class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
            :class="{ 'cursor-pointer': result.session_id }"
            @click="result.session_id ? viewSession(result) : null"
          >
            <td class="px-4 py-2.5 text-primary">{{ result.email }}</td>
            <td class="px-4 py-2.5">
              <span :class="statusColor[result.status] ?? 'text-dim'" class="uppercase text-xs tracking-wider">
                {{ result.status }}
              </span>
            </td>
            <td class="px-4 py-2.5 text-dim">{{ result.sent_at ? new Date(result.sent_at).toLocaleString() : '—' }}</td>
            <td class="px-4 py-2.5 text-dim">{{ result.clicked_at ? new Date(result.clicked_at).toLocaleString() : '—' }}</td>
            <td class="px-4 py-2.5 text-dim">{{ result.captured_at ? new Date(result.captured_at).toLocaleString() : '—' }}</td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
