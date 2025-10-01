<template>
  <div class="max-w-3xl mx-auto p-4">
    <!-- Header -->
    <h2 class="text-xl font-semibold mb-4">Groups</h2>

    <!-- Error -->
    <ErrorMsg v-if="err" :text="err" class="mb-3" />

    <!-- Create group -->
    <div class="mb-4 p-3 bg-white rounded border">
      <h3 class="font-medium mb-2">Create Group</h3>
      <div class="flex flex-col gap-2 md:flex-row">
        <input
          v-model="groupName"
          class="input flex-1"
          placeholder="Group name"
        />
        <input
          v-model="memberIdsRaw"
          class="input flex-1"
          placeholder="Member IDs (comma separated)"
        />
        <button class="btn" @click="createGroup" :disabled="loading">
          Create
        </button>
      </div>
      <div class="text-xs text-gray-500 mt-1">
        Note: include yourself in the member list; otherwise the server may reject the request.
      </div>
    </div>

    <!-- List -->
    <LoadingSpinner v-if="loading" />
    <ul v-else class="space-y-2">
      <li
        v-for="g in groups"
        :key="g.id"
        class="p-3 bg-white rounded border"
      >
        <div class="flex items-center justify-between">
          <div class="min-w-0">
            <div class="font-medium truncate">
              {{ g.name || ('Group ' + (g.id || '').slice(0, 8)) }}
            </div>
            <div class="text-sm text-gray-500">
              conversation_id: {{ g.conversation_id || '(unknown)' }}
            </div>
          </div>

          <router-link
            v-if="g.conversation_id"
            class="text-blue-600 hover:underline"
            :to="{ name: 'chat', params: { type: 'conv', id: g.conversation_id } }"
            title="Open group chat"
          >
            Open
          </router-link>
        </div>
      </li>

      <li v-if="!groups.length" class="p-6 text-center text-gray-500 bg-white rounded border">
        No groups yet. Create one above.
      </li>
    </ul>
  </div>
</template>

<script setup>
// English-only. Uses axios directly and injects Authorization header per call.
// Matches your backend routing: GET /groups, POST /groups (name + members/member_ids).
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../services/axios'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ErrorMsg from '../components/ErrorMsg.vue'

const router = useRouter()

const groups = ref([])
const loading = ref(false)
const err = ref('')

const groupName = ref('')
const memberIdsRaw = ref('')

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

onMounted(load)

async function load() {
  loading.value = true
  err.value = ''
  groups.value = []
  try {
    // Backend canonical list
    const res = await axios.get('/groups', {
      headers: { ...getAuthHeaders() }
    })
    const payload = unwrap(res)

    // Accept multiple shapes
    groups.value =
      payload?.items ||
      payload?.groups ||
      (Array.isArray(payload) ? payload : []) ||
      []
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to load groups'
    }
  } finally {
    loading.value = false
  }
}

async function createGroup() {
  err.value = ''

  // Parse comma separated IDs → ['u1','u2',...]
  const members = memberIdsRaw.value
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)

  if (!groupName.value || members.length === 0) {
    err.value = 'Group name and member list are required.'
    return
  }

  loading.value = true
  try {
    // Be tolerant to backend naming: send both "members" and "member_ids"
    await axios.post(
      '/groups',
      { name: groupName.value, members, member_ids: members },
      { headers: { ...getAuthHeaders() } }
    )

    await load()
    groupName.value = ''
    memberIdsRaw.value = ''
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to create group'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.input {
  width: 100%;
  border: 1px solid #ddd;
  padding: .5rem .75rem;
  border-radius: .375rem;
  outline: none;
}
.input:focus {
  border-color: #a5b4fc;
  box-shadow: 0 0 0 3px rgba(99,102,241,.2);
}
.btn {
  background: #111827;
  color: #fff;
  padding: .5rem .75rem;
  border-radius: .375rem;
}
.btn:disabled {
  opacity: .6;
  cursor: not-allowed;
}
</style>
