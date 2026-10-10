<template>
  <div class="video-info">
    <div class="video-info__thumb">
      <img v-if="video.thumbnail_url" :src="video.thumbnail_url" :alt="video.title" />
      <div v-else class="video-info__thumb-placeholder">
        <span>No thumbnail</span>
      </div>
    </div>
    <div class="video-info__meta">
      <PlatformBadge :platform="video.platform" />
      <h2 class="video-info__title">{{ video.title || 'Untitled video' }}</h2>
      <div class="video-info__details">
        <span v-if="video.duration_ms">{{ formatMs(video.duration_ms) }}</span>
        <span v-if="video.width && video.height">{{ video.width }}×{{ video.height }}</span>
        <span>{{ video.status }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Video } from '@/types'
import PlatformBadge from './PlatformBadge.vue'
import { formatMs } from '@/utils/format'

defineProps<{ video: Video }>()
</script>

<style scoped>
.video-info { display: flex; gap: 16px; align-items: flex-start; }
.video-info__thumb { width: 160px; flex-shrink: 0; }
.video-info__thumb img { width: 100%; border-radius: var(--radius); aspect-ratio: 16/9; object-fit: cover; }
.video-info__thumb-placeholder {
  width: 100%; aspect-ratio: 16/9;
  background: var(--surface2); border-radius: var(--radius);
  display: flex; align-items: center; justify-content: center;
  color: var(--text-muted); font-size: 12px;
}
.video-info__meta { flex: 1; display: flex; flex-direction: column; gap: 8px; }
.video-info__title { font-size: 16px; font-weight: 600; line-height: 1.4; }
.video-info__details { display: flex; gap: 12px; color: var(--text-muted); font-size: 13px; }
</style>
