<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import {
  getCampaign, startCampaign, completeCampaign, cancelCampaign, sendTestEmail, listCampaignResults,
  getMiragedSession, exportMiragedSessionCookies, updateCampaign,
  getTemplate, getSMTPProfile, getTargetList, listTargets,
  listTemplates, listSMTPProfiles, listMiraged, enrollMiraged,
  createTemplate, createSMTPProfile, listPhishlets, createPhishlet,
  type Campaign, type CampaignResult, type MiragedSession,
  type EmailTemplate, type SMTPProfile, type TargetList, type MiragedConnection, type Phishlet,
} from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import MiragedForm from '../components/MiragedForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import PageHeader from '../components/PageHeader.vue'
import TabBar from '../components/TabBar.vue'
import TemplateForm from '../components/TemplateForm.vue'
import SMTPForm from '../components/SMTPForm.vue'
import TargetListPicker from '../components/TargetListPicker.vue'
import SettingsSection from '../components/SettingsSection.vue'
import PhishletForm from '../components/PhishletForm.vue'

const route = useRoute()
const { confirm } = useConfirm()
const id = route.params.id as string

const campaign = ref<Campaign | null>(null)
const results = ref<CampaignResult[]>([])
const error = ref('')
const starting = ref(false)
const completing = ref(false)
const cancelling = ref(false)
const showTestEmail = ref(false)
const testEmailAddress = ref('')
const testEmailSending = ref(false)
const testEmailStatus = ref('')

const selectedSession = ref<MiragedSession | null>(null)
const sessionLoading = ref(false)
const sessionError = ref('')

const activeTab = ref('Results')

const settingsTemplate = ref<EmailTemplate | null>(null)
const settingsSmtp = ref<SMTPProfile | null>(null)
const settingsTargetList = ref<TargetList | null>(null)
const settingsTargetCount = ref(0)
const settingsMiraged = ref<MiragedConnection | null>(null)

type SettingsSection = 'template' | 'smtp' | 'targets' | 'miraged' | 'phishlet' | 'general' | null
const expandedSection = ref<SettingsSection>(null)
const settingsSaving = ref(false)
const settingsError = ref('')

const allTemplates = ref<EmailTemplate[]>([])
const allSmtpProfiles = ref<SMTPProfile[]>([])
const allMiraged = ref<MiragedConnection[]>([])
const allPhishlets = ref<Phishlet[]>([])

const editTemplateId = ref('')
const editSmtpId = ref('')
const editTargetListId = ref('')
const targetListPicker = ref<InstanceType<typeof TargetListPicker> | null>(null)
const editMiragedId = ref('')
const editPhishlet = ref('')
const editName = ref('')
const editRedirectUrl = ref('')
const editSendRate = ref('')

const showNewTemplate = ref(false)
const showNewSmtp = ref(false)
const showNewPhishlet = ref(false)
const showNewMiraged = ref(false)
const newPhishletYaml = ref('')
const newTemplate = ref({ name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' })
const newSmtp = ref({ name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' })
const newMiraged = ref({ name: '', address: '', secret_hostname: '', token: '' })
const createLoading = ref(false)

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

async function loadSettings() {
  if (!campaign.value) return
  try {
    const [tmpl, smtp, list, targets, conns] = await Promise.all([
      getTemplate(campaign.value.template_id),
      getSMTPProfile(campaign.value.smtp_profile_id),
      getTargetList(campaign.value.target_list_id),
      listTargets(campaign.value.target_list_id),
      campaign.value.miraged_id ? listMiraged() : Promise.resolve([]),
    ])
    settingsTemplate.value = tmpl
    settingsSmtp.value = smtp
    settingsTargetList.value = list
    settingsTargetCount.value = targets?.length ?? 0
    if (conns.length > 0) allMiraged.value = conns
    settingsMiraged.value = conns.find(c => c.id === campaign.value!.miraged_id) ?? null
  } catch { /* ignore — settings are supplementary */ }
}

function expandSection(section: SettingsSection) {
  if (expandedSection.value === section) {
    expandedSection.value = null
    return
  }
  settingsError.value = ''
  showNewTemplate.value = false
  showNewSmtp.value = false
  showNewPhishlet.value = false
  showNewMiraged.value = false
  expandedSection.value = section

  if (!campaign.value) return

  // Pre-select current values and load options for the section (cached after first load).
  if (section === 'template') {
    editTemplateId.value = campaign.value.template_id
    if (allTemplates.value.length === 0) listTemplates().then(t => allTemplates.value = t)
  } else if (section === 'smtp') {
    editSmtpId.value = campaign.value.smtp_profile_id
    if (allSmtpProfiles.value.length === 0) listSMTPProfiles().then(s => allSmtpProfiles.value = s)
  } else if (section === 'targets') {
    editTargetListId.value = campaign.value.target_list_id
  } else if (section === 'miraged') {
    editMiragedId.value = campaign.value.miraged_id
    if (allMiraged.value.length === 0) listMiraged().then(m => allMiraged.value = m)
  } else if (section === 'phishlet') {
    editPhishlet.value = campaign.value.phishlet
    if (allPhishlets.value.length === 0) listPhishlets().then(p => allPhishlets.value = p)
  } else if (section === 'general') {
    editName.value = campaign.value.name
    editRedirectUrl.value = campaign.value.redirect_url
    editSendRate.value = String(campaign.value.send_rate)
  }
}

async function saveSection(section: SettingsSection) {
  if (!campaign.value || !section) return
  settingsSaving.value = true
  settingsError.value = ''
  try {
    const updates: Record<string, unknown> = {}
    if (section === 'template' && editTemplateId.value !== campaign.value.template_id) {
      updates.template_id = editTemplateId.value
    } else if (section === 'smtp' && editSmtpId.value !== campaign.value.smtp_profile_id) {
      updates.smtp_profile_id = editSmtpId.value
    } else if (section === 'targets' && editTargetListId.value !== campaign.value.target_list_id) {
      updates.target_list_id = editTargetListId.value
    } else if (section === 'miraged' && editMiragedId.value !== campaign.value.miraged_id) {
      updates.miraged_id = editMiragedId.value
    } else if (section === 'phishlet' && editPhishlet.value !== campaign.value.phishlet) {
      updates.phishlet = editPhishlet.value
    } else if (section === 'general') {
      if (editName.value !== campaign.value.name) updates.name = editName.value
      if (editRedirectUrl.value !== campaign.value.redirect_url) updates.redirect_url = editRedirectUrl.value
      if (Number(editSendRate.value) !== campaign.value.send_rate) updates.send_rate = Number(editSendRate.value)
    }
    if (Object.keys(updates).length === 0) { expandedSection.value = null; return }
    campaign.value = await updateCampaign(id, updates)
    expandedSection.value = null

    // Update summaries from cached dropdown data instead of re-fetching everything.
    if (section === 'template') {
      settingsTemplate.value = allTemplates.value.find(t => t.id === campaign.value!.template_id) ?? settingsTemplate.value
    } else if (section === 'smtp') {
      settingsSmtp.value = allSmtpProfiles.value.find(s => s.id === campaign.value!.smtp_profile_id) ?? settingsSmtp.value
    } else if (section === 'targets') {
      await loadSettings()
    } else if (section === 'miraged') {
      settingsMiraged.value = allMiraged.value.find(m => m.id === campaign.value!.miraged_id) ?? settingsMiraged.value
    }
  } catch (e: any) {
    settingsError.value = e.message
  } finally {
    settingsSaving.value = false
  }
}

async function createNewPhishlet() {
  createLoading.value = true
  settingsError.value = ''
  try {
    const phishlet = await createPhishlet(newPhishletYaml.value)
    allPhishlets.value.unshift(phishlet)
    editPhishlet.value = phishlet.name
    showNewPhishlet.value = false
    newPhishletYaml.value = ''
  } catch (e: any) {
    settingsError.value = e.message
  } finally {
    createLoading.value = false
  }
}

async function createNewTemplate() {
  createLoading.value = true
  settingsError.value = ''
  try {
    const tmpl = await createTemplate(newTemplate.value)
    allTemplates.value.unshift(tmpl)
    editTemplateId.value = tmpl.id
    showNewTemplate.value = false
    newTemplate.value = { name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' }
  } catch (e: any) {
    settingsError.value = e.message
  } finally {
    createLoading.value = false
  }
}

async function createNewMiraged() {
  createLoading.value = true
  settingsError.value = ''
  try {
    const conn = await enrollMiraged(newMiraged.value)
    allMiraged.value.unshift(conn)
    editMiragedId.value = conn.id
    showNewMiraged.value = false
    newMiraged.value = { name: '', address: '', secret_hostname: '', token: '' }
  } catch (e: any) {
    settingsError.value = e.message
  } finally {
    createLoading.value = false
  }
}

async function createNewSmtp() {
  createLoading.value = true
  settingsError.value = ''
  try {
    const profile = await createSMTPProfile({
      ...newSmtp.value,
      port: parseInt(newSmtp.value.port) || 587,
    })
    allSmtpProfiles.value.unshift(profile)
    editSmtpId.value = profile.id
    showNewSmtp.value = false
    newSmtp.value = { name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' }
  } catch (e: any) {
    settingsError.value = e.message
  } finally {
    createLoading.value = false
  }
}

async function start() {
  if (!await confirm('Start this campaign? Emails will begin sending immediately.', { label: 'Start', variant: 'primary' })) return
  starting.value = true
  error.value = ''
  try {
    await startCampaign(id)
    await load()
    startStreaming()
  } catch (e: any) {
    error.value = e.message
  } finally {
    starting.value = false
  }
}

async function complete() {
  if (!await confirm('Complete this campaign? The lure and phishlet will be disabled.', { label: 'Complete', variant: 'primary' })) return
  completing.value = true
  error.value = ''
  try {
    await completeCampaign(id)
    await load()
    stopStreaming()
  } catch (e: any) {
    error.value = e.message
  } finally {
    completing.value = false
  }
}

async function cancel() {
  if (!await confirm('Cancel this campaign? The lure and phishlet will be disabled.', { label: 'Cancel', variant: 'danger' })) return
  cancelling.value = true
  error.value = ''
  try {
    await cancelCampaign(id)
    await load()
    stopStreaming()
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
    await load()
  } catch (e: any) {
    testEmailStatus.value = e.message
  } finally {
    testEmailSending.value = false
  }
}

let eventSource: EventSource | null = null

function startStreaming() {
  stopStreaming()
  eventSource = new EventSource(`/api/campaigns/${id}/stream`)

  eventSource.addEventListener('result.updated', (e) => {
    const event = JSON.parse(e.data)
    const result = event.result as CampaignResult
    const idx = results.value.findIndex(r => r.id === result.id)
    if (idx !== -1) {
      results.value[idx] = result
    }
  })

  eventSource.addEventListener('campaign.status', (e) => {
    const event = JSON.parse(e.data)
    if (campaign.value) {
      campaign.value.status = event.status
    }
    // Stop streaming when the campaign is no longer active.
    if (event.status !== 'active') {
      stopStreaming()
    }
  })

  eventSource.onerror = () => {
    // EventSource.CLOSED means the server rejected the connection (e.g., 404).
    // The browser will not reconnect in this case.
    if (eventSource?.readyState === EventSource.CLOSED) {
      stopStreaming()
    }
  }
}

function stopStreaming() {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
}

const statusColor: Record<string, string> = {
  pending: 'text-dim',
  sent: 'text-muted',
  failed: 'text-danger',
  clicked: 'text-amber',
  captured: 'text-amber',
  completed: 'text-teal',
}

const sentCount = computed(() => results.value.filter(result => result.status !== 'pending').length)
const totalCount = computed(() => results.value.length)
const clickedCount = computed(() => results.value.filter(result => ['clicked', 'captured', 'completed'].includes(result.status)).length)
const capturedCount = computed(() => results.value.filter(result => result.status === 'captured' || result.status === 'completed').length)
const completedCount = computed(() => results.value.filter(result => result.status === 'completed').length)
const captureRate = computed(() => {
  const delivered = sentCount.value
  return delivered > 0 ? completedCount.value / delivered : 0
})

async function viewSession(result: CampaignResult) {
  if (!result.session_id || !campaign.value?.miraged_id) return
  stopStreaming()
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
  if (isActive.value) startStreaming()
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
  if (isActive.value) startStreaming()
})

onUnmounted(stopStreaming)
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
            <span v-if="clickedCount > 0" class="text-xs text-amber font-mono">{{ clickedCount }} clicked</span>
            <span v-if="capturedCount > 0" class="text-xs text-amber font-mono">{{ capturedCount }} captured</span>
            <span v-if="completedCount > 0" class="text-xs text-teal font-mono">{{ completedCount }} completed</span>
            <span v-if="captureRate > 0" class="text-xs text-dim font-mono">({{ (captureRate * 100).toFixed(1) }}%)</span>
          </div>
          <span v-if="campaign?.lure_url" class="text-xs text-dim font-mono select-all">{{ campaign.lure_url }}</span>
        </div>
      </template>
      <AppButton v-if="results.length > 0" variant="secondary" @click="exportCSV">Export CSV</AppButton>
      <AppButton v-if="isDraft" variant="secondary" @click="showTestEmail = !showTestEmail">Send Test</AppButton>
      <AppButton v-if="isDraft" :disabled="starting" @click="start">
        {{ starting ? 'Starting...' : 'Start Campaign' }}
      </AppButton>
      <AppButton v-if="isActive" :disabled="completing" @click="complete">
        {{ completing ? 'Completing...' : 'Complete Campaign' }}
      </AppButton>
      <AppButton v-if="isActive" variant="danger" :disabled="cancelling" @click="cancel">
        {{ cancelling ? 'Cancelling...' : 'Cancel Campaign' }}
      </AppButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <!-- Test email -->
    <Card v-if="showTestEmail" class="p-5 mb-4">
      <form @submit.prevent="sendTest" class="flex flex-col gap-7">
        <div class="text-xs font-mono text-dim uppercase tracking-wider">Send Test Email</div>
        <AppInput v-model="testEmailAddress" type="email" placeholder="Recipient email" required />
        <div class="flex gap-2 justify-end">
          <AppButton variant="ghost" @click="showTestEmail = false">Cancel</AppButton>
          <AppButton type="submit" :disabled="testEmailSending || !testEmailAddress">
            {{ testEmailSending ? 'Sending...' : 'Send' }}
          </AppButton>
        </div>
      </form>
      <div v-if="testEmailStatus" class="text-xs font-mono mt-2" :class="testEmailStatus.startsWith('Test email') ? 'text-teal' : 'text-danger'">
        {{ testEmailStatus }}
      </div>
    </Card>

    <TabBar :tabs="['Results', 'Settings']" :modelValue="activeTab" @update:modelValue="(t: string) => { activeTab = t as any; if (t === 'Settings') loadSettings() }" />

    <!-- Settings tab -->
    <div v-if="activeTab === 'Settings'" class="flex flex-col gap-4">
      <ErrorBanner :message="settingsError" />

      <!-- General: Name, Send Rate -->
      <SettingsSection label="General" :editable="isDraft" :expanded="expandedSection === 'general'" :saving="settingsSaving"
        @change="expandSection('general')" @cancel="expandedSection = null" @save="saveSection('general')">
        <template #editor>
          <div class="flex flex-col gap-7">
            <AppInput v-model="editName" placeholder="Name" required />
            <AppInput v-model="editRedirectUrl" placeholder="Redirect URL (post-capture destination)" />
            <AppInput v-model="editSendRate" placeholder="Send Rate (per minute)" type="number" min="1" required />
          </div>
        </template>
        <template #summary>
          <div class="text-primary">{{ campaign?.name }}</div>
          <div class="text-xs text-dim">{{ campaign?.send_rate }} emails/min</div>
          <div v-if="campaign?.redirect_url" class="text-xs text-dim">Redirect: {{ campaign.redirect_url }}</div>
        </template>
      </SettingsSection>

      <!-- Template -->
      <SettingsSection label="Email Template" :editable="isDraft" :expanded="expandedSection === 'template'" :saving="settingsSaving"
        @change="expandSection('template')" @cancel="expandedSection = null" @save="saveSection('template')">
        <template #editor>
          <template v-if="!showNewTemplate">
            <AppSelect v-model="editTemplateId" label="Select a template">
              <option value="" disabled></option>
              <option v-for="t in allTemplates" :key="t.id" :value="t.id">{{ t.name }} — {{ t.subject }}</option>
            </AppSelect>
          </template>
          <div v-else class="flex flex-col gap-7">
            <TemplateForm v-model="newTemplate" min-editor-height="150px" />
            <div class="flex gap-2 justify-end">
              <AppButton variant="ghost" @click="showNewTemplate = false">Cancel</AppButton>
              <AppButton :disabled="createLoading" @click="createNewTemplate">{{ createLoading ? 'Creating...' : 'Create' }}</AppButton>
            </div>
          </div>
        </template>
        <template #create-new>
          <button v-if="!showNewTemplate" @click="showNewTemplate = true" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">+ Create new template</button>
        </template>
        <template #summary>
          <div class="text-primary">{{ settingsTemplate?.name || campaign?.template_id }}</div>
          <div v-if="settingsTemplate" class="text-xs text-dim mt-0.5">Subject: {{ settingsTemplate.subject }}</div>
        </template>
      </SettingsSection>

      <!-- SMTP Profile -->
      <SettingsSection label="SMTP Profile" :editable="isDraft" :expanded="expandedSection === 'smtp'" :saving="settingsSaving"
        @change="expandSection('smtp')" @cancel="expandedSection = null" @save="saveSection('smtp')">
        <template #editor>
          <template v-if="!showNewSmtp">
            <AppSelect v-model="editSmtpId" label="Select an SMTP profile">
              <option value="" disabled></option>
              <option v-for="s in allSmtpProfiles" :key="s.id" :value="s.id">{{ s.name }} ({{ s.host }}:{{ s.port }})</option>
            </AppSelect>
          </template>
          <div v-else class="flex flex-col gap-7">
            <SMTPForm v-model="newSmtp" />
            <div class="flex gap-2 justify-end">
              <AppButton variant="ghost" @click="showNewSmtp = false">Cancel</AppButton>
              <AppButton :disabled="createLoading" @click="createNewSmtp">{{ createLoading ? 'Creating...' : 'Create' }}</AppButton>
            </div>
          </div>
        </template>
        <template #create-new>
          <button v-if="!showNewSmtp" @click="showNewSmtp = true" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">+ Create new profile</button>
        </template>
        <template #summary>
          <div class="text-primary">{{ settingsSmtp?.name || campaign?.smtp_profile_id }}</div>
          <div v-if="settingsSmtp" class="text-xs text-dim mt-0.5">{{ settingsSmtp.host }}:{{ settingsSmtp.port }} &middot; {{ settingsSmtp.from_addr }}</div>
        </template>
      </SettingsSection>

      <!-- Target List -->
      <SettingsSection label="Target List" :editable="isDraft" :expanded="expandedSection === 'targets'" :saving="settingsSaving"
        @change="expandSection('targets')" @cancel="expandedSection = null" @save="saveSection('targets')">
        <template #editor>
          <TargetListPicker ref="targetListPicker" v-model="editTargetListId" />
        </template>
        <template #create-new>
          <button @click="targetListPicker?.startCreateNew()" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">+ Create new list</button>
        </template>
        <template #summary>
          <div class="text-primary">{{ settingsTargetList?.name || campaign?.target_list_id }}</div>
          <div v-if="settingsTargetCount" class="text-xs text-dim mt-0.5">{{ settingsTargetCount }} {{ settingsTargetCount === 1 ? 'target' : 'targets' }}</div>
        </template>
      </SettingsSection>

      <!-- Phishlet -->
      <SettingsSection v-if="campaign?.phishlet" label="Phishlet" :editable="isDraft" :expanded="expandedSection === 'phishlet'" :saving="settingsSaving"
        @change="expandSection('phishlet')" @cancel="expandedSection = null" @save="saveSection('phishlet')">
        <template #editor>
          <template v-if="!showNewPhishlet">
            <AppSelect v-model="editPhishlet" label="Select a phishlet">
              <option value="" disabled></option>
              <option v-for="p in allPhishlets" :key="p.id" :value="p.name">{{ p.name }}</option>
            </AppSelect>
          </template>
          <PhishletForm v-else v-model="newPhishletYaml" :loading="createLoading"
            @submit="createNewPhishlet" @cancel="showNewPhishlet = false" />
        </template>
        <template #create-new>
          <button v-if="!showNewPhishlet" @click="showNewPhishlet = true" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">+ Create new phishlet</button>
        </template>
        <template #summary>
          <div class="text-primary">{{ campaign.phishlet }}</div>
        </template>
      </SettingsSection>

      <!-- Miraged Connection -->
      <SettingsSection v-if="campaign?.miraged_id" label="Miraged Connection" :editable="isDraft" :expanded="expandedSection === 'miraged'" :saving="settingsSaving"
        @change="expandSection('miraged')" @cancel="expandedSection = null" @save="saveSection('miraged')">
        <template #editor>
          <template v-if="!showNewMiraged">
            <AppSelect v-model="editMiragedId" label="Select a connection">
              <option value="" disabled></option>
              <option v-for="m in allMiraged" :key="m.id" :value="m.id">{{ m.name }} ({{ m.address }})</option>
            </AppSelect>
          </template>
          <div v-else class="flex flex-col gap-7">
            <MiragedForm v-model="newMiraged" />
            <div class="flex gap-2 justify-end">
              <AppButton variant="ghost" @click="showNewMiraged = false">Cancel</AppButton>
              <AppButton :disabled="createLoading" @click="createNewMiraged">{{ createLoading ? 'Enrolling...' : 'Enroll' }}</AppButton>
            </div>
          </div>
        </template>
        <template #create-new>
          <button v-if="!showNewMiraged" @click="showNewMiraged = true" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">+ Enroll new server</button>
        </template>
        <template #summary>
          <div class="text-primary">{{ settingsMiraged?.name || campaign?.miraged_id }}</div>
          <div v-if="settingsMiraged" class="text-xs text-dim mt-0.5">{{ settingsMiraged.address }}</div>
        </template>
      </SettingsSection>
    </div>

    <!-- Results tab -->
    <template v-if="activeTab === 'Results'">

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
          <div class="grid grid-cols-2 gap-3 [&>*:only-child]:col-span-2">
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
          <div class="grid grid-cols-2 gap-3 [&>*:only-child]:col-span-2">
            <div v-for="(value, key) in selectedSession.custom" :key="key" class="px-3 py-2 bg-bg border border-edge">
              <div class="text-xs text-dim font-mono">{{ key }}</div>
              <div class="text-sm text-primary font-mono select-all">{{ value }}</div>
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
            <div class="grid grid-cols-2 gap-3 [&>*:only-child]:col-span-2">
              <div v-for="(value, name) in cookies" :key="name" class="px-3 py-2 bg-bg border border-edge">
                <div class="text-xs text-dim font-mono">{{ name }}</div>
                <div class="text-sm text-primary font-mono select-all break-all">{{ value }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Details -->
        <div>
          <div class="text-xs font-mono text-dim uppercase tracking-wider mb-2">Details</div>
          <div class="grid grid-cols-2 gap-3 [&>*:only-child]:col-span-2">
            <div v-if="selectedSession.remote_addr" class="px-3 py-2 bg-bg border border-edge">
              <div class="text-xs text-dim font-mono">IP Address</div>
              <div class="text-sm text-primary font-mono select-all">{{ selectedSession.remote_addr }}</div>
            </div>
            <div v-if="selectedSession.phishlet" class="px-3 py-2 bg-bg border border-edge">
              <div class="text-xs text-dim font-mono">Phishlet</div>
              <div class="text-sm text-primary font-mono">{{ selectedSession.phishlet }}</div>
            </div>
            <div v-if="selectedSession.started_at" class="px-3 py-2 bg-bg border border-edge">
              <div class="text-xs text-dim font-mono">Started</div>
              <div class="text-sm text-primary font-mono">{{ new Date(selectedSession.started_at).toLocaleString() }}</div>
            </div>
            <div v-if="selectedSession.user_agent" class="px-3 py-2 bg-bg border border-edge col-span-2">
              <div class="text-xs text-dim font-mono">User Agent</div>
              <div class="text-sm text-primary font-mono select-all break-all">{{ selectedSession.user_agent }}</div>
            </div>
          </div>
        </div>
      </div>
    </Card>

    <EmptyState v-if="results.length === 0" message="No results yet." />

    <DataTable v-else :columns="[{ label: 'Email' }, { label: 'Status' }, { label: 'Sent' }, { label: 'Clicked' }, { label: 'Captured' }]">
      <DataTableRow
        v-for="(result, i) in results"
        :key="result.id"
        :index="i"
        :clickable="!!result.session_id"
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
      </DataTableRow>
    </DataTable>

    </template>
  </div>
</template>
