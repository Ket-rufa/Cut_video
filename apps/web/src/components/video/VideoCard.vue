<template>
  <div :class="['video-card', `video-card--${viewMode}`, 'card-base']">
    <!-- Thumbnail wrap with duration & source -->
    <div class="card-thumb-wrap" @click="emit('open', video)">
      <img :src="video.thumbnail_url" :alt="video.title" class="card-thumb" />
      <span class="card-duration">{{ formatTimeShort(video.duration_ms) }}</span>
      <span class="card-platform-badge" :class="`platform--${video.platform}`">
        {{ platformLabel(video.platform) }}
      </span>

      <button
        class="fav-btn"
        :class="{ 'fav-btn--active': video.is_favorite }"
        @click.stop="emit('toggleFav', video.id)"
        title="Yêu thích"
      >
        ★
      </button>
    </div>

    <!-- Details -->
    <div class="card-content">
      <div class="title-row">
        <h4 class="card-title" :title="video.title" @click="emit('open', video)">
          {{ video.title }}
        </h4>

        <!-- 3-Dots Action Menu -->
        <AppDropdown align="right">
          <button class="dropdown-item" @click="emit('open', video)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            Mở video
          </button>
          <button class="dropdown-item" @click="emit('cut', video)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="6" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><line x1="20" y1="4" x2="8.12" y2="15.88"/>
            </svg>
            Cắt video
          </button>
          <button class="dropdown-item" @click="emit('rename', video)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
            </svg>
            Đổi tên
          </button>
          <button class="dropdown-item" @click="emit('moveFolder', video)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/>
            </svg>
            Chuyển thư mục
          </button>
          <button class="dropdown-item" @click="emit('download', video)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            Tải xuống
          </button>
          <button class="dropdown-item dropdown-item--danger" @click="emit('delete', video.id)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
            Xóa
          </button>
        </AppDropdown>
      </div>

      <!-- Folder & Tags -->
      <div class="tags-folder-row">
        <span class="folder-badge">
          📁 {{ video.folder_name || 'Mặc định' }}
        </span>
        <div class="card-tags">
          <span v-for="tag in (video.tags || []).slice(0, 2)" :key="tag" class="tag-badge">
            #{{ tag }}
          </span>
          <span v-if="(video.tags || []).length > 2" class="tag-more">
            +{{ (video.tags || []).length - 2 }}
          </span>
        </div>
      </div>

      <!-- Date & Size Footer -->
      <div class="card-footer-row">
        <span class="created-date">{{ formatDate(video.created_at) }}</span>
        <span class="file-size">{{ formatBytes(video.file_size || 0) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Video } from '@/types'
import { formatTimeShort, formatDate, formatBytes, platformLabel } from '@/utils/format'
import AppDropdown from '@/components/common/AppDropdown.vue'

withDefaults(
  defineProps<{
    video: Video
    viewMode?: 'grid' | 'list'
  }>(),
  {
    viewMode: 'grid',
  }
)

const emit = defineEmits<{
  open: [video: Video]
  cut: [video: Video]
  rename: [video: Video]
  moveFolder: [video: Video]
  download: [video: Video]
  delete: [id: string]
  toggleFav: [id: string]
}>()
</script>

<style scoped>
.video-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.15s ease;
  background: #FFFFFF;
}

.video-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: #B5D5FE;
}

/* Grid layout */
.video-card--grid {
  width: 100%;
}

.video-card--grid .card-thumb-wrap {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000;
  cursor: pointer;
  overflow: hidden;
}

.card-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s ease;
}

.video-card:hover .card-thumb {
  transform: scale(1.03);
}

.card-duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.8);
  color: #FFFFFF;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono);
}

.card-platform-badge {
  position: absolute;
  top: 6px;
  left: 6px;
  padding: 2px 6px;
  font-size: 10px;
  font-weight: 700;
  border-radius: 4px;
  text-transform: uppercase;
  color: #FFFFFF;
  background: rgba(0, 0, 0, 0.7);
}

.platform--youtube { background: #DC2626; }
.platform--tiktok { background: #0F172A; }
.platform--douyin { background: #161823; }
.platform--facebook { background: #1D4ED8; }
.platform--instagram { background: #DB2777; }
.platform--twitter { background: #0F172A; }

.fav-btn {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.15s ease;
}

.fav-btn--active {
  color: #FBBF24;
  background: rgba(0, 0, 0, 0.75);
}

/* Card content */
.card-content {
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.card-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  cursor: pointer;
}

.card-title:hover {
  color: var(--primary);
}

.tags-folder-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.folder-badge {
  font-size: 11px;
  color: var(--text-secondary);
  background: var(--surface-secondary);
  padding: 2px 6px;
  border-radius: 4px;
}

.card-tags {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tag-badge {
  font-size: 11px;
  color: var(--primary);
  background: var(--primary-light);
  padding: 2px 6px;
  border-radius: 4px;
}

.tag-more {
  font-size: 10px;
  color: var(--text-muted);
}

.card-footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 6px;
  border-top: 1px solid var(--border-subtle);
  font-size: 11px;
  color: var(--text-muted);
}

/* List view mode */
.video-card--list {
  flex-direction: row;
  align-items: center;
}

.video-card--list .card-thumb-wrap {
  width: 160px;
  min-width: 160px;
  height: 95px;
  aspect-ratio: auto;
}

.video-card--list .card-content {
  flex: 1;
}

/* Dropdown menu items */
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  background: transparent;
  color: var(--text-primary);
  font-size: 13px;
  border-radius: 4px;
  width: 100%;
  text-align: left;
  transition: background 0.15s ease;
}

.dropdown-item:hover {
  background: var(--surface-secondary);
}

.dropdown-item--danger {
  color: var(--danger);
}

.dropdown-item--danger:hover {
  background: var(--danger-light);
}
</style>
