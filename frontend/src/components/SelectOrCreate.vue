<script setup lang="ts">
import { ref } from 'vue'
import AddButton from './AddButton.vue'
import AppButton from './AppButton.vue'
import AppSelect from './AppSelect.vue'
import Card from './Card.vue'

defineProps<{
  modelValue: string
  label: string
  items: { value: string; label: string }[]
  emptyMessage: string
  createLabel: string
  loading?: boolean
}>()

defineEmits<{
  'update:modelValue': [value: string]
  'create': []
}>()

const creating = ref(false)

function startCreate() {
  creating.value = true
}

function cancelCreate() {
  creating.value = false
}

defineExpose({ creating, startCreate, cancelCreate })
</script>

<template>
  <Card class="p-7">
    <slot name="heading" />

    <template v-if="!creating">
      <template v-if="items.length > 0">
        <AppSelect :modelValue="modelValue" @update:modelValue="$emit('update:modelValue', $event)" :label="label">
          <option v-for="item in items" :key="item.value" :value="item.value">{{ item.label }}</option>
        </AppSelect>

        <slot name="detail" />

        <div class="mt-8 pt-6 border-t border-edge">
          <AddButton variant="link" @click="startCreate">{{ createLabel }}</AddButton>
        </div>
      </template>
      <div v-else>
        <p class="text-sm text-dim font-mono">{{ emptyMessage }}</p>
        <div class="mt-8 pt-6 border-t border-edge">
          <AddButton variant="link" @click="startCreate">{{ createLabel }}</AddButton>
        </div>
      </div>
    </template>

    <div v-else class="flex flex-col gap-7">
      <slot name="form" />
      <div class="flex gap-2 justify-end">
        <AppButton variant="ghost" @click="cancelCreate">Cancel</AppButton>
        <AppButton :disabled="loading" @click="$emit('create')">Create</AppButton>
      </div>
    </div>
  </Card>
</template>
