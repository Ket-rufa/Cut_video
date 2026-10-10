import { api } from './api'
import type { Clip } from '@/types'
import { DEFAULT_CLIPS } from '@/utils/mockData'

let localClips: Record<string, Clip[]> = {
  'vid-demo-01': [...DEFAULT_CLIPS]
}

export const clipService = {
  async listByVideoId(videoId: string): Promise<Clip[]> {
    try {
      return await api.get<Clip[]>(`/videos/${videoId}/clips`)
    } catch {
      if (!localClips[videoId]) {
        localClips[videoId] = videoId === 'vid-demo-01' ? [...DEFAULT_CLIPS] : []
      }
      return [...localClips[videoId]]
    }
  },

  async createClip(videoId: string, clip: Omit<Clip, 'id' | 'created_at'>): Promise<Clip> {
    try {
      return await api.post<Clip>(`/videos/${videoId}/clips`, clip)
    } catch {
      const newClip: Clip = {
        ...clip,
        id: 'clip-' + Date.now().toString(36),
        created_at: new Date().toISOString(),
      }
      if (!localClips[videoId]) {
        localClips[videoId] = []
      }
      localClips[videoId].push(newClip)
      return newClip
    }
  },

  async deleteClip(clipId: string, videoId?: string): Promise<void> {
    try {
      await api.delete(`/clips/${clipId}`)
    } catch {
      if (videoId && localClips[videoId]) {
        localClips[videoId] = localClips[videoId].filter(c => c.id !== clipId)
      } else {
        for (const vid in localClips) {
          localClips[vid] = localClips[vid].filter(c => c.id !== clipId)
        }
      }
    }
  },

  downloadClip(clip: Clip) {
    const link = document.createElement('a')
    if (clip.video_id) {
      const params = new URLSearchParams()
      params.set('start_ms', String(clip.start_time_ms ?? 0))
      params.set('end_ms', String(clip.end_time_ms ?? (clip.duration_ms || 30000)))
      params.set('download', '1')
      if (clip.title) params.set('title', `${clip.title.replace(/\s+/g, '_')}_${clip.id}`)
      if (clip.speed && clip.speed !== 1) params.set('speed', String(clip.speed))
      if (clip.filter && clip.filter.preset !== 'none') {
        params.set('filter', JSON.stringify(clip.filter))
      }
      link.href = `/api/v1/videos/${clip.video_id}/file?${params.toString()}`
    } else {
      link.href = clip.file_path || 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4'
    }
    link.download = `${clip.title.replace(/\s+/g, '_')}_${clip.id}.mp4`
    link.target = '_blank'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  },

  downloadAllClips(clips: Clip[]) {
    // Sequentially trigger or notify
    clips.forEach((clip, index) => {
      setTimeout(() => {
        clipService.downloadClip(clip)
      }, index * 400)
    })
  }
}
