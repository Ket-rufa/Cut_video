<template>
  <aside class="sidebar">
    <!-- Header / Logo -->
    <div class="sidebar__brand">
      <div class="brand-icon">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="6" cy="6" r="3"/>
          <circle cx="6" cy="18" r="3"/>
          <line x1="20" y1="4" x2="8.12" y2="15.88"/>
          <line x1="14.47" y1="14.48" x2="20" y2="20"/>
          <line x1="8.12" y1="8.12" x2="12" y2="12"/>
        </svg>
      </div>
      <div class="brand-text">
        <span class="brand-title">CUT VIDEO MXH</span>
        <span class="brand-subtitle">Smart Video Cutter</span>
      </div>
    </div>

    <!-- Main Navigation Menu -->
    <nav class="sidebar__nav">
      <RouterLink to="/" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
            <polyline points="9 22 9 12 15 12 15 22"/>
          </svg>
        </span>
        <span class="nav-label">Trang chủ</span>
      </RouterLink>

      <RouterLink to="/editor" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/editor') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
            <line x1="7" y1="2" x2="7" y2="22"/>
            <line x1="17" y1="2" x2="17" y2="22"/>
            <line x1="2" y1="12" x2="22" y2="12"/>
            <line x1="2" y1="7" x2="7" y2="7"/>
            <line x1="2" y1="17" x2="7" y2="17"/>
            <line x1="17" y1="17" x2="22" y2="17"/>
            <line x1="17" y1="7" x2="22" y2="7"/>
          </svg>
        </span>
        <span class="nav-label">Cắt video</span>
      </RouterLink>

      <RouterLink to="/saved" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/saved') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="23 7 16 12 23 17 23 7"/>
            <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
          </svg>
        </span>
        <span class="nav-label">Video đã lưu</span>
      </RouterLink>

      <RouterLink to="/library" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/library') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/>
          </svg>
        </span>
        <span class="nav-label">Thư viện</span>
      </RouterLink>

      <RouterLink to="/social" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/social') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="18" cy="5" r="3"/>
            <circle cx="6" cy="12" r="3"/>
            <circle cx="18" cy="19" r="3"/>
            <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/>
            <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/>
          </svg>
        </span>
        <span class="nav-label">Mạng xã hội</span>
      </RouterLink>

      <RouterLink to="/history" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/history') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </span>
        <span class="nav-label">Lịch sử</span>
      </RouterLink>

      <RouterLink to="/settings" class="nav-item" :class="{ 'nav-item--active': isRouteActive('/settings') }">
        <span class="nav-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
        </span>
        <span class="nav-label">Cài đặt</span>
      </RouterLink>
    </nav>

    <!-- BỘ LỌC NHANH -->
    <div class="sidebar__section">
      <div class="section-heading">BỘ LỌC NHANH</div>
      <div class="filter-list">
        <button class="filter-item" :class="{ 'filter-item--active': activeQuickFilter === 'all' }" @click="selectQuickFilter('all')">
          <span class="filter-name">Tất cả</span>
          <span class="filter-count">{{ videoStore.countAll }}</span>
        </button>
        <button class="filter-item" :class="{ 'filter-item--active': activeQuickFilter === 'untagged' }" @click="selectQuickFilter('untagged')">
          <span class="filter-name">Chưa gắn thẻ</span>
          <span class="filter-count">{{ videoStore.countUntagged }}</span>
        </button>
        <button class="filter-item" :class="{ 'filter-item--active': activeQuickFilter === 'favorites' }" @click="selectQuickFilter('favorites')">
          <span class="filter-name">Yêu thích</span>
          <span class="filter-count">{{ videoStore.countFavorites }}</span>
        </button>
        <button class="filter-item" :class="{ 'filter-item--active': activeQuickFilter === 'trash' }" @click="selectQuickFilter('trash')">
          <span class="filter-name">Rác</span>
          <span class="filter-count">{{ videoStore.countTrash }}</span>
        </button>
      </div>
    </div>

    <!-- Cuối Sidebar: Dung lượng & Profile -->
    <div class="sidebar__footer">
      <!-- Storage usage -->
      <div class="storage-box">
        <div class="storage-meta">
          <span class="storage-title">Dung lượng</span>
          <span class="storage-val">{{ videoStore.storageFormatted }} / 50 GB</span>
        </div>
        <div class="progress-track">
          <div class="progress-bar" :style="{ width: videoStore.storagePercent + '%' }"></div>
        </div>
      </div>

      <!-- User Profile -->
      <div class="user-profile">
        <div class="user-avatar">
          <img src="https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=80&q=80" alt="Big Hải" />
          <span class="online-indicator"></span>
        </div>
        <div class="user-info">
          <span class="user-name">Big Hải</span>
          <span class="user-role">Admin</span>
        </div>
        <button class="user-options-btn" title="Tài khoản">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="9 18 15 12 9 6"/>
          </svg>
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useVideoStore } from '@/stores/video.store'

const route = useRoute()
const router = useRouter()
const videoStore = useVideoStore()
const activeQuickFilter = ref('all')

onMounted(() => {
  if (videoStore.videos.length === 0) {
    videoStore.fetchAllVideos()
  }
})

function isRouteActive(path: string): boolean {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

function selectQuickFilter(filter: string) {
  activeQuickFilter.value = filter
  router.push({ path: '/saved', query: { filter } })
}
</script>

<style scoped>
.sidebar {
  width: 236px;
  min-width: 236px;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--border);
  height: 100vh;
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  user-select: none;
  z-index: 20;
}

/* Brand Logo */
.sidebar__brand {
  height: 64px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--border);
}

.brand-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  background: var(--primary);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 6px rgba(22, 119, 255, 0.3);
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-title {
  font-size: 14px;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text-primary);
  line-height: 1.2;
}

.brand-subtitle {
  font-size: 11px;
  color: var(--text-muted);
}

/* Navigation */
.sidebar__nav {
  padding: 16px 12px 10px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.15s ease;
  text-decoration: none;
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  transition: color 0.15s ease;
}

.nav-item:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.nav-item:hover .nav-icon {
  color: var(--text-primary);
}

/* Menu active state: background xanh nhạt, text xanh, icon xanh */
.nav-item--active {
  background: var(--primary-light) !important;
  color: var(--primary) !important;
  font-weight: 600;
}

.nav-item--active .nav-icon {
  color: var(--primary) !important;
}

/* Bộ lọc nhanh */
.sidebar__section {
  padding: 12px 12px;
  border-top: 1px solid var(--border-subtle);
}

.section-heading {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  padding: 0 12px 8px;
  text-transform: uppercase;
}

.filter-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.filter-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  transition: all 0.15s ease;
  text-align: left;
}

.filter-item:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.filter-item--active {
  background: var(--surface-secondary);
  color: var(--primary);
  font-weight: 600;
}

.filter-count {
  font-size: 12px;
  color: var(--text-muted);
  background: var(--surface-secondary);
  padding: 1px 7px;
  border-radius: 10px;
}

.filter-item--active .filter-count {
  background: var(--primary-light);
  color: var(--primary);
  font-weight: 600;
}

/* Footer / Storage & Profile */
.sidebar__footer {
  margin-top: auto;
  padding: 16px 14px;
  border-top: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.storage-box {
  background: var(--surface-secondary);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.storage-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.storage-title {
  color: var(--text-secondary);
  font-weight: 500;
}

.storage-val {
  color: var(--text-primary);
  font-weight: 600;
  font-family: var(--font-sans);
}

.progress-track {
  height: 6px;
  background: #E2E8F0;
  border-radius: 3px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: var(--primary);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px;
}

.user-avatar {
  position: relative;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: visible;
  flex-shrink: 0;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border);
}

.online-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 10px;
  height: 10px;
  background: var(--success);
  border-radius: 50%;
  border: 2px solid #FFFFFF;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  line-height: 1.25;
}

.user-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.user-role {
  font-size: 11px;
  color: var(--text-muted);
}

.user-options-btn {
  background: transparent;
  color: var(--text-muted);
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-options-btn:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}
</style>
