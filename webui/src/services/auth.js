import { computed, reactive } from 'vue'
import {
  readToken as readPersistedToken,
  clearAuthArtifacts,
  persistAuthToken,
  getMyProfile,
  doLogin as apiLogin,
  abortActiveRequests,
} from './api'

const state = reactive({
  token: '',
  user: null,
  isReady: false,
  isLoading: false,
  status: 'idle', // idle | loading | authenticated | anonymous
})

let initPromise = null

function setToken(token, { emit = true } = {}) {
  const normalized = typeof token === 'string' ? token.trim() : String(token || '').trim()
  if (!normalized) {
    clearState({ emit })
    return
  }

  persistAuthToken(normalized, { emit })
  state.token = normalized
}

function clearState({ emit = true } = {}) {
  abortActiveRequests('Clearing auth state')
  clearAuthArtifacts({ emit })
  state.token = ''
  state.user = null
  state.status = 'anonymous'
}

async function syncUserProfile() {
  if (!state.token) {
    state.user = null
    return null
  }

  const profile = await getMyProfile()
  state.user = profile
  if (typeof window !== 'undefined') {
    localStorage.setItem('name', profile?.name || '')
    localStorage.setItem('me', JSON.stringify(profile || {}))
  }
  return profile
}

async function hydrateFromStorage() {
  const token = readPersistedToken()
  if (token) {
    setToken(token, { emit: false })
    await syncUserProfile().catch(() => {
      clearState({ emit: true })
    })
    state.status = state.token ? 'authenticated' : 'anonymous'
  } else {
    clearState({ emit: false })
  }
}

export async function ensureAuthReady() {
  if (state.isReady && !state.isLoading) return state
  if (!initPromise) {
    initPromise = (async () => {
      state.isLoading = true
      state.status = 'loading'
      await hydrateFromStorage()
      state.isReady = true
      state.isLoading = false
      if (!state.token && state.status !== 'authenticated') {
        state.status = 'anonymous'
      }
      return state
    })().finally(() => { initPromise = null })
  }
  return initPromise
}

export async function login(name) {
  state.isLoading = true
  state.status = 'loading'
  clearState({ emit: false })
  try {
    const result = await apiLogin(name, { persist: false })
    const token = result?.token || result
    setToken(token, { emit: true })
    await syncUserProfile()
    state.isReady = true
    state.status = 'authenticated'
    return state
  } finally {
    state.isLoading = false
    if (!state.token && state.status !== 'authenticated') {
      state.status = 'anonymous'
    }
  }
}

export function logout() {
  clearState({ emit: true })
  state.isReady = true
  state.isLoading = false
  state.status = 'anonymous'
}

export async function refreshProfile() {
  if (!state.token) {
    clearState({ emit: false })
    return null
  }

  try {
    return await syncUserProfile()
  } catch (e) {
    clearState({ emit: true })
    throw e
  }
}

export const isAuthenticated = computed(() => !!state.token)
export const isReady = computed(() => state.isReady)
export const isAuthLoading = computed(() => state.isLoading)
export const currentUser = computed(() => state.user)

export function useAuthState() {
  return {
    state,
    isAuthenticated,
    isReady,
    isAuthLoading,
    currentUser,
    ensureAuthReady,
    login,
    logout,
    refreshProfile,
  }
}

export default {
  state,
  isAuthenticated,
  isReady,
  isAuthLoading,
  currentUser,
  ensureAuthReady,
  login,
  logout,
  refreshProfile,
}