<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <div>
    <PageHeader title="Campaigns" subtitle="Active operations">
      <AppButton>+ New Campaign</AppButton>
    </PageHeader>
    <EmptyState message="No campaigns. Create one to begin an operation." />
  </div>
</template>
