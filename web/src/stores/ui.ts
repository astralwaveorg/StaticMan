import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
  const showLogin = ref(false)
  const showCommand = ref(false)
  const commandQuery = ref('')
  const commandMode = ref<'name' | 'content'>('name')

  function openLogin() { showLogin.value = true }
  function closeLogin() { showLogin.value = false }
  function openCommand(query = '', mode?: 'name' | 'content') {
    commandQuery.value = query
    if (mode) commandMode.value = mode
    showCommand.value = true
  }
  function closeCommand() { showCommand.value = false }

  return {
    showLogin, showCommand,
    commandQuery, commandMode,
    openLogin, closeLogin, openCommand, closeCommand
  }
})
