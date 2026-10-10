<template>
  <div
    class="video-player-card"
    ref="playerContainerRef"
    :class="{
      'is-fullscreen': isFullscreen,
      'is-vertical': isVertical
    }"
  >
    <!-- Top Info Overlay (Resolution & Orientation Badge) -->
    <div v-if="videoWidth > 0" class="player-top-overlay">
      <div class="ratio-indicator-chip" :class="{ 'chip--vertical': isVertical }">
        <span class="ratio-icon">{{ isVertical ? '📱' : '🖥️' }}</span>
        <span class="ratio-text">{{ detectedRatioBadge }} &bull; {{ videoWidth }}x{{ videoHeight }}</span>
      </div>
      <button
        v-if="isFullscreen"
        class="fullscreen-exit-btn"
        @click="toggleFullscreen"
        title="Thoát toàn màn hình (Esc)"
      >
        ✕ Thoát toàn màn hình
      </button>
    </div>

    <!-- Video Element Container -->
    <div
      class="video-screen-wrap"
      :style="containerWrapStyle"
      @click="togglePlay"
    >
      <video
        ref="videoEl"
        class="video-screen"
        :class="{ 'video-screen--cover': fitMode === 'cover' }"
        :src="src"
        :style="videoScreenStyle"
        preload="metadata"
        playsinline
        @timeupdate="onTimeUpdate"
        @loadedmetadata="onLoadedMetadata"
        @ended="onEnded"
        @play="onVideoPlay"
        @pause="onVideoPause"
        @error="onVideoError"
      />

      <!-- Dubbed Audio Element (Synchronized Voiceover) -->
      <audio
        ref="dubbedAudioEl"
        v-if="hasDubbedAudio"
        :src="dubbedAudioSrc"
        preload="auto"
      />

      <!-- Error overlay if failed to load -->
      <div v-if="hasError" class="video-error-overlay" @click.stop>
        <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#EF4444" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        <p class="error-msg-text">{{ errorMessage }}</p>
        <button class="error-retry-btn" @click.stop="retryLoad">Thử tải lại</button>
      </div>

      <!-- Center Big Play Overlay when paused and no error -->
      <div v-else-if="!isPlaying" class="center-play-overlay">
        <button class="center-play-btn" @click.stop="togglePlay">
          <svg width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
        </button>
      </div>

      <!-- Live Subtitle Overlay -->
      <div v-if="showSubtitles && currentSubtitleText" class="video-subtitle-overlay" @click.stop="togglePlay">
        <span class="subtitle-text">{{ currentSubtitleText }}</span>
      </div>
    </div>

    <!-- Video Controls Toolbar -->
    <div class="player-controls">
      <!-- Row 1: Full-width Scrubber Progress Bar -->
      <div class="controls-scrubber-bar">
        <input
          type="range"
          min="0"
          :max="durationMs || 100"
          :value="currentTimeMs"
          step="100"
          @input="onSeekInput"
          class="progress-slider"
        />
      </div>

      <!-- Row 2: Action Buttons & Tools Toolbar -->
      <div class="controls-bottom-bar">
        <!-- Left: Play/Pause, Volume, Time -->
        <div class="controls-left">
          <button class="ctrl-btn play-btn" @click="togglePlay" :title="isPlaying ? 'Tạm dừng (Space)' : 'Phát (Space)'">
            <svg v-if="isPlaying" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <rect x="6" y="4" width="4" height="16"/>
              <rect x="14" y="4" width="4" height="16"/>
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
          </button>

          <!-- Volume -->
          <div class="volume-control">
            <button class="ctrl-btn" @click="toggleMute" title="Âm lượng">
              <svg v-if="isMuted || volume === 0" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="1" y1="1" x2="23" y2="23"/>
                <path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"/>
                <path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"/>
                <line x1="12" y1="19" x2="12" y2="23"/>
                <line x1="8" y1="23" x2="16" y2="23"/>
              </svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
                <path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/>
              </svg>
            </button>
            <input
              type="range"
              min="0"
              max="1"
              step="0.05"
              :value="isMuted ? 0 : volume"
              @input="onVolumeChange"
              class="volume-slider"
            />
          </div>

          <!-- Time counter: 00:24 / 03:15 -->
          <div class="time-display">
            <span class="current-time">{{ formatTimeShort(currentTimeMs) }}</span>
            <span class="time-sep">/</span>
            <span class="duration-time">{{ formatTimeShort(durationMs) }}</span>
          </div>
        </div>

        <!-- Right: Aspect Ratio, Fit Mode, Speed, Subtitle, Dubbing, Fullscreen -->
        <div class="controls-right">
          <!-- Aspect Ratio Selector Dropdown -->
          <div class="ratio-selector" title="Tỉ lệ khung xem video">
            <select v-model="selectedRatio" class="ratio-dropdown">
              <option value="auto">📐 {{ detectedRatioBadge }}</option>
              <option value="9:16">📱 9:16</option>
              <option value="3:4">📱 3:4</option>
              <option value="16:9">🖥️ 16:9</option>
              <option value="1:1">⬜ 1:1</option>
              <option value="original">🎞️ Gốc</option>
            </select>
          </div>

          <!-- Fit Mode: FIT (Chứa vừa) vs FILL (Lấp đầy) -->
          <button
            class="ctrl-btn fit-btn"
            :class="{ 'fit-btn--cover': fitMode === 'cover' }"
            @click="toggleFitMode"
            :title="fitMode === 'contain' ? 'Đang ở chế độ FIT (Chứa vừa 100% không cắt). Bấm để chuyển FILL (Lấp đầy)' : 'Đang ở chế độ FILL (Lấp đầy). Bấm để chuyển FIT (Chứa vừa 100%)'"
          >
            <span class="fit-badge">{{ fitMode === 'contain' ? 'FIT' : 'FILL' }}</span>
          </button>

          <!-- Speed Indicator / Quick dropdown -->
          <div class="speed-selector">
            <select v-model="playbackRate" @change="onRateChange" class="speed-dropdown" title="Tốc độ phát">
              <option :value="0.5">0.5x</option>
              <option :value="0.75">0.75x</option>
              <option :value="1.0">1.0x</option>
              <option :value="1.25">1.25x</option>
              <option :value="1.5">1.5x</option>
              <option :value="2.0">2.0x</option>
              <option :value="4.0">4.0x</option>
            </select>
          </div>

          <!-- Subtitle CC Toggle -->
          <button
            class="ctrl-btn cc-btn"
            :class="{ 'cc-btn--active': showSubtitles, 'cc-btn--has-sub': hasSubtitles }"
            @click="toggleSubtitles"
            :title="showSubtitles ? 'Tắt phụ đề (CC)' : 'Bật phụ đề (CC)'"
          >
            <span class="cc-badge">CC</span>
          </button>

          <!-- Dubbed Audio Voiceover Toggle -->
          <button
            v-if="hasDubbedAudio"
            class="ctrl-btn dub-btn"
            :class="{ 'dub-btn--active': isPlayingDubbedAudio }"
            @click="toggleDubbedAudio"
            :title="isPlayingDubbedAudio ? 'Đang bật lồng tiếng Tiếng Việt (Bấm để tắt)' : 'Bật lồng tiếng Tiếng Việt AI (Khớp nhịp thoại)'"
          >
            <span class="dub-badge">🎙️ VI</span>
          </button>

          <!-- Fullscreen Button -->
          <button
            class="ctrl-btn fullscreen-btn"
            :class="{ 'fullscreen-btn--active': isFullscreen }"
            @click="toggleFullscreen"
            :title="isFullscreen ? 'Thoát toàn màn hình (Esc hoặc F)' : 'Toàn màn hình (F)'"
          >
            <!-- Compress icon when in fullscreen -->
            <svg v-if="isFullscreen" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
            </svg>
            <!-- Expand icon when in normal window -->
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { formatTimeShort } from '@/utils/format'
import { useEditorStore } from '@/stores/editor.store'

const props = defineProps<{
  src: string
  seekMs?: number
  initialDurationMs?: number
}>()

const emit = defineEmits<{
  timeUpdate: [ms: number]
  durationLoaded: [ms: number]
}>()

const editorStore = useEditorStore()

const videoEl = ref<HTMLVideoElement | null>(null)
const playerContainerRef = ref<HTMLDivElement | null>(null)

const isPlaying = ref(false)
const currentTimeMs = ref(0)
const durationMs = ref(props.initialDurationMs || editorStore.duration || 195000)
const volume = ref(1.0)
const isMuted = ref(false)
const playbackRate = ref(1.0)
const hasError = ref(false)
const errorMessage = ref('')

// Dimension & Orientation State
const videoWidth = ref(0)
const videoHeight = ref(0)
const isFullscreen = ref(false)
const selectedRatio = ref<'auto' | '9:16' | '3:4' | '16:9' | '1:1' | 'original'>('auto')
const fitMode = ref<'contain' | 'cover'>('contain')

const isVertical = computed(() => videoHeight.value > videoWidth.value)
const naturalAspectRatio = computed(() => {
  if (videoWidth.value > 0 && videoHeight.value > 0) {
    return videoWidth.value / videoHeight.value
  }
  return 16 / 9
})

const detectedRatioBadge = computed(() => {
  if (videoWidth.value === 0) return 'Tự động'
  const r = naturalAspectRatio.value
  if (Math.abs(r - 9 / 16) < 0.06) return '9:16 Dọc'
  if (Math.abs(r - 3 / 4) < 0.06) return '3:4 Dọc'
  if (Math.abs(r - 1) < 0.06) return '1:1 Vuông'
  if (Math.abs(r - 16 / 9) < 0.06) return '16:9 Ngang'
  if (Math.abs(r - 4 / 3) < 0.06) return '4:3'
  return isVertical.value ? 'Dọc' : 'Ngang'
})

function toggleFitMode() {
  fitMode.value = fitMode.value === 'contain' ? 'cover' : 'contain'
}

// Adaptive Container Style
const containerWrapStyle = computed(() => {
  if (isFullscreen.value) {
    return {
      aspectRatio: 'unset',
      width: '100%',
      height: 'calc(100vh - 56px)',
      maxHeight: 'calc(100vh - 56px)',
    }
  }

  if (selectedRatio.value === '9:16') {
    return {
      aspectRatio: '9 / 16',
      maxHeight: '480px',
      margin: '0 auto',
      width: 'auto',
    }
  }
  if (selectedRatio.value === '3:4') {
    return {
      aspectRatio: '3 / 4',
      maxHeight: '480px',
      margin: '0 auto',
      width: 'auto',
    }
  }
  if (selectedRatio.value === '1:1') {
    return {
      aspectRatio: '1 / 1',
      maxHeight: '440px',
      margin: '0 auto',
      width: 'auto',
    }
  }
  if (selectedRatio.value === '16:9') {
    return {
      aspectRatio: '16 / 9',
      width: '100%',
    }
  }
  if (selectedRatio.value === 'original' && videoWidth.value > 0 && videoHeight.value > 0) {
    return {
      aspectRatio: `${videoWidth.value} / ${videoHeight.value}`,
      maxHeight: isVertical.value ? '480px' : 'none',
      margin: isVertical.value ? '0 auto' : '0',
      width: isVertical.value ? 'auto' : '100%',
    }
  }

  // Auto mode
  if (isVertical.value && videoWidth.value > 0 && videoHeight.value > 0) {
    return {
      aspectRatio: `${videoWidth.value} / ${videoHeight.value}`,
      maxHeight: '480px',
      margin: '0 auto',
      width: 'auto',
    }
  }

  // Default landscape
  return {
    aspectRatio: '16 / 9',
    width: '100%',
  }
})

const videoScreenStyle = computed(() => {
  return {
    filter: activeFilterCss.value,
    objectFit: fitMode.value,
  }
})

// Sync durationMs reactively with editorStore or prop
watch(
  () => props.initialDurationMs,
  newVal => {
    if (newVal && newVal > 0) {
      durationMs.value = newVal
    }
  }
)

watch(
  () => editorStore.duration,
  newVal => {
    if (newVal && newVal > 0) {
      durationMs.value = newVal
    }
  }
)

watch(
  () => props.src,
  () => {
    hasError.value = false
    errorMessage.value = ''
    if (videoEl.value) {
      videoEl.value.load()
    }
  }
)

// Real-time CSS filter from store
const activeFilterCss = computed(() => editorStore.activeFilterCss)

// Subtitles integration
const showSubtitles = computed(() => editorStore.showSubtitles)
const hasSubtitles = computed(() => !!editorStore.subtitles && !!editorStore.subtitles.cues && editorStore.subtitles.cues.length > 0)
const currentSubtitleText = computed(() => editorStore.currentSubtitleText)

function toggleSubtitles() {
  editorStore.toggleShowSubtitles()
}

// Dubbed Audio Voiceover Integration
const dubbedAudioEl = ref<HTMLAudioElement | null>(null)
const dubbing = computed(() => editorStore.dubbing)
const hasDubbedAudio = computed(() => !!dubbing.value && !!dubbing.value.audio_url)
const dubbedAudioSrc = computed(() => dubbing.value?.audio_url || '')
const isPlayingDubbedAudio = computed(() => editorStore.isPlayingDubbedAudio)

function toggleDubbedAudio() {
  editorStore.togglePlayDubbedAudio()
  syncAudioVolume()
}

function syncAudioVolume() {
  if (!videoEl.value) return
  if (isPlayingDubbedAudio.value) {
    videoEl.value.volume = 0.15 // Background music ducking
    if (dubbedAudioEl.value) {
      dubbedAudioEl.value.volume = 1.0
      if (isPlaying.value) {
        dubbedAudioEl.value.currentTime = videoEl.value.currentTime
        dubbedAudioEl.value.play().catch(() => {})
      }
    }
  } else {
    videoEl.value.volume = isMuted.value ? 0 : volume.value
    if (dubbedAudioEl.value) {
      dubbedAudioEl.value.pause()
    }
  }
}

watch(isPlayingDubbedAudio, () => {
  syncAudioVolume()
})

function onVideoPlay() {
  isPlaying.value = true
  if (isPlayingDubbedAudio.value && dubbedAudioEl.value && videoEl.value) {
    dubbedAudioEl.value.currentTime = videoEl.value.currentTime
    dubbedAudioEl.value.play().catch(() => {})
  }
}

function onVideoPause() {
  isPlaying.value = false
  if (dubbedAudioEl.value) {
    dubbedAudioEl.value.pause()
  }
}

// Synchronize speed dynamically based on current playing clip
watch(
  () => editorStore.currentPlayingClip?.speed,
  newSpeed => {
    if (newSpeed && videoEl.value) {
      videoEl.value.playbackRate = newSpeed
      playbackRate.value = newSpeed
      if (dubbedAudioEl.value) {
        dubbedAudioEl.value.playbackRate = newSpeed
      }
    }
  },
  { immediate: true }
)

function togglePlay() {
  const v = videoEl.value
  if (!v) return
  if (hasError.value) {
    retryLoad()
    return
  }
  if (v.paused) {
    v.play().catch(err => {
      console.warn('Playback error:', err)
    })
  } else {
    v.pause()
  }
}

function onTimeUpdate() {
  const v = videoEl.value
  if (!v) return
  currentTimeMs.value = Math.round(v.currentTime * 1000)
  emit('timeUpdate', currentTimeMs.value)

  // Auto-sync speed when crossing clip boundaries during playback
  const playingClip = editorStore.currentPlayingClip
  if (playingClip && Math.abs(v.playbackRate - playingClip.speed) > 0.01) {
    v.playbackRate = playingClip.speed
    playbackRate.value = playingClip.speed
    if (dubbedAudioEl.value) {
      dubbedAudioEl.value.playbackRate = playingClip.speed
    }
  }

  // Periodic drift check for dubbed audio (keep within 150ms)
  if (dubbedAudioEl.value && isPlayingDubbedAudio.value && isPlaying.value) {
    const diff = Math.abs(dubbedAudioEl.value.currentTime - v.currentTime)
    if (diff > 0.15) {
      dubbedAudioEl.value.currentTime = v.currentTime
    }
  }
}

function onLoadedMetadata() {
  const v = videoEl.value
  if (!v) return
  hasError.value = false
  errorMessage.value = ''

  if (v.videoWidth && v.videoHeight) {
    videoWidth.value = v.videoWidth
    videoHeight.value = v.videoHeight
  }

  if (v.duration && !isNaN(v.duration) && isFinite(v.duration)) {
    durationMs.value = Math.round(v.duration * 1000)
    emit('durationLoaded', durationMs.value)
  }
}

function onVideoError() {
  const v = videoEl.value
  hasError.value = true
  if (v?.error) {
    const code = v.error.code
    if (code === 2) {
      errorMessage.value = 'Lỗi mạng: Không thể tải luồng video từ máy chủ.'
    } else if (code === 4) {
      errorMessage.value = 'Không tìm thấy nguồn video hoặc định dạng video không được hỗ trợ.'
    } else {
      errorMessage.value = `Lỗi phát video (Mã lỗi: ${code}).`
    }
  } else {
    errorMessage.value = 'Không thể tải được video.'
  }
}

function retryLoad() {
  hasError.value = false
  errorMessage.value = ''
  if (videoEl.value) {
    videoEl.value.load()
    videoEl.value.play().catch(() => {})
  }
}

function onEnded() {
  isPlaying.value = false
  if (dubbedAudioEl.value) {
    dubbedAudioEl.value.pause()
  }
}

function onSeekInput(e: Event) {
  const target = e.target as HTMLInputElement
  const ms = Number(target.value)
  seekTo(ms)
}

function seekTo(ms: number) {
  const v = videoEl.value
  if (!v) return
  v.currentTime = ms / 1000
  currentTimeMs.value = ms
  if (dubbedAudioEl.value) {
    dubbedAudioEl.value.currentTime = ms / 1000
  }
  emit('timeUpdate', ms)
}

function onVolumeChange(e: Event) {
  const target = e.target as HTMLInputElement
  volume.value = Number(target.value)
  isMuted.value = volume.value === 0
  if (videoEl.value) {
    if (isPlayingDubbedAudio.value) {
      videoEl.value.volume = 0.15
    } else {
      videoEl.value.volume = volume.value
    }
  }
}

function toggleMute() {
  if (!videoEl.value) return
  isMuted.value = !isMuted.value
  videoEl.value.muted = isMuted.value
}

function onRateChange() {
  const rate = Number(playbackRate.value)
  if (videoEl.value) {
    videoEl.value.playbackRate = rate
  }
  if (editorStore.selectedTimelineClip) {
    editorStore.setClipSpeed(editorStore.selectedTimelineClip.id, rate)
  }
}

function updateFullscreenState() {
  isFullscreen.value = !!document.fullscreenElement
}

function toggleFullscreen() {
  const el = playerContainerRef.value
  if (!el) return
  if (!document.fullscreenElement) {
    if (el.requestFullscreen) {
      el.requestFullscreen().catch(() => {})
    } else if ((el as any).webkitRequestFullscreen) {
      (el as any).webkitRequestFullscreen()
    }
  } else {
    if (document.exitFullscreen) {
      document.exitFullscreen().catch(() => {})
    } else if ((document as any).webkitExitFullscreen) {
      (document as any).webkitExitFullscreen()
    }
  }
}

// Keyboard shortcuts handler
function onKeyDown(e: KeyboardEvent) {
  // Do not trigger shortcuts when typing in inputs or textareas
  const activeEl = document.activeElement
  const tagName = activeEl?.tagName.toLowerCase()
  if (tagName === 'input' || tagName === 'textarea' || tagName === 'select') {
    return
  }

  if (e.code === 'Space') {
    e.preventDefault()
    togglePlay()
  } else if (e.code === 'KeyF') {
    e.preventDefault()
    toggleFullscreen()
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'z' && !e.shiftKey) {
    e.preventDefault()
    editorStore.undo()
  } else if ((e.ctrlKey || e.metaKey) && (e.key.toLowerCase() === 'y' || (e.key.toLowerCase() === 'z' && e.shiftKey))) {
    e.preventDefault()
    editorStore.redo()
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
    e.preventDefault()
    editorStore.splitClipAtCurrentTime()
  } else if (e.code === 'ArrowLeft') {
    e.preventDefault()
    const target = Math.max(0, currentTimeMs.value - 1000)
    seekTo(target)
  } else if (e.code === 'ArrowRight') {
    e.preventDefault()
    const target = Math.min(durationMs.value, currentTimeMs.value + 1000)
    seekTo(target)
  }
}

watch(
  () => props.seekMs,
  newMs => {
    if (newMs !== undefined && Math.abs(currentTimeMs.value - newMs) > 200) {
      seekTo(newMs)
    }
  }
)

defineExpose({ seekTo, togglePlay })

onMounted(() => {
  if (videoEl.value) {
    videoEl.value.volume = volume.value
    if (videoEl.value.videoWidth && videoEl.value.videoHeight) {
      videoWidth.value = videoEl.value.videoWidth
      videoHeight.value = videoEl.value.videoHeight
    }
  }
  window.addEventListener('keydown', onKeyDown)
  document.addEventListener('fullscreenchange', updateFullscreenState)
  document.addEventListener('webkitfullscreenchange', updateFullscreenState)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  document.removeEventListener('fullscreenchange', updateFullscreenState)
  document.removeEventListener('webkitfullscreenchange', updateFullscreenState)
})
</script>

<style scoped>
.video-player-card {
  background: #000000;
  border-radius: var(--radius-lg, 8px);
  overflow: hidden;
  box-shadow: var(--shadow-md, 0 4px 6px -1px rgba(0, 0, 0, 0.1));
  display: flex;
  flex-direction: column;
  position: relative;
  transition: all 0.2s ease;
}

/* Top Overlay Badge for Resolution & Ratio */
.player-top-overlay {
  position: absolute;
  top: 10px;
  left: 12px;
  right: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  pointer-events: none;
  z-index: 15;
  opacity: 0.9;
  transition: opacity 0.2s ease;
}

.ratio-indicator-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: rgba(15, 23, 42, 0.75);
  backdrop-filter: blur(6px);
  color: #E2E8F0;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 20px;
  padding: 3px 10px;
  font-size: 11px;
  font-weight: 600;
  pointer-events: auto;
}

.ratio-indicator-chip.chip--vertical {
  border-color: rgba(59, 130, 246, 0.4);
  color: #93C5FD;
  background: rgba(30, 58, 138, 0.6);
}

.fullscreen-exit-btn {
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(6px);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 6px;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  pointer-events: auto;
  transition: all 0.15s ease;
}

.fullscreen-exit-btn:hover {
  background: #DC2626;
  border-color: #DC2626;
}

.video-screen-wrap {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000000;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  overflow: hidden;
}

.video-screen {
  width: 100%;
  height: 100%;
  object-fit: contain;
  transition: filter 0.15s ease;
  display: block;
}

.video-screen--cover {
  object-fit: cover !important;
}

.center-play-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.25);
  pointer-events: none;
}

.video-error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(4px);
  padding: 24px;
  text-align: center;
  z-index: 5;
}

.error-msg-text {
  color: #F87171;
  font-size: 14px;
  font-weight: 500;
  max-width: 400px;
  line-height: 1.5;
  margin: 0;
}

.error-retry-btn {
  background: #2563EB;
  color: #FFFFFF;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.15s ease;
}

.error-retry-btn:hover {
  background: #1D4ED8;
}

.center-play-btn {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: rgba(22, 119, 255, 0.9);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-left: 4px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  pointer-events: auto;
  transition: transform 0.15s ease, background 0.15s ease;
}

.center-play-btn:hover {
  transform: scale(1.08);
  background: var(--primary, #1677FF);
}

/* Controls Toolbar - 2-row layout (Row 1: full-width scrubber, Row 2: buttons) */
.player-controls {
  height: auto;
  background: #111827;
  padding: 0;
  display: flex;
  flex-direction: column;
  color: #F3F4F6;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
  width: 100%;
  box-sizing: border-box;
}

/* Row 1: Full-width Scrubber Bar */
.controls-scrubber-bar {
  width: 100%;
  height: 12px;
  padding: 2px 8px 0 8px;
  display: flex;
  align-items: center;
  box-sizing: border-box;
}

.progress-slider {
  width: 100%;
  height: 4px;
  accent-color: var(--primary, #1677FF);
  cursor: pointer;
  border-radius: 2px;
  transition: height 0.15s ease;
  margin: 0;
}

.controls-scrubber-bar:hover .progress-slider {
  height: 6px;
}

/* Row 2: Bottom Toolbar Action Buttons */
.controls-bottom-bar {
  height: 38px;
  padding: 0 8px 4px 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  box-sizing: border-box;
  gap: 6px;
}

.controls-left {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.ctrl-btn {
  background: transparent;
  color: #E5E7EB;
  width: 28px;
  height: 28px;
  border-radius: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  flex-shrink: 0;
  border: none;
  cursor: pointer;
  padding: 0;
}

.ctrl-btn:hover {
  color: #FFFFFF;
  background: rgba(255, 255, 255, 0.14);
}

.play-btn {
  color: #FFFFFF;
  background: rgba(255, 255, 255, 0.1);
}

.volume-control {
  display: flex;
  align-items: center;
  gap: 3px;
}

.volume-slider {
  width: 44px;
  height: 3px;
  accent-color: var(--primary, #1677FF);
  cursor: pointer;
}

.time-display {
  display: flex;
  align-items: center;
  gap: 3px;
  font-size: 11px;
  font-weight: 500;
  font-family: var(--font-mono, monospace);
  white-space: nowrap;
}

.current-time {
  color: #FFFFFF;
}

.time-sep {
  color: #6B7280;
}

.duration-time {
  color: #9CA3AF;
}

/* Right Controls */
.controls-right {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
  margin-left: auto;
}

/* Aspect Ratio Selector */
.ratio-selector {
  display: flex;
  align-items: center;
}

.ratio-dropdown {
  background: rgba(255, 255, 255, 0.1);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-radius: 5px;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 4px;
  cursor: pointer;
  outline: none;
  transition: all 0.15s ease;
  max-width: 82px;
}

.ratio-dropdown:hover {
  background: rgba(255, 255, 255, 0.16);
  border-color: rgba(255, 255, 255, 0.3);
}

.ratio-dropdown option {
  background: #1F2937;
  color: #FFFFFF;
}

/* Fit Mode Button */
.fit-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 5px;
  height: 26px;
  min-width: 28px;
  padding: 0 4px;
  cursor: pointer;
}

.fit-badge {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.5px;
  font-family: var(--font-mono, monospace);
  color: #93C5FD;
}

.fit-btn--cover {
  background: rgba(59, 130, 246, 0.2);
  border-color: #3B82F6;
}

.fit-btn--cover .fit-badge {
  color: #60A5FA;
}

.speed-dropdown {
  background: rgba(255, 255, 255, 0.1);
  color: #FFFFFF;
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 5px;
  font-size: 11px;
  font-weight: 500;
  padding: 2px 4px;
  cursor: pointer;
}

.speed-dropdown option {
  background: #1F2937;
  color: #FFFFFF;
}

/* Subtitle Overlay */
.video-subtitle-overlay {
  position: absolute;
  bottom: 24px;
  left: 5%;
  right: 5%;
  display: flex;
  justify-content: center;
  align-items: center;
  pointer-events: none;
  z-index: 10;
}

.subtitle-text {
  background: rgba(0, 0, 0, 0.75);
  color: #FFFFFF;
  font-size: 15px;
  font-weight: 500;
  line-height: 1.4;
  padding: 6px 14px;
  border-radius: 6px;
  text-align: center;
  max-width: 90%;
  backdrop-filter: blur(4px);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  letter-spacing: 0.2px;
}

/* CC Toggle Button */
.cc-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.cc-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 1px 3px;
  border-radius: 3px;
  border: 1.2px solid currentColor;
  line-height: 1;
  letter-spacing: 0.5px;
}

.cc-btn--active {
  color: #3B82F6 !important;
}

.cc-btn--has-sub:not(.cc-btn--active) {
  opacity: 0.7;
}

/* Dubbed Audio Toggle Button */
.dub-btn {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.dub-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 4px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.08);
  line-height: 1;
}

.dub-btn--active .dub-badge {
  background: #10B981;
  color: #FFFFFF;
  border-color: #10B981;
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.fullscreen-btn--active {
  color: #3B82F6 !important;
}

/* ==========================================================================
   FULLSCREEN MODE STYLING (Comprehensive :fullscreen and .is-fullscreen rules)
   Guarantees vertical videos (3:4, 9:16) scale cleanly, centered with letterbox,
   without cropping face/body/subtitles, and keeps controls visible.
   ========================================================================== */

:fullscreen.video-player-card,
:-webkit-full-screen.video-player-card,
.video-player-card.is-fullscreen {
  position: fixed !important;
  inset: 0 !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  bottom: 0 !important;
  width: 100vw !important;
  height: 100vh !important;
  max-width: 100vw !important;
  max-height: 100vh !important;
  background: #000000 !important;
  border-radius: 0 !important;
  display: flex !important;
  flex-direction: column !important;
  justify-content: space-between !important;
  overflow: hidden !important;
  padding: 0 !important;
  margin: 0 !important;
  z-index: 999999 !important;
}

:fullscreen .video-screen-wrap,
:-webkit-full-screen .video-screen-wrap,
.video-player-card.is-fullscreen .video-screen-wrap {
  flex: 1 1 auto !important;
  width: 100% !important;
  height: calc(100vh - 64px) !important;
  max-height: calc(100vh - 64px) !important;
  aspect-ratio: unset !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
  background: #000000 !important;
  position: relative !important;
  overflow: hidden !important;
  margin: 0 !important;
}

:fullscreen .video-screen,
:-webkit-full-screen .video-screen,
.video-player-card.is-fullscreen .video-screen {
  max-width: 100% !important;
  max-height: calc(100vh - 64px) !important;
  width: auto !important;
  height: 100% !important;
  object-fit: contain !important;
  margin: auto !important;
  display: block !important;
}

:fullscreen .video-screen--cover,
:-webkit-full-screen .video-screen--cover,
.video-player-card.is-fullscreen .video-screen--cover {
  width: 100% !important;
  height: 100% !important;
  object-fit: cover !important;
}

:fullscreen .player-controls,
:-webkit-full-screen .player-controls,
.video-player-card.is-fullscreen .player-controls {
  position: relative !important;
  bottom: 0 !important;
  left: 0 !important;
  right: 0 !important;
  width: 100% !important;
  height: auto !important;
  flex-shrink: 0 !important;
  background: rgba(15, 23, 42, 0.96) !important;
  backdrop-filter: blur(12px) !important;
  border-top: 1px solid rgba(255, 255, 255, 0.12) !important;
  z-index: 100 !important;
  padding: 4px 16px 8px 16px !important;
}

:fullscreen .controls-scrubber-bar,
:-webkit-full-screen .controls-scrubber-bar,
.video-player-card.is-fullscreen .controls-scrubber-bar {
  padding: 0 4px !important;
  height: 16px !important;
}

:fullscreen .controls-bottom-bar,
:-webkit-full-screen .controls-bottom-bar,
.video-player-card.is-fullscreen .controls-bottom-bar {
  height: 44px !important;
  padding: 0 !important;
}

:fullscreen .video-subtitle-overlay,
:-webkit-full-screen .video-subtitle-overlay,
.video-player-card.is-fullscreen .video-subtitle-overlay {
  bottom: 76px !important;
  z-index: 50 !important;
}

:fullscreen .subtitle-text,
:-webkit-full-screen .subtitle-text,
.video-player-card.is-fullscreen .subtitle-text {
  font-size: 22px !important;
  padding: 8px 24px !important;
  max-width: 80% !important;
  border-radius: 8px !important;
  background: rgba(0, 0, 0, 0.8) !important;
}
</style>

