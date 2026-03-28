<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <div>
    <PageHeader title="Email Templates">
      <AppButton>New Template</AppButton>
    </PageHeader>
    <EmptyState message="No email templates yet. Create one to get started." />
  </div>
</template>
