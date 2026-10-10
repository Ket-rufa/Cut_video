import { api } from './api'
import type { VideoSubtitle, SubtitleCue, AITranslationResult, AIConfigStatus, VideoDubbing, DubbingVoice } from '@/types'

export const aiService = {
  async translateMeta(videoId: string, apply: boolean = false): Promise<AITranslationResult> {
    const q = apply ? '?apply=1' : ''
    return await api.post<AITranslationResult>(`/videos/${videoId}/translate-meta${q}`, {})
  },

  async generateSubtitles(videoId: string, range?: { start_ms?: number; end_ms?: number }): Promise<VideoSubtitle> {
    return await api.post<VideoSubtitle>(`/videos/${videoId}/subtitles/generate`, range || {})
  },

  async updateSubtitles(videoId: string, cues: SubtitleCue[]): Promise<VideoSubtitle> {
    return await api.put<VideoSubtitle>(`/videos/${videoId}/subtitles`, { cues })
  },

  async translateCue(videoId: string, cue: { index: number; text: string; start_ms?: number; end_ms?: number }): Promise<{ original_text: string; translated_text: string; splitted_cues?: SubtitleCue[] }> {
    return await api.post<{ original_text: string; translated_text: string; splitted_cues?: SubtitleCue[] }>(`/videos/${videoId}/subtitles/translate-cue`, cue)
  },

  async getSubtitles(videoId: string): Promise<VideoSubtitle | null> {
    try {
      return await api.get<VideoSubtitle | null>(`/videos/${videoId}/subtitles`)
    } catch {
      return null
    }
  },

  getDownloadSubtitleUrl(videoId: string): string {
    return `/api/v1/videos/${videoId}/subtitles/download`
  },

  getCueAudioUrl(videoId: string, index: number, voice?: string): string {
    const vParam = voice ? `?voice=${encodeURIComponent(voice)}` : ''
    return `/api/v1/videos/${videoId}/subtitles/cues/${index}/audio${vParam}`
  },

  async getDubbingVoices(): Promise<DubbingVoice[]> {
    try {
      const res = await api.get<{ voices: DubbingVoice[] }>('/dubbing/voices')
      return res?.voices || []
    } catch {
      return []
    }
  },

  getVoicePreviewUrl(voiceId: string): string {
    return `/api/v1/dubbing/preview?voice=${encodeURIComponent(voiceId)}`
  },

  async generateDubbing(videoId: string, params?: { start_ms?: number; end_ms?: number; bg_volume?: number; voice?: string }): Promise<VideoDubbing> {
    return await api.post<VideoDubbing>(`/videos/${videoId}/dubbing/generate`, params || {})
  },

  async getDubbing(videoId: string): Promise<VideoDubbing | null> {
    try {
      return await api.get<VideoDubbing>(`/videos/${videoId}/dubbing`)
    } catch {
      return null
    }
  },

  async getAIConfig(): Promise<AIConfigStatus> {
    return await api.get<AIConfigStatus>('/settings/ai')
  },

  async updateAIConfig(apiKey: string): Promise<{ success: boolean; message: string }> {
    return await api.post<{ success: boolean; message: string }>('/settings/ai', { api_key: apiKey })
  },

  async testAIConnection(apiKey?: string): Promise<{ success: boolean; message: string }> {
    return await api.post<{ success: boolean; message: string }>('/settings/ai/test', { api_key: apiKey || '' })
  },
}

