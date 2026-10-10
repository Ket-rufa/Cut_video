<template>
  <div class="filter-panel" v-if="selectedClip">
    <div class="panel-header">
      <!-- Tab Switcher -->
      <div class="tab-switcher">
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': currentTab === 'presets' }"
          @click="currentTab = 'presets'"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2a10 10 0 0 1 0 20z"/>
          </svg>
          Bộ lọc màu
        </button>
        <button
          class="tab-btn"
          :class="{ 'tab-btn--active': currentTab === 'adjust' }"
          @click="currentTab = 'adjust'"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="4" y1="21" x2="4" y2="14"/>
            <line x1="4" y1="10" x2="4" y2="3"/>
            <line x1="12" y1="21" x2="12" y2="12"/>
            <line x1="12" y1="8" x2="12" y2="3"/>
            <line x1="20" y1="21" x2="20" y2="16"/>
            <line x1="20" y1="12" x2="20" y2="3"/>
            <line x1="1" y1="14" x2="7" y2="14"/>
            <line x1="9" y1="8" x2="15" y2="8"/>
            <line x1="17" y1="16" x2="23" y2="16"/>
          </svg>
          Điều chỉnh
        </button>
      </div>

      <!-- Reset Action -->
      <div class="header-right">
        <button class="reset-link" @click="handleReset" title="Đặt lại màu sắc về mặc định">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="1 4 1 10 7 10"/>
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
          </svg>
          Đặt lại màu
        </button>
      </div>
    </div>

    <!-- Tab 1: Presets Cards -->
    <div v-if="currentTab === 'presets'" class="presets-gallery">
      <button
        v-for="preset in FILTER_PRESETS"
        :key="preset.id"
        class="preset-card"
        :class="{ 'preset-card--active': currentFilter.preset === preset.id }"
        @click="selectPreset(preset.id)"
      >
        <div class="preset-swatch" :style="preset.previewStyle">
          <span v-if="currentFilter.preset === preset.id" class="check-icon">✓</span>
        </div>
        <span class="preset-name">{{ preset.name }}</span>
      </button>
    </div>

    <!-- Tab 2: Manual Adjustments Sliders -->
    <div v-else class="adjustments-form">
      <!-- Brightness -->
      <div class="adj-row">
        <div class="adj-info">
          <span class="adj-label">Độ sáng (Brightness)</span>
          <span class="adj-val">{{ formatVal(currentFilter.brightness) }}</span>
        </div>
        <input
          type="range"
          min="-100"
          max="100"
          step="1"
          :value="currentFilter.brightness"
          @input="onAdjustChange('brightness', $event)"
          class="adj-slider"
        />
      </div>

      <!-- Contrast -->
      <div class="adj-row">
        <div class="adj-info">
          <span class="adj-label">Độ tương phản (Contrast)</span>
          <span class="adj-val">{{ formatVal(currentFilter.contrast) }}</span>
        </div>
        <input
          type="range"
          min="-100"
          max="100"
          step="1"
          :value="currentFilter.contrast"
          @input="onAdjustChange('contrast', $event)"
          class="adj-slider"
        />
      </div>

      <!-- Saturation -->
      <div class="adj-row">
        <div class="adj-info">
          <span class="adj-label">Độ bão hòa (Saturation)</span>
          <span class="adj-val">{{ formatVal(currentFilter.saturation) }}</span>
        </div>
        <input
          type="range"
          min="-100"
          max="100"
          step="1"
          :value="currentFilter.saturation"
          @input="onAdjustChange('saturation', $event)"
          class="adj-slider"
        />
      </div>

      <!-- Temperature -->
      <div class="adj-row">
        <div class="adj-info">
          <span class="adj-label">Nhiệt độ màu (Temperature)</span>
          <span class="adj-val">{{ formatVal(currentFilter.temperature) }}</span>
        </div>
        <input
          type="range"
          min="-100"
          max="100"
          step="1"
          :value="currentFilter.temperature"
          @input="onAdjustChange('temperature', $event)"
          class="adj-slider adj-slider--temp"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useEditorStore } from '@/stores/editor.store'
import { FILTER_PRESETS, DEFAULT_CLIP_FILTER } from '@/utils/filter'
import type { FilterPresetType, ClipFilterConfig } from '@/types'

const editorStore = useEditorStore()

const currentTab = ref<'presets' | 'adjust'>('presets')

const selectedClip = computed(() => editorStore.selectedTimelineClip)
const currentFilter = computed<ClipFilterConfig>(() => {
  return selectedClip.value?.filter || { ...DEFAULT_CLIP_FILTER }
})

function selectPreset(presetId: FilterPresetType) {
  if (selectedClip.value) {
    editorStore.setClipFilterPreset(selectedClip.value.id, presetId)
  }
}

function onAdjustChange(field: keyof ClipFilterConfig, e: Event) {
  if (!selectedClip.value) return
  const val = Number((e.target as HTMLInputElement).value)
  editorStore.setClipFilterAdjustment(selectedClip.value.id, { [field]: val })
}

function handleReset() {
  if (selectedClip.value) {
    editorStore.resetClipFilter(selectedClip.value.id)
  }
}

function formatVal(val: number): string {
  if (val > 0) return `+${val}`
  return `${val}`
}
</script>

<style scoped>
.filter-panel {
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
  border-bottom: 1px solid var(--border-subtle, #F1F5F9);
  padding-bottom: 8px;
}

.tab-switcher {
  display: flex;
  align-items: center;
  gap: 4px;
  background: var(--surface-hover, #F8FAFC);
  padding: 2px;
  border-radius: 6px;
  border: 1px solid var(--border, #E2E8F0);
}

.tab-btn {
  height: 28px;
  padding: 0 10px;
  border-radius: 4px;
  background: transparent;
  border: none;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #64748B);
  display: inline-flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: var(--text-primary, #1E293B);
}

.tab-btn--active {
  background: #FFFFFF;
  color: var(--primary, #1677FF);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
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

/* Presets Gallery */
.presets-gallery {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(68px, 1fr));
  gap: 8px;
}

.preset-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  padding: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-card:hover {
  background: var(--surface-hover, #F8FAFC);
  border-color: #E2E8F0;
}

.preset-card--active {
  border-color: var(--primary, #1677FF) !important;
  background: #F0F7FF !important;
}

.preset-swatch {
  width: 100%;
  height: 42px;
  border-radius: 4px;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: inset 0 0 2px rgba(0, 0, 0, 0.2);
}

.check-icon {
  color: #FFFFFF;
  font-size: 14px;
  font-weight: 800;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.6);
}

.preset-name {
  font-size: 11px;
  font-weight: 500;
  color: var(--text-primary, #1E293B);
  text-align: center;
}

.preset-card--active .preset-name {
  color: var(--primary, #1677FF);
  font-weight: 600;
}

/* Adjustments Form */
.adjustments-form {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 20px;
}

@media (max-width: 768px) {
  .adjustments-form {
    grid-template-columns: 1fr;
  }
}

.adj-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.adj-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
}

.adj-label {
  color: var(--text-secondary, #64748B);
  font-weight: 500;
}

.adj-val {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
  font-weight: 600;
  color: var(--primary, #1677FF);
}

.adj-slider {
  width: 100%;
  height: 4px;
  accent-color: var(--primary, #1677FF);
  cursor: pointer;
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
