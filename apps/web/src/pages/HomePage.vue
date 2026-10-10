<template>
  <div class="home-dashboard">
    <!-- Hero / Quick Action Banner -->
    <div class="hero-banner card-base">
      <div class="hero-content">
        <h2 class="hero-title">Lưu & Cắt Video Mạng Xã Hội Chuyên Nghiệp</h2>
        <p class="hero-subtitle">
          Dán link từ YouTube, TikTok, Douyin, Facebook, Instagram hoặc tải lên file MP4 để bắt đầu cắt đoạn video theo ý muốn.
        </p>

        <!-- Quick URL Bar -->
        <div class="quick-input-row">
          <div class="input-with-icon">
            <svg class="url-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
              <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
            </svg>
            <input
              v-model="inputUrl"
              type="url"
              class="url-input"
              placeholder="Dán đường link video YouTube, TikTok, Douyin, Facebook, Instagram..."
              @keydown.enter="handleQuickResolve"
              :disabled="loading"
            />
          </div>
          <AppButton variant="primary" size="md" :loading="loading" @click="handleQuickResolve">
            Phân tích & Cắt
          </AppButton>
        </div>
      </div>
    </div>

    <!-- 4 SaaS Metrics Cards -->
    <div class="metrics-grid">
      <div class="metric-card card-base">
        <div class="metric-icon-wrap icon--blue">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="23 7 16 12 23 17 23 7"/>
            <rect x="1" y="5" width="15" height="14" rx="2" ry="2"/>
          </svg>
        </div>
        <div class="metric-info">
          <span class="metric-val">{{ videoStore.countAll }}</span>
          <span class="metric-label">Tổng video đã lưu</span>
        </div>
      </div>

      <div class="metric-card card-base">
        <div class="metric-icon-wrap icon--green">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="6" cy="6" r="3"/>
            <circle cx="6" cy="18" r="3"/>
            <line x1="20" y1="4" x2="8.12" y2="15.88"/>
          </svg>
        </div>
        <div class="metric-info">
          <span class="metric-val">{{ videoStore.totalClipsCount }}</span>
          <span class="metric-label">Đoạn đã cắt thành công</span>
        </div>
      </div>

      <div class="metric-card card-base">
        <div class="metric-icon-wrap icon--orange">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
        </div>
        <div class="metric-info">
          <span class="metric-val">{{ videoStore.storageFormatted }}</span>
          <span class="metric-label">Dung lượng sử dụng ({{ videoStore.storagePercent }}%)</span>
        </div>
      </div>

      <div class="metric-card card-base">
        <div class="metric-icon-wrap icon--purple">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div class="metric-info">
          <span class="metric-val">{{ videoStore.totalHoursSaved }}</span>
          <span class="metric-label">Thời gian biên tập tiết kiệm</span>
        </div>
      </div>
    </div>

    <!-- Continue Editing Current Video Feature Box -->
    <div class="continue-editing-card card-base">
      <div class="continue-thumb-box">
        <img :src="currentVideo.thumbnail_url" alt="Current video" class="continue-thumb" />
        <span class="duration-badge">{{ formatTimeShort(currentVideo.duration_ms) }}</span>
      </div>
      <div class="continue-info">
        <span class="badge-tag">Đang xử lý gần đây</span>
        <h3 class="continue-title">{{ currentVideo.title }}</h3>
        <p class="continue-desc">{{ currentVideo.description }}</p>
        <div class="continue-meta">
          <span>📁 Thư mục: {{ currentVideo.folder_name || 'Du lịch' }}</span>
          <span>•</span>
          <span>✂️ {{ editorStore.clips.length > 0 ? editorStore.clips.length : videoStore.totalClipsCount }} đoạn cắt đã sẵn sàng</span>
        </div>
      </div>
      <div class="continue-action">
        <RouterLink to="/editor" class="continue-btn">
          Mở phòng Cắt Video →
        </RouterLink>
      </div>
    </div>

    <!-- Recent Videos Section -->
    <div class="recent-section">
      <div class="section-header-row">
        <h3 class="section-title">Video gần đây</h3>
        <RouterLink to="/saved" class="view-all-link">Xem tất cả ({{ videoStore.videos.length }}) →</RouterLink>
      </div>

      <div class="recent-grid">
        <VideoCard
          v-for="video in videoStore.videos.slice(0, 4)"
          :key="video.id"
          :video="video"
          view-mode="grid"
          @open="openVideo"
          @cut="openEditor"
          @toggle-fav="videoStore.toggleFavorite"
          @delete="videoStore.deleteVideo"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useVideoStore } from '@/stores/video.store'
import { useEditorStore } from '@/stores/editor.store'
import { formatTimeShort } from '@/utils/format'
import type { Video } from '@/types'
import AppButton from '@/components/common/AppButton.vue'
import VideoCard from '@/components/video/VideoCard.vue'

const router = useRouter()
const videoStore = useVideoStore()
const editorStore = useEditorStore()

const inputUrl = ref('')
const loading = ref(false)

const currentVideo = computed(() => editorStore.currentVideo)

onMounted(() => {
  videoStore.fetchAllVideos()
})

function handleQuickResolve() {
  const url = inputUrl.value.trim()
  if (!url) return
  videoStore.openAddModal(url)
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
.home-dashboard {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Hero Banner */
.hero-banner {
  padding: 32px;
  background: linear-gradient(135deg, #FFFFFF 0%, #F0F7FF 100%);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.hero-content {
  max-width: 680px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.hero-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.25;
}

.hero-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.quick-input-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 8px;
}

.input-with-icon {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}

.url-icon {
  position: absolute;
  left: 14px;
  color: var(--text-muted);
}

.url-input {
  width: 100%;
  height: 42px;
  padding: 0 16px 0 42px;
  font-size: 14px;
  border-radius: var(--radius-sm);
  background: #FFFFFF;
}

/* Metrics Grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.metric-card {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 14px;
  background: #FFFFFF;
}

.metric-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon--blue { background: #E6F4FF; color: var(--primary); }
.icon--green { background: #DCFCE7; color: var(--success); }
.icon--orange { background: #FEF3C7; color: #D97706; }
.icon--purple { background: #F3E8FF; color: #7C3AED; }

.metric-info {
  display: flex;
  flex-direction: column;
}

.metric-val {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.metric-label {
  font-size: 12px;
  color: var(--text-muted);
}

/* Continue Editing Card */
.continue-editing-card {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  gap: 20px;
  background: #FFFFFF;
  border: 1.5px solid #D0E2FF;
}

.continue-thumb-box {
  position: relative;
  width: 140px;
  height: 85px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  flex-shrink: 0;
  background: #000;
}

.continue-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.duration-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 3px;
  font-family: var(--font-mono);
}

.continue-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.badge-tag {
  font-size: 11px;
  font-weight: 600;
  color: var(--primary);
  background: var(--primary-light);
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
}

.continue-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.continue-desc {
  font-size: 13px;
  color: var(--text-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.continue-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-muted);
}

.continue-btn {
  display: inline-flex;
  align-items: center;
  padding: 10px 18px;
  background: var(--primary);
  color: #FFFFFF;
  font-weight: 600;
  font-size: 13px;
  border-radius: var(--radius-sm);
  transition: background 0.15s ease;
  white-space: nowrap;
}

.continue-btn:hover {
  background: var(--primary-hover);
  color: #FFFFFF;
}

/* Recent Videos Section */
.recent-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.view-all-link {
  font-size: 13px;
  font-weight: 600;
}

.recent-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

@media (max-width: 1100px) {
  .metrics-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .recent-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .continue-editing-card {
    flex-direction: column;
    align-items: flex-start;
  }
  .recent-grid {
    grid-template-columns: 1fr;
  }
}
</style>
