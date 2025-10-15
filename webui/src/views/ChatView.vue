<template>
  <div class="wrap">
    <div class="bar">
      <h2 class="h5">
        Chat <small>({{ type }} / {{ id }})</small>
      </h2>
      <RouterLink class="link" to="/conversations">Back</RouterLink>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />

    <section class="panel">
      <div ref="scrollbox" class="scroll" @scroll.passive="onScroll">
        <LoadingSpinner v-if="loading && messages.length === 0" />
        <div v-for="m in messages" :key="m.id" class="msg" :class="{ mine: (m.senderId || m.sender_id) === meId }">
          <div class="bubble" :class="{ mine: (m.senderId || m.sender_id) === meId }">
            {{ m.content }}
          </div>
          <div class="meta">{{ formatMeta(m) }}</div>
        </div>
        <LoadingSpinner v-if="loadingMore" small />
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
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import api, { startConversation, getMessages, sendMessage } from '../services/api'

const route = useRoute()
const router = useRouter()
const type = computed(() => route.params.type)
const id = computed(() => route.params.id)

const messages = ref([])
const loading = ref(false)
const loadingMore = ref(false)
const sending = ref(false)
const err = ref('')
const draft = ref('')
const scrollbox = ref(null)
const beforeCursor = ref(null)

const meId = ref('')
try { meId.value = JSON.parse(localStorage.getItem('me') || '{}')?.id || '' } catch {}

const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))
const conversationId = ref(null)

/** 格式化消息元信息 */
function formatMeta(m) {
  const who = String(m.senderId ?? m.sender_id ?? '').slice(0, 8)
  const created = m.createdAt ?? m.created_at ?? ''
  const d = new Date(created)
  const when = isNaN(d.getTime()) ? String(created).slice(0, 19) : d.toLocaleString()
  return `${who} · ${when}`
}

async function resolveConversationId() {
  if (type.value === 'conv') {
    conversationId.value = String(id.value)
    return
  }
  if (type.value === 'private') {
    const res = await startConversation({ user_id: String(id.value) })
    conversationId.value =
      res?.conversationId ?? res?.id ?? res?.conversation_id ?? res?.cid ?? String(res)
    return
  }
  if (type.value === 'group') {
    const res = await startConversation({ group_id: String(id.value) })
    conversationId.value =
      res?.conversationId ?? res?.id ?? res?.conversation_id ?? res?.cid ?? String(res)
    return
  }
  throw new Error('Unknown chat type')
}

/** 首次加载或刷新 */
async function load() {
  if (!authed()) { router.replace('/login'); return }
  loading.value = true
  err.value = ''
  try {
    await resolveConversationId()
    const data = await getMessages({ conversationId: conversationId.value, limit: 20 })
    const msgs = Array.isArray(data) ? data : (data?.messages ?? [])
    messages.value = msgs
    // 更新 beforeCursor 为最早消息的时间
    if (msgs.length > 0) {
      beforeCursor.value = msgs[0].createdAt ?? msgs[0].created_at
    }
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load messages'
  } finally {
    loading.value = false
    await nextTick()
    if (scrollbox.value) scrollbox.value.scrollTop = scrollbox.value.scrollHeight
  }
}

/** 加载更早的历史记录 */
async function loadMore() {
  if (loadingMore.value || !beforeCursor.value) return
  loadingMore.value = true
  try {
    const data = await getMessages({
      conversationId: conversationId.value,
      limit: 20,
      beforeCursor: beforeCursor.value,
    })
    const older = Array.isArray(data) ? data : (data?.messages ?? [])
    if (older.length > 0) {
      beforeCursor.value = older[0].createdAt ?? older[0].created_at
      messages.value = [...older, ...messages.value]
      await nextTick()
      // 维持滚动位置（防止跳动）
      if (scrollbox.value) {
        const prevHeight = scrollbox.value.scrollHeight
        scrollbox.value.scrollTop = scrollbox.value.scrollHeight - prevHeight + 10
      }
    }
  } catch (e) {
    console.error('Load more failed:', e)
  } finally {
    loadingMore.value = false
  }
}

/** 发送消息（乐观插入） */
async function onSend() {
  const text = draft.value.trim()
  if (!text || sending.value) return
  if (!authed()) { router.replace('/login'); return }

  try {
    if (!conversationId.value) await resolveConversationId()
    sending.value = true
    const tempId = `tmp_${Date.now()}_${Math.random().toString(16).slice(2)}`
    const optimistic = {
      id: tempId,
      content: text,
      senderId: meId.value,
      createdAt: new Date().toISOString(),
    }
    messages.value = [...messages.value, optimistic]
    draft.value = ''
    await nextTick()
    if (scrollbox.value) scrollbox.value.scrollTop = scrollbox.value.scrollHeight

    const res = await sendMessage({
      conversationId: conversationId.value,
      content: text,
      type: 'text',
      status: 'sent',
    })
    const real = res?.resource ?? res
    if (real && real.id) {
      messages.value = messages.value.map(m => m.id === tempId ? real : m)
    } else {
      await load()
    }
  } catch (e) {
    messages.value = messages.value.filter(m => !String(m.id).startsWith('tmp_'))
    err.value = e?.response?.data?.message || e?.message || 'Failed to send message'
  } finally {
    sending.value = false
    await nextTick()
    if (scrollbox.value) scrollbox.value.scrollTop = scrollbox.value.scrollHeight
  }
}

/** 滚动事件：接近顶部加载更多 */
function onScroll() {
  if (!scrollbox.value) return
  if (scrollbox.value.scrollTop < 50 && !loadingMore.value) {
    loadMore()
  }
}

watch([type, id], load, { immediate: true })
</script>

<style scoped>
.wrap{ max-width:900px; margin:0 auto; padding:12px 16px; }
.bar{
  display:flex; align-items:center; justify-content:space-between;
  padding:8px 0; border-bottom:1px solid #e2e8f0; margin-bottom:8px;
}
.link{ color:#2563eb; text-decoration:none; }
.link:hover{ text-decoration:underline; }

.panel{
  background:#fff; border:1px solid #e2e8f0; border-radius:14px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
  overflow:hidden;
}
.scroll{ height:64vh; overflow:auto; background:#f8fafc; padding:12px; }

/* messages */
.msg{ display:flex; gap:.5rem; margin:.25rem 0; align-items:flex-end; }
.msg.mine{ flex-direction: row-reverse; }
.bubble{
  max-width:70ch; padding:.5rem .75rem; border-radius:.75rem;
  background:#ffffff; box-shadow:0 1px 1px rgba(0,0,0,.06); word-break:break-word;
}
.bubble.mine{
  background: linear-gradient(135deg,#22c55e,#3b82f6); color:#fff;
}
.meta{ font-size:.75rem; color:#6b7280; margin:0 .25rem; user-select:none; }

.composer{
  display:grid; grid-template-columns:1fr auto; gap:10px;
  padding:10px; border-top:1px solid #e2e8f0; background:#fff;
}
.input{
  min-height:44px; resize:vertical; border-radius:10px; border:1px solid #cbd5e1;
  padding:10px 12px; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }
.btn{
  border:0; border-radius:10px; color:#fff; padding:10px 16px;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65 }
</style>
