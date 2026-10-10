<template>
  <header class="header">
    <!-- Left: Search Bar -->
    <div class="header__search">
      <div class="search-input-wrapper">
        <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input
          type="text"
          class="search-field"
          placeholder="Tìm kiếm video, tiêu đề, thẻ..."
          v-model="videoStore.searchQuery"
          @keydown.meta.k.prevent="focusSearch"
          @keydown.ctrl.k.prevent="focusSearch"
          ref="searchInputRef"
        />
        <span class="search-shortcut">⌘ K</span>
      </div>
    </div>

    <!-- Right: Actions & Add Video -->
    <div class="header__actions">
      <!-- Theme Toggle -->
      <button class="action-btn" :title="isDark ? 'Chuyển sang Giao diện Sáng' : 'Chuyển sang Giao diện Tối'" @click="toggleTheme">
        <svg v-if="!isDark" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
        </svg>
        <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="5"/>
          <line x1="12" y1="1" x2="12" y2="3"/>
          <line x1="12" y1="21" x2="12" y2="23"/>
          <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
          <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
          <line x1="1" y1="12" x2="3" y2="12"/>
          <line x1="21" y1="12" x2="23" y2="12"/>
          <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
          <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
        </svg>
      </button>

      <!-- Notification -->
      <button class="action-btn notification-btn" title="Thông báo">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
        </svg>
        <span class="notification-badge"></span>
      </button>

      <!-- User Avatar -->
      <div class="user-avatar-wrap" title="Big Hải (Admin)">
        <img src="https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=80&q=80" alt="Avatar" />
      </div>

      <!-- Primary CTA: + Thêm video -->
      <AppButton variant="primary" size="md" @click="emit('openAddVideo')">
        <template #icon-left>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
        </template>
        + Thêm video
      </AppButton>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useVideoStore } from '@/stores/video.store'
import AppButton from '@/components/common/AppButton.vue'

const emit = defineEmits<{
  openAddVideo: []
}>()

const videoStore = useVideoStore()
const searchInputRef = ref<HTMLInputElement | null>(null)
const isDark = ref(false)

function focusSearch() {
  searchInputRef.value?.focus()
}

function toggleTheme() {
  isDark.value = !isDark.value
  if (isDark.value) {
    document.documentElement.setAttribute('data-theme', 'dark')
  } else {
    document.documentElement.removeAttribute('data-theme')
  }
}

function handleGlobalKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    focusSearch()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<style scoped>
.header {
  height: 64px;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 15;
}

/* Search bar */
.header__search {
  flex: 1;
  max-width: 440px;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: var(--text-muted);
  pointer-events: none;
}

.search-field {
  width: 100%;
  height: 38px;
  padding: 0 48px 0 36px;
  background: var(--surface-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 13px;
  color: var(--text-primary);
  transition: all 0.15s ease;
}

.search-field:focus {
  background: #FFFFFF;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-light);
}

.search-field::placeholder {
  color: var(--text-muted);
}

.search-shortcut {
  position: absolute;
  right: 10px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 500;
  color: var(--text-muted);
  pointer-events: none;
  font-family: var(--font-mono);
}

/* Actions */
.header__actions {
  display: flex;
  align-items: center;
  gap: 14px;
}

.action-btn {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--border);
  position: relative;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
  border-color: #CBD5E1;
}

.notification-btn {
  position: relative;
}

.notification-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 7px;
  height: 7px;
  background: var(--danger);
  border-radius: 50%;
  border: 1.5px solid #FFFFFF;
}

.user-avatar-wrap {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid var(--border);
  cursor: pointer;
}

.user-avatar-wrap img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
