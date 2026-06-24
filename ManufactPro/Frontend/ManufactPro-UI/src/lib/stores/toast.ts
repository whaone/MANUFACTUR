import { writable } from 'svelte/store'

export type ToastKind = 'success' | 'error' | 'warning' | 'info'

export interface Toast {
  id: string
  kind: ToastKind
  title: string
  lines: string[]
  timeout: number
}

function createToastStore() {
  const { subscribe, update } = writable<Toast[]>([])

  function dismiss(id: string) {
    update((list) => list.filter((t) => t.id !== id))
  }

  function push(kind: ToastKind, title: string, detail?: string | string[], timeout = 5000) {
    const id = crypto.randomUUID()
    const lines = detail == null ? [] : Array.isArray(detail) ? detail : [detail]
    update((list) => [...list, { id, kind, title, lines, timeout }])
    if (timeout > 0 && typeof window !== 'undefined') {
      setTimeout(() => dismiss(id), timeout)
    }
    return id
  }

  return {
    subscribe,
    dismiss,
    success: (title: string, detail?: string | string[]) => push('success', title, detail),
    error: (title: string, detail?: string | string[]) => push('error', title, detail, 8000),
    warning: (title: string, detail?: string | string[]) => push('warning', title, detail, 8000),
    info: (title: string, detail?: string | string[]) => push('info', title, detail),
  }
}

export const toast = createToastStore()
