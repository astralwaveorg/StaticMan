import { ref } from 'vue'

const STORAGE_KEY = 'staticman_favorites'

export interface FavoriteItem {
  path: string
  name: string
  type: 'file' | 'directory'
  addedAt: number
}

function loadFavorites(): FavoriteItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      return JSON.parse(raw)
    }
  } catch {}
  return []
}

function saveFavorites(items: FavoriteItem[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(items))
}

const favorites = ref<FavoriteItem[]>(loadFavorites())

export function useFavorites() {
  function isFavorite(path: string): boolean {
    return favorites.value.some(f => f.path === path)
  }

  function toggleFavorite(item: Omit<FavoriteItem, 'addedAt'>) {
    const index = favorites.value.findIndex(f => f.path === item.path)
    if (index !== -1) {
      favorites.value.splice(index, 1)
    } else {
      favorites.value.push({ ...item, addedAt: Date.now() })
    }
    saveFavorites(favorites.value)
  }

  function addFavorite(item: Omit<FavoriteItem, 'addedAt'>) {
    if (!isFavorite(item.path)) {
      favorites.value.push({ ...item, addedAt: Date.now() })
      saveFavorites(favorites.value)
    }
  }

  function removeFavorite(path: string) {
    const index = favorites.value.findIndex(f => f.path === path)
    if (index !== -1) {
      favorites.value.splice(index, 1)
      saveFavorites(favorites.value)
    }
  }

  function getFavorites(): FavoriteItem[] {
    return favorites.value
  }

  return {
    favorites,
    isFavorite,
    toggleFavorite,
    addFavorite,
    removeFavorite,
    getFavorites,
  }
}
