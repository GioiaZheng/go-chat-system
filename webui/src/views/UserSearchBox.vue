<template>
  <div>
    <!-- Search bar -->
    <div class="flex gap-2">
      <input
        v-model.trim="q"
        @keyup.enter="search"
        :disabled="loading"
        class="input flex-1"
        :placeholder="placeholder"
      />
      <button class="btn" @click="search" :disabled="loading">
        <span v-if="loading">Searching…</span>
        <span v-else>Search</span>
      </button>
    </div>

    <!-- Error (inline and also emitted upward) -->
    <p v-if="err" class="mt-2 text-sm text-red-600">{{ err }}</p>

    <!-- Results -->
    <ul v-if="items.length" class="mt-2 space-y-1">
      <li
        v-for="u in items"
        :key="u.id"
        class="p-2 bg-white border rounded flex justify-between items-center"
      >
        <div class="min-w-0">
          <div class="font-medium truncate">
            {{ u.username || u.email || u.name || ('User ' + (u.id || '').slice(0,8)) }}
          </div>
          <div class="text-xs text-gray-500 truncate">
            id: {{ u.id }}
          </div>
        </div>
        <button class="text-blue-600 hover:underline" @click="$emit('select', u)">
          Select
        </button>
      </li>
    </ul>

    <!-- Empty state -->
    <p v-else-if="queried && !loading" class="mt-2 text-sm text-gray-500">
      No users found.
    </p>
  </div>
</template>

<script setup>
// English-only UI & comments.
// This component uses axios directly and injects Authorization header per call.
// It unwraps both {code,data} and plain payloads, and emits 'error' upward on failures.

import { ref } from 'vue'
import axios from '../services/axios'

const props = defineProps({
  placeholder: { type: String, default: 'Search users…' }
})
defineEmits(['select', 'error'])

const q = ref('')
const items = ref([])
const loading = ref(false)
const err = ref('')
const queried = ref(false)

/** Attach Authorization header (token saved at login). */
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

async function search() {
  err.value = ''
  queried.value = true
  items.value = []

  const query = q.value
  if (!query) return

  loading.value = true
  try {
    // Backend route: GET /users/search?q=...
    const res = await axios.get('/users/search', {
      params: { q: query },
      headers: { ...getAuthHeaders() }
    })
    const payload = unwrap(res)

    // Accept multiple shapes: { items: [...] } or direct array
    items.value = payload?.items || (Array.isArray(payload) ? payload : [])
  } catch (e) {
    // Normalize error; also emit upward for parent UIs (e.g., toast)
    err.value =
      e?.response?.data?.message ||
      e?.message ||
      'Failed to search users'
    // Optional: special hint on 401
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
    }
    // Let parent optionally react (show toast / redirect)
    // Do not block UI if parent has no listener.
    // eslint-disable-next-line vue/require-explicit-emits
    // (we declared 'error' in defineEmits above)
    // @ts-ignore
    emit?.('error', err.value)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.input {
  width: 100%;
  border: 1px solid #ddd;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
  outline: none;
}
.input:focus {
  border-color: #a5b4fc;
  box-shadow: 0 0 0 3px rgba(99,102,241,.2);
}
.btn {
  background: #111827;
  color: #fff;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
}
.btn:disabled {
  opacity: .6;
  cursor: not-allowed;
}
</style>
