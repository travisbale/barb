<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <div>
    <PageHeader title="SMTP Profiles">
      <AppButton>Add Profile</AppButton>
    </PageHeader>
    <EmptyState message="No SMTP profiles configured. Add one to send emails." />
  </div>
</template>
