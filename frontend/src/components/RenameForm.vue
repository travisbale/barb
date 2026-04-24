<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import AppButton from './AppButton.vue'
import AppInput from './AppInput.vue'
import FormCard from './FormCard.vue'

const props = withDefaults(defineProps<{
  value: string
  saving?: boolean
  label?: string
}>(), {
  label: 'Settings',
})

const emit = defineEmits<{
  save: [value: string]
}>()

const edit = ref(props.value)
const dirty = computed(() => edit.value !== props.value)

// Resync the input when the parent's value changes (e.g. after a save commits).
watch(() => props.value, (v) => {
  edit.value = v
})

function cancel() {
  edit.value = props.value
}

function submit() {
  if (dirty.value) emit('save', edit.value)
}
</script>

<template>
  <FormCard @submit="submit">
    <h3 v-if="label" class="text-sm font-medium text-primary">{{ label }}</h3>
    <AppInput v-model="edit" placeholder="Name" required />
    <template #actions>
      <AppButton variant="ghost" type="button" :disabled="!dirty" @click="cancel">Cancel</AppButton>
      <AppButton type="submit" :disabled="!dirty || saving">{{ saving ? 'Saving...' : 'Save' }}</AppButton>
    </template>
  </FormCard>
</template>
