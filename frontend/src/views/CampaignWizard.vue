<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  listMiraged, enrollMiraged, listMiragedDNSProviders, pushMiragedPhishlet, enableMiragedPhishlet,
  listPhishlets, createPhishlet,
  listTargetLists,
  listTemplates, createTemplate, previewTemplate,
  listSMTPProfiles, createSMTPProfile,
  createCampaign,
  type MiragedConnection, type Phishlet,
  type TargetList, type EmailTemplate, type SMTPProfile,
  type PreviewResult,
} from '../api/client'
import WizardShell from '../components/WizardShell.vue'
import AddButton from '../components/AddButton.vue'
import AppButton from '../components/AppButton.vue'
import HTMLPreview from '../components/HTMLPreview.vue'
import AppInput from '../components/AppInput.vue'
import AppSelect from '../components/AppSelect.vue'
import MiragedForm from '../components/MiragedForm.vue'
import PhishletForm from '../components/PhishletForm.vue'
import TemplateForm from '../components/TemplateForm.vue'
import SMTPForm from '../components/SMTPForm.vue'
import TargetListPicker from '../components/TargetListPicker.vue'
import Card from '../components/Card.vue'
import ErrorBanner from '../components/ErrorBanner.vue'

const router = useRouter()
const step = ref(0)
const error = ref('')
const loading = ref(false)

// --- Step data ---
const connections = ref<MiragedConnection[]>([])
const localPhishlets = ref<Phishlet[]>([])
const targetLists = ref<TargetList[]>([])
const templates = ref<EmailTemplate[]>([])
const smtpProfiles = ref<SMTPProfile[]>([])
const dnsProviders = ref<string[]>([])

// --- Selections ---
const selectedConnectionId = ref('')
const selectedPhishletName = ref('')
const phishletHostname = ref('')
const phishletDnsProvider = ref('')
const phishletEnabled = ref(false)
const selectedTargetListId = ref('')
const selectedTemplateId = ref('')
const selectedSmtpId = ref('')
const campaignName = ref('')
const redirectUrl = ref('')
const sendRate = ref('10')

// --- Create-new toggles ---
const showNewConnection = ref(false)
const showNewPhishlet = ref(false)
const showNewTemplate = ref(false)
const showNewSmtp = ref(false)

// --- Create-new forms ---
const newConnection = ref({ name: '', address: '', secret_hostname: '', token: '' })
const newPhishletYaml = ref('')
const newTemplate = ref({ name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' })
const newSmtp = ref({ name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' })

// --- Computed ---
const steps = computed(() => {
  const base = ['Infrastructure', 'Phishlet', 'Targets', 'Template', 'SMTP', 'Review']
  if (!selectedConnectionId.value) {
    return ['Infrastructure', 'Targets', 'Template', 'SMTP', 'Review']
  }
  return base
})

const effectiveStep = computed(() => {
  if (!selectedConnectionId.value && step.value >= 1) {
    return step.value + 1
  }
  return step.value
})

const selectedTargetList = computed(() => targetLists.value.find(l => l.id === selectedTargetListId.value))
const selectedTemplate = computed(() => templates.value.find(t => t.id === selectedTemplateId.value))
const selectedSmtp = computed(() => smtpProfiles.value.find(p => p.id === selectedSmtpId.value))
const selectedConnection = computed(() => connections.value.find(c => c.id === selectedConnectionId.value))

const targetListPicker = ref<InstanceType<typeof TargetListPicker> | null>(null)
const targetCount = ref(0)
const previewResult = ref<PreviewResult | null>(null)

// --- Load data ---
async function loadAll() {
  try {
    const [conns, phish, lists, tmpls, profiles] = await Promise.all([
      listMiraged(),
      listPhishlets(),
      listTargetLists(),
      listTemplates(),
      listSMTPProfiles(),
    ])
    connections.value = conns ?? []
    localPhishlets.value = phish ?? []
    targetLists.value = lists ?? []
    templates.value = tmpls ?? []
    smtpProfiles.value = profiles ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

watch(selectedConnectionId, async (id) => {
  dnsProviders.value = []
  if (!id) return
  try {
    dnsProviders.value = await listMiragedDNSProviders(id)
  } catch { /* DNS providers are optional */ }
})

onMounted(loadAll)

// --- Navigation ---
function canAdvance(): boolean {
  switch (effectiveStep.value) {
    case 0: return !showNewConnection.value
    case 1: return !!selectedPhishletName.value && !!phishletHostname.value && !showNewPhishlet.value
    case 2: return !!selectedTargetListId.value
    case 3: return !!selectedTemplateId.value && !showNewTemplate.value
    case 4: return !!selectedSmtpId.value && !showNewSmtp.value
    case 5: return !!campaignName.value && !!redirectUrl.value
    default: return false
  }
}

async function next() {
  error.value = ''

  // Enable the phishlet on miraged before advancing past the phishlet step.
  if (effectiveStep.value === 1 && !phishletEnabled.value) {
    await enableSelectedPhishlet()
    if (!phishletEnabled.value) return
  }

  step.value++
}

function back() {
  error.value = ''
  step.value--
}

// --- Inline creation handlers ---
async function createNewConnection() {
  loading.value = true
  error.value = ''
  try {
    const conn = await enrollMiraged(newConnection.value)
    connections.value.unshift(conn)
    selectedConnectionId.value = conn.id
    showNewConnection.value = false
    newConnection.value = { name: '', address: '', secret_hostname: '', token: '' }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createNewPhishlet() {
  loading.value = true
  error.value = ''
  try {
    const phishlet = await createPhishlet(newPhishletYaml.value)
    localPhishlets.value.unshift(phishlet)
    selectedPhishletName.value = phishlet.name
    showNewPhishlet.value = false
    newPhishletYaml.value = ''
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function enableSelectedPhishlet() {
  if (!selectedConnectionId.value || !selectedPhishletName.value || !phishletHostname.value) return
  loading.value = true
  error.value = ''
  try {
    // Push the phishlet YAML to miraged first so it's available for enabling.
    const localPhishlet = localPhishlets.value.find(p => p.name === selectedPhishletName.value)
    if (localPhishlet) {
      await pushMiragedPhishlet(selectedConnectionId.value, localPhishlet.yaml)
    }

    // Then enable it
    await enableMiragedPhishlet(selectedConnectionId.value, selectedPhishletName.value, phishletHostname.value, phishletDnsProvider.value)
    phishletEnabled.value = true
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createNewTemplate() {
  loading.value = true
  error.value = ''
  try {
    const tmpl = await createTemplate(newTemplate.value)
    templates.value.unshift(tmpl)
    selectedTemplateId.value = tmpl.id
    showNewTemplate.value = false
    newTemplate.value = { name: '', subject: '', html_body: '', text_body: '', envelope_sender: '' }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

watch(selectedTemplateId, async (id) => {
  previewResult.value = null
  if (!id) return
  try {
    previewResult.value = await previewTemplate(id, {
      first_name: 'Jane', last_name: 'Doe', email: 'jane@example.com', url: 'https://phish.example.com/lure',
    })
  } catch (e: any) {
    error.value = e.message
  }
})

async function createNewSmtp() {
  loading.value = true
  error.value = ''
  try {
    const profile = await createSMTPProfile({
      ...newSmtp.value,
      port: parseInt(newSmtp.value.port) || 587,
    })
    smtpProfiles.value.unshift(profile)
    selectedSmtpId.value = profile.id
    showNewSmtp.value = false
    newSmtp.value = { name: '', host: '', port: '587', username: '', password: '', from_addr: '', from_name: '' }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

// --- Final submission ---
async function submit() {
  loading.value = true
  error.value = ''
  try {
    const campaign = await createCampaign({
      name: campaignName.value,
      template_id: selectedTemplateId.value,
      smtp_profile_id: selectedSmtpId.value,
      target_list_id: selectedTargetListId.value,
      miraged_id: selectedConnectionId.value,
      phishlet: selectedPhishletName.value,
      redirect_url: redirectUrl.value,
      send_rate: parseInt(sendRate.value) || 10,
    })
    router.push(`/campaigns/${campaign.id}`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <div class="mb-6">
      <h1>New Campaign</h1>
    </div>

    <ErrorBanner v-model="error" />

    <WizardShell :steps="steps" :currentStep="step" @back="back">
      <!-- Step 0: Infrastructure -->
      <Card v-if="effectiveStep === 0" class="p-7">
        <h6 class="mb-7">Miraged Server</h6>

        <template v-if="!showNewConnection">
          <AppSelect v-model="selectedConnectionId" label="Select a connection">
            <option value="">None (manual lure URL)</option>
            <option v-for="conn in connections" :key="conn.id" :value="conn.id">{{ conn.name }} ({{ conn.address }})</option>
          </AppSelect>

          <div class="mt-8 pt-6 border-t border-edge">
            <AddButton variant="link" @click="showNewConnection = true">Enroll new server</AddButton>
          </div>
        </template>

        <div v-else class="flex flex-col gap-7">
          <MiragedForm v-model="newConnection" />
          <div class="flex gap-2 justify-end">
            <AppButton variant="ghost" @click="showNewConnection = false">Cancel</AppButton>
            <AppButton :disabled="loading" @click="createNewConnection">{{ loading ? 'Enrolling...' : 'Enroll' }}</AppButton>
          </div>
        </div>
      </Card>

      <!-- Step 1: Phishlet (only if miraged selected) -->
      <template v-else-if="effectiveStep === 1">
        <PhishletForm v-if="showNewPhishlet" v-model="newPhishletYaml" :loading="loading"
          @submit="createNewPhishlet" @cancel="showNewPhishlet = false" />

        <Card v-else class="p-7">
          <h6 class="mb-7">Phishlet Configuration</h6>

          <div class="flex flex-col gap-7">
            <AppSelect v-model="selectedPhishletName" label="Select a phishlet">
              <option value="" disabled></option>
              <option v-for="p in localPhishlets" :key="p.id" :value="p.name">{{ p.name }}</option>
            </AppSelect>
            <div class="grid grid-cols-2 gap-7">
              <AppInput v-model="phishletHostname" placeholder="Hostname" required />
              <AppSelect v-model="phishletDnsProvider" label="DNS provider" :disabled="dnsProviders.length === 0">
                <option value="">{{ dnsProviders.length === 0 ? 'No providers configured' : 'None' }}</option>
                <option v-for="p in dnsProviders" :key="p" :value="p">{{ p }}</option>
              </AppSelect>
            </div>
          </div>

          <div class="mt-8 pt-6 border-t border-edge">
            <AddButton variant="link" @click="showNewPhishlet = true">Create new phishlet</AddButton>
          </div>
        </Card>
      </template>

      <!-- Step 2: Targets -->
      <Card v-else-if="effectiveStep === 2" class="p-7">
        <h6 class="mb-7">Target List</h6>
        <TargetListPicker ref="targetListPicker" v-model="selectedTargetListId" @update:targetCount="targetCount = $event">
          <div class="mt-8 pt-6 border-t border-edge">
            <AddButton variant="link" @click="targetListPicker?.startCreateNew()">Create new list</AddButton>
          </div>
        </TargetListPicker>
      </Card>

      <!-- Step 3: Template -->
      <Card v-else-if="effectiveStep === 3" class="p-7">
        <h6 class="mb-7">Email Template</h6>

        <template v-if="!showNewTemplate">
          <AppSelect v-model="selectedTemplateId" label="Select a template">
            <option value="" disabled></option>
            <option v-for="tmpl in templates" :key="tmpl.id" :value="tmpl.id">{{ tmpl.name }} — {{ tmpl.subject }}</option>
          </AppSelect>

          <div v-if="previewResult?.html_body" class="mt-7 border-t border-edge pt-6">
            <h6 class="mb-2">Preview</h6>
            <HTMLPreview :srcdoc="previewResult.html_body" />
          </div>

          <div class="mt-8 pt-6 border-t border-edge">
            <AddButton variant="link" @click="showNewTemplate = true">Create new template</AddButton>
          </div>
        </template>

        <div v-else class="flex flex-col gap-7">
          <TemplateForm v-model="newTemplate" min-editor-height="150px" />
          <div class="flex gap-2 justify-end">
            <AppButton variant="ghost" @click="showNewTemplate = false">Cancel</AppButton>
            <AppButton :disabled="loading" @click="createNewTemplate">Create</AppButton>
          </div>
        </div>
      </Card>

      <!-- Step 4: SMTP -->
      <Card v-else-if="effectiveStep === 4" class="p-7">
        <h6 class="mb-7">SMTP Profile</h6>

        <template v-if="!showNewSmtp">
          <AppSelect v-model="selectedSmtpId" label="Select an SMTP profile">
            <option value="" disabled></option>
            <option v-for="profile in smtpProfiles" :key="profile.id" :value="profile.id">{{ profile.name }} ({{ profile.host }})</option>
          </AppSelect>

          <div class="mt-8 pt-6 border-t border-edge">
            <AddButton variant="link" @click="showNewSmtp = true">Create new profile</AddButton>
          </div>
        </template>

        <div v-else class="flex flex-col gap-7">
          <SMTPForm v-model="newSmtp" />
          <div class="flex gap-2 justify-end">
            <AppButton variant="ghost" @click="showNewSmtp = false">Cancel</AppButton>
            <AppButton :disabled="loading" @click="createNewSmtp">Create</AppButton>
          </div>
        </div>
      </Card>

      <!-- Step 5: Review -->
      <Card v-else-if="effectiveStep === 5" class="p-7">
        <h6 class="mb-7">Campaign Settings</h6>

        <div class="flex flex-col gap-7">
          <AppInput v-model="campaignName" placeholder="Campaign name" required />
          <div class="grid grid-cols-[1fr_auto] gap-7">
            <AppInput v-model="redirectUrl" placeholder="Redirect URL (post-capture destination)" required />
            <AppInput v-model="sendRate" type="number" placeholder="Send rate (per min)" class="w-60" />
          </div>

          <div class="border-t border-edge pt-7 mt-3">
            <table class="w-full text-sm font-mono">
              <tbody>
                <tr v-if="selectedConnection" class="border-b border-edge/50">
                  <td class="py-2 text-dim uppercase tracking-wider text-xs w-32">Miraged</td>
                  <td class="py-2 text-primary">{{ selectedConnection.name }}</td>
                </tr>
                <tr v-if="selectedPhishletName" class="border-b border-edge/50">
                  <td class="py-2 text-dim uppercase tracking-wider text-xs">Phishlet</td>
                  <td class="py-2 text-primary">{{ selectedPhishletName }} <span v-if="phishletHostname" class="text-dim">({{ phishletHostname }})</span></td>
                </tr>
                <tr v-if="selectedTargetList" class="border-b border-edge/50">
                  <td class="py-2 text-dim uppercase tracking-wider text-xs">Targets</td>
                  <td class="py-2 text-primary">{{ selectedTargetList.name }} <span class="text-dim">({{ targetCount }} targets)</span></td>
                </tr>
                <tr v-if="selectedTemplate" class="border-b border-edge/50">
                  <td class="py-2 text-dim uppercase tracking-wider text-xs">Template</td>
                  <td class="py-2 text-primary">{{ selectedTemplate.name }}</td>
                </tr>
                <tr v-if="selectedSmtp">
                  <td class="py-2 text-dim uppercase tracking-wider text-xs">SMTP</td>
                  <td class="py-2 text-primary">{{ selectedSmtp.name }} <span class="text-dim">({{ selectedSmtp.host }})</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </Card>

      <template #actions>
        <AppButton v-if="effectiveStep < 5" :disabled="!canAdvance()" @click="next">
          Next
        </AppButton>
        <template v-else>
          <AppButton :disabled="loading || !canAdvance()" @click="submit">
            {{ loading ? 'Creating...' : 'Create Campaign' }}
          </AppButton>
        </template>
      </template>
    </WizardShell>
  </div>
</template>
