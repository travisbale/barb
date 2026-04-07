<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard, type DashboardStats } from '../api/client'
import AppButton from '../components/AppButton.vue'
import Card from '../components/Card.vue'
import MetricCard from '../components/MetricCard.vue'
import ErrorBanner from '../components/ErrorBanner.vue'

const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const error = ref('')

const isEmpty = computed(() =>
  stats.value != null &&
  stats.value.campaigns.total === 0 &&
  stats.value.miraged_count === 0
)

async function load() {
  try {
    stats.value = await getDashboard()
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="mb-8">
      <h1>Dashboard</h1>
      <p class="text-sm text-dim font-mono mt-1">Operations overview</p>
    </div>

    <ErrorBanner v-model="error" />

    <!-- Empty state onboarding -->
    <Card v-if="isEmpty" class="p-10 text-center">
      <div class="font-mono text-2xl font-bold tracking-widest text-amber uppercase mb-5">Welcome to Barb</div>
      <p class="text-sm text-muted font-mono mb-10">
        Get started by creating your first campaign. The wizard will walk you through connecting to a miraged server, configuring a phishlet, importing targets, and launching.
      </p>
      <AppButton @click="router.push('/campaigns/new')">Create Your First Campaign</AppButton>
    </Card>

    <div v-else-if="stats" class="flex flex-col gap-6">
      <!-- Summary cards -->
      <div class="grid grid-cols-3 gap-4">
        <MetricCard label="Click Rate"
          :value="`${stats.total_emails_sent > 0 ? ((stats.total_clicks / stats.total_emails_sent) * 100).toFixed(1) : '0.0'}%`"
          :subtitle="`${stats.total_clicks} total clicks`" />
        <MetricCard label="Completion Rate"
          :value="`${stats.total_emails_sent > 0 ? ((stats.total_completions / stats.total_emails_sent) * 100).toFixed(1) : '0.0'}%`"
          :subtitle="`${stats.total_completions} total completions`" />
        <MetricCard label="Miraged"
          :value="String(stats.miraged_count)"
          :subtitle="stats.miraged_count === 1 ? 'connection' : 'connections'" />
      </div>

      <!-- Active campaign cards -->
      <div v-if="stats.active_campaigns.length > 0">
        <h6 class="mb-3">Active Campaigns</h6>
        <div class="grid grid-cols-2 gap-4">
          <Card
            v-for="campaign in stats.active_campaigns"
            :key="campaign.id"
            class="p-5 cursor-pointer hover:border-amber/30 transition-colors"
            @click="router.push(`/campaigns/${campaign.id}`)"
          >
            <div class="text-sm font-mono font-medium text-primary mb-4">{{ campaign.name }}</div>

            <!-- Progress bar -->
            <div class="w-full h-1.5 bg-bg rounded-full overflow-hidden mb-3">
              <div class="h-full flex">
                <div
                  class="bg-green"
                  :style="{ width: campaign.total > 0 ? ((campaign.completed / campaign.total) * 100) + '%' : '0%' }"
                ></div>
                <div
                  class="bg-teal"
                  :style="{ width: campaign.total > 0 ? ((campaign.captured / campaign.total) * 100) + '%' : '0%' }"
                ></div>
                <div
                  class="bg-muted"
                  :style="{ width: campaign.total > 0 ? (((campaign.sent - campaign.captured - campaign.completed) / campaign.total) * 100) + '%' : '0%' }"
                ></div>
                <div
                  v-if="campaign.failed > 0"
                  class="bg-danger"
                  :style="{ width: ((campaign.failed / campaign.total) * 100) + '%' }"
                ></div>
              </div>
            </div>

            <div class="flex items-center gap-4 text-xs font-mono">
              <span class="text-muted">{{ campaign.sent + campaign.failed }}/{{ campaign.total }} sent</span>
              <span v-if="campaign.captured > 0" class="text-teal">{{ campaign.captured }} captured</span>
              <span v-if="campaign.completed > 0" class="text-green">{{ campaign.completed }} completed</span>
              <span v-if="campaign.failed > 0" class="text-danger">{{ campaign.failed }} failed</span>
            </div>
          </Card>
        </div>
      </div>

      <!-- Empty state for no active campaigns -->
      <Card v-else class="p-8 text-center">
        <div class="text-sm text-dim font-mono">No active campaigns</div>
      </Card>

      <!-- Recent captures -->
      <div v-if="stats.recent_captures.length > 0">
        <h6 class="mb-3">Recent Captures</h6>
        <Card class="overflow-hidden">
          <table class="w-full text-sm font-mono">
            <thead>
              <tr class="border-b border-edge text-dim text-left uppercase tracking-wider">
                <th class="px-4 py-2.5 font-medium">Email</th>
                <th class="px-4 py-2.5 font-medium">Campaign</th>
                <th class="px-4 py-2.5 font-medium">Captured</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(capture, i) in stats.recent_captures"
                :key="i"
                class="border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors cursor-pointer"
                @click="router.push(`/campaigns/${capture.campaign_id}?session=${capture.session_id}`)"
              >
                <td class="px-4 py-2.5 text-primary">{{ capture.email }}</td>
                <td class="px-4 py-2.5 text-muted">{{ capture.campaign_name }}</td>
                <td class="px-4 py-2.5 text-dim">{{ new Date(capture.captured_at).toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </Card>
      </div>
    </div>
  </div>
</template>
