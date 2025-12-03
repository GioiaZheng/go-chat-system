<!-- src/views/ChatView.vue -->
<template>
  <div class="page">
    <!-- TOP BAR -->
    <header class="topbar">
      <button class="back" @click="router.back()">←</button>

      <div class="title">
        <strong>{{ headerTitle }}</strong>
        <small class="muted">
          {{ isGroup ? 'Group Chat' : 'Private Chat' }}
        </small>
      </div>
    </header>

    <!-- MAIN CONTENT -->
    <section class="content">
      <ErrorMsg v-if="err" :text="err" class="mb-2" />

      <div ref="scrollbox" class="scroll">
        <div
          v-for="m in messages"
          :key="m.id"
          class="row"
          :class="{ mine: isMine(m) }"
        >
          <!-- LEFT AVATAR  -->
          <div
            v-if="!isMine(m)"
            class="avatar"
            :style="{ backgroundImage: `url('${avatarFor(m)}')` }"
          ></div>

          <!-- MESSAGE BLOCK -->
          <div class="bubble-wrap" :class="{ mine: isMine(m) }">
            <!-- Sender Name (group only) -->
            <div class="who" v-if="showSenderName && !isMine(m)">
              {{ displayNameFor(m) }}
            </div>

            <!-- Bubble -->
            <div class="bubble" :class="{ mine: isMine(m) }">
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
            </div>

            <!-- COMMENT CHIP -->
            <div
              v-if="m._myLastComment"
              class="comment-chip"
              :class="{ mine: isMine(m) }"
            >
              {{ m._myLastComment }}
            </div>

            <!-- REMOVE COMMENT BUTTON -->
            <button
              v-if="m._myLastComment"
              class="btn-xs btn-secondary"
              @click="removeComment(m)"
            >
              Remove
            </button>

            <!-- REACTIONS -->
            <div class="actions">
              <button class="icon-btn" @click="toggleCommentBox(m)">💬</button>

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
            </div>

            <!-- COMMENT BOX -->
            <div v-if="m._showCommentBox" class="comment-box">
              <input
                v-model="m._commentDraft"
                class="comment-input"
                type="text"
                placeholder="Write a comment…"
                @keyup.enter="sendComment(m)"
              />
              <button class="btn-xs" @click="sendComment(m)">Send</button>
            </div>
          </div>

          <!-- RIGHT AVATAR (me) -->
          <div
            v-if="isMine(m)"
            class="avatar mine"
            :style="{ backgroundImage: `url('${myAvatar}')` }"
          ></div>
        </div>
      </div>

      <!-- COMPOSER -->
      <div class="composer">
        <textarea
          v-model="draft"
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
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import {
  isAuthed,
  getMyProfile,
  getMyConversations,
  getMessages,
  sendMessage,
  sendImageMessage,
  getAvatarUrl,
  absUrl,
  commentMessage,
  uncommentMessage,
} from '@/services/api'

const route = useRoute()
const router = useRouter()

// basic states
const convId = computed(() => String(route.params.id || ''))
const err = ref('')
const loading = ref(false)
const sending = ref(false)
const draft = ref('')
const messages = ref([])
const scrollbox = ref(null)
const imageInput = ref(null)

const me = ref(null)
const meId = computed(() => String(me.value?.id || ''))
const myAvatar = computed(() => getAvatarUrl(me.value || {}))

const currentConv = ref(null)
const isGroup = ref(false)
const participants = computed(() => currentConv.value?.participants || [])
const peer = computed(
  () => participants.value.find(u => String(u.id) !== meId.value) || null
)

// ---- Avatar helpers ----
function avatarFor(m) {
  if (isMine(m)) return myAvatar.value

  const sender = participants.value.find(u => String(u.id) === String(m.senderId))
  if (sender) return getAvatarUrl(sender)

  return '/default-avatar.png'
}

function displayNameFor(m) {
  const s = participants.value.find(u => String(u.id) === String(m.senderId))
  return s?.name || 'User'
}

// ---- Header title ----
const headerTitle = computed(() => {
  if (!currentConv.value) return 'Chat'
  if (!isGroup.value) return peer.value?.name || 'Chat'
  return currentConv.value.name || 'Group'
})

const showSenderName = computed(() => isGroup.value)

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

    const conv = items.find(c => String(c.id) === convId.value)
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

  // 处理图片：如果 type === 'image'，content 就是图片 URL
  const fileRel =
    raw.fileUrl ||
    raw.file_url ||
    (raw.type === 'image' ? raw.content : null)

  return {
    id: raw.id,
    content: raw.type === 'image' ? '' : (raw.content || raw.text || ''),
    type: raw.type === 'image' ? 'image' : 'text',
    fileAbsUrl: fileRel ? absUrl(fileRel) : null,
    senderId: String(senderId),
    _ts: new Date(ts).toISOString(),
    _showCommentBox: false,
    _commentDraft: '',
    _myLastComment: '',
    _myReactions: [],
  }
}

// ---- Load messages ----
async function loadMessages() {
  loading.value = true
  err.value = ''
  try {
    const data = await getMessages({ conversationId: convId.value, limit: 150 })

    // 兼容 envelope：{code,data:{messages}} 或 {messages} 或直接数组
    const list =
      data?.data?.messages ||
      data?.messages ||
      (Array.isArray(data) ? data : [])

    messages.value = list
      .map(normalizeMessage)
      .sort((a, b) => a._ts.localeCompare(b._ts))

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

async function onSend() {
  const t = draft.value.trim()
  if (!t) return

  sending.value = true
  err.value = ''
  try {
    await sendMessage({ conversationId: convId.value, content: t })
    draft.value = ''
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
    await sendImageMessage({ conversationId: convId.value, file })
    await loadMessages()
  } catch (e2) {
    err.value = e2?.response?.data?.message || 'Failed to send image'
  } finally {
    if (imageInput.value) {
      imageInput.value.value = ''
    }
  }
}

// ---- Comments ----
function toggleCommentBox(m) {
  m._showCommentBox = !m._showCommentBox
}

async function sendComment(m) {
  const c = m._commentDraft.trim()
  if (!c) return

  try {
    await commentMessage(m.id, {
      type: 'text',
      content: c,
    })
    m._myLastComment = c
    m._commentDraft = ''
    m._showCommentBox = false
  } catch (e) {
    err.value = e?.response?.data?.message || 'Failed to add comment'
  }
}

async function removeComment(m) {
  try {
    await uncommentMessage(m.id)
    m._myLastComment = ''
  } catch (e) {
    err.value = e?.response?.data?.message || 'Failed to remove comment'
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

// ---- Bootstrap ----
async function bootstrap() {
  if (!isAuthed()) {
    return router.replace('/login')
  }

  // 兼容 me profile envelope
  const prof = await getMyProfile()
  me.value = prof?.data?.user || prof?.user || prof || null

  await loadConversationMeta()
  await loadMessages()
}

onMounted(bootstrap)
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
  font-weight: 800;
  color: #1e293b;
  display: flex;
  flex-direction: column;
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
  border-radius: 50%;
  border: 1px solid #e2e8f0;

  /* 用背景图来显示头像，避免任何拉伸 */
  background-size: cover;        /* 等比例放大裁剪 */
  background-position: center;   /* 取中间 */
  background-repeat: no-repeat;
  flex-shrink: 0;
}

.avatar.mine {
  margin-left: 4px;    /* 自己的头像和气泡之间留一点空 */
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

.img {
  max-width: 260px;
  border-radius: 10px;
}

.meta {
  font-size: 0.7rem;
  color: #6b7280;
  margin-top: 3px;
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
  gap: 10px;
  padding: 10px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e1e5eb;
  align-items: center;
}

.input {
  flex: 1;
  padding: 10px;
  border-radius: 10px;
  border: 1px solid #cbd5e1;
  resize: none;
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
</style>
