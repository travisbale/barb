<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { changePassword } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'

const router = useRouter()
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  loading.value = true
  error.value = ''
  try {
    await changePassword(currentPassword.value, newPassword.value)
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
        <div class="font-mono text-xs text-dim mt-1 tracking-wider">Change your password to continue</div>
      </div>

      <form @submit.prevent="submit" class="flex flex-col gap-5">
        <AppInput v-model="currentPassword" placeholder="Current password" type="password" required autofocus />
        <AppInput v-model="newPassword" placeholder="New password (8+ characters)" type="password" required />
        <AppInput v-model="confirmPassword" placeholder="Confirm new password" type="password" required />

        <div v-if="error" class="text-xs text-danger font-mono">{{ error }}</div>

        <AppButton type="submit" :disabled="loading" class="w-full py-3 justify-center">
          {{ loading ? 'Changing...' : 'Change Password' }}
        </AppButton>
      </form>
    </div>
  </div>
</template>
