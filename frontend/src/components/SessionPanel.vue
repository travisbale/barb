<script setup lang="ts">
import { ref } from 'vue'
import { getMiragedSession, exportMiragedSessionCookies, type MiragedSession, type CampaignResult } from '../api/client'
import { resultStatusColor } from '../utils/results'
import AppButton from './AppButton.vue'
import Card from './Card.vue'
import ErrorBanner from './ErrorBanner.vue'
import FieldValue from './FieldValue.vue'

const props = defineProps<{ miragedId: string }>()
const emit = defineEmits<{ close: [] }>()

const session = ref<MiragedSession | null>(null)
const result = ref<CampaignResult | null>(null)
const loading = ref(false)
const error = ref('')

async function open(sessionId: string, campaignResult?: CampaignResult) {
  session.value = null
  result.value = campaignResult ?? null
  error.value = ''
  loading.value = true
  try {
    session.value = await getMiragedSession(props.miragedId, sessionId)
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function close() {
  session.value = null
  result.value = null
  error.value = ''
  emit('close')
}

function downloadCookies() {
  if (!session.value) return
  const url = exportMiragedSessionCookies(props.miragedId, session.value.id)
  window.open(url, '_blank')
}

defineExpose({ open, close })
</script>

<template>
  <Card v-if="session || loading || error" class="p-7 mb-4 relative animate-in">
    <button @click="close" class="absolute top-5 right-5 text-dim hover:text-primary transition-colors text-2xl leading-none">&times;</button>

    <div v-if="loading" class="text-sm font-mono text-dim">Loading session...</div>
    <ErrorBanner v-else-if="error" v-model="error" />

    <div v-else-if="session" class="flex flex-col [&>*+*]:border-t [&>*+*]:border-edge [&>*+*]:pt-8 [&>*+*]:mt-8">
      <!-- Session details -->
      <div>
        <h6 class="mb-5">Session Details</h6>
        <div class="grid grid-cols-2 gap-x-8 gap-y-5">
          <FieldValue v-if="result" label="Target">{{ result.email }}</FieldValue>
          <FieldValue v-if="result" label="Status" :selectable="false"><span class="uppercase" :class="resultStatusColor(result.status)">{{ result.status }}</span></FieldValue>
          <FieldValue v-if="session.remote_addr" label="IP Address">{{ session.remote_addr }}</FieldValue>
          <FieldValue v-if="session.started_at" label="Started" :selectable="false">{{ new Date(session.started_at).toLocaleString() }}</FieldValue>
          <FieldValue v-if="result?.sent_at" label="Sent" :selectable="false">{{ new Date(result.sent_at).toLocaleString() }}</FieldValue>
          <FieldValue v-if="result?.clicked_at" label="Clicked" :selectable="false">{{ new Date(result.clicked_at).toLocaleString() }}</FieldValue>
          <FieldValue v-if="result?.captured_at" label="Captured" :selectable="false">{{ new Date(result.captured_at).toLocaleString() }}</FieldValue>
          <FieldValue v-if="session.user_agent" label="User Agent" class="col-span-2">{{ session.user_agent }}</FieldValue>
        </div>
      </div>

      <!-- Credentials and captured fields -->
      <div v-if="session.username || session.password || (session.custom && Object.keys(session.custom).length > 0)">
        <h6 class="mb-5">Credentials</h6>
        <div class="grid grid-cols-2 gap-x-8 gap-y-5">
          <FieldValue v-if="session.username" label="Username">{{ session.username }}</FieldValue>
          <FieldValue v-if="session.password" label="Password">{{ session.password }}</FieldValue>
          <FieldValue v-for="(value, key) in session.custom" :key="key" :label="String(key)">{{ value }}</FieldValue>
        </div>
      </div>

      <!-- Cookies -->
      <div v-if="session.cookie_tokens && Object.keys(session.cookie_tokens).length > 0">
        <div class="flex items-center justify-between mb-3 -mt-2">
          <h6>Cookies</h6>
          <AppButton variant="secondary" @click="downloadCookies">Export</AppButton>
        </div>
        <div class="flex flex-col gap-3">
          <template v-for="(cookies, domain) in session.cookie_tokens" :key="domain">
            <FieldValue v-for="(value, name) in cookies" :key="`${domain}-${name}`" :label="String(name)">{{ value }}</FieldValue>
          </template>
        </div>
      </div>

      <!-- HTTP tokens -->
      <div v-if="session.http_tokens && Object.keys(session.http_tokens).length > 0">
        <h6 class="mb-5">HTTP Tokens</h6>
        <div class="flex flex-col gap-3">
          <FieldValue v-for="(value, name) in session.http_tokens" :key="name" :label="String(name)">{{ value }}</FieldValue>
        </div>
      </div>

      <!-- Body tokens -->
      <div v-if="session.body_tokens && Object.keys(session.body_tokens).length > 0">
        <h6 class="mb-5">Body Tokens</h6>
        <div class="flex flex-col gap-3">
          <FieldValue v-for="(value, name) in session.body_tokens" :key="name" :label="String(name)">{{ value }}</FieldValue>
        </div>
      </div>
    </div>
  </Card>
</template>
