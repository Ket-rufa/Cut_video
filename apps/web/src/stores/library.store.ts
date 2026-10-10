import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Folder, Tag } from '@/types'
import { folderService } from '@/services/folder.service'
import { tagService } from '@/services/tag.service'
import { DEFAULT_FOLDERS, DEFAULT_TAGS } from '@/utils/mockData'

export const useLibraryStore = defineStore('library', () => {
  const folders = ref<Folder[]>([...DEFAULT_FOLDERS])
  const tags = ref<Tag[]>([...DEFAULT_TAGS])
  const selectedFolderId = ref<string>('f-all')
  const loading = ref(false)

  async function fetchAll() {
    loading.value = true
    try {
      const [f, t] = await Promise.all([folderService.list(), tagService.list()])
      if (f && f.length > 0) folders.value = f
      if (t && t.length > 0) tags.value = t
    } finally {
      loading.value = false
    }
  }

  async function addFolder(name: string, color?: string, icon?: string): Promise<Folder> {
    const created = await folderService.create({ name, color, icon })
    folders.value.push(created)
    return created
  }

  async function removeFolder(id: string) {
    await folderService.delete(id)
    folders.value = folders.value.filter(f => f.id !== id)
    if (selectedFolderId.value === id) {
      selectedFolderId.value = 'f-all'
    }
  }

  async function addTag(name: string, color?: string): Promise<Tag> {
    const created = await tagService.create({ name, color })
    tags.value.push(created)
    return created
  }

  async function removeTag(id: string) {
    await tagService.delete(id)
    tags.value = tags.value.filter(t => t.id !== id)
  }

  return {
    folders,
    tags,
    selectedFolderId,
    loading,
    fetchAll,
    addFolder,
    removeFolder,
    addTag,
    removeTag,
  }
})
