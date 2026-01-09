<!-- User search helper: wraps the search API, displays results, and emits selections with basic error feedback. -->
<template>
  <div class="user-search">
    <div class="bar">
      <input
        v-model.trim="q"
        type="text"
        class="input"
        :placeholder="placeholder || 'Search users…'"
        @keyup.enter="onSearch"
      />
      <button class="btn" @click="onSearch" :disabled="loading">
        <span v-if="loading" class="spinner" aria-hidden="true"></span>
        <span>{{ loading ? 'Searching' : 'Search' }}</span>
      </button>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />

    <div v-if="loading" class="loading">
      <span class="spinner" aria-hidden="true"></span>
      Searching users…
    </div>

    <ul v-else class="list">
      <li
        v-for="u in users"
        :key="String(u.id || u.user_id || u._id)"
        class="item"
        @click="$emit('select', u)"
      >
        <span v-if="!avatar(u)" class="avatar-fallback avatar-circle">{{ initials(u) }}</span>
        <img v-else :src="avatar(u)" class="avatar avatar-circle" alt="avatar" />
        <div class="meta">
          <div class="name">{{ displayName(u) }}</div>
          <div v-if="handle(u)" class="sub">{{ handle(u) }}</div>
        </div>
      </li>
      <li v-if="!users.length && !err" class="empty">
        No matches yet. Try another name or username.
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ErrorMsg from '@/components/ErrorMsg.vue'
import { searchUsers, getAvatarUrl, preferredDisplayName, safeUsername, initialsFor } from '@/services/api'

const props = defineProps({
  placeholder: String,
})
const emit = defineEmits(['select', 'error'])

const q = ref('')
const users = ref([])
const loading = ref(false)
const err = ref('')

const avatar = (u) => getAvatarUrl(u)
const displayName = (u) => preferredDisplayName(u)
const handle = (u) => safeUsername(u)
const initials = (u) => initialsFor({ name: displayName(u) }, 'U')

async function onSearch() {
  if (!q.value) {
    users.value = []
    return
  }
  loading.value = true
  err.value = ''
  try {
    const res = await searchUsers(q.value)
    users.value = Array.isArray(res) ? res : res?.users || []
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Search failed'
    emit('error', err.value)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.user-search {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.bar {
  display: flex;
  gap: 8px;
}
.input {
  flex: 1;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  padding: 0.5rem 0.75rem;
}
.btn {
  border: 0;
  border-radius: 10px;
  color: #fff;
  padding: 0.5rem 0.9rem;
  background: #34d399;
  box-shadow: 0 0.5rem 1rem rgba(16, 185, 129, 0.2);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: transform 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}
.btn:hover {
  background: #10b981;
  box-shadow: 0 0.7rem 1.2rem rgba(16, 185, 129, 0.32);
  transform: translateY(-1px);
}
.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.btn:disabled:hover {
  background: #34d399;
  box-shadow: 0 0.5rem 1rem rgba(16, 185, 129, 0.2);
  transform: none;
}
.loading {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #475569;
  font-size: 0.9rem;
}
.spinner {
  width: 0.9rem;
  height: 0.9rem;
  border: 2px solid rgba(15, 23, 42, 0.25);
  border-top-color: transparent;
  border-radius: 50%;
  display: inline-block;
  animation: spin 0.7s linear infinite;
}
.list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 6px;
}
.item {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 8px 10px;
  cursor: pointer;
}
.item:hover {
  background: #f1f5f9;
}

.meta {
  flex: 1;
  min-width: 0;
}
.name {
  font-weight: 600;
  color: #0f172a;
}
.sub {
  color: #64748b;
  font-size: 0.9rem;
}
.empty {
  text-align: center;
  color: #64748b;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
