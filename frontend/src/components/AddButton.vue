<script setup lang="ts">
import AppButton from './AppButton.vue'
import IconPlus from './IconPlus.vue'

defineProps<{
  disabled?: boolean
}>()
</script>

<template>
  <AppButton :disabled="disabled"><IconPlus /> <slot /></AppButton>
</template>
