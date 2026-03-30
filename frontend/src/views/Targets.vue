<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { listTargetLists, createTargetList, deleteTargetList, type TargetList } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import AddButton from '../components/AddButton.vue'

const router = useRouter()
const { confirm } = useConfirm()
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
  if (!await confirm('Delete this target list?')) return
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
    <PageHeader title="Target Lists" :subtitle="`${lists.length} lists`">
      <AddButton @click="showCreate = true">New List</AddButton>
    </PageHeader>

    <ErrorBanner :message="error" />

    <Card v-if="showCreate" class="p-7 mb-4">
      <form @submit.prevent="create" class="flex gap-3 items-center">
        <AppInput v-model="newName" placeholder="List name..." autofocus class="flex-1" />
        <AppButton type="submit">Create</AppButton>
        <AppButton variant="ghost" @click="showCreate = false">Cancel</AppButton>
      </form>
    </Card>

    <EmptyState v-if="lists.length === 0 && !showCreate" message="No target lists. Create one to begin." />

    <Card v-else-if="lists.length > 0">
      <div
        v-for="(list, i) in lists"
        :key="list.id"
        :style="{ animationDelay: `${i * 30}ms` }"
        class="animate-in flex items-center justify-between px-4 py-3 border-b border-edge last:border-0 hover:bg-surface-hover cursor-pointer transition-colors"
        @click="router.push(`/targets/${list.id}`)"
      >
        <div class="flex items-center gap-3">
          <span class="text-amber text-xs font-mono">&#x25C9;</span>
          <div>
            <div class="text-sm font-medium text-primary">{{ list.name }}</div>
            <div class="text-xs text-dim font-mono mt-0.5">{{ new Date(list.created_at).toLocaleDateString() }}</div>
          </div>
        </div>
        <button @click.stop="remove(list.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
      </div>
    </Card>
  </div>
</template>
