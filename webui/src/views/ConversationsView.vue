<!-- src/views/ConversationsView.vue: Conversation directory showing private and group chats. -->
<template>
  <div class="page">
    <header class="topbar">
      <div>
        <div class="title">Chats</div>
      </div>
    </header>

    <section class="workspace">
      <div class="content-inner">
        <div v-if="loading" class="loading">
          <span class="spinner"></span> Loading conversations…
        </div>
        <template v-else>
          <ErrorMsg v-if="err" :text="err" class="mb-2" />
          <button v-if="err" class="btn btn-sm btn-outline-secondary mb-3" @click="load">
            Retry
          </button>

          <div v-else class="panel">
            <aside class="list-pane">
              <div class="list-header">
                <input
                  type="search"
                  class="search"
                  placeholder="search conversations"
                  aria-label="Search conversations"
                />
              </div>

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
                <div class="section-head second">
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
            </aside>

            <div class="conversation-pane">
              <div class="preview-ghost">
                <span class="ghost-icon">💬</span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'

import { getMyConversations, getAvatarUrl, deleteConversation } from '@/services/api'
import { ensureAuthReady, isAuthenticated, currentUser } from '@/services/auth'

const router = useRouter()

const convs = ref([])
const loading = ref(false)
const err = ref('')

// Helper to read the current user identifier as a string.
const me = computed(() => currentUser.value)
const myId = () => String(me.value?.id ?? '')

// Derived conversation lists by type.
const privateConvs = computed(() => convs.value.filter(c => c.type !== 'group'))
const groupConvs = computed(() => convs.value.filter(c => c.type === 'group'))

// Format conversation timestamps for list display.
function fmtTime(t) {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return ''
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${m}/${day}`
}

// Resolve the display name for a conversation.
// Private chats show the peer name; groups show the conversation name.
function displayName(c) {
  if (c.type === 'group') return c.name || 'Group'

  const list = c.participants || []
  const other = list.find(u => String(u.id) !== myId())

  return other?.name || 'Chat'
}

// Resolve the avatar for private or group conversations.
// Private chats mirror the peer avatar; groups use their own avatar.
function avatarFor(c) {
  if (c.type === 'group') {
    return getAvatarUrl(c)
  }

  const list = c.participants || []
  const other = list.find(u => String(u.id) !== myId())
  return getAvatarUrl(other || {})
}

// Fallback initials when no avatar is available.
function initials(c) {
  const name = displayName(c)
  const match = name.match(/\b\w/g) || ['U']
  return match.slice(0, 2).join('').toUpperCase()
}

// Load conversations and normalize varying backend payload shapes.
async function load() {

  await ensureAuthReady()
  if (!isAuthenticated.value) {
    convs.value = []
    loading.value = false
    return
  }

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

// Navigation helpers for the chat list.
function open(c) {
  router.push({
    name: 'chat',
    params: { type: c.type === 'group' ? 'group' : 'conv', id: c.id },
  })
}

// Remove a conversation after user confirmation.
async function warnDelete(c) {
  if (!confirm(`Delete chat:\n"${displayName(c)}"?`)) return
  try {
    await deleteConversation(c.id)
    await load()
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to delete conversation'
  }
}

// Lifecycle hooks and event subscriptions.
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
  min-height: 100%;
  height: 100%;
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  background: #e5e7eb;
  color: #1f2937;
}
.topbar {
  height: 56px;
  display: grid;
  place-items: center;
  padding: 0 18px;
  border-bottom: 1px solid rgba(20, 100, 60, 0.06);
  background: #f8fafc;
  position: sticky;
  top: 0;
  z-index: 1;
}
.title {
  font-weight: 800;
  font-size: 1.3rem;
  color: #0f172a;
  text-align: center;
}

.workspace {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 12px 12px 16px;
}

.content-inner {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1 1 auto;
  min-height: 0;
}

.panel {
  flex: 1 1 auto;
  min-height: 0;
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  background: #f8fafc;
  border: 1px solid #d9dde3;
}
.list-pane {
  flex: 0 0 320px;
  background: #f7f7f7;
  border-right: 1px solid #d9dde3;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 10px;
}

.list-header {
  padding: 4px 0 8px;
}

.search {
  width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: var(--radius-control);
  padding: 10px 12px;
  background: #ffffff;
  font-size: 0.95rem;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.search:focus {
  outline: none;
  border-color: #22c55e;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.15);
}

.section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 4px 4px 2px;
}

.section-head.second {
  margin-top: 4px;
  border-top: 1px solid #e5e7eb;
  padding-top: 10px;
}

.section-title {
  margin: 0;
  font-weight: 700;
  color: #0f172a;
  font-size: 1.02rem;
}

.badge {
  display: inline-flex;
  min-width: 28px;
  height: 22px;
  border-radius: 0;
  padding: 0 8px;
  align-items: center;
  justify-content: center;
  border: 1px solid #cbd5e1;
  background: #fff;
  font-size: 0.82rem;
  color: #475569;
  font-weight: 600;
}
.badge--secondary {
  background: #f1f5f9;
  border-color: #d8e0e8;
}

.list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.item {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 0;
  padding: 10px 12px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 12px;
  transition: background 0.15s ease, border-color 0.15s ease, transform 0.12s ease;
}

.item:hover {
  background: #f8fafc;
  border-color: #d1d5db;
  transform: translateY(-1px);
}
.left {
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #d1d5db;
  background: #fff;
}
.avatar-fallback {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #d9f5e8;
  color: #0f766e;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #b2e5cd;
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
  color: #111827;
  font-size: 1rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  font-size: 0.85rem;
  color: #6b7280;
  white-space: nowrap;
}
.preview {
  color: #6b7280;
  font-size: 0.92rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.del {
  border: 0;
  border-radius: var(--radius-control);
  padding: 0.35rem 0.7rem;
  font-size: 0.8rem;
  color: #fff;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  transition: 0.2s;
}
.del:hover {
  opacity: 0.9;
}
.empty {
  text-align: center;
  color: #6b7280;
  padding: 16px 0;
}
.conversation-pane {
  flex: 1 1 auto;
  min-width: 0;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px;
  border-left: 1px solid #d9dde3;
}

.preview-ghost {
  text-align: center;
  color: #7a7a7a;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ghost-icon {
  font-size: 1.6rem;
}

.ghost-title {
  margin: 0;
  font-weight: 700;
  color: #4a4a4a;
}

.ghost-sub {
  margin: 0;
  font-size: 0.95rem;
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

@media (max-width: 992px) {
  .topbar {
    height: auto;
    align-items: flex-start;
    grid-template-columns: 1fr;
    padding: 10px 14px;
  }

  .workspace {
    padding: 12px;
  }
  .panel {
    flex-direction: column;
    min-height: auto;
  }


  .list-pane {
    flex-basis: auto;
    border-right: 0;
    border-bottom: 1px solid #e2e8f0;
  }

  .conversation-pane {
    min-height: 200px;
  }
}
</style>
