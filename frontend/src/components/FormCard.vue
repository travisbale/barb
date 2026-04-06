<script setup lang="ts">
import Card from './Card.vue'

defineEmits<{
  submit: []
}>()
</script>

<template>
  <Card class="p-7 mb-4 animate-in">
    <form @submit.prevent="$emit('submit')">
      <div class="flex flex-col gap-7">
        <slot />
      </div>
      <div class="flex justify-between mt-7">
        <div>
          <slot name="toolbar" />
        </div>
        <div class="flex gap-2">
          <slot name="actions" />
        </div>
      </div>
    </form>
  </Card>
</template>
