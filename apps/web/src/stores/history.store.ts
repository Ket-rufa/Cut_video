import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { HistoryItem } from '@/types'
import { historyService } from '@/services/history.service'
import { DEFAULT_HISTORY_ITEMS } from '@/utils/mockData'

export const useHistoryStore = defineStore('history', () => {
  const items = ref<HistoryItem[]>([...DEFAULT_HISTORY_ITEMS])
  const loading = ref(false)

  async function fetchHistory() {
    loading.value = true
    try {
      const data = await historyService.list()
      if (data && data.length > 0) {
        items.value = data
      }
    } finally {
      loading.value = false
    }
  }

  async function addItem(item: Omit<HistoryItem, 'id' | 'created_at'>) {
    const created = await historyService.add(item)
    items.value.unshift(created)
  }

  async function removeItem(id: string) {
    await historyService.delete(id)
    items.value = items.value.filter(i => i.id !== id)
  }

  return { items, loading, fetchHistory, addItem, removeItem }
})
