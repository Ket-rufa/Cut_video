<template>
  <span :class="['badge', `badge--${platform}`]">{{ label }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { VideoPlatform } from '@/types'

const props = defineProps<{ platform: VideoPlatform }>()

const label = computed(() => {
  const map: Record<VideoPlatform, string> = {
    youtube: 'YouTube',
    tiktok: 'TikTok',
    douyin: 'Douyin',
    facebook: 'Facebook',
    instagram: 'Instagram',
    twitter: 'Twitter (X)',
    generic: 'Direct URL',
    unknown: 'Unknown',
  }
  return map[props.platform] ?? props.platform
})
</script>

<style scoped>
.badge {
  display: inline-block; padding: 2px 8px;
  border-radius: 4px; font-size: 11px; font-weight: 600;
  letter-spacing: 0.02em; text-transform: uppercase;
}
.badge--youtube   { background: #ff0000; color: #fff; }
.badge--tiktok    { background: #000; color: #fff; border: 1px solid #444; }
.badge--douyin    { background: #161823; color: #fff; border: 1px solid #334155; }
.badge--facebook  { background: #1877f2; color: #fff; }
.badge--instagram { background: #e1306c; color: #fff; }
.badge--twitter   { background: #0f172a; color: #fff; }
.badge--generic   { background: var(--surface-secondary); color: var(--text-secondary); border: 1px solid var(--border); }
.badge--unknown   { background: var(--surface-secondary); color: var(--text-muted); }
</style>
