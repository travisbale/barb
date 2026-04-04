<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getTargetList, type TargetList } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import TargetEditor from '../components/TargetEditor.vue'

const route = useRoute()
const id = route.params.id as string

const list = ref<TargetList | null>(null)
const error = ref('')

async function load() {
  try {
    list.value = await getTargetList(id)
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader
      :title="list?.name ?? '...'"
      :breadcrumbs="[{ label: 'Target Lists', to: '/targets' }, { label: list?.name ?? '...' }]"
    />

    <TargetEditor :list-id="id" />
  </div>
</template>
