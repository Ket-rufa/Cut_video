<template>
  <div class="app-layout">
    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Right Content Column -->
    <div class="app-layout__body">
      <!-- Header -->
      <AppHeader @open-add-video="videoStore.openAddModal()" />

      <!-- Social Platform Filter Chips (home & list views) -->
      <SocialTabs v-if="showSocialTabs" />

      <!-- Main Content / Page View -->
      <main class="app-layout__content">
        <slot>
          <RouterView />
        </slot>
      </main>
    </div>

    <!-- Global Add Video Modal -->
    <AddVideoModal
      v-model="videoStore.isAddModalOpen"
      :initial-url="videoStore.addModalInitialUrl"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import SocialTabs from '@/components/layout/SocialTabs.vue'
import AddVideoModal from '@/components/video/AddVideoModal.vue'
import { useVideoStore } from '@/stores/video.store'

const route = useRoute()
const videoStore = useVideoStore()

const showSocialTabs = computed(() => {
  return route.path === '/' || route.path === '/saved' || route.path === '/social'
})
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  width: 100%;
  max-width: 100vw;
  background: var(--bg);
  overflow-x: hidden;
}

.app-layout__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0; /* Prevent flex overflow */
  max-width: 100%;
  overflow-x: hidden;
}

.app-layout__content {
  flex: 1;
  padding: 20px 24px;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}
</style>
