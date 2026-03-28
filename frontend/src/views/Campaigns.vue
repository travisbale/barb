<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  listCampaigns, createCampaign, deleteCampaign,
  listTargetLists, listTemplates, listSMTPProfiles,
  listMiraged, listMiragedPhishlets,
  type Campaign, type TargetList, type EmailTemplate, type SMTPProfile,
  type MiragedConnection, type MiragedPhishlet,
} from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const router = useRouter()
const campaigns = ref<Campaign[]>([])
const showCreate = ref(false)
const error = ref('')

// Options for dropdowns.
const targetLists = ref<TargetList[]>([])
const templates = ref<EmailTemplate[]>([])
const smtpProfiles = ref<SMTPProfile[]>([])
const miragedConnections = ref<MiragedConnection[]>([])
const phishlets = ref<MiragedPhishlet[]>([])

const form = ref({ name: '', template_id: '', smtp_profile_id: '', target_list_id: '', miraged_id: '', phishlet: '' })

async function load() {
  try {
    campaigns.value = await listCampaigns() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function openCreate() {
  try {
    const [lists, tmpls, profiles, connections] = await Promise.all([
      listTargetLists(),
      listTemplates(),
      listSMTPProfiles(),
      listMiraged(),
    ])
    targetLists.value = lists ?? []
    templates.value = tmpls ?? []
    smtpProfiles.value = profiles ?? []
    miragedConnections.value = connections ?? []
    showCreate.value = true
  } catch (e: any) {
    error.value = e.message
  }
}

async function onMiragedChange(id: string) {
  form.value.miraged_id = id
  form.value.phishlet = ''
  phishlets.value = []
  if (!id) return
  try {
    phishlets.value = (await listMiragedPhishlets(id)) ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function create() {
  try {
    const campaign = await createCampaign(form.value)
    campaigns.value.unshift(campaign)
    form.value = { name: '', template_id: '', smtp_profile_id: '', target_list_id: '', miraged_id: '', phishlet: '' }
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
      <AddButton @click="openCreate">New Campaign</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showCreate" class="p-7 mb-4">
      <form @submit.prevent="create" class="flex flex-col gap-7">
        <AppInput v-model="form.name" placeholder="Campaign name (required)" required />

        <div class="grid grid-cols-3 gap-5">
          <AppSelect v-model="form.target_list_id" label="Target list" required>
            <option value="" disabled></option>
            <option v-for="list in targetLists" :key="list.id" :value="list.id">{{ list.name }}</option>
          </AppSelect>

          <AppSelect v-model="form.template_id" label="Email template" required>
            <option value="" disabled></option>
            <option v-for="tmpl in templates" :key="tmpl.id" :value="tmpl.id">{{ tmpl.name }}</option>
          </AppSelect>

          <AppSelect v-model="form.smtp_profile_id" label="SMTP profile" required>
            <option value="" disabled></option>
            <option v-for="profile in smtpProfiles" :key="profile.id" :value="profile.id">{{ profile.name }}</option>
          </AppSelect>
        </div>

        <div class="grid grid-cols-2 gap-5">
          <AppSelect :modelValue="form.miraged_id" @update:modelValue="onMiragedChange" label="Miraged server">
            <option value="">None (manual lure URL)</option>
            <option v-for="conn in miragedConnections" :key="conn.id" :value="conn.id">{{ conn.name }}</option>
          </AppSelect>

          <AppSelect v-model="form.phishlet" label="Phishlet" :disabled="!form.miraged_id">
            <option value="" disabled></option>
            <option v-for="p in phishlets" :key="p.name" :value="p.name">{{ p.name }}{{ p.enabled ? '' : ' (disabled)' }}</option>
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
