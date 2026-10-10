<template>
  <div class="library-page">
    <!-- Top Bar -->
    <div class="page-top-bar card-base">
      <div class="title-group">
        <h2 class="page-title">Thư viện video & clip</h2>
        <span class="page-sub">Quản lý cây thư mục và phân loại nội dung video</span>
      </div>

      <AppButton variant="primary" size="md" @click="showNewFolderModal = true">
        <template #icon-left>
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
        </template>
        Tạo thư mục mới
      </AppButton>
    </div>

    <!-- 2-Column Split: Folder Tree (Left 280px) & Folder Contents (Right) -->
    <div class="library-layout">
      <!-- Left Folder Tree -->
      <aside class="folder-tree-card card-base">
        <div class="tree-header">
          <h4 class="card-title">Thư mục phân loại</h4>
        </div>

        <div class="folder-tree-list">
          <button
            v-for="folder in libraryStore.folders"
            :key="folder.id"
            :class="['tree-item', { 'tree-item--active': selectedFolderId === folder.id }]"
            @click="selectFolder(folder.id)"
          >
            <span class="folder-icon">{{ folder.icon || '📁' }}</span>
            <span class="folder-name">{{ folder.name }}</span>
            <span class="folder-count-pill">{{ getFolderCount(folder.id) }}</span>
            <button
              v-if="folder.id !== 'f-all'"
              class="folder-del-btn"
              @click.stop="deleteFolder(folder.id)"
              title="Xóa thư mục"
            >
              ×
            </button>
          </button>
        </div>

        <!-- Tags Management List -->
        <div class="tags-header">
          <h4 class="card-title">Quản lý thẻ (Tags)</h4>
        </div>

        <div class="tags-tree-list">
          <span
            v-for="tag in libraryStore.tags"
            :key="tag.id"
            class="tag-pill"
          >
            #{{ tag.name }} ({{ tag.count || 0 }})
            <button class="tag-del-btn" @click="deleteTag(tag.id)">×</button>
          </span>
        </div>
      </aside>

      <!-- Right Folder Content Area -->
      <section class="folder-content-area">
        <!-- Sub-tabs: Video trong thư mục vs Đoạn clip trong thư mục -->
        <div class="content-tabs card-base">
          <div class="tabs-row">
            <button
              :class="['content-tab-btn', { 'content-tab-btn--active': activeTab === 'videos' }]"
              @click="activeTab = 'videos'"
            >
              Video ({{ folderVideos.length }})
            </button>
            <button
              :class="['content-tab-btn', { 'content-tab-btn--active': activeTab === 'clips' }]"
              @click="activeTab = 'clips'"
            >
              Đoạn cắt clips ({{ folderClips.length }})
            </button>
          </div>
        </div>

        <!-- Tab 1: Videos in Folder -->
        <div v-if="activeTab === 'videos'" class="videos-tab-content">
          <div v-if="folderVideos.length > 0" class="videos-grid">
            <VideoCard
              v-for="video in folderVideos"
              :key="video.id"
              :video="video"
              view-mode="grid"
              @open="openVideo"
              @cut="openEditor"
              @toggle-fav="videoStore.toggleFavorite"
              @delete="videoStore.deleteVideo"
            />
          </div>
          <div v-else class="empty-folder-box card-base">
            <span class="empty-icon">📂</span>
            <h4>Chưa có video nào trong thư mục này</h4>
            <p>Hãy chuyển video vào thư mục này từ trang "Video đã lưu" hoặc dán liên kết mới.</p>
          </div>
        </div>

        <!-- Tab 2: Clips in Folder -->
        <div v-else class="clips-tab-content">
          <div v-if="folderClips.length > 0" class="clips-grid">
            <div
              v-for="clip in folderClips"
              :key="clip.id"
              class="library-clip-card card-base"
            >
              <div class="clip-preview-box">
                <img :src="clip.thumbnail || defaultThumb" class="clip-preview-img" alt="clip" />
                <span class="clip-duration">{{ formatDurationSec(clip.duration_ms) }}</span>
              </div>
              <div class="clip-details">
                <h4 class="clip-name">{{ clip.title }}</h4>
                <span class="clip-time">{{ formatTimeShort(clip.start_time_ms) }} - {{ formatTimeShort(clip.end_time_ms) }}</span>
                <div class="clip-actions-row">
                  <AppButton variant="secondary" size="sm" @click="downloadClip(clip)">Tải đoạn</AppButton>
                  <AppButton variant="primary" size="sm" @click="openEditorWithClip(clip)">Xem</AppButton>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="empty-folder-box card-base">
            <span class="empty-icon">✂️</span>
            <h4>Chưa có đoạn cắt nào</h4>
            <p>Mở video trong phòng Cắt Video để tạo các đoạn cắt mới.</p>
          </div>
        </div>
      </section>
    </div>

    <!-- New Folder Modal -->
    <AppModal v-model="showNewFolderModal" title="Tạo thư mục mới">
      <div class="modal-form">
        <label class="form-label">Tên thư mục:</label>
        <input v-model="newFolderName" type="text" class="form-input" placeholder="Ví dụ: Du lịch, Shorts, Review..." />

        <label class="form-label" style="margin-top: 10px;">Biểu tượng:</label>
        <div class="icon-selector">
          <button
            v-for="icon in ['📁', '✈️', '🎙️', '⭐', '🎵', '📢', '🎬', '🔥']"
            :key="icon"
            :class="['icon-choice', { 'icon-choice--active': newFolderIcon === icon }]"
            @click="newFolderIcon = icon"
          >
            {{ icon }}
          </button>
        </div>
      </div>
      <template #footer>
        <AppButton variant="secondary" @click="showNewFolderModal = false">Hủy</AppButton>
        <AppButton variant="primary" :disabled="!newFolderName.trim()" @click="createFolder">
          Tạo thư mục
        </AppButton>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useVideoStore } from '@/stores/video.store'
import { useEditorStore } from '@/stores/editor.store'
import { useLibraryStore } from '@/stores/library.store'
import type { Video, Clip } from '@/types'
import { formatTimeShort, formatDurationSec } from '@/utils/format'
import VideoCard from '@/components/video/VideoCard.vue'
import AppButton from '@/components/common/AppButton.vue'
import AppModal from '@/components/common/AppModal.vue'

const router = useRouter()
const videoStore = useVideoStore()
const editorStore = useEditorStore()
const libraryStore = useLibraryStore()

const selectedFolderId = ref('f-all')
const activeTab = ref<'videos' | 'clips'>('videos')

const defaultThumb = 'https://images.unsplash.com/photo-1528127269322-539801943592?auto=format&fit=crop&w=320&q=80'

// New Folder Modal
const showNewFolderModal = ref(false)
const newFolderName = ref('')
const newFolderIcon = ref('📁')

const folderVideos = computed(() => {
  if (selectedFolderId.value === 'f-all') {
    return videoStore.videos
  }
  return videoStore.videos.filter(v => v.folder_id === selectedFolderId.value)
})

const folderClips = computed(() => {
  return editorStore.clips
})

function getFolderCount(folderId: string): number {
  if (folderId === 'f-all') return videoStore.videos.length
  return videoStore.videos.filter(v => v.folder_id === folderId).length
}

function selectFolder(id: string) {
  selectedFolderId.value = id
}

async function createFolder() {
  if (!newFolderName.value.trim()) return
  await libraryStore.addFolder(newFolderName.value.trim(), '#1677FF', newFolderIcon.value)
  newFolderName.value = ''
  showNewFolderModal.value = false
}

function deleteFolder(id: string) {
  if (confirm('Bạn có chắc chắn muốn xóa thư mục này?')) {
    libraryStore.removeFolder(id)
  }
}

function deleteTag(id: string) {
  libraryStore.removeTag(id)
}

function openVideo(video: Video) {
  editorStore.setVideo(video)
  router.push('/editor')
}

function openEditor(video: Video) {
  editorStore.setVideo(video)
  router.push('/editor')
}

function openEditorWithClip(clip: Clip) {
  editorStore.selectClip(clip)
  router.push('/editor')
}

function downloadClip(clip: Clip) {
  editorStore.downloadSingleClip(clip)
}
</script>

<style scoped>
.library-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.page-top-bar {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
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

/* 2-Column layout */
.library-layout {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 20px;
  align-items: start;
}

/* Left tree card */
.folder-tree-card {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #FFFFFF;
}

.tree-header, .tags-header {
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 8px;
}

.tags-header {
  margin-top: 14px;
}

.folder-tree-list {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background: transparent;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s ease;
  text-align: left;
  position: relative;
}

.tree-item:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.tree-item--active {
  background: var(--primary-light) !important;
  color: var(--primary) !important;
  font-weight: 600;
}

.folder-icon {
  font-size: 15px;
}

.folder-name {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-count-pill {
  font-size: 11px;
  background: var(--surface-secondary);
  color: var(--text-muted);
  padding: 1px 6px;
  border-radius: 10px;
}

.tree-item--active .folder-count-pill {
  background: #FFFFFF;
  color: var(--primary);
  font-weight: 700;
}

.folder-del-btn {
  opacity: 0;
  background: transparent;
  color: var(--text-muted);
  font-size: 14px;
  line-height: 1;
  padding: 2px 4px;
  border-radius: 4px;
  transition: opacity 0.15s;
}

.tree-item:hover .folder-del-btn {
  opacity: 1;
}

.folder-del-btn:hover {
  color: var(--danger);
}

/* Tags list */
.tags-tree-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding-top: 4px;
}

.tag-pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--surface-secondary);
  color: var(--primary);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid var(--border);
}

.tag-del-btn {
  background: transparent;
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
}

.tag-del-btn:hover {
  color: var(--danger);
}

/* Content Area */
.folder-content-area {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.content-tabs {
  padding: 8px 14px;
  background: #FFFFFF;
}

.tabs-row {
  display: flex;
  gap: 8px;
}

.content-tab-btn {
  padding: 6px 14px;
  background: transparent;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  transition: all 0.15s ease;
}

.content-tab-btn:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

.content-tab-btn--active {
  background: var(--primary-light);
  color: var(--primary);
  font-weight: 600;
}

.videos-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.clips-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.library-clip-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #FFFFFF;
}

.clip-preview-box {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000;
}

.clip-preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.clip-duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono);
}

.clip-details {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.clip-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.clip-time {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

.clip-actions-row {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.empty-folder-box {
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

.empty-folder-box h4 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-folder-box p {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* Modal */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.form-input {
  height: 38px;
  padding: 0 12px;
  font-size: 14px;
}

.icon-selector {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.icon-choice {
  width: 38px;
  height: 38px;
  font-size: 18px;
  background: var(--surface-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}

.icon-choice--active {
  border-color: var(--primary);
  background: var(--primary-light);
}

@media (max-width: 900px) {
  .library-layout {
    grid-template-columns: 1fr;
  }
  .videos-grid, .clips-grid {
    grid-template-columns: 1fr;
  }
}
</style>
