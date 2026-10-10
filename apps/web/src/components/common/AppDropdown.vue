<template>
  <div class="dropdown-wrapper" ref="dropdownRef">
    <div class="dropdown-trigger" @click.stop="toggle">
      <slot name="trigger">
        <button class="icon-trigger-btn" title="Tùy chọn">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="12" cy="5" r="2"/>
            <circle cx="12" cy="12" r="2"/>
            <circle cx="12" cy="19" r="2"/>
          </svg>
        </button>
      </slot>
    </div>

    <Transition name="drop-fade">
      <div v-if="isOpen" :class="['dropdown-menu', `align-${align}`]" @click="closeOnClick && close()">
        <slot />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

withDefaults(
  defineProps<{
    align?: 'left' | 'right'
    closeOnClick?: boolean
  }>(),
  {
    align: 'right',
    closeOnClick: true,
  }
)

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

function toggle() {
  isOpen.value = !isOpen.value
}

function close() {
  isOpen.value = false
}

function handleOutsideClick(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    close()
  }
}

onMounted(() => {
  window.addEventListener('click', handleOutsideClick)
})

onUnmounted(() => {
  window.removeEventListener('click', handleOutsideClick)
})

defineExpose({ close, toggle })
</script>

<style scoped>
.dropdown-wrapper {
  position: relative;
  display: inline-block;
}

.icon-trigger-btn {
  background: transparent;
  color: var(--text-secondary);
  width: 30px;
  height: 30px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.icon-trigger-btn:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
  border-color: var(--border);
}

.dropdown-menu {
  position: absolute;
  top: calc(100% + 4px);
  z-index: 100;
  min-width: 170px;
  background: #FFFFFF;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  box-shadow: var(--shadow-md);
  padding: 4px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dropdown-menu.align-right {
  right: 0;
}

.dropdown-menu.align-left {
  left: 0;
}

/* Transitions */
.drop-fade-enter-active,
.drop-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.drop-fade-enter-from,
.drop-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
