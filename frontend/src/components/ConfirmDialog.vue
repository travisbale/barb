<script setup lang="ts">
import { useConfirm } from '../composables/useConfirm'

const { visible, message, confirmLabel, confirmVariant, respond } = useConfirm()

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') respond(false)
  if (e.key === 'Enter') respond(true)
}
</script>

<template>
  <Teleport to="body">
    <Transition name="confirm">
      <div
        v-if="visible"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @keydown="onKeydown"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-bg/80 backdrop-blur-sm" @click="respond(false)" />

        <!-- Dialog -->
        <div class="relative bg-surface border border-edge p-6 w-full max-w-sm confirm-enter">
          <div class="font-mono text-xs text-dim uppercase tracking-widest mb-4">Confirm</div>
          <p class="font-mono text-sm text-primary leading-relaxed mb-6">{{ message }}</p>
          <div class="flex gap-3 justify-end">
            <button
              @click="respond(false)"
              class="px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase text-muted hover:text-primary bg-transparent transition-colors"
            >Cancel</button>
            <button
              ref="confirmBtn"
              @click="respond(true)"
              autofocus
              class="px-3 py-1.5 text-sm font-mono font-medium tracking-wide uppercase hover:brightness-110 transition-all"
              :class="confirmVariant === 'danger' ? 'bg-danger text-white' : 'bg-amber text-black'"
            >{{ confirmLabel }}</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.confirm-enter-from {
  opacity: 0;
}
.confirm-enter-from .relative {
  transform: translateY(8px);
}
.confirm-enter-active {
  transition: opacity 0.15s ease-out;
}
.confirm-enter-active .relative {
  transition: transform 0.15s ease-out;
}
.confirm-leave-active {
  transition: opacity 0.1s ease-in;
}
.confirm-leave-to {
  opacity: 0;
}

.confirm-enter {
  animation: dialogIn 0.15s ease-out both;
}

@keyframes dialogIn {
  from {
    opacity: 0;
    transform: translateY(8px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}
</style>
