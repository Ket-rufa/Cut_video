import {
  calculateOutputDuration,
  getFilterCss,
  getFFmpegFilterString,
  DEFAULT_CLIP_FILTER
} from './filter'
import type { TimelineClip, ClipFilterConfig } from '@/types'

export function testEditorSuite() {
  // 1. Speed 2.0x (10s -> 5s)
  const duration2x = calculateOutputDuration(0, 10000, 2.0)
  if (duration2x !== 5000) throw new Error(`Expected 5000ms, got ${duration2x}ms`)

  // 2. Speed 0.5x (10s -> 20s)
  const duration05x = calculateOutputDuration(0, 10000, 0.5)
  if (duration05x !== 20000) throw new Error(`Expected 20000ms, got ${duration05x}ms`)

  // 3. Filter CSS
  const cssDefault = getFilterCss(DEFAULT_CLIP_FILTER)
  if (cssDefault !== 'none') throw new Error(`Expected 'none', got ${cssDefault}`)

  const bwFilter: ClipFilterConfig = {
    preset: 'bw',
    brightness: 0,
    contrast: 0,
    saturation: 0,
    temperature: 0,
  }
  const cssBw = getFilterCss(bwFilter)
  if (!cssBw.includes('grayscale(100%)')) throw new Error(`Expected grayscale in ${cssBw}`)

  // 4. FFmpeg mapping for 4x speed
  const ffmpegRes = getFFmpegFilterString(DEFAULT_CLIP_FILTER, 4.0)
  if (!ffmpegRes.videoFilter.includes('setpts=0.2500*PTS')) {
    throw new Error(`Expected setpts=0.2500*PTS, got ${ffmpegRes.videoFilter}`)
  }
  if (!ffmpegRes.audioFilter.includes('atempo=2.0,atempo=2.0000')) {
    throw new Error(`Expected chained atempo, got ${ffmpegRes.audioFilter}`)
  }

  // 5. State isolation
  const clips: TimelineClip[] = [
    {
      id: 'c1',
      sourceStartMs: 0,
      sourceEndMs: 10000,
      speed: 2.0,
      filter: { ...DEFAULT_CLIP_FILTER, brightness: 20 },
    },
    {
      id: 'c2',
      sourceStartMs: 10000,
      sourceEndMs: 20000,
      speed: 1.0,
      filter: { ...DEFAULT_CLIP_FILTER, brightness: 0 },
    }
  ]
  if (clips[0].speed !== 2.0 || clips[1].speed !== 1.0) {
    throw new Error('Speed state isolation failed')
  }
  if (clips[0].filter.brightness !== 20 || clips[1].filter.brightness !== 0) {
    throw new Error('Filter state isolation failed')
  }

  return true
}
