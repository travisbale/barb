<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { getCampaign, startCampaign, cancelCampaign, listCampaignResults, type Campaign, type CampaignResult } from '../api/client'
import AppButton from '../components/AppButton.vue'
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
      <AppButton v-if="isDraft" :disabled="starting" @click="start">
        {{ starting ? 'Starting...' : 'Start Campaign' }}
      </AppButton>
      <AppButton v-if="isActive" variant="danger" :disabled="cancelling" @click="cancel">
        {{ cancelling ? 'Cancelling...' : 'Cancel Campaign' }}
      </AppButton>
    </PageHeader>

    <ErrorBanner :message="error" />

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
