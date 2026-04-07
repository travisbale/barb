<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { countResults, resultsToCSV, resultStatusColor } from '../utils/results'
import {
  getCampaign, startCampaign, completeCampaign, cancelCampaign, sendTestEmail, listCampaignResults,
  updateCampaign,
  getTemplate, getSMTPProfile, getTargetList, listTargets,
  listTemplates, listSMTPProfiles, listMiraged, enrollMiraged,
  createTemplate, createSMTPProfile, listPhishlets, createPhishlet,
  type Campaign, type CampaignResult,
  type EmailTemplate, type SMTPProfile, type TargetList, type MiragedConnection, type Phishlet,
} from '../api/client'
import AddButton from '../components/AddButton.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import MiragedForm from '../components/MiragedForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import MetricCard from '../components/MetricCard.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import PageHeader from '../components/PageHeader.vue'
import TabBar from '../components/TabBar.vue'
import TemplateForm from '../components/TemplateForm.vue'
import SMTPForm from '../components/SMTPForm.vue'
import TargetListPicker from '../components/TargetListPicker.vue'
import SettingsSection from '../components/SettingsSection.vue'
import PhishletForm from '../components/PhishletForm.vue'
import SessionPanel from '../components/SessionPanel.vue'

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

const sessionPanel = ref<InstanceType<typeof SessionPanel> | null>(null)

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
const settingsLoaded = ref(false)

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
  if (!campaign.value || settingsLoaded.value) return
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
    settingsLoaded.value = true
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

    if (section === 'template') {
      settingsTemplate.value = allTemplates.value.find(t => t.id === campaign.value!.template_id) ?? settingsTemplate.value
    } else if (section === 'smtp') {
      settingsSmtp.value = allSmtpProfiles.value.find(s => s.id === campaign.value!.smtp_profile_id) ?? settingsSmtp.value
    } else if (section === 'targets') {
      settingsLoaded.value = false
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

async function createInline(apiCall: () => Promise<any>, onSuccess: (item: any) => void) {
  createLoading.value = true
  settingsError.value = ''
  try {
    onSuccess(await apiCall())
  } catch (e: any) {
    settingsError.value = e.message
  } finally {
    createLoading.value = false
  }
}

const createNewPhishlet = () => createInline(
  () => createPhishlet(newPhishletYaml.value),
  (p) => { allPhishlets.value.unshift(p); editPhishlet.value = p.name; showNewPhishlet.value = false; newPhishletYaml.value = '' },
)

const createNewTemplate = () => createInline(
  () => createTemplate(newTemplate.value),
  (t) => { allTemplates.value.unshift(t); editTemplateId.value = t.id; showNewTemplate.value = false; newTemplate.value = { name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' } },
)

const createNewMiraged = () => createInline(
  () => enrollMiraged(newMiraged.value),
  (c) => { allMiraged.value.unshift(c); editMiragedId.value = c.id; showNewMiraged.value = false; newMiraged.value = { name: '', address: '', secret_hostname: '', token: '' } },
)

const createNewSmtp = () => createInline(
  () => createSMTPProfile({ ...newSmtp.value, port: parseInt(newSmtp.value.port) || 587 }),
  (s) => { allSmtpProfiles.value.unshift(s); editSmtpId.value = s.id; showNewSmtp.value = false; newSmtp.value = { name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' } },
)

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
  if (!await confirm('Cancel this campaign? The lure and phishlet will be disabled.', { label: 'Confirm', variant: 'danger' })) return
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


const counts = computed(() => countResults(results.value))
const completionRate = computed(() => counts.value.sent > 0 ? counts.value.completed / counts.value.sent : 0)

function openResult(result: CampaignResult) {
  if (!result.session_id) return
  stopStreaming()
  sessionPanel.value?.open(result.session_id, result)
}

function onSessionClose() {
  if (isActive.value) startStreaming()
}

function exportCSV() {
  if (results.value.length === 0) return

  const csv = resultsToCSV(results.value)
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

  const sessionId = route.query.session as string
  if (sessionId) {
    const result = results.value.find(r => r.session_id === sessionId)
    if (result) openResult(result)
  }
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
        <div class="flex items-center gap-3 mt-1">
          <span class="text-xs font-mono uppercase tracking-wider" :class="{
            'text-dim': campaign?.status === 'draft',
            'text-teal': campaign?.status === 'active',
            'text-amber': campaign?.status === 'paused',
            'text-muted': campaign?.status === 'completed',
            'text-danger': campaign?.status === 'cancelled',
          }">{{ campaign?.status }}</span>
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
      <AppButton v-if="isActive" :disabled="completing" @click="complete">
        {{ completing ? 'Completing...' : 'Complete Campaign' }}
      </AppButton>
    </PageHeader>

    <ErrorBanner v-model="error" />

    <!-- Test email -->
    <Card v-if="showTestEmail" class="p-5 mb-4">
      <form @submit.prevent="sendTest" class="flex flex-col gap-7">
        <h6>Send Test Email</h6>
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

    <!-- Campaign stats -->
    <div v-if="campaign && campaign.status !== 'draft'" class="grid grid-cols-3 gap-4 mb-6">
      <MetricCard label="Click Rate"
        :value="`${counts.sent > 0 ? ((counts.clicked / counts.sent) * 100).toFixed(1) : '0.0'}%`"
        :subtitle="`${counts.clicked} of ${counts.sent} clicked`" />
      <MetricCard label="Capture Rate"
        :value="`${counts.sent > 0 ? ((counts.captured / counts.sent) * 100).toFixed(1) : '0.0'}%`"
        :subtitle="`${counts.captured} of ${counts.sent} captured`" />
      <MetricCard label="Completion Rate"
        :value="`${(completionRate * 100).toFixed(1)}%`"
        :subtitle="`${counts.completed} of ${counts.sent} completed`" />
    </div>

    <TabBar :tabs="['Results', 'Settings']" :modelValue="activeTab" @update:modelValue="(t: string) => { activeTab = t as any; if (t === 'Settings') loadSettings() }" />

    <!-- Settings tab -->
    <div v-if="activeTab === 'Settings'" class="flex flex-col gap-4">
      <ErrorBanner v-model="settingsError" />

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
          <AddButton v-if="!showNewTemplate" variant="link" @click="showNewTemplate = true">Create new template</AddButton>
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
          <AddButton v-if="!showNewSmtp" variant="link" @click="showNewSmtp = true">Create new profile</AddButton>
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
          <AddButton variant="link" @click="targetListPicker?.startCreateNew()">Create new list</AddButton>
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
          <AddButton v-if="!showNewPhishlet" variant="link" @click="showNewPhishlet = true">Create new phishlet</AddButton>
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
          <AddButton v-if="!showNewMiraged" variant="link" @click="showNewMiraged = true">Enroll new server</AddButton>
        </template>
        <template #summary>
          <div class="text-primary">{{ settingsMiraged?.name || campaign?.miraged_id }}</div>
          <div v-if="settingsMiraged" class="text-xs text-dim mt-0.5">{{ settingsMiraged.address }}</div>
        </template>
      </SettingsSection>
    </div>

    <!-- Results tab -->
    <template v-if="activeTab === 'Results'">

    <SessionPanel v-if="campaign?.miraged_id" ref="sessionPanel" :miraged-id="campaign.miraged_id" @close="onSessionClose" />

    <EmptyState v-if="results.length === 0" message="No results yet." />

    <DataTable v-else :columns="[{ label: 'Email' }, { label: 'Status' }, { label: 'Sent' }, { label: 'Clicked' }, { label: 'Captured' }]">
      <DataTableRow
        v-for="(result, i) in results"
        :key="result.id"
        :index="i"
        :clickable="!!result.session_id"
        @click="openResult(result)"
      >
        <td class="text-primary">{{ result.email }}</td>
        <td>
          <span :class="resultStatusColor(result.status)" class="uppercase text-xs tracking-wider">
            {{ result.status }}
          </span>
        </td>
        <td class="text-dim">{{ result.sent_at ? new Date(result.sent_at).toLocaleString() : '—' }}</td>
        <td class="text-dim">{{ result.clicked_at ? new Date(result.clicked_at).toLocaleString() : '—' }}</td>
        <td class="text-dim">{{ result.captured_at ? new Date(result.captured_at).toLocaleString() : '—' }}</td>
      </DataTableRow>
    </DataTable>

    </template>
  </div>
</template>
