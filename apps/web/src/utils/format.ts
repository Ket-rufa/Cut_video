export function formatMs(ms: number): string {
  if (!ms || ms < 0) return '00:00.000'
  const totalSec = ms / 1000
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = Math.floor(totalSec % 60)
  const msPart = Math.floor(ms % 1000)
  if (h > 0) {
    return `${pad(h)}:${pad(m)}:${pad(s)}.${padMs(msPart)}`
  }
  return `${pad(m)}:${pad(s)}.${padMs(msPart)}`
}

export function formatTimeShort(ms: number): string {
  if (!ms || ms < 0) return '00:00'
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  if (h > 0) {
    return `${pad(h)}:${pad(m)}:${pad(s)}`
  }
  return `${pad(m)}:${pad(s)}`
}

export function formatDurationSec(ms: number): string {
  if (!ms || ms < 0) return '0s'
  const sec = Math.round(ms / 1000)
  if (sec < 60) return `${sec}s`
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return s > 0 ? `${m}m ${s}s` : `${m}m`
}

export function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    const day = pad(d.getDate())
    const month = pad(d.getMonth() + 1)
    const year = d.getFullYear()
    const hours = pad(d.getHours())
    const mins = pad(d.getMinutes())
    return `${day}/${month}/${year} ${hours}:${mins}`
  } catch {
    return dateStr
  }
}

function pad(n: number): string {
  return n.toString().padStart(2, '0')
}

function padMs(n: number): string {
  return n.toString().padStart(3, '0')
}

export function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

export function formatMbps(bytesPerSec: number): string {
  if (!bytesPerSec || bytesPerSec <= 0) return '0 Mbps'
  const mbps = (bytesPerSec * 8) / 1000000
  return `${mbps.toFixed(1)} Mbps`
}

export function platformLabel(p: string): string {
  const map: Record<string, string> = {
    youtube: 'YouTube',
    tiktok: 'TikTok',
    douyin: 'Douyin',
    facebook: 'Facebook',
    instagram: 'Instagram',
    twitter: 'Twitter (X)',
    generic: 'Direct URL',
    unknown: 'Khác',
  }
  return map[p] ?? p
}
