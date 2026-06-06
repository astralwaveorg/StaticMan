import { defineStore } from 'pinia'
import { ref } from 'vue'
import { setToken, clearToken, isLoggedIn as checkLoggedIn, authenticate } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const authenticated = ref(checkLoggedIn())

  async function login(password: string): Promise<boolean> {
    try {
      const { data } = await authenticate(password)
      setToken(data.token)
      authenticated.value = true
      // 触发全局登录成功事件，通知各组件刷新内容
      window.dispatchEvent(new CustomEvent('auth:login'))
      return true
    } catch { return false }
  }

  function logout() {
    clearToken()
    authenticated.value = false
    location.reload()
  }

  return { authenticated, login, logout }
})
