<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { listCampaigns, deleteCampaign, type Campaign } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import IconTrash from '../components/IconTrash.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const router = useRouter()
const { confirm } = useConfirm()
const campaigns = ref<Campaign[]>([])
const error = ref('')

async function load() {
  try {
    campaigns.value = await listCampaigns() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(id: string) {
  if (!await confirm('Delete this campaign?')) return
  try {
    await deleteCampaign(id)
    campaigns.value = campaigns.value.filter(c => c.id !== id)
  } catch (e: any) {
    error.value = e.message
  }
}

const statusColor: Record<string, string> = {
  draft: 'text-dim',
  active: 'text-teal',
  paused: 'text-amber',
  completed: 'text-muted',
  cancelled: 'text-danger',
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Campaigns" :subtitle="`${campaigns.length} campaigns`">
      <AddButton @click="router.push('/campaigns/new')">New Campaign</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <EmptyState v-if="campaigns.length === 0" message="No campaigns. Create one to begin an operation." />

    <Card v-else>
      <div
        v-for="(campaign, i) in campaigns"
        :key="campaign.id"
        :style="{ animationDelay: `${i * 30}ms` }"
        class="animate-in flex items-center justify-between px-4 py-3 border-b border-edge last:border-0 hover:bg-surface-hover cursor-pointer transition-colors"
        @click="router.push(`/campaigns/${campaign.id}`)"
      >
        <div class="flex items-center gap-3">
          <span class="text-xs font-mono uppercase tracking-wider" :class="statusColor[campaign.status] ?? 'text-dim'">
            {{ campaign.status }}
          </span>
          <div>
            <div class="text-sm font-medium text-primary">{{ campaign.name }}</div>
            <div class="text-xs text-dim font-mono mt-0.5">{{ new Date(campaign.created_at).toLocaleDateString() }}</div>
          </div>
        </div>
        <button @click.stop="remove(campaign.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
      </div>
    </Card>
  </div>
</template>
