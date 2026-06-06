import { ref } from 'vue'

const STORAGE_KEY = 'staticman_history'
const MAX_ITEMS = 20

export interface HistoryItem {
  path: string
  name: string
  type: 'file' | 'directory'
  timestamp: number
}

function loadHistory(): HistoryItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      return JSON.parse(raw)
    }
  } catch {}
  return []
}

function saveHistory(items: HistoryItem[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items.slice(0, MAX_ITEMS)))
}

const history = ref<HistoryItem[]>(loadHistory())

export function useHistory() {
  function addHistory(item: Omit<HistoryItem, 'timestamp'>) {
    const existingIndex = history.value.findIndex(h => h.path === item.path)
    const newItem: HistoryItem = { ...item, timestamp: Date.now() }

    if (existingIndex !== -1) {
      // 更新已有项并移到顶部
      history.value.splice(existingIndex, 1)
    }
    history.value.unshift(newItem)

    // 限制数量
    if (history.value.length > MAX_ITEMS) {
      history.value = history.value.slice(0, MAX_ITEMS)
    }

    saveHistory(history.value)
  }

  function getHistory(): HistoryItem[] {
    return history.value
  }

  function clearHistory() {
    history.value = []
    localStorage.removeItem(STORAGE_KEY)
  }

  function formatRelativeTime(timestamp: number): string {
    const diff = Date.now() - timestamp
    const seconds = Math.floor(diff / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (seconds < 60) return '刚刚'
    if (minutes < 60) return `${minutes}分钟前`
    if (hours < 24) return `${hours}小时前`
    if (days < 7) return `${days}天前`
    return new Date(timestamp).toLocaleDateString('zh-CN')
  }

  return {
    history,
    addHistory,
    getHistory,
    clearHistory,
    formatRelativeTime,
  }
}
