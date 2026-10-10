export type JobType = 'download' | 'trim' | 'crop' | 'convert' | 'export' | 'thumbnail'
export type JobStatus = 'queued' | 'processing' | 'completed' | 'failed' | 'cancelled'

export interface VideoJob {
  id: string
  video_id: string
  type: JobType
  status: JobStatus
  progress: number
  payload: Record<string, unknown>
  error_message: string
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface TrimPayload {
  start_ms: number
  end_ms: number
  format: string
  quality: string
}
