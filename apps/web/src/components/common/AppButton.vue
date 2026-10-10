<template>
  <button
    :class="[
      'app-btn',
      `app-btn--${variant}`,
      `app-btn--${size}`,
      { 'app-btn--loading': loading, 'app-btn--block': block }
    ]"
    :disabled="disabled || loading"
    v-bind="$attrs"
  >
    <span v-if="loading" class="app-btn__spinner" />
    <slot name="icon-left" />
    <slot />
    <slot name="icon-right" />
  </button>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'outline' | 'danger' | 'ghost' | 'success'
    size?: 'sm' | 'md' | 'lg'
    loading?: boolean
    disabled?: boolean
    block?: boolean
  }>(),
  {
    variant: 'primary',
    size: 'md',
    loading: false,
    disabled: false,
    block: false,
  }
)
</script>

<style scoped>
.app-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 500;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  transition: all 0.15s ease;
  user-select: none;
  white-space: nowrap;
  line-height: 1.4;
}

/* Sizes */
.app-btn--sm {
  padding: 6px 12px;
  font-size: 13px;
  border-radius: 6px;
}
.app-btn--md {
  padding: 8px 16px;
  font-size: 14px;
  border-radius: var(--radius-sm);
}
.app-btn--lg {
  padding: 10px 20px;
  font-size: 15px;
  border-radius: var(--radius-md);
}
.app-btn--block {
  width: 100%;
}

/* Variants */
.app-btn--primary {
  background: var(--primary);
  color: #FFFFFF;
  border-color: var(--primary);
  box-shadow: 0 1px 2px rgba(22, 119, 255, 0.2);
}
.app-btn--primary:hover:not(:disabled) {
  background: var(--primary-hover);
  border-color: var(--primary-hover);
}

.app-btn--secondary {
  background: var(--surface-secondary);
  color: var(--text-primary);
  border-color: var(--border);
}
.app-btn--secondary:hover:not(:disabled) {
  background: #E2E8F0;
}

.app-btn--outline {
  background: #FFFFFF;
  color: var(--text-primary);
  border-color: var(--border);
  box-shadow: var(--shadow-xs);
}
.app-btn--outline:hover:not(:disabled) {
  border-color: var(--primary-border);
  color: var(--primary);
  background: var(--primary-light);
}

.app-btn--ghost {
  background: transparent;
  color: var(--text-secondary);
}
.app-btn--ghost:hover:not(:disabled) {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.app-btn--danger {
  background: var(--danger);
  color: #FFFFFF;
}
.app-btn--danger:hover:not(:disabled) {
  background: #DC2626;
}

.app-btn--success {
  background: var(--success);
  color: #FFFFFF;
}
.app-btn--success:hover:not(:disabled) {
  background: #15803D;
}

.app-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  box-shadow: none;
}

.app-btn__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
