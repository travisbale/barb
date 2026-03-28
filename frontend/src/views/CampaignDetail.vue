<script setup lang="ts">
import { useRoute } from 'vue-router'
import EmptyState from '../components/EmptyState.vue'

const route = useRoute()
const id = route.params.id
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-lg font-mono font-semibold tracking-tight text-primary">Campaign {{ id }}</h1>
      <p class="text-xs text-dim font-mono mt-0.5">Campaign details</p>
    </div>
    <EmptyState message="Campaign view coming soon." />
  </div>
</template>
