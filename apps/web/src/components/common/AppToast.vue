<template>
  <Teleport to="body">
    <div class="toast-container">
      <div
        v-for="t in toasts"
        :key="t.id"
        :class="['toast', `toast--${t.type}`]"
      >
        {{ t.message }}
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Toast { id: number; message: string; type: 'success' | 'error' | 'info' }
const toasts = ref<Toast[]>([])
let nextId = 0

function show(message: string, type: Toast['type'] = 'info') {
  const id = ++nextId
  toasts.value.push({ id, message, type })
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, 3500)
}

defineExpose({ show })
</script>

<style scoped>
.toast-container {
  position: fixed; bottom: 24px; right: 24px;
  display: flex; flex-direction: column; gap: 8px; z-index: 9999;
}
.toast {
  padding: 10px 16px; border-radius: var(--radius);
  font-size: 13px; font-weight: 500; max-width: 320px;
  animation: slideIn 0.2s ease;
}
.toast--success { background: var(--success); color: #fff; }
.toast--error   { background: var(--error); color: #fff; }
.toast--info    { background: var(--surface2); color: var(--text); border: 1px solid var(--border); }
@keyframes slideIn { from { opacity: 0; transform: translateX(20px); } to { opacity: 1; } }
</style>
