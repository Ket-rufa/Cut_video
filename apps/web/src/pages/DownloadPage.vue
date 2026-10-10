<template>
  <div class="download-page">
    <header class="download-page__header">
      <RouterLink to="/" class="back-link">← Home</RouterLink>
      <h1>Download</h1>
    </header>
    <main class="download-page__main">
      <div v-if="loading" class="state">Loading job…</div>
      <div v-else-if="error" class="state state--error">{{ error }}</div>
      <template v-else-if="job">
        <div class="job-card">
          <div class="job-card__row">
            <span class="label">Job ID</span>
            <span class="mono">{{ job.id }}</span>
          </div>
          <div class="job-card__row">
            <span class="label">Status</span>
            <span :class="['status-badge', `status-badge--${job.status}`]">{{ job.status }}</span>
          </div>
          <div v-if="job.status === 'processing' || job.status === 'queued'" class="job-card__row">
            <AppProgress :value="job.progress" />
          </div>
          <div v-if="job.error_message" class="job-card__error">{{ job.error_message }}</div>
        </div>

        <div v-if="exportRecord" class="export-card">
          <h2>Export Ready</h2>
          <div class="export-card__row">
            <span class="label">Format</span>
            <span>{{ exportRecord.format }}</span>
          </div>
          <div class="export-card__row">
            <span class="label">Duration</span>
            <span>{{ formatMs(exportRecord.end_ms - exportRecord.start_ms) }}</span>
          </div>
          <div class="export-card__row">
            <span class="label">File size</span>
            <span>{{ formatBytes(exportRecord.file_size) }}</span>
          </div>
          <a :href="downloadUrl" download class="download-btn">⬇ Download MP4</a>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import type { VideoJob, VideoExport } from '@/types'
import { jobService } from '@/services/job.service'
import AppProgress from '@/components/common/AppProgress.vue'
import { formatMs, formatBytes } from '@/utils/format'

const route = useRoute()
const job = ref<VideoJob | null>(null)
const exportRecord = ref<VideoExport | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
let interval: ReturnType<typeof setInterval> | null = null

const downloadUrl = computed(() =>
  exportRecord.value ? jobService.getDownloadUrl(exportRecord.value.id) : '#'
)

onMounted(async () => {
  const jobId = route.params.jobId as string
  await loadJob(jobId)
  if (job.value?.status !== 'completed' && job.value?.status !== 'failed') {
    interval = setInterval(() => loadJob(jobId), 2000)
  }
})

onUnmounted(() => {
  if (interval) clearInterval(interval)
})

async function loadJob(id: string) {
  try {
    job.value = await jobService.getJob(id)
    if (job.value.status === 'completed') {
      if (interval) { clearInterval(interval); interval = null }
      try {
        exportRecord.value = await jobService.getExport(id)
      } catch {
        // export may be keyed by export ID — backend now also accepts job ID
      }
    }
    if (job.value.status === 'failed') {
      if (interval) { clearInterval(interval); interval = null }
    }
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.download-page { display: flex; flex-direction: column; min-height: 100vh; }
.download-page__header {
  display: flex; align-items: center; gap: 16px;
  padding: 14px 24px; border-bottom: 1px solid var(--border); background: var(--surface);
}
.download-page__header h1 { font-size: 16px; font-weight: 700; }
.back-link { color: var(--text-muted); font-size: 14px; }
.download-page__main { padding: 32px 24px; display: flex; flex-direction: column; gap: 20px; max-width: 500px; }
.state { color: var(--text-muted); }
.state--error { color: var(--error); }
.job-card, .export-card {
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--radius-lg); padding: 20px;
  display: flex; flex-direction: column; gap: 12px;
}
.export-card h2 { font-size: 15px; font-weight: 600; color: var(--success); }
.job-card__row, .export-card__row { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.label { color: var(--text-muted); font-size: 13px; }
.mono { font-family: monospace; font-size: 12px; }
.job-card__error { color: var(--error); font-size: 13px; }
.status-badge {
  font-size: 12px; font-weight: 600; padding: 2px 8px; border-radius: 4px; text-transform: capitalize;
}
.status-badge--queued    { background: var(--surface2); color: var(--text-muted); }
.status-badge--processing { background: #1d4ed8; color: #fff; }
.status-badge--completed  { background: var(--success); color: #fff; }
.status-badge--failed     { background: var(--error); color: #fff; }
.download-btn {
  display: block; text-align: center; padding: 12px;
  background: var(--success); color: #fff; border-radius: var(--radius);
  font-weight: 600; text-decoration: none; margin-top: 8px;
}
.download-btn:hover { opacity: 0.9; }
</style>
