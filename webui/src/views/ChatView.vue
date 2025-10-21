<template>
  <div class="layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="brand">WASAText</div>

      <nav class="nav">
        <div class="sec">CONVERSATIONS</div>
        <RouterLink class="link" to="/conversations">All Conversations</RouterLink>
        <RouterLink class="link" to="/groups">Groups</RouterLink>

        <div class="sec mt">SETTINGS</div>
        <RouterLink class="link" to="/profile">Profile</RouterLink>
        <button class="link link-btn" @click="logout">Logout</button>
      </nav>
    </aside>

    <!-- Main -->
    <main class="main">
      <header class="topbar">
        <div class="title">
          <strong>Chat</strong>
          <small v-if="route.params.type && route.params.id" class="muted">
            ({{ route.params.type }} / {{ route.params.id }})
          </small>
        </div>
        <div class="who">Signed in as {{ meName }}</div>
      </header>

      <section class="content">
        <ErrorMsg v-if="err" :text="err" class="mb-2" />

        <div ref="scrollbox" class="scroll">
          <LoadingSpinner v-if="loading" />
          <div
            v-for="m in messages"
            :key="m._local_id || m.id"
            class="msg"
            :class="{ mine: isMine(m) }"
          >
            <div class="bubble" :class="{ mine: isMine(m) }">
              {{ m.content }}
            </div>
            <div class="meta">
              {{ shortId(m.sender_id) }} · {{ fmtTime(m.created_at || m.createdAt) }}
              <span v-if="m._status==='sending'"> · sending…</span>
              <span v-if="m._status==='error'" class="err"> · failed</span>
            </div>
          </div>
        </div>

        <div class="composer">
          <textarea
            v-model="draft"
            class="input"
            placeholder="Type a message…"
            rows="1"
            @keyup.enter.exact.prevent="onSend"
          ></textarea>
          <button class="btn" :disabled="!draft.trim() || sending" @click="onSend">
            {{ sending ? 'Sending…' : 'Send' }}
          </button>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import api, { getMessages, sendMessage, startConversation } from '@/services/api'

const route = useRoute()
const router = useRouter()

/* ---------- auth & me ---------- */
const meName = ref(localStorage.getItem('name') || 'user')
const meId = ref('')
try { meId.value = JSON.parse(localStorage.getItem('me') || '{}')?.id || '' } catch {}

/* ---------- state ---------- */
const convId = ref('')                 // 解析后的会话ID
const messages = ref([])
const loading = ref(false)
const sending = ref(false)
const err = ref('')
const draft = ref('')
const scrollbox = ref(null)
let pollTimer = null

/* ---------- helpers ---------- */
const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))
const isMine = (m) => (m?.sender_id || m?.from_user_id || '') === meId.value
const shortId = (s='') => String(s).slice(0, 8)
const fmtTime = (t) => String(t || '').toString().slice(0, 19)

/** 保证拿到会话ID
 *  - /chat/conv/:id      => 直接用
 *  - /chat/private/:uid  => startConversation({ user_id: uid })
 *  - /chat/group/:gid    => startConversation({ group_id: gid })
 */
async function resolveConversationId() {
  const type = route.params.type
  const id = String(route.params.id || '')
  err.value = ''

  if (type === 'conv') { convId.value = id; return }

  try {
    if (type === 'private') {
      const data = await startConversation({ user_id: id })
      convId.value = data?.conversationId ?? data?.id ?? data?.conversation_id ?? String(data)
      return
    }
    if (type === 'group') {
      const data = await startConversation({ group_id: id })
      convId.value = data?.conversationId ?? data?.id ?? data?.conversation_id ?? String(data)
      return
    }
    throw new Error('Unknown chat type')
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to resolve conversation'
  }
}

async function loadMessages() {
  if (!convId.value) return
  loading.value = true; err.value = ''
  try {
    const data = await getMessages({ conversationId: convId.value, limit: 50 })
    // 兼容 {items} 或 {messages} 或 直接数组
    const arr = data?.items || data?.messages || (Array.isArray(data) ? data : [])
    messages.value = Array.isArray(arr) ? arr : []
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load messages'
  } finally {
    loading.value = false
    await nextTick()
    if (scrollbox.value) scrollbox.value.scrollTop = scrollbox.value.scrollHeight
  }
}

async function refreshLoop() {
  clearInterval(pollTimer)
  pollTimer = setInterval(loadMessages, 5000)
}

async function bootstrap() {
  if (!authed()) { router.replace('/login'); return }
  await resolveConversationId()
  await loadMessages()
  refreshLoop()
}

onMounted(bootstrap)
onUnmounted(() => { clearInterval(pollTimer) })
watch(() => [route.params.type, route.params.id], bootstrap)

/* ---------- send ---------- */
async function onSend() {
  const text = draft.value.trim()
  if (!text || !convId.value) return

  // 乐观插入
  const temp = {
    _local_id: 'tmp_' + Math.random().toString(36).slice(2),
    sender_id: meId.value,
    content: text,
    created_at: new Date().toISOString(),
    _status: 'sending'
  }
  messages.value.push(temp)
  draft.value = ''
  await nextTick()
  if (scrollbox.value) scrollbox.value.scrollTop = scrollbox.value.scrollHeight

  sending.value = true
  try {
    await sendMessage({ conversationId: convId.value, content: text, type: 'text' })
    // 重新拉一次，替换临时消息
    await loadMessages()
  } catch (e) {
    // 标记失败
    const idx = messages.value.findIndex(m => m._local_id === temp._local_id)
    if (idx >= 0) messages.value[idx]._status = 'error'
    err.value = e?.response?.data?.message || e?.message || 'Failed to send'
  } finally {
    sending.value = false
  }
}

/* ---------- logout ---------- */
function logout() {
  try {
    localStorage.clear()
    sessionStorage.clear()
  } finally {
    router.replace('/login')
  }
}
</script>

<style scoped>
/* =============== Layout =============== */
.layout{
  min-height:100vh;
  display:grid;
  grid-template-columns: 240px 1fr;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
  color:#0f172a;
}

/* Sidebar */
.sidebar{
  background:#f8fafc;
  border-right:1px solid #e2e8f0;
  padding:14px 12px;
}
.brand{
  height:44px; display:flex; align-items:center; padding:0 8px;
  font-weight:800; letter-spacing:.4px;
}
.nav{ padding:8px 4px; display:flex; flex-direction:column; gap:6px; }
.sec{ font-size:.78rem; color:#64748b; padding:6px 8px; }
.mt{ margin-top:8px; }
.link{
  display:block; padding:8px 10px; border-radius:8px; color:#0f172a; text-decoration:none;
}
.link:hover{ background:#eef2f7; }
.link-btn{ background:none; border:0; text-align:left; }

/* Main */
.main{ display:flex; flex-direction:column; }
.topbar{
  height:56px; display:flex; align-items:center; justify-content:space-between;
  padding:0 18px; border-bottom:1px solid rgba(20,100,60,.08); background:#fff8; backdrop-filter: blur(6px);
}
.title{ font-size:1.05rem; font-weight:800; }
.muted{ color:#64748b; font-weight:500; margin-left:.35rem }
.who{ color:#64748b; font-size:.95rem }

.content{
  max-width:960px; margin:0 auto; padding:12px 18px; width:100%;
}

/* Message list */
.scroll{
  height:64vh; overflow:auto; background:#f8fafc;
  border:1px solid #e2e8f0; border-radius:14px; padding:12px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
}

.msg{ display:flex; gap:.5rem; margin:.25rem 0; align-items:flex-end; }
.msg.mine{ flex-direction: row-reverse; }
.bubble{
  max-width:70ch; padding:.55rem .8rem; border-radius:.85rem;
  background:#ffffff; box-shadow:0 1px 1px rgba(0,0,0,.06); word-break:break-word;
}
.bubble.mine{
  background: linear-gradient(135deg,#22c55e,#3b82f6); color:#fff;
}
.meta{ font-size:.75rem; color:#6b7280; margin:0 .25rem; user-select:none; }
.meta .err{ color:#dc2626 }

/* Composer */
.composer{
  display:grid; grid-template-columns:1fr auto; gap:10px;
  margin-top:10px; padding:10px; border:1px solid #e2e8f0; background:#fff; border-radius:14px;
}
.input{
  min-height:44px; resize:vertical; border-radius:10px; border:1px solid #cbd5e1;
  padding:10px 12px; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }
.btn{
  border:0; border-radius:10px; color:#fff; padding:.65rem 1rem;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65 }
</style>
