<template>
  <div class="editor-toolbar">
    <!-- Trim Mode Toggle -->
    <button
      class="toolbar-btn"
      :class="{ 'toolbar-btn--active': activeTool === 'trim' }"
      @click="toggleTool('trim')"
      title="Cắt / Điều chỉnh vùng chọn đoạn video"
    >
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M6 3v18"/>
        <path d="M18 3v18"/>
        <rect x="6" y="8" width="12" height="8" rx="1"/>
      </svg>
      <span>Cắt / Trim</span>
    </button>

    <!-- Split Button -->
    <button
      class="toolbar-btn toolbar-btn--action"
      @click="handleSplit"
      title="Tách đoạn tại vị trí con trỏ (Ctrl+B)"
    >
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="6" cy="6" r="3"/>
        <circle cx="6" cy="18" r="3"/>
        <line x1="20" y1="4" x2="8.12" y2="15.88"/>
        <line x1="14.47" y1="14.48" x2="20" y2="20"/>
        <line x1="8.12" y1="8.12" x2="12" y2="12"/>
      </svg>
      <span>Tách</span>
      <kbd class="shortcut-kbd">Ctrl+B</kbd>
    </button>

    <!-- Speed Tool Toggle -->
    <button
      class="toolbar-btn"
      :class="{ 'toolbar-btn--active': activeTool === 'speed' }"
      :disabled="!hasSelectedClip"
      @click="toggleTool('speed')"
      :title="hasSelectedClip ? 'Điều chỉnh tốc độ phát' : 'Hãy chọn một đoạn video để chỉnh sửa'"
    >
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <polyline points="12 6 12 12 16 14"/>
      </svg>
      <span>Tốc độ</span>
      <span v-if="currentClipSpeed !== 1" class="tool-badge">{{ currentClipSpeed }}x</span>
    </button>

    <!-- Filter Tool Toggle -->
    <button
      class="toolbar-btn"
      :class="{ 'toolbar-btn--active': activeTool === 'filter' }"
      :disabled="!hasSelectedClip"
      @click="toggleTool('filter')"
      :title="hasSelectedClip ? 'Điều chỉnh màu sắc đoạn video' : 'Hãy chọn một đoạn video để chỉnh sửa'"
    >
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <path d="M12 2a10 10 0 0 1 0 20z" fill="currentColor"/>
      </svg>
      <span>Bộ lọc</span>
      <span v-if="hasActiveFilter" class="tool-badge-dot">●</span>
    </button>

    <!-- Subtitle AI Tool Toggle -->
    <button
      class="toolbar-btn toolbar-btn--ai"
      :class="{ 'toolbar-btn--active': activeTool === 'subtitle' }"
      @click="toggleTool('subtitle')"
      title="Tạo phụ đề tự động bằng AI Gemini (Speech-to-Text & Dịch)"
    >
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        <line x1="9" y1="10" x2="15" y2="10"/>
        <line x1="12" y1="7" x2="12" y2="13"/>
      </svg>
      <span>Phụ đề AI</span>
      <span v-if="hasSubtitles" class="tool-badge-dot tool-badge-dot--purple">●</span>
    </button>

    <!-- Quick Feedback Alert -->
    <div v-if="splitToast" class="split-toast" :class="{ 'split-toast--error': splitToast.isError }">
      {{ splitToast.msg }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useEditorStore } from '@/stores/editor.store'

const editorStore = useEditorStore()

const splitToast = ref<{ msg: string; isError?: boolean } | null>(null)
let toastTimer: number | null = null

const activeTool = computed(() => editorStore.activeTool)
const hasSelectedClip = computed(() => !!editorStore.selectedTimelineClip)
const currentClipSpeed = computed(() => editorStore.selectedTimelineClip?.speed ?? 1.0)
const hasSubtitles = computed(() => !!editorStore.subtitles && !!editorStore.subtitles.cues && editorStore.subtitles.cues.length > 0)
const hasActiveFilter = computed(() => {
  const f = editorStore.selectedTimelineClip?.filter
  if (!f) return false
  return f.preset !== 'none' || f.brightness !== 0 || f.contrast !== 0 || f.saturation !== 0 || f.temperature !== 0
})

function toggleTool(tool: 'trim' | 'split' | 'speed' | 'filter' | 'subtitle') {
  editorStore.setActiveTool(tool)
}

function handleSplit() {
  const res = editorStore.splitClipAtCurrentTime()
  if (res.success) {
    showToast('Đã tách đoạn tại vị trí playhead!', false)
  } else {
    showToast(res.message || 'Không thể tách đoạn', true)
  }
}

function showToast(msg: string, isError: boolean = false) {
  if (toastTimer) clearTimeout(toastTimer)
  splitToast.value = { msg, isError }
  toastTimer = window.setTimeout(() => {
    splitToast.value = null
  }, 2200)
}
</script>

<style scoped>
.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: var(--surface-hover, #F8FAFC);
  border: 1px solid var(--border, #E2E8F0);
  border-radius: var(--radius-sm, 6px);
  position: relative;
  flex-wrap: wrap;
}

.toolbar-btn {
  height: 34px;
  padding: 0 14px;
  border-radius: 6px;
  background: #FFFFFF;
  border: 1px solid var(--border, #E2E8F0);
  color: var(--text-primary, #1E293B);
  font-size: 13px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.toolbar-btn:hover:not(:disabled) {
  border-color: #B5D5FE;
  background: #F0F7FF;
  color: var(--primary, #1677FF);
}

.toolbar-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
  background: #F1F5F9;
  border-color: #E2E8F0;
  color: #94A3B8;
}

.toolbar-btn--active {
  background: var(--primary, #1677FF) !important;
  border-color: var(--primary, #1677FF) !important;
  color: #FFFFFF !important;
  box-shadow: 0 1px 3px rgba(22, 119, 255, 0.3);
}

.toolbar-btn--action:hover {
  background: #F0FDF4;
  border-color: #86EFAC;
  color: #15803D;
}

.shortcut-kbd {
  font-size: 10px;
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-muted, #64748B);
  padding: 1px 4px;
  border-radius: 3px;
  font-family: var(--font-mono, monospace);
  font-weight: 600;
  margin-left: 2px;
}

.toolbar-btn--active .shortcut-kbd {
  background: rgba(255, 255, 255, 0.25);
  color: #FFFFFF;
}

.tool-badge {
  font-size: 11px;
  font-weight: 700;
  background: #EFF6FF;
  color: var(--primary, #1677FF);
  padding: 1px 5px;
  border-radius: 4px;
}

.toolbar-btn--active .tool-badge {
  background: rgba(255, 255, 255, 0.25);
  color: #FFFFFF;
}

.tool-badge-dot {
  font-size: 8px;
  color: #F59E0B;
}

.tool-badge-dot--purple {
  color: #8B5CF6;
}

.toolbar-btn--ai {
  border-color: rgba(139, 92, 246, 0.3);
}

.toolbar-btn--ai:hover {
  background: rgba(139, 92, 246, 0.08);
  border-color: #8B5CF6;
  color: #6D28D9;
}

.toolbar-btn--ai.toolbar-btn--active {
  background: #8B5CF6;
  border-color: #7C3AED;
  color: #FFFFFF;
}

.split-toast {
  position: absolute;
  right: 12px;
  top: 50%;
  transform: translateY(-50%);
  background: #10B981;
  color: #FFFFFF;
  font-size: 12px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  animation: fadeIn 0.2s ease;
  z-index: 20;
}

.split-toast--error {
  background: #EF4444;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-40%); }
  to { opacity: 1; transform: translateY(-50%); }
}
</style>
