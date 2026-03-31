<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboard, type DashboardStats } from '../api/client'
import AppButton from '../components/AppButton.vue'
import Card from '../components/Card.vue'
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
      <h1 class="font-mono text-xl font-bold text-primary tracking-wide">Dashboard</h1>
      <p class="text-sm text-dim font-mono mt-1">Operations overview</p>
    </div>

    <ErrorBanner :message="error" />

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
      <div class="grid grid-cols-4 gap-4">
        <Card class="p-5">
          <div class="text-xs font-mono text-dim uppercase tracking-wider">Campaigns</div>
          <div class="text-2xl font-mono font-bold text-primary mt-1">{{ stats.campaigns.total }}</div>
          <div class="text-xs font-mono text-dim mt-2 flex gap-3">
            <template v-if="stats.campaigns.active || stats.campaigns.draft || stats.campaigns.completed">
              <span v-if="stats.campaigns.active" class="text-teal">{{ stats.campaigns.active }} active</span>
              <span v-if="stats.campaigns.draft">{{ stats.campaigns.draft }} draft</span>
              <span v-if="stats.campaigns.completed">{{ stats.campaigns.completed }} done</span>
            </template>
            <span v-else>total campaigns</span>
          </div>
        </Card>

        <Card class="p-5">
          <div class="text-xs font-mono text-dim uppercase tracking-wider">Active Now</div>
          <div class="text-2xl font-mono font-bold" :class="stats.campaigns.active > 0 ? 'text-teal' : 'text-dim'">{{ stats.campaigns.active }}</div>
          <div class="text-xs font-mono text-dim mt-2">
            {{ stats.campaigns.active === 1 ? 'campaign running' : 'campaigns running' }}
          </div>
        </Card>

        <Card class="p-5">
          <div class="text-xs font-mono text-dim uppercase tracking-wider">Total Captures</div>
          <div class="text-2xl font-mono font-bold text-teal mt-1">{{ stats.total_captures }}</div>
          <div class="text-xs font-mono text-dim mt-2">lifetime sessions</div>
        </Card>

        <Card class="p-5">
          <div class="text-xs font-mono text-dim uppercase tracking-wider">Miraged</div>
          <div class="text-2xl font-mono font-bold text-primary mt-1">{{ stats.miraged_count }}</div>
          <div class="text-xs font-mono text-dim mt-2">
            {{ stats.miraged_count === 1 ? 'connection' : 'connections' }}
          </div>
        </Card>
      </div>

      <!-- Active campaign cards -->
      <div v-if="stats.active_campaigns.length > 0">
        <h2 class="text-xs font-mono text-dim uppercase tracking-wider mb-3">Active Campaigns</h2>
        <div class="grid grid-cols-2 gap-4">
          <Card
            v-for="campaign in stats.active_campaigns"
            :key="campaign.id"
            class="p-5 cursor-pointer hover:border-amber/30 transition-colors"
            @click="router.push(`/campaigns/${campaign.id}`)"
          >
            <div class="flex items-center justify-between mb-4">
              <div class="text-sm font-mono font-medium text-primary">{{ campaign.name }}</div>
              <span class="text-xs font-mono uppercase tracking-wider text-teal">active</span>
            </div>

            <!-- Progress bar -->
            <div class="w-full h-1.5 bg-bg rounded-full overflow-hidden mb-3">
              <div class="h-full flex">
                <div
                  class="bg-teal"
                  :style="{ width: campaign.total > 0 ? ((campaign.captured / campaign.total) * 100) + '%' : '0%' }"
                ></div>
                <div
                  class="bg-muted"
                  :style="{ width: campaign.total > 0 ? ((campaign.sent / campaign.total) * 100) + '%' : '0%' }"
                ></div>
                <div
                  v-if="campaign.failed > 0"
                  class="bg-danger"
                  :style="{ width: ((campaign.failed / campaign.total) * 100) + '%' }"
                ></div>
              </div>
            </div>

            <div class="flex items-center gap-4 text-xs font-mono">
              <span class="text-muted">{{ campaign.sent + campaign.captured + campaign.failed }}/{{ campaign.total }} sent</span>
              <span v-if="campaign.captured > 0" class="text-teal">{{ campaign.captured }} captured</span>
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
        <h2 class="text-xs font-mono text-dim uppercase tracking-wider mb-3">Recent Captures</h2>
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
                class="border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
              >
                <td class="px-4 py-2.5 text-teal">{{ capture.email }}</td>
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
