<template>
  <div class="video-info-card card-base">
    <div class="info-header">
      <div class="header-title-box">
        <h3 class="card-title">Thông tin video</h3>
        <span v-if="translationInfo" class="detected-lang-pill" :title="`Đã dịch từ ${translationInfo.source_language}`">
          🌐 {{ translationInfo.source_language }}
        </span>
      </div>

      <div class="header-actions">
        <!-- AI Translate Button -->
        <button
          class="ai-translate-btn"
          :class="{ 'btn--loading': isTranslating }"
          :disabled="isTranslating"
          @click="handleTranslateMeta"
          title="Tự động nhận diện ngôn ngữ & dịch tiêu đề, mô tả, thẻ sang Tiếng Việt bằng AI"
        >
          <svg v-if="!isTranslating" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <line x1="2" y1="12" x2="22" y2="12"/>
            <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
          </svg>
          <span v-if="isTranslating" class="spinner-sm"></span>
          <span>{{ isTranslating ? 'Đang dịch...' : (translationInfo ? 'Dịch lại' : 'Dịch AI') }}</span>
        </button>

        <button class="edit-btn" @click="isEditing = !isEditing" :title="isEditing ? 'Lưu' : 'Chỉnh sửa'">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
          </svg>
          <span>{{ isEditing ? 'Xong' : 'Sửa' }}</span>
        </button>
      </div>
    </div>

    <!-- Toggle Original vs Vietnamese if translated -->
    <div v-if="translationInfo" class="translation-toggle-bar">
      <span class="toggle-hint">Hiển thị:</span>
      <div class="toggle-btn-group">
        <button
          class="sub-toggle-pill"
          :class="{ 'sub-toggle-pill--active': showingTranslated }"
          @click="toggleTranslatedView(true)"
        >
          Tiếng Việt
        </button>
        <button
          class="sub-toggle-pill"
          :class="{ 'sub-toggle-pill--active': !showingTranslated }"
          @click="toggleTranslatedView(false)"
        >
          Gốc ({{ translationInfo.source_language }})
        </button>
      </div>
    </div>

    <div class="info-body">
      <!-- Thumbnail & Primary Meta -->
      <div class="thumb-meta-row">
        <div class="thumb-box">
          <img :src="video.thumbnail_url" :alt="video.title" class="thumb-img" />
          <span class="source-badge" :class="`source--${video.platform}`">
            {{ platformLabel(video.platform) }}
          </span>
        </div>

        <div class="title-meta-col">
          <input
            v-if="isEditing"
            v-model="editTitle"
            type="text"
            class="title-input"
            @blur="saveTitle"
            @keydown.enter="saveTitle"
          />
          <h4 v-else class="video-title" :title="video.title">{{ video.title }}</h4>

          <div class="meta-specs">
            <span class="spec-item">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
              {{ formatTimeShort(video.duration_ms) }}
            </span>
            <span class="spec-dot">•</span>
            <span class="spec-item">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
                <line x1="8" y1="21" x2="16" y2="21"/>
                <line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
              {{ video.width }}x{{ video.height }}
            </span>
          </div>
        </div>
      </div>

      <!-- Tags list with + button -->
      <div class="tags-section">
        <span class="field-label">Thẻ:</span>
        <div class="tags-list">
          <span v-for="(tag, idx) in tags" :key="idx" class="tag-chip">
            #{{ tag }}
            <button v-if="isEditing" class="tag-remove" @click="removeTag(idx)">×</button>
          </span>

          <!-- Add Tag Input / Button -->
          <div v-if="isAddingTag" class="tag-input-wrap">
            <input
              ref="tagInputRef"
              v-model="newTagText"
              type="text"
              class="tag-input"
              placeholder="nhập thẻ..."
              @keydown.enter="addTag"
              @blur="addTag"
            />
          </div>
          <button v-else class="add-tag-btn" @click="showAddTagInput" title="Thêm thẻ">
            +
          </button>
        </div>
      </div>

      <!-- Description -->
      <div class="desc-section">
        <div class="desc-header">
          <span class="field-label">Mô tả:</span>
          <button
            v-if="!isEditing && isLongDesc"
            class="toggle-desc-btn"
            @click="isDescExpanded = !isDescExpanded"
          >
            <span>{{ isDescExpanded ? 'Rút gọn' : 'Xem thêm' }}</span>
            <svg
              class="toggle-desc-icon"
              :class="{ 'icon--rotated': isDescExpanded }"
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>
        </div>

        <textarea
          v-if="isEditing"
          v-model="editDesc"
          class="desc-textarea"
          rows="3"
          @blur="saveDesc"
        />
        <div
          v-else
          class="desc-text"
          :class="{ 'desc-text--collapsed': !isDescExpanded && isLongDesc }"
        >
          {{ video.description || 'Chưa có mô tả cho video này.' }}
        </div>
      </div>

      <!-- Folder & Created Date Details -->
      <div class="details-footer">
        <div class="detail-row">
          <span class="detail-label">Lưu vào thư mục:</span>
          <span class="detail-value folder-val">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z"/>
            </svg>
            {{ video.folder_name || 'Du lịch' }}
          </span>
        </div>

        <div class="detail-row">
          <span class="detail-label">Ngày tạo:</span>
          <span class="detail-value">{{ formatDate(video.created_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import type { Video, AITranslationResult } from '@/types'
import { formatTimeShort, formatDate, platformLabel } from '@/utils/format'
import { useEditorStore } from '@/stores/editor.store'

const props = defineProps<{ video: Video }>()
const editorStore = useEditorStore()

const isEditing = ref(false)
const editTitle = ref(props.video.title)
const editDesc = ref(props.video.description)
const isAddingTag = ref(false)
const newTagText = ref('')
const tagInputRef = ref<HTMLInputElement | null>(null)

// AI Translation state
const translationInfo = ref<AITranslationResult | null>(null)
const showingTranslated = ref(true)
const isTranslating = computed(() => editorStore.isTranslatingMeta)

// Sync when video changes
watch(
  () => props.video.id,
  () => {
    translationInfo.value = null
    showingTranslated.value = true
    editTitle.value = props.video.title
    editDesc.value = props.video.description
  }
)

async function handleTranslateMeta() {
  const res = await editorStore.translateMeta(true)
  if (res) {
    translationInfo.value = res
    showingTranslated.value = true
    editTitle.value = res.translated_title
    editDesc.value = res.translated_description
  }
}

function toggleTranslatedView(showVi: boolean) {
  if (!translationInfo.value) return
  showingTranslated.value = showVi
  if (showVi) {
    editorStore.updateVideoMeta({
      title: translationInfo.value.translated_title,
      description: translationInfo.value.translated_description,
      tags: translationInfo.value.translated_tags,
    })
    editTitle.value = translationInfo.value.translated_title
    editDesc.value = translationInfo.value.translated_description
  } else {
    editorStore.updateVideoMeta({
      title: translationInfo.value.original_title,
      description: translationInfo.value.original_description,
      tags: translationInfo.value.original_tags,
    })
    editTitle.value = translationInfo.value.original_title
    editDesc.value = translationInfo.value.original_description
  }
}

const isDescExpanded = ref(false)
const isLongDesc = computed(() => {
  const desc = props.video?.description || ''
  return desc.length > 100 || desc.split('\n').length > 3
})

const tags = computed(() => props.video.tags || ['du lịch', 'việt nam', 'vlog'])

function saveTitle() {
  if (editTitle.value.trim()) {
    editorStore.updateVideoMeta({ title: editTitle.value.trim() })
  }
}

function saveDesc() {
  editorStore.updateVideoMeta({ description: editDesc.value.trim() })
}

function showAddTagInput() {
  isAddingTag.value = true
  nextTick(() => {
    tagInputRef.value?.focus()
  })
}

function addTag() {
  const t = newTagText.value.replace(/^#/, '').trim()
  if (t && !tags.value.includes(t)) {
    const updated = [...tags.value, t]
    editorStore.updateVideoMeta({ tags: updated })
  }
  newTagText.value = ''
  isAddingTag.value = false
}

function removeTag(idx: number) {
  const updated = [...tags.value]
  updated.splice(idx, 1)
  editorStore.updateVideoMeta({ tags: updated })
}
</script>

<style scoped>
.video-info-card {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.info-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: 10px;
}

.header-title-box {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.detected-lang-pill {
  font-size: 11px;
  font-weight: 600;
  color: #1D4ED8;
  background: #EFF6FF;
  border: 1px solid #BFDBFE;
  padding: 1px 7px;
  border-radius: 12px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-translate-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  background: #EFF6FF;
  border: 1px solid #BFDBFE;
  border-radius: 6px;
  color: #1D4ED8;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
}

.ai-translate-btn:hover:not(:disabled) {
  background: #DBEAFE;
  border-color: #93C5FD;
  color: #1E40AF;
}

.ai-translate-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner-sm {
  width: 12px;
  height: 12px;
  border: 2px solid #93C5FD;
  border-top-color: #1D4ED8;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.translation-toggle-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  font-size: 12px;
}

.toggle-hint {
  color: #64748B;
  font-size: 11px;
}

.toggle-btn-group {
  display: flex;
  gap: 4px;
}

.sub-toggle-pill {
  border: none;
  background: transparent;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  color: #64748B;
  cursor: pointer;
  transition: all 0.15s ease;
}

.sub-toggle-pill--active {
  background: #FFFFFF;
  color: #1E293B;
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

.edit-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  background: var(--surface-secondary);
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 500;
  transition: all 0.15s ease;
}

.edit-btn:hover {
  background: var(--primary-light);
  color: var(--primary);
  border-color: var(--primary-border);
}

.info-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.thumb-meta-row {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.thumb-box {
  position: relative;
  width: 100px;
  height: 62px;
  border-radius: var(--radius-sm);
  overflow: hidden;
  flex-shrink: 0;
  border: 1px solid var(--border);
}

.thumb-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.source-badge {
  position: absolute;
  bottom: 3px;
  left: 3px;
  padding: 1px 5px;
  font-size: 9px;
  font-weight: 700;
  border-radius: 3px;
  text-transform: uppercase;
  background: rgba(0, 0, 0, 0.7);
  color: #FFFFFF;
}

.source--youtube { background: #DC2626; }
.source--tiktok { background: #0F172A; }
.source--douyin { background: #161823; }
.source--facebook { background: #1D4ED8; }
.source--instagram { background: #DB2777; }
.source--twitter { background: #0F172A; }

.title-meta-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.video-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.title-input {
  width: 100%;
  height: 32px;
  font-size: 13px;
  padding: 0 8px;
}

.meta-specs {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
  font-size: 12px;
}

.spec-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.spec-dot {
  font-size: 10px;
}

/* Tags Section */
.tags-section {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.field-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.tags-list {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.tag-chip {
  background: var(--surface-secondary);
  color: var(--primary);
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 1px solid #D0E2FF;
}

.tag-remove {
  background: transparent;
  color: var(--text-muted);
  font-size: 13px;
  line-height: 1;
  padding: 0;
  cursor: pointer;
}

.tag-remove:hover {
  color: var(--danger);
}

.add-tag-btn {
  width: 22px;
  height: 22px;
  border-radius: 4px;
  background: var(--surface-secondary);
  border: 1px dashed var(--border);
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.15s ease;
}

.add-tag-btn:hover {
  background: var(--primary-light);
  color: var(--primary);
  border-color: var(--primary);
}

.tag-input {
  width: 90px;
  height: 24px;
  padding: 0 6px;
  font-size: 12px;
  border-radius: 4px;
}

/* Description */
.desc-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.desc-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toggle-desc-btn {
  background: transparent;
  border: none;
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  transition: all 0.15s ease;
}

.toggle-desc-btn:hover {
  background: #E8F0FE;
  color: #1765CC;
}

.toggle-desc-icon {
  transition: transform 0.2s ease;
}

.toggle-desc-icon.icon--rotated {
  transform: rotate(180deg);
}

.desc-text {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  background: var(--surface-secondary);
  padding: 8px 10px;
  border-radius: var(--radius-sm);
  white-space: pre-wrap;
  word-break: break-word;
  transition: all 0.25s ease;
}

.desc-text--collapsed {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  max-height: 4.8em;
  position: relative;
}

.desc-textarea {
  width: 100%;
  padding: 8px;
  font-size: 13px;
}

/* Details Footer */
.details-footer {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 8px;
  border-top: 1px solid var(--border-subtle);
  font-size: 12px;
}

.detail-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.detail-label {
  color: var(--text-muted);
}

.detail-value {
  color: var(--text-primary);
  font-weight: 500;
}

.folder-val {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--primary);
  font-weight: 600;
}
</style>
