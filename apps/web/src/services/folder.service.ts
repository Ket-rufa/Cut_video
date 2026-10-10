import { api } from './api'
import type { Folder } from '@/types'
import { DEFAULT_FOLDERS } from '@/utils/mockData'

let localFolders: Folder[] = [...DEFAULT_FOLDERS]

export const folderService = {
  async list(): Promise<Folder[]> {
    try {
      return await api.get<Folder[]>('/folders')
    } catch {
      return [...localFolders]
    }
  },

  async create(data: { name: string; color?: string; icon?: string }): Promise<Folder> {
    try {
      return await api.post<Folder>('/folders', data)
    } catch {
      const newFolder: Folder = {
        id: 'f-' + Date.now().toString(36),
        name: data.name,
        color: data.color || '#1677FF',
        icon: data.icon || '📁',
        count: 0,
        created_at: new Date().toISOString(),
      }
      localFolders.push(newFolder)
      return newFolder
    }
  },

  async delete(id: string): Promise<void> {
    try {
      await api.delete(`/folders/${id}`)
    } catch {
      localFolders = localFolders.filter(f => f.id !== id)
    }
  }
}
