<script setup lang="ts">
import Card from './Card.vue'

defineEmits<{
  submit: []
}>()
</script>

<template>
  <Card class="p-7 mb-4">
    <form @submit.prevent="$emit('submit')">
      <div class="flex flex-col gap-7">
        <slot />
      </div>
      <div class="flex gap-2 justify-end mt-7">
        <slot name="actions" />
      </div>
    </form>
  </Card>
</template>
