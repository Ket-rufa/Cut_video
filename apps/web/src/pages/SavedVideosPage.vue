<template>
  <div class="saved-videos-page">
    <!-- Header Controls: Title, Search, Filters, Sort, Grid/List -->
    <div class="page-top-bar card-base">
      <div class="title-filter-left">
        <h2 class="page-title">Video đã lưu ({{ filteredVideos.length }})</h2>
      </div>

      <!-- Filters & Actions -->
      <div class="filter-actions-right">
        <!-- Search -->
        <div class="search-box">
          <svg class="search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="videoStore.searchQuery"
            type="text"
            class="search-input"
            placeholder="Tìm theo tiêu đề, tag..."
          />
        </div>

        <!-- Filter Folder -->
        <select v-model="videoStore.selectedFolder" class="filter-select" title="Lọc theo thư mục">
          <option value="all">Tất cả thư mục</option>
          <option v-for="folder in libraryStore.folders" :key="folder.id" :value="folder.id">
            {{ folder.name }} ({{ folder.count || 0 }})
          </option>
        </select>

        <!-- Filter Tag -->
        <select v-model="videoStore.selectedTag" class="filter-select" title="Lọc theo thẻ">
          <option value="all">
            Tất cả thẻ{{ videoStore.availableTags.length > 0 ? ` (${videoStore.availableTags.length})` : '' }}
          </option>
          <option v-if="videoStore.countUntagged > 0" value="untagged">
            Chưa gắn thẻ ({{ videoStore.countUntagged }})
          </option>
          <option v-for="tag in videoStore.availableTags" :key="tag.name" :value="tag.name">
            #{{ tag.name }} ({{ tag.count }})
          </option>
        </select>

        <!-- Sort -->
        <select v-model="videoStore.sortBy" class="filter-select" title="Sắp xếp">
          <option value="newest">Mới nhất</option>
          <option value="oldest">Cũ nhất</option>
          <option value="duration">Thời lượng dài nhất</option>
          <option value="title">Tên (A-Z)</option>
        </select>

        <!-- Grid/List Toggle -->
        <div class="view-toggle">
          <button
            :class="['toggle-btn', { 'toggle-btn--active': videoStore.viewMode === 'grid' }]"
            @click="videoStore.viewMode = 'grid'"
            title="Dạng lưới"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="7" height="7"/>
              <rect x="14" y="3" width="7" height="7"/>
              <rect x="14" y="14" width="7" height="7"/>
              <rect x="3" y="14" width="7" height="7"/>
            </svg>
          </button>
          <button
            :class="['toggle-btn', { 'toggle-btn--active': videoStore.viewMode === 'list' }]"
            @click="videoStore.viewMode = 'list'"
            title="Dạng danh sách"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Videos Grid / List Display -->
    <div v-if="filteredVideos.length > 0" :class="['videos-container', `videos-container--${videoStore.viewMode}`]">
      <VideoCard
        v-for="video in filteredVideos"
        :key="video.id"
        :video="video"
        :view-mode="videoStore.viewMode"
        @open="openVideo"
        @cut="cutVideo"
        @rename="promptRename"
        @move-folder="promptMoveFolder"
        @download="downloadVideo"
        @delete="deleteVideo"
        @toggle-fav="videoStore.toggleFavorite"
      />
    </div>

    <!-- Empty State -->
    <div v-else class="empty-state-box card-base">
      <span class="empty-icon">📁</span>
      <h4 class="empty-title">Không tìm thấy video nào</h4>
      <p class="empty-desc">Thử thay đổi bộ lọc hoặc thêm video mới từ thanh tác vụ phía trên.</p>
    </div>

    <!-- Rename Modal -->
    <AppModal v-model="showRenameModal" title="Đổi tên video">
      <div class="modal-form">
        <label class="form-label">Tên video mới:</label>
        <input v-model="newTitleInput" type="text" class="form-input" @keydown.enter="saveRename" />
      </div>
      <template #footer>
        <AppButton variant="secondary" @click="showRenameModal = false">Hủy</AppButton>
        <AppButton variant="primary" @click="saveRename">Lưu thay đổi</AppButton>
      </template>
    </AppModal>

    <!-- Move Folder Modal -->
    <AppModal v-model="showMoveFolderModal" title="Chuyển thư mục">
      <div class="modal-form">
        <label class="form-label">Chọn thư mục đích:</label>
        <select v-model="selectedTargetFolder" class="form-select">
          <option v-for="f in libraryStore.folders" :key="f.id" :value="f">
            {{ f.name }}
          </option>
        </select>
      </div>
      <template #footer>
        <AppButton variant="secondary" @click="showMoveFolderModal = false">Hủy</AppButton>
        <AppButton variant="primary" @click="saveMoveFolder">Chuyển</AppButton>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useVideoStore } from '@/stores/video.store'
import { useEditorStore } from '@/stores/editor.store'
import { useLibraryStore } from '@/stores/library.store'
import type { Video, Folder } from '@/types'
import VideoCard from '@/components/video/VideoCard.vue'
import AppButton from '@/components/common/AppButton.vue'
import AppModal from '@/components/common/AppModal.vue'

const router = useRouter()
const videoStore = useVideoStore()
const editorStore = useEditorStore()
const libraryStore = useLibraryStore()

const filteredVideos = computed(() => videoStore.filteredVideos)

onMounted(() => {
  videoStore.fetchAllVideos()
  libraryStore.fetchAll()
})

// Modals
const showRenameModal = ref(false)
const targetVideoForRename = ref<Video | null>(null)
const newTitleInput = ref('')

const showMoveFolderModal = ref(false)
const targetVideoForMove = ref<Video | null>(null)
const selectedTargetFolder = ref<Folder | null>(null)

function openVideo(video: Video) {
  editorStore.setVideo(video)
  router.push('/editor')
}

function cutVideo(video: Video) {
  editorStore.setVideo(video)
  router.push('/editor')
}

function promptRename(video: Video) {
  targetVideoForRename.value = video
  newTitleInput.value = video.title
  showRenameModal.value = true
}

function saveRename() {
  if (targetVideoForRename.value && newTitleInput.value.trim()) {
    videoStore.updateVideo(targetVideoForRename.value.id, { title: newTitleInput.value.trim() })
  }
  showRenameModal.value = false
}

function promptMoveFolder(video: Video) {
  targetVideoForMove.value = video
  selectedTargetFolder.value = libraryStore.folders.find(f => f.id === video.folder_id) || libraryStore.folders[0]
  showMoveFolderModal.value = true
}

function saveMoveFolder() {
  if (targetVideoForMove.value && selectedTargetFolder.value) {
    videoStore.updateVideo(targetVideoForMove.value.id, {
      folder_id: selectedTargetFolder.value.id,
      folder_name: selectedTargetFolder.value.name,
    })
  }
  showMoveFolderModal.value = false
}

function downloadVideo(video: Video) {
  const link = document.createElement('a')
  link.href = video.source_path || 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4'
  link.download = `${video.title.replace(/\s+/g, '_')}.mp4`
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function deleteVideo(id: string) {
  if (confirm('Bạn có chắc chắn muốn xóa video này khỏi danh sách?')) {
    videoStore.deleteVideo(id)
  }
}
</script>

<style scoped>
.saved-videos-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Top bar */
.page-top-bar {
  padding: 16px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
  background: #FFFFFF;
}

.title-filter-left {
  display: flex;
  align-items: center;
}

.filter-actions-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 10px;
  color: var(--text-muted);
}

.search-input {
  height: 36px;
  padding: 0 12px 0 32px;
  font-size: 13px;
  width: 200px;
  background: var(--surface-secondary);
}

.filter-select {
  height: 36px;
  padding: 0 10px;
  font-size: 13px;
  background: var(--surface-secondary);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

/* View toggle buttons */
.view-toggle {
  display: flex;
  background: var(--surface-secondary);
  padding: 2px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.toggle-btn {
  width: 32px;
  height: 30px;
  border-radius: 4px;
  background: transparent;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.toggle-btn--active {
  background: #FFFFFF;
  color: var(--primary);
  box-shadow: var(--shadow-xs);
}

/* Videos display */
.videos-container--grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 18px;
}

.videos-container--list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Empty State */
.empty-state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 60px 20px;
  background: #FFFFFF;
}

.empty-icon {
  font-size: 40px;
  margin-bottom: 12px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-desc {
  font-size: 13px;
  color: var(--text-muted);
  margin-top: 4px;
}

/* Modal form elements */
.modal-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.form-input, .form-select {
  height: 38px;
  padding: 0 12px;
  font-size: 14px;
  width: 100%;
}

@media (max-width: 1200px) {
  .videos-container--grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 900px) {
  .videos-container--grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .videos-container--grid {
    grid-template-columns: 1fr;
  }
}
</style>
