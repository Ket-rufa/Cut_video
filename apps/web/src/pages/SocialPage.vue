<template>
  <div class="social-page">
    <!-- Top Bar -->
    <div class="page-top-bar card-base">
      <div class="title-group">
        <h2 class="page-title">Mạng xã hội & Nguồn video</h2>
        <span class="page-sub">Phân loại và quản lý video nhập từ các nền tảng mạng xã hội</span>
      </div>
    </div>

    <!-- Platform Stats Grid -->
    <div class="platforms-summary-grid">
      <div
        v-for="p in platformStats"
        :key="p.id"
        :class="['platform-card', 'card-base', { 'platform-card--active': videoStore.activePlatform === p.id }]"
        @click="videoStore.setPlatformFilter(p.id)"
      >
        <div class="platform-header">
          <span class="p-icon" v-html="p.icon"></span>
          <span class="p-badge" :class="`badge--${p.id}`">{{ p.name }}</span>
        </div>
        <div class="p-count-row">
          <span class="p-count">{{ p.count }}</span>
          <span class="p-unit">video</span>
        </div>
        <span class="p-desc">{{ p.desc }}</span>
      </div>
    </div>

    <!-- Filtered Videos Grid -->
    <div class="section-title-wrap">
      <h3 class="section-title">
        Danh sách video nguồn {{ getPlatformName(videoStore.activePlatform) }} ({{ platformVideos.length }})
      </h3>
    </div>

    <div v-if="platformVideos.length > 0" class="videos-grid">
      <VideoCard
        v-for="video in platformVideos"
        :key="video.id"
        :video="video"
        view-mode="grid"
        @open="openVideo"
        @cut="openEditor"
        @toggle-fav="videoStore.toggleFavorite"
        @delete="videoStore.deleteVideo"
      />
    </div>

    <div v-else class="empty-platform card-base">
      <span class="empty-icon">📺</span>
      <h4>Không có video nào từ nguồn này</h4>
      <p>Nhấp vào "+ Thêm video" ở thanh tiêu đề để nhập link video từ nền tảng này.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useVideoStore } from '@/stores/video.store'
import { useEditorStore } from '@/stores/editor.store'
import type { Video } from '@/types'
import VideoCard from '@/components/video/VideoCard.vue'

const router = useRouter()
const videoStore = useVideoStore()
const editorStore = useEditorStore()

const platformStats = computed(() => [
  {
    id: 'all',
    name: 'Tất cả nguồn',
    count: videoStore.videos.length,
    desc: 'Tổng hợp từ tất cả các mạng xã hội',
    icon: `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="m4.93 4.93 4.24 4.24"/><path d="m14.83 9.17 4.24-4.24"/></svg>`,
  },
  {
    id: 'youtube',
    name: 'YouTube',
    count: videoStore.videos.filter(v => v.platform === 'youtube').length,
    desc: 'Video dài, Vlog du lịch, Shorts',
    icon: `<svg width="22" height="22" viewBox="0 0 24 24" fill="#DC2626"><path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z"/></svg>`,
  },
  {
    id: 'tiktok',
    name: 'TikTok',
    count: videoStore.videos.filter(v => v.platform === 'tiktok').length,
    desc: 'Video dọc ngắn, viral trends',
    icon: `<svg width="20" height="20" viewBox="0 0 24 24" fill="#0F172A"><path d="M12.525.02c1.31-.02 2.61-.01 3.91-.02.08 1.53.63 3.09 1.75 4.17 1.12 1.11 2.7 1.62 4.24 1.79v4.03c-1.44-.05-2.89-.35-4.2-.97-.57-.26-1.1-.59-1.62-1.01-.03 3.4.03 6.8-.03 10.2-.09 1.61-.63 3.22-1.66 4.47-1.57 1.94-4.04 3.04-6.52 2.87-2.73-.13-5.26-1.74-6.42-4.19-1.24-2.55-.86-5.75.92-7.91 1.75-2.17 4.54-3.13 7.27-2.45.02 1.48-.01 2.96.01 4.44-1.35-.39-2.88-.13-3.97.74-1.05.81-1.55 2.22-1.26 3.51.27 1.25 1.4 2.22 2.68 2.37 1.28.16 2.6-.47 3.23-1.57.4-1.01.4-2.16.4-3.26V.02h2.09z"/></svg>`,
  },
  {
    id: 'douyin',
    name: 'Douyin (抖音)',
    count: videoStore.videos.filter(v => v.platform === 'douyin').length,
    desc: 'Video ngắn Douyin Trung Quốc, hot trend',
    icon: `<svg width="20" height="20" viewBox="0 0 24 24" fill="#161823"><path d="M12.525.02c1.31-.02 2.61-.01 3.91-.02.08 1.53.63 3.09 1.75 4.17 1.12 1.11 2.7 1.62 4.24 1.79v4.03c-1.44-.05-2.89-.35-4.2-.97-.57-.26-1.1-.59-1.62-1.01-.03 3.4.03 6.8-.03 10.2-.09 1.61-.63 3.22-1.66 4.47-1.57 1.94-4.04 3.04-6.52 2.87-2.73-.13-5.26-1.74-6.42-4.19-1.24-2.55-.86-5.75.92-7.91 1.75-2.17 4.54-3.13 7.27-2.45.02 1.48-.01 2.96.01 4.44-1.35-.39-2.88-.13-3.97.74-1.05.81-1.55 2.22-1.26 3.51.27 1.25 1.4 2.22 2.68 2.37 1.28.16 2.6-.47 3.23-1.57.4-1.01.4-2.16.4-3.26V.02h2.09z"/></svg>`,
  },
  {
    id: 'facebook',
    name: 'Facebook',
    count: videoStore.videos.filter(v => v.platform === 'facebook').length,
    desc: 'Watch videos, Reels, livestream',
    icon: `<svg width="22" height="22" viewBox="0 0 24 24" fill="#1D4ED8"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/></svg>`,
  },
  {
    id: 'instagram',
    name: 'Instagram',
    count: videoStore.videos.filter(v => v.platform === 'instagram').length,
    desc: 'Reels, stories, bài đăng',
    icon: `<svg width="22" height="22" viewBox="0 0 24 24" fill="#DB2777"><path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069z"/></svg>`,
  },
  {
    id: 'twitter',
    name: 'Twitter (X)',
    count: videoStore.videos.filter(v => v.platform === 'twitter').length,
    desc: 'Tin tức, khoảnh khắc ngắn',
    icon: `<svg width="20" height="20" viewBox="0 0 24 24" fill="#0F172A"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/></svg>`,
  },
])

const platformVideos = computed(() => {
  if (!videoStore.activePlatform || videoStore.activePlatform === 'all') {
    return videoStore.videos
  }
  return videoStore.videos.filter(v => v.platform === videoStore.activePlatform)
})

function getPlatformName(plat: string): string {
  const map: Record<string, string> = {
    all: 'tất cả',
    youtube: 'YouTube',
    tiktok: 'TikTok',
    douyin: 'Douyin',
    facebook: 'Facebook',
    instagram: 'Instagram',
    twitter: 'Twitter (X)',
    generic: 'Khác',
  }
  return map[plat] || plat
}

function openVideo(video: Video) {
  editorStore.setVideo(video)
  router.push('/editor')
}

function openEditor(video: Video) {
  editorStore.setVideo(video)
  router.push('/editor')
}
</script>

<style scoped>
.social-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-top-bar {
  padding: 16px 20px;
  background: #FFFFFF;
}

.title-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.page-sub {
  font-size: 13px;
  color: var(--text-muted);
}

.platforms-summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 14px;
}

.platform-card {
  padding: 16px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  cursor: pointer;
  transition: all 0.15s ease;
  background: #FFFFFF;
}

.platform-card:hover {
  transform: translateY(-2px);
  border-color: var(--primary);
  box-shadow: var(--shadow-md);
}

.platform-card--active {
  border-color: var(--primary) !important;
  background: var(--primary-light) !important;
  box-shadow: 0 0 0 1px var(--primary);
}

.platform-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.p-badge {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
}

.p-count-row {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.p-count {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.p-unit {
  font-size: 11px;
  color: var(--text-muted);
}

.p-desc {
  font-size: 11px;
  color: var(--text-muted);
  line-height: 1.3;
}

.section-title-wrap {
  margin-top: 8px;
}

.videos-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
}

.empty-platform {
  padding: 60px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  background: #FFFFFF;
}

.empty-icon {
  font-size: 36px;
  margin-bottom: 8px;
}

.empty-platform h4 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-platform p {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}

@media (max-width: 1200px) {
  .platforms-summary-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  .videos-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .platforms-summary-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .videos-grid {
    grid-template-columns: 1fr;
  }
}
</style>
