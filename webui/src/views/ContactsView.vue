<!-- src/views/ContactsView.vue: Contact directory for launching private chats. -->
<template>
  <div class="page">
    <header class="topbar">
      <div class="title">Contacts</div>
    </header>

    <section class="content">
      <div class="content-inner">
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

      <div v-if="loading" class="loading">
        <span class="spinner" aria-hidden="true"></span>
        Loading contacts…
      </div>

      <ErrorMsg v-else-if="err" :text="err" class="mb-2" />
      <button v-if="err" class="btn btn-outline mb-3" @click="load">Retry</button>

      <ul v-else class="list">
        <li v-for="u in users" :key="asId(u)" class="item">
          <div class="left">
            <span v-if="!avatar(u)" class="avatar-fallback avatar-circle">{{ initials(u) }}</span>
            <img v-else :src="avatar(u)" class="avatar avatar-circle" alt="avatar" />
            <div class="meta">
              <div class="name">{{ displayName(u) }}</div>
              <div v-if="usernameLabel(u)" class="sub">{{ usernameLabel(u) }}</div>
            </div>
          </div>
          <button class="btn" :disabled="creatingId===asId(u)" @click="message(u)">
            {{ creatingId===asId(u) ? 'Opening…' : 'Message' }}
          </button>
        </li>
        <li v-if="!users.length" class="empty">No users.</li>
      </ul>
      </div>
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
  preferredDisplayName,
  safeUsername,
  initialsFor,
} from '@/services/api'

const router = useRouter()

const q = ref('')
const loading = ref(false)
const err = ref('')
const users = ref([])
const creatingId = ref('') // Prevent duplicate conversation creation on rapid clicks

const asId          = (u) => String(u.id ?? u.user_id ?? u._id ?? '')
const displayName   = (u) => preferredDisplayName(u)
const usernameLabel = (u) => safeUsername(u)
const avatar        = (u) => getAvatarUrl({ avatarUri: u.avatarUri ?? u.avatar_uri ?? u.avatar_url ?? u.avatar })
const initials      = (u) => initialsFor({ name: displayName(u) }, 'U')

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

    if (cid) {
      window.dispatchEvent(new CustomEvent('conversations:reload'))
      router.push({ name: 'chat', params: { type: 'conv', id: cid } })
    }
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

.content{
  flex:1 1 auto;
  margin:0;
  padding:16px 20px;
  width:100%;
  display:flex;
  justify-content:flex-start;
}

.content-inner{
  width:min(100%, 820px);
  display:flex;
  flex-direction:column;
  gap:14px;
}

.search{ display:flex; gap:8px; width:100%; }
.input{
  flex:1; border:1px solid #cbd5e1; border-radius:10px; padding:.5rem .75rem; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }
.btn{
  border:0; border-radius:10px; color:#fff; padding:.5rem .9rem; white-space:nowrap;
  background:#34d399;
  box-shadow:0 .5rem 1rem rgba(16,185,129,.2);
  transition:transform .15s ease, box-shadow .15s ease, background .15s ease;
}
.btn:hover{
  background:#10b981;
  box-shadow:0 .7rem 1.2rem rgba(16,185,129,.32);
  transform:translateY(-1px);
}
.btn:disabled { opacity:.6; cursor:not-allowed; }
.btn:disabled:hover { background:#34d399; box-shadow:0 .5rem 1rem rgba(16,185,129,.2); transform:none; }
.btn-outline{
  background:#fff; color:#334155; border:1px solid #cbd5e1; box-shadow:none;
}
.btn.icon-only{
  width:36px; height:36px; padding:0;
  display:inline-flex; align-items:center; justify-content:center;
}
.btn.icon-only .btn__label{ display:none; }

.loading{ color:#475569; display:flex; align-items:center; gap:.5rem }
.spinner{
  width:1rem; height:1rem; border:2px solid rgba(15,23,42,.25); border-top-color:transparent; border-radius:50%;
  display:inline-block; animation:spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.list{ list-style:none; padding:0; margin:0; display:grid; gap:12px }
.item{
  display:grid; align-items:center; gap:14px;
  grid-template-columns:auto 1fr auto;
  background:#fff; border:1px solid #e2e8f0; border-radius:var(--radius-control);
  padding:14px 16px;
  box-shadow:0 4px 12px rgba(15,23,42,.06);
  transition:background .15s ease, box-shadow .15s ease, border-color .15s ease;
}
.item:hover{
  background:#f8fafc;
  border-color:#d7dde5;
  box-shadow:0 6px 16px rgba(15,23,42,.08);
}
.left{ display:flex; align-items:center; gap:12px; }
.avatar,
.avatar-fallback{
  width:40px;
  height:40px;
  border-radius:50%;
  flex:0 0 40px;
}
.avatar{
  object-fit:cover;
  border:1px solid #e2e8f0;
}
.avatar-fallback{
  display:inline-flex;
  align-items:center;
  justify-content:center;
  background:#e0f7ee;
  color:#0f766e;
  font-weight:700;
}
.meta{ flex:1; min-width:0; }
.name{ font-weight:600; color:#0f172a }
.sub{ color:#64748b; font-size:.92rem }
.empty{ text-align:center; color:#64748b }
</style>
