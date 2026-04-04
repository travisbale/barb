<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { listTargetLists, createTargetList, type TargetList } from '../api/client'
import AppButton from './AppButton.vue'
import AppInput from './AppInput.vue'
import AppSelect from './AppSelect.vue'
import TargetEditor from './TargetEditor.vue'

defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:targetCount': [value: number]
}>()

const targetLists = ref<TargetList[]>([])
const targetCount = ref(0)
const creatingList = ref(false)
const newListName = ref('')
const loading = ref(false)
const error = ref('')

async function loadLists() {
  try {
    targetLists.value = await listTargetLists() ?? []
  } catch { /* ignore */ }
}

loadLists()

watchEffect(() => emit('update:targetCount', targetCount.value))

async function createNewList() {
  loading.value = true
  error.value = ''
  try {
    const list = await createTargetList(newListName.value)
    targetLists.value.unshift(list)
    emit('update:modelValue', list.id)
    creatingList.value = false
    newListName.value = ''
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function startCreateNew() {
  creatingList.value = true
}

defineExpose({ startCreateNew, targetCount })
</script>

<template>
  <div>
    <div v-if="error" class="text-xs font-mono text-danger mb-3">{{ error }}</div>

    <template v-if="!creatingList">
      <AppSelect :modelValue="modelValue" @update:modelValue="$emit('update:modelValue', $event)" label="Select a target list">
        <option value="" disabled></option>
        <option v-for="list in targetLists" :key="list.id" :value="list.id">{{ list.name }}</option>
      </AppSelect>

      <div v-if="modelValue" class="mt-7">
        <TargetEditor :list-id="modelValue" compact @update:count="targetCount = $event" />
      </div>

      <slot />
    </template>

    <div v-else class="flex flex-col gap-7">
      <AppInput v-model="newListName" placeholder="List name" />
      <div class="flex gap-2 justify-end">
        <AppButton variant="ghost" @click="creatingList = false">Cancel</AppButton>
        <AppButton :disabled="loading" @click="createNewList">{{ loading ? 'Creating...' : 'Create' }}</AppButton>
      </div>
    </div>
  </div>
</template>
