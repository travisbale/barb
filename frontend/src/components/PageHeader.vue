<script setup lang="ts">
import Breadcrumbs from './Breadcrumbs.vue'

defineProps<{
  title: string
  subtitle?: string
  breadcrumbs?: { label: string; to?: string }[]
}>()
</script>

<template>
  <div class="mb-8">
    <Breadcrumbs v-if="breadcrumbs" :crumbs="breadcrumbs" />
    <div class="flex items-center justify-between">
      <div>
        <h1>{{ title }}</h1>
        <p v-if="subtitle" class="text-sm text-dim font-mono mt-1">{{ subtitle }}</p>
        <slot name="subtitle" />
      </div>
      <div class="flex gap-2">
        <slot />
      </div>
    </div>
  </div>
</template>
