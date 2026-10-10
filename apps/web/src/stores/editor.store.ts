import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Video, Clip, AISuggestion, TimelineClip, ClipFilterConfig, FilterPresetType, VideoSubtitle, SubtitleCue, AITranslationResult, VideoDubbing, DubbingVoice } from '@/types'
import { DEFAULT_CURRENT_VIDEO, DEFAULT_CLIPS, DEFAULT_AI_SUGGESTIONS } from '@/utils/mockData'
import { clipService } from '@/services/clip.service'
import { historyService } from '@/services/history.service'
import { aiService } from '@/services/ai.service'
import { DEFAULT_CLIP_FILTER, getFilterCss, calculateOutputDuration } from '@/utils/filter'

export interface EditorSnapshot {
  timelineClips: TimelineClip[]
  selectedTimelineClipId: string | null
  startTime: number
  endTime: number
  currentTime: number
}

export const useEditorStore = defineStore('editor', () => {
  const currentVideo = ref<Video>({ ...DEFAULT_CURRENT_VIDEO })
  const currentTime = ref(24000)      // ms (initial 00:24)
  const startTime = ref(24000)        // ms (clipStart: 00:00:24)
  const endTime = ref(52000)          // ms (clipEnd: 00:00:52)
  const duration = ref(195000)        // ms (total video duration: 03:15)
  const isPlaying = ref(false)
  const volume = ref(1.0)
  const playbackRate = ref(1.0)
  
  // Saved library clips
  const clips = ref<Clip[]>([...DEFAULT_CLIPS])
  const selectedClip = ref<Clip | null>(null)
  const aiSuggestions = ref<AISuggestion[]>([...DEFAULT_AI_SUGGESTIONS])
  
  // Timeline Multi-Clip State (EDL - Edit Decision List)
  const timelineClips = ref<TimelineClip[]>([
    {
      id: 'timeline-clip-1',
      sourceStartMs: 0,
      sourceEndMs: 195000,
      speed: 1.0,
      filter: { ...DEFAULT_CLIP_FILTER },
    }
  ])
  const selectedTimelineClipId = ref<string | null>('timeline-clip-1')
  const activeTool = ref<'trim' | 'split' | 'speed' | 'filter' | 'subtitle' | null>(null)

  // AI Subtitles, Dubbing & Translation
  const subtitles = ref<VideoSubtitle | null>(null)
  const showSubtitles = ref(true)
  const isGeneratingSubtitles = ref(false)
  const isTranslatingMeta = ref(false)
  const dubbing = ref<VideoDubbing | null>(null)
  const isGeneratingDubbing = ref(false)
  const isPlayingDubbedAudio = ref(false)
  const availableVoices = ref<DubbingVoice[]>([])
  const selectedVoiceId = ref<string>('male_deep')

  // Subtitle cue matching currentTime
  const currentSubtitleCue = computed<SubtitleCue | null>(() => {
    if (!subtitles.value || !subtitles.value.cues || !showSubtitles.value) return null
    const ct = currentTime.value
    return subtitles.value.cues.find(c => ct >= c.start_ms && ct <= c.end_ms) || null
  })
  const currentSubtitleText = computed(() => currentSubtitleCue.value?.text || '')

  // Undo / Redo Stacks
  const undoStack = ref<EditorSnapshot[]>([])
  const redoStack = ref<EditorSnapshot[]>([])
  const canUndo = computed(() => undoStack.value.length > 0)
  const canRedo = computed(() => redoStack.value.length > 0)

  const selectedFormat = ref('mp4')
  const selectedQuality = ref('original')

  // Computed clip duration (ms and sec) for current trim selection
  const clipDuration = computed(() => Math.max(0, endTime.value - startTime.value))
  const clipDurationSec = computed(() => Math.round(clipDuration.value / 1000))

  // Active selected timeline clip
  const selectedTimelineClip = computed<TimelineClip | null>(() => {
    if (!selectedTimelineClipId.value) {
      return timelineClips.value[0] || null
    }
    return timelineClips.value.find(c => c.id === selectedTimelineClipId.value) || timelineClips.value[0] || null
  })

  // Clip currently under playhead position
  const currentPlayingClip = computed<TimelineClip | null>(() => {
    const time = currentTime.value
    const found = timelineClips.value.find(c => time >= c.sourceStartMs && time <= c.sourceEndMs)
    return found || selectedTimelineClip.value || timelineClips.value[0] || null
  })

  // Total timeline output duration accounting for speeds of all clips
  const totalTimelineOutputDurationMs = computed(() => {
    return timelineClips.value.reduce((acc, c) => {
      return acc + calculateOutputDuration(c.sourceStartMs, c.sourceEndMs, c.speed)
    }, 0)
  })

  // Real-time CSS filter string for currently playing/selected clip
  const activeFilterCss = computed(() => {
    const clip = currentPlayingClip.value || selectedTimelineClip.value
    return getFilterCss(clip?.filter)
  })

  // --- Snapshot Helper ---
  function takeSnapshot(): EditorSnapshot {
    return {
      timelineClips: JSON.parse(JSON.stringify(timelineClips.value)),
      selectedTimelineClipId: selectedTimelineClipId.value,
      startTime: startTime.value,
      endTime: endTime.value,
      currentTime: currentTime.value,
    }
  }

  function pushUndoSnapshot() {
    undoStack.value.push(takeSnapshot())
    if (undoStack.value.length > 50) {
      undoStack.value.shift()
    }
    redoStack.value = []
  }

  function undo(): boolean {
    if (undoStack.value.length === 0) return false
    const prev = undoStack.value.pop()!
    redoStack.value.push(takeSnapshot())
    if (redoStack.value.length > 50) {
      redoStack.value.shift()
    }

    timelineClips.value = prev.timelineClips
    selectedTimelineClipId.value = prev.selectedTimelineClipId
    startTime.value = prev.startTime
    endTime.value = prev.endTime
    currentTime.value = prev.currentTime
    return true
  }

  function redo(): boolean {
    if (redoStack.value.length === 0) return false
    const next = redoStack.value.pop()!
    undoStack.value.push(takeSnapshot())
    if (undoStack.value.length > 50) {
      undoStack.value.shift()
    }

    timelineClips.value = next.timelineClips
    selectedTimelineClipId.value = next.selectedTimelineClipId
    startTime.value = next.startTime
    endTime.value = next.endTime
    currentTime.value = next.currentTime
    return true
  }

  function setVideo(video: Video) {
    currentVideo.value = { ...video }
    const totalMs = video.duration_ms || 195000
    duration.value = totalMs
    startTime.value = 0
    endTime.value = Math.min(30000, totalMs)
    currentTime.value = 0

    undoStack.value = []
    redoStack.value = []

    // Initialize timeline with single full-length clip
    initTimelineClips(totalMs)
    loadClips(video.id)
    loadSubtitles(video.id)
    loadDubbing(video.id)
  }

  function initTimelineClips(totalMs: number) {
    const clipId = 'tclip-' + Date.now().toString(36)
    timelineClips.value = [
      {
        id: clipId,
        sourceStartMs: 0,
        sourceEndMs: totalMs > 0 ? totalMs : 195000,
        speed: 1.0,
        filter: { ...DEFAULT_CLIP_FILTER },
      }
    ]
    selectedTimelineClipId.value = clipId
  }

  async function loadClips(videoId: string) {
    const loaded = await clipService.listByVideoId(videoId)
    if (loaded && loaded.length > 0) {
      clips.value = loaded
    } else {
      clips.value = []
    }
  }

  async function loadSubtitles(videoId: string) {
    if (!videoId) return
    try {
      const sub = await aiService.getSubtitles(videoId)
      subtitles.value = sub
    } catch {
      subtitles.value = null
    }
  }

  async function generateSubtitles(rangeOnly: boolean = false, startMs?: number, endMs?: number) {
    if (!currentVideo.value.id) return null
    isGeneratingSubtitles.value = true
    try {
      const s = startMs !== undefined ? startMs : startTime.value
      const e = endMs !== undefined ? endMs : endTime.value
      const range = rangeOnly ? { start_ms: s, end_ms: e } : undefined
      const sub = await aiService.generateSubtitles(currentVideo.value.id, range)
      subtitles.value = sub
      showSubtitles.value = true
      return sub
    } finally {
      isGeneratingSubtitles.value = false
    }
  }

  async function updateCueText(cueIndex: number, newText: string) {
    if (!currentVideo.value.id || !subtitles.value) return
    const updatedCues = subtitles.value.cues.map(c => {
      if (c.index === cueIndex) {
        return { ...c, text: newText.trim() }
      }
      return c
    })
    subtitles.value.cues = updatedCues
    try {
      const updatedSub = await aiService.updateSubtitles(currentVideo.value.id, updatedCues)
      subtitles.value = updatedSub
    } catch (e) {
      console.error('Failed to save updated cue:', e)
    }
  }

  async function translateSingleCue(cueIndex: number) {
    if (!currentVideo.value.id || !subtitles.value) return
    const targetCue = subtitles.value.cues.find(c => c.index === cueIndex)
    if (!targetCue) return
    try {
      const res = await aiService.translateCue(currentVideo.value.id, {
        index: targetCue.index,
        text: targetCue.text,
        start_ms: targetCue.start_ms,
        end_ms: targetCue.end_ms,
      })
      if (res && res.translated_text) {
        if (res.splitted_cues && res.splitted_cues.length > 1) {
          const newCues: SubtitleCue[] = []
          for (const c of subtitles.value.cues) {
            if (c.index === cueIndex) {
              newCues.push(...res.splitted_cues)
            } else {
              newCues.push(c)
            }
          }
          newCues.forEach((c, idx) => { c.index = idx + 1 })
          subtitles.value.cues = newCues
          const updatedSub = await aiService.updateSubtitles(currentVideo.value.id, newCues)
          subtitles.value = updatedSub
        } else {
          await updateCueText(cueIndex, res.translated_text)
        }
      }
    } catch (e) {
      console.error('Failed to translate cue:', e)
    }
  }


  async function loadDubbing(videoId: string) {
    try {
      const d = await aiService.getDubbing(videoId)
      dubbing.value = d
    } catch {
      dubbing.value = null
    }
  }

  async function loadDubbingVoices() {
    try {
      const v = await aiService.getDubbingVoices()
      if (v && v.length > 0) {
        availableVoices.value = v
      }
    } catch {
      // Keep existing
    }
  }

  function setSelectedVoice(voiceId: string) {
    selectedVoiceId.value = voiceId
  }

  async function generateDubbing(params?: { start_ms?: number; end_ms?: number; bg_volume?: number; voice?: string }) {
    if (!currentVideo.value.id) return null
    isGeneratingDubbing.value = true
    try {
      const voiceToUse = params?.voice || selectedVoiceId.value || 'male_deep'
      const res = await aiService.generateDubbing(currentVideo.value.id, {
        ...params,
        voice: voiceToUse,
      })
      dubbing.value = res
      return res
    } finally {
      isGeneratingDubbing.value = false
    }
  }

  function togglePlayDubbedAudio() {
    isPlayingDubbedAudio.value = !isPlayingDubbedAudio.value
  }

  async function translateMeta(apply: boolean = true): Promise<AITranslationResult | null> {
    if (!currentVideo.value.id) return null
    isTranslatingMeta.value = true
    try {
      const res = await aiService.translateMeta(currentVideo.value.id, apply)
      if (apply && res) {
        currentVideo.value.title = res.translated_title
        currentVideo.value.description = res.translated_description
        if (res.translated_tags && res.translated_tags.length > 0) {
          currentVideo.value.tags = res.translated_tags
        }
      }
      return res
    } finally {
      isTranslatingMeta.value = false
    }
  }

  function toggleShowSubtitles() {
    showSubtitles.value = !showSubtitles.value
  }

  function setDuration(ms: number) {
    if (ms > 0) {
      duration.value = ms
      if (endTime.value === 0 || endTime.value > ms) {
        endTime.value = ms
      }
      // If only 1 timeline clip exists and reaches old duration, adjust it
      if (timelineClips.value.length === 1 && timelineClips.value[0].sourceStartMs === 0) {
        timelineClips.value[0].sourceEndMs = ms
      }
    }
  }

  function setRange(startMs: number, endMs: number) {
    const safeStart = Math.max(0, Math.min(startMs, duration.value - 500))
    const safeEnd = Math.max(safeStart + 500, Math.min(endMs, duration.value))
    startTime.value = safeStart
    endTime.value = safeEnd
    currentTime.value = Math.max(safeStart, Math.min(currentTime.value, safeEnd))

    if (selectedTimelineClipId.value) {
      const clip = timelineClips.value.find(c => c.id === selectedTimelineClipId.value)
      if (clip) {
        clip.sourceStartMs = safeStart
        clip.sourceEndMs = safeEnd
      }
    }
  }

  function setStartTime(ms: number) {
    const clamped = Math.max(0, Math.min(ms, endTime.value - 500))
    startTime.value = clamped
    currentTime.value = Math.max(clamped, currentTime.value)
    if (selectedTimelineClipId.value) {
      const clip = timelineClips.value.find(c => c.id === selectedTimelineClipId.value)
      if (clip) {
        clip.sourceStartMs = clamped
      }
    }
  }

  function setEndTime(ms: number) {
    const clamped = Math.max(startTime.value + 500, Math.min(ms, duration.value))
    endTime.value = clamped
    if (selectedTimelineClipId.value) {
      const clip = timelineClips.value.find(c => c.id === selectedTimelineClipId.value)
      if (clip) {
        clip.sourceEndMs = clamped
      }
    }
  }

  function setCurrentTime(ms: number) {
    const safeTime = Math.max(0, Math.min(ms, duration.value))
    currentTime.value = safeTime

    // Auto-select the timeline clip containing this playhead position if not already selected
    if (timelineClips.value.length > 1) {
      const containingClip = timelineClips.value.find(
        c => safeTime >= c.sourceStartMs && safeTime <= c.sourceEndMs
      )
      if (containingClip && containingClip.id !== selectedTimelineClipId.value) {
        selectedTimelineClipId.value = containingClip.id
        startTime.value = containingClip.sourceStartMs
        endTime.value = containingClip.sourceEndMs
      }
    }
  }

  function setVolume(v: number) {
    volume.value = Math.max(0, Math.min(1, v))
  }

  function setPlaybackRate(rate: number) {
    playbackRate.value = rate
  }

  function resetRange() {
    pushUndoSnapshot()
    startTime.value = 0
    endTime.value = duration.value > 0 ? duration.value : 60000
    currentTime.value = 0
  }

  function applyAISuggestion(suggestion: AISuggestion) {
    pushUndoSnapshot()
    startTime.value = suggestion.start_ms
    endTime.value = Math.min(suggestion.end_ms, duration.value)
    currentTime.value = suggestion.start_ms
  }

  // --- Multi-Clip / Split / Speed / Filter Actions ---

  function selectTimelineClip(clipId: string) {
    selectedTimelineClipId.value = clipId
    const found = timelineClips.value.find(c => c.id === clipId)
    if (found) {
      startTime.value = found.sourceStartMs
      endTime.value = found.sourceEndMs
      if (currentTime.value < found.sourceStartMs || currentTime.value > found.sourceEndMs) {
        currentTime.value = found.sourceStartMs
      }
    }
  }

  function setActiveTool(tool: 'trim' | 'split' | 'speed' | 'filter' | 'subtitle' | null) {
    if (activeTool.value === tool) {
      activeTool.value = null
    } else {
      activeTool.value = tool
    }
  }

  /**
   * Split timeline clip at currentTime (or specific ms).
   * Tolerance: minimum 300ms from start and end of the clip to avoid tiny slivers.
   */
  function splitClipAtCurrentTime(targetMs?: number): { success: boolean; message?: string; newClipId?: string } {
    const splitPoint = targetMs !== undefined ? targetMs : currentTime.value
    const toleranceMs = 300

    // Find the clip containing the split point
    const clipIndex = timelineClips.value.findIndex(
      c => splitPoint > c.sourceStartMs + toleranceMs && splitPoint < c.sourceEndMs - toleranceMs
    )

    if (clipIndex === -1) {
      return {
        success: false,
        message: 'Không thể tách quá gần đầu/cuối đoạn video hoặc ngoài phạm vi clip (tối thiểu 0.3s).',
      }
    }

    pushUndoSnapshot()

    const targetClip = timelineClips.value[clipIndex]
    const clip1Id = targetClip.id
    const clip2Id = 'tclip-' + Date.now().toString(36) + '-' + Math.random().toString(36).substring(2, 5)

    const clip1: TimelineClip = {
      ...targetClip,
      id: clip1Id,
      sourceStartMs: targetClip.sourceStartMs,
      sourceEndMs: splitPoint,
      // preserve filter & speed
      filter: { ...targetClip.filter },
    }

    const clip2: TimelineClip = {
      id: clip2Id,
      sourceStartMs: splitPoint,
      sourceEndMs: targetClip.sourceEndMs,
      speed: targetClip.speed,
      filter: { ...targetClip.filter },
    }

    // Replace target clip with the two new clips
    timelineClips.value.splice(clipIndex, 1, clip1, clip2)

    // Select the right clip (Clip 2) as required by UX
    selectedTimelineClipId.value = clip2Id
    startTime.value = clip2.sourceStartMs
    endTime.value = clip2.sourceEndMs

    return {
      success: true,
      newClipId: clip2Id,
    }
  }

  function setClipSpeed(clipId: string, newSpeed: number) {
    const safeSpeed = Math.max(0.25, Math.min(4.0, Number(newSpeed.toFixed(2))))
    const clip = timelineClips.value.find(c => c.id === clipId)
    if (clip && clip.speed !== safeSpeed) {
      pushUndoSnapshot()
      clip.speed = safeSpeed
    }
  }

  function resetClipSpeed(clipId: string) {
    const clip = timelineClips.value.find(c => c.id === clipId)
    if (clip && clip.speed !== 1.0) {
      pushUndoSnapshot()
      clip.speed = 1.0
    }
  }

  function setClipFilterPreset(clipId: string, preset: FilterPresetType) {
    const clip = timelineClips.value.find(c => c.id === clipId)
    if (clip && clip.filter.preset !== preset) {
      pushUndoSnapshot()
      clip.filter.preset = preset
    }
  }

  function setClipFilterAdjustment(clipId: string, adjustments: Partial<ClipFilterConfig>) {
    const clip = timelineClips.value.find(c => c.id === clipId)
    if (clip) {
      pushUndoSnapshot()
      clip.filter = {
        ...clip.filter,
        ...adjustments,
      }
    }
  }

  function resetClipFilter(clipId: string) {
    const clip = timelineClips.value.find(c => c.id === clipId)
    if (clip) {
      pushUndoSnapshot()
      clip.filter = { ...DEFAULT_CLIP_FILTER }
    }
  }

  function deleteTimelineClip(clipId: string) {
    if (timelineClips.value.length <= 1) {
      return
    }
    const idx = timelineClips.value.findIndex(c => c.id === clipId)
    if (idx !== -1) {
      pushUndoSnapshot()
      timelineClips.value.splice(idx, 1)
      const nextClip = timelineClips.value[Math.min(idx, timelineClips.value.length - 1)]
      selectedTimelineClipId.value = nextClip.id
      startTime.value = nextClip.sourceStartMs
      endTime.value = nextClip.sourceEndMs
    }
  }

  /**
   * Unsplit / Merge selected clip with adjacent clip.
   */
  function unsplitSelectedClip(clipId?: string): { success: boolean; message?: string } {
    if (timelineClips.value.length <= 1) {
      return { success: false, message: 'Chưa có đoạn nào bị tách để hủy.' }
    }

    const targetId = clipId || selectedTimelineClipId.value
    const idx = timelineClips.value.findIndex(c => c.id === targetId)
    if (idx === -1) {
      return { success: false, message: 'Không tìm thấy đoạn được chọn.' }
    }

    pushUndoSnapshot()

    if (idx > 0) {
      // Merge with previous clip
      const prevClip = timelineClips.value[idx - 1]
      const currentClip = timelineClips.value[idx]
      prevClip.sourceEndMs = currentClip.sourceEndMs
      timelineClips.value.splice(idx, 1)
      selectedTimelineClipId.value = prevClip.id
      startTime.value = prevClip.sourceStartMs
      endTime.value = prevClip.sourceEndMs
    } else {
      // Merge with next clip (idx === 0)
      const currentClip = timelineClips.value[0]
      const nextClip = timelineClips.value[1]
      nextClip.sourceStartMs = currentClip.sourceStartMs
      timelineClips.value.splice(0, 1)
      selectedTimelineClipId.value = nextClip.id
      startTime.value = nextClip.sourceStartMs
      endTime.value = nextClip.sourceEndMs
    }

    return { success: true }
  }

  /**
   * Merge all split clips back into one single continuous clip.
   */
  function unsplitAllClips(): { success: boolean } {
    if (timelineClips.value.length <= 1) {
      return { success: false }
    }
    pushUndoSnapshot()
    initTimelineClips(duration.value)
    return { success: true }
  }

  // --- Clip Library Actions ---

  async function addCurrentClip(customTitle?: string): Promise<Clip> {
    const clipIndex = clips.value.length + 1
    const padIndex = clipIndex.toString().padStart(2, '0')
    const activeTimeline = selectedTimelineClip.value
    const speed = activeTimeline ? activeTimeline.speed : 1.0
    const filter = activeTimeline ? { ...activeTimeline.filter } : undefined

    const title = customTitle?.trim() || `Đoạn cắt ${padIndex} (${Math.round(startTime.value/1000)}s - ${Math.round(endTime.value/1000)}s)`
    
    const newClipData = {
      video_id: currentVideo.value.id,
      title,
      start_time_ms: startTime.value,
      end_time_ms: endTime.value,
      duration_ms: clipDuration.value,
      thumbnail: currentVideo.value.thumbnail_url,
      status: 'ready' as const,
      file_path: currentVideo.value.source_path,
      speed,
      filter,
    }

    const created = await clipService.createClip(currentVideo.value.id, newClipData)
    clips.value.push(created)
    selectedClip.value = created

    // Add to history
    await historyService.add({
      video_id: currentVideo.value.id,
      video_title: currentVideo.value.title,
      action: 'Cắt video',
      clip_name: `${created.title} (${Math.round(created.duration_ms / 1000)}s)`,
      status: 'Đã cắt',
      file_size: Math.round((created.duration_ms / (currentVideo.value.duration_ms || 1)) * (currentVideo.value.file_size || 50000000)),
    })

    return created
  }

  async function removeClip(clipId: string) {
    await clipService.deleteClip(clipId, currentVideo.value.id)
    clips.value = clips.value.filter(c => c.id !== clipId)
    if (selectedClip.value?.id === clipId) {
      selectedClip.value = null
    }
  }

  function selectClip(clip: Clip) {
    selectedClip.value = clip
    startTime.value = clip.start_time_ms
    endTime.value = clip.end_time_ms
    currentTime.value = clip.start_time_ms
  }

  function downloadSingleClip(clip: Clip) {
    clipService.downloadClip(clip)
    historyService.add({
      video_id: currentVideo.value.id,
      video_title: currentVideo.value.title,
      action: 'Tải đoạn',
      clip_name: clip.title,
      status: 'Đã tải xuống',
      file_size: Math.round((clip.duration_ms / 1000) * 350000),
    })
  }

  function downloadAll() {
    clipService.downloadAllClips(clips.value)
    historyService.add({
      video_id: currentVideo.value.id,
      video_title: currentVideo.value.title,
      action: 'Tải tất cả',
      clip_name: `Toàn bộ ${clips.value.length} đoạn cắt`,
      status: 'Đã tải xuống',
      file_size: Math.round(clips.value.reduce((acc, c) => acc + c.duration_ms, 0) / 1000 * 350000),
    })
  }

  function updateVideoMeta(updates: Partial<Video>) {
    currentVideo.value = { ...currentVideo.value, ...updates }
  }

  return {
    currentVideo,
    currentTime,
    startTime,
    endTime,
    duration,
    isPlaying,
    volume,
    playbackRate,
    clips,
    selectedClip,
    aiSuggestions,
    selectedFormat,
    selectedQuality,
    clipDuration,
    clipDurationSec,
    timelineClips,
    selectedTimelineClipId,
    selectedTimelineClip,
    currentPlayingClip,
    totalTimelineOutputDurationMs,
    activeFilterCss,
    activeTool,
    subtitles,
    showSubtitles,
    isGeneratingSubtitles,
    isTranslatingMeta,
    dubbing,
    isGeneratingDubbing,
    isPlayingDubbedAudio,
    availableVoices,
    selectedVoiceId,
    loadDubbingVoices,
    setSelectedVoice,
    loadDubbing,
    generateDubbing,
    togglePlayDubbedAudio,
    currentSubtitleCue,
    currentSubtitleText,
    loadSubtitles,
    generateSubtitles,
    updateCueText,
    translateSingleCue,
    translateMeta,
    toggleShowSubtitles,
    undoStack,
    redoStack,
    canUndo,
    canRedo,
    undo,
    redo,
    setRange,
    setVideo,
    setDuration,
    setStartTime,
    setEndTime,
    setCurrentTime,
    setVolume,
    setPlaybackRate,
    resetRange,
    applyAISuggestion,
    addCurrentClip,
    removeClip,
    selectClip,
    downloadSingleClip,
    downloadAll,
    updateVideoMeta,
    initTimelineClips,
    selectTimelineClip,
    setActiveTool,
    splitClipAtCurrentTime,
    setClipSpeed,
    resetClipSpeed,
    setClipFilterPreset,
    setClipFilterAdjustment,
    resetClipFilter,
    deleteTimelineClip,
    unsplitSelectedClip,
    unsplitAllClips,
  }
})
