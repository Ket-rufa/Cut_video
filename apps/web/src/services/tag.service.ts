import { api } from './api'
import type { Tag } from '@/types'
import { DEFAULT_TAGS } from '@/utils/mockData'

let localTags: Tag[] = [...DEFAULT_TAGS]

export const tagService = {
  async list(): Promise<Tag[]> {
    try {
      return await api.get<Tag[]>('/tags')
    } catch {
      return [...localTags]
    }
  },

  async create(data: { name: string; color?: string }): Promise<Tag> {
    try {
      return await api.post<Tag>('/tags', data)
    } catch {
      const cleanName = data.name.replace(/^#/, '').trim()
      const newTag: Tag = {
        id: 't-' + Date.now().toString(36),
        name: cleanName,
        color: data.color || '#1677FF',
        count: 1,
      }
      localTags.push(newTag)
      return newTag
    }
  },

  async delete(id: string): Promise<void> {
    try {
      await api.delete(`/tags/${id}`)
    } catch {
      localTags = localTags.filter(t => t.id !== id)
    }
  }
}
