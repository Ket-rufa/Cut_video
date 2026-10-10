<template>
  <div class="subtitle-panel">
    <!-- Header with status and actions -->
    <div class="panel-header">
      <div class="header-left">
        <div class="panel-icon">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            <path d="M8 9h8"/>
            <path d="M8 13h5"/>
          </svg>
        </div>
        <div>
          <h4 class="panel-title">Phụ đề AI Gemini 3.5 Flash-Lite</h4>
          <p class="panel-subtitle">Tự động nghe âm thanh, chuyển ngữ và đồng bộ thời gian chuẩn xác</p>
        </div>
      </div>

      <div class="header-actions">
        <!-- Toggle CC Display on Video Screen -->
        <label class="cc-switch" title="Bật/Tắt hiển thị phụ đề trên màn hình video">
          <input type="checkbox" :checked="showSubtitles" @change="toggleShowSubtitles" />
          <span class="switch-slider"></span>
          <span class="switch-label">Hiển thị trên video (CC)</span>
        </label>

        <!-- Download SRT Button if has subtitles -->
        <button
          v-if="hasSubtitles"
          class="btn-sub btn-secondary"
          @click="downloadSrt"
          title="Tải về file phụ đề định dạng .SRT"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          Tải .SRT
        </button>
      </div>
    </div>

    <!-- Generating State (Loader) -->
    <div v-if="isGenerating" class="generating-box">
      <div class="spinner-gemini">
        <div class="pulse-ring"></div>
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#8B5CF6" stroke-width="2">
          <path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/>
        </svg>
      </div>
      <div class="generating-text">
        <h5>Gemini 1.5 Flash đang phân tích giọng nói & dịch thuật...</h5>
        <p>Gemini 3.5 Flash-Lite đang nghe, chia từng câu thoại và tạo phụ đề tiếng Việt đồng bộ. Video dài có thể mất vài phút.</p>
      </div>
    </div>

    <!-- Configuration / Generate Controls -->
    <div v-else-if="!hasSubtitles || isExpandedSettings" class="generate-controls-card">
      <div class="controls-row">
        <!-- Target Language -->
        <div class="control-item">
          <label class="control-label">Ngôn ngữ đích</label>
          <div class="control-input-wrap">
            <span class="flag-icon">🇻🇳</span>
            <input type="text" class="input-readonly" value="Tiếng Việt (vi)" readonly />
          </div>
        </div>

        <!-- Scope Selection -->
        <div class="control-item">
          <label class="control-label">Phạm vi video</label>
          <div class="scope-radio-group">
            <label class="radio-label" :class="{ 'radio-label--active': scope === 'clip' }">
              <input type="radio" v-model="scope" value="clip" />
              <span>Đoạn chọn ({{ formatTimeShort(currentClipStart) }} - {{ formatTimeShort(currentClipEnd) }})</span>
            </label>
            <label class="radio-label" :class="{ 'radio-label--active': scope === 'full' }">
              <input type="radio" v-model="scope" value="full" />
              <span>Toàn bộ video ({{ formatTimeShort(totalDurationMs) }})</span>
            </label>
          </div>
        </div>
      </div>

      <!-- Generate Trigger Button -->
      <div class="generate-btn-row">
        <button class="btn-generate-ai" @click="handleGenerate" :disabled="isGenerating">
          <span class="ai-sparkle">✨</span>
          <span>Bắt đầu tạo phụ đề với Gemini AI</span>
        </button>
        <button
          v-if="hasSubtitles"
          class="btn-cancel-settings"
          @click="isExpandedSettings = false"
        >
          Đóng cài đặt
        </button>
      </div>

      <div class="note-box">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#6B7280" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="16" x2="12" y2="12"/>
          <line x1="12" y1="8" x2="12.01" y2="8"/>
        </svg>
        <span>
          Gemini 1.5 Flash tự động phát hiện ngôn ngữ gốc (Tiếng Trung, Anh, v.v.), nhận diện giọng đọc thoại và đồng bộ thời gian từng câu sang Tiếng Việt.
        </span>
      </div>
    </div>

    <!-- Subtitle Cues Timeline List (When available) -->
    <div v-if="hasSubtitles && !isGenerating" class="subtitles-result-area">
      <!-- AI Dubbing / Voiceover Banner -->
      <div class="dubbing-banner-card">
        <div class="dubbing-banner-left">
          <div class="dubbing-icon">🎙️</div>
          <div>
            <div class="dubbing-title-row">
              <h5 class="dubbing-heading">Lồng tiếng AI (TTS Khớp nhịp câu nói)</h5>
              <span v-if="hasDubbedAudio" class="dubbed-tag">Đã lồng tiếng</span>
            </div>
            <p class="dubbing-desc">
              Chuyển phụ đề tiếng Việt thành giọng đọc AI và co giãn tốc độ (tempo sync) khớp chuẩn xác từng câu thoại.
            </p>
          </div>
        </div>

        <div class="dubbing-banner-actions">
          <div v-if="isGeneratingDubbing" class="dubbing-loading-badge">
            <span class="dub-spinner"></span>
            <span>Đang tạo giọng đọc & căn nhịp thoại...</span>
          </div>

          <template v-else>
            <!-- Voice selection & preview -->
            <div class="voice-selector-box">
              <span class="voice-label">Giọng đọc:</span>
              <select v-model="selectedVoiceId" class="voice-select" title="Chọn giọng đọc AI cho video">
                <option v-for="v in voices" :key="v.id" :value="v.id">
                  {{ v.icon }} {{ v.name }}
                </option>
              </select>
              <button
                type="button"
                class="btn-preview-voice"
                :class="{ 'btn-preview-voice--active': isPlayingVoicePreview }"
                @click="togglePreviewVoice"
                title="Bấm để nghe thử mẫu giọng này"
              >
                <span v-if="isPlayingVoicePreview">⏹️ Dừng</span>
                <span v-else>▶️ Nghe thử</span>
              </button>
            </div>

            <div class="dubbing-scope-select">
              <select v-model="dubbingScope" class="scope-select">
                <option value="clip">Đoạn chọn ({{ formatTimeShort(currentClipStart) }} - {{ formatTimeShort(currentClipEnd) }})</option>
                <option value="full">Toàn bộ video ({{ formatTimeShort(totalDurationMs) }})</option>
              </select>
            </div>

            <button
              class="btn-generate-dub"
              @click="handleGenerateDubbing"
              title="Chuyển văn bản thành giọng đọc và tự động khớp nhịp thoại"
            >
              <span>🎙️</span>
              <span>{{ hasDubbedAudio ? 'Lồng tiếng lại' : 'Tạo lồng tiếng AI' }}</span>
            </button>

            <template v-if="hasDubbedAudio">
              <button
                class="btn-sub btn-voice-toggle"
                :class="{ 'btn-voice-toggle--active': isPlayingDubbedAudio }"
                @click="togglePlayDubbedAudio"
                title="Bật/Tắt phát giọng lồng tiếng kèm video"
              >
                <span>{{ isPlayingDubbedAudio ? '🔊 Đang bật lồng tiếng' : '🔈 Bật lồng tiếng' }}</span>
              </button>

              <button
                class="btn-sub btn-secondary"
                @click="downloadDubbedAudio"
                title="Tải về file MP3 lồng tiếng độc lập"
              >
                Tải Audio (.mp3)
              </button>

              <button
                v-if="dubbing?.video_url"
                class="btn-sub btn-secondary"
                @click="downloadDubbedVideo"
                title="Tải về video đã ghép lồng tiếng (.mp4)"
              >
                Tải Video (.mp4)
              </button>
            </template>
          </template>
        </div>
      </div>

      <div class="result-toolbar">
        <div class="cue-count-badge">
          <span>{{ cues.length }} câu phụ đề</span>
        </div>

        <div class="toolbar-right">
          <input
            type="text"
            v-model="searchQuery"
            placeholder="Tìm kiếm câu thoại..."
            class="cue-search-input"
          />
          <button
            class="btn-sub btn-re-generate"
            @click="handleReGenerateDirect"
            :disabled="isGenerating"
            title="Bấm để AI Gemini tạo lại toàn bộ phụ đề cho video"
          >
            <span v-if="isGenerating" class="generating-inline-spinner">⏳ Đang tạo...</span>
            <template v-else>
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="1 4 1 10 7 10"/>
                <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
              </svg>
              <span>Tạo lại phụ đề</span>
            </template>
          </button>
          <button
            class="btn-sub btn-settings-gear"
            @click="isExpandedSettings = !isExpandedSettings"
            :title="isExpandedSettings ? 'Đóng tùy chọn' : 'Tùy chỉnh phạm vi video'"
          >
            ⚙️
          </button>
        </div>
      </div>

      <!-- Cue List -->
      <div class="cues-list" ref="cuesListRef">
        <div
          v-for="cue in filteredCues"
          :key="cue.index"
          class="cue-item"
          :class="{ 'cue-item--active': isCueActive(cue) }"
          @click="seekToCue(cue)"
        >
          <div class="cue-time">
            <span class="time-badge">{{ formatTimeShort(cue.start_ms) }}</span>
            <span class="time-arrow">→</span>
            <span class="time-badge">{{ formatTimeShort(cue.end_ms) }}</span>
          </div>

          <div class="cue-content">
            <template v-if="editingCueIndex === cue.index">
              <div class="cue-edit-wrap" @click.stop>
                <input
                  v-model="editingCueText"
                  class="cue-edit-input"
                  @keydown.enter.prevent="saveEditCue(cue)"
                  @keydown.esc="cancelEditCue"
                  ref="cueInputRef"
                />
                <div class="cue-edit-btns">
                  <button class="btn-cue-save" @click.stop="saveEditCue(cue)" title="Lưu">✓</button>
                  <button class="btn-cue-cancel" @click.stop="cancelEditCue" title="Hủy">✕</button>
                </div>
              </div>
            </template>
            <p
              v-else
              class="cue-text"
              @dblclick.stop="startEditCue(cue)"
              :title="'Nhấp đúp hoặc bấm biểu tượng bút để chỉnh sửa lời thoại'"
            >
              {{ cue.text }}
            </p>
          </div>

          <div class="cue-actions">
            <!-- Edit cue button -->
            <button
              class="cue-action-icon-btn"
              @click.stop="startEditCue(cue)"
              title="Sửa nội dung câu này"
            >
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
              </svg>
            </button>

            <!-- Translate button if contains foreign text -->
            <button
              v-if="hasForeignText(cue.text)"
              class="cue-translate-btn"
              @click.stop="translateCue(cue)"
              :disabled="translatingCueIndex === cue.index"
              title="Dịch câu này sang Tiếng Việt"
            >
              <span v-if="translatingCueIndex === cue.index" class="cue-translating-txt">Đang dịch...</span>
              <span v-else>🌐 Dịch</span>
            </button>

            <!-- TTS Listen button -->
            <button
              class="cue-tts-btn"
              :class="{ 'cue-tts-btn--active': playingCueIndex === cue.index }"
              @click.stop="playCueTTS(cue)"
              :title="'Nghe thử giọng đọc câu: ' + cue.text"
            >
              <span v-if="playingCueIndex === cue.index" class="tts-wave-txt">🔊 Đang phát...</span>
              <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
                <path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/>
              </svg>
            </button>

            <!-- Jump to position button -->
            <button
              class="cue-play-btn"
              @click.stop="seekToCue(cue)"
              title="Nhảy tới đoạn này"
            >
              <svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5 3 19 12 5 21 5 3"/>
              </svg>
            </button>
          </div>
        </div>

        <div v-if="filteredCues.length === 0" class="empty-search">
          Không tìm thấy phụ đề phù hợp với "{{ searchQuery }}"
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch, onMounted } from 'vue'
import { useEditorStore } from '@/stores/editor.store'
import { aiService } from '@/services/ai.service'
import { formatTimeShort } from '@/utils/format'
import type { SubtitleCue } from '@/types'

const editorStore = useEditorStore()

const scope = ref<'clip' | 'full'>('full')
const dubbingScope = ref<'clip' | 'full'>('full')
const searchQuery = ref('')
const isExpandedSettings = ref(false)
const cuesListRef = ref<HTMLDivElement | null>(null)

// Inline edit state
const editingCueIndex = ref<number | null>(null)
const editingCueText = ref('')
const cueInputRef = ref<HTMLInputElement | null>(null)
const translatingCueIndex = ref<number | null>(null)

function hasForeignText(text: string): boolean {
  return /[\u4e00-\u9fa5\u0400-\u04FF\u3040-\u30ff]/.test(text)
}

function startEditCue(cue: SubtitleCue) {
  editingCueIndex.value = cue.index
  editingCueText.value = cue.text
  nextTick(() => {
    cueInputRef.value?.focus()
  })
}

async function saveEditCue(cue: SubtitleCue) {
  if (editingCueText.value.trim()) {
    await editorStore.updateCueText(cue.index, editingCueText.value.trim())
  }
  editingCueIndex.value = null
}

function cancelEditCue() {
  editingCueIndex.value = null
  editingCueText.value = ''
}

async function translateCue(cue: SubtitleCue) {
  translatingCueIndex.value = cue.index
  try {
    await editorStore.translateSingleCue(cue.index)
  } finally {
    translatingCueIndex.value = null
  }
}

const isGenerating = computed(() => editorStore.isGeneratingSubtitles)
const subtitles = computed(() => editorStore.subtitles)
const hasSubtitles = computed(() => !!subtitles.value && subtitles.value.cues && subtitles.value.cues.length > 0)
const showSubtitles = computed(() => editorStore.showSubtitles)
const cues = computed(() => subtitles.value?.cues || [])

// Dubbing state
const dubbing = computed(() => editorStore.dubbing)
const hasDubbedAudio = computed(() => !!dubbing.value && !!dubbing.value.audio_url)
const isGeneratingDubbing = computed(() => editorStore.isGeneratingDubbing)
const isPlayingDubbedAudio = computed(() => editorStore.isPlayingDubbedAudio)

// Voice selection & preview
const voices = computed(() => {
  if (editorStore.availableVoices && editorStore.availableVoices.length > 0) {
    return editorStore.availableVoices
  }
  return [
    { id: 'male_deep', name: 'Nam Trầm Ấm (Review Phim)', gender: 'male', tag: 'Gợi ý', icon: '👨' },
    { id: 'female_standard', name: 'Nữ Tiêu Chuẩn (Google)', gender: 'female', tag: 'Tự nhiên', icon: '👩' },
    { id: 'female_young', name: 'Nữ Trẻ Trung (Vlog / Ngọt Ngào)', gender: 'female', tag: 'Ngọt ngào', icon: '👧' },
    { id: 'male_news', name: 'Nam Năng Động (Tin Tức / Phóng Sự)', gender: 'male', tag: 'Rõ ràng', icon: '🎙️' },
    { id: 'male_story', name: 'Nam Kể Chuyện (Đêm Khuya / Trầm)', gender: 'male', tag: 'Trầm ấm', icon: '🌙' },
    { id: 'robot', name: 'Robot AI (Công Nghệ Sci-Fi)', gender: 'robot', tag: 'Hiệu ứng', icon: '🤖' },
    { id: 'fun_chipmunk', name: 'Hoạt Hình / Nhí Nhảnh', gender: 'fun', tag: 'Hài hước', icon: '🐿️' },
  ]
})

const selectedVoiceId = computed({
  get: () => editorStore.selectedVoiceId || 'male_deep',
  set: (val: string) => editorStore.setSelectedVoice(val)
})

const isPlayingVoicePreview = ref(false)
let previewAudioElement: HTMLAudioElement | null = null

function togglePreviewVoice() {
  if (previewAudioElement) {
    previewAudioElement.pause()
    previewAudioElement = null
    isPlayingVoicePreview.value = false
    return
  }

  const url = aiService.getVoicePreviewUrl(selectedVoiceId.value)
  const audio = new Audio(url)
  previewAudioElement = audio
  isPlayingVoicePreview.value = true

  audio.onended = () => {
    isPlayingVoicePreview.value = false
    previewAudioElement = null
  }
  audio.onerror = () => {
    isPlayingVoicePreview.value = false
    previewAudioElement = null
  }
  audio.play()
}

onMounted(() => {
  editorStore.loadDubbingVoices()
})

// Single cue audio preview state
const playingCueIndex = ref<number | null>(null)
let cueAudioElement: HTMLAudioElement | null = null

function playCueTTS(cue: SubtitleCue) {
  const videoId = editorStore.currentVideo?.id
  if (!videoId) return

  if (cueAudioElement) {
    cueAudioElement.pause()
    cueAudioElement = null
  }

  if (playingCueIndex.value === cue.index) {
    playingCueIndex.value = null
    return
  }

  playingCueIndex.value = cue.index
  const audioUrl = aiService.getCueAudioUrl(videoId, cue.index, selectedVoiceId.value)
  const audio = new Audio(audioUrl)
  cueAudioElement = audio

  audio.onended = () => {
    playingCueIndex.value = null
    cueAudioElement = null
  }
  audio.onerror = () => {
    playingCueIndex.value = null
    cueAudioElement = null
  }
  audio.play()
}

function togglePlayDubbedAudio() {
  editorStore.togglePlayDubbedAudio()
}

async function handleGenerateDubbing() {
  const isClip = dubbingScope.value === 'clip'
  const startMs = isClip ? currentClipStart.value : 0
  const endMs = isClip ? currentClipEnd.value : 0

  await editorStore.generateDubbing({
    start_ms: startMs,
    end_ms: endMs,
    bg_volume: 0.15,
    voice: selectedVoiceId.value,
  })
}

function downloadDubbedAudio() {
  if (!dubbing.value?.audio_url) return
  const id = editorStore.currentVideo?.id || 'video'
  const a = document.createElement('a')
  a.href = dubbing.value.audio_url
  a.download = `long_tieng_${id}.mp3`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function downloadDubbedVideo() {
  if (!dubbing.value?.video_url) return
  const id = editorStore.currentVideo?.id || 'video'
  const a = document.createElement('a')
  a.href = dubbing.value.video_url
  a.download = `video_long_tieng_${id}.mp4`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

const currentClipStart = computed(() => editorStore.startTime)
const currentClipEnd = computed(() => editorStore.endTime)
const totalDurationMs = computed(() => editorStore.duration || 0)
const currentTime = computed(() => editorStore.currentTime)

const filteredCues = computed(() => {
  if (!searchQuery.value.trim()) return cues.value
  const q = searchQuery.value.toLowerCase()
  return cues.value.filter((c: SubtitleCue) => c.text.toLowerCase().includes(q))
})

function isCueActive(cue: SubtitleCue): boolean {
  return currentTime.value >= cue.start_ms && currentTime.value <= cue.end_ms
}

function seekToCue(cue: SubtitleCue) {
  editorStore.setCurrentTime(cue.start_ms)
}

function toggleShowSubtitles() {
  editorStore.toggleShowSubtitles()
}

async function handleGenerate() {
  const isClip = scope.value === 'clip'
  const startMs = isClip ? currentClipStart.value : 0
  const endMs = isClip ? currentClipEnd.value : 0

  await editorStore.generateSubtitles(isClip, startMs, endMs)
  isExpandedSettings.value = false
}

async function handleReGenerateDirect() {
  const isClip = scope.value === 'clip'
  const startMs = isClip ? currentClipStart.value : 0
  const endMs = isClip ? currentClipEnd.value : 0

  await editorStore.generateSubtitles(isClip, startMs, endMs)
  isExpandedSettings.value = false
}

function downloadSrt() {
  if (!subtitles.value) return
  const id = editorStore.currentVideo?.id
  if (!id) return

  const url = `/api/v1/videos/${id}/subtitles/download`
  const a = document.createElement('a')
  a.href = url
  a.download = `subtitles_${id}.srt`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

// Auto scroll active cue into view
watch(
  () => editorStore.currentSubtitleCue,
  cue => {
    if (!cue || !cuesListRef.value) return
    nextTick(() => {
      const activeEl = cuesListRef.value?.querySelector('.cue-item--active') as HTMLElement | null
      if (activeEl) {
        activeEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      }
    })
  }
)
</script>

<style scoped>
.subtitle-panel {
  background: #FFFFFF;
  border: 1px solid var(--border, #E2E8F0);
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Header */
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-light, #F1F5F9);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.panel-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: rgba(139, 92, 246, 0.12);
  color: #8B5CF6;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #0F172A);
  margin: 0;
  line-height: 1.2;
}

.panel-subtitle {
  font-size: 12px;
  color: var(--text-secondary, #64748B);
  margin: 3px 0 0 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* CC Switch */
.cc-switch {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary, #475569);
  cursor: pointer;
  user-select: none;
}

.cc-switch input {
  display: none;
}

.switch-slider {
  width: 32px;
  height: 18px;
  background: #CBD5E1;
  border-radius: 10px;
  position: relative;
  transition: background 0.2s;
}

.switch-slider::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 14px;
  height: 14px;
  background: #FFFFFF;
  border-radius: 50%;
  transition: transform 0.2s;
}

.cc-switch input:checked + .switch-slider {
  background: #8B5CF6;
}

.cc-switch input:checked + .switch-slider::after {
  transform: translateX(14px);
}

/* Generating State */
.generating-box {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  background: #F8F7FF;
  border: 1px dashed #C4B5FD;
  border-radius: 8px;
}

.spinner-gemini {
  position: relative;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pulse-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: rgba(139, 92, 246, 0.25);
  animation: pulse 1.5s infinite ease-out;
}

@keyframes pulse {
  0% { transform: scale(0.8); opacity: 1; }
  100% { transform: scale(1.4); opacity: 0; }
}

.generating-text h5 {
  font-size: 13px;
  font-weight: 600;
  color: #6D28D9;
  margin: 0 0 4px 0;
}

.generating-text p {
  font-size: 12px;
  color: #7C3AED;
  margin: 0;
  line-height: 1.4;
}

/* Generate Controls */
.generate-controls-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #FAFAFA;
  border: 1px solid var(--border, #E2E8F0);
  border-radius: 8px;
  padding: 14px;
}

.controls-row {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: 16px;
}

@media (max-width: 640px) {
  .controls-row {
    grid-template-columns: 1fr;
  }
}

.control-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.control-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #64748B);
}

.control-input-wrap {
  display: flex;
  align-items: center;
  background: #FFFFFF;
  border: 1px solid var(--border, #CBD5E1);
  border-radius: 6px;
  padding: 0 10px;
  height: 36px;
  gap: 6px;
}

.flag-icon {
  font-size: 14px;
}

.input-readonly {
  border: none;
  background: transparent;
  font-size: 13px;
  color: var(--text-primary, #0F172A);
  font-weight: 500;
  outline: none;
  width: 100%;
}

.scope-radio-group {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.radio-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #FFFFFF;
  border: 1px solid var(--border, #CBD5E1);
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary, #334155);
  cursor: pointer;
  transition: all 0.15s;
}

.radio-label input {
  accent-color: #8B5CF6;
}

.radio-label--active {
  border-color: #8B5CF6;
  background: rgba(139, 92, 246, 0.06);
  color: #6D28D9;
}

.generate-btn-row {
  display: flex;
  gap: 10px;
  align-items: center;
}

.btn-generate-ai {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: #FFFFFF;
  border: none;
  border-radius: 6px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 2px 6px rgba(124, 58, 237, 0.25);
  transition: opacity 0.15s, transform 0.15s;
}

.btn-generate-ai:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
}

.btn-generate-ai:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.ai-sparkle {
  font-size: 15px;
}

.btn-cancel-settings {
  background: transparent;
  border: 1px solid var(--border, #CBD5E1);
  color: var(--text-secondary, #64748B);
  border-radius: 6px;
  padding: 8px 14px;
  font-size: 12px;
  cursor: pointer;
}

.note-box {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 11.5px;
  color: #64748B;
  background: #F1F5F9;
  padding: 8px 12px;
  border-radius: 6px;
  line-height: 1.4;
}

/* Results Area */
.subtitles-result-area {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.result-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.cue-count-badge {
  font-size: 12px;
  font-weight: 600;
  color: #6D28D9;
  background: rgba(139, 92, 246, 0.1);
  padding: 3px 8px;
  border-radius: 4px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cue-search-input {
  height: 30px;
  padding: 0 10px;
  border: 1px solid var(--border, #CBD5E1);
  border-radius: 6px;
  font-size: 12px;
  outline: none;
  min-width: 170px;
}

.cue-search-input:focus {
  border-color: #8B5CF6;
}

.btn-sub {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border, #CBD5E1);
  background: #FFFFFF;
  color: var(--text-primary, #334155);
}

.btn-sub:hover {
  background: #F8FAFC;
}

.btn-secondary {
  border-color: #8B5CF6;
  color: #6D28D9;
  background: rgba(139, 92, 246, 0.05);
}

.btn-secondary:hover {
  background: rgba(139, 92, 246, 0.12);
}

/* Cues List */
.cues-list {
  max-height: 240px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
  border: 1px solid var(--border, #E2E8F0);
  border-radius: 6px;
  padding: 6px;
  background: #FFFFFF;
}

.cue-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 10px;
  border-radius: 6px;
  background: #F8FAFC;
  cursor: pointer;
  border-left: 3px solid transparent;
  transition: background 0.15s, border-color 0.15s;
}

.cue-item:hover {
  background: #F1F5F9;
}

.cue-item--active {
  background: #F5F3FF !important;
  border-left-color: #8B5CF6;
}

.cue-time {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.time-badge {
  font-size: 11px;
  font-family: monospace;
  font-weight: 600;
  color: #64748B;
  background: #E2E8F0;
  padding: 2px 5px;
  border-radius: 4px;
}

.cue-item--active .time-badge {
  background: #DDD6FE;
  color: #5B21B6;
}

.time-arrow {
  font-size: 10px;
  color: #94A3B8;
}

.cue-content {
  flex: 1;
}

.cue-text {
  font-size: 13px;
  color: #1E293B;
  margin: 0;
  line-height: 1.35;
}

.cue-item--active .cue-text {
  font-weight: 600;
  color: #6D28D9;
}

.cue-edit-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.cue-edit-input {
  flex: 1;
  font-size: 13px;
  padding: 4px 8px;
  border: 1px solid #8B5CF6;
  border-radius: 4px;
  outline: none;
  background: #FFFFFF;
  color: #1E293B;
  box-shadow: 0 0 0 2px rgba(139, 92, 246, 0.15);
}

.cue-edit-btns {
  display: flex;
  gap: 4px;
}

.btn-cue-save {
  background: #10B981;
  color: #FFFFFF;
  border: none;
  border-radius: 4px;
  padding: 3px 7px;
  font-size: 11px;
  font-weight: bold;
  cursor: pointer;
}

.btn-cue-save:hover {
  background: #059669;
}

.btn-cue-cancel {
  background: #EF4444;
  color: #FFFFFF;
  border: none;
  border-radius: 4px;
  padding: 3px 7px;
  font-size: 11px;
  font-weight: bold;
  cursor: pointer;
}

.btn-cue-cancel:hover {
  background: #DC2626;
}

.cue-action-icon-btn {
  background: transparent;
  border: none;
  color: #94A3B8;
  padding: 4px;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: all 0.15s;
}

.cue-action-icon-btn:hover {
  color: #8B5CF6;
  background: rgba(139, 92, 246, 0.08);
}

.cue-translate-btn {
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  color: #059669;
  padding: 3px 7px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 3px;
  transition: all 0.15s;
}

.cue-translate-btn:hover:not(:disabled) {
  background: #10B981;
  color: #FFFFFF;
}

.cue-translate-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.cue-translating-txt {
  font-size: 10px;
}

.cue-play-btn {
  background: transparent;
  border: none;
  color: #94A3B8;
  padding: 4px;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
}

.cue-item:hover .cue-play-btn,
.cue-item--active .cue-play-btn {
  color: #8B5CF6;
}

.cue-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.cue-tts-btn {
  background: rgba(139, 92, 246, 0.08);
  border: 1px solid rgba(139, 92, 246, 0.2);
  color: #7C3AED;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  font-weight: 500;
  transition: all 0.15s;
}

.cue-tts-btn:hover {
  background: #7C3AED;
  color: #FFFFFF;
}

.cue-tts-btn--active {
  background: #7C3AED !important;
  color: #FFFFFF !important;
  animation: pulse 1s infinite alternate;
}

.tts-wave-txt {
  font-size: 11px;
  font-weight: 600;
}

/* Dubbing Banner */
.dubbing-banner-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px;
  border-radius: 8px;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.06) 0%, rgba(59, 130, 246, 0.06) 100%);
  border: 1px solid rgba(139, 92, 246, 0.2);
  margin-bottom: 12px;
}

.dubbing-banner-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dubbing-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.dubbing-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dubbing-heading {
  font-size: 13px;
  font-weight: 700;
  color: #4C1D95;
  margin: 0;
}

.dubbed-tag {
  background: #10B981;
  color: #FFFFFF;
  font-size: 10px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 10px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.dubbing-desc {
  font-size: 11.5px;
  color: #6B7280;
  margin: 2px 0 0 0;
  line-height: 1.4;
}

.dubbing-banner-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.voice-selector-box {
  display: flex;
  align-items: center;
  gap: 6px;
  background: #FFFFFF;
  padding: 3px 6px 3px 8px;
  border-radius: 6px;
  border: 1px solid #DDD6FE;
  box-shadow: 0 1px 2px rgba(124, 58, 237, 0.05);
}

.voice-label {
  font-size: 11px;
  font-weight: 700;
  color: #6D28D9;
  white-space: nowrap;
}

.voice-select {
  padding: 4px 6px;
  border-radius: 5px;
  border: 1px solid #E2E8F0;
  font-size: 11.5px;
  font-weight: 500;
  background: #F8FAFC;
  color: #1E293B;
  outline: none;
  cursor: pointer;
  max-width: 200px;
}

.voice-select:hover, .voice-select:focus {
  border-color: #8B5CF6;
  background: #FFFFFF;
}

.btn-preview-voice {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 5px;
  border: 1px solid #C4B5FD;
  background: #F5F3FF;
  color: #6D28D9;
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.btn-preview-voice:hover {
  background: #EDE9FE;
  border-color: #8B5CF6;
}

.btn-preview-voice--active {
  background: #EF4444 !important;
  color: #FFFFFF !important;
  border-color: #DC2626 !important;
}

.dubbing-scope-select .scope-select {
  padding: 5px 8px;
  border-radius: 6px;
  border: 1px solid var(--border, #CBD5E1);
  font-size: 11.5px;
  background: #FFFFFF;
  color: #334155;
  outline: none;
}

.btn-generate-dub {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 6px;
  border: none;
  background: linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%);
  color: #FFFFFF;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 2px 6px rgba(139, 92, 246, 0.3);
  transition: transform 0.15s, box-shadow 0.15s;
}

.btn-generate-dub:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(139, 92, 246, 0.4);
}

.btn-voice-toggle {
  border-color: #10B981;
  color: #047857;
  background: rgba(16, 185, 129, 0.08);
  font-weight: 600;
}

.btn-voice-toggle--active {
  background: #10B981 !important;
  color: #FFFFFF !important;
  border-color: #10B981 !important;
}

.dubbing-loading-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #7C3AED;
}

.dub-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(139, 92, 246, 0.2);
  border-top-color: #7C3AED;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes pulse {
  from { opacity: 0.8; }
  to { opacity: 1; }
}

.empty-search {
  text-align: center;
  padding: 24px;
  font-size: 12px;
  color: #94A3B8;
}
</style>
