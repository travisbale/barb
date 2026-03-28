<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <div>
    <PageHeader title="SMTP Profiles" subtitle="Mail relay configuration">
      <AppButton>+ Add Profile</AppButton>
    </PageHeader>
    <EmptyState message="No SMTP profiles. Add one to enable email delivery." />
  </div>
</template>
