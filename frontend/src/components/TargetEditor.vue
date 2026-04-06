<script setup lang="ts">
import { ref, watch } from 'vue'
import { useConfirm } from '../composables/useConfirm'
import { listTargets, addTarget, deleteTarget, importTargetsCSV, type Target } from '../api/client'
import AppButton from './AppButton.vue'
import AppInput from './AppInput.vue'
import DeleteButton from './DeleteButton.vue'
import AddButton from './AddButton.vue'
import FormCard from './FormCard.vue'
import ErrorBanner from './ErrorBanner.vue'
import EmptyState from './EmptyState.vue'
import DataTable from './DataTable.vue'
import DataTableRow from './DataTableRow.vue'

const props = defineProps<{
  listId: string
  compact?: boolean
}>()

const emit = defineEmits<{
  'update:count': [value: number]
}>()

const { confirm } = useConfirm()
const targets = ref<Target[]>([])
const showAdd = ref(false)
const importing = ref(false)
const error = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const form = ref({ email: '', first_name: '', last_name: '', department: '', position: '' })

async function load() {
  try {
    targets.value = await listTargets(props.listId) ?? []
    emit('update:count', targets.value.length)
  } catch (e: any) {
    error.value = e.message
  }
}

watch(() => props.listId, () => {
  if (props.listId) load()
}, { immediate: true })

async function add() {
  if (!form.value.email.trim()) return
  error.value = ''
  try {
    const target = await addTarget(props.listId, form.value)
    targets.value.push(target)
    form.value = { email: '', first_name: '', last_name: '', department: '', position: '' }
    showAdd.value = false
    emit('update:count', targets.value.length)
  } catch (e: any) {
    error.value = e.message
  }
}

async function handleImport(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  importing.value = true
  error.value = ''
  try {
    const result = await importTargetsCSV(props.listId, file)
    targets.value = await listTargets(props.listId) ?? []
    emit('update:count', targets.value.length)
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
    emit('update:count', targets.value.length)
  } catch (e: any) {
    error.value = e.message
  }
}
</script>

<template>
  <div>
    <ErrorBanner v-model="error" />

    <!-- Actions bar -->
    <div class="flex items-center gap-2 mb-4">
      <span class="text-xs font-mono text-dim flex-1">{{ targets.length }} {{ targets.length === 1 ? 'target' : 'targets' }}</span>
      <AppButton variant="secondary" :disabled="importing" @click="fileInput?.click()">
        {{ importing ? 'Importing...' : 'Import CSV' }}
      </AppButton>
      <AddButton @click="showAdd = true">Add Target</AddButton>
      <input ref="fileInput" type="file" accept=".csv" class="hidden" @change="handleImport" />
    </div>

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
    <DataTable v-else-if="targets.length > 0" :columns="compact
      ? [{ label: 'Email' }, { label: 'Name' }, { label: '', width: 'w-12' }]
      : [{ label: 'Email' }, { label: 'Name' }, { label: 'Dept' }, { label: 'Position' }, { label: '', width: 'w-12' }]
    ">
      <DataTableRow
        v-for="(target, i) in compact ? targets.slice(0, 10) : targets"
        :key="target.id"
        :index="i"
      >
        <td class="px-4 py-2.5 text-primary">{{ target.email }}</td>
        <td class="px-4 py-2.5 text-muted">{{ [target.first_name, target.last_name].filter(Boolean).join(' ') || '—' }}</td>
        <template v-if="!compact">
          <td class="px-4 py-2.5 text-dim">{{ target.department || '—' }}</td>
          <td class="px-4 py-2.5 text-dim">{{ target.position || '—' }}</td>
        </template>
        <td class="px-4 py-2.5 text-right" @click.stop>
          <DeleteButton @click="remove(target.id)" />
        </td>
      </DataTableRow>
    </DataTable>
    <div v-if="compact && targets.length > 10" class="text-xs font-mono text-dim mt-2 px-4">
      and {{ targets.length - 10 }} more...
    </div>
  </div>
</template>
