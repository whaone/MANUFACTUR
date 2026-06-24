import { writable } from 'svelte/store'

function createOfflineStore() {
  const { subscribe, set } = writable(typeof navigator !== 'undefined' ? !navigator.onLine : false)

  if (typeof window !== 'undefined') {
    const setOffline = () => set(true)
    const setOnline = () => set(false)

    window.addEventListener('online', setOnline)
    window.addEventListener('offline', setOffline)

    return { subscribe }
  }

  return { subscribe }
}

export const isOffline = createOfflineStore()
