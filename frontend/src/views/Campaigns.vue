<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <div>
    <PageHeader title="Campaigns">
      <AppButton>New Campaign</AppButton>
    </PageHeader>
    <EmptyState message="No campaigns yet. Create one to get started." />
  </div>
</template>
