<template>
  <div class="social-tabs-container">
    <div class="social-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="['social-tab', `social-tab--${tab.id}`, { 'social-tab--active': activeTab === tab.id }]"
        :style="{
          '--platform-color': tab.color,
          '--platform-bg': tab.bgLight,
          '--platform-shadow': tab.shadowColor
        }"
        @click="selectTab(tab.id)"
      >
        <!-- Square Brand Icon Badge -->
        <span class="tab-icon-box" v-html="tab.icon"></span>
        <span class="tab-label">{{ tab.label }}</span>
        <span v-if="tab.count !== undefined" class="tab-count">{{ tab.count }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useVideoStore } from '@/stores/video.store'

const videoStore = useVideoStore()

const activeTab = computed(() => videoStore.activePlatform)

const tabs = computed(() => [
  {
    id: 'all',
    label: 'Tất cả',
    color: '#1677FF',
    bgLight: '#EFF6FF',
    shadowColor: 'rgba(22, 119, 255, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#1677FF"/>
      <rect x="7" y="7" width="5.5" height="5.5" rx="1.5" fill="white"/>
      <rect x="15.5" y="7" width="5.5" height="5.5" rx="1.5" fill="white"/>
      <rect x="7" y="15.5" width="5.5" height="5.5" rx="1.5" fill="white"/>
      <rect x="15.5" y="15.5" width="5.5" height="5.5" rx="1.5" fill="white"/>
    </svg>`,
    count: videoStore.videos.length,
  },
  {
    id: 'youtube',
    label: 'YouTube',
    color: '#FF0000',
    bgLight: '#FEF2F2',
    shadowColor: 'rgba(255, 0, 0, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#FF0000"/>
      <path d="M11 9L19 14L11 19V9Z" fill="white"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'youtube').length,
  },
  {
    id: 'tiktok',
    label: 'TikTok',
    color: '#000000',
    bgLight: '#F1F5F9',
    shadowColor: 'rgba(0, 0, 0, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#000000"/>
      <path d="M18.8 9.5c.8.6 1.7.9 2.7 1v2.5c-1-.05-2-.35-2.7-.9v5.6c0 2.7-2.2 4.8-4.8 4.8s-4.8-2.1-4.8-4.8c0-2.7 2.2-4.8 4.8-4.8.3 0 .6.03.9.1v2.6c-.3-.07-.6-.1-.9-.1-1.3 0-2.3 1-2.3 2.2s1 2.2 2.3 2.2 2.3-1 2.3-2.2V6h2.5c0 1.6 1 3 2.3 3.5z" fill="#00F2FE"/>
      <path d="M18.2 10.1c.8.6 1.7.9 2.7 1v2.5c-1-.05-2-.35-2.7-.9v5.6c0 2.7-2.2 4.8-4.8 4.8s-4.8-2.1-4.8-4.8c0-2.7 2.2-4.8 4.8-4.8.3 0 .6.03.9.1v2.6c-.3-.07-.6-.1-.9-.1-1.3 0-2.3 1-2.3 2.2s1 2.2 2.3 2.2 2.3-1 2.3-2.2V6h2.5c0 1.6 1 3 2.3 3.5z" fill="#FE2C55"/>
      <path d="M18.5 9.8c.8.6 1.7.9 2.7 1v2.5c-1-.05-2-.35-2.7-.9v5.6c0 2.7-2.2 4.8-4.8 4.8s-4.8-2.1-4.8-4.8c0-2.7 2.2-4.8 4.8-4.8.3 0 .6.03.9.1v2.6c-.3-.07-.6-.1-.9-.1-1.3 0-2.3 1-2.3 2.2s1 2.2 2.3 2.2 2.3-1 2.3-2.2V6h2.5c0 1.6 1 3 2.3 3.5z" fill="#FFFFFF"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'tiktok').length,
  },
  {
    id: 'douyin',
    label: 'Douyin (抖音)',
    color: '#161823',
    bgLight: '#F1F5F9',
    shadowColor: 'rgba(22, 24, 35, 0.25)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#161823"/>
      <path d="M19 10.2c.7.5 1.5.8 2.4.9v2.2c-.9-.05-1.8-.3-2.4-.8v4.9c0 2.4-2 4.2-4.2 4.2s-4.2-1.9-4.2-4.2 1.9-4.2 4.2-4.2c.3 0 .5.03.8.1v2.3c-.3-.06-.5-.1-.8-.1-1.1 0-2 .9-2 1.9s.9 1.9 2 1.9 2-.9 2-1.9V7.1h2.2c0 1.4.9 2.6 2 3.1z" fill="#00F2FE"/>
      <path d="M18.5 10.7c.7.5 1.5.8 2.4.9v2.2c-.9-.05-1.8-.3-2.4-.8v4.9c0 2.4-2 4.2-4.2 4.2s-4.2-1.9-4.2-4.2 1.9-4.2 4.2-4.2c.3 0 .5.03.8.1v2.3c-.3-.06-.5-.1-.8-.1-1.1 0-2 .9-2 1.9s.9 1.9 2 1.9 2-.9 2-1.9V7.1h2.2c0 1.4.9 2.6 2 3.1z" fill="#FE2C55"/>
      <path d="M18.7 10.4c.7.5 1.5.8 2.4.9v2.2c-.9-.05-1.8-.3-2.4-.8v4.9c0 2.4-2 4.2-4.2 4.2s-4.2-1.9-4.2-4.2 1.9-4.2 4.2-4.2c.3 0 .5.03.8.1v2.3c-.3-.06-.5-.1-.8-.1-1.1 0-2 .9-2 1.9s.9 1.9 2 1.9 2-.9 2-1.9V7.1h2.2c0 1.4.9 2.6 2 3.1z" fill="#FFFFFF"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'douyin').length,
  },
  {
    id: 'facebook',
    label: 'Facebook',
    color: '#1877F2',
    bgLight: '#EFF6FF',
    shadowColor: 'rgba(24, 119, 242, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#1877F2"/>
      <path d="M18 14.2h-2.6v8.8h-3.6v-8.8H9.8v-3.1h2V9c0-2.5 1.5-3.8 3.7-3.8 1.1 0 2.1.08 2.3.11v2.7h-1.6c-1.2 0-1.4.6-1.4 1.4v1.7h3l-.4 3.1z" fill="white"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'facebook').length,
  },
  {
    id: 'instagram',
    label: 'Instagram',
    color: '#E1306C',
    bgLight: '#FDF2F8',
    shadowColor: 'rgba(225, 48, 108, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <radialGradient id="ig-rg" cx="20%" cy="110%" r="130%">
          <stop offset="0%" stop-color="#FFD600"/>
          <stop offset="30%" stop-color="#FF0100"/>
          <stop offset="60%" stop-color="#D800B9"/>
          <stop offset="100%" stop-color="#405DE6"/>
        </radialGradient>
      </defs>
      <rect width="28" height="28" rx="6" fill="url(#ig-rg)"/>
      <rect x="7.5" y="7.5" width="13" height="13" rx="3.5" stroke="white" stroke-width="1.8" fill="none"/>
      <circle cx="14" cy="14" r="3.2" stroke="white" stroke-width="1.8" fill="none"/>
      <circle cx="17.6" cy="10.4" r="0.9" fill="white"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'instagram').length,
  },
  {
    id: 'twitter',
    label: 'Twitter (X)',
    color: '#0F172A',
    bgLight: '#F8FAFC',
    shadowColor: 'rgba(15, 23, 42, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#0F1419"/>
      <path d="M18.8 7.5h2.2l-4.8 5.5 5.6 7.5h-4.4l-3.5-4.5-4 4.5H7.7l5.1-5.9-5.4-7.1h4.5l3.1 4.1 3.8-4.1zm-.8 11.7h1.2L11.7 8.7h-1.3l7.6 10.5z" fill="white"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'twitter').length,
  },
  {
    id: 'generic',
    label: 'Khác',
    color: '#64748B',
    bgLight: '#F1F5F9',
    shadowColor: 'rgba(100, 116, 139, 0.2)',
    icon: `<svg width="28" height="28" viewBox="0 0 28 28" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect width="28" height="28" rx="6" fill="#64748B"/>
      <path d="M12 16a3.5 3.5 0 0 0 5 0l2.5-2.5a3.5 3.5 0 0 0-5-5L13.2 9.7M16 12a3.5 3.5 0 0 0-5 0L8.5 14.5a3.5 3.5 0 0 0 5 5l1.3-1.2" stroke="white" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"/>
    </svg>`,
    count: videoStore.videos.filter(v => v.platform === 'generic' || v.platform === 'unknown').length,
  },
])

function selectTab(id: string) {
  videoStore.setPlatformFilter(id)
}
</script>

<style scoped>
.social-tabs-container {
  padding: 10px 24px;
  background: #FFFFFF;
  border-bottom: 1px solid var(--border);
  overflow-x: auto;
}

.social-tabs {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: max-content;
}

/* Square / Rectangular Tab Style */
.social-tab {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 6px 12px 6px 7px;
  background: #FFFFFF;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-sm); /* Hình vuông bo nhẹ góc 8px (SaaS standard) */
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 600;
  box-shadow: var(--shadow-xs);
  transition: all 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  user-select: none;
  cursor: pointer;
}

.social-tab:hover {
  border-color: #CBD5E1;
  background: #F8FAFC;
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* Square Icon Box */
.tab-icon-box {
  width: 28px;
  height: 28px;
  border-radius: 6px; /* Vuông vắn bo góc 6px */
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
  transition: transform 0.18s ease;
}

.social-tab:hover .tab-icon-box {
  transform: scale(1.04);
}

.tab-label {
  font-weight: 600;
  letter-spacing: -0.01em;
  color: #1E293B;
  transition: color 0.18s ease;
}

/* Square-rounded count badge */
.tab-count {
  font-size: 11px;
  background: #F1F5F9;
  color: #64748B;
  padding: 2px 7px;
  border-radius: 4px; /* Hình vuông góc nhẹ */
  font-weight: 700;
  font-family: var(--font-mono);
  transition: all 0.18s ease;
}

/* Tab active: Viền & nền theo màu nhận diện thương hiệu của từng mạng xã hội */
.social-tab--active {
  border-color: var(--platform-color) !important;
  background: var(--platform-bg) !important;
  box-shadow: 0 2px 8px var(--platform-shadow) !important;
}

.social-tab--active .tab-label {
  color: #0F172A;
  font-weight: 700;
}

.social-tab--active .tab-count {
  background: var(--platform-color);
  color: #FFFFFF;
}

.social-tab--active .tab-icon-box {
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.18);
  transform: scale(1.06);
}

/* Dark mode adjustments */
[data-theme="dark"] .social-tabs-container {
  background: #1E293B;
}

[data-theme="dark"] .social-tab {
  background: #0F172A;
  border-color: #334155;
  color: #F8FAFC;
}

[data-theme="dark"] .social-tab .tab-label {
  color: #F8FAFC;
}

[data-theme="dark"] .social-tab--active {
  background: rgba(255, 255, 255, 0.06) !important;
}
</style>
