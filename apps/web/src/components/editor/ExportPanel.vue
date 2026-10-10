<template>
  <div class="export-panel">
    <h3 class="export-panel__title">Export</h3>

    <div class="export-panel__field">
      <label>Start</label>
      <span class="mono">{{ formatMs(startMs) }}</span>
    </div>
    <div class="export-panel__field">
      <label>End</label>
      <span class="mono">{{ formatMs(endMs) }}</span>
    </div>
    <div class="export-panel__field">
      <label>Duration</label>
      <span class="mono">{{ formatMs(endMs - startMs) }}</span>
    </div>

    <div class="export-panel__divider" />

    <div class="export-panel__field">
      <label>Format</label>
      <select v-model="format">
        <option value="mp4">MP4</option>
      </select>
    </div>

    <div class="export-panel__field">
      <label>Quality</label>
      <select v-model="quality">
        <option value="original">Original</option>
        <option value="1080p">1080p</option>
        <option value="720p">720p</option>
        <option value="480p">480p</option>
      </select>
    </div>

    <div class="export-panel__divider" />

    <!-- Job status -->
    <div v-if="job" class="export-panel__status">
      <div class="export-panel__field">
        <label>Status</label>
        <span :class="['status-badge', `status-badge--${job.status}`]">{{ job.status }}</span>
      </div>
      <AppProgress v-if="job.status === 'processing' || job.status === 'queued'" :value="job.progress" />
      <div v-if="job.status === 'failed' && job.error_message" class="export-panel__error">{{ job.error_message }}</div>
    </div>

    <!-- Download button -->
    <div v-if="exportReady && currentExport" class="export-panel__download">
      <a :href="downloadUrl" download class="download-link">
        ⬇ Download MP4
      </a>
    </div>

    <AppButton
      v-else
      variant="primary"
      :loading="loading"
      :disabled="!canExport"
      style="width: 100%; justify-content: center; margin-top: 8px;"
      @click="emit('export', { format, quality })"
    >
      Export Video
    </AppButton>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { VideoJob, VideoExport } from '@/types'
import AppButton from '@/components/common/AppButton.vue'
import AppProgress from '@/components/common/AppProgress.vue'
import { formatMs } from '@/utils/format'
import { jobService } from '@/services/job.service'

const props = defineProps<{
  startMs: number
  endMs: number
  job?: VideoJob | null
  currentExport?: VideoExport | null
  loading?: boolean
}>()

const emit = defineEmits<{ export: [{ format: string; quality: string }] }>()

const format = ref('mp4')
const quality = ref('original')

const canExport = computed(() => props.endMs > props.startMs && !props.loading)
const exportReady = computed(() => props.job?.status === 'completed' && props.currentExport)
const downloadUrl = computed(() =>
  props.currentExport ? jobService.getDownloadUrl(props.currentExport.id) : '#'
)
</script>

<style scoped>
.export-panel {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius-lg); padding: 20px;
  display: flex; flex-direction: column; gap: 12px;
}
.export-panel__title { font-size: 15px; font-weight: 600; }
.export-panel__field { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.export-panel__field label { color: var(--text-muted); font-size: 13px; }
.export-panel__field select {
  background: var(--surface2); border: 1px solid var(--border);
  color: var(--text); padding: 4px 8px; border-radius: var(--radius);
}
.export-panel__divider { height: 1px; background: var(--border); margin: 4px 0; }
.mono { font-family: monospace; font-size: 13px; }
.status-badge {
  font-size: 12px; font-weight: 600; padding: 2px 8px; border-radius: 4px;
  text-transform: capitalize;
}
.status-badge--queued { background: var(--surface2); color: var(--text-muted); }
.status-badge--processing { background: #1d4ed8; color: #fff; }
.status-badge--completed { background: var(--success); color: #fff; }
.status-badge--failed { background: var(--error); color: #fff; }
.export-panel__error { color: var(--error); font-size: 12px; }
.download-link {
  display: block; text-align: center; padding: 10px;
  background: var(--success); color: #fff; border-radius: var(--radius);
  font-weight: 600; text-decoration: none;
}
.download-link:hover { opacity: 0.9; }
</style>
