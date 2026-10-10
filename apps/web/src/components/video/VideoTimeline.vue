<template>
  <div class="timeline-section card-base">
    <!-- Section Header: Title, Summary, & Toolbar -->
    <div class="timeline-header-area">
      <div class="timeline-header">
        <div class="header-title-wrap">
          <div class="section-badge-icon">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="6" cy="6" r="3"/>
              <circle cx="6" cy="18" r="3"/>
              <line x1="20" y1="4" x2="8.12" y2="15.88"/>
              <line x1="14.47" y1="14.48" x2="20" y2="20"/>
              <line x1="8.12" y1="8.12" x2="12" y2="12"/>
            </svg>
          </div>
          <h3 class="section-title">Trình Edit & Cắt Video</h3>
        </div>

        <div class="header-summary">
          <span class="summary-item">
            Đoạn chọn: <strong>{{ formatTimeDetailed(startTime) }}</strong> - <strong>{{ formatTimeDetailed(endTime) }}</strong>
          </span>
          <span class="summary-sep">•</span>
          <span class="summary-item duration-badge">
            Thời lượng đoạn: <strong>{{ formatTimeDetailed(clipDuration) }} ({{ clipDurationSec }}s)</strong>
          </span>
          <span v-if="timelineClips.length > 1" class="summary-sep">•</span>
          <span v-if="timelineClips.length > 1" class="summary-item output-badge">
            Tổng xuất (EDL): <strong>{{ formatDurationSec(totalOutputDurationMs) }}</strong>
          </span>
        </div>
      </div>

      <!-- Editor Toolbar (Trim, Split, Speed, Filter) -->
      <EditorToolbar />
    </div>

    <!-- Active Tool Sub-Panels (Inline beneath toolbar) -->
    <div v-if="activeTool === 'speed'" class="tool-subpanel">
      <SpeedPanel />
    </div>

    <div v-if="activeTool === 'filter'" class="tool-subpanel">
      <FilterPanel />
    </div>

    <div v-if="activeTool === 'subtitle'" class="tool-subpanel">
      <SubtitlePanel />
    </div>

    <!-- Timeline Scrubber & Filmstrip -->
    <div class="timeline-track-wrap" ref="trackContainerRef" @mousedown="onTrackMouseDown">
      <!-- Time Markers Scale (Ruler) -->
      <div class="time-markers-ruler">
        <div
          v-for="marker in timeMarkers"
          :key="marker.ms"
          class="time-marker-tick"
          :style="{ left: marker.pct + '%' }"
        >
          <span class="marker-label">{{ marker.label }}</span>
          <span class="marker-bar"></span>
        </div>
      </div>

      <!-- Filmstrip thumbnails background -->
      <div class="filmstrip-bg">
        <div
          v-for="i in 16"
          :key="i"
          class="film-frame"
          :style="{ backgroundImage: `url(${thumbnailUrl})` }"
        ></div>
      </div>

      <!-- Multi-Clip Blocks Layer (Shown only after splitting into 2+ clips) -->
      <div v-if="timelineClips.length > 1" class="timeline-clips-layer">
        <div
          v-for="(clip, idx) in timelineClips"
          :key="clip.id"
          class="timeline-clip-block"
          :class="{
            'timeline-clip-block--selected': selectedClipId === clip.id,
            'timeline-clip-block--has-speed': clip.speed !== 1,
            'timeline-clip-block--has-filter': hasClipFilter(clip)
          }"
          :style="getClipBlockStyle(clip)"
          @mousedown.stop="onClipBlockMouseDown(clip, $event)"
          @click.stop="onClipClick(clip.id)"
          :title="`Đoạn ${idx + 1}: ${formatTimeShort(clip.sourceStartMs)} - ${formatTimeShort(clip.sourceEndMs)}`"
        >
          <!-- Clip Info Badges -->
          <div class="clip-badges-wrap">
            <span class="clip-index-tag">#{{ idx + 1 }}</span>
            <span v-if="clip.speed !== 1" class="clip-badge-pill clip-badge-speed">
              {{ clip.speed }}x
            </span>
            <span v-if="hasClipFilter(clip)" class="clip-badge-pill clip-badge-filter" title="Đang áp dụng bộ lọc màu">
              ◐
            </span>
          </div>

          <!-- Quick Delete / Unsplit Clip Action -->
          <button
            v-if="selectedClipId === clip.id"
            class="clip-delete-btn"
            @click.stop="editorStore.unsplitSelectedClip(clip.id)"
            title="Xóa / Ghép lại đoạn này"
          >
            ×
          </button>
        </div>
      </div>

      <!-- Selected Trim Range Overlay (Blue Transparent) - Draggable to move whole frame -->
      <div
        class="selection-range"
        :class="{ 'selection-range--dragging': draggingType === 'range' }"
        :style="selectionRangeStyle"
        @mousedown.stop="startDrag('range', $event)"
        @touchstart.stop="startDrag('range', $event)"
        title="Kéo để di chuyển cả khung cắt (giữ nguyên độ dài)"
      >
        <span class="range-move-grip">⋮⋮</span>
        <span class="range-duration-tag">{{ clipDurationSec }}s</span>
      </div>

      <!-- Left Handle (Start) -->
      <div
        class="timeline-handle handle-left"
        :style="{ left: startPct + '%' }"
        @mousedown.stop="startDrag('start', $event)"
        @touchstart.stop="startDrag('start', $event)"
        title="Kéo để chỉnh điểm bắt đầu"
      >
        <div class="handle-grip"></div>
        <span class="handle-tooltip">{{ formatTimeShort(startTime) }}</span>
      </div>

      <!-- Right Handle (End) -->
      <div
        class="timeline-handle handle-right"
        :style="{ left: endPct + '%' }"
        @mousedown.stop="startDrag('end', $event)"
        @touchstart.stop="startDrag('end', $event)"
        title="Kéo để chỉnh điểm kết thúc"
      >
        <div class="handle-grip"></div>
        <span class="handle-tooltip">{{ formatTimeShort(endTime) }}</span>
      </div>

      <!-- Moving Playhead -->
      <div
        class="timeline-playhead"
        :style="{ left: playheadPct + '%' }"
        @mousedown.stop="startPlayheadDrag"
        @touchstart.stop="startPlayheadDrag"
      >
        <div class="playhead-cap">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
            <polygon points="12 22 2 4 22 4 12 22"/>
          </svg>
        </div>
        <div class="playhead-line"></div>
      </div>
    </div>

    <!-- Inputs & Actions Toolbar Below Timeline -->
    <div class="timeline-controls-footer">
      <div class="inputs-group">
        <!-- Start time input -->
        <div class="time-input-field">
          <label class="input-label">Thời gian bắt đầu</label>
          <div class="input-box">
            <input
              type="text"
              class="field-input"
              v-model="startInputText"
              @blur="commitStartInput"
              @keydown.enter="commitStartInput"
            />
            <div class="step-buttons">
              <button class="step-btn" @click="stepTime('start', 1000)">▲</button>
              <button class="step-btn" @click="stepTime('start', -1000)">▼</button>
            </div>
          </div>
        </div>

        <!-- End time input -->
        <div class="time-input-field">
          <label class="input-label">Thời gian kết thúc</label>
          <div class="input-box">
            <input
              type="text"
              class="field-input"
              v-model="endInputText"
              @blur="commitEndInput"
              @keydown.enter="commitEndInput"
            />
            <div class="step-buttons">
              <button class="step-btn" @click="stepTime('end', 1000)">▲</button>
              <button class="step-btn" @click="stepTime('end', -1000)">▼</button>
            </div>
          </div>
        </div>

        <!-- Duration display -->
        <div class="time-input-field">
          <label class="input-label">Thời lượng đoạn cắt</label>
          <div class="input-box input-box--readonly">
            <span class="field-readonly">{{ formatTimeDetailed(clipDuration) }} ({{ clipDurationSec }}s)</span>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="actions-group">
        <AppButton variant="secondary" size="md" @click="resetTimeline">
          <template #icon-left>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="1 4 1 10 7 10"/>
              <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
            </svg>
          </template>
          Đặt lại
        </AppButton>

        <AppButton
          variant="primary"
          size="md"
          :loading="isSavingClip"
          @click="handleSaveClip"
        >
          <template #icon-left>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="6" cy="6" r="3"/>
              <circle cx="6" cy="18" r="3"/>
              <line x1="20" y1="4" x2="8.12" y2="15.88"/>
              <line x1="14.47" y1="14.48" x2="20" y2="20"/>
              <line x1="8.12" y1="8.12" x2="12" y2="12"/>
            </svg>
          </template>
          Cắt và lưu đoạn này
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useEditorStore } from '@/stores/editor.store'
import { formatTimeShort, formatDurationSec } from '@/utils/format'
import type { TimelineClip } from '@/types'
import AppButton from '@/components/common/AppButton.vue'
import EditorToolbar from '@/components/video/EditorToolbar.vue'
import SpeedPanel from '@/components/video/SpeedPanel.vue'
import FilterPanel from '@/components/video/FilterPanel.vue'
import SubtitlePanel from '@/components/video/SubtitlePanel.vue'

const editorStore = useEditorStore()

const trackContainerRef = ref<HTMLDivElement | null>(null)
const isSavingClip = ref(false)

const startTime = computed(() => editorStore.startTime)
const endTime = computed(() => editorStore.endTime)
const duration = computed(() => editorStore.duration || 195000)
const currentTime = computed(() => editorStore.currentTime)
const clipDuration = computed(() => editorStore.clipDuration)
const clipDurationSec = computed(() => editorStore.clipDurationSec)

const timelineClips = computed(() => editorStore.timelineClips)
const selectedClipId = computed(() => editorStore.selectedTimelineClipId)
const activeTool = computed(() => editorStore.activeTool)
const totalOutputDurationMs = computed(() => editorStore.totalTimelineOutputDurationMs)

const thumbnailUrl = computed(() => editorStore.currentVideo.thumbnail_url)

// Percentages
const startPct = computed(() => (startTime.value / duration.value) * 100)
const endPct = computed(() => (endTime.value / duration.value) * 100)
const playheadPct = computed(() => (currentTime.value / duration.value) * 100)

const selectionRangeStyle = computed(() => ({
  left: startPct.value + '%',
  width: Math.max(0, endPct.value - startPct.value) + '%',
}))

function getClipBlockStyle(clip: TimelineClip) {
  const leftPct = (clip.sourceStartMs / duration.value) * 100
  const widthPct = Math.max(0, ((clip.sourceEndMs - clip.sourceStartMs) / duration.value) * 100)
  return {
    left: leftPct + '%',
    width: widthPct + '%',
  }
}

function hasClipFilter(clip: TimelineClip): boolean {
  const f = clip.filter
  if (!f) return false
  return f.preset !== 'none' || f.brightness !== 0 || f.contrast !== 0 || f.saturation !== 0 || f.temperature !== 0
}

function onClipClick(clipId: string) {
  editorStore.selectTimelineClip(clipId)
}

function onClipBlockMouseDown(clip: TimelineClip, e: MouseEvent) {
  editorStore.selectTimelineClip(clip.id)
  const pct = getPctFromEvent(e)
  const ms = Math.round((pct / 100) * duration.value)
  editorStore.setCurrentTime(ms)
}

// Time markers along ruler with dynamic step calculation to avoid squished labels
const timeMarkers = computed(() => {
  const d = duration.value || 195000
  const list: { ms: number; label: string; pct: number }[] = []
  
  let step = 30000 // default 30s
  if (d > 7200000) step = 1800000      // > 2h: step 30m
  else if (d > 3600000) step = 600000  // > 1h: step 10m
  else if (d > 1200000) step = 300000  // > 20m: step 5m
  else if (d > 600000) step = 120000   // > 10m: step 2m
  else if (d > 180000) step = 60000    // > 3m: step 1m
  else if (d > 60000) step = 20000     // > 1m: step 20s
  else step = 5000                     // <= 1m: step 5s
  
  for (let ms = 0; ms <= d; ms += step) {
    list.push({
      ms,
      label: formatTimeShort(ms),
      pct: (ms / d) * 100,
    })
  }

  // Ensure end duration marker exists if far enough from last step
  const lastMs = list[list.length - 1]?.ms || 0
  if (d - lastMs > step * 0.4) {
    list.push({
      ms: d,
      label: formatTimeShort(d),
      pct: 100,
    })
  }

  return list
})

// Input text models
const startInputText = ref(formatTimeDetailed(startTime.value))
const endInputText = ref(formatTimeDetailed(endTime.value))

watch(startTime, val => {
  startInputText.value = formatTimeDetailed(val)
})

watch(endTime, val => {
  endInputText.value = formatTimeDetailed(val)
})

function formatTimeDetailed(ms: number): string {
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${pad(h)}:${pad(m)}:${pad(s)}`
}

function parseTimeToMs(text: string): number | null {
  const parts = text.trim().split(':').map(Number)
  if (parts.some(isNaN)) return null
  if (parts.length === 3) {
    return (parts[0] * 3600 + parts[1] * 60 + parts[2]) * 1000
  }
  if (parts.length === 2) {
    return (parts[0] * 60 + parts[1]) * 1000
  }
  if (parts.length === 1) {
    return parts[0] * 1000
  }
  return null
}

function commitStartInput() {
  const parsed = parseTimeToMs(startInputText.value)
  if (parsed !== null) {
    editorStore.setStartTime(parsed)
  } else {
    startInputText.value = formatTimeDetailed(startTime.value)
  }
}

function commitEndInput() {
  const parsed = parseTimeToMs(endInputText.value)
  if (parsed !== null) {
    editorStore.setEndTime(parsed)
  } else {
    endInputText.value = formatTimeDetailed(endTime.value)
  }
}

function stepTime(target: 'start' | 'end', deltaMs: number) {
  if (target === 'start') {
    editorStore.setStartTime(startTime.value + deltaMs)
  } else {
    editorStore.setEndTime(endTime.value + deltaMs)
  }
}

function resetTimeline() {
  editorStore.resetRange()
}

async function handleSaveClip() {
  isSavingClip.value = true
  try {
    await editorStore.addCurrentClip()
  } finally {
    isSavingClip.value = false
  }
}

// Dragging logic
let draggingType: 'start' | 'end' | 'playhead' | 'range' | null = null
let initialDragClientX = 0
let initialDragStartMs = 0
let initialDragEndMs = 0
let initialDragDurationMs = 0

function getPctFromEvent(e: MouseEvent | TouchEvent): number {
  if (!trackContainerRef.value) return 0
  const rect = trackContainerRef.value.getBoundingClientRect()
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  const pct = ((clientX - rect.left) / rect.width) * 100
  return Math.max(0, Math.min(100, pct))
}

function startDrag(type: 'start' | 'end' | 'range', e: MouseEvent | TouchEvent) {
  draggingType = type
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  initialDragClientX = clientX
  initialDragStartMs = startTime.value
  initialDragEndMs = endTime.value
  initialDragDurationMs = Math.max(1000, initialDragEndMs - initialDragStartMs)

  window.addEventListener('mousemove', onDragging)
  window.addEventListener('mouseup', stopDrag)
  window.addEventListener('touchmove', onDragging)
  window.addEventListener('touchend', stopDrag)
  e.preventDefault()
}

function startPlayheadDrag(e: MouseEvent | TouchEvent) {
  draggingType = 'playhead'
  window.addEventListener('mousemove', onDragging)
  window.addEventListener('mouseup', stopDrag)
  window.addEventListener('touchmove', onDragging)
  window.addEventListener('touchend', stopDrag)
  e.preventDefault()
}

function onDragging(e: MouseEvent | TouchEvent) {
  if (!draggingType) return

  if (draggingType === 'range') {
    if (!trackContainerRef.value) return
    const rect = trackContainerRef.value.getBoundingClientRect()
    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
    const deltaX = clientX - initialDragClientX
    const deltaMs = Math.round((deltaX / rect.width) * duration.value)

    let newStart = initialDragStartMs + deltaMs
    newStart = Math.max(0, Math.min(newStart, duration.value - initialDragDurationMs))
    let newEnd = newStart + initialDragDurationMs

    editorStore.setRange(newStart, newEnd)
    return
  }

  const pct = getPctFromEvent(e)
  const ms = Math.round((pct / 100) * duration.value)

  if (draggingType === 'start') {
    editorStore.setStartTime(ms)
  } else if (draggingType === 'end') {
    editorStore.setEndTime(ms)
  } else if (draggingType === 'playhead') {
    editorStore.setCurrentTime(ms)
  }
}

function stopDrag() {
  draggingType = null
  window.removeEventListener('mousemove', onDragging)
  window.removeEventListener('mouseup', stopDrag)
  window.removeEventListener('touchmove', onDragging)
  window.removeEventListener('touchend', stopDrag)
}

function onTrackMouseDown(e: MouseEvent) {
  if (draggingType) return
  const pct = getPctFromEvent(e)
  const ms = Math.round((pct / 100) * duration.value)
  editorStore.setCurrentTime(ms)
}
</script>

<style scoped>
.timeline-section {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: #FFFFFF;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

.timeline-header-area {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.timeline-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.header-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-badge-icon {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  background: var(--primary-light, #E6F4FF);
  color: var(--primary, #1677FF);
  display: flex;
  align-items: center;
  justify-content: center;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1E293B);
}

.header-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary, #64748B);
}

.header-summary strong {
  color: var(--text-primary, #1E293B);
  font-family: var(--font-mono, monospace);
}

.summary-sep {
  color: var(--text-muted, #94A3B8);
}

.duration-badge {
  background: var(--primary-light, #E6F4FF);
  color: var(--primary, #1677FF);
  padding: 2px 8px;
  border-radius: 4px;
}

.duration-badge strong {
  color: var(--primary, #1677FF);
}

.output-badge {
  background: #F0FDF4;
  color: #15803D;
  padding: 2px 8px;
  border-radius: 4px;
}

.output-badge strong {
  color: #15803D;
}

.tool-subpanel {
  margin-bottom: 2px;
}

/* Timeline Scrubber Track */
.timeline-track-wrap {
  position: relative;
  height: 74px;
  background: #0F172A;
  border-radius: var(--radius-sm, 6px);
  overflow: visible;
  user-select: none;
  cursor: pointer;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* Time Markers Ruler */
.time-markers-ruler {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 18px;
  background: rgba(15, 23, 42, 0.95);
  border-top-left-radius: var(--radius-sm, 6px);
  border-top-right-radius: var(--radius-sm, 6px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  pointer-events: none;
  z-index: 2;
}

.time-marker-tick {
  position: absolute;
  top: 0;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
}

.marker-label {
  font-size: 10px;
  color: #94A3B8;
  font-family: var(--font-mono, monospace);
  line-height: 14px;
}

.marker-bar {
  width: 1px;
  height: 4px;
  background: #64748B;
}

/* Filmstrip frames */
.filmstrip-bg {
  position: absolute;
  top: 18px;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  overflow: hidden;
  border-bottom-left-radius: var(--radius-sm, 6px);
  border-bottom-right-radius: var(--radius-sm, 6px);
}

.film-frame {
  flex: 1;
  height: 100%;
  background-size: cover;
  background-position: center;
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  opacity: 0.6;
}

/* Timeline Clips Layer */
.timeline-clips-layer {
  position: absolute;
  top: 18px;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 4;
}

.timeline-clip-block {
  position: absolute;
  top: 0;
  bottom: 0;
  pointer-events: auto;
  cursor: pointer;
  box-sizing: border-box;
  border-right: 2px dashed rgba(255, 255, 255, 0.7);
  background: transparent;
  transition: all 0.15s ease;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 4px 6px;
}

.timeline-clip-block:hover {
  background: rgba(255, 255, 255, 0.08);
}

.timeline-clip-block--selected {
  border-top: 2px solid #38BDF8;
  border-bottom: 2px solid #38BDF8;
  background: rgba(255, 255, 255, 0.04);
}

.clip-badges-wrap {
  display: flex;
  align-items: center;
  gap: 4px;
}

.clip-index-tag {
  font-size: 10px;
  font-weight: 700;
  color: #FFFFFF;
  background: rgba(0, 0, 0, 0.6);
  padding: 1px 4px;
  border-radius: 2px;
  font-family: var(--font-mono, monospace);
}

.clip-badge-pill {
  font-size: 9px;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 2px;
  font-family: var(--font-mono, monospace);
}

.clip-badge-speed {
  background: #2563EB;
  color: #FFFFFF;
}

.clip-badge-filter {
  background: #D97706;
  color: #FFFFFF;
}

.clip-delete-btn {
  background: rgba(239, 68, 68, 0.9);
  color: #FFFFFF;
  border: none;
  border-radius: 50%;
  width: 16px;
  height: 16px;
  font-size: 12px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: transform 0.15s ease;
}

.clip-delete-btn:hover {
  transform: scale(1.15);
  background: #DC2626;
}

/* Selection Range (Draggable Frame) */
.selection-range {
  position: absolute;
  top: 18px;
  bottom: 0;
  background: rgba(22, 119, 255, 0.28);
  border-top: 2px solid var(--primary, #1677FF);
  border-bottom: 2px solid var(--primary, #1677FF);
  pointer-events: auto;
  cursor: grab;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  z-index: 6;
  user-select: none;
  transition: background 0.15s ease;
}

.selection-range:hover {
  background: rgba(22, 119, 255, 0.38);
}

.selection-range--dragging {
  cursor: grabbing !important;
  background: rgba(22, 119, 255, 0.48) !important;
}

.range-move-grip {
  color: rgba(255, 255, 255, 0.85);
  font-size: 14px;
  letter-spacing: -2px;
  font-weight: 900;
  user-select: none;
  pointer-events: none;
}

.range-duration-tag {
  background: rgba(22, 119, 255, 0.95);
  color: #FFFFFF;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: var(--font-mono, monospace);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
  pointer-events: none;
}

/* Drag Handles */
.timeline-handle {
  position: absolute;
  top: 18px;
  bottom: 0;
  width: 14px;
  background: var(--primary, #1677FF);
  cursor: ew-resize;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
  transition: background 0.15s ease;
}

.timeline-handle:hover {
  background: var(--primary-hover, #4096FF);
}

.handle-left {
  transform: translateX(0);
  border-top-left-radius: 4px;
  border-bottom-left-radius: 4px;
}

.handle-right {
  transform: translateX(-100%);
  border-top-right-radius: 4px;
  border-bottom-right-radius: 4px;
}

.handle-grip {
  width: 2px;
  height: 14px;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 1px;
}

.handle-tooltip {
  position: absolute;
  bottom: calc(100% + 4px);
  left: 50%;
  transform: translateX(-50%);
  background: #1E293B;
  color: #FFFFFF;
  font-size: 10px;
  padding: 2px 5px;
  border-radius: 3px;
  white-space: nowrap;
  font-family: var(--font-mono, monospace);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.15s ease;
  z-index: 25;
}

.timeline-handle:hover .handle-tooltip {
  opacity: 1;
}

/* Playhead */
.timeline-playhead {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 2px;
  background: #EF4444;
  z-index: 12;
  cursor: ew-resize;
  transform: translateX(-50%);
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.6);
}

.playhead-cap {
  position: absolute;
  top: -6px;
  left: 50%;
  transform: translateX(-50%);
  color: #EF4444;
  display: flex;
  align-items: center;
  justify-content: center;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.4));
}

.playhead-line {
  width: 100%;
  height: 100%;
  background: #EF4444;
}

/* Controls Footer */
.timeline-controls-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 16px;
  padding-top: 4px;
}

.inputs-group {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-wrap: wrap;
}

.time-input-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.input-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #64748B);
}

.input-box {
  display: flex;
  align-items: center;
  background: #FFFFFF;
  border: 1px solid var(--border, #E2E8F0);
  border-radius: var(--radius-sm, 6px);
  padding: 2px 4px;
  height: 36px;
}

.field-input {
  width: 85px;
  border: none;
  background: transparent;
  font-size: 13px;
  font-family: var(--font-mono, monospace);
  font-weight: 600;
  color: var(--text-primary, #1E293B);
  text-align: center;
  padding: 0 4px;
}

.field-input:focus {
  outline: none;
  box-shadow: none;
}

.step-buttons {
  display: flex;
  flex-direction: column;
  gap: 1px;
}

.step-btn {
  background: transparent;
  border: none;
  color: var(--text-muted, #94A3B8);
  font-size: 8px;
  padding: 1px 3px;
  cursor: pointer;
  line-height: 1;
}

.step-btn:hover {
  color: var(--text-primary, #1E293B);
}

.input-box--readonly {
  background: var(--surface-secondary, #F8FAFC);
  padding: 0 10px;
}

.field-readonly {
  font-size: 13px;
  font-family: var(--font-mono, monospace);
  font-weight: 600;
  color: var(--primary, #1677FF);
}

.actions-group {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
