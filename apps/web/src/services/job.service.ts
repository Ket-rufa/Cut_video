import { api } from './api'
import type { VideoJob, VideoExport } from '@/types'

export interface TrimRequest {
  start_ms: number
  end_ms: number
  format?: string
  quality?: string
}

export const jobService = {
  createTrimJob: (videoId: string, req: TrimRequest) =>
    api.post<VideoJob>(`/videos/${videoId}/trim`, req),

  createExportJob: (videoId: string, req: TrimRequest) =>
    api.post<VideoJob>(`/videos/${videoId}/export`, req),

  getJob: (id: string) =>
    api.get<VideoJob>(`/jobs/${id}`),

  getJobProgress: (id: string) =>
    api.get<{ job_id: string; status: string; progress: number }>(`/jobs/${id}/progress`),

  listJobs: () =>
    api.get<VideoJob[]>('/jobs'),

  getExport: (id: string) =>
    api.get<VideoExport>(`/exports/${id}`),

  getDownloadUrl: (id: string) =>
    `${import.meta.env.VITE_API_URL || '/api/v1'}/exports/${id}/download`,
}
