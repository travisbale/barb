<script setup lang="ts">
import Card from '../components/Card.vue'
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-lg font-mono font-semibold tracking-tight text-primary">Settings</h1>
      <p class="text-xs text-dim font-mono mt-0.5">Console configuration</p>
    </div>

    <Card class="p-5">
      <h2 class="text-sm font-mono font-semibold text-muted uppercase tracking-wider mb-3">Mirage Connection</h2>
      <p class="text-sm text-dim font-mono leading-relaxed">
        Configure the mTLS connection to your miraged instance.
        The console uses the Mirage API to create lures and monitor captured sessions.
      </p>
    </Card>
  </div>
</template>
