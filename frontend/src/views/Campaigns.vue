<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { listCampaigns, deleteCampaign, type Campaign } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import DeleteButton from '../components/DeleteButton.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
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

    <ErrorBanner v-model="error" />

    <EmptyState v-if="campaigns.length === 0" message="No campaigns. Create one to begin an operation." />

    <DataTable v-else :columns="[{ label: 'Name' }, { label: 'Status' }, { label: 'Created' }, { label: '', width: 'w-16' }]">
      <DataTableRow
        v-for="(campaign, i) in campaigns"
        :key="campaign.id"
        :index="i"
        clickable
        @click="router.push(`/campaigns/${campaign.id}`)"
      >
        <td class="text-primary">{{ campaign.name }}</td>
        <td>
          <span class="text-xs uppercase tracking-wider" :class="statusColor[campaign.status] ?? 'text-dim'">{{ campaign.status }}</span>
        </td>
        <td class="text-dim">{{ new Date(campaign.created_at).toLocaleDateString() }}</td>
        <td class="text-right">
          <DeleteButton @click.stop="remove(campaign.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
