import { api } from './api'
import type { HistoryItem } from '@/types'
import { DEFAULT_HISTORY_ITEMS } from '@/utils/mockData'

let localHistory: HistoryItem[] = [...DEFAULT_HISTORY_ITEMS]

export const historyService = {
  async list(): Promise<HistoryItem[]> {
    try {
      return await api.get<HistoryItem[]>('/history')
    } catch {
      return [...localHistory]
    }
  },

  async add(item: Omit<HistoryItem, 'id' | 'created_at'>): Promise<HistoryItem> {
    try {
      return await api.post<HistoryItem>('/history', item)
    } catch {
      const newItem: HistoryItem = {
        ...item,
        id: 'hist-' + Date.now().toString(36),
        created_at: new Date().toISOString(),
      }
      localHistory.unshift(newItem)
      return newItem
    }
  },

  async delete(id: string): Promise<void> {
    localHistory = localHistory.filter(h => h.id !== id)
  }
}
