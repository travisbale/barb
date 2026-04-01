<script setup lang="ts">
import AppInput from './AppInput.vue'

defineProps<{
  modelValue: {
    name: string
    address: string
    secret_hostname: string
    token: string
  }
}>()

defineEmits<{
  'update:modelValue': [value: any]
}>()
</script>

<template>
  <div class="flex flex-col gap-7">
    <div class="grid grid-cols-2 gap-7">
      <AppInput :modelValue="modelValue.name" @update:modelValue="$emit('update:modelValue', { ...modelValue, name: $event })" placeholder="Name" required />
      <AppInput :modelValue="modelValue.address" @update:modelValue="$emit('update:modelValue', { ...modelValue, address: $event })" placeholder="Address (host:port)" required />
    </div>
    <div class="grid grid-cols-2 gap-7">
      <AppInput :modelValue="modelValue.secret_hostname" @update:modelValue="$emit('update:modelValue', { ...modelValue, secret_hostname: $event })" placeholder="Secret hostname" required />
      <AppInput :modelValue="modelValue.token" @update:modelValue="$emit('update:modelValue', { ...modelValue, token: $event })" placeholder="Invite token" required />
    </div>
  </div>
</template>
