import { defineStore } from 'pinia'
import { ref } from 'vue'
import { setKey, clearKey, isLoggedIn as checkLoggedIn, authenticate } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const authenticated = ref(checkLoggedIn())

  async function login(password: string): Promise<boolean> {
    try {
      const { data } = await authenticate(password)
      setKey(data.key)
      authenticated.value = true
      // 触发全局登录成功事件，通知各组件刷新内容
      window.dispatchEvent(new CustomEvent('auth:login'))
      return true
    } catch { return false }
  }

  function logout() {
    clearKey()
    authenticated.value = false
    location.reload()
  }

  return { authenticated, login, logout }
})
