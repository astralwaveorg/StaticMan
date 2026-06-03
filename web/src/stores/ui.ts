import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
  const showLogin = ref(false)
  const showCommand = ref(false)

  function openLogin() { showLogin.value = true }
  function closeLogin() { showLogin.value = false }
  function openCommand() { showCommand.value = true }
  function closeCommand() { showCommand.value = false }

  return { showLogin, showCommand, openLogin, closeLogin, openCommand, closeCommand }
})
