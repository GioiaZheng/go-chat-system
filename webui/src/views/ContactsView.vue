<!-- src/views/ContactsView.vue: Contact directory for launching private chats. -->
<template>
  <div class="page">
    <section class="workspace">
      <div class="content-inner">
        <div class="panel">
          <div class="list-pane">
            <div class="list-search">
              <input
                v-model.trim="q"
                class="search"
                type="text"
                placeholder="Search by name/username"
                @keyup.enter="onSearch"
              />
              <button class="btn btn-search" @click="onSearch">Search</button>
            </div>

            <div v-if="loading" class="loading">
              <span class="spinner" aria-hidden="true"></span>
              Loading contacts…
            </div>

            <ErrorMsg v-else-if="err" :text="err" class="mb-2" />
            <button v-if="err" class="btn btn-outline mb-3" @click="load">Retry</button>

            <ul v-else class="list" role="list">
              <li
                v-for="u in sortedUsers"
                :key="asId(u)"
                class="item"
                :class="{ active: activeId === asId(u), disabled: creatingId === asId(u) }"
                role="button"
                tabindex="0"
                @click="handleSelect(u)"
                @keydown.enter.prevent="handleSelect(u)"
                @keydown.space.prevent="handleSelect(u)"
              >
                <div class="left">
                  <span v-if="!avatar(u)" class="avatar-fallback avatar-circle">{{ initials(u) }}</span>
                  <img v-else :src="avatar(u)" class="avatar avatar-circle" alt="avatar" />
                </div>
                <div class="info">
                  <div class="top">
                    <div class="name">{{ displayName(u) }}</div>
                    <div class="time">{{ lastTimeFor(u) }}</div>
                  </div>
                  <div class="bottom">
                    <div class="preview">
                      {{ lastPreviewFor(u) }}
                    </div>
                  </div>
                </div>
              </li>
              <li v-if="!sortedUsers.length" class="empty">
                <span class="empty-icon" aria-hidden="true">👥</span>
                <span>No contacts yet — try a search 👋</span>
              </li>
            </ul>
          </div>
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
const activeId = ref('')
const conversationIndex = ref({})

const PREVIEW_LIMIT = 20

const asId          = (u) => String(u.id ?? u.user_id ?? u._id ?? '')
const displayName   = (u) => preferredDisplayName(u)
const avatar        = (u) => getAvatarUrl({ avatarUri: u.avatarUri ?? u.avatar_uri ?? u.avatar_url ?? u.avatar })
const initials      = (u) => initialsFor({ name: displayName(u) }, 'U')
const me = computed(() => currentUser.value)
const myId = () => String(me.value?.id ?? '')
const sortedUsers = computed(() => {
  const items = Array.isArray(users.value) ? [...users.value] : []
  return items.sort((a, b) => {
    const metaA = conversationMetaFor(a)
    const metaB = conversationMetaFor(b)
    const timeA = metaA?.lastTime ? new Date(metaA.lastTime).getTime() : 0
    const timeB = metaB?.lastTime ? new Date(metaB.lastTime).getTime() : 0

    if (timeA && !timeB) return -1
    if (!timeA && timeB) return 1
    if (timeA && timeB) return timeB - timeA
    return displayName(a).localeCompare(displayName(b))
  })
})

function cleanPreviewText(value = '') {
  return String(value || '')
    .replace(/\[forwarded message\]|\[forwarded\]|forwarded message:?/gi, '')
    .trim()
}

function formatPreviewText(value = '') {
  const text = cleanPreviewText(value)
  if (!text) return ''
  return text.split('\n')[0].slice(0, PREVIEW_LIMIT).trim()
}

function previewForMessage(raw = {}) {
  if (raw?.deleted || raw?.isDeleted) return '[Deleted]'

  const type = String(raw?.type || raw?.Type || '').toLowerCase()
  const fileUrl = raw?.fileUrl ?? raw?.file_url ?? raw?.FileUrl
  const content = raw?.content ?? raw?.Content ?? ''
  const text = formatPreviewText(content)
  const isFile = type === 'file'
  const isImage = type === 'image'
  const mediaLabel = isFile ? '[file]' : (isImage || fileUrl ? '[image]' : '')

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
  const preview = String(conversationMetaFor(u)?.lastPreview || '').trim()
  if (!preview || /no messages yet/i.test(preview)) return '— no messages yet —'
  return preview
}

function lastTimeFor(u) {
  const meta = conversationMetaFor(u)
  return formatTimestamp(meta?.lastTime)
}

function handleSelect(u) {
  if (creatingId.value === asId(u)) return
  openChat(u)
}

async function openChat (u) {
  err.value = ''
  const id = asId(u)
  if (!id) return
  activeId.value = id
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
  min-height: 100vh;
  height: 100vh;
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  background: #e5e7eb;
  color: #1f2937;
}
.workspace{
  flex:1 1 auto;
  margin:0;
  padding:16px;
  width:100%;
  display:flex;
  flex-direction:column;
  min-height:0;
}

.content-inner{
  width:100%;
  display:flex;
  flex-direction:column;
  gap:12px;
  flex:1 1 auto;
  min-height:0;
}

.list-search{
  display:flex;
  align-items:center;
  gap:8px;
  padding: 0 0 8px;
}
.search{
  flex:1;
  border:1px solid #e5e7eb;
  border-radius:var(--radius-control);
  padding:10px 12px;
  outline:none;
  background:#fff;
  font-size:var(--font-primary);
  transition:border-color .15s ease, box-shadow .15s ease;
}
.search:focus{
  border-color:#32d583;
  box-shadow:0 0 0 3px rgba(50, 213, 131, 0.2);
}
.btn{
  border:0;
  border-radius:var(--radius-control);
  color:#fff;
  padding:0.45rem 0.75rem;
  white-space:nowrap;
  background:#32d583;
  box-shadow:0 .35rem .8rem rgba(50, 213, 131, 0.22);
  transition:transform .15s ease, box-shadow .15s ease, background .15s ease;
  font-weight:700;
}
.btn:hover{
  background:#c6f6d5;
  color:#065f46;
  box-shadow:0 .5rem 1rem rgba(15, 118, 110, 0.18);
  transform:translateY(-1px);
}
.btn:disabled { opacity:.6; cursor:not-allowed; }
.btn:disabled:hover { background:#32d583; color:#fff; box-shadow:0 .35rem .8rem rgba(50, 213, 131, 0.22); transform:none; }
.btn-search{
  background:#f1f5f9;
  color:#0f172a;
  border:1px solid #e2e8f0;
  box-shadow:none;
  padding:8px 12px;
}
.btn-search:hover{
  background:#c6f6d5;
  border-color:#32d583;
  color:#065f46;
  box-shadow:0 8px 16px rgba(15,23,42,0.06);
}
.btn-outline{
  background:#e5e7eb; color:#1f2937; border:1px solid #e5e7eb; box-shadow:none;
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

.panel{
  flex:1 1 auto;
  min-height:0;
  display:flex;
  flex-direction:row;
  background:var(--panel);
  border:1px solid var(--border);
  overflow:hidden;
}
.list-pane{
  flex:1 1 auto;
  background:var(--panel);
  border-right:0;
  display:flex;
  flex-direction:column;
  padding:16px;
  gap:10px;
  min-height:0;
  overflow:hidden;
}
.list{
  list-style:none;
  padding:0;
  margin:0;
  display:flex;
  flex-direction:column;
  gap:10px;
  min-height:0;
  flex:1 1 auto;
  overflow-y:auto;
}
.item{
  background:#ffffff;
  border:1px solid var(--border);
  border-radius:var(--radius-control);
  padding:14px 16px;
  display:grid;
  grid-template-columns:auto 1fr auto;
  align-items:center;
  gap:14px;
  min-height:64px;
  overflow:hidden;
  transition:background .15s ease, box-shadow .15s ease, border-color .15s ease;
  position:relative;
  cursor:pointer;
}
.item:hover{
  background:#c6f6d5;
  border-color:rgba(50, 213, 131, 0.6);
  box-shadow:0 6px 16px rgba(15,23,42,0.08);
}
.item.active{
  background:#32d583;
  border-color:#32d583;
}
.item.active::before{
  content:'';
  position:absolute;
  left:0;
  top:10px;
  bottom:10px;
  width:3px;
  background:#16a34a;
}
.item.active .name{
  color:#0f172a;
}
.item.active .time,
.item.active .preview{
  color:#065f46;
}
.item.disabled{
  cursor:not-allowed;
  opacity:.7;
}
.left{ display:flex; align-items:center; justify-content:center; }
.avatar,
.avatar-fallback{
  width:32px;
  height:32px;
  border-radius:50%;
  flex:0 0 32px;
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
.info{ flex:1; min-width:0; display:flex; flex-direction:column; gap:2px; }
.top{
  display:flex;
  align-items:center;
  gap:12px;
}
.name{ font-weight:600; color:var(--text); font-size:var(--font-primary); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; min-width:0; }
.time{
  color:var(--muted);
  font-size:.85rem;
  white-space:nowrap;
  margin-left:auto;
}
.preview{ color:var(--muted); font-size:.92rem; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.empty{
  text-align:center;
  color:#64748b;
  display:flex;
  align-items:center;
  justify-content:center;
  gap:8px;
  font-weight:600;
}
.empty-icon{
  font-size:1rem;
}
@media (max-width: 900px){
  .panel{
    flex-direction:column;
  }
  .list-pane{
    flex:0 0 auto;
    border-right:0;
    border-bottom:0;
  }
}
</style>
