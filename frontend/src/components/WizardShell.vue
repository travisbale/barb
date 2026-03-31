<script setup lang="ts">
defineProps<{
  steps: string[]
  currentStep: number
}>()

defineEmits<{
  back: []
  next: []
  complete: []
}>()
</script>

<template>
  <div>
    <!-- Step indicator -->
    <div class="flex items-center gap-2 mb-8">
      <template v-for="(label, i) in steps" :key="i">
        <div v-if="i > 0" class="flex-1 h-px" :class="i <= currentStep ? 'bg-amber/40' : 'bg-edge'" />
        <div class="flex items-center gap-2">
          <div
            class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-mono font-bold transition-colors"
            :class="i < currentStep ? 'bg-amber text-bg' : i === currentStep ? 'border-2 border-amber text-amber' : 'border border-edge text-dim'"
          >{{ i + 1 }}</div>
          <span
            class="text-xs font-mono uppercase tracking-wider hidden sm:inline transition-colors"
            :class="i <= currentStep ? 'text-primary' : 'text-dim'"
          >{{ label }}</span>
        </div>
      </template>
    </div>

    <!-- Step content -->
    <slot />

    <!-- Navigation -->
    <div class="flex justify-between mt-8">
      <button
        v-if="currentStep > 0"
        @click="$emit('back')"
        class="text-sm font-mono text-muted hover:text-primary transition-colors uppercase tracking-wide"
      >Back</button>
      <div v-else />

      <div class="flex gap-2">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>
