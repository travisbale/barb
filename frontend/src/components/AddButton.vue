<script setup lang="ts">
import AppButton from './AppButton.vue'
import IconPlus from './IconPlus.vue'

defineProps<{
  variant?: 'primary' | 'link'
  disabled?: boolean
}>()
</script>

<template>
  <AppButton :variant="variant" :disabled="disabled">
    <IconPlus :size="variant === 'link' ? 11 : 14" /> 
    <slot />
  </AppButton>
</template>
