<script setup lang="ts">
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <div>
    <PageHeader title="Email Templates" subtitle="Reusable message templates">
      <AppButton>+ New Template</AppButton>
    </PageHeader>
    <EmptyState message="No templates. Create one to compose phishing emails." />
  </div>
</template>
