<template>
  <div class="progress-wrap">
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: `${clampedValue}%` }" />
    </div>
    <span class="progress-label">{{ clampedValue }}%</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ value: number }>()
const clampedValue = computed(() => Math.max(0, Math.min(100, Math.round(props.value))))
</script>

<style scoped>
.progress-wrap { display: flex; align-items: center; gap: 10px; }
.progress-bar {
  flex: 1; height: 6px; background: var(--surface2);
  border-radius: 3px; overflow: hidden;
}
.progress-fill {
  height: 100%; background: var(--accent);
  border-radius: 3px; transition: width 0.3s ease;
}
.progress-label { font-size: 12px; color: var(--text-muted); min-width: 32px; }
</style>
