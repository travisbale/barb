<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  listCampaigns, createCampaign, deleteCampaign,
  listTargetLists, listTemplates, listSMTPProfiles,
  type Campaign, type TargetList, type EmailTemplate, type SMTPProfile,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'

const router = useRouter()
const campaigns = ref<Campaign[]>([])
const showCreate = ref(false)
const error = ref('')

// Options for dropdowns.
const targetLists = ref<TargetList[]>([])
const templates = ref<EmailTemplate[]>([])
const smtpProfiles = ref<SMTPProfile[]>([])

const form = ref({ name: '', template_id: '', smtp_profile_id: '', target_list_id: '' })

async function load() {
  try {
    campaigns.value = await listCampaigns() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function openCreate() {
  try {
    const [lists, tmpls, profiles] = await Promise.all([
      listTargetLists(),
      listTemplates(),
      listSMTPProfiles(),
    ])
    targetLists.value = lists ?? []
    templates.value = tmpls ?? []
    smtpProfiles.value = profiles ?? []
    showCreate.value = true
  } catch (e: any) {
    error.value = e.message
  }
}

async function create() {
  try {
    const campaign = await createCampaign(form.value)
    campaigns.value.unshift(campaign)
    form.value = { name: '', template_id: '', smtp_profile_id: '', target_list_id: '' }
    showCreate.value = false
    router.push(`/campaigns/${campaign.id}`)
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(id: string) {
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
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Campaigns" :subtitle="`${campaigns.length} campaigns`">
      <AppButton @click="openCreate">+ New Campaign</AppButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showCreate" class="p-5 mb-4">
      <form @submit.prevent="create" class="flex flex-col gap-4">
        <AppInput v-model="form.name" placeholder="Campaign name (required)" required />

        <div class="grid grid-cols-3 gap-3">
          <AppSelect v-model="form.target_list_id" required>
            <option value="" disabled>Target list...</option>
            <option v-for="l in targetLists" :key="l.id" :value="l.id">{{ l.name }}</option>
          </AppSelect>

          <AppSelect v-model="form.template_id" required>
            <option value="" disabled>Email template...</option>
            <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
          </AppSelect>

          <AppSelect v-model="form.smtp_profile_id" required>
            <option value="" disabled>SMTP profile...</option>
            <option v-for="p in smtpProfiles" :key="p.id" :value="p.id">{{ p.name }}</option>
          </AppSelect>
        </div>

        <div class="flex gap-2 pt-1">
          <AppButton type="submit">Create</AppButton>
          <AppButton variant="ghost" @click="showCreate = false">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="campaigns.length === 0 && !showCreate" message="No campaigns. Create one to begin an operation." />

    <Card v-else-if="campaigns.length > 0">
      <div
        v-for="(c, i) in campaigns"
        :key="c.id"
        :style="{ animationDelay: `${i * 30}ms` }"
        class="animate-in flex items-center justify-between px-4 py-3 border-b border-edge last:border-0 hover:bg-surface-hover cursor-pointer transition-colors"
        @click="router.push(`/campaigns/${c.id}`)"
      >
        <div class="flex items-center gap-3">
          <span class="text-xs font-mono uppercase tracking-wider" :class="statusColor[c.status] ?? 'text-dim'">
            {{ c.status }}
          </span>
          <div>
            <div class="text-sm font-medium text-primary">{{ c.name }}</div>
            <div class="text-xs text-dim font-mono mt-0.5">{{ new Date(c.created_at).toLocaleDateString() }}</div>
          </div>
        </div>
        <AppButton variant="danger" @click.stop="remove(c.id)">Del</AppButton>
      </div>
    </Card>
  </div>
</template>
