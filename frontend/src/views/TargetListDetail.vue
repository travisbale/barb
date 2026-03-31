<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { getTargetList, listTargets, addTarget, deleteTarget, importTargetsCSV, type TargetList, type Target } from '../api/client'
import AppButton from '../components/AppButton.vue'
import IconTrash from '../components/IconTrash.vue'
import AppInput from '../components/AppInput.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import Card from '../components/Card.vue'
import FormCard from '../components/FormCard.vue'
import AddButton from '../components/AddButton.vue'
import PageHeader from '../components/PageHeader.vue'

const route = useRoute()
const { confirm } = useConfirm()
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
  if (!await confirm('Delete this target?')) return
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
    <PageHeader
      :title="list?.name ?? '...'"
      :subtitle="`${targets.length} targets`"
      :breadcrumbs="[{ label: 'Target Lists', to: '/targets' }, { label: list?.name ?? '...' }]"
    >
      <AppButton variant="secondary" :disabled="importing" @click="fileInput?.click()">
        {{ importing ? 'Importing...' : 'Import CSV' }}
      </AppButton>
      <AddButton @click="showAdd = true">Add Target</AddButton>
      <input ref="fileInput" type="file" accept=".csv" class="hidden" @change="importCSV" />
    </PageHeader>

    <ErrorBanner :message="error" />

    <!-- Add target form -->
    <FormCard v-if="showAdd" @submit="add">
      <div class="grid grid-cols-2 gap-7">
        <AppInput v-model="form.email" type="email" placeholder="Email" required class="col-span-2" />
        <AppInput v-model="form.first_name" placeholder="First name" />
        <AppInput v-model="form.last_name" placeholder="Last name" />
        <AppInput v-model="form.department" placeholder="Department" />
        <AppInput v-model="form.position" placeholder="Position" />
      </div>
      <template #actions>
        <AppButton variant="ghost" @click="showAdd = false">Cancel</AppButton>
        <AppButton type="submit">Add</AppButton>
      </template>
    </FormCard>

    <EmptyState v-if="targets.length === 0 && !showAdd" message="No targets. Add manually or import a CSV." />

    <!-- Target table -->
    <Card v-else-if="targets.length > 0" class="overflow-hidden">
      <table class="w-full text-sm font-mono">
        <thead>
          <tr class="border-b border-edge text-dim text-left uppercase tracking-wider">
            <th class="px-4 py-2.5 font-medium">Email</th>
            <th class="px-4 py-2.5 font-medium">Name</th>
            <th class="px-4 py-2.5 font-medium">Dept</th>
            <th class="px-4 py-2.5 font-medium">Position</th>
            <th class="px-4 py-2.5 font-medium w-16"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(target, i) in targets"
            :key="target.id"
            :style="{ animationDelay: `${i * 20}ms` }"
            class="animate-in border-b border-edge/50 last:border-0 hover:bg-surface-hover transition-colors"
          >
            <td class="px-4 py-2.5 text-primary">{{ target.email }}</td>
            <td class="px-4 py-2.5 text-muted">{{ [target.first_name, target.last_name].filter(Boolean).join(' ') || '—' }}</td>
            <td class="px-4 py-2.5 text-dim">{{ target.department || '—' }}</td>
            <td class="px-4 py-2.5 text-dim">{{ target.position || '—' }}</td>
            <td class="px-4 py-2.5 text-right">
              <button @click="remove(target.id)" class="text-dim hover:text-danger transition-colors"><IconTrash /></button>
            </td>
          </tr>
        </tbody>
      </table>
    </Card>
  </div>
</template>
