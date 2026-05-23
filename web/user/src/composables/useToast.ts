import { ref, readonly } from 'vue'

export type ToastType = 'success' | 'error' | 'warn' | 'info'

export interface Toast {
  id: number
  message: string
  type: ToastType
}

const toasts = ref<Toast[]>([])
let _id = 0

export function useToast() {
  function show(message: string, type: ToastType = 'info', duration = 3500) {
    const id = ++_id
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  function success(msg: string) { show(msg, 'success') }
  function error(msg: string) { show(msg, 'error') }
  function warn(msg: string) { show(msg, 'warn') }
  function info(msg: string) { show(msg, 'info') }
  function dismiss(id: number) { toasts.value = toasts.value.filter(t => t.id !== id) }

  return { toasts: readonly(toasts), show, success, error, warn, info, dismiss }
}
