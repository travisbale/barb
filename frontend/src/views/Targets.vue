<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listTargetLists, createTargetList, deleteTargetList, type TargetList } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'

const router = useRouter()
const lists = ref<TargetList[]>([])
const showCreate = ref(false)
const newName = ref('')
const error = ref('')

async function load() {
  try {
    lists.value = await listTargetLists() ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function create() {
  if (!newName.value.trim()) return
  try {
    const list = await createTargetList(newName.value.trim())
    lists.value.unshift(list)
    newName.value = ''
    showCreate.value = false
    router.push(`/targets/${list.id}`)
  } catch (e: any) {
    error.value = e.message
  }
}

async function remove(id: string) {
  try {
    await deleteTargetList(id)
    lists.value = lists.value.filter(l => l.id !== id)
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <PageHeader title="Target Lists">
      <AppButton @click="showCreate = true">New List</AppButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showCreate" class="p-4 mb-4">
      <form @submit.prevent="create" class="flex gap-3">
        <AppInput v-model="newName" placeholder="List name" autofocus class="flex-1" />
        <AppButton type="submit">Create</AppButton>
        <AppButton variant="ghost" @click="showCreate = false">Cancel</AppButton>
      </form>
    </Card>

    <EmptyState v-if="lists.length === 0 && !showCreate" message="No target lists yet. Create one to get started." />

    <Card v-else-if="lists.length > 0" class="divide-y divide-gray-200">
      <div
        v-for="list in lists"
        :key="list.id"
        class="flex items-center justify-between px-4 py-3 hover:bg-gray-50 cursor-pointer"
        @click="router.push(`/targets/${list.id}`)"
      >
        <div>
          <div class="text-sm font-medium">{{ list.name }}</div>
          <div class="text-xs text-gray-400">{{ new Date(list.created_at).toLocaleDateString() }}</div>
        </div>
        <AppButton variant="danger" @click.stop="remove(list.id)">Delete</AppButton>
      </div>
    </Card>
  </div>
</template>
