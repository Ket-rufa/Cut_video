<template>
  <div class="speed-panel" v-if="selectedClip">
    <div class="panel-header">
      <div class="header-left">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10"/>
          <polyline points="12 6 12 12 16 14"/>
        </svg>
        <span class="panel-title">Tốc độ phát — Đoạn {{ selectedClipIndex + 1 }}</span>
      </div>
      <div class="header-right">
        <button class="reset-link" @click="handleReset" title="Đặt lại tốc độ về 1.0x">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="1 4 1 10 7 10"/>
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
          </svg>
          Đặt lại (1.0x)
        </button>
      </div>
    </div>

    <!-- Presets list -->
    <div class="speed-presets">
      <button
        v-for="preset in presets"
        :key="preset"
        class="preset-btn"
        :class="{ 'preset-btn--active': currentSpeed === preset }"
        @click="setSpeed(preset)"
      >
        {{ preset }}x
      </button>
    </div>

    <!-- Continuous Slider -->
    <div class="speed-slider-wrap">
      <span class="slider-min">0.25x</span>
      <input
        type="range"
        min="0.25"
        max="4.0"
        step="0.05"
        :value="currentSpeed"
        @input="onSliderChange"
        class="speed-range-slider"
      />
      <span class="slider-max">4.0x</span>
      <span class="current-speed-tag">{{ currentSpeed.toFixed(2) }}x</span>
    </div>

    <!-- Duration Impact Info -->
    <div class="duration-impact-info">
      <span class="impact-item">Thời lượng gốc: <strong>{{ formatDurationSec(sourceDurationMs) }}</strong></span>
      <span class="impact-sep">➔</span>
      <span class="impact-item impact-item--highlight">
        Sau khi đổi tốc độ: <strong>{{ formatDurationSec(outputDurationMs) }}</strong>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useEditorStore } from '@/stores/editor.store'
import { calculateOutputDuration } from '@/utils/filter'
import { formatDurationSec } from '@/utils/format'

const editorStore = useEditorStore()

const presets = [0.5, 0.75, 1.0, 1.25, 1.5, 2.0]

const selectedClip = computed(() => editorStore.selectedTimelineClip)
const selectedClipIndex = computed(() => {
  if (!selectedClip.value) return 0
  return editorStore.timelineClips.findIndex((c: any) => c.id === selectedClip.value?.id)
})

const currentSpeed = computed(() => selectedClip.value?.speed ?? 1.0)
const sourceDurationMs = computed(() => {
  if (!selectedClip.value) return 0
  return Math.max(0, selectedClip.value.sourceEndMs - selectedClip.value.sourceStartMs)
})
const outputDurationMs = computed(() => {
  if (!selectedClip.value) return 0
  return calculateOutputDuration(
    selectedClip.value.sourceStartMs,
    selectedClip.value.sourceEndMs,
    currentSpeed.value
  )
})

function setSpeed(speed: number) {
  if (selectedClip.value) {
    editorStore.setClipSpeed(selectedClip.value.id, speed)
  }
}

function onSliderChange(e: Event) {
  const target = e.target as HTMLInputElement
  setSpeed(Number(target.value))
}

function handleReset() {
  if (selectedClip.value) {
    editorStore.resetClipSpeed(selectedClip.value.id)
  }
}
</script>

<style scoped>
.speed-panel {
  padding: 12px 16px;
  background: #FFFFFF;
  border: 1px solid var(--border, #E2E8F0);
  border-radius: var(--radius-sm, 6px);
  display: flex;
  flex-direction: column;
  gap: 12px;
  animation: slideDown 0.15s ease;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #1E293B);
}

.reset-link {
  background: transparent;
  color: var(--text-secondary, #64748B);
  font-size: 12px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  border-radius: 4px;
  padding: 2px 6px;
  transition: all 0.15s ease;
}

.reset-link:hover {
  color: var(--primary, #1677FF);
  background: #F0F7FF;
}

.speed-presets {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.preset-btn {
  height: 30px;
  padding: 0 12px;
  border-radius: 4px;
  background: var(--surface-hover, #F8FAFC);
  border: 1px solid var(--border, #E2E8F0);
  color: var(--text-primary, #1E293B);
  font-size: 12px;
  font-weight: 600;
  font-family: var(--font-mono, monospace);
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-btn:hover {
  border-color: #B5D5FE;
  background: #F0F7FF;
  color: var(--primary, #1677FF);
}

.preset-btn--active {
  background: var(--primary, #1677FF) !important;
  border-color: var(--primary, #1677FF) !important;
  color: #FFFFFF !important;
}

.speed-slider-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.slider-min, .slider-max {
  font-size: 11px;
  color: var(--text-muted, #64748B);
  font-family: var(--font-mono, monospace);
}

.speed-range-slider {
  flex: 1;
  height: 5px;
  accent-color: var(--primary, #1677FF);
  cursor: pointer;
}

.current-speed-tag {
  min-width: 48px;
  text-align: right;
  font-size: 13px;
  font-weight: 700;
  font-family: var(--font-mono, monospace);
  color: var(--primary, #1677FF);
}

.duration-impact-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary, #64748B);
  background: var(--surface-hover, #F8FAFC);
  padding: 6px 10px;
  border-radius: 4px;
  border: 1px solid #EEF2F6;
}

.impact-sep {
  color: var(--text-muted, #94A3B8);
}

.impact-item--highlight strong {
  color: var(--primary, #1677FF);
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
