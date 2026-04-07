<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { listTargetLists, createTargetList, deleteTargetList, type TargetList } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import AppButton from '../components/AppButton.vue'
import DeleteButton from '../components/DeleteButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import FormCard from '../components/FormCard.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
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

    <ErrorBanner v-model="error" />

    <FormCard v-if="showCreate" @submit="create">
      <AppInput v-model="newName" placeholder="List name" autofocus />
      <template #actions>
        <AppButton variant="ghost" @click="showCreate = false">Cancel</AppButton>
        <AppButton type="submit">Create</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="lists.length === 0 && !showCreate" message="No target lists. Create one to begin." />

    <DataTable v-else-if="lists.length > 0" :columns="[{ label: 'Name' }, { label: 'Created' }, { label: '', width: 'w-16' }]">
      <DataTableRow
        v-for="(list, i) in lists"
        :key="list.id"
        :index="i"
        clickable
        @click="router.push(`/targets/${list.id}`)"
      >
        <td class="text-primary">{{ list.name }}</td>
        <td class="text-dim">{{ new Date(list.created_at).toLocaleDateString() }}</td>
        <td class="text-right">
          <DeleteButton @click.stop="remove(list.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
