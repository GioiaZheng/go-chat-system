<template>
  <div class="max-w-xl mx-auto p-4">
    <h2 class="text-xl font-semibold mb-4">My Profile</h2>

    <!-- Error -->
    <ErrorMsg v-if="err" :text="err" class="mb-3" />

    <!-- Profile card -->
    <div v-if="me" class="space-y-3 bg-white p-4 rounded border">
      <div class="flex items-center gap-3">
        <img
          v-if="me.photo_url"
          :src="me.photo_url"
          alt="avatar"
          class="w-12 h-12 rounded object-cover border"
        />
        <div class="text-sm text-gray-500">Logged in as:</div>
      </div>

      <div><span class="font-medium">id:</span> {{ me.id }}</div>
      <div><span class="font-medium">username:</span> {{ me.username }}</div>
      <div><span class="font-medium">name:</span> {{ me.name }}</div>
      <div><span class="font-medium">email:</span> {{ me.email }}</div>
      <div><span class="font-medium">gender:</span> {{ me.gender }}</div>

      <!-- Change username -->
      <div class="mt-2">
        <label class="block mb-1 text-sm">Change username</label>
        <input v-model="newName" class="input w-full" placeholder="new username" />
        <button class="btn mt-2" @click="setUsername" :disabled="loading || !newName">
          Save
        </button>
      </div>

      <!-- Set photo (multipart upload) -->
      <div class="mt-2">
        <label class="block mb-1 text-sm">Set photo</label>
        <input type="file" @change="onFile" />
        <button class="btn mt-2" @click="setPhoto" :disabled="loading || !file">
          Upload
        </button>
        <p class="text-xs text-gray-500 mt-1">
          Tip: you can also use preset mode via backend (e.g., <code>?preset=avatar7</code>) if enabled.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
// English-only UI and comments.
// This component uses axios directly, without changing axios.js.
// It injects Authorization header per request and unwraps both {code,data} and plain payloads.

import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../services/axios'
import ErrorMsg from '../components/ErrorMsg.vue'

const router = useRouter()

const me = ref(null)
const newName = ref('')
const file = ref(null)
const loading = ref(false)
const err = ref('')

/** Attach Authorization header (token was stored at login). */
function getAuthHeaders() {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

/** Unwrap both { code, data } and plain payloads. */
function unwrap(res) {
  const d = res?.data
  if (d && typeof d === 'object' && 'data' in d) return d.data
  return d
}

/** Load current profile and cache it in localStorage('me'). */
async function loadProfile() {
  try {
    const res = await axios.get('/users/me', { headers: { ...getAuthHeaders() } })
    const payload = unwrap(res)
    // Accept shapes: { user: {...} } or direct object
    me.value = payload?.user || payload || null
    if (me.value) localStorage.setItem('me', JSON.stringify(me.value))
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to load profile'
    }
  }
}

onMounted(loadProfile)

function onFile(e) {
  file.value = e.target.files?.[0] || null
}

/** PUT /users/set_username { username } */
async function setUsername() {
  if (!newName.value) return
  loading.value = true
  err.value = ''
  try {
    await axios.put(
      '/users/set_username',
      { username: newName.value },
      { headers: { ...getAuthHeaders() } }
    )
    await loadProfile()
    newName.value = ''
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to set username'
    }
  } finally {
    loading.value = false
  }
}

/** PUT /users/set_photo (multipart, field "upload") */
async function setPhoto() {
  if (!file.value) return
  loading.value = true
  err.value = ''
  try {
    const form = new FormData()
    form.append('upload', file.value, file.value.name)

    await axios.put('/users/set_photo', form, {
      headers: { ...getAuthHeaders(), 'Content-Type': 'multipart/form-data' }
    })

    await loadProfile()
    file.value = null
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to set photo'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.input { border:1px solid #ddd; padding:.5rem .75rem; border-radius:.375rem; outline: none; }
.input:focus { border-color:#a5b4fc; box-shadow:0 0 0 3px rgba(99,102,241,.2); }
.btn { background:#111827; color:#fff; padding:.5rem .75rem; border-radius:.375rem; }
.btn:disabled { opacity:.6; cursor:not-allowed; }
</style>
