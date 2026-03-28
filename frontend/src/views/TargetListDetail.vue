<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getTargetList, listTargets, addTarget, deleteTarget, importTargetsCSV, type TargetList, type Target } from '../api/client'
import AppButton from '../components/AppButton.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'

const route = useRoute()
const router = useRouter()
const id = route.params.id as string

const list = ref<TargetList | null>(null)
const targets = ref<Target[]>([])
const showAdd = ref(false)
const error = ref('')

const form = ref({ email: '', first_name: '', last_name: '', department: '', position: '' })

async function load() {
  try {
    list.value = await getTargetList(id)
    targets.value = await listTargets(id) ?? []
  } catch (e: any) {
    error.value = e.message
  }
}

async function add() {
  if (!form.value.email.trim()) return
  try {
    const target = await addTarget(id, form.value)
    targets.value.push(target)
    form.value = { email: '', first_name: '', last_name: '', department: '', position: '' }
    showAdd.value = false
  } catch (e: any) {
    error.value = e.message
  }
}

const fileInput = ref<HTMLInputElement | null>(null)
const importing = ref(false)

async function importCSV(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  importing.value = true
  try {
    const result = await importTargetsCSV(id, file)
    targets.value = await listTargets(id) ?? []
    error.value = ''
    alert(`Imported ${result.imported} targets`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    importing.value = false
    input.value = ''
  }
}

async function remove(targetId: string) {
  try {
    await deleteTarget(targetId)
    targets.value = targets.value.filter(t => t.id !== targetId)
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="flex items-center gap-3 mb-6">
      <button @click="router.push('/targets')" class="text-gray-400 hover:text-gray-600">&larr;</button>
      <h1 class="text-2xl font-semibold">{{ list?.name ?? 'Loading...' }}</h1>
      <span class="text-sm text-gray-400">{{ targets.length }} targets</span>
    </div>

    <ErrorBanner :message="error" />

    <div class="mb-4 flex gap-3">
      <AppButton @click="showAdd = true">Add Target</AppButton>
      <AppButton variant="secondary" :disabled="importing" @click="fileInput?.click()">
        {{ importing ? 'Importing...' : 'Import CSV' }}
      </AppButton>
      <input ref="fileInput" type="file" accept=".csv" class="hidden" @change="importCSV" />
    </div>

    <Card v-if="showAdd" class="p-4 mb-4">
      <form @submit.prevent="add" class="grid grid-cols-2 gap-3">
        <AppInput v-model="form.email" type="email" placeholder="Email (required)" required class="col-span-2" />
        <AppInput v-model="form.first_name" placeholder="First name" />
        <AppInput v-model="form.last_name" placeholder="Last name" />
        <AppInput v-model="form.department" placeholder="Department" />
        <AppInput v-model="form.position" placeholder="Position" />
        <div class="col-span-2 flex gap-3">
          <AppButton type="submit">Add</AppButton>
          <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        </div>
      </form>
    </Card>

    <EmptyState v-if="targets.length === 0 && !showAdd" message="No targets yet. Add one manually or import a CSV." />

    <Card v-else-if="targets.length > 0" class="overflow-hidden">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 text-gray-500 text-left">
          <tr>
            <th class="px-4 py-2 font-medium">Email</th>
            <th class="px-4 py-2 font-medium">Name</th>
            <th class="px-4 py-2 font-medium">Department</th>
            <th class="px-4 py-2 font-medium">Position</th>
            <th class="px-4 py-2 font-medium"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr v-for="t in targets" :key="t.id" class="hover:bg-gray-50">
            <td class="px-4 py-2">{{ t.email }}</td>
            <td class="px-4 py-2">{{ [t.first_name, t.last_name].filter(Boolean).join(' ') || '-' }}</td>
            <td class="px-4 py-2 text-gray-500">{{ t.department || '-' }}</td>
            <td class="px-4 py-2 text-gray-500">{{ t.position || '-' }}</td>
            <td class="px-4 py-2 text-right">
              <AppButton variant="danger" @click="remove(t.id)">Delete</AppButton>
            </td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
