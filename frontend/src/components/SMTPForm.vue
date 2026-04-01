<script setup lang="ts">
import AppInput from './AppInput.vue'

defineProps<{
  modelValue: {
    name: string
    host: string
    port: string
    username: string
    password: string
    from_addr: string
    from_name: string
  }
}>()

defineEmits<{
  'update:modelValue': [value: any]
}>()
</script>

<template>
  <div class="flex flex-col gap-7">
    <div class="grid grid-cols-2 gap-7">
      <AppInput :modelValue="modelValue.name" @update:modelValue="$emit('update:modelValue', { ...modelValue, name: $event })" placeholder="Profile name" />
      <AppInput :modelValue="modelValue.host" @update:modelValue="$emit('update:modelValue', { ...modelValue, host: $event })" placeholder="SMTP host" />
    </div>
    <div class="grid grid-cols-2 gap-7">
      <AppInput :modelValue="modelValue.port" @update:modelValue="$emit('update:modelValue', { ...modelValue, port: $event })" placeholder="Port" type="number" />
      <AppInput :modelValue="modelValue.from_addr" @update:modelValue="$emit('update:modelValue', { ...modelValue, from_addr: $event })" placeholder="From address" />
    </div>
    <div class="grid grid-cols-2 gap-7">
      <AppInput :modelValue="modelValue.username" @update:modelValue="$emit('update:modelValue', { ...modelValue, username: $event })" placeholder="Username" />
      <AppInput :modelValue="modelValue.password" @update:modelValue="$emit('update:modelValue', { ...modelValue, password: $event })" placeholder="Password" type="password" />
    </div>
    <AppInput :modelValue="modelValue.from_name" @update:modelValue="$emit('update:modelValue', { ...modelValue, from_name: $event })" placeholder="From name" />
  </div>
</template>
