<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  listTargetLists, createTargetList, addTarget, importTargetsCSV, listTargets,
  type TargetList, type Target,
} from '../api/client'
import AppButton from './AppButton.vue'
import AppInput from './AppInput.vue'
import AppSelect from './AppSelect.vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const targetLists = ref<TargetList[]>([])
const targetCount = ref(0)
const targetPreview = ref<Target[]>([])
const showTargetPreview = ref(false)
const showAddTargets = ref(false)
const creatingList = ref(false)
const newListName = ref('')
const newTarget = ref({ email: '', first_name: '', last_name: '', department: '', position: '' })
const loading = ref(false)
const error = ref('')

async function loadLists() {
  try {
    targetLists.value = await listTargetLists() ?? []
  } catch { /* ignore */ }
}

loadLists()

watch(() => props.modelValue, async (id) => {
  targetPreview.value = []
  targetCount.value = 0
  showTargetPreview.value = false
  showAddTargets.value = false
  if (!id) return
  try {
    const targets = await listTargets(id)
    targetPreview.value = targets ?? []
    targetCount.value = targetPreview.value.length
  } catch { /* ignore */ }
}, { immediate: true })

async function createNewList() {
  loading.value = true
  error.value = ''
  try {
    const list = await createTargetList(newListName.value)
    targetLists.value.unshift(list)
    emit('update:modelValue', list.id)
    creatingList.value = false
    showAddTargets.value = true
    newListName.value = ''
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleCsvImport(event: Event) {
  if (!props.modelValue) return
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  loading.value = true
  error.value = ''
  try {
    await importTargetsCSV(props.modelValue, file)
    const targets = await listTargets(props.modelValue)
    targetPreview.value = targets ?? []
    targetCount.value = targetPreview.value.length
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
    input.value = ''
  }
}

function startCreateNew() {
  creatingList.value = true
}

defineExpose({ startCreateNew, targetCount })

async function addNewTarget() {
  if (!props.modelValue || !newTarget.value.email) return
  loading.value = true
  error.value = ''
  try {
    await addTarget(props.modelValue, newTarget.value)
    const targets = await listTargets(props.modelValue)
    targetPreview.value = targets ?? []
    targetCount.value = targetPreview.value.length
    newTarget.value = { email: '', first_name: '', last_name: '', department: '', position: '' }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div>
    <div v-if="error" class="text-xs font-mono text-danger mb-3">{{ error }}</div>

    <template v-if="!creatingList">
      <AppSelect :modelValue="modelValue" @update:modelValue="$emit('update:modelValue', $event)" label="Select a target list">
        <option value="" disabled></option>
        <option v-for="list in targetLists" :key="list.id" :value="list.id">{{ list.name }}</option>
      </AppSelect>

      <!-- Selected list: preview + actions -->
      <div v-if="modelValue" class="mt-4">
        <div class="flex items-center gap-3 text-xs font-mono text-dim">
          <span>{{ targetCount }} {{ targetCount === 1 ? 'target' : 'targets' }}</span>
          <button v-if="targetCount > 0" @click="showTargetPreview = !showTargetPreview" class="text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">
            {{ showTargetPreview ? 'Hide' : 'Preview' }}
          </button>
        </div>

        <!-- Target preview (first 10) -->
        <div v-if="showTargetPreview && targetPreview.length > 0" class="mt-3 border border-edge overflow-hidden">
          <table class="w-full text-xs font-mono">
            <tbody>
              <tr
                v-for="target in targetPreview.slice(0, 10)"
                :key="target.id"
                class="border-b border-edge/50 last:border-0"
              >
                <td class="px-3 py-1.5 text-primary">{{ target.email }}</td>
                <td class="px-3 py-1.5 text-muted">{{ [target.first_name, target.last_name].filter(Boolean).join(' ') || '\u2014' }}</td>
              </tr>
            </tbody>
          </table>
          <div v-if="targetPreview.length > 10" class="px-3 py-1.5 text-xs font-mono text-dim border-t border-edge/50">
            and {{ targetPreview.length - 10 }} more...
          </div>
        </div>

        <!-- Add more targets -->
        <div class="mt-4">
          <button @click="showAddTargets = !showAddTargets" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">
            {{ showAddTargets ? 'Hide' : '+ Add targets' }}
          </button>
        </div>

        <div v-if="showAddTargets" class="mt-4 border-t border-edge pt-4 flex flex-col gap-4">
          <label class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase border border-edge text-muted hover:text-amber hover:border-amber/30 cursor-pointer transition-all duration-150 self-start">
            Import CSV
            <input type="file" accept=".csv" class="hidden" @change="handleCsvImport" />
          </label>

          <form @submit.prevent="addNewTarget" class="flex gap-3 items-center">
            <AppInput v-model="newTarget.email" type="email" placeholder="Email (required)" required class="flex-1" />
            <AppInput v-model="newTarget.first_name" placeholder="First name" class="flex-1" />
            <AppInput v-model="newTarget.last_name" placeholder="Last name" class="flex-1" />
            <AppButton type="submit" :disabled="loading || !newTarget.email">Add</AppButton>
          </form>
        </div>
      </div>

      <!-- Slot for callers to place actions (create-new, cancel, save) -->
      <slot />
    </template>

    <div v-else class="flex gap-3 items-end">
      <AppInput v-model="newListName" placeholder="List name" class="flex-1" />
      <AppButton variant="ghost" @click="creatingList = false">Cancel</AppButton>
      <AppButton :disabled="loading" @click="createNewList">{{ loading ? 'Creating...' : 'Create' }}</AppButton>
    </div>
  </div>
</template>
