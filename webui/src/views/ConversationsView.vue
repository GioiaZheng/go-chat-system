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
                  v-model.trim="search"
                  type="search"
                  class="search"
                  placeholder="search conversations"
                  aria-label="Search conversations"
                />
              </div>

              <div class="section">
                <div class="section-head">
                  <h3 class="section-title text-secondary">Direct</h3>
                  <span class="badge">{{ privateConvs.length }}</span>
                </div>

                <ul class="list">
                  <li
                    v-for="c in privateConvs"
                    :key="c.id"
                    class="item"
                    :class="{ active: selectedId === String(c.id) }"
                    @click="open(c)"
                  >
                    <div class="left">
                      <span v-if="!avatarFor(c)" class="avatar-fallback avatar-circle">{{ initials(c) }}</span>
                      <img v-else class="avatar avatar-circle" :src="avatarFor(c)" alt="avatar" />
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

                  <li v-if="!privateConvs.length" class="empty">No direct chats yet.</li>
                </ul>
              </div>
              <div class="section">
                <div class="section-head second">
                  <h3 class="section-title text-secondary">Groups</h3>
                  <span class="badge badge--secondary">{{ groupConvs.length }}</span>
                </div>

                <ul class="list">
                  <li
                    v-for="c in groupConvs"
                    :key="c.id"
                    class="item"
                    :class="{ active: selectedId === String(c.id) }"
                    @click="open(c)"
                  >
                    <div class="left">
                      <span v-if="!avatarFor(c)" class="avatar-fallback avatar-circle">{{ initials(c) }}</span>
                      <img v-else class="avatar avatar-circle" :src="avatarFor(c)" alt="avatar" />
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
                  <li v-if="!groupConvs.length" class="empty">▸ You have no group chats. Create one ➕</li>
                </ul>
              </div>
            </aside>

            <div class="conversation-pane">
              <div class="preview-empty">
                <div class="preview-icon" aria-hidden="true">
                  <svg viewBox="0 0 48 48" role="img" focusable="false">
                    <path
                      d="M15 18h18M15 24h12M6 22c0-7.732 6.268-14 14-14h8c7.732 0 14 6.268 14 14s-6.268 14-14 14h-8l-8 6v-8c-3.314-2.564-6-7.106-6-12z"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2.2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                </div>
                <h2 class="preview-title">Choose a chat to view messages</h2>
                <p class="preview-description">
                  Your chats will appear here once you start a conversation.
                </p>
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

import { getMyConversations, getAvatarUrl, deleteConversation, preferredDisplayName, initialsFor, normalizeUser } from '@/services/api'
import { hydrateConversationList, upsertConversationMeta } from '@/services/conversationStore'
import { ensureAuthReady, isAuthenticated, currentUser } from '@/services/auth'

const router = useRouter()

const convs = ref([])
const loading = ref(false)
const err = ref('')
const search = ref('')
const selectedId = ref('')
let refreshTimer = null

const PREVIEW_LIMIT = 30

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

  const type = String(raw?.type || '').toLowerCase()
  if (type === 'image' || raw?.imageUrl || raw?.image_url) return '📷 Photo'
  if (
    type === 'reaction' ||
    type === 'emoji' ||
    raw?.reaction ||
    raw?.reactionType ||
    raw?.reaction_type ||
    raw?.emoji
  ) {
    return '😃 Emoji'
  }

  const content =
    raw?.content ||
    raw?.text ||
    raw?.body ||
    raw?.message ||
    raw?.Content ||
    raw?.Text ||
    raw?.Body ||
    ''

  return formatPreviewText(content)
}

function senderProfileFromRaw(raw = {}) {
  const senderRaw =
    raw.sender || raw.user || raw.author || raw.from || raw.owner || raw.created_by || raw.createdBy || {}

  const senderIdValue =
    raw.senderId || raw.sender_id || raw.userId || senderRaw.id || senderRaw.userId || senderRaw.user_id

  const senderName =
    senderRaw.name ||
    senderRaw.username ||
    senderRaw.displayName ||
    senderRaw.display_name ||
    senderRaw.fullName ||
    senderRaw.full_name ||
    raw.senderName ||
    raw.sender_name ||
    ''

  const normalized = normalizeUser({
    ...senderRaw,
    id: senderIdValue ?? senderRaw.id,
    name: senderName,
  })

  const senderId = normalized.id ? String(normalized.id) : ''

  return senderId
    ? {
        id: senderId,
        name: normalized.name || normalized.username || senderName,
        username: normalized.username,
      }
    : { id: '', name: normalized.name || senderName }
}

function conversationPreviewForMessage(raw = {}, conv = null) {
  const base = previewForMessage(raw)
  if (!base) return ''
  const isGroupType =
    conv?.type === 'group' ||
    !!(conv?.groupId || conv?.group_id || conv?.group?.id)
  if (!isGroupType) return base
  const sender = senderProfileFromRaw(raw)
  const senderName = preferredDisplayName(sender || {}) || sender?.name || sender?.username || ''
  return `${senderName || 'Someone'}: ${base}`
}

// Helper to read the current user identifier as a string.
const me = computed(() => currentUser.value)
const myId = () => String(me.value?.id ?? '')

// Derived conversation lists by type with inline search support.
const filteredConversations = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return convs.value

  return convs.value.filter((c) => {
    const name = (displayName(c) || '').toLowerCase()
    const preview = (c.last_preview || '').toLowerCase()
    return name.includes(q) || preview.includes(q)
  })
})

const privateConvs = computed(() => filteredConversations.value.filter(c => c.type !== 'group'))
const groupConvs = computed(() => filteredConversations.value.filter(c => c.type === 'group'))

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
  if (c.type === 'group') return preferredDisplayName(c)

  const list = c.participants || []
  const other = list.find(u => String(u.id) !== myId())

  return preferredDisplayName(other || {})
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
  return initialsFor({ name: displayName(c) }, 'U')
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
        const isGroupType = c?.type === 'group' || !!(c?.groupId || c?.group_id || c?.group?.id)
        const last = pickLastMessage(c) || {}
        const previewContent = conversationPreviewForMessage(last, {
          ...c,
          type: isGroupType ? 'group' : c?.type,
        })

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
          type: isGroupType ? 'group' : c?.type,
          participants: (c.participants || []).map(normalizeUser),
          last_preview: previewContent || (isGroupType ? 'Say hello 👋' : 'No messages yet'),
          last_time: time,
        }
      })
      .sort((a, b) => new Date(b.last_time || 0) - new Date(a.last_time || 0))
    hydrateConversationList(convs.value)
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load conversations'
  } finally {
    loading.value = false
  }
}

// Navigation helpers for the chat list.
function open(c) {
  if (c) {
    selectedId.value = String(c.id || '')
    upsertConversationMeta(c)
    window.dispatchEvent(
      new CustomEvent('conversations:hydrate', {
        detail: {
          conversationId: String(c.id || ''),
          meta: c,
        },
      })
    )
  }
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

  const bumpedName = detail.name
  const bumpedAvatar = detail.avatar

  if (targetId && Array.isArray(convs.value) && convs.value.length > 0) {
    const bumpedTime = detail.lastTime
    const bumpedPreview = detail.lastPreview
    convs.value = convs.value
      .map(c =>
        String(c.id) === targetId
          ? {
              ...c,
              ...(bumpedTime ? { last_time: bumpedTime } : {}),
              ...(bumpedPreview ? { last_preview: bumpedPreview } : {}),
              ...(bumpedName ? { name: bumpedName } : {}),
              ...(bumpedAvatar ? { avatar: bumpedAvatar } : {}),
            }
          : c
      )
      .sort((a, b) => new Date(b.last_time || 0) - new Date(a.last_time || 0))
    upsertConversationMeta({
      id: targetId,
      ...(bumpedName ? { name: bumpedName } : {}),
      ...(bumpedAvatar ? { avatar: bumpedAvatar } : {}),
    })
    return
  }

  load()
}

onMounted(async () => {
  await load()
  window.addEventListener('auth:changed', load)
  window.addEventListener('conversations:refresh', handleRefreshEvent)
  window.addEventListener('conversations:reload', load)
  refreshTimer = setInterval(() => {
    if (!isAuthenticated.value) return
    load()
  }, 15000)
})

onUnmounted(() => {
  window.removeEventListener('auth:changed', load)
  window.removeEventListener('conversations:refresh', handleRefreshEvent)
  window.removeEventListener('conversations:reload', load)
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<style scoped>
.page {
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
  font-weight: 700;
  font-size: var(--font-title);
  color: #0f172a;
  text-align: center;
}

.workspace {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 16px;
  overflow: hidden;
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
  display: flex;
  flex-direction: row;
  background: var(--panel);
  border: 1px solid var(--border);
  overflow: hidden;
}
.list-pane {
  flex: 0 0 320px;
  background: var(--panel);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 10px;
  min-height: 0;
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
  font-size: var(--font-primary);
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
  min-height: 0;
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
  font-size: var(--font-secondary);
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
  font-size: var(--font-secondary);
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
  gap: 10px;
  flex: 1 1 auto;
  overflow-y: auto;
  max-height: 100%;
  min-height: 0;
}

.item {
  background: #ffffff;
  border: 1px solid var(--border);
  border-radius: var(--radius-control);
  padding: 14px 16px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 14px;
  transition: background 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease;
  position: relative;
}

.item:hover {
  background: rgba(34, 197, 94, 0.1);
  border-color: rgba(34, 197, 94, 0.35);
  box-shadow: 0 6px 16px rgba(15, 23, 42, 0.08);
}

.item.active {
  background: #dcfce7;
  border-color: #22c55e;
}

.item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 10px;
  bottom: 10px;
  width: 3px;
  background: #22c55e;
}

.left {
  display: flex;
  align-items: center;
  justify-content: center;
}

.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.top {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 12px;
}

.name {
  font-weight: 600;
  color: #0f172a;
  font-size: var(--font-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.handle {
  color: #64748b;
  font-size: var(--font-secondary);
}

.time {
  font-size: 0.92rem;
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
  border-radius: var(--radius-control);
  padding: 0.35rem 0.7rem;
  font-size: var(--font-secondary);
  color: #fff;
  background: linear-gradient(135deg, #ef4444, #dc2626);
  transition: 0.2s;
  opacity: 0;
  pointer-events: none;
}

.item:hover .del,
.item:focus-within .del {
  opacity: 1;
  pointer-events: auto;
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
  background: var(--panel);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 24px;
  border-left: 1px solid var(--border);
}

.preview-empty {
  max-width: 520px;
  margin-top: 12px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #1f2937;
}

.preview-icon {
  width: 64px;
  height: 64px;
  border-radius: 22px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #16a34a;
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.18), rgba(22, 163, 74, 0.1));
  box-shadow: 0 10px 24px rgba(22, 163, 74, 0.12);
}
.preview-icon svg {
  width: 32px;
  height: 32px;
}

.preview-title {
  font-size: 1.4rem;
  font-weight: 800;
  margin: 0;
  color: #1f2937;
}

.preview-description {
  margin: 0;
  color: #475569;
  font-size: var(--font-primary);
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
