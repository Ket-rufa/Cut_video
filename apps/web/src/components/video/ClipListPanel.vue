<template>
  <div class="clip-list-panel card-base">
    <!-- Panel Header -->
    <div class="panel-header">
      <div class="header-left">
        <h3 class="panel-title">Danh sách đoạn cắt ({{ clips.length }})</h3>
      </div>
      <div class="header-right">
        <button class="sort-btn" @click="toggleSort" :title="sortOrder === 'asc' ? 'Sắp xếp: Tăng dần' : 'Sắp xếp: Giảm dần'">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <polyline points="19 12 12 19 5 12"/>
          </svg>
          {{ sortOrder === 'asc' ? 'Theo thứ tự' : 'Mới nhất' }}
        </button>
      </div>
    </div>

    <!-- Clips Scrollable List -->
    <div class="clips-container">
      <div v-if="clips.length === 0" class="empty-clips">
        <div class="empty-icon">✂️</div>
        <p class="empty-text">Chưa có đoạn cắt nào.</p>
        <span class="empty-sub">Chọn điểm đầu & cuối trên timeline rồi bấm "Cắt và lưu đoạn này".</span>
      </div>

      <div
        v-for="(clip, idx) in sortedClips"
        :key="clip.id"
        :class="['clip-card', { 'clip-card--active': selectedClip?.id === clip.id }]"
        @click="selectClip(clip)"
      >
        <!-- Clip Index Badge -->
        <span class="clip-index">{{ formatIndex(idx + 1) }}</span>

        <!-- Clip Thumbnail Preview -->
        <div class="clip-thumb-wrap">
          <img :src="clip.thumbnail || defaultThumb" :alt="clip.title" class="clip-thumb" />
          <span class="clip-duration-badge">{{ formatDurationSec(clip.duration_ms) }}</span>
        </div>

        <!-- Clip Information -->
        <div class="clip-info">
          <span class="clip-title" :title="clip.title">{{ clip.title }}</span>
          <div class="clip-timestamps">
            <span class="time-range">{{ formatTimeShort(clip.start_time_ms) }} - {{ formatTimeShort(clip.end_time_ms) }}</span>
            <span class="duration-tag">{{ formatDurationSec(clip.duration_ms) }}</span>
            <span v-if="clip.speed && clip.speed !== 1" class="clip-speed-pill">{{ clip.speed }}x</span>
            <span v-if="clip.filter && clip.filter.preset !== 'none'" class="clip-filter-pill">◐</span>
          </div>
        </div>

        <!-- 3-Dots Action Menu -->
        <div class="clip-actions" @click.stop>
          <AppDropdown align="right">
            <button class="dropdown-item" @click="selectClip(clip)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3"/>
              </svg>
              Phát đoạn này
            </button>
            <button class="dropdown-item" @click="downloadClip(clip)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="7 10 12 15 17 10"/>
                <line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
              Tải xuống
            </button>
            <button class="dropdown-item dropdown-item--danger" @click="removeClip(clip.id)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="3 6 5 6 21 6"/>
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
              </svg>
              Xóa đoạn
            </button>
          </AppDropdown>
        </div>
      </div>
    </div>

    <!-- Panel Footer CTA -->
    <div class="panel-footer">
      <AppButton
        variant="primary"
        block
        size="md"
        :disabled="clips.length === 0"
        @click="downloadAllClips"
      >
        <template #icon-left>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
        </template>
        Tải tất cả đoạn cắt ({{ clips.length }})
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Clip } from '@/types'
import { formatTimeShort, formatDurationSec } from '@/utils/format'
import { useEditorStore } from '@/stores/editor.store'
import AppButton from '@/components/common/AppButton.vue'
import AppDropdown from '@/components/common/AppDropdown.vue'

const editorStore = useEditorStore()

const defaultThumb = 'https://images.unsplash.com/photo-1528127269322-539801943592?auto=format&fit=crop&w=320&q=80'
const sortOrder = ref<'asc' | 'desc'>('asc')

const clips = computed(() => editorStore.clips)
const selectedClip = computed(() => editorStore.selectedClip)

const sortedClips = computed(() => {
  const list = [...clips.value]
  if (sortOrder.value === 'desc') {
    return list.reverse()
  }
  return list
})

function toggleSort() {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

function formatIndex(num: number): string {
  return num.toString().padStart(2, '0')
}

function selectClip(clip: Clip) {
  editorStore.selectClip(clip)
}

function removeClip(id: string) {
  editorStore.removeClip(id)
}

function downloadClip(clip: Clip) {
  editorStore.downloadSingleClip(clip)
}

function downloadAllClips() {
  editorStore.downloadAll()
}
</script>

<style scoped>
.clip-list-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  max-height: 520px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-subtle);
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.sort-btn {
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.15s ease;
}

.sort-btn:hover {
  background: var(--surface-secondary);
  color: var(--text-primary);
}

/* Clips scroll container */
.clips-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.empty-clips {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 32px 16px;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 28px;
  margin-bottom: 8px;
}

.empty-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.empty-sub {
  font-size: 12px;
  margin-top: 4px;
  line-height: 1.4;
}

/* Clip item card */
.clip-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  background: #FFFFFF;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.clip-card:hover {
  border-color: #B5D5FE;
  background: #F8FAFC;
}

/* Active clip state: border màu xanh */
.clip-card--active {
  border-color: var(--primary) !important;
  background: #F0F7FF !important;
  box-shadow: 0 0 0 1px var(--primary);
}

.clip-index {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-muted);
  width: 18px;
  text-align: center;
  font-family: var(--font-mono);
}

.clip-card--active .clip-index {
  color: var(--primary);
}

.clip-thumb-wrap {
  position: relative;
  width: 58px;
  height: 38px;
  border-radius: 4px;
  overflow: hidden;
  flex-shrink: 0;
  background: #000;
}

.clip-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.clip-duration-badge {
  position: absolute;
  bottom: 2px;
  right: 2px;
  background: rgba(0, 0, 0, 0.75);
  color: #FFFFFF;
  font-size: 9px;
  font-weight: 600;
  padding: 1px 3px;
  border-radius: 2px;
  font-family: var(--font-mono);
}

.clip-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.clip-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.clip-timestamps {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
}

.time-range {
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

.duration-tag {
  color: var(--primary);
  font-weight: 600;
}

.clip-speed-pill {
  background: #EFF6FF;
  color: var(--primary);
  font-size: 10px;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 3px;
  font-family: var(--font-mono);
}

.clip-filter-pill {
  color: #D97706;
  font-size: 11px;
}

.clip-actions {
  display: flex;
  align-items: center;
}

/* Dropdown items */
.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
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

/* Panel Footer */
.panel-footer {
  padding: 12px 14px;
  border-top: 1px solid var(--border-subtle);
  background: var(--surface-hover);
}
</style>
