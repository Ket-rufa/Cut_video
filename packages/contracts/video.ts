export type VideoPlatform = 'youtube' | 'tiktok' | 'douyin' | 'facebook' | 'instagram' | 'generic' | 'unknown'
export type VideoStatus = 'pending' | 'resolved' | 'downloaded' | 'failed'

export interface Video {
  id: string
  source_url: string
  platform: VideoPlatform
  external_id: string
  title: string
  description: string
  thumbnail_url: string
  duration_ms: number
  width: number
  height: number
  source_path: string
  status: VideoStatus
  created_at: string
  updated_at: string
}

export interface VideoFormat {
  quality: string
  url: string
  width: number
  height: number
  mime_type: string
}
