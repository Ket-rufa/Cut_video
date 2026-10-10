<template>
  <div class="ai-suggestions-card card-base">
    <!-- Header -->
    <div class="ai-header">
      <div class="ai-badge-title">
        <span class="ai-icon">✨</span>
        <div class="title-group">
          <h4 class="ai-title">Các đoạn đề xuất (AI)</h4>
          <span class="ai-subtitle">Gợi ý các khoảnh khắc nổi bật trong video</span>
        </div>
      </div>
      <span class="ai-tag">Tự động phát hiện</span>
    </div>

    <!-- Horizontal scrollable card list -->
    <div class="cards-scroll-track">
      <div
        v-for="item in suggestions"
        :key="item.id"
        class="suggestion-card"
        @click="applySuggestion(item)"
        title="Nhấp để đặt vùng cắt này trên timeline"
      >
        <div class="card-thumb-wrap">
          <img :src="item.thumbnail || defaultThumb" :alt="item.title" class="card-thumb" />
          <span class="duration-badge">{{ item.duration_s }}s</span>
          <div class="hover-play-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
          </div>
        </div>

        <div class="card-meta">
          <span class="card-title" :title="item.title">{{ item.title }}</span>
          <span class="card-time-range">{{ formatTimeShort(item.start_ms) }} - {{ formatTimeShort(item.end_ms) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AISuggestion } from '@/types'
import { formatTimeShort } from '@/utils/format'
import { useEditorStore } from '@/stores/editor.store'

const editorStore = useEditorStore()
const defaultThumb = 'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&w=240&q=80'

const suggestions = computed(() => editorStore.aiSuggestions)

function applySuggestion(item: AISuggestion) {
  editorStore.applyAISuggestion(item)
}
</script>

<style scoped>
.ai-suggestions-card {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: #FFFFFF;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

.ai-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.ai-badge-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ai-icon {
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.title-group {
  display: flex;
  flex-direction: column;
}

.ai-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.ai-subtitle {
  font-size: 12px;
  color: var(--text-muted);
}

.ai-tag {
  font-size: 11px;
  font-weight: 600;
  color: #7C3AED;
  background: #F3E8FF;
  padding: 3px 8px;
  border-radius: 4px;
}

/* Horizontal track */
.cards-scroll-track {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 6px;
  min-width: 0;
  max-width: 100%;
  scrollbar-width: thin;
}

.suggestion-card {
  width: 170px;
  min-width: 170px;
  background: #FFFFFF;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.suggestion-card:hover {
  border-color: var(--primary);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.card-thumb-wrap {
  position: relative;
  width: 100%;
  aspect-ratio: 16/9;
  background: #000000;
  overflow: hidden;
}

.card-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s ease;
}

.suggestion-card:hover .card-thumb {
  transform: scale(1.05);
}

.duration-badge {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.75);
  color: #FFFFFF;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 3px;
  font-family: var(--font-mono);
}

.hover-play-icon {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(22, 119, 255, 0.4);
  color: #FFFFFF;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.suggestion-card:hover .hover-play-icon {
  opacity: 1;
}

.card-meta {
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.card-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-time-range {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono);
}
</style>
