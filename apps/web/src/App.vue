<template>
  <AppLayout>
    <RouterView />
  </AppLayout>
  <AppToast ref="toastRef" />
</template>

<script setup lang="ts">
import { ref, provide } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import AppToast from '@/components/common/AppToast.vue'

const toastRef = ref<InstanceType<typeof AppToast> | null>(null)

function showToast(message: string, type: 'success' | 'error' | 'info' = 'success') {
  toastRef.value?.show(message, type)
}

provide('toast', { showToast })
</script>
