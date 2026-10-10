import { createRouter, createWebHistory } from 'vue-router'
import HomePage from '@/pages/HomePage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomePage,
    },
    {
      path: '/editor',
      name: 'editor-default',
      component: () => import('@/pages/EditorPage.vue'),
    },
    {
      path: '/editor/:id',
      name: 'editor',
      component: () => import('@/pages/EditorPage.vue'),
    },
    {
      path: '/saved',
      name: 'saved',
      component: () => import('@/pages/SavedVideosPage.vue'),
    },
    {
      path: '/library',
      name: 'library',
      component: () => import('@/pages/LibraryPage.vue'),
    },
    {
      path: '/social',
      name: 'social',
      component: () => import('@/pages/SocialPage.vue'),
    },
    {
      path: '/history',
      name: 'history',
      component: () => import('@/pages/HistoryPage.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/pages/SettingsPage.vue'),
    },
    {
      path: '/download/:jobId',
      name: 'download',
      component: () => import('@/pages/DownloadPage.vue'),
    },
  ],
})

export default router
