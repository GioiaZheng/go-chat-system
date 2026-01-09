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
          <button class="btn btn-secondary" @click="onSearch">Search</button>
        </div>

        <div class="contacts-layout">
          <div>
            <div v-if="loading" class="loading">
              <span class="spinner" aria-hidden="true"></span>
              Loading contacts…
            </div>

            <ErrorMsg v-else-if="err" :text="err" class="mb-2" />
            <button v-if="err" class="btn btn-outline mb-3" @click="load">Retry</button>

            <ul v-else class="list">
              <li v-for="u in users" :key="asId(u)" class="item">
                <button
                  class="item-button"
                  type="button"
                  :disabled="creatingId === asId(u)"
                  @click="openChat(u)"
                >
                  <div class="left">
                    <span v-if="!avatar(u)" class="avatar-fallback avatar-circle">{{ initials(u) }}</span>
                    <img v-else :src="avatar(u)" class="avatar avatar-circle" alt="avatar" />
                    <div class="meta">
                      <div class="name">{{ displayName(u) }}</div>
                      <div class="preview">
                        {{ lastPreviewFor(u) }}
                      </div>
                    </div>
                  </div>
                  <div class="meta-right">
                    <span class="time">{{ lastTimeFor(u) }}</span>
                  </div>
                </button>
              </li>
              <li v-if="!users.length" class="empty">No users.</li>
            </ul>
          </div>

          <aside class="empty-state" aria-live="polite">
            <div class="empty-title">Select a contact to start chatting</div>
            <div class="empty-subtitle">
              Choose someone from the list to open a conversation.
            </div>
          </aside>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import {
  listContacts,
  searchUsers,
  startPrivateConversation,
  getAvatarUrl,
  preferredDisplayName,
  initialsFor,
  getMyConversations,
  normalizeUser,
} from '@/services/api'
import { ensureAuthReady, isAuthenticated, currentUser } from '@/services/auth'

const router = useRouter()

const q = ref('')
const loading = ref(false)
const err = ref('')
const users = ref([])
const creatingId = ref('') // Prevent duplicate conversation creation on rapid clicks
const conversationIndex = ref({})

const PREVIEW_LIMIT = 40

const asId          = (u) => String(u.id ?? u.user_id ?? u._id ?? '')
const displayName   = (u) => preferredDisplayName(u)
const avatar        = (u) => getAvatarUrl({ avatarUri: u.avatarUri ?? u.avatar_uri ?? u.avatar_url ?? u.avatar })
const initials      = (u) => initialsFor({ name: displayName(u) }, 'U')
const me = computed(() => currentUser.value)
const myId = () => String(me.value?.id ?? '')

function cleanPreviewText(value = '') {
  return String(value || '').trim()
}

function formatPreviewText(value = '') {
  const text = cleanPreviewText(value)
  if (!text) return ''
  return text.split('\n')[0].slice(0, PREVIEW_LIMIT).trim()
}

function previewForMessage(raw = {}) {
  if (!raw) return ''
  const type = String(raw?.type || raw?.Type || '').toLowerCase()
  const fileUrl = raw?.fileUrl ?? raw?.file_url ?? raw?.FileUrl
  const content = raw?.content ?? raw?.Content ?? ''
  const text = formatPreviewText(content)
  const isImage = type === 'image'
  const isFile = type === 'file' || (!isImage && !!fileUrl)
  const mediaLabel = isImage ? '[image]' : (isFile ? '[file]' : '')

  if (mediaLabel && text) return `${mediaLabel} ${text}`
  if (mediaLabel) return mediaLabel
  return text
}

function messageTime(raw = {}) {
  return (
    raw?.createdAt ||
    raw?.created_at ||
    raw?.CreatedAt ||
    raw?.timestamp ||
    raw?.Timestamp ||
    null
  )
}

function messageTimeValue(raw = {}) {
  const value = messageTime(raw)
  if (!value) return null
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return null
  return d.getTime()
}

function pickLastMessage(c) {
  if (!c) return null

  const fromKnownFields = c.lastMessage || c.last_message
  if (fromKnownFields) return fromKnownFields

  const list =
    c.messages ||
    c.Messages ||
    c.messageList ||
    c.message_list ||
    c.items ||
    c.list

  if (!Array.isArray(list) || !list.length) return null

  let best = null
  let bestTime = null
  list.forEach((msg) => {
    const ts = messageTimeValue(msg)
    if (ts !== null) {
      if (bestTime === null || ts > bestTime) {
        bestTime = ts
        best = msg
      }
      return
    }
    if (!best) {
      best = msg
    }
  })

  return best || list[list.length - 1]
}

function formatTimestamp(value) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const now = new Date()
  const isToday =
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}/${day}`
}

function buildConversationIndex(items = []) {
  const index = {}
  const meId = myId()

  ;(items || []).forEach((c) => {
    const isGroupType = c?.type === 'group' || !!(c?.groupId || c?.group_id || c?.group?.id)
    if (isGroupType) return

    const participants =
      c.participants ||
      c.members ||
      c.memberList ||
      c.users ||
      []
    const normalized = Array.isArray(participants) ? participants.map(normalizeUser) : []
    const other = normalized.find(u => String(u.id) !== meId)
    if (!other?.id) return

    const last = pickLastMessage(c)
    const preview = last ? previewForMessage(last) : ''
    const lastTime = last ? messageTime(last) : ''
    const conversationId = c?.id || c?.conversationId || c?.conversation_id

    index[String(other.id)] = {
      conversationId,
      lastPreview: preview || 'No messages yet',
      lastTime: lastTime || '',
    }
  })

  conversationIndex.value = index
}

async function load () {
  loading.value = true
  err.value = ''
  try {
    const [list, convs] = await Promise.all([
      listContacts(), // Backend already excludes the current user
      getMyConversations(),
    ])
    users.value = Array.isArray(list) ? list : []
    const items = convs?.data?.items || convs?.items || (Array.isArray(convs) ? convs : []) || []
    buildConversationIndex(items)
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
    const [arr, convs] = await Promise.all([
      searchUsers(keyword),
      getMyConversations(),
    ])
    users.value = Array.isArray(arr) ? arr : []
    const items = convs?.data?.items || convs?.items || (Array.isArray(convs) ? convs : []) || []
    buildConversationIndex(items)
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Search failed'
  } finally {
    loading.value = false
  }
}

function conversationMetaFor(u) {
  return conversationIndex.value[asId(u)] || null
}

function lastPreviewFor(u) {
  return conversationMetaFor(u)?.lastPreview || 'No messages yet'
}

function lastTimeFor(u) {
  const meta = conversationMetaFor(u)
  return formatTimestamp(meta?.lastTime)
}

async function openChat (u) {
  err.value = ''
  const id = asId(u)
  if (!id) return
  const existing = conversationMetaFor(u)
  if (existing?.conversationId) {
    router.push({ name: 'chat', params: { type: 'conv', id: existing.conversationId } })
    return
  }
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
  ensureAuthReady().then(() => {
    if (!isAuthenticated.value) {
      router.replace('/login')
      return
    }
    load()
  })
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
  width:min(100%, 1080px);
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
.btn-secondary{
  background:#475569;
  box-shadow:none;
}
.btn-secondary:hover{
  background:#334155;
  box-shadow:none;
}
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

.contacts-layout{
  display:grid;
  grid-template-columns:minmax(280px, 1fr) minmax(260px, 360px);
  gap:20px;
  align-items:start;
}
.list{ list-style:none; padding:0; margin:0; display:grid; gap:12px }
.item{
  background:#fff;
  border:1px solid #e2e8f0;
  border-radius:var(--radius-control);
  box-shadow:0 4px 12px rgba(15,23,42,.06);
  overflow:hidden;
}
.item-button{
  width:100%;
  border:0;
  background:transparent;
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap:14px;
  padding:14px 16px;
  text-align:left;
  cursor:pointer;
  transition:background .15s ease, box-shadow .15s ease, border-color .15s ease;
}
.item-button:hover{
  background:#f8fafc;
}
.item-button:disabled{
  cursor:not-allowed;
  opacity:.7;
}
.left{ display:flex; align-items:center; gap:12px; min-width:0; flex:1; }
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
.preview{ color:#64748b; font-size:.92rem; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.meta-right{
  color:#94a3b8;
  font-size:.85rem;
  white-space:nowrap;
  flex:0 0 auto;
}
.empty{ text-align:center; color:#64748b }
.empty-state{
  border:1px dashed #d8e1ee;
  border-radius:16px;
  padding:20px;
  background:#f8fafc;
  color:#475569;
}
.empty-title{
  font-weight:600;
  margin-bottom:6px;
  color:#0f172a;
}
.empty-subtitle{
  font-size:.92rem;
  color:#64748b;
}
@media (max-width: 900px){
  .contacts-layout{
    grid-template-columns:1fr;
  }
  .empty-state{
    order:-1;
  }
}
</style>
