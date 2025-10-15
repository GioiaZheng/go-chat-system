<template>
  <div class="page">
    <header class="topbar">
      <div class="brand">WASA <span class="grad">Chat</span></div>
      <div class="who" v-if="me">Signed in as {{ me }}</div>
    </header>

    <main class="container">
      <h2 class="title">My Conversations</h2>

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <div v-if="!loading && conversations.length === 0" class="empty">
        No conversations yet. Start a new chat from the Groups page.
      </div>

      <div v-else class="list" :aria-busy="loading">
        <article
          v-for="c in conversations"
          :key="c.id || c.conversation_id"
          class="item"
          role="button"
          @click="open(c)"
        >
          <div class="name">
            {{ c.name || 'Conversation' }}
            <div v-if="c.lastMessage?.content" class="last">
              {{ c.lastMessage.content }}
            </div>
          </div>
          <div class="meta">
            {{ displayTime(c) }}
          </div>
        </article>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '../components/ErrorMsg.vue'
import api, { getMyConversations } from '../services/api'

const router = useRouter()
const me = ref(localStorage.getItem('name') || 'user')
const conversations = ref([])
const loading = ref(false)
const err = ref('')
let timer = null

const hasToken = () =>
  !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

function normalizeList(raw) {
  const arr = Array.isArray(raw) ? raw : (raw?.items ?? [])
  // 轻量归一化：确保存在 id/name/lastMessage
  return arr.map(x => ({
    id: x.id ?? x.conversation_id ?? x.cid ?? x._id,
    name: x.name ?? x.title ?? 'Conversation',
    avatarUri: x.avatarUri ?? x.avatar_url ?? x.photo ?? null,
    lastMessage: x.lastMessage ?? x.last_message ?? null,
    updated_at: x.updated_at ?? x.updatedAt ?? null,
  })).filter(x => !!x.id)
}

async function load () {
  if (!hasToken()) { router.replace('/login'); return }
  if (loading.value) return
  loading.value = true
  err.value = ''
  try {
    const data = await getMyConversations() // 兼容 api 默认导出 & 命名导出
    conversations.value = normalizeList(data)
  } catch (e) {
    err.value = e?.response?.data?.error || e?.message || 'Failed to load'
  } finally {
    loading.value = false
  }
}

function displayTime(c) {
  const t = c.lastMessage?.createdAt || c.updated_at
  if (!t) return ''
  const d = new Date(t)
  // 防止无效日期
  return isNaN(d.getTime()) ? String(t).slice(0, 19) : d.toLocaleString()
}

function open(c) {
  router.push({ name: 'chat', params: { type: 'conv', id: c.id } })
}

function onVisibilityChange () {
  if (document.visibilityState === 'visible') load()
}

onMounted(() => {
  if (!hasToken()) { router.replace('/login'); return }
  load()
  // 5s 轮询（简单可靠）
  timer = setInterval(load, 5000)
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  document.removeEventListener('visibilitychange', onVisibilityChange)
})
</script>

<style scoped>
.page{
  min-height:100vh;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
  color:#0f172a;
}
.topbar{
  height:56px;
  display:flex;
  align-items:center;
  justify-content:space-between;
  padding:0 18px;
  border-bottom:1px solid rgba(20,100,60,.08);
  background:#fff8;
  backdrop-filter: blur(6px);
}
.brand{ font-weight:800; letter-spacing:.4px; }
.grad{
  background: linear-gradient(90deg,#22c55e,#3b82f6);
  -webkit-background-clip:text; background-clip:text; color:transparent;
}
.who{ color:#64748b; font-size:.95rem }

.container{ max-width:960px; margin:0 auto; padding:18px; }
.title{ font-size:1.5rem; font-weight:800; color:#334155; margin:10px 0 14px; }

.empty{
  color:#64748b; background:#fff; border:1px dashed #cbd5e1;
  border-radius:12px; padding:18px; text-align:center;
}

.list[aria-busy="true"]{ opacity:.7 }
.item{
  background:#fff;
  border:1px solid #e2e8f0;
  border-radius:14px;
  padding:14px 16px;
  display:flex; align-items:center; justify-content:space-between;
  margin-bottom:10px;
  box-shadow: 0 4px 14px rgba(2,6,23,.06);
}
.item:hover { background:#fbfeff; border-color:#dbeafe; cursor:pointer; }
.name{ font-weight:600; color:#0f172a }
.name .last{ font-weight: 400; color:#60708b; font-size: 13px; margin-top: 4px; }
.meta{ color:#64748b; font-size:.9rem; white-space:nowrap; }
</style>
