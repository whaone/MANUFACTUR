import { writable } from 'svelte/store'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

interface AuthUser {
  id: string
  workspace_id: string
  name: string
  email: string
  role: string
}

interface AuthState {
  authenticated: boolean
  user: AuthUser | null
  access_token: string | null
  refresh_token: string | null
}

const EMPTY: AuthState = { authenticated: false, user: null, access_token: null, refresh_token: null }

function loadState(): AuthState {
  if (typeof localStorage === 'undefined') return { ...EMPTY }
  try {
    const s = localStorage.getItem('manufactpro_auth')
    return s ? { ...EMPTY, ...JSON.parse(s) } : { ...EMPTY }
  } catch {
    return { ...EMPTY }
  }
}

function createAuthStore() {
  const initial = loadState()
  const { subscribe, set } = writable<AuthState>(initial)

  function persist(state: AuthState) {
    localStorage.setItem('manufactpro_auth', JSON.stringify(state))
  }

  return {
    subscribe,

    getToken(): string | null {
      return loadState().access_token
    },

    async login(email: string, password: string): Promise<void> {
      const res = await fetch(`${API_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Login failed' }))
        throw new Error(err.error ?? 'Login failed')
      }
      const data = await res.json()
      const state: AuthState = {
        authenticated: true,
        user: data.user,
        access_token: data.access_token,
        refresh_token: data.refresh_token ?? null,
      }
      set(state)
      persist(state)
    },

    // Called after OAuth callback — token already obtained, fetch /me to get user info
    async setTokens(accessToken: string, refreshToken: string): Promise<void> {
      const res = await fetch(`${API_URL}/api/auth/me`, {
        headers: { Authorization: `Bearer ${accessToken}` },
      })
      if (!res.ok) throw new Error('OAuth login: could not fetch user')
      const user: AuthUser = await res.json()
      const state: AuthState = { authenticated: true, user, access_token: accessToken, refresh_token: refreshToken || null }
      set(state)
      persist(state)
    },

    // Exchange refresh_token for a fresh token pair. Returns true on success.
    async refresh(): Promise<boolean> {
      const rt = loadState().refresh_token
      if (!rt) return false
      try {
        const res = await fetch(`${API_URL}/api/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: rt }),
        })
        if (!res.ok) return false
        const data = await res.json()
        const prev = loadState()
        const state: AuthState = {
          authenticated: true,
          user: data.user ?? prev.user,
          access_token: data.access_token,
          refresh_token: data.refresh_token ?? rt,
        }
        set(state)
        persist(state)
        return true
      } catch {
        return false
      }
    },

    logout() {
      set({ ...EMPTY })
      persist({ ...EMPTY })
      window.location.href = '/'
    },
  }
}

export const auth = createAuthStore()

// Derived helpers for backward compat
import { derived } from 'svelte/store'
export const isAuthenticated = derived(auth, ($a) => $a.authenticated)
