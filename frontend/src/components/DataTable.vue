<script setup lang="ts">
defineProps<{
  columns: { label: string; width?: string }[]
}>()
</script>

<template>
  <div class="bg-surface border border-edge overflow-hidden">
    <table class="w-full text-sm font-mono">
      <thead>
        <tr class="border-b border-edge text-dim text-left uppercase tracking-wider">
          <th
            v-for="(col, i) in columns"
            :key="i"
            :class="col.width"
            class="px-4 py-2.5 font-medium"
          >{{ col.label }}</th>
        </tr>
      </thead>
      <tbody>
        <slot />
      </tbody>
    </table>
  </div>
</template>
