<template>
  <div class="editor-workspace">
    <!-- Main Workspace Top Grid: Player 45%, Info 30%, Clips 25% -->
    <div class="workspace-grid">
      <!-- Left / Center: Video Player (45%) -->
      <div class="workspace-col workspace-col--player">
        <VideoPlayer
          ref="playerRef"
          :src="videoStreamUrl"
          :initial-duration-ms="editorStore.duration"
          :seek-ms="editorStore.currentTime"
          @time-update="onTimeUpdate"
          @duration-loaded="onDurationLoaded"
        />
      </div>

      <!-- Center: Video Information (30%) -->
      <div class="workspace-col workspace-col--info">
        <VideoInfoCard :video="editorStore.currentVideo" />
      </div>

      <!-- Right: Clip List Panel (25%) -->
      <div class="workspace-col workspace-col--clips">
        <ClipListPanel />
      </div>
    </div>

    <!-- Bottom: Video Timeline & Cutter Section -->
    <div class="workspace-bottom">
      <VideoTimeline />

      <!-- AI Suggestions Section Below Timeline -->
      <AISuggestions />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useEditorStore } from '@/stores/editor.store'
import { useVideoStore } from '@/stores/video.store'
import VideoPlayer from '@/components/video/VideoPlayer.vue'
import VideoInfoCard from '@/components/video/VideoInfoCard.vue'
import ClipListPanel from '@/components/video/ClipListPanel.vue'
import VideoTimeline from '@/components/video/VideoTimeline.vue'
import AISuggestions from '@/components/video/AISuggestions.vue'

const route = useRoute()
const editorStore = useEditorStore()
const videoStore = useVideoStore()

const playerRef = ref<InstanceType<typeof VideoPlayer> | null>(null)

const videoStreamUrl = computed(() => {
  if (editorStore.currentVideo.id) {
    return `/api/v1/videos/${editorStore.currentVideo.id}/file`
  }
  if (editorStore.currentVideo.source_url && editorStore.currentVideo.source_url.startsWith('http')) {
    return editorStore.currentVideo.source_url
  }
  return 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4'
})

onMounted(async () => {
  const id = route.params.id as string
  if (id) {
    try {
      const v = await videoStore.fetchById(id)
      editorStore.setVideo(v)
      await editorStore.loadSubtitles(id)
    } catch {
      // Fallback already handled in store
    }
  }
})

function onTimeUpdate(ms: number) {
  editorStore.setCurrentTime(ms)
}

function onDurationLoaded(ms: number) {
  editorStore.setDuration(ms)
}
</script>

<style scoped>
.editor-workspace {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
}

/* 3-Column Grid: Proportional fr layout */
.workspace-grid {
  display: grid;
  grid-template-columns: minmax(0, 4.4fr) minmax(0, 3fr) minmax(0, 2.6fr);
  gap: 20px;
  align-items: stretch;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

.workspace-col {
  display: flex;
  flex-direction: column;
}

.workspace-col--player {
  min-width: 0;
}

.workspace-col--info {
  min-width: 0;
}

.workspace-col--clips {
  min-width: 0;
}

/* Bottom Stack */
.workspace-bottom {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
  max-width: 100%;
  box-sizing: border-box;
}

/* Responsive Breakdown */
@media (max-width: 1200px) {
  .workspace-grid {
    grid-template-columns: 1fr 1fr;
  }
  .workspace-col--clips {
    grid-column: span 2;
  }
}

@media (max-width: 768px) {
  .workspace-grid {
    grid-template-columns: 1fr;
  }
  .workspace-col--clips {
    grid-column: span 1;
  }
}
</style>
