<!-- src/views/ConversationsView.vue: Conversation directory showing private and group chats. -->
<template>
  <div class="page">
    <div v-if="toastMessage" class="toast" role="status" aria-live="polite">
      <span class="checkmark">✓</span>
      <span>{{ toastMessage }}</span>
    </div>

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
            <ConversationList
              v-model:search="search"
              :private-convs="privateConvs"
              :group-convs="groupConvs"
              :selected-id="selectedId"
              :display-name="displayName"
              :avatar-for="avatarFor"
              :initials="initials"
              :fmt-time="fmtTime"
              :show-delete="true"
              variant="split"
              empty-direct-text="No chats yet — start in Contacts 👋"
              empty-group-text="No groups yet — create one 👋"
              @select="open"
              @delete="warnDelete"
            />

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
import ConversationList from '@/components/ConversationList.vue'

import {
  getMyConversations,
  getAvatarUrl,
  deleteConversation,
  leaveGroup,
  getGroupIdForConversation,
  preferredDisplayName,
  initialsFor,
  normalizeUser,
} from '@/services/api'
import { hydrateConversationList, removeConversation, upsertConversationMeta } from '@/services/conversationStore'
import { ensureAuthReady, isAuthenticated, currentUser } from '@/services/auth'

const router = useRouter()

const convs = ref([])
const loading = ref(false)
const err = ref('')
const search = ref('')
const selectedId = ref('')
const toastMessage = ref('')
let refreshTimer = null
let toastTimer = null

const PREVIEW_LIMIT = 20

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

function conversationPreviewForMessage(raw = {}, conv = null) {
  const base = previewForMessage(raw)
  if (!base) return ''
  return base
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

    convs.value = (items || [])
      .map(c => {
        const isGroupType = c?.type === 'group' || !!(c?.groupId || c?.group_id || c?.group?.id)
        const last = pickLastMessage(c) || {}
        const previewContent = conversationPreviewForMessage(last, {
          ...c,
          type: isGroupType ? 'group' : c?.type,
        })

        const time = messageTime(last)

        return {
          ...c,
          type: isGroupType ? 'group' : c?.type,
          participants: (c.participants || []).map(normalizeUser),
          last_preview: previewContent || 'No messages yet',
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

function matchesConversationId(conversation = {}, targetId = '') {
  const target = String(targetId || '')
  if (!target) return false
  const conversationId = String(
    conversation?.id ||
      conversation?.conversationId ||
      conversation?.conversation_id ||
      ''
  )
  const groupId = String(
    conversation?.groupId ||
      conversation?.group_id ||
      conversation?.group?.id ||
      ''
  )
  return target === conversationId || (groupId && target === groupId)
}

// Remove a conversation after user confirmation.
async function warnDelete(c) {
  if (!c) return
  if (c.type === 'group') {
    err.value = ''
    try {
      const groupId =
        c.groupId || c.group_id || c?.group?.id || (await getGroupIdForConversation(c.id))
      if (!groupId) throw new Error('Group not found')
      await leaveGroup(groupId)
      const targetId = String(c.id || '')
      convs.value = convs.value.filter(conv => String(conv.id) !== targetId)
      removeConversation(targetId)
      if (selectedId.value === targetId) selectedId.value = ''
      window.dispatchEvent(
        new CustomEvent('conversations:remove', {
          detail: { conversationId: targetId },
        })
      )
    } catch (e) {
      err.value = e?.response?.data?.message || e?.message || 'Failed to leave group'
    }
    return
  }
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

const handleRemoveEvent = e => {
  const detail = e?.detail || {}
  const targetId = detail.conversationId ? String(detail.conversationId) : ''
  if (!targetId) return
  const matched = convs.value.filter(c => matchesConversationId(c, targetId))
  convs.value = convs.value.filter(c => !matchesConversationId(c, targetId))
  const idsToRemove = new Set([targetId])
  matched.forEach(c => {
    const candidateIds = [
      c?.id,
      c?.conversationId,
      c?.conversation_id,
      c?.groupId,
      c?.group_id,
      c?.group?.id,
    ]
    candidateIds.forEach(id => {
      const key = String(id || '')
      if (key) idsToRemove.add(key)
    })
  })
  idsToRemove.forEach(id => removeConversation(id))
}

onMounted(async () => {
  const welcomeName = sessionStorage.getItem('toast:welcome')
  if (welcomeName) {
    toastMessage.value = `Welcome back, ${welcomeName}!`
    sessionStorage.removeItem('toast:welcome')
    toastTimer = setTimeout(() => {
      toastMessage.value = ''
      toastTimer = null
    }, 3000)
  }
  await load()
  window.addEventListener('auth:changed', load)
  window.addEventListener('conversations:refresh', handleRefreshEvent)
  window.addEventListener('conversations:remove', handleRemoveEvent)
  window.addEventListener('conversations:reload', load)
  refreshTimer = setInterval(() => {
    if (!isAuthenticated.value) return
    load()
  }, 15000)
})

onUnmounted(() => {
  window.removeEventListener('auth:changed', load)
  window.removeEventListener('conversations:refresh', handleRefreshEvent)
  window.removeEventListener('conversations:remove', handleRemoveEvent)
  window.removeEventListener('conversations:reload', load)
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  if (toastTimer) {
    clearTimeout(toastTimer)
    toastTimer = null
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
.workspace {
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 16px;
  overflow: hidden;
}
.toast {
  position: fixed;
  top: 20px;
  right: 20px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #ecfdf3;
  border: 1px solid #bbf7d0;
  color: #166534;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 600;
  font-size: 0.95rem;
  box-shadow: 0 12px 26px rgba(34, 197, 94, 0.18);
  z-index: 10;
}
.checkmark {
  font-size: 1.1rem;
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
