<!-- src/views/ChatView.vue -->
<template>
  <div class="page">
    <!-- TOP BAR -->
    <header class="topbar">
      <button class="back" @click="router.back()">←</button>

      <div class="title">
        <div class="title-row">
          <div
            class="header-avatar"
            :style="headerAvatar ? { backgroundImage: `url('${headerAvatar}')` } : {}"
          >
            <span v-if="!headerAvatar">{{ headerTitle[0] || 'C' }}</span>
          </div>
          <div>
            <strong>{{ headerTitle }}</strong>
            <small class="muted">
              {{ isGroup ? 'Group Chat' : 'Private Chat' }}
            </small>
          </div>
        </div>
      </div>
    </header>

    <!-- MAIN CONTENT -->
    <section class="content">
      <ErrorMsg v-if="err" :text="err" class="mb-2" />
      <p v-else-if="notice" class="notice">{{ notice }}</p>
      <div ref="scrollbox" class="scroll">
        <div
          v-for="m in messages"
          :key="m.id"
          class="row"
          :id="`msg-${m.id}`"
          :class="{ mine: isMine(m), highlight: replyHighlightId === String(m.id) }"
        >
          <!-- LEFT AVATAR  -->
          <div v-if="!isMine(m)" class="avatar" :class="{ placeholder: !avatarFor(m) }">
            <img
              v-if="avatarFor(m)"
              class="avatar-img"
              :src="avatarFor(m)"
              :alt="avatarInitial(m)"
            />
            <span v-else class="avatar-initial">{{ avatarInitial(m) }}</span>
          </div>
          <!-- MESSAGE BLOCK -->
          <div class="bubble-wrap" :class="{ mine: isMine(m) }">
            <!-- Sender Name (group only) -->
            <div class="who" v-if="showSenderName && !isMine(m)">
              {{ displayNameFor(m) }}
            </div>

            <!-- Bubble -->
            <div class="bubble" :class="{ mine: isMine(m) }">
              <button
                v-if="m._replyPreview"
                class="inline-reply"
                type="button"
                @click="jumpToMessage(m.replyToId)"
              >
                <div v-if="m._replyFrom" class="reply-from">{{ m._replyFrom }}</div>
                <div class="reply-text">{{ m._replyPreview }}</div>
              </button>

              <template v-if="m.type === 'image' && m.fileAbsUrl">
                <img :src="m.fileAbsUrl" class="img" />
              </template>

              <template v-else>
                {{ m.content }}
              </template>
            </div>

            <!-- Timestamp -->
            <div class="meta">
              {{ fmtTime(m._ts) }}
              <span v-if="tickText(m)" class="ticks">{{ tickText(m) }}</span>
            </div>

            <!-- COMMENT CHIP -->
            <div
              v-if="m._myLastComment"
              class="comment-chip"
              :class="{ mine: isMine(m) }"
            >
              {{ m._myLastComment }}
            </div>

            <!-- REACTIONS / ACTIONS -->
            <div class="actions">
              <button class="icon-btn" title="Reply" @click="setReplyTarget(m)">↩️</button>
              <button class="icon-btn" title="Forward" @click="openForwardPicker(m)">🔗</button>
              <button
                class="icon-btn"
                :class="{ active: m._myReactions.includes('👍') }"
                @click="toggleReaction(m, '👍')"
              >
                👍
              </button>

              <button
                class="icon-btn"
                :class="{ active: m._myReactions.includes('❤️') }"
                @click="toggleReaction(m, '❤️')"
              >
                ❤️
              </button>

              <button
                class="icon-btn"
                :class="{ active: m._myReactions.includes('😂') }"
                @click="toggleReaction(m, '😂')"
              >
                😂
              </button>

              <button
                v-if="isMine(m)"
                class="icon-btn"
                :disabled="deletingMessageId === String(m.id)"
                title="Delete message"
                @click="confirmDeleteMessage(m)"
              >
                {{ deletingMessageId === String(m.id) ? '⌛' : '🗑️' }}
              </button>

            </div>
          </div>

          <!-- RIGHT AVATAR (me) -->
          <div v-if="isMine(m)" class="avatar mine" :class="{ placeholder: !myAvatar }">
            <img
              v-if="myAvatar"
              class="avatar-img"
              :src="myAvatar"
              :alt="avatarInitial(m)"
            />
            <span v-else class="avatar-initial">{{ avatarInitial(m) }}</span>
          </div>
        </div>
      </div>

      <!-- COMPOSER -->
      <div class="composer">

        <div v-if="replyTarget" class="reply-banner">
          Replying to {{ nameForSender(replyTarget.senderId) || 'message' }}:
          <span class="reply-snippet">
            {{
              replyTarget.content ||
                replyTarget._replyPreview ||
                (replyTarget.type === 'image' ? '[image]' : 'message')
            }}
          </span>

          <button class="btn-xs btn-secondary" type="button" @click="clearReplyTarget">Cancel</button>
        </div>
        <div class="composer-row">
          <textarea
            v-model="draft"
            ref="composerInput"
            class="input"
            placeholder="Type a message…"
            rows="1"
            @keyup.enter.exact.prevent="onSend"
          ></textarea>

          <!-- Attach button for images -->
          <button
            type="button"
            class="icon-btn attach"
            @click="triggerImagePicker"
            title="Send image"
          >
            📎
          </button>

          <!-- Hidden file input -->
          <input
            ref="imageInput"
            type="file"
            class="filepick"
            accept="image/*"
            @change="onPickImage"
          />

          <button class="btn" :disabled="!canSend" @click="onSend">
            {{ sending ? 'Sending…' : 'Send' }}
          </button>
        </div>
      </div>

      <!-- Forward picker -->
      <div v-if="forwardPanelOpen" class="forward-overlay">
        <div class="forward-modal">
          <header class="forward-header">
            <div>
              <strong>Forward message</strong>
              <div class="muted small">Select a chat to forward this message.</div>
            </div>
            <button class="close-btn" type="button" @click="closeForwardPicker">✕</button>
          </header>

          <input
            v-model="forwardSearch"
            class="forward-search"
            type="text"
            placeholder="Search chats"
          />

          <div class="forward-body">
            <p v-if="forwardLoading" class="muted">Loading conversations…</p>
            <ErrorMsg v-else-if="forwardError" :text="forwardError" />

            <template v-else>
              <button
                v-for="c in filteredForwardList"
                :key="c.id"
                class="forward-item"
                type="button"
                @click="forwardToConversation(c.id)"
              >
                <div
                  class="forward-avatar"
                  :style="
                    avatarForConversation(c, meId) ? { backgroundImage: `url('${avatarForConversation(c, meId)}')` } : {}
                  "
                >
                  <span v-if="!avatarForConversation(c, meId)">
                    {{ titleForConversation(c, meId)[0] || 'C' }}
                  </span>
                </div>

                <div class="forward-meta">
                  <div class="forward-name">{{ titleForConversation(c, meId) }}</div>
                  <div class="forward-type muted small">
                    {{ c.type === 'group' ? 'Group chat' : 'Direct chat' }}
                  </div>
                </div>
              </button>

              <p v-if="!filteredForwardList.length" class="muted">No conversations found.</p>
            </template>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import {
  isAuthed,
  getMyProfile,
  getMyConversations,
  getConversationMembers,
  getMessages,
  sendMessage,
  sendImageMessage,
  getAvatarUrl,
  absUrl,
  deleteMessage,
  commentMessage,
  uncommentMessage,
  forwardMessage,
  ticksFor,
  titleForConversation,
  avatarForConversation,
} from '@/services/api'

const route = useRoute()
const router = useRouter()

// basic states
const convId = computed(() => String(route.params.id || ''))
const err = ref('')
const notice = ref('')
const loading = ref(false)
const sending = ref(false)
const draft = ref('')
const messages = ref([])
const scrollbox = ref(null)
const imageInput = ref(null)
const replyTarget = ref(null)
const composerInput = ref(null)
const replyHighlightId = ref('')
const deletingMessageId = ref('')

const me = ref(null)
const meId = computed(() => String(me.value?.id || ''))
const myAvatar = computed(() => getAvatarUrl(me.value || {}))

const currentConv = ref(null)
const isGroup = ref(false)
const participants = computed(() => currentConv.value?.participants || [])
const peer = computed(
  () => participants.value.find(u => String(u.id) !== meId.value) || null
)

// forward modal state
const forwardPanelOpen = ref(false)
const forwardLoading = ref(false)
const forwardError = ref('')
const forwardSearch = ref('')
const forwardList = ref([])
const forwardTargetMessage = ref(null)

// ---- Avatar helpers ----
function avatarFor(m) {
  if (isMine(m)) return myAvatar.value

  const sender = participants.value.find(u => String(u.id) === String(m.senderId))
  if (sender) return getAvatarUrl(sender)

  return ''
}

function avatarInitial(m) {
  if (isMine(m)) return (me.value?.name || 'Me')[0] || 'M'
  const sender = participants.value.find(u => String(u.id) === String(m.senderId))
  const name = sender?.name || sender?.username || 'U'
  return (name[0] || 'U').toUpperCase()
}

function displayNameFor(m) {
  const s = participants.value.find(u => String(u.id) === String(m.senderId))
  return s?.name || 'User'
}

function nameForSender(userId) {
  if (!userId) return ''
  if (String(userId) === meId.value) return me.value?.name || 'Me'
  const s = participants.value.find(u => String(u.id) === String(userId))
  return s?.name || s?.username || 'User'
}

// ---- Header title ----
const headerTitle = computed(() => {
  if (!currentConv.value) return 'Chat'
  
  const title = titleForConversation(currentConv.value, meId.value)
  if (title && title !== 'Chat') return title

  if (!isGroup.value) return peer.value?.name || peer.value?.username || 'Chat'
  return currentConv.value.name || 'Group'
})

const headerAvatar = computed(() => {
  if (!currentConv.value) return ''
  if (isGroup.value) return getAvatarUrl(currentConv.value)
  return getAvatarUrl(peer.value || {})
})

const showSenderName = computed(() => isGroup.value)

async function refreshProfile() {
  // 兼容 me profile envelope
  const prof = await getMyProfile()
  me.value = prof?.data?.user || prof?.user || prof || null
}

// load conversation meta (for title, participants, type)
async function loadConversationMeta() {
  try {
    const raw = await getMyConversations()
    // 兼容 payload：可能是 {code,data:{items}}、{items} 或直接数组
    const items =
      raw?.data?.items ||
      raw?.items ||
      (Array.isArray(raw) ? raw : []) ||
      []

    let conv = items.find(c => String(c.id) === convId.value)

    if (conv) {
      const hasParticipants = Array.isArray(conv.participants) && conv.participants.length >= 2

      if (!hasParticipants) {
        try {
          const members = await getConversationMembers(convId.value)
          const memberList =
            (Array.isArray(members) && members) ||
            members?.data?.items ||
            members?.items ||
            []

          if (Array.isArray(memberList) && memberList.length > 0) {
            conv = { ...conv, participants: memberList }
          }
        } catch (memberErr) {
          console.error('loadConversationMeta members fallback failed', memberErr)
        }
      }
    }

    currentConv.value = conv || null
    isGroup.value = conv?.type === 'group'
  } catch (e) {
    // 不阻止消息加载，只在控制台看问题
    console.error('loadConversationMeta failed', e)
  }
}

// ---- Time format ----
function fmtTime(ts) {
  if (!ts) return ''
  return ts.replace('T', ' ').slice(0, 19)
}

// ---- Mine? ----
function isMine(m) {
  return String(m.senderId) === meId.value
}

// ---- Normalize message from API ----
function normalizeMessage(raw) {
  const senderId = raw.senderId || raw.sender_id || raw.userId
  const ts = raw.createdAt || raw.created_at || new Date().toISOString()
  const replyToId = raw.replyToId || raw.reply_to_id || null

  const replyContent = raw.replyTo?.content || raw.replyTo?.text || ''
  const replyType = raw.replyTo?.type || ''
  const replyPreview = replyContent || (replyType === 'image' ? '[image]' : '')

  // 处理图片：如果 type === 'image'，content 就是图片 URL
  const fileRel =
    raw.fileUrl ||
    raw.file_url ||
    raw.imageUrl ||
    raw.image_url ||
    raw.file ||
    (raw.type === 'image' ? raw.content : null)

  return {
    id: raw.id,
    content: raw.type === 'image' ? '' : (raw.content || raw.text || ''),
    type: raw.type === 'image' ? 'image' : 'text',
    fileAbsUrl: fileRel ? absUrl(fileRel) : null,
    senderId: String(senderId),
    _ts: new Date(ts).toISOString(),
    replyToId: replyToId ? String(replyToId) : '',
    _showCommentBox: false,
    _commentDraft: '',
    _myLastComment: '',
    _myReactions: [],
    _replyPreview: replyPreview,
    _replyFrom: raw.replyTo?.sender?.name || '',
  }
}

// ---- Load messages ----
async function loadMessages() {
  loading.value = true
  err.value = ''
  notice.value = ''
  try {
    const data = await getMessages({ conversationId: convId.value, limit: 150 })

    // 兼容 envelope：{code,data:{messages}} 或 {messages} 或直接数组
    const list =
      data?.data?.messages ||
      data?.messages ||
      (Array.isArray(data) ? data : [])

    const mapped = list
      .map(normalizeMessage)
      .sort((a, b) => a._ts.localeCompare(b._ts))
    const byId = new Map(mapped.map(m => [String(m.id), m]))

    mapped.forEach(m => {
      if (!m._replyPreview && m.replyToId) {
        const target = byId.get(String(m.replyToId))
        if (target) {
          const preview = target.content || (target.type === 'image' ? '[image]' : '')
          m._replyPreview = preview
          m._replyFrom = nameForSender(target.senderId)
        }
      }

      if (m._replyPreview && !m._replyFrom && m.replyToId) {
        const target = byId.get(String(m.replyToId))
        if (target) {
          m._replyFrom = nameForSender(target.senderId)
        }
      }
    })

    messages.value = mapped

    await nextTick()
    if (scrollbox.value) {
      scrollbox.value.scrollTop = scrollbox.value.scrollHeight
    }
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load messages'
  } finally {
    loading.value = false
  }
}

// ---- Send text ----
const canSend = computed(() => !!draft.value.trim() && !sending.value)

async function confirmDeleteMessage(m) {
  if (!m || !m.id) return
  if (!isMine(m)) return
  const ok = window.confirm('Delete this message?')
  if (!ok) return

  deletingMessageId.value = String(m.id)
  err.value = ''
  notice.value = ''
  try {
    await deleteMessage(m.id)
    await loadMessages()
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to delete message'
  } finally {
    deletingMessageId.value = ''
  }
}

async function onSend() {
  const t = draft.value.trim()
  if (!t) return

  sending.value = true
  err.value = ''
  notice.value = ''
  try {
    await sendMessage({
      conversationId: convId.value,
      content: t,
      replyToId: replyTarget.value?.id,
    })
    draft.value = ''
    replyTarget.value = null
    await loadMessages()
  } catch (e) {
    err.value = e?.response?.data?.message || 'Failed to send'
  } finally {
    sending.value = false
  }
}

// ---- Send image ----
function triggerImagePicker() {
  if (imageInput.value) {
    imageInput.value.click()
  }
}

async function onPickImage(e) {
  const file = e.target.files?.[0]
  if (!file) return

  err.value = ''
  try {
    await sendImageMessage({
      conversationId: convId.value,
      file,
      replyToId: replyTarget.value?.id,
    })
    await loadMessages()
  } catch (e2) {
    err.value = e2?.response?.data?.message || 'Failed to send image'
  } finally {
    if (imageInput.value) {
      imageInput.value.value = ''
    }
  }
}

// ---- Reactions (use comment/uncomment backend) ----
async function toggleReaction(m, emoji) {
  if (!Array.isArray(m._myReactions)) m._myReactions = []

  const has = m._myReactions.includes(emoji)

  try {
    if (has) {
      // remove all reactions/comments on this message for simplicity
      await uncommentMessage(m.id)
      m._myReactions = []
    } else {
      await commentMessage(m.id, {
        type: 'emoji',
        content: emoji,
      })
      m._myReactions = [emoji]
    }
  } catch (e) {
    err.value = e?.response?.data?.message || 'Failed to react'
  }
}


async function loadForwardList() {
  forwardLoading.value = true
  forwardError.value = ''
  try {
    const raw = await getMyConversations()
    forwardList.value =
      raw?.data?.items || raw?.items || (Array.isArray(raw) ? raw : []) || []
  } catch (e) {
    forwardError.value =
      e?.response?.data?.message || e?.message || 'Failed to load conversations'
  } finally {
    forwardLoading.value = false
  }
}

function openForwardPicker(m) {
  forwardTargetMessage.value = m
  forwardPanelOpen.value = true
  forwardSearch.value = ''
  forwardError.value = ''
  loadForwardList()
}

function closeForwardPicker() {
  forwardPanelOpen.value = false
  forwardTargetMessage.value = null
}

const filteredForwardList = computed(() => {
  const q = forwardSearch.value.trim().toLowerCase()
  if (!q) return forwardList.value
  return forwardList.value.filter(c => {
    const title = titleForConversation(c, meId.value).toLowerCase()
    const type = (c.type || '').toLowerCase()
    return title.includes(q) || type.includes(q)
  })
})

async function forwardToConversation(targetConvId) {
  if (!forwardTargetMessage.value) return
  forwardError.value = ''
  notice.value = ''
  try {
    await forwardMessage(forwardTargetMessage.value.id, targetConvId)
    notice.value = 'Message forwarded successfully.'
    closeForwardPicker()
  } catch (e) {
    forwardError.value =
      e?.response?.data?.message || e?.message || 'Failed to forward message'
  }
}

function focusComposer() {
  nextTick(() => {
    if (composerInput.value) {
      composerInput.value.focus()
      composerInput.value.scrollTop = composerInput.value.scrollHeight
    }
  })
}

function setReplyTarget(m) {
  replyTarget.value = m
  replyHighlightId.value = String(m?.id || '')
  focusComposer()
}

function clearReplyTarget() {
  replyTarget.value = null
  replyHighlightId.value = ''
}

function jumpToMessage(targetId) {
  if (!targetId) return
  const el = document.getElementById(`msg-${targetId}`)
  if (!el) return

  replyHighlightId.value = String(targetId)
  el.scrollIntoView({ behavior: 'smooth', block: 'center' })

  setTimeout(() => {
    if (replyHighlightId.value === String(targetId)) {
      replyHighlightId.value = ''
    }
  }, 1500)
}

function tickText(m) {
  const v = ticksFor(m, meId.value)
  if (v === 3) return '✓✓ read'
  if (v === 2) return '✓✓ delivered'
  if (v === 1) return '✓ sent'
  if (v === 0) return '…'
  return ''
}


// ---- Bootstrap ----
async function bootstrap() {
  if (!isAuthed()) {
    return router.replace('/login')
  }

  await refreshProfile()

  await loadConversationMeta()
  await loadMessages()
}

function handleAuthChanged() {
  refreshProfile().catch(() => {})
}

onMounted(() => {
  window.addEventListener('auth:changed', handleAuthChanged)
  bootstrap()
})

onBeforeUnmount(() => {
  window.removeEventListener('auth:changed', handleAuthChanged)
})

watch(convId, async () => {
  await loadConversationMeta()
  await loadMessages()
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #f2f5f8;
}

.topbar {
  position: relative;
  display: flex;
  align-items: center;
  gap: 14px;
  height: 56px;
  padding: 0 16px;
  border-bottom: 1px solid #e2e8f0;
  background: #fff;
}

.back {
  border: none;
  background: none;
  cursor: pointer;
  font-size: 1.3rem;
}

.title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-weight: 800;
  color: #1e293b;
  display: flex;
  flex-direction: column;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: center;
}

.header-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #e2e8f0;
  background-size: cover;
  background-position: center;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #334155;
}

.muted {
  font-size: 0.75rem;
  color: #64748b;
}

.content {
  max-width: 900px;
  margin: 0 auto;
  padding: 16px;
}

.notice {
  background: #ecfeff;
  color: #0f766e;
  border: 1px solid #99f6e4;
  border-radius: 10px;
  padding: 8px 10px;
  margin-bottom: 8px;
}

.scroll {
  height: 65vh;
  overflow-y: auto;
  background: #f8fafc;
  border-radius: 14px;
  padding: 12px;
  border: 1px solid #e1e5eb;
}

.row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 8px 0;
}

.row.mine {
  justify-content: flex-end;
}
.avatar {
  width: 36px;
  height: 36px;
  aspect-ratio: 1;
  border-radius: 50%;
  overflow: hidden;

  flex-shrink: 0;
  display: grid;
  place-items: center;
  background: #e2e8f0;
}

.avatar.mine {
  margin-left: 4px;    /* 自己的头像和气泡之间留一点空 */
}
.avatar.placeholder {
  background: #e2e8f0;
  border: 1px solid #e2e8f0;
  color: #475569;
  font-weight: 700;
}

.avatar-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #e2e8f0;
  object-position: center;
  display: block;
}

.avatar-initial {
  font-size: 0.9rem;
}

.bubble-wrap {
  display: flex;
  flex-direction: column;
  max-width: 70%;
}

.bubble-wrap.mine {
  align-items: flex-end;
}

.who {
  font-size: 0.8rem;
  margin-bottom: 3px;
  color: #475569;
}

.bubble {
  padding: 8px 12px;
  border-radius: 14px;
  background: #ffffff;
  max-width: 100%;
  word-wrap: break-word;
}

.bubble.mine {
  background: #95ec69;
  color: #000;
}

.inline-reply {
  width: 100%;
  text-align: left;
  border: none;
  outline: none;
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 6px;
  padding: 6px 10px;
  border-left: 3px solid #3b82f6;
  background: #eef2f7;
  font-size: 0.86rem;
  color: #334155;
  border-radius: 8px;
  cursor: pointer;
}

.inline-reply:hover {
  background: #e2e8f0;
}

.inline-reply .reply-from {
  font-weight: 600;
  color: #0f172a;
}

.inline-reply .reply-text {
  color: #475569;
}

.img {
  max-width: 260px;
  border-radius: 10px;
}

.meta {
  font-size: 0.7rem;
  color: #6b7280;
  margin-top: 3px;
}

.ticks {
  margin-left: 6px;
  color: #0f766e;
}

.row.highlight .bubble {
  box-shadow: 0 0 0 2px #22c55e, 0 8px 20px rgba(0, 0, 0, 0.1);
}

.comment-chip {
  margin-top: 6px;
  background: #ffffff;
  border-radius: 12px;
  padding: 4px 12px;
  font-size: 0.82rem;
  max-width: 80%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.comment-chip.mine {
  background: #d9f7be;
}

.actions {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}

.icon-btn {
  border: none;
  background: none;
  cursor: pointer;
  font-size: 1.05rem;
  padding: 3px;
  border-radius: 6px;
}

.icon-btn.active {
  background: rgba(34, 197, 94, 0.2);
}

.comment-box {
  margin-top: 6px;
  display: flex;
  gap: 6px;
}

.comment-input {
  flex: 1;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 5px 8px;
}

.btn-xs {
  border: none;
  padding: 5px 12px;
  border-radius: 8px;
  background: #22c55e;
  color: white;
  cursor: pointer;
  font-size: 0.8rem;
}

.btn-secondary {
  background: #e2e8f0;
  color: #475569;
}

.composer {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e1e5eb;
}

.composer-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.input {
  flex: 1;
  padding: 10px;
  border-radius: 10px;
  border: 1px solid #cbd5e1;
  resize: none;
}

.reply-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
  padding: 6px 8px;
  margin-bottom: 6px;
  width: 100%;
}

.reply-snippet {
  color: #0f172a;
  font-weight: 500;
}

.btn {
  border: none;
  border-radius: 10px;
  background: #22c55e;
  padding: 10px 16px;
  color: white;
  white-space: nowrap;
}

.filepick {
  display: none;
}

.attach {
  font-size: 1.2rem;
}


.forward-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  padding: 12px;
}

.forward-modal {
  width: min(460px, 92vw);
  max-height: 80vh;
  background: #fff;
  border-radius: 14px;
  padding: 14px;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.forward-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.forward-search {
  width: 100%;
  padding: 10px;
  border-radius: 10px;
  border: 1px solid #cbd5e1;
}

.forward-body {
  overflow-y: auto;
  max-height: 55vh;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.forward-item {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  background: #f8fafc;
  cursor: pointer;
  transition: 0.2s;
}

.forward-item:hover {
  background: #eef2ff;
  border-color: #c7d2fe;
}

.forward-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: #e2e8f0;
  background-size: cover;
  background-position: center;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #475569;
}

.forward-meta {
  text-align: left;
}

.forward-name {
  font-weight: 700;
  color: #0f172a;
}

.forward-type {
  margin-top: 2px;
}

.close-btn {
  border: none;
  background: #f1f5f9;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  cursor: pointer;
}

</style>
