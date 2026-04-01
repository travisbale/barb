<script setup lang="ts">
import Card from './Card.vue'
import AppButton from './AppButton.vue'

defineProps<{
  label: string
  editable?: boolean
  expanded?: boolean
  saving?: boolean
}>()

defineEmits<{
  change: []
  cancel: []
  save: []
}>()
</script>

<template>
  <Card class="p-5">
    <div class="flex items-center justify-between mb-1">
      <div class="text-xs font-mono text-dim uppercase tracking-wider">{{ label }}</div>
      <button v-if="editable && !expanded" @click="$emit('change')" class="text-xs font-mono text-amber hover:text-amber-dim transition-colors uppercase tracking-wider">Change</button>
    </div>

    <div v-if="expanded" class="mt-7">
      <slot name="editor" />

      <div class="flex items-center justify-between mt-7">
        <div>
          <slot name="create-new" />
        </div>
        <div class="flex gap-2">
          <AppButton variant="ghost" @click="$emit('cancel')">Cancel</AppButton>
          <AppButton :disabled="saving" @click="$emit('save')">{{ saving ? 'Saving...' : 'Save' }}</AppButton>
        </div>
      </div>
    </div>

    <div v-else class="mt-2 text-sm font-mono">
      <slot name="summary" />
    </div>
  </Card>
</template>
