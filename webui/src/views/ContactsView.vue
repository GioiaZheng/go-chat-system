<!-- src/views/ContactsView.vue: Contact directory for launching private chats. -->
<template>
  <div class="page">
    <header class="topbar">
      <div class="title">Contacts</div>

      <div class="search">
        <input
          v-model.trim="q"
          class="input"
          type="text"
          placeholder="Search by name/username"
          @keyup.enter="onSearch"
        />
        <button class="btn" @click="onSearch">Search</button>
      </div>
    </header>

    <section class="content">
      <div v-if="loading" class="loading">
        <span class="spinner" aria-hidden="true"></span>
        Loading contacts…
      </div>

      <ErrorMsg v-else-if="err" :text="err" class="mb-2" />
      <button v-if="err" class="btn btn-outline mb-3" @click="load">Retry</button>

      <ul v-else class="list">
        <li v-for="u in users" :key="asId(u)" class="item">
          <div class="left">
            <span v-if="!avatar(u)" class="avatar-fallback">{{ initials(u) }}</span>
            <img v-else :src="avatar(u)" class="avatar" alt="avatar" />
            <div class="meta">
              <div class="name">{{ displayName(u) }}</div>
              <div class="sub">@{{ usernameOf(u) }}</div>
            </div>
          </div>
          <button class="btn" :disabled="creatingId===asId(u)" @click="message(u)">
            {{ creatingId===asId(u) ? 'Opening…' : 'Message' }}
          </button>
        </li>
        <li v-if="!users.length" class="empty">No users.</li>
      </ul>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import {
  isAuthed,
  listContacts,
  searchUsers,
  startPrivateConversation,
  getAvatarUrl,
} from '@/services/api'

const router = useRouter()

const q = ref('')
const loading = ref(false)
const err = ref('')
const users = ref([])
const creatingId = ref('') // Prevent duplicate conversation creation on rapid clicks

const asId        = (u) => String(u.id ?? u.user_id ?? u._id ?? '')
const displayName = (u) => String(u.name ?? u.username ?? '(user)')
const usernameOf  = (u) => String(u.username ?? u.name ?? '').toLowerCase()
const avatar      = (u) => getAvatarUrl({ avatarUri: u.avatarUri ?? u.avatar_uri ?? u.avatar_url ?? u.avatar })
const initials    = (u) => (displayName(u).match(/\b\w/g) || ['U']).slice(0,2).join('').toUpperCase()

async function load () {
  loading.value = true
  err.value = ''
  try {
    const list = await listContacts()  // Backend already excludes the current user
    users.value = Array.isArray(list) ? list : []
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load contacts'
  } finally {
    loading.value = false
  }
}

async function onSearch () {
  const keyword = q.value || ''
  if (!keyword) { load(); return }
  loading.value = true
  err.value = ''
  try {
    const arr = await searchUsers(keyword)
    users.value = Array.isArray(arr) ? arr : []
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Search failed'
  } finally {
    loading.value = false
  }
}

async function message (u) {
  err.value = ''
  const id = asId(u)
  if (!id) return
  creatingId.value = id
  try {
    const data = await startPrivateConversation(u)
    const cid =
      data?.conversation?.id ||
      data?.conversationId ||
      data?.conversation_id ||
      data?.id ||
      String(data)

    if (cid) router.push({ name: 'chat', params: { type: 'conv', id: cid } })
    else throw new Error('Invalid conversation response')
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to start chat'
  } finally {
    creatingId.value = ''
  }
}

// Restore the full list on empty input (with debouncing)
let debounceT = null
watch(q, (val) => {
  if (debounceT) clearTimeout(debounceT)
  debounceT = setTimeout(() => {
    if (val === '') load()
  }, 250)
})

onMounted(() => {
  if (!isAuthed()) { router.replace('/login'); return }
  load()
})
</script>

<style scoped>
.page{
  min-height:100%;
  height:100%;
  width:100%;
  min-width:0;
  flex:1 1 auto;
  display:flex;
  flex-direction:column;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
}
.topbar{
  height:56px; display:flex; align-items:center; justify-content:space-between;
  padding:0 18px; border-bottom:1px solid rgba(20,100,60,.08); background:#fff8; backdrop-filter: blur(6px);
}
.title{
  font-weight:800; color:#0f172a;
}

.search{ display:flex; gap:8px; }
.input{
  width:260px; border:1px solid #cbd5e1; border-radius:10px; padding:.5rem .75rem; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }
.btn{
  border:0; border-radius:10px; color:#fff; padding:.5rem .9rem; white-space:nowrap;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled { opacity:.6; cursor:not-allowed; }
.btn-outline{
  background:#fff; color:#334155; border:1px solid #cbd5e1; box-shadow:none;
}
.content{ flex:1 1 auto; max-width:1100px; margin:0 auto; padding:16px; width:100%; }

.loading{ color:#475569; display:flex; align-items:center; gap:.5rem }
.spinner{
  width:1rem; height:1rem; border:2px solid rgba(15,23,42,.25); border-top-color:transparent; border-radius:50%;
  display:inline-block; animation:spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.list{ list-style:none; padding:0; margin:0; display:grid; gap:10px }
.item{
  display:flex; align-items:center; gap:12px; justify-content:space-between;
  background:#fff; border:1px solid #e2e8f0; border-radius:12px; padding:10px 12px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
}
.left{ display:flex; align-items:center; gap:12px; }
.avatar{ width:36px; height:36px; border-radius:50%; object-fit:cover; border:1px solid #e2e8f0; }
.avatar-fallback{
  width:36px; height:36px; border-radius:50%; display:inline-flex; align-items:center; justify-content:center;
  background:#e0f7ee; color:#0f766e; font-weight:700; border:1px solid #a7f3d0;
}
.meta{ flex:1; min-width:0; }
.name{ font-weight:600; color:#0f172a }
.sub{ color:#64748b; font-size:.92rem }
.empty{ text-align:center; color:#64748b }
</style>
