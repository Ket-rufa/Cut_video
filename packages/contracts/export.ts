export type ExportFormat = 'mp4' | 'webm'
export type ExportQuality = 'original' | '1080p' | '720p' | '480p'

export interface VideoExport {
  id: string
  video_id: string
  job_id: string
  start_ms: number
  end_ms: number
  width: number
  height: number
  format: ExportFormat
  quality: ExportQuality
  file_size: number
  output_path: string
  expires_at: string
  created_at: string
}
