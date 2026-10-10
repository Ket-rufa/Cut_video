import { api } from './api'
import type { Video, VideoAnalysisResult, DownloadJobProgress } from '@/types'
import { DEFAULT_CURRENT_VIDEO, DEFAULT_SAVED_VIDEOS } from '@/utils/mockData'

let localVideos: Video[] = [...DEFAULT_SAVED_VIDEOS]

export function extractTagsFromVideo(v: Partial<Video>): string[] {
  const set = new Set<string>()

  // Explicit tags
  if (Array.isArray(v.tags)) {
    for (const t of v.tags) {
      const clean = t.replace(/^#/, '').trim()
      if (clean) set.add(clean)
    }
  }

  // Extract #hashtags from description and title
  const text = `${v.title || ''} ${v.description || ''}`
  const matches = text.match(/#([\p{L}\w_-]+)/gu)
  if (matches) {
    for (const m of matches) {
      const clean = m.replace(/^#/, '').trim()
      if (clean && clean.length >= 2) set.add(clean)
    }
  }

  return Array.from(set)
}

export const videoService = {
  // Real Analysis API
  async analyze(url: string): Promise<VideoAnalysisResult> {
    return await api.post<VideoAnalysisResult>('/videos/analyze', { url })
  },

  // Real Download Job creation
  async createDownload(analysisId: string, quality: string, url?: string): Promise<{ jobId: string; status: string; videoId?: string }> {
    return await api.post<{ jobId: string; status: string; videoId?: string }>('/videos/download', {
      analysisId,
      quality,
      url,
    })
  },

  // Poll download progress
  async getDownloadJob(jobId: string): Promise<DownloadJobProgress> {
    return await api.get<DownloadJobProgress>(`/videos/jobs/${jobId}`)
  },

  // Cancel download job
  async cancelDownloadJob(jobId: string): Promise<void> {
    await api.post(`/videos/jobs/${jobId}/cancel`, {})
  },

  // Direct video file stream URL
  getVideoStreamUrl(videoId: string): string {
    return `/api/v1/videos/${videoId}/file`
  },

  async resolve(url: string): Promise<Video> {
    try {
      return await api.post<Video>('/videos/resolve', { url })
    } catch {
      // Local fallback for quick preview / demo
      const id = 'vid-' + Date.now().toString(36)
      const isYt = url.includes('youtube') || url.includes('youtu.be')
      const isTt = url.includes('tiktok')
      const isDy = url.includes('douyin')
      const isFb = url.includes('facebook') || url.includes('fb.watch')
      const isIg = url.includes('instagram')
      const isTw = url.includes('twitter') || url.includes('x.com')

      const plat = isYt ? 'youtube' : isTt ? 'tiktok' : isDy ? 'douyin' : isFb ? 'facebook' : isIg ? 'instagram' : isTw ? 'twitter' : 'generic'

      const newVideo: Video = {
        id,
        source_url: url,
        platform: plat,
        external_id: id,
        title: `Video ${plat.toUpperCase()} mới lưu - ${new Date().toLocaleDateString('vi-VN')}`,
        description: 'Video nhập từ mạng xã hội, sẵn sàng để cắt và lưu trữ.',
        thumbnail_url: DEFAULT_CURRENT_VIDEO.thumbnail_url,
        duration_ms: 195000,
        width: 1280,
        height: 720,
        source_path: DEFAULT_CURRENT_VIDEO.source_path,
        status: 'resolved',
        folder_id: 'f-travel',
        folder_name: 'Du lịch',
        tags: ['mới', 'mạng xã hội'],
        file_size: 45000000,
        is_favorite: false,
        is_trash: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }
      localVideos.unshift(newVideo)
      return newVideo
    }
  },

  async getById(id: string): Promise<Video> {
    try {
      const v = await api.get<Video>(`/videos/${id}`)
      if (v) {
        // Ensure source_path points to streaming endpoint if empty or relative
        if (!v.source_path || v.source_path.startsWith('/') || v.source_path.includes('\\')) {
          v.source_path = `/api/v1/videos/${v.id}/file`
        }
        v.tags = extractTagsFromVideo(v)
        return v
      }
    } catch {
      // fallback
    }
    const found = localVideos.find(v => v.id === id)
    if (found) return found
    return DEFAULT_CURRENT_VIDEO
  },

  async list(filters?: { platform?: string; folder_id?: string; search?: string }): Promise<Video[]> {
    try {
      const params = new URLSearchParams()
      if (filters?.platform && filters.platform !== 'all') params.append('platform', filters.platform)
      if (filters?.folder_id) params.append('folder_id', filters.folder_id)
      if (filters?.search) params.append('search', filters.search)
      const q = params.toString() ? `?${params.toString()}` : ''
      const res = await api.get<Video[]>(`/videos${q}`)
      if (res && res.length > 0) {
        return res.map(v => ({
          ...v,
          tags: extractTagsFromVideo(v),
          source_path: v.source_path ? `/api/v1/videos/${v.id}/file` : v.source_url,
        }))
      }
    } catch {
      // fallback
    }
    let res = [...localVideos]
    if (filters?.platform && filters.platform !== 'all') {
      res = res.filter(v => v.platform === filters.platform)
    }
    if (filters?.folder_id && filters.folder_id !== 'f-all') {
      res = res.filter(v => v.folder_id === filters.folder_id)
    }
    if (filters?.search) {
      const s = filters.search.toLowerCase()
      res = res.filter(v => v.title.toLowerCase().includes(s) || v.tags?.some(t => t.toLowerCase().includes(s)))
    }
    return res.map(v => ({
      ...v,
      tags: extractTagsFromVideo(v),
    }))
  },

  async update(id: string, updates: Partial<Video>): Promise<Video> {
    try {
      return await api.post<Video>(`/videos/${id}`, updates)
    } catch {
      const idx = localVideos.findIndex(v => v.id === id)
      if (idx >= 0) {
        localVideos[idx] = { ...localVideos[idx], ...updates, updated_at: new Date().toISOString() }
        return localVideos[idx]
      }
      return { ...DEFAULT_CURRENT_VIDEO, ...updates }
    }
  },

  async delete(id: string): Promise<void> {
    try {
      await api.delete(`/videos/${id}`)
    } catch {
      // ignore
    }
    localVideos = localVideos.filter(v => v.id !== id)
  },

  async upload(file: File): Promise<Video> {
    const id = 'vid-upload-' + Date.now().toString(36)
    const localUrl = URL.createObjectURL(file)
    const newVideo: Video = {
      id,
      source_url: localUrl,
      platform: 'generic',
      external_id: id,
      title: file.name.replace(/\.[^/.]+$/, ''),
      description: `Video tải lên từ thiết bị (${file.type || 'video/mp4'})`,
      thumbnail_url: DEFAULT_CURRENT_VIDEO.thumbnail_url,
      duration_ms: 120000,
      width: 1920,
      height: 1080,
      source_path: localUrl,
      status: 'resolved',
      folder_id: 'f-travel',
      folder_name: 'Du lịch',
      tags: ['upload', 'thiết bị'],
      file_size: file.size,
      is_favorite: false,
      is_trash: false,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
    localVideos.unshift(newVideo)
    return newVideo
  }
}
