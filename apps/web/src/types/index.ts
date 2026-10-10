export type VideoPlatform = 'youtube' | 'tiktok' | 'douyin' | 'facebook' | 'instagram' | 'twitter' | 'generic' | 'unknown'
export type VideoStatus = 'pending' | 'resolved' | 'downloaded' | 'failed'
export type JobType = 'download' | 'trim' | 'crop' | 'convert' | 'export' | 'thumbnail'
export type JobStatus = 'queued' | 'processing' | 'completed' | 'failed' | 'cancelled'
export type ExportFormat = 'mp4' | 'webm'
export type ExportQuality = 'original' | '1080p' | '720p' | '480p'

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
  folder_id?: string
  folder_name?: string
  tags?: string[]
  file_size?: number
  is_favorite?: boolean
  is_trash?: boolean
  created_at: string
  updated_at: string
}

export type FilterPresetType = 'none' | 'vivid' | 'warm' | 'cool' | 'cinematic' | 'vintage' | 'bw'

export interface ClipFilterConfig {
  preset: FilterPresetType
  brightness: number  // -100 to +100 (default: 0)
  contrast: number    // -100 to +100 (default: 0)
  saturation: number  // -100 to +100 (default: 0)
  temperature: number // -100 to +100 (default: 0)
}

export interface TimelineClip {
  id: string
  sourceStartMs: number
  sourceEndMs: number
  speed: number         // 0.25 to 4.0 (default: 1.0)
  filter: ClipFilterConfig
}

export interface Clip {
  id: string
  video_id: string
  title: string
  start_time_ms: number
  end_time_ms: number
  duration_ms: number
  file_path?: string
  thumbnail?: string
  status: 'ready' | 'processing' | 'downloaded'
  created_at: string
  speed?: number
  filter?: ClipFilterConfig
}

export interface Folder {
  id: string
  name: string
  icon?: string
  color?: string
  count?: number
  created_at: string
}

export interface Tag {
  id: string
  name: string
  color?: string
  count?: number
}

export interface HistoryItem {
  id: string
  video_id: string
  video_title: string
  action: 'Cắt video' | 'Lưu video' | 'Tải đoạn' | 'Tải tất cả' | 'Xóa đoạn'
  clip_name?: string
  status: 'Đã cắt' | 'Đang xử lý' | 'Thất bại' | 'Đã tải xuống'
  file_size: number
  created_at: string
}

export interface AISuggestion {
  id: string
  title: string
  start_ms: number
  end_ms: number
  duration_s: number
  thumbnail?: string
}

export interface VideoJob {
  id: string
  video_id: string
  type: JobType
  status: JobStatus
  progress: number
  payload: Record<string, unknown>
  error_message?: string
  started_at: string | null
  completed_at: string | null
  created_at: string
  updated_at: string
}

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

export interface ApiResponse<T> {
  data: T
}

export interface ApiError {
  error: {
    code: string
    message: string
  }
}

export interface VideoFormatInfo {
  formatId: string
  container: string
  quality: string
  width?: number
  height?: number
  fps?: number
  hasVideo: boolean
  hasAudio: boolean
  audioOnly: boolean
  fileSize?: number
}

export interface VideoAnalysisResult {
  analysisId: string
  platform: VideoPlatform
  sourceUrl: string
  videoId: string
  title: string
  description: string
  thumbnail: string
  duration: number
  durationMs: number
  author: {
    name: string
    avatar?: string
  }
  formats: VideoFormatInfo[]
}

export interface DownloadJobProgress {
  id: string
  videoId?: string
  platform?: string
  status: 'QUEUED' | 'ANALYZING' | 'DOWNLOADING' | 'MERGING' | 'SAVING' | 'COMPLETED' | 'FAILED' | 'CANCELLED' | string
  progress: number
  downloadedBytes: number
  totalBytes: number
  speed: number
  eta: number
  error?: string
}

export interface SubtitleCue {
  index: number
  start_ms: number
  end_ms: number
  text: string
}

export interface VideoSubtitle {
  id: string
  video_id: string
  language_code: string
  source_language: string
  format: string
  content: string
  cues: SubtitleCue[]
  created_at: string
  updated_at: string
}

export interface AITranslationResult {
  source_language: string
  target_language: string
  original_title: string
  translated_title: string
  original_description: string
  translated_description: string
  original_tags: string[]
  translated_tags: string[]
}

export interface AIConfigStatus {
  has_key: boolean
  is_configured?: boolean
  masked_key: string
  model: string
}

export interface VideoDubbing {
  video_id: string
  audio_url: string
  video_url?: string
  cue_count: number
  start_ms: number
  end_ms: number
  duration_ms: number
  status: string
  created_at: string
}

export interface DubbingVoice {
  id: string
  name: string
  gender: string
  description: string
  tag: string
  icon: string
  sample_text: string
}

