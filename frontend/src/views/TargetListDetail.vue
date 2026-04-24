<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getTargetList, updateTargetList, type TargetList } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import RenameForm from '../components/RenameForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import TargetEditor from '../components/TargetEditor.vue'

const route = useRoute()
const id = route.params.id as string

const list = ref<TargetList | null>(null)
const error = ref('')
const saving = ref(false)

async function load() {
  try {
    list.value = await getTargetList(id)
  } catch (e: any) {
    error.value = e.message
  }
}

async function saveName(name: string) {
  saving.value = true
  error.value = ''
  try {
    list.value = await updateTargetList(id, { name })
  } catch (e: any) {
    error.value = e.message
  } finally {
    saving.value = false
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

    <ErrorBanner v-model="error" />

    <RenameForm v-if="list" :value="list.name" :saving="saving" @save="saveName" />

    <div v-if="list" class="mt-12">
      <TargetEditor :list-id="id" />
    </div>
  </div>
</template>
