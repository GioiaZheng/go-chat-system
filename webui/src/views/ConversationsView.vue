<template>
  <div class="page">
    <header class="topbar">
      <div class="title">Chats</div>
    </header>

    <section class="content">
      <div v-if="loading" class="loading">
        <span class="spinner"></span> Loading conversations…
      </div>
      <template v-else>
        <ErrorMsg v-if="err" :text="err" class="mb-2" />
        <button v-if="err" class="btn btn-sm btn-outline-secondary mb-3" @click="load">
          Retry
        </button>

        <div v-else class="sections">
          <div class="section">
            <div class="section-head">
              <h3 class="section-title">Private Chats</h3>
              <span class="badge">{{ privateConvs.length }}</span>
            </div>

            <ul class="list">
              <li
                v-for="c in privateConvs"
                :key="c.id"
                class="item"
                @click="open(c)"
              >
                <div class="left">
                  <span v-if="!avatarFor(c)" class="avatar-fallback">{{ initials(c) }}</span>
                  <img v-else class="avatar" :src="avatarFor(c)" alt="avatar" />
                </div>

                <div class="info">
                  <div class="top">
                    <div class="name">{{ displayName(c) }}</div>
                    <div class="time">{{ fmtTime(c.last_time) }}</div>
                  </div>
                  <div class="bottom">
                    <div class="preview">{{ c.last_preview || 'No messages yet' }}</div>
                  </div>
                </div>

                <button class="del" @click.stop="warnDelete(c)">Delete</button>
              </li>

              <li v-if="!privateConvs.length" class="empty">No private chats yet.</li>
            </ul>
          </div>
          <div class="section">
            <div class="section-head">
              <h3 class="section-title">Group Chats</h3>
              <span class="badge badge--secondary">{{ groupConvs.length }}</span>
            </div>
            <ul class="list">
              <li
                v-for="c in groupConvs"
                :key="c.id"
                class="item"
                @click="open(c)"
              >
                <div class="left">
                  <span v-if="!avatarFor(c)" class="avatar-fallback">{{ initials(c) }}</span>
                  <img v-else class="avatar" :src="avatarFor(c)" alt="avatar" />
                </div>

                <div class="info">
                  <div class="top">
                    <div class="name">{{ displayName(c) }}</div>
                    <div class="time">{{ fmtTime(c.last_time) }}</div>
                  </div>
                  <div class="bottom">
                    <div class="preview">{{ c.last_preview || 'No messages yet' }}</div>
                  </div>
                </div>

                <button class="del" @click.stop="warnDelete(c)">Delete</button>
              </li>

              <li v-if="!groupConvs.length" class="empty">No group chats yet.</li>
            </ul>
          </div>
        </div>
      </template>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'

import {
  getMyConversations,
  getMyProfile,
  getAvatarUrl,
  deleteConversation,
} from '@/services/api'

const router = useRouter()

const convs = ref([])
const me = ref(null)
const loading = ref(false)
const err = ref('')

// 当前用户 id
const myId = () => String(me.value?.id ?? '')

// 分类：私聊 & 群聊
const privateConvs = computed(() => convs.value.filter(c => c.type !== 'group'))
const groupConvs = computed(() => convs.value.filter(c => c.type === 'group'))

// -----------------------------------------------------
// time format
// -----------------------------------------------------
function fmtTime(t) {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return ''
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${m}/${day}`
}


// -----------------------------------------------------
// ✔ 会话显示名称
// private → 对方名字（来自 participants）
// group   → 会话 name
// -----------------------------------------------------
function displayName(c) {
  if (c.type === 'group') return c.name || 'Group'

  const list = c.participants || []
  const other = list.find(u => String(u.id) !== myId())

  return other?.name || 'Chat'
}


// -----------------------------------------------------
// ✔ avatar
// private → 对方头像
// group   → 群头像
// -----------------------------------------------------
function avatarFor(c) {
  if (c.type === 'group') {
    return getAvatarUrl(c)
  }

  const list = c.participants || []
  const other = list.find(u => String(u.id) !== myId())
  return getAvatarUrl(other || {})
}


// fallback initials
function initials(c) {
  const name = displayName(c)
  const match = name.match(/\b\w/g) || ['U']
  return match.slice(0, 2).join('').toUpperCase()
}


// -----------------------------------------------------
// ✔ load conversations（从后端格式自动归一化）
// -----------------------------------------------------
async function load() {
  loading.value = true
  err.value = ''

  try {
    const res = await getMyConversations()
    const root = res?.data?.items || res?.data || res
    const items = Array.isArray(root) ? root : (root?.items ?? root?.list ?? [])

    const pickLastMessage = c => {
      if (!c) return null

      const fromKnownFields =
        c.lastMessage ||
        c.last_message ||
        c.last ||
        c.lastmessage ||
        c.lastMsg ||
        c.last_msg ||
        (Array.isArray(c.messages) ? c.messages[c.messages.length - 1] : null)

      if (fromKnownFields) return fromKnownFields

      if (typeof c.lastMessageContent === 'string') return { content: c.lastMessageContent }
      if (typeof c.last_message_content === 'string') return { content: c.last_message_content }

      return null
    }

    convs.value = (items || [])
      .map(c => {
        const last = pickLastMessage(c) || {}
        const previewContent =
          last.type === 'image' ? '[Image]' :
          last.type === 'file'  ? '[File]'  :
          last.content ||
          last.text ||
          last.body ||
          last.message ||
          last.Content ||
          last.Text ||
          last.Body || ''

        const time =
          last.createdAt ||
          last.created_at ||
          last.CreatedAt ||
          last.timestamp ||
          last.Timestamp ||
          c.updatedAt || c.updated_at || c.UpdatedAt ||
          c.createdAt || c.created_at || c.CreatedAt || null

        return {
          ...c,
          last_preview: previewContent || 'No messages yet',
          last_time: time,
        }
      })
      .sort((a, b) => new Date(b.last_time || 0) - new Date(a.last_time || 0))
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load conversations'
  } finally {
    loading.value = false
  }
}


// -----------------------------------------------------
// open chat
// -----------------------------------------------------
function open(c) {
  router.push({
    name: 'chat',
    params: { type: c.type === 'group' ? 'group' : 'conv', id: c.id },
  })
}


// -----------------------------------------------------
// delete conversation
// -----------------------------------------------------
async function warnDelete(c) {
  if (!confirm(`Delete chat:\n"${displayName(c)}"?`)) return
  try {
    await deleteConversation(c.id)
    await load()
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to delete conversation'
  }
}


// -----------------------------------------------------
// mounted
// -----------------------------------------------------
const handleRefreshEvent = e => {
  const detail = e?.detail || {}
  const targetId = detail.conversationId ? String(detail.conversationId) : ''

  if (targetId && Array.isArray(convs.value) && convs.value.length > 0) {
    const bumpedTime = detail.lastTime || new Date().toISOString()
    const bumpedPreview = detail.lastPreview

    convs.value = convs.value
      .map(c =>
        String(c.id) === targetId
          ? {
              ...c,
              last_time: bumpedTime,
              ...(bumpedPreview ? { last_preview: bumpedPreview } : {}),
            }
          : c
      )
      .sort((a, b) => new Date(b.last_time || 0) - new Date(a.last_time || 0))
  }

  load()
}

onMounted(async () => {
  try {
    me.value = await getMyProfile()
  } catch {}

  await load()
  window.addEventListener('auth:changed', load)
  window.addEventListener('conversations:refresh', handleRefreshEvent)
})

onUnmounted(() => {
  window.removeEventListener('auth:changed', load)
  window.removeEventListener('conversations:refresh', handleRefreshEvent)
})
</script>


<style scoped>
.page {
  min-height: 100vh;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
}
.topbar {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 18px;
  border-bottom: 1px solid rgba(20, 100, 60, 0.08);
  background: #fff8;
  backdrop-filter: blur(6px);
  position: sticky;
  top: 0;
  z-index: 1;
}
.title {
  font-weight: 800;
  color: #0f172a;
}
.content {
  max-width: 1100px;
  margin: 0 auto;
  padding: 16px;
  display: grid;
  gap: 16px;
}
.sections {
  display: grid;
  gap: 18px;
}
.section {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 14px;
  box-shadow: 0 6px 18px rgba(2, 6, 23, 0.06);
}
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 10px;
}
.section-title {
  margin: 0;
  font-weight: 700;
  color: #0f172a;
}
.badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 30px;
  padding: 2px 10px;
  border-radius: 999px;
  font-weight: 700;
  color: #0f172a;
  background: #e0f2fe;
}
.badge--secondary {
  background: #ecfdf3;
  color: #166534;
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 10px;
}
.item {
  display: flex;
  align-items: center;
  gap: 12px;
  justify-content: space-between;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 10px 12px;
  cursor: pointer;
  box-shadow: 0 6px 18px rgba(2, 6, 23, 0.05);
  transition: background 0.2s;
}
.item:hover {
  background: #f8fafc;
}
.left {
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #e2e8f0;
  background: #fff;
}
.avatar-fallback {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: #e0f7ee;
  color: #0f766e;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #a7f3d0;
}
.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.name {
  font-weight: 600;
  color: #0f172a;
  font-size: 1rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.time {
  font-size: 0.85rem;
  color: #64748b;
  white-space: nowrap;
}
.preview {
  color: #64748b;
  font-size: 0.92rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.del {
  border: 0;
  border-radius: 8px;
  padding: 0.25rem 0.55rem;
  font-size: 0.8rem;
  color: #fff;
  background: linear-gradient(135deg, #ef4444, #f97316);
  transition: 0.2s;
}
.del:hover {
  opacity: 0.9;
}
.empty {
  text-align: center;
  color: #64748b;
  padding: 20px 0;
}
.loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #475569;
}
.spinner {
  width: 1rem;
  height: 1rem;
  border: 2px solid rgba(15, 23, 42, 0.25);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
