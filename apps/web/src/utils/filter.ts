import type { ClipFilterConfig, FilterPresetType } from '@/types'

export const DEFAULT_CLIP_FILTER: ClipFilterConfig = {
  preset: 'none',
  brightness: 0,
  contrast: 0,
  saturation: 0,
  temperature: 0,
}

export interface FilterPresetInfo {
  id: FilterPresetType
  name: string
  previewStyle: string
  values: {
    brightness: number
    contrast: number
    saturation: number
    temperature: number
  }
}

export const FILTER_PRESETS: FilterPresetInfo[] = [
  {
    id: 'none',
    name: 'Gốc',
    previewStyle: 'background: linear-gradient(135deg, #64748b 0%, #334155 100%);',
    values: { brightness: 0, contrast: 0, saturation: 0, temperature: 0 },
  },
  {
    id: 'vivid',
    name: 'Tươi sáng',
    previewStyle: 'background: linear-gradient(135deg, #f59e0b 0%, #ec4899 50%, #3b82f6 100%);',
    values: { brightness: 10, contrast: 15, saturation: 30, temperature: 5 },
  },
  {
    id: 'warm',
    name: 'Ấm áp',
    previewStyle: 'background: linear-gradient(135deg, #fb923c 0%, #d97706 100%);',
    values: { brightness: 5, contrast: 10, saturation: 15, temperature: 40 },
  },
  {
    id: 'cool',
    name: 'Lạnh',
    previewStyle: 'background: linear-gradient(135deg, #38bdf8 0%, #1d4ed8 100%);',
    values: { brightness: 0, contrast: 10, saturation: -5, temperature: -40 },
  },
  {
    id: 'cinematic',
    name: 'Điện ảnh',
    previewStyle: 'background: linear-gradient(135deg, #0f766e 0%, #1e1b4b 100%);',
    values: { brightness: -5, contrast: 25, saturation: -15, temperature: -10 },
  },
  {
    id: 'vintage',
    name: 'Vintage',
    previewStyle: 'background: linear-gradient(135deg, #d97706 0%, #78350f 100%);',
    values: { brightness: 5, contrast: -10, saturation: -20, temperature: 25 },
  },
  {
    id: 'bw',
    name: 'Đen trắng',
    previewStyle: 'background: linear-gradient(135deg, #94a3b8 0%, #0f172a 100%);',
    values: { brightness: 5, contrast: 20, saturation: -100, temperature: 0 },
  },
]

/**
 * Generates the CSS filter string for real-time video element preview.
 */
export function getFilterCss(filter?: ClipFilterConfig): string {
  if (!filter) return 'none'

  // Combine preset adjustments with manual adjustments if preset is active
  const presetInfo = FILTER_PRESETS.find(p => p.id === filter.preset)
  const base = presetInfo ? presetInfo.values : { brightness: 0, contrast: 0, saturation: 0, temperature: 0 }

  const b = Math.max(-100, Math.min(100, base.brightness + (filter.brightness || 0)))
  const c = Math.max(-100, Math.min(100, base.contrast + (filter.contrast || 0)))
  const s = Math.max(-100, Math.min(100, base.saturation + (filter.saturation || 0)))
  const t = Math.max(-100, Math.min(100, base.temperature + (filter.temperature || 0)))

  const cssFilters: string[] = []

  // Brightness: -100..100 -> 0.0 .. 2.0 (1.0 is neutral)
  const brightnessVal = 1 + b / 100
  if (Math.abs(brightnessVal - 1) > 0.001) {
    cssFilters.push(`brightness(${brightnessVal.toFixed(2)})`)
  }

  // Contrast: -100..100 -> 0.0 .. 2.0 (1.0 is neutral)
  const contrastVal = 1 + c / 100
  if (Math.abs(contrastVal - 1) > 0.001) {
    cssFilters.push(`contrast(${contrastVal.toFixed(2)})`)
  }

  // Saturation: -100..100 -> 0.0 .. 2.0 (1.0 is neutral)
  const saturateVal = Math.max(0, 1 + s / 100)
  if (filter.preset === 'bw') {
    cssFilters.push(`grayscale(100%)`)
  } else if (Math.abs(saturateVal - 1) > 0.001) {
    cssFilters.push(`saturate(${saturateVal.toFixed(2)})`)
  }

  // Temperature simulation in CSS
  if (filter.preset === 'vintage') {
    cssFilters.push(`sepia(0.4)`)
  } else if (t > 0) {
    // Warm: slight sepia & hue rotate towards warm
    const sepiaAmt = (t / 100) * 0.25
    cssFilters.push(`sepia(${sepiaAmt.toFixed(2)})`)
  } else if (t < 0) {
    // Cool: hue rotate slightly towards cyan/blue
    const hueDeg = (t / 100) * 15
    cssFilters.push(`hue-rotate(${hueDeg.toFixed(1)}deg)`)
  }

  return cssFilters.length > 0 ? cssFilters.join(' ') : 'none'
}

/**
 * Calculates output duration based on source start/end and speed multiplier.
 */
export function calculateOutputDuration(sourceStartMs: number, sourceEndMs: number, speed: number): number {
  const srcDuration = Math.max(0, sourceEndMs - sourceStartMs)
  const safeSpeed = speed > 0 ? speed : 1
  return Math.round(srcDuration / safeSpeed)
}

/**
 * Maps normalized filter config and speed to FFmpeg filter string options.
 */
export function getFFmpegFilterString(filter: ClipFilterConfig, speed: number): { videoFilter: string; audioFilter: string } {
  const videoFilters: string[] = []
  const audioFilters: string[] = []

  // Video Speed: setpts=(1/speed)*PTS
  if (speed > 0 && Math.abs(speed - 1.0) > 0.001) {
    const ptsFactor = (1.0 / speed).toFixed(4)
    videoFilters.push(`setpts=${ptsFactor}*PTS`)

    // Audio Speed: chain atempo filters (supported range per filter is 0.5 to 2.0)
    let currentSpeed = speed
    while (currentSpeed > 2.0) {
      audioFilters.push('atempo=2.0')
      currentSpeed /= 2.0
    }
    while (currentSpeed < 0.5) {
      audioFilters.push('atempo=0.5')
      currentSpeed /= 0.5
    }
    if (Math.abs(currentSpeed - 1.0) > 0.001) {
      audioFilters.push(`atempo=${currentSpeed.toFixed(4)}`)
    }
  }

  // Color Filter Mapping to FFmpeg `eq` and `colorbalance`
  const presetInfo = FILTER_PRESETS.find(p => p.id === filter.preset)
  const base = presetInfo ? presetInfo.values : { brightness: 0, contrast: 0, saturation: 0, temperature: 0 }

  const b = Math.max(-100, Math.min(100, base.brightness + (filter.brightness || 0)))
  const c = Math.max(-100, Math.min(100, base.contrast + (filter.contrast || 0)))
  const s = Math.max(-100, Math.min(100, base.saturation + (filter.saturation || 0)))
  const t = Math.max(-100, Math.min(100, base.temperature + (filter.temperature || 0)))

  // FFmpeg eq filter: brightness (-1.0 to 1.0, 0 is default), contrast (-1000 to 1000, 1.0 is default), saturation (0 to 3.0, 1.0 is default)
  const ffBrightness = (b / 200).toFixed(2) // -0.5 to +0.5
  const ffContrast = (1 + c / 100).toFixed(2)
  const ffSaturation = filter.preset === 'bw' ? '0' : Math.max(0, 1 + s / 100).toFixed(2)

  const eqParts: string[] = []
  if (Math.abs(Number(ffBrightness)) > 0.01) eqParts.push(`brightness=${ffBrightness}`)
  if (Math.abs(Number(ffContrast) - 1.0) > 0.01) eqParts.push(`contrast=${ffContrast}`)
  if (Math.abs(Number(ffSaturation) - 1.0) > 0.01) eqParts.push(`saturation=${ffSaturation}`)

  if (eqParts.length > 0) {
    videoFilters.push(`eq=${eqParts.join(':')}`)
  }

  // Temperature in FFmpeg: colorbalance (rs=red shadows, rm=red midtones, rh=red highlights, bs=blue shadows, bm=blue midtones, bh=blue highlights)
  if (Math.abs(t) > 5) {
    const factor = (t / 100) * 0.3
    if (factor > 0) {
      // Warm: boost red, reduce blue
      videoFilters.push(`colorbalance=rm=${factor.toFixed(2)}:gm=0.0:bm=-${(factor * 0.8).toFixed(2)}`)
    } else {
      // Cool: boost blue, reduce red
      const absFactor = Math.abs(factor)
      videoFilters.push(`colorbalance=rm=-${(absFactor * 0.8).toFixed(2)}:gm=0.0:bm=${absFactor.toFixed(2)}`)
    }
  }

  return {
    videoFilter: videoFilters.join(','),
    audioFilter: audioFilters.join(','),
  }
}
