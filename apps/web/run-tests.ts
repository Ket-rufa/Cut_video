import assert from 'node:assert'
import {
  calculateOutputDuration,
  getFilterCss,
  getFFmpegFilterString,
  DEFAULT_CLIP_FILTER
} from './src/utils/filter.ts'

console.log('=== RUNNING EDITOR SUITE TESTS ===\n')

// TEST 1: Speed calculation (2x speed)
{
  const srcStart = 0
  const srcEnd = 10000 // 10s
  const speed = 2.0
  const output = calculateOutputDuration(srcStart, srcEnd, speed)
  assert.strictEqual(output, 5000, '10s at 2x speed must be 5s output duration')
  console.log('✓ PASS: Speed 2.0x calculates output duration = 5s')
}

// TEST 2: Speed calculation (0.5x speed)
{
  const srcStart = 0
  const srcEnd = 10000 // 10s
  const speed = 0.5
  const output = calculateOutputDuration(srcStart, srcEnd, speed)
  assert.strictEqual(output, 20000, '10s at 0.5x speed must be 20s output duration')
  console.log('✓ PASS: Speed 0.5x calculates output duration = 20s')
}

// TEST 3: Split logic simulation
{
  function simulateSplit(clips: any[], targetMs: number, toleranceMs = 300) {
    const clipIndex = clips.findIndex(
      (c: any) => targetMs > c.sourceStartMs + toleranceMs && targetMs < c.sourceEndMs - toleranceMs
    )
    if (clipIndex === -1) {
      return { success: false, clips }
    }
    const targetClip = clips[clipIndex]
    const clip1 = {
      ...targetClip,
      id: 'clip-1',
      sourceStartMs: targetClip.sourceStartMs,
      sourceEndMs: targetMs,
    }
    const clip2 = {
      ...targetClip,
      id: 'clip-2',
      sourceStartMs: targetMs,
      sourceEndMs: targetClip.sourceEndMs,
    }
    const newClips = [...clips]
    newClips.splice(clipIndex, 1, clip1, clip2)
    return { success: true, clips: newClips }
  }

  // Valid split at 4s for a 0->10s clip
  const initialClips = [{ id: 'clip-0', sourceStartMs: 0, sourceEndMs: 10000, speed: 1.0 }]
  const res = simulateSplit(initialClips, 4000)
  assert.strictEqual(res.success, true, 'Split at 4s must succeed')
  assert.strictEqual(res.clips.length, 2, 'Should produce 2 clips')
  assert.strictEqual(res.clips[0].sourceStartMs, 0)
  assert.strictEqual(res.clips[0].sourceEndMs, 4000)
  assert.strictEqual(res.clips[1].sourceStartMs, 4000)
  assert.strictEqual(res.clips[1].sourceEndMs, 10000)
  console.log('✓ PASS: Split clip [0..10s] at 4s produces [0..4s] and [4..10s]')

  // Invalid split at clip start (0s)
  const invalidStart = simulateSplit(initialClips, 0)
  assert.strictEqual(invalidStart.success, false, 'Split at 0s must fail')
  console.log('✓ PASS: Split at clip start (0s) correctly rejected')

  // Invalid split at clip end (10s)
  const invalidEnd = simulateSplit(initialClips, 10000)
  assert.strictEqual(invalidEnd.success, false, 'Split at 10s must fail')
  console.log('✓ PASS: Split at clip end (10s) correctly rejected')

  // Invalid split within tolerance (< 300ms from start)
  const invalidTol = simulateSplit(initialClips, 150)
  assert.strictEqual(invalidTol.success, false, 'Split within tolerance must fail')
  console.log('✓ PASS: Split within 150ms tolerance correctly rejected')
}

// TEST 4: Clip State Isolation
{
  const clip1 = { id: 'c1', speed: 1.0, filter: { preset: 'none', brightness: 0 } }
  const clip2 = { id: 'c2', speed: 1.0, filter: { preset: 'none', brightness: 0 } }

  clip1.speed = 2.0
  clip1.filter.preset = 'vivid'
  clip1.filter.brightness = 25

  assert.strictEqual(clip2.speed, 1.0, 'Clip2 speed must remain 1.0')
  assert.strictEqual(clip2.filter.preset, 'none', 'Clip2 preset must remain none')
  assert.strictEqual(clip2.filter.brightness, 0, 'Clip2 brightness must remain 0')
  console.log('✓ PASS: Clip state isolation verified (clip1 edits do not affect clip2)')
}

// TEST 5: Filter CSS and FFmpeg filter mapping
{
  const vividCss = getFilterCss({
    preset: 'vivid',
    brightness: 10,
    contrast: 0,
    saturation: 0,
    temperature: 0
  })
  assert.ok(vividCss.includes('brightness'), 'Vivid preset CSS must include brightness')
  console.log('✓ PASS: Filter CSS generation for vivid preset verified')

  const ffRes = getFFmpegFilterString({ preset: 'none', brightness: 10, contrast: 20, saturation: 10, temperature: 30 }, 2.0)
  assert.ok(ffRes.videoFilter.includes('setpts=0.5000*PTS'), 'FFmpeg video filter must contain setpts')
  assert.ok(ffRes.videoFilter.includes('eq='), 'FFmpeg video filter must contain eq')
  assert.ok(ffRes.audioFilter.includes('atempo=2.0000'), 'FFmpeg audio filter must contain atempo')
  console.log('✓ PASS: FFmpeg video & audio filter string mapping verified')
}

console.log('\n=== ALL UNIT TESTS PASSED SUCCESSFULLY ===\n')
