<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()
const id = route.params.id
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold mb-6">Campaign {{ id }}</h1>
    <div class="bg-white rounded-lg border border-gray-200 p-8 text-center text-gray-500">
      Campaign details will be shown here.
    </div>
  </div>
</template>
