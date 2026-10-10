import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Video } from '@/types'
import { videoService } from '@/services/video.service'
import { historyService } from '@/services/history.service'
import { formatBytes } from '@/utils/format'
import { DEFAULT_SAVED_VIDEOS, DEFAULT_CURRENT_VIDEO } from '@/utils/mockData'

export const useVideoStore = defineStore('video', () => {
  const currentVideo = ref<Video>({ ...DEFAULT_CURRENT_VIDEO })
  const videos = ref<Video[]>([...DEFAULT_SAVED_VIDEOS])
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  // Filters & State
  const activePlatform = ref<string>('all')
  const searchQuery = ref('')
  const selectedFolder = ref<string>('all')
  const selectedTag = ref<string>('all')
  const viewMode = ref<'grid' | 'list'>('grid')
  const sortBy = ref<'newest' | 'oldest' | 'duration' | 'title'>('newest')

  // Global Add Video Modal State
  const isAddModalOpen = ref(false)
  const addModalInitialUrl = ref('')

  function openAddModal(url?: string) {
    addModalInitialUrl.value = url || ''
    isAddModalOpen.value = true
  }

  const filteredVideos = computed(() => {
    let list = [...videos.value]
    
    // Platform filter
    if (activePlatform.value && activePlatform.value !== 'all') {
      list = list.filter(v => v.platform === activePlatform.value)
    }
    
    // Folder filter
    if (selectedFolder.value && selectedFolder.value !== 'all' && selectedFolder.value !== 'f-all') {
      list = list.filter(v => v.folder_id === selectedFolder.value)
    }
    
    // Tag filter
    if (selectedTag.value && selectedTag.value !== 'all') {
      if (selectedTag.value === 'untagged') {
        list = list.filter(v => !v.tags || v.tags.length === 0)
      } else {
        const targetTag = selectedTag.value.replace(/^#/, '').trim().toLowerCase()
        list = list.filter(v => 
          Array.isArray(v.tags) && v.tags.some(t => t.replace(/^#/, '').trim().toLowerCase() === targetTag)
        )
      }
    }
    
    // Search query
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase().trim()
      list = list.filter(v => 
        v.title.toLowerCase().includes(q) || 
        v.description.toLowerCase().includes(q) ||
        (Array.isArray(v.tags) && v.tags.some(t => t.toLowerCase().includes(q)))
      )
    }
    
    // Sort
    if (sortBy.value === 'newest') {
      list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    } else if (sortBy.value === 'oldest') {
      list.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
    } else if (sortBy.value === 'duration') {
      list.sort((a, b) => (b.duration_ms || 0) - (a.duration_ms || 0))
    } else if (sortBy.value === 'title') {
      list.sort((a, b) => a.title.localeCompare(b.title))
    }
    
    return list
  })

  // Dynamic real tags extracted from saved videos with counts
  const availableTags = computed(() => {
    const tagCountMap = new Map<string, number>()
    for (const v of videos.value) {
      if (Array.isArray(v.tags)) {
        for (const t of v.tags) {
          const cleaned = t.replace(/^#/, '').trim()
          if (cleaned) {
            tagCountMap.set(cleaned, (tagCountMap.get(cleaned) || 0) + 1)
          }
        }
      }
    }
    return Array.from(tagCountMap.entries())
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
  })

  // Quick filter counts
  const countAll = computed(() => videos.value.length)
  const countUntagged = computed(() => videos.value.filter(v => !v.tags || v.tags.length === 0).length)
  const countFavorites = computed(() => videos.value.filter(v => v.is_favorite).length)
  const countTrash = computed(() => videos.value.filter(v => v.is_trash).length)

  // Real Storage & Time Metrics
  const totalClipsCount = ref(0)
  const maxStorageBytes = 50 * 1024 * 1024 * 1024 // 50 GB limit

  const totalStorageBytes = computed(() => {
    return videos.value.reduce((acc, v) => acc + (v.file_size || 0), 0)
  })

  const storageFormatted = computed(() => {
    return formatBytes(totalStorageBytes.value)
  })

  const storagePercent = computed(() => {
    if (maxStorageBytes === 0) return 0
    return Math.min(100, Number(((totalStorageBytes.value / maxStorageBytes) * 100).toFixed(1)))
  })

  const storageFreeFormatted = computed(() => {
    return formatBytes(Math.max(0, maxStorageBytes - totalStorageBytes.value))
  })

  const totalDurationMs = computed(() => {
    return videos.value.reduce((acc, v) => acc + (v.duration_ms || 0), 0)
  })

  const totalHoursSaved = computed(() => {
    const hours = (totalDurationMs.value / (1000 * 3600)).toFixed(1)
    return `~${hours} giờ`
  })

  async function fetchHistoryMetrics() {
    try {
      const hist = await historyService.list()
      if (Array.isArray(hist)) {
        const cutLogs = hist.filter(h => h.action?.includes('Cắt') || h.status?.includes('cắt'))
        totalClipsCount.value = cutLogs.length > 0 ? cutLogs.length : hist.length
      }
    } catch {
      // ignore
    }
  }

  async function fetchAllVideos() {
    loading.value = true
    error.value = null
    try {
      const data = await videoService.list()
      if (data && data.length > 0) {
        videos.value = data
      }
      await fetchHistoryMetrics()
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function resolve(url: string): Promise<Video> {
    loading.value = true
    error.value = null
    try {
      const video = await videoService.resolve(url)
      currentVideo.value = video
      // Prepend if not exists
      if (!videos.value.some(v => v.id === video.id)) {
        videos.value.unshift(video)
      }
      return video
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function upload(file: File): Promise<Video> {
    loading.value = true
    error.value = null
    try {
      const video = await videoService.upload(file)
      currentVideo.value = video
      videos.value.unshift(video)
      return video
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchById(id: string): Promise<Video> {
    loading.value = true
    error.value = null
    try {
      const video = await videoService.getById(id)
      currentVideo.value = video
      return video
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteVideo(id: string) {
    await videoService.delete(id)
    videos.value = videos.value.filter(v => v.id !== id)
  }

  async function updateVideo(id: string, updates: Partial<Video>) {
    const updated = await videoService.update(id, updates)
    const idx = videos.value.findIndex(v => v.id === id)
    if (idx >= 0) {
      videos.value[idx] = { ...videos.value[idx], ...updated }
    }
    if (currentVideo.value.id === id) {
      currentVideo.value = { ...currentVideo.value, ...updated }
    }
    return updated
  }

  function toggleFavorite(id: string) {
    const v = videos.value.find(item => item.id === id)
    if (v) {
      v.is_favorite = !v.is_favorite
    }
  }

  function setPlatformFilter(platform: string) {
    activePlatform.value = platform
  }

  function setFolderFilter(folderId: string) {
    selectedFolder.value = folderId
  }

  return {
    currentVideo,
    videos,
    loading,
    error,
    activePlatform,
    searchQuery,
    selectedFolder,
    selectedTag,
    viewMode,
    sortBy,
    isAddModalOpen,
    addModalInitialUrl,
    openAddModal,
    filteredVideos,
    availableTags,
    countAll,
    countUntagged,
    countFavorites,
    countTrash,
    totalStorageBytes,
    maxStorageBytes,
    storageFormatted,
    storagePercent,
    storageFreeFormatted,
    totalDurationMs,
    totalHoursSaved,
    totalClipsCount,
    fetchHistoryMetrics,
    fetchAllVideos,
    resolve,
    upload,
    fetchById,
    deleteVideo,
    updateVideo,
    toggleFavorite,
    setPlatformFilter,
    setFolderFilter,
  }
})
