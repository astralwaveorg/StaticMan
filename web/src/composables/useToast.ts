import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: number
  message: string
  type: ToastType
}

const toasts = ref<Toast[]>([])
let idCounter = 0

export function useToast() {
  function show(message: string, type: ToastType = 'info', duration = 3000) {
    const id = ++idCounter
    const toast: Toast = { id, message, type }
    toasts.value.push(toast)
    if (duration > 0) {
      setTimeout(() => {
        remove(id)
      }, duration)
    }
    return id
  }

  function remove(id: number) {
    const idx = toasts.value.findIndex(t => t.id === id)
    if (idx !== -1) {
      toasts.value.splice(idx, 1)
    }
  }

  function success(message: string, duration?: number) {
    return show(message, 'success', duration)
  }
  function error(message: string, duration?: number) {
    return show(message, 'error', duration)
  }
  function warning(message: string, duration?: number) {
    return show(message, 'warning', duration)
  }
  function info(message: string, duration?: number) {
    return show(message, 'info', duration)
  }

  return { toasts, show, remove, success, error, warning, info }
}

export { toasts }
