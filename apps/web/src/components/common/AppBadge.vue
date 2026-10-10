<template>
  <span :class="['badge', `badge--${variant}`, `badge--${size}`]">
    <slot name="icon" />
    <slot />
  </span>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: 'primary' | 'success' | 'warning' | 'danger' | 'neutral' | 'youtube' | 'tiktok' | 'douyin' | 'facebook' | 'instagram' | 'twitter'
    size?: 'sm' | 'md'
  }>(),
  {
    variant: 'neutral',
    size: 'sm',
  }
)
</script>

<style scoped>
.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-weight: 500;
  border-radius: 4px;
  white-space: nowrap;
  line-height: 1.2;
}

.badge--sm {
  padding: 3px 7px;
  font-size: 11px;
}

.badge--md {
  padding: 4px 9px;
  font-size: 12px;
}

/* Variants */
.badge--primary {
  background: var(--primary-light);
  color: var(--primary);
  border: 1px solid var(--primary-border);
}

.badge--success {
  background: var(--success-light);
  color: var(--success);
}

.badge--warning {
  background: var(--warning-light);
  color: #B45309;
}

.badge--danger {
  background: var(--danger-light);
  color: var(--danger);
}

.badge--neutral {
  background: var(--surface-secondary);
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

/* Social Badges */
.badge--youtube {
  background: #FEE2E2;
  color: #DC2626;
  border: 1px solid #FECACA;
}

.badge--tiktok {
  background: #F1F5F9;
  color: #0F172A;
  border: 1px solid #CBD5E1;
}

.badge--douyin {
  background: #F8FAFC;
  color: #161823;
  border: 1px solid #CBD5E1;
}

.badge--facebook {
  background: #EFF6FF;
  color: #1D4ED8;
  border: 1px solid #BFDBFE;
}

.badge--instagram {
  background: #FDF2F8;
  color: #DB2777;
  border: 1px solid #FBCFE8;
}

.badge--twitter {
  background: #F8FAFC;
  color: #0F172A;
  border: 1px solid #E2E8F0;
}
</style>
