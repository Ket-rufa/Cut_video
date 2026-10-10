import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { VideoJob, VideoExport } from '@/types'
import { jobService } from '@/services/job.service'

export const useJobStore = defineStore('job', () => {
  const currentJob = ref<VideoJob | null>(null)
  const currentExport = ref<VideoExport | null>(null)
  const allJobs = ref<VideoJob[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  let pollInterval: ReturnType<typeof setInterval> | null = null

  async function createTrimJob(videoId: string, startMs: number, endMs: number, format: string, quality: string) {
    loading.value = true
    error.value = null
    try {
      const job = await jobService.createTrimJob(videoId, {
        start_ms: startMs,
        end_ms: endMs,
        format,
        quality,
      })
      currentJob.value = job
      startPolling(job.id)
      return job
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  function startPolling(jobId: string) {
    stopPolling()
    pollInterval = setInterval(async () => {
      try {
        const job = await jobService.getJob(jobId)
        currentJob.value = job
        if (job.status === 'completed') {
          stopPolling()
          // Fetch export
          try {
            const exp = await jobService.getExport(job.id)
            currentExport.value = exp
          } catch {
            // export may be keyed by export id, try fetching by job id via jobs endpoint
          }
        } else if (job.status === 'failed' || job.status === 'cancelled') {
          stopPolling()
        }
      } catch (e) {
        // continue polling
      }
    }, 2000)
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  async function fetchAllJobs() {
    try {
      allJobs.value = await jobService.listJobs()
    } catch (e: any) {
      error.value = e.message
    }
  }

  function reset() {
    stopPolling()
    currentJob.value = null
    currentExport.value = null
    error.value = null
  }

  return {
    currentJob, currentExport, allJobs, loading, error,
    createTrimJob, startPolling, stopPolling, fetchAllJobs, reset,
  }
})
