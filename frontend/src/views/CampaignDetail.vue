<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getCampaign, listCampaignResults, type Campaign, type CampaignResult } from '../api/client'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const campaign = ref<Campaign | null>(null)
const results = ref<CampaignResult[]>([])
const error = ref('')

async function load() {
  try {
    campaign.value = await getCampaign(id)
    results.value = await listCampaignResults(id) ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

const statusColor: Record<string, string> = {
  pending: 'text-dim',
  sent: 'text-muted',
  clicked: 'text-amber',
  captured: 'text-teal',
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center gap-3 mb-8">
      <button
        @click="router.push('/campaigns')"
        class="text-dim hover:text-amber font-mono text-sm transition-colors"
      >&larr;</button>
      <div>
        <h1 class="text-xl font-mono font-semibold tracking-tight text-primary">
          {{ campaign?.name ?? '...' }}
        </h1>
        <div class="flex items-center gap-3 mt-1">
          <span class="text-xs font-mono uppercase tracking-wider" :class="{
            'text-dim': campaign?.status === 'draft',
            'text-teal': campaign?.status === 'active',
            'text-amber': campaign?.status === 'paused',
            'text-muted': campaign?.status === 'completed',
          }">{{ campaign?.status }}</span>
          <span class="text-xs text-dim font-mono">{{ results.length }} targets</span>
        </div>
      </div>
    </div>

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
            v-for="(r, i) in results"
            :key="r.id"
            :style="{ animationDelay: `${i * 20}ms` }"
            class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
          >
            <td class="px-4 py-2.5 text-primary">{{ r.email }}</td>
            <td class="px-4 py-2.5">
              <span :class="statusColor[r.status] ?? 'text-dim'" class="uppercase text-xs tracking-wider">
                {{ r.status }}
              </span>
            </td>
            <td class="px-4 py-2.5 text-dim">{{ r.sent_at ? new Date(r.sent_at).toLocaleString() : '—' }}</td>
            <td class="px-4 py-2.5 text-dim">{{ r.clicked_at ? new Date(r.clicked_at).toLocaleString() : '—' }}</td>
            <td class="px-4 py-2.5 text-dim">{{ r.captured_at ? new Date(r.captured_at).toLocaleString() : '—' }}</td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
