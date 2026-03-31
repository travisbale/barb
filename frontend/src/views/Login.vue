<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { login } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'

const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  error.value = ''
  try {
    await login(username.value, password.value)
    router.push('/')
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center">
    <div class="w-full max-w-sm">
      <div class="text-center mb-8">
        <div class="font-mono text-2xl font-bold tracking-widest text-amber uppercase">Barb</div>
        <div class="font-mono text-xs text-dim mt-1 tracking-wider">Operations Console</div>
      </div>

      <form @submit.prevent="submit" class="flex flex-col gap-5">
        <AppInput v-model="username" placeholder="Username" required autofocus />
        <AppInput v-model="password" placeholder="Password" type="password" required />

        <div v-if="error" class="text-xs text-danger font-mono">{{ error }}</div>

        <AppButton type="submit" :disabled="loading" class="w-full justify-center">
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </AppButton>
      </form>
    </div>
  </div>
</template>
