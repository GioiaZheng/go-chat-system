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
        {{ loading ? '...' : 'Search' }}
      </button>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />

    <ul v-if="!loading" class="list">
      <li
        v-for="u in users"
        :key="String(u.id || u.user_id || u._id)"
        class="item"
        @click="$emit('select', u)"
      >
        <span v-if="!avatar(u)" class="avatar-fallback">{{ initials(u) }}</span>
        <img v-else :src="avatar(u)" class="avatar" alt="avatar" />
        <div class="meta">
          <div class="name">{{ u.name || u.username || '(user)' }}</div>
          <div class="sub">@{{ u.username || '-' }}</div>
        </div>
      </li>
      <li v-if="!users.length && !err" class="empty">No results.</li>
    </ul>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import ErrorMsg from '@/components/ErrorMsg.vue'
import { searchUsers, getAvatarUrl } from '@/services/api'

const props = defineProps({
  placeholder: String,
})
const emit = defineEmits(['select', 'error'])

const q = ref('')
const users = ref([])
const loading = ref(false)
const err = ref('')

const avatar = (u) => getAvatarUrl(u)
const initials = (u) => {
  const name = u?.name || u?.username || 'U'
  const match = String(name).match(/\b\w/g) || ['U']
  return match.slice(0, 2).join('').toUpperCase()
}

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
  background-image: linear-gradient(135deg, #22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow: 0 0.3rem 0.8rem rgba(34, 197, 94, 0.25);
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
.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #e2e8f0;
}
.avatar-fallback {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #e0f7ee;
  color: #0f766e;
  font-weight: 700;
  border: 1px solid #a7f3d0;
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
</style>
