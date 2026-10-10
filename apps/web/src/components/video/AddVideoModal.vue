<template>
  <AppModal
    :model-value="modelValue"
    @update:model-value="emit('update:modelValue', $event)"
    title="Thêm video mới"
    width="580px"
  >
    <!-- Tabs Header -->
    <div class="modal-tabs">
      <button
        :class="['tab-btn', { 'tab-btn--active': activeTab === 'url' }]"
        @click="activeTab = 'url'"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
        </svg>
        Dán liên kết (URL)
      </button>

      <button
        :class="['tab-btn', { 'tab-btn--active': activeTab === 'upload' }]"
        @click="activeTab = 'upload'"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
        Tải tệp lên
      </button>
    </div>

    <!-- TAB 1: Nhập URL -->
    <div v-if="activeTab === 'url'" class="tab-content">
      <!-- Input State (no analysis yet) -->
      <div v-if="!analysisResult && !activeJob" class="form-group">
        <label class="form-label">Dán link video từ YouTube, TikTok hoặc Douyin</label>
        <div class="url-input-box">
          <input
            v-model="urlInput"
            type="url"
            class="input-field"
            placeholder="Dán link YouTube, TikTok, Douyin hoặc Link trực tiếp CDN (CoCoCut)..."
            :disabled="loading"
            @keydown.enter="handleAnalyze"
          />
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>

        <!-- Supported platforms badges -->
        <div class="platforms-guide">
          <span class="guide-title">Nền tảng hỗ trợ chính thức:</span>
          <div class="platform-chips">
            <span class="chip chip--yt">YouTube (Video & Shorts)</span>
            <span class="chip chip--tt">TikTok (Chuẩn & Link rút gọn)</span>
            <span class="chip chip--dy">Douyin (douyin.com & Link rút gọn)</span>
            <span class="chip chip--generic">Link trực tiếp (CDN / MP4 / CoCoCut)</span>
          </div>
        </div>

        <div class="modal-actions">
          <AppButton variant="secondary" @click="close">Hủy</AppButton>
          <AppButton variant="primary" :loading="loading" @click="handleAnalyze">
            <template #icon-left>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="11" cy="11" r="8"/>
                <line x1="21" y1="21" x2="16.65" y2="16.65"/>
              </svg>
            </template>
            Phân tích video
          </AppButton>
        </div>
      </div>

      <!-- Analysis Preview State -->
      <div v-else-if="analysisResult && !activeJob" class="analysis-preview-card">
        <div class="analysis-header">
          <div class="thumb-wrap">
            <img :src="analysisResult.thumbnail" :alt="analysisResult.title" class="preview-thumb" />
            <span class="duration-badge">{{ formatTimeShort(analysisResult.durationMs) }}</span>
          </div>
          <div class="meta-wrap">
            <span class="platform-tag" :class="`tag--${analysisResult.platform}`">{{ analysisResult.platform.toUpperCase() }}</span>
            <h4 class="video-preview-title" :title="analysisResult.title">{{ analysisResult.title }}</h4>
            <span class="author-name" v-if="analysisResult.author?.name">👤 Kênh / Tác giả: {{ analysisResult.author.name }}</span>
          </div>
        </div>

        <!-- Quality Select -->
        <div class="quality-select-group">
          <label class="form-label">Chọn độ phân giải tải về:</label>
          <select v-model="selectedQuality" class="form-select">
            <option
              v-for="fmt in analysisResult.formats"
              :key="fmt.formatId + fmt.quality"
              :value="fmt.quality"
            >
              {{ fmt.quality.toUpperCase() }} ({{ fmt.container.toUpperCase() }})
              {{ fmt.width && fmt.height ? `• ${fmt.width}x${fmt.height}` : '' }}
              {{ fmt.fileSize ? `• ~${formatBytes(fmt.fileSize)}` : '' }}
            </option>
          </select>
        </div>

        <div v-if="error" class="error-msg">{{ error }}</div>

        <div class="modal-actions">
          <AppButton variant="secondary" @click="resetAnalysis">Chọn URL khác</AppButton>
          <AppButton variant="primary" :loading="loading" @click="handleStartDownload">
            <template #icon-left>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="7 10 12 15 17 10"/>
                <line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
            </template>
            Tải video ({{ selectedQuality.toUpperCase() }})
          </AppButton>
        </div>
      </div>

      <!-- Downloading Progress State -->
      <div v-else-if="activeJob" class="job-progress-card">
        <div class="job-status-row">
          <div class="status-info">
            <span class="status-badge" :class="`badge--${(activeJob.status || '').toLowerCase()}`">{{ getStatusText(activeJob.status) }}</span>
            <span class="job-sub">
              {{ (activeJob.status || '').toUpperCase() === 'COMPLETED' ? 'Video đã được tải và lưu trữ thành công!' : 'Đang xử lý tải video về máy chủ...' }}
            </span>
          </div>
          <span class="progress-percent">{{ activeJob.progress }}%</span>
        </div>

        <!-- Progress Bar -->
        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" :style="{ width: activeJob.progress + '%' }"></div>
        </div>

        <!-- Job Metrics (Bytes, Speed, ETA) -->
        <div class="job-metrics-row" v-if="activeJob.totalBytes > 0 || activeJob.speed > 0">
          <span v-if="activeJob.downloadedBytes > 0">
            {{ formatBytes(activeJob.downloadedBytes) }} / {{ formatBytes(activeJob.totalBytes) }}
          </span>
          <span v-if="activeJob.speed > 0">Tốc độ: {{ formatBytes(activeJob.speed) }}/s ({{ formatMbps(activeJob.speed) }})</span>
          <span v-if="activeJob.eta > 0">Còn lại: ~{{ activeJob.eta }}s</span>
        </div>

        <div v-if="activeJob.error" class="error-msg">{{ activeJob.error }}</div>

        <div class="modal-actions">
          <AppButton
            variant="secondary"
            :disabled="(activeJob.status || '').toUpperCase() === 'COMPLETED' || (activeJob.status || '').toUpperCase() === 'FAILED'"
            @click="handleCancelJob"
          >
            Hủy tải
          </AppButton>
          <AppButton
            v-if="(activeJob.status || '').toUpperCase() === 'COMPLETED'"
            variant="primary"
            @click="finishAndOpenEditor"
          >
            Mở phòng cắt ngay →
          </AppButton>
        </div>
      </div>
    </div>

    <!-- TAB 2: Upload file -->
    <div v-else class="tab-content">
      <div
        class="dropzone"
        :class="{ 'dropzone--dragging': isDragging }"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="handleFileDrop"
        @click="triggerFileInput"
      >
        <input
          ref="fileInputRef"
          type="file"
          accept="video/*"
          style="display: none"
          @change="handleFileChange"
        />

        <div class="dropzone-icon">
          <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="17 8 12 3 7 8"/>
            <line x1="12" y1="3" x2="12" y2="15"/>
          </svg>
        </div>

        <div class="dropzone-text">
          <span class="dropzone-primary">Kéo & thả file video vào đây</span>
          <span class="dropzone-secondary">hoặc <span class="highlight">Chọn video từ máy tính</span></span>
        </div>

        <span class="dropzone-hint">Hỗ trợ MP4, WebM, MOV lên tới 500MB</span>
      </div>

      <!-- Upload Progress -->
      <div v-if="selectedFile" class="file-preview-card">
        <div class="file-info-row">
          <div class="file-name-meta">
            <span class="file-name">{{ selectedFile.name }}</span>
            <span class="file-size">{{ formatBytes(selectedFile.size) }}</span>
          </div>
          <span class="file-progress-percent">{{ uploadProgress }}%</span>
        </div>

        <div class="progress-bar-wrap">
          <div class="progress-bar-fill" :style="{ width: uploadProgress + '%' }"></div>
        </div>
      </div>

      <div class="modal-actions">
        <AppButton variant="secondary" @click="close">Hủy</AppButton>
        <AppButton
          variant="primary"
          :disabled="!selectedFile || loading"
          :loading="loading"
          @click="startUpload"
        >
          Tải lên và Mở Video
        </AppButton>
      </div>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useVideoStore } from '@/stores/video.store'
import { useEditorStore } from '@/stores/editor.store'
import { videoService } from '@/services/video.service'
import { formatBytes, formatMbps, formatTimeShort } from '@/utils/format'
import type { VideoAnalysisResult, DownloadJobProgress } from '@/types'
import AppModal from '@/components/common/AppModal.vue'
import AppButton from '@/components/common/AppButton.vue'

const props = withDefaults(defineProps<{
  modelValue: boolean
  initialUrl?: string
}>(), {
  initialUrl: '',
})
const emit = defineEmits<{ 'update:modelValue': [val: boolean] }>()

const router = useRouter()
const videoStore = useVideoStore()
const editorStore = useEditorStore()

const activeTab = ref<'url' | 'upload'>('url')
const urlInput = ref('')
const loading = ref(false)
const error = ref<string | null>(null)

// Analysis & Job states
const analysisResult = ref<VideoAnalysisResult | null>(null)
const selectedQuality = ref('1080p')
const activeJob = ref<DownloadJobProgress | null>(null)
let pollInterval: ReturnType<typeof setInterval> | null = null

// Upload state
const isDragging = ref(false)
const selectedFile = ref<File | null>(null)
const uploadProgress = ref(0)
const fileInputRef = ref<HTMLInputElement | null>(null)
watch(() => props.initialUrl, (newVal) => {
  if (newVal) {
    urlInput.value = newVal
    activeTab.value = 'url'
    handleAnalyze()
  }
})

watch(() => props.modelValue, (isOpen) => {
  if (isOpen && props.initialUrl && !analysisResult.value && !activeJob.value) {
    urlInput.value = props.initialUrl
    activeTab.value = 'url'
    handleAnalyze()
  }
})

function close() {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
  emit('update:modelValue', false)
  error.value = null
  analysisResult.value = null
  activeJob.value = null
  uploadProgress.value = 0
}

function resetAnalysis() {
  analysisResult.value = null
  activeJob.value = null
  error.value = null
}

function getStatusText(status: string): string {
  const s = (status || '').toUpperCase()
  const map: Record<string, string> = {
    QUEUED: 'Đang xếp hàng',
    ANALYZING: 'Đang phân tích',
    DOWNLOADING: 'Đang tải xuống dữ liệu thật',
    MERGING: 'Đang ghép hình ảnh & âm thanh (FFmpeg)',
    SAVING: 'Đang lưu vào kho video',
    COMPLETED: 'Hoàn tất tải về',
    FAILED: 'Thất bại',
    CANCELLED: 'Đã hủy',
  }
  return map[s] || status
}

// 1. Phân tích video thật
async function handleAnalyze() {
  let url = urlInput.value.trim()
  if (!url) {
    error.value = 'Vui lòng nhập đường link video!'
    return
  }

  // Tự động bóc tách link nếu người dùng dán cả đoạn văn bản chia sẻ
  const urlMatch = url.match(/https?:\/\/[^\s]+/i)
  if (urlMatch) {
    url = urlMatch[0]
    urlInput.value = url
  }

  loading.value = true
  error.value = null
  try {
    const result = await videoService.analyze(url)
    analysisResult.value = result
    if (result.formats && result.formats.length > 0) {
      selectedQuality.value = result.formats[0].quality
    }
  } catch (e: any) {
    error.value = e.message || 'Không thể phân tích video này.'
  } finally {
    loading.value = false
  }
}

// 2. Bắt đầu tải video thật & Poll progress
async function handleStartDownload() {
  if (!analysisResult.value) return
  loading.value = true
  error.value = null

  try {
    const { jobId } = await videoService.createDownload(
      analysisResult.value.analysisId,
      selectedQuality.value,
      analysisResult.value.sourceUrl
    )

    activeJob.value = {
      id: jobId,
      status: 'QUEUED',
      progress: 0,
      downloadedBytes: 0,
      totalBytes: 0,
      speed: 0,
      eta: 0,
    }

    // Start polling every 1 second
    if (pollInterval) clearInterval(pollInterval)
    pollInterval = setInterval(async () => {
      try {
        const job = await videoService.getDownloadJob(jobId)
        activeJob.value = job

        const s = (job.status || '').toUpperCase()
        if (s === 'COMPLETED') {
          if (pollInterval) {
            clearInterval(pollInterval)
            pollInterval = null
          }
          // Refresh videos list
          await videoStore.fetchAllVideos()
          if (job.videoId) {
            try {
              const v = await videoStore.fetchById(job.videoId)
              editorStore.setVideo(v)
            } catch {
              // Ignore if not found immediately
            }
          }
        } else if (s === 'FAILED' || s === 'CANCELLED') {
          if (pollInterval) {
            clearInterval(pollInterval)
            pollInterval = null
          }
          error.value = job.error || 'Tải video thất bại'
        }
      } catch (err: any) {
        // Poll retry
      }
    }, 1000)

  } catch (e: any) {
    error.value = e.message || 'Không thể khởi tạo tiến trình tải.'
  } finally {
    loading.value = false
  }
}

async function handleCancelJob() {
  if (!activeJob.value) return
  try {
    await videoService.cancelDownloadJob(activeJob.value.id)
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
    activeJob.value.status = 'CANCELLED'
  } catch (e: any) {
    error.value = e.message
  }
}

function finishAndOpenEditor() {
  const vid = activeJob.value?.videoId
  close()
  if (vid) {
    router.push(`/editor/${vid}`)
  } else {
    router.push('/editor')
  }
}

// Upload handlers
function triggerFileInput() {
  fileInputRef.value?.click()
}

function handleFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files[0]) {
    selectedFile.value = input.files[0]
  }
}

function handleFileDrop(e: DragEvent) {
  isDragging.value = false
  if (e.dataTransfer?.files && e.dataTransfer.files[0]) {
    selectedFile.value = e.dataTransfer.files[0]
  }
}

async function startUpload() {
  if (!selectedFile.value) return
  loading.value = true
  uploadProgress.value = 0

  const interval = setInterval(() => {
    if (uploadProgress.value < 90) {
      uploadProgress.value += 15
    }
  }, 100)

  try {
    const video = await videoStore.upload(selectedFile.value)
    uploadProgress.value = 100
    clearInterval(interval)
    setTimeout(() => {
      editorStore.setVideo(video)
      close()
      selectedFile.value = null
      uploadProgress.value = 0
      router.push(`/editor`)
    }, 400)
  } catch (e: any) {
    clearInterval(interval)
    error.value = e.message || 'Tải video thất bại.'
  } finally {
    loading.value = false
  }
}

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
})
</script>

<style scoped>
.modal-tabs {
  display: flex;
  background: var(--surface-secondary);
  padding: 4px;
  border-radius: var(--radius-sm);
  margin-bottom: 20px;
  gap: 4px;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 16px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.tab-btn--active {
  background: #FFFFFF;
  color: var(--primary);
  font-weight: 600;
  box-shadow: var(--shadow-xs);
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.form-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.url-input-box {
  display: flex;
}

.input-field {
  width: 100%;
  height: 42px;
  padding: 0 14px;
  font-size: 13px;
}

.error-msg {
  color: var(--danger);
  font-size: 12px;
  line-height: 1.4;
}

.platforms-guide {
  background: var(--surface-secondary);
  padding: 12px 14px;
  border-radius: var(--radius-sm);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.guide-title {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
}

.platform-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border);
  background: #FFFFFF;
}

.chip--yt { color: #DC2626; border-color: #FECACA; background: #FEF2F2; }
.chip--tt { color: #0F172A; }
.chip--dy { color: #161823; border-color: #CBD5E1; background: #F8FAFC; }

/* Analysis Preview */
.analysis-preview-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: var(--surface-secondary);
  padding: 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}

.analysis-header {
  display: flex;
  gap: 14px;
}

.thumb-wrap {
  position: relative;
  width: 140px;
  aspect-ratio: 16/9;
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: #000;
  flex-shrink: 0;
}

.preview-thumb {
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
  padding: 1px 4px;
  border-radius: 2px;
}

.meta-wrap {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.platform-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 3px;
  width: fit-content;
}
.tag--youtube { background: #FEE2E2; color: #DC2626; }
.tag--tiktok { background: #E2E8F0; color: #0F172A; }
.tag--douyin { background: #161823; color: #FFFFFF; }

.video-preview-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.author-name {
  font-size: 12px;
  color: var(--text-muted);
}

.quality-select-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 4px;
}

.form-select {
  height: 38px;
  padding: 0 10px;
  font-size: 13px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: #FFFFFF;
}

/* Job Progress Card */
.job-progress-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: var(--surface-secondary);
  padding: 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}

.job-status-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-badge {
  font-size: 12px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
}
.badge--downloading { background: #DBEAFE; color: #1D4ED8; }
.badge--merging { background: #FEF3C7; color: #D97706; }
.badge--saving { background: #F3E8FF; color: #7C3AED; }
.badge--completed { background: #DCFCE7; color: #16A34A; }
.badge--failed { background: #FEE2E2; color: #DC2626; }
.badge--cancelled { background: #E2E8F0; color: #475569; }

.job-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.progress-percent {
  font-size: 20px;
  font-weight: 700;
  color: var(--primary);
}

.job-metrics-row {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-muted);
}

.progress-bar-wrap {
  height: 8px;
  background: #E2E8F0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 4px;
  transition: width 0.25s ease;
}

/* Dropzone */
.dropzone {
  border: 2px dashed var(--border);
  border-radius: var(--radius-md);
  padding: 32px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  background: var(--surface-hover);
  transition: all 0.15s ease;
}

.dropzone:hover,
.dropzone--dragging {
  border-color: var(--primary);
  background: var(--primary-light);
}

.dropzone-icon {
  color: var(--primary);
}

.dropzone-text {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.dropzone-primary {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.dropzone-secondary {
  font-size: 13px;
  color: var(--text-secondary);
}

.highlight {
  color: var(--primary);
  text-decoration: underline;
  cursor: pointer;
}

.dropzone-hint {
  font-size: 11px;
  color: var(--text-muted);
}

.file-preview-card {
  background: var(--surface-secondary);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.file-info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.file-name-meta {
  display: flex;
  flex-direction: column;
}

.file-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.file-size {
  font-size: 11px;
  color: var(--text-muted);
}

.file-progress-percent {
  font-size: 13px;
  font-weight: 600;
  color: var(--primary);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}
</style>
