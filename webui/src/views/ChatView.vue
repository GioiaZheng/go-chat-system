<!-- src/views/ChatView.vue: Chat experience with in-place conversation switching. -->
<template>
  <div class="page">
    <!-- Top bar: back navigation and counterpart details. -->
    <header class="topbar">
      <button
        class="back"
        type="button"
        aria-label="Go back"
        title="Back"
        @click="router.back()"
      >
        ←
      </button>

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
              {{ isGroup ? 'Group' : 'Direct message' }}
            </small>
          </div>
        </div>
      </div>
    </header>

    <!-- Main content area: errors/notices, conversation rail, and chat canvas. -->
    <section class="content">
      <div class="content-inner">
        <ErrorMsg v-if="err" :text="err" class="mb-2" />
        <p v-else-if="notice" class="notice" role="status" aria-live="polite">{{ notice }}</p>
        <div class="chat-layout" :class="{ 'has-group': isGroup }">
          <aside class="conv-rail">
            <div class="conv-rail__header">
              <div>
                <div class="conv-rail__title">Conversations</div>
                <div class="conv-rail__hint">Switch chats without leaving the page.</div>
            </div>
            <span class="conv-rail__badge">{{ filteredConversations.length }}</span>
          </div>

          <div class="conv-rail__search">
            <input
              v-model.trim="convSearch"
              type="text"
              placeholder="Search by name or group"
              aria-label="Search conversations"
            />
          </div>

          <div class="conv-rail__list" role="list">
            <p v-if="convLoading" class="muted">Loading conversations…</p>
            <ErrorMsg v-else-if="convErr" :text="convErr" />
            <template v-else>
              <button
                v-for="c in filteredConversations"
                :key="c.id"
                type="button"
                class="conv-pill"
                :class="{ active: String(c.id) === convId }"
                @click="switchConversation(c)"
              >
                <div
                  class="conv-pill__avatar"
                  :class="{ placeholder: !avatarForConversation(c, meId) }"
                  :style="avatarForConversation(c, meId) ? { backgroundImage: `url('${avatarForConversation(c, meId)}')` } : {}"
                >
                  <span v-if="!avatarForConversation(c, meId)">{{ titleForConversation(c, meId)[0] || 'C' }}</span>
                </div>

                <div class="conv-pill__meta">
                  <div class="conv-pill__top">
                    <span class="conv-pill__name">{{ titleForConversation(c, meId) }}</span>
                    <span class="conv-pill__time">{{ convTime(c.last_time) }}</span>
                  </div>
                  <div class="conv-pill__bottom">{{ c.last_preview || 'No messages yet' }}</div>
                </div>

                <span class="conv-pill__type" :class="{ group: c.type === 'group' }">
                  {{ c.type === 'group' ? 'Group' : 'Direct' }}
                </span>
              </button>

              <p v-if="!filteredConversations.length" class="muted">No conversations found.</p>
            </template>
          </div>
        </aside>
        <div class="chat-main">
          <div ref="scrollbox" class="scroll">
            <template v-if="messages.length">
              <div
                v-for="m in messages"
                :key="m.id"
                class="row"
                :id="`msg-${m.id}`"
                :class="{ mine: isMine(m), highlight: replyHighlightId === String(m.id) }"
              >
                <!-- Left avatar for incoming messages. -->
                <div
                  v-if="!isMine(m)"
                  class="avatar"
                  :class="{ placeholder: !avatarFor(m) }"
                  :style="avatarBg(avatarFor(m))"
                >
                  <span v-if="!avatarFor(m)" class="avatar-initial">{{ avatarInitial(m) }}</span>
                </div>
                <!-- Message block with reply preview, content, and timestamp. -->
                <div class="bubble-wrap" :class="{ mine: isMine(m) }">
                  <!-- Sender label (group conversations only). -->
                  <div class="who" v-if="showSenderName && !isMine(m)">
                    {{ displayNameFor(m) }}
                  </div>
                  <!-- Bubble content (text or image). -->
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

                    <div v-if="m.fileAbsUrl" class="img-wrap">
                      <img :src="m.fileAbsUrl" class="img" />
                    </div>

                    <div v-if="m.content" class="text-block">
                      {{ m.content }}
                    </div>
                  </div>
                  <!-- Timestamp and delivery markers. -->
                  <div class="meta">
                    {{ fmtTime(m._ts) }}
                    <span v-if="tickText(m)" class="ticks">{{ tickText(m) }}</span>
                  </div>

                  <!-- Comment summary for the current user. -->
                  <div
                    v-if="m._myLastComment"
                    class="comment-chip"
                    :class="{ mine: isMine(m) }"
                  >
                    {{ m._myLastComment }}
                  </div>

                  <div
                    v-if="m._myReactions && m._myReactions.length"
                    class="my-reactions"
                    role="group"
                    aria-label="Your reactions"
                  >
                    <span class="my-reactions__label">You:</span>
                    <span v-for="emoji in m._myReactions" :key="emoji" class="my-reactions__pill">
                      {{ emoji }}
                    </span>
                  </div>

                <!-- Message actions: reply, forward, react, delete. -->
                  <div class="actions">
                    <button
                      class="icon-btn"
                      type="button"
                      aria-label="Reply to message"
                      title="Reply"
                      @click="setReplyTarget(m)"
                    >
                      ↩️
                    </button>
                    <button
                      class="icon-btn"
                      type="button"
                      aria-label="Forward message"
                      title="Forward"
                      @click="openForwardPicker(m)"
                    >
                      🔗
                    </button>
                    <button
                      class="icon-btn"
                      type="button"
                      :class="{ active: m._myReactions.includes('👍') }"
                      :aria-pressed="m._myReactions.includes('👍')"
                      aria-label="Toggle thumbs up reaction"
                      @click="toggleReaction(m, '👍')"
                    >
                      👍
                    </button>

                    <button
                      class="icon-btn"
                      type="button"
                      :class="{ active: m._myReactions.includes('❤️') }"
                      :aria-pressed="m._myReactions.includes('❤️')"
                      aria-label="Toggle heart reaction"
                      @click="toggleReaction(m, '❤️')"
                    >
                      ❤️
                    </button>

                    <button
                      class="icon-btn"
                      type="button"
                      :class="{ active: m._myReactions.includes('😂') }"
                      :aria-pressed="m._myReactions.includes('😂')"
                      aria-label="Toggle laugh reaction"
                      @click="toggleReaction(m, '😂')"
                    >
                      😂
                    </button>

                    <button
                      v-if="isMine(m)"
                      class="icon-btn"
                      :disabled="deletingMessageId === String(m.id)"
                      type="button"
                      title="Delete message"
                      aria-label="Delete message"
                      @click="confirmDeleteMessage(m)"
                    >
                      {{ deletingMessageId === String(m.id) ? '⌛' : '🗑️' }}
                    </button>
                  </div>
                </div>

                <!-- Right avatar when the current user sent the message. -->
                <div
                  v-if="isMine(m)"
                  class="avatar mine"
                  :class="{ placeholder: !myAvatar }"
                  :style="avatarBg(myAvatar)"
                >
                  <span v-if="!myAvatar" class="avatar-initial">{{ avatarInitial(m) }}</span>
                </div>
              </div>
            </template>
            <div v-else class="empty-thread" role="status" aria-live="polite">
              <h2>No messages yet</h2>
              <p class="muted">
                {{ isGroup ? 'Say hello 👋 so everyone can join in.' : 'Say hello 👋 to start the conversation.' }}
              </p>
            </div>
          </div>

          <!-- Composer for text and image messages. -->
          <div class="composer">

            <div v-if="replyTarget" class="reply-banner">
              Replying to {{ nameForSender(replyTarget.senderId, replyTarget) || 'message' }}:
              <span class="reply-snippet">
                {{
                  replyTarget.content ||
                    replyTarget._replyPreview ||
                    (replyTarget.type === 'image' ? '[image]' : 'message')
                }}
              </span>

              <button class="btn-xs btn-secondary" type="button" @click="clearReplyTarget">Cancel</button>
            </div>
            <div v-if="imagePreview" class="attach-preview" aria-label="Selected image preview">
              <img :src="imagePreview" alt="Selected upload" class="attach-thumb" />
              <div class="attach-meta">
                <div class="attach-name">{{ imageFile?.name || 'Image' }}</div>
                <button class="btn-xs btn-secondary" type="button" @click="clearImageSelection">
                  Remove
                </button>
              </div>
            </div>
            <div class="composer-row">
              <textarea
                v-model="draft"
                ref="composerInput"
                class="input"
                placeholder="Type a message…"
                aria-label="Message input"
                rows="1"
                @keyup.enter.exact.prevent="onSend"
              ></textarea>

              <!-- Attach button for images -->
              <button
                type="button"
                class="icon-btn attach"
                aria-label="Attach image"
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

              <button class="btn" type="button" aria-label="Send message" :disabled="!canSend" @click="onSend">
                {{ sending ? 'Sending…' : 'Send' }}
              </button>
            </div>
          </div>

          <!-- Forward picker for choosing a destination chat. -->
          <div v-if="forwardPanelOpen" class="forward-overlay">
            <div class="forward-modal">
              <header class="forward-header">
                <div>
                  <strong>Forward message</strong>
                  <div class="muted small">Select a chat to forward this message.</div>
                </div>
                <button
                  class="close-btn"
                  type="button"
                  aria-label="Close forward picker"
                  @click="closeForwardPicker"
                >
                  ✕
                </button>
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
                      :class="{ placeholder: !avatarForConversation(c, meId) }"
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
                <div class="forward-new">
                  <div class="forward-new__title">Forward to a new user</div>
                  <UserSearch
                    placeholder="Search users"
                    @select="forwardToUser"
                    @error="forwardError = $event || ''"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

          <aside v-if="isGroup" class="group-panel">
            <div class="group-card">
              <div class="group-header">
                <div
                  class="group-avatar"
                  :class="{ placeholder: !groupInfo?.avatar }"
                  :style="groupInfo?.avatar ? { backgroundImage: `url('${groupInfo.avatar}')` } : {}"
                >
                  <span v-if="!groupInfo?.avatar">{{ (groupInfo?.name || headerTitle)[0] || 'G' }}</span>
                </div>
                <div class="group-meta">
                  <div class="group-name">{{ groupInfo?.name || headerTitle }}</div>
                  <div class="group-sub">Conversation ID: {{ convId }}</div>
                  <div class="group-sub">Members: {{ groupMembers.length }}</div>
                </div>
                <button class="link" type="button" @click="triggerGroupPhoto">Change photo</button>
                <input
                  ref="groupPhotoInput"
                  type="file"
                  class="filepick"
                  accept="image/*"
                  @change="onPickGroupPhoto"
                />
              </div>

              <div class="field inline">
                <input
                  v-model.trim="groupNameDraft"
                  class="input"
                  placeholder="Group name"
                  :disabled="groupBusy"
                />
                <button class="btn sm" type="button" :disabled="groupBusy || !groupNameDraft" @click="onRenameGroup">
                  {{ groupBusy ? 'Saving…' : 'Save' }}
                </button>
              </div>

              <div class="members-block">
                <div class="members-title">Members ({{ groupMembers.length }})</div>
                <p v-if="groupLoading" class="muted">Loading group info…</p>
                <ErrorMsg v-else-if="groupErr" :text="groupErr" />
                <div v-else class="member-scroll" role="list">
                  <ul class="member-list">
                    <li v-for="u in groupMembers" :key="u.id" class="member-item" role="listitem">
                      <div class="member-left">
                        <div
                          class="member-avatar"
                          :class="{ placeholder: !u.avatar }"
                          :style="u.avatar ? { backgroundImage: `url('${u.avatar}')` } : {}"
                        >
                          <span v-if="!u.avatar">{{ (u.name || 'User')[0] || 'U' }}</span>
                        </div>
                        <div class="member-info">
                          <div class="member-name">{{ u.name || 'User' }}</div>
                          <div class="member-sub">ID: {{ u.id }}</div>
                        </div>
                      </div>
                      <button
                        v-if="String(u.id) !== meId"
                        class="link danger"
                        type="button"
                        :disabled="groupBusy"
                        @click="onRemoveMember(u.id)"
                      >
                        Remove
                      </button>
                    </li>
                    <li v-if="!groupMembers.length" class="muted">No members found.</li>
                  </ul>
                </div>
              </div>
              <div class="members-add">
                <div class="members-title">Add members</div>
                <UserSearch
                  placeholder="Search users to add"
                  :class="{ disabled: addingMember }"
                  @select="onSelectNewMember"
                  @error="groupErr = $event || ''"
                />
              </div>

              <button class="btn danger leave" type="button" :disabled="groupBusy" @click="onLeaveGroup">
                Leave group
              </button>
              <p v-if="groupNotice" class="notice small" role="status" aria-live="polite">{{ groupNotice }}</p>
            </div>
          </aside>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import UserSearch from '@/components/UserSearch.vue'
import {
  getMyConversations,
  getConversationMembers,
  getMessages,
  getGroupsList,
  getGroupDetail,
  setGroupName,
  setGroupPhoto,
  addToGroup,
  removeFromGroup,
  leaveGroup,
  sendMessage,
  sendImageMessage,
  startConversation,
  getAvatarUrl,
  absUrl,
  deleteMessage,
  commentMessage,
  uncommentMessage,
  forwardMessage,
  isAbortError,
  ticksFor,
  titleForConversation,
  avatarForConversation,
  normalizeUser,
} from '@/services/api'

import { ensureAuthReady, isAuthenticated, currentUser } from '@/services/auth'

const route = useRoute()
const router = useRouter()

// Core reactive state used across the page.
const convId = computed(() => String(route.params.id || ''))
const err = ref('')
const notice = ref('')
const loading = ref(false)
const sending = ref(false)
const draft = ref('')
const messages = ref([])
const scrollbox = ref(null)
const imageInput = ref(null)
const imageFile = ref(null)
const imagePreview = ref('')
const replyTarget = ref(null)
const composerInput = ref(null)
const replyHighlightId = ref('')
const deletingMessageId = ref('')

// Conversation rail data and filtering.
const convList = ref([])
const convLoading = ref(false)
const convErr = ref('')
const convSearch = ref('')

const me = computed(() => currentUser.value)
const meId = computed(() => String(me.value?.id || ''))
const myAvatar = computed(() => getAvatarUrl(me.value || {}))

const currentConv = ref(null)
const isGroup = ref(false)
const groupInfo = ref(null)
const groupLoading = ref(false)
const groupBusy = ref(false)
const groupErr = ref('')
const groupNotice = ref('')
const groupNameDraft = ref('')
const addingMember = ref(false)
const groupPhotoInput = ref(null)
const groupMembers = computed(() => {
  const src = groupInfo.value?.members?.length ? groupInfo.value.members : participants.value
  return normalizeMembers(src)
})
const groupId = computed(() => {
  if (!isGroup.value) return ''
  return (
    groupInfo.value?.id ||
    currentConv.value?.groupId ||
    currentConv.value?.group_id ||
    currentConv.value?.group?.id ||
    ''
  )
})
const participants = computed(() => currentConv.value?.participants || [])
const peer = computed(
  () => participants.value.find(u => String(u.id) !== meId.value) || null
)

// Forward modal state.
const forwardPanelOpen = ref(false)
const forwardLoading = ref(false)
const forwardError = ref('')
const forwardSearch = ref('')
const forwardList = ref([])
const forwardTargetMessage = ref(null)

// Conversation rail helpers.
function convTime(t) {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric' })
}

function normalizeConversationList(items = []) {
  const pickLastMessage = (c) => {
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

  return (items || [])
    .map((c) => {
      const isGroupType =
        c?.type === 'group' ||
        !!(c?.groupId || c?.group_id || c?.group?.id)

      const last = pickLastMessage(c) || {}
      const previewContent =
        last.type === 'image'
          ? '[Image]'
          : last.type === 'file'
          ? '[File]'
          : last.content || last.text || last.body || last.message || last.Content || last.Text || last.Body || ''

      const time =
        last.createdAt ||
        last.created_at ||
        last.CreatedAt ||
        last.timestamp ||
        last.Timestamp ||
        c.updatedAt ||
        c.updated_at ||
        c.UpdatedAt ||
        c.createdAt ||
        c.created_at ||
        c.CreatedAt ||
        null

      return {
        ...c,
        type: isGroupType ? 'group' : c?.type,
        last_preview: previewContent || 'No messages yet',
        last_time: time,
      }
    })
    .sort((a, b) => new Date(b.last_time || 0) - new Date(a.last_time || 0))
}

function normalizeParticipantList(list = []) {
  return (list || [])
    .map(normalizeUser)
    .map(u => ({ ...u, avatar: u.avatarUrl || u.avatar }))
    .filter(u => u.id)
}

async function loadConversationList() {
  convErr.value = ''
  await ensureAuthReady()
  if (!isAuthenticated.value) {
    convList.value = []
    convLoading.value = false
    return
  }

  convLoading.value = true
  try {
    const raw = await getMyConversations()
    const items = raw?.data?.items || raw?.items || (Array.isArray(raw) ? raw : []) || []
    convList.value = normalizeConversationList(items)
  } catch (e) {
    convErr.value = e?.response?.data?.message || e?.message || 'Failed to load conversations'
  } finally {
    convLoading.value = false
  }
}

const filteredConversations = computed(() => {
  const q = convSearch.value.trim().toLowerCase()
  if (!q) return convList.value
  return convList.value.filter((c) => {
    const title = titleForConversation(c, meId.value).toLowerCase()
    const preview = (c.last_preview || '').toLowerCase()
    return title.includes(q) || preview.includes(q)
  })
})

function switchConversation(c) {
  if (!c || String(c.id) === convId.value) return
  router.push({ name: 'chat', params: { type: c.type === 'group' ? 'group' : 'conv', id: c.id } })
}


// Sender resolution helpers to keep author data accurate per userId.
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

  const senderTag = senderRaw.tag || senderRaw.role || senderRaw.label || senderRaw.title || ''

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
        avatar: normalized.avatarUrl,
        tag: senderTag || normalized.tag,
      }
    : { id: '', name: normalized.name || senderName, avatar: normalized.avatarUrl, tag: senderTag }
}

function resolveSender(userId, msg = null) {
  const id = String(userId || msg?.senderId || '')
  if (!id) return null

  const fromParticipants = participants.value.find(u => String(u.id) === id)
  if (fromParticipants) return fromParticipants

  if (msg?._sender && String(msg._sender.id) === id) return msg._sender

  const fromMessages = messages.value.find(m => String(m.senderId) === id && m._sender)
  if (fromMessages) return fromMessages._sender

  return null
}

// Avatar helpers.
function avatarFor(m) {
  if (isMine(m)) return myAvatar.value

  const sender = resolveSender(m.senderId, m)
  if (sender?.avatar) return sender.avatar
  if (sender) return getAvatarUrl(sender)

  return ''
}

function avatarInitial(m) {
  if (isMine(m)) return (me.value?.name || 'Me')[0] || 'M'
  const sender = resolveSender(m.senderId, m)
  const name = sender?.name || sender?.username || sender?.tag || sender?.id || 'U'
  return (name[0] || 'U').toUpperCase()
}

function avatarBg(src) {
  return src ? { backgroundImage: `url('${src}')` } : {}
}

function displayNameFor(m) {
  const s = resolveSender(m.senderId, m)
  if (s?.name) return s.name
  if (s?.username) return s.username
  if (s?.tag) return s.tag
  return s?.id || String(m.senderId || '')
}
function nameForSender(userId, msg = null) {
  if (!userId && !msg?.senderId) return ''
  if (String(userId || msg?.senderId) === meId.value) return me.value?.name || 'Me'

  const s = resolveSender(userId, msg)
  if (s?.name) return s.name
  if (s?.username) return s.username
  if (s?.tag) return s.tag
  return s?.id || String(userId || msg?.senderId || '')
}

// Header avatar and title helpers.
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

// Keep the header avatar/title in sync when conversation metadata refreshes.
watch(convList, () => {
  const fresh = convList.value.find(c => String(c.id) === convId.value)
  if (fresh) {
    currentConv.value = { ...(currentConv.value || {}), ...fresh }
  }
})

// Load conversation metadata (for title, participants, and type).
async function loadConversationMeta() {
  await ensureAuthReady()
  if (!isAuthenticated.value) return

  try {
    if (!convList.value.length) {
      await loadConversationList()
    }

    let conv = convList.value.find(c => String(c.id) === convId.value)

    if (!conv) {
      const raw = await getMyConversations()
      // Accept payloads shaped as {code,data:{items}}, {items}, or a raw array
      const items =
        raw?.data?.items ||
        raw?.items ||
        (Array.isArray(raw) ? raw : []) ||
        []

      convList.value = normalizeConversationList(items)
      conv = convList.value.find(c => String(c.id) === convId.value)
    }

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
            conv = { ...conv, participants: normalizeParticipantList(memberList) }
          }
        } catch (memberErr) {
          if (!isAbortError(memberErr)) {
            console.error('loadConversationMeta members fallback failed', memberErr)
          }
        }
      }

      const normalizedParticipants = normalizeParticipantList(conv.participants || [])
      if (normalizedParticipants.length) {
        conv = { ...conv, participants: normalizedParticipants }
      }
    }

    currentConv.value = conv || null
    isGroup.value =
      conv?.type === 'group' || !!(conv?.groupId || conv?.group_id || conv?.group?.id)
    if (isGroup.value) {
      await loadGroupPanel()
    } else {
      groupInfo.value = null
    }
  } catch (e) {
    if (!isAbortError(e)) {
      // Let message loading continue; surface diagnostics in the console only
      console.error('loadConversationMeta failed', e)
    }
  }
}

// Format ISO timestamps for UI display.
function fmtTime(ts) {
  if (!ts) return ''
  return ts.replace('T', ' ').slice(0, 19)
}

// Determine whether the message belongs to the current user.
function isMine(m) {
  return String(m.senderId) === meId.value
}

function looksLikeFileUrl(val) {
  if (!val) return false
  return /^https?:\/\//i.test(val) || String(val).startsWith('/')
}

// Normalize heterogeneous message payloads from the API.
function normalizeMessage(raw) {
  const senderProfile = senderProfileFromRaw(raw)
  const senderId = senderProfile.id
  const ts = raw.createdAt || raw.created_at || new Date().toISOString()
  const replyToId = raw.replyToId || raw.reply_to_id || null

  const replyContent = raw.replyTo?.content || raw.replyTo?.text || ''
  const replyType = raw.replyTo?.type || ''
  const replyPreview = replyContent || (replyType === 'image' ? '[image]' : '')

  const read = Boolean(raw.read || raw.read_at || raw.readAt)

  // Image handling: when type === 'image', content is the image URL
  const possibleFileRel =
    raw.fileUrl ||
    raw.file_url ||
    raw.imageUrl ||
    raw.image_url ||
    raw.file

  const contentIsUrl = raw.type === 'image' && looksLikeFileUrl(raw.content)
  const fileRel = possibleFileRel || (contentIsUrl ? raw.content : null)

  const commentList =
    raw.comments ||
    raw.Comments ||
    raw.replies ||
    raw.Replies ||
    []

  const byMe = commentList.filter(c =>
    String(c?.senderId || c?.userId || c?.user_id || '') === meId.value
  )

  const myEmojis = byMe
    .filter(c => (c?.type || '').toLowerCase() === 'emoji' && c?.content)
    .map(c => c.content)

  const myLastComment = [...byMe]
    .reverse()
    .find(c => (c?.type || '').toLowerCase() !== 'emoji' && (c?.content || c?.text))

  const replySenderProfile = senderProfileFromRaw({ sender: raw.replyTo?.sender || raw.reply_to?.sender || {} })

  const contentText =
    raw.type === 'image'
      ? raw.caption || raw.text || (contentIsUrl ? '' : raw.content || '')
      : raw.content || raw.text || ''

  return {
    id: raw.id,
    read,
    content: contentText,
    type: raw.type === 'image' ? 'image' : 'text',
    fileAbsUrl: fileRel ? absUrl(fileRel) : null,
    senderId: String(senderId || ''),
      _sender: senderProfile.id
        ? {
            id: senderProfile.id,
            name: senderProfile.name,
            username: senderProfile.username,
            avatar: senderProfile.avatar,
            tag: senderProfile.tag,
          }
        : null,
    _ts: new Date(ts).toISOString(),
    replyToId: replyToId ? String(replyToId) : '',
    _showCommentBox: false,
    _commentDraft: '',
    _myLastComment:
      (myLastComment?.content || myLastComment?.text || '').trim(),
    _myReactions: Array.from(new Set(myEmojis)),
    _replyPreview: replyPreview,
    _replyFrom: replySenderProfile.name || replySenderProfile.tag || replySenderProfile.id || '',
  }
}

// Load messages for the active conversation and hydrate reply previews.
async function loadMessages() {
  await ensureAuthReady()
  if (!isAuthenticated.value) return

  loading.value = true
  err.value = ''
  notice.value = ''
  try {
    const data = await getMessages({ conversationId: convId.value, limit: 150 })

    // Accept message envelopes: {code,data:{messages}}, {messages}, or a raw array
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
          m._replyFrom = nameForSender(target.senderId, target)
        }
      }

      if (m._replyPreview && !m._replyFrom && m.replyToId) {
        const target = byId.get(String(m.replyToId))
        if (target) {
          m._replyFrom = nameForSender(target.senderId, target)
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

function normalizeMembers(list) {
  return normalizeParticipantList(list).map(u => ({
    id: u.id,
    name: u.name || u.username || u.id,
    avatar: u.avatar,
    username: u.username,
  }))
}

async function resolveGroupId() {
  if (!isGroup.value) return ''
  if (groupId.value) return groupId.value

  try {
    const list = await getGroupsList()
    const arr = Array.isArray(list) ? list : (list?.items ?? list?.groups ?? list?.list ?? [])
    const hit = (arr || []).find(
      g => String(g?.conversationId ?? g?.conversation_id ?? '') === convId.value
    )
    if (hit?.id || hit?.group_id) {
      groupInfo.value = {
        id: String(hit.id ?? hit.group_id ?? ''),
        name: hit.name || '',
        avatar: getAvatarUrl(hit || {}),
        members: normalizeMembers(hit.members || []),
      }
      return groupInfo.value.id
    }
  } catch (e) {
    if (!isAbortError(e)) console.error('resolveGroupId failed', e)
  }
  return ''
}

async function loadGroupPanel() {
  if (!isGroup.value) return
  groupLoading.value = true
  groupErr.value = ''
  groupNotice.value = ''

  try {
    const gid = groupId.value || (await resolveGroupId())
    if (!gid) {
      groupErr.value = 'Group info unavailable.'
      return
    }

    const rawDetail = await getGroupDetail(gid)
    const detail = rawDetail?.group || rawDetail || {}
    const membersRaw =
      detail?.members ||
      detail?.participants ||
      detail?.data?.members ||
      detail?.data?.items ||
      detail?.group?.members ||
      []

    const members = normalizeMembers(membersRaw)

    if (membersRaw?.length) {
      currentConv.value = { ...(currentConv.value || {}), participants: normalizeParticipantList(membersRaw) }
    }

    groupInfo.value = {
      id: String(detail?.id ?? detail?.group_id ?? gid),
      name: detail?.name || detail?.title || detail?.group?.name || headerTitle.value,
      avatar: getAvatarUrl(detail || detail?.group || {}),
      members,
    }

    currentConv.value = {
      ...(currentConv.value || {}),
      name: groupInfo.value.name,
      avatar: groupInfo.value.avatar,
      participants: membersRaw?.length
        ? normalizeParticipantList(membersRaw)
        : currentConv.value?.participants || [],
    }

    currentConv.value = {
      ...(currentConv.value || {}),
      name: groupInfo.value.name,
      avatar: groupInfo.value.avatar,
      participants: membersRaw?.length
        ? normalizeParticipantList(membersRaw)
        : currentConv.value?.participants || [],
    }

    groupNameDraft.value = groupInfo.value.name || ''
  } catch (e) {
    groupErr.value = e?.response?.data?.message || e?.message || 'Failed to load group info'
  } finally {
    groupLoading.value = false
  }
}

// Send text messages and handle deletion.
const canSend = computed(() => (!!draft.value.trim() || !!imageFile.value) && !sending.value)

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
  const file = imageFile.value
  if (!t && !file) return

  sending.value = true
  err.value = ''
  notice.value = ''
  try {
    let sent
    if (file) {
      sent = await sendImageMessage({
        conversationId: convId.value,
        file,
        caption: t,
        replyToId: replyTarget.value?.id,
      })
    } else {
      sent = await sendMessage({
        conversationId: convId.value,
        content: t,
        replyToId: replyTarget.value?.id,
      })
    }

    const normalized = normalizeMessage(sent?.message || sent || {})
    normalized._localStatus = 1
    messages.value = [...messages.value, normalized]
    await nextTick()
    if (scrollbox.value) {
      scrollbox.value.scrollTop = scrollbox.value.scrollHeight
    }

    draft.value = ''
    replyTarget.value = null
    clearImageSelection()
    await loadMessages()
    bumpConversationList({ lastPreview: normalized.content || forwardPreview(normalized) })
  } catch (e) {
    err.value = e?.response?.data?.message || 'Failed to send'
  } finally {
    sending.value = false
  }
}

// Send image messages via the file picker.
function triggerImagePicker() {
  if (imageInput.value) {
    imageInput.value.click()
  }
}

async function onPickImage(e) {
  const file = e.target.files?.[0]
  if (!file) return

  imageFile.value = file
  imagePreview.value = URL.createObjectURL(file)
  if (imageInput.value) {
    imageInput.value.value = ''
  }
}

function clearImageSelection() {
  imageFile.value = null
  imagePreview.value = ''
  if (imageInput.value) {
    imageInput.value.value = ''
  }
}

// Toggle emoji reactions using the comment/uncomment endpoints.
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
    const preview = forwardPreview(forwardTargetMessage.value)
    window.dispatchEvent(
      new CustomEvent('conversations:refresh', {
        detail: {
          conversationId: String(targetConvId),
          lastTime: new Date().toISOString(),
          lastPreview: preview,
        },
      })
    )
    closeForwardPicker()
  } catch (e) {
    forwardError.value =
      e?.response?.data?.message || e?.message || 'Failed to forward message'
  }
}

async function forwardToUser(user) {
  const userId = String(user?.id || user?.userId || user?.user_id || '')
  if (!userId || !forwardTargetMessage.value) return

  forwardError.value = ''
  try {
    const res = await startConversation({ memberIds: [userId] })
    const cid =
      res?.conversationId ||
      res?.conversation_id ||
      res?.conversation?.id ||
      res?.id ||
      res?._id

    if (!cid) throw new Error('Failed to create conversation')
    await forwardToConversation(String(cid))
  } catch (e) {
    forwardError.value = e?.response?.data?.message || e?.message || 'Failed to forward message'
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

function forwardPreview(m) {
  if (!m) return ''
  if (m.type === 'image') return '[Image]'
  if (m.type === 'file') return '[File]'
  if (typeof m.content === 'string' && m.content.trim()) return m.content
  if (typeof m.body === 'string' && m.body.trim()) return m.body
  if (typeof m.message === 'string' && m.message.trim()) return m.message
  if (typeof m.text === 'string' && m.text.trim()) return m.text
  return '[Forwarded message]'
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
  if (!isMine(m)) return ''
  if (m.read) return '✓✓'
  const v = Math.max(ticksFor(m, meId.value), m?._localStatus || 0)
  if (v >= 2) return '✓✓'
  return '✓'
}

// Group management helpers.
function triggerGroupPhoto() {
  if (groupPhotoInput.value) groupPhotoInput.value.click()
}

async function onPickGroupPhoto(e) {
  const file = e?.target?.files?.[0]
  if (e?.target) e.target.value = ''
  if (!file || !groupId.value) return

  groupErr.value = ''
  groupNotice.value = ''
  groupBusy.value = true
  try {
    await setGroupPhoto(groupId.value, file)
    groupNotice.value = 'Group photo updated.'
    await loadGroupPanel()
    bumpConversationList()
  } catch (er) {
    groupErr.value = er?.response?.data?.message || er?.message || 'Failed to update group photo'
  } finally {
    groupBusy.value = false
  }
}

async function onRenameGroup() {
  if (!groupId.value) return
  if (!groupNameDraft.value) { groupErr.value = 'Name is required.'; return }
  groupBusy.value = true
  groupErr.value = ''
  groupNotice.value = ''
  try {
    await setGroupName(groupId.value, groupNameDraft.value)
    currentConv.value = { ...(currentConv.value || {}), name: groupNameDraft.value }
    groupNotice.value = 'Group name saved.'
    await loadGroupPanel()
    bumpConversationList()
  } catch (e) {
    groupErr.value = e?.response?.data?.message || e?.message || 'Failed to rename group'
  } finally {
    groupBusy.value = false
  }
}

async function onSelectNewMember(user) {
  const userId = String(user?.id || user?.userId || user?.user_id || '')
  if (!groupId.value || !userId) return
  addingMember.value = true
  groupErr.value = ''
  groupNotice.value = ''
  try {
    await addToGroup(groupId.value, [userId])
    groupNotice.value = 'Member added.'
    await loadGroupPanel()
  } catch (e) {
    groupErr.value = e?.response?.data?.message || e?.message || 'Failed to add member'
  } finally {
    addingMember.value = false
  }
}

async function onRemoveMember(userId) {
  if (!groupId.value || !userId) return
  if (!confirm('Remove this member from the group?')) return
  groupBusy.value = true
  groupErr.value = ''
  groupNotice.value = ''
  try {
    await removeFromGroup(groupId.value, userId)
    groupNotice.value = 'Member removed.'
    await loadGroupPanel()
  } catch (e) {
    groupErr.value = e?.response?.data?.message || e?.message || 'Failed to remove member'
  } finally {
    groupBusy.value = false
  }
}

async function onLeaveGroup() {
  if (!groupId.value) return
  if (!confirm('Leave this group?')) return
  groupBusy.value = true
  groupErr.value = ''
  groupNotice.value = ''
  try {
    await leaveGroup(groupId.value)
    router.push('/conversations')
  } catch (e) {
    groupErr.value = e?.response?.data?.message || e?.message || 'Failed to leave group'
  } finally {
    groupBusy.value = false
  }
}

function bumpConversationList(extra = {}) {
  if (!convId.value) return
  window.dispatchEvent(
    new CustomEvent('conversations:refresh', {
      detail: {
        conversationId: convId.value,
        lastTime: new Date().toISOString(),
        ...(extra || {}),
      },
    })
  )
}

// Initialize the view once authentication is confirmed.
async function bootstrap() {
  await ensureAuthReady()
  if (!isAuthenticated.value) {
    return router.replace('/login')
  }

  await loadConversationList()
  await loadConversationMeta()
  await loadMessages()
}

onMounted(() => {
  window.addEventListener('conversations:refresh', loadConversationList)
  bootstrap()
})

onBeforeUnmount(() => {
  window.removeEventListener('conversations:refresh', loadConversationList)
})

watch(convId, async () => {
  await loadConversationList()
  await loadConversationMeta()
  await loadMessages()
})
</script>

<style scoped>
.page {
  --avatar-bg: #e0f7ee;
  --avatar-border: #a7f3d0;
  --avatar-text: #0f766e;
  min-height: 100%;
  height: 100%;
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  background: #e5e7eb;
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
  justify-content: flex-start;
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
  flex: 1 1 auto;
  min-width: 0;
  text-align: center;
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
  background: var(--avatar-bg);
  background-size: cover;
  background-position: center;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: var(--avatar-text);
  border: 1px solid var(--avatar-border);
}

.muted {
  font-size: 0.75rem;
  color: #64748b;
}

.content {
  width: 100%;
  padding: 12px 12px 16px;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
}

.content-inner {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1 1 auto;
  min-height: 0;
}

.chat-layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 12px;
  align-items: stretch;
  flex: 1 1 auto;
  min-height: 0;
}

.chat-layout.has-group {
  grid-template-columns: 320px minmax(0, 1fr) 300px;
}

.conv-rail {
  background: #f8fafc;
  border-right: 1px solid #d9dde3;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
  min-height: 0;
}

.conv-rail__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.conv-rail__title {
  font-weight: 800;
  color: #0f172a;
}

.conv-rail__hint {
  color: #64748b;
  font-size: 0.9rem;
}

.conv-rail__badge {
  background: #e0f2fe;
  color: #0f172a;
  border-radius: 0;
  padding: 4px 10px;
  font-weight: 700;
}

.conv-rail__search input {
  width: 100%;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-control);
  padding: 10px 12px;
  font-size: 0.95rem;
}

.conv-rail__list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  overflow: auto;
  padding-right: 2px;
}

.conv-pill {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 0;
  padding: 12px;
  background: #f8fafc;
  cursor: pointer;
  transition: border 0.2s, background 0.2s;
}

.conv-pill:hover {
  border-color: #c7d2fe;
  background: #eef2ff;
}

.conv-pill.active {
  border-color: #3b82f6;
  background: #e0f2fe;
}

.conv-pill__avatar {
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background: #e2e8f0;
  background-size: cover;
  background-position: center;
  display: grid;
  place-items: center;
  font-weight: 700;
  color: #475569;
  border: 1px solid #d1d5db;
}

.conv-pill__avatar.placeholder {
  background: #e0f7ee;
  color: #0f766e;
  border-color: #a7f3d0;
}

.conv-pill__meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.conv-pill__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 6px;
}

.conv-pill__name {
  font-weight: 700;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-pill__time {
  color: #94a3b8;
  font-size: 0.85rem;
}

.conv-pill__bottom {
  color: #64748b;
  font-size: 0.92rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-pill__type {
  font-size: 0.85rem;
  color: #0f172a;
  background: #e2e8f0;
  border-radius: 0;
  padding: 4px 8px;
  font-weight: 600;
}

.conv-pill__type.group {
  background: #ecfdf3;
  color: #166534;
}

.chat-main {
  background: #fff;
  border: 1px solid #d9dde3;
  border-radius: 0;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.group-panel {
  position: sticky;
  top: 68px;
}

.group-card {
  background: #fff;
  border: 1px solid #d9dde3;
  border-radius: 0;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: calc(100vh - 90px);
  overflow: auto;
}

.group-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 10px;
  align-items: center;
}

.group-avatar {
  width: 54px;
  height: 54px;
  border-radius: 50%;
  background: var(--avatar-bg);
  border: 1px solid var(--avatar-border);
  display: grid;
  place-items: center;
  background-size: cover;
  background-position: center;
  color: var(--avatar-text);
  font-weight: 700;
}

.group-avatar.placeholder {
  border: 1px solid var(--avatar-border);
}

.group-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.group-name {
  font-weight: 700;
  color: #0f172a;
}

.group-sub {
  color: #64748b;
  font-size: 0.85rem;
}

.field.inline {
  display: flex;
  gap: 8px;
  align-items: center;
}

.field .input {
  flex: 1;
}

.members-block,
.members-add {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0;
  padding: 10px;
  display: grid;
  gap: 8px;
}

.member-scroll {
  max-height: 260px;
  overflow: auto;
  padding-right: 4px;
}

.members-add .disabled {
  opacity: 0.6;
  pointer-events: none;
}

.members-title {
  font-weight: 700;
  color: #0f172a;
}

.member-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 8px;
}

.member-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border: 1px solid #e2e8f0;
  border-radius: 0;
  padding: 8px;
  background: #f8fafc;
}

.member-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.member-avatar {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  background: var(--avatar-bg);
  border: 1px solid var(--avatar-border);
  display: grid;
  place-items: center;
  background-size: cover;
  background-position: center;
  font-weight: 700;
  color: var(--avatar-text);
}

.member-avatar.placeholder {
  border: 1px solid var(--avatar-border);
}

.member-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.members-add :deep(.list) {
  max-height: 220px;
  overflow: auto;
  padding-right: 4px;
}

.member-name {
  font-weight: 600;
  color: #0f172a;
}

.member-sub {
  color: #64748b;
  font-size: 0.85rem;
}

.link {
  background: none;
  border: none;
  color: #2563eb;
  cursor: pointer;
  padding: 0;
}

.link.danger {
  color: #dc2626;
}

.notice {
  background: #ecfeff;
  color: #0f766e;
  border: 1px solid #99f6e4;
  border-radius: 0;
  padding: 8px 10px;
  margin-bottom: 8px;
}

.scroll {
  height: 65vh;
  overflow-y: auto;
  background: #f8fafc;
  border-radius: 0;
  padding: 12px;
  border: 1px solid #e1e5eb;
}

.empty-thread {
  min-height: 50vh;
  display: grid;
  place-items: center;
  text-align: center;
  gap: 8px;
  color: #0f172a;
}

.empty-thread h2 {
  margin: 0;
  font-size: 1.2rem;
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
  width: 42px;
  height: 42px;
  border-radius: 50%;
  overflow: hidden;

  flex-shrink: 0;
  display: grid;
  place-items: center;
  background: var(--avatar-bg);
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  border: 1px solid var(--avatar-border);
}

.avatar.mine {
  margin-left: 4px;    /* Keep a small gap between my avatar and bubble */
}
.avatar.placeholder {
  background: var(--avatar-bg);
  border: 1px solid var(--avatar-border);
  color: var(--avatar-text);
  font-weight: 700;
}

.avatar-initial {
  font-size: 0.9rem;
}

.bubble-wrap {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
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
  display: inline-flex;
  flex-direction: column;
  padding: 8px 12px;
  border-radius: var(--radius-bubble);
  background: #ffffff;
  width: fit-content;
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
  border-radius: var(--radius-control);
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

.img-wrap {
  display: inline-flex;
  margin-bottom: 6px;
}

.img {
  max-width: 260px;
  border-radius: var(--radius-bubble);
}

.text-block {
  white-space: pre-wrap;
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
  border-radius: 0;
  padding: 4px 12px;
  font-size: 0.82rem;
  max-width: 80%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.comment-chip.mine {
  background: #d9f7be;
}

.my-reactions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 6px;
}

.my-reactions__label {
  color: #475569;
  font-size: 0.82rem;
}

.my-reactions__pill {
  border-radius: 0;
  background: #e0f2fe;
  border: 1px solid #bfdbfe;
  padding: 2px 8px;
  font-size: 0.9rem;
  line-height: 1.2;
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
  border-radius: var(--radius-control);
}

.icon-btn:focus-visible,
.btn:focus-visible,
.btn-xs:focus-visible,
.link:focus-visible,
.close-btn:focus-visible,
.forward-item:focus-visible {
  outline: 2px solid #2563eb;
  outline-offset: 2px;
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
  border-radius: var(--radius-control);
  padding: 5px 8px;
}

.btn-xs {
  border: none;
  padding: 5px 12px;
  border-radius: var(--radius-control);
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
  border-radius: 0;
  border: 1px solid #e1e5eb;
}

.attach-preview {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #f8fafc;
  border: 1px dashed #cbd5e1;
  padding: 8px;
  border-radius: var(--radius-control);
  margin-bottom: 4px;
}

.attach-thumb {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: var(--radius-bubble);
  border: 1px solid #e2e8f0;
}

.attach-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.attach-name {
  font-weight: 600;
  color: #0f172a;
}

.composer-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.input {
  flex: 1;
  padding: 10px;
  border-radius: var(--radius-control);
  border: 1px solid #cbd5e1;
  resize: none;
}

.reply-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f1f5f9;
  border: 1px solid #cbd5e1;
  border-radius: 0;
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
  border-radius: var(--radius-control);
  background: #22c55e;
  padding: 10px 16px;
  color: white;
  white-space: nowrap;
}

.btn.sm {
  padding: 8px 12px;
  font-size: 0.92rem;
}

.btn.danger {
  background: linear-gradient(135deg, #ef4444, #f97316);
}

.btn.leave {
  width: 100%;
  text-align: center;
}

.notice.small {
  font-size: 0.9rem;
  margin: 0;
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
  border-radius: 0;
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
  border-radius: var(--radius-control);
  border: 1px solid #cbd5e1;
}

.forward-body {
  overflow-y: auto;
  max-height: 55vh;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.forward-new {
  border-top: 1px solid #e2e8f0;
  padding-top: 10px;
}

.forward-new__title {
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 6px;
}

.forward-item {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 10px;
  border: 1px solid #e2e8f0;
  border-radius: 0;
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

.forward-avatar.placeholder {
  background: #e0f7ee;
  color: #0f766e;
  border: 1px solid #a7f3d0;
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
