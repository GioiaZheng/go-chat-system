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

      <div class="title-row">
        <span v-if="!headerAvatar" class="header-avatar avatar-fallback avatar-circle">{{ headerTitle[0] || 'C' }}</span>
        <img v-else class="header-avatar avatar avatar-circle" :src="headerAvatar" alt="avatar" />
        <div class="title text-title">{{ headerTitle }}</div>
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
              <div class="conv-rail__title text-title">Conversations</div>
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
                  <span
                    v-if="!avatarForConversation(c, meId)"
                    class="conv-pill__avatar avatar-fallback avatar-circle"
                  >
                    {{ titleForConversation(c, meId)[0] || 'C' }}
                  </span>
                  <img
                    v-else
                    class="conv-pill__avatar avatar avatar-circle"
                    :src="avatarForConversation(c, meId)"
                    alt="avatar"
                  />

                  <div class="conv-pill__meta">
                    <div class="conv-pill__top">
                      <span class="conv-pill__name text-primary">{{ titleForConversation(c, meId) }}</span>
                      <span class="conv-pill__time text-secondary">{{ convTime(c.last_time) }}</span>
                    </div>
                    <div class="conv-pill__bottom text-secondary">{{ c.last_preview || 'No messages yet' }}</div>
                  </div>
                </button>

                <p v-if="!filteredConversations.length" class="muted">No conversations found.</p>
              </template>
            </div>
          </aside>

          <div class="chat-main">
            <div ref="scrollbox" class="scroll" @scroll="handleScroll">
              <template v-if="messages.length">
              <div
                v-for="m in messages"
                :key="m.id"
                class="row"
                :id="`msg-${m.id}`"
                :class="{ mine: isMine(m), highlight: replyHighlightId === String(m.id) }"
              >
                <!-- Left avatar for incoming messages. -->
                <span
                  v-if="!isMine(m) && !avatarFor(m)"
                  class="avatar-fallback avatar-circle"
                >
                  {{ avatarInitial(m) }}
                </span>
                <img
                  v-else-if="!isMine(m)"
                  class="avatar avatar-circle"
                  :src="avatarFor(m)"
                  alt="avatar"
                />
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
                      <div class="reply-body">
                        <img v-if="m._replyImage" :src="m._replyImage" alt="Reply image" class="reply-thumb" />
                        <div class="reply-text">{{ m._replyPreview }}</div>
                      </div>
                    </button>

                    <div v-if="m.fileAbsUrl" class="img-wrap">
                      <img :src="m.fileAbsUrl" class="img" />
                    </div>

                    <div v-if="m.content" class="text-block text-primary">
                      {{ m.content }}
                    </div>
                  </div>
                  <!-- Timestamp and delivery markers. -->
                  <div class="meta text-secondary">
                    {{ fmtTime(m._ts) }}
                    <span v-if="tickText(m)" class="ticks">{{ tickText(m) }}</span>
                  </div>

                  <!-- Comment summary for the current user. -->
                  <div
                    v-if="m._myLastComment"
                    class="comment-chip text-secondary"
                    :class="{ mine: isMine(m) }"
                  >
                    {{ m._myLastComment }}
                  </div>

                  <div
                    v-if="m._reactions && m._reactions.length"
                    class="reactions"
                    role="group"
                    aria-label="Reactions"
                  >
                    <span v-for="emoji in m._reactions" :key="emoji" class="reaction-pill">
                      {{ emoji }}
                    </span>
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
                <span
                  v-if="isMine(m) && !myAvatar"
                  class="avatar avatar-fallback avatar-circle mine"
                >
                  {{ avatarInitial(m) }}
                </span>
                <img
                  v-else-if="isMine(m)"
                  class="avatar avatar-circle mine"
                  :src="myAvatar"
                  alt="avatar"
                />
              </div>
            </template>
            <div v-else class="empty-thread" role="status" aria-live="polite">
              <p class="muted">{{ isGroup ? 'Say hello to the group 👋' : 'Say hello 👋 to start the chat.' }}</p>
            </div>
          </div>

          <!-- Composer for text and image messages. -->
          <div class="composer">

            <div v-if="replyTarget" class="reply-banner">
              Replying to {{ nameForSender(replyTarget.senderId, replyTarget) || 'message' }}:
              <span class="reply-snippet">
                <img
                  v-if="replyTarget.fileAbsUrl"
                  :src="replyTarget.fileAbsUrl"
                  alt="Reply image"
                  class="reply-thumb"
                />
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
          <div v-if="forwardPanelOpen" class="forward-overlay" @click.self="closeForwardPicker">
            <div class="forward-modal">
              <header class="forward-header">
                <strong>Forward message</strong>
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
                    <span
                      v-if="!avatarForConversation(c, meId)"
                      class="forward-avatar avatar-fallback avatar-circle"
                    >
                      {{ titleForConversation(c, meId)[0] || 'C' }}
                    </span>
                    <img
                      v-else
                      class="forward-avatar avatar avatar-circle"
                      :src="avatarForConversation(c, meId)"
                      alt="avatar"
                    />

                    <div class="forward-meta">
                      <div class="forward-name">{{ titleForConversation(c, meId) }}</div>
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
                <span
                  v-if="!groupInfo?.avatar"
                  class="group-avatar avatar-fallback avatar-circle"
                >
                  {{ (groupInfo?.name || headerTitle)[0] || 'G' }}
                </span>
                <img
                  v-else
                  class="group-avatar avatar avatar-circle"
                  :src="groupInfo.avatar"
                  alt="avatar"
                />
                <div class="group-meta">
                  <div class="group-name">{{ groupInfo?.name || headerTitle }}</div>
                  <div class="group-sub">{{ groupMembers.length === 1 ? '1 member' : `${groupMembers.length} members` }}</div>
                </div>
                <button class="link muted" type="button" @click="triggerGroupPhoto">Change photo</button>
                <input
                  ref="groupPhotoInput"
                  type="file"
                  class="filepick"
                  accept="image/*"
                  @change="onPickGroupPhoto"
                />
              </div>

            <div class="field inline group-rename">
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
              <div class="section-label text-secondary">Members</div>
              <p v-if="groupLoading" class="muted">Loading group info…</p>
              <ErrorMsg v-else-if="groupErr" :text="groupErr" />
              <div v-else class="member-scroll" role="list">
                <ul class="member-list">
                  <li v-for="u in groupMembers" :key="u.id" class="member-item" role="listitem">
                    <div class="member-left">
                      <span
                        v-if="!u.avatar"
                        class="member-avatar avatar-fallback avatar-circle"
                      >
                        {{ initialsFor(u, 'U') }}
                      </span>
                      <img
                        v-else
                        class="member-avatar avatar avatar-circle"
                        :src="u.avatar"
                        alt="avatar"
                      />
                      <div class="member-info">
                        <div class="member-name">{{ u.name || 'Unknown' }}</div>
                        <div v-if="safeUsername(u)" class="member-sub">{{ safeUsername(u) }}</div>
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
              <div class="section-label text-secondary">Add members</div>
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
  preferredDisplayName,
  initialsFor,
  safeUsername,
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
const pinnedToBottom = ref(true)
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

function scrollToBottom(force = false) {
  return nextTick(() => {
    if (!scrollbox.value) return

    const distance =
      scrollbox.value.scrollHeight - (scrollbox.value.scrollTop + scrollbox.value.clientHeight)

    if (force || distance < 48 || pinnedToBottom.value) {
      scrollbox.value.scrollTop = scrollbox.value.scrollHeight
      pinnedToBottom.value = true
    }
  })
}

function handleScroll() {
  if (!scrollbox.value) return
  const distance = scrollbox.value.scrollHeight - (scrollbox.value.scrollTop + scrollbox.value.clientHeight)
  pinnedToBottom.value = distance < 64
}

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
        participants: normalizeParticipantList(c.participants || c.members || []),
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

async function refreshConversations() {
  await loadConversationList()
  const fresh = convList.value.find(c => String(c.id) === convId.value)
  if (fresh) {
    currentConv.value = { ...(currentConv.value || {}), ...fresh }
  }
}

function handleConversationRefresh(e) {
  const detail = e?.detail || {}
  const targetId = detail.conversationId ? String(detail.conversationId) : ''
  const bumpedTime = detail.lastTime || ''
  const bumpedPreview = detail.lastPreview
  const bumpedName = detail.name
  const bumpedAvatar = detail.avatar

  let updated = false
  if (targetId && convList.value?.length) {
    convList.value = convList.value
      .map(c => {
        if (String(c.id) !== targetId) return c
        updated = true
        return {
          ...c,
          ...(bumpedTime ? { last_time: bumpedTime } : {}),
          ...(bumpedPreview ? { last_preview: bumpedPreview } : {}),
          ...(bumpedName ? { name: bumpedName } : {}),
          ...(bumpedAvatar ? { avatar: bumpedAvatar } : {}),
        }
      })
      .sort((a, b) => new Date(b.last_time || 0) - new Date(a.last_time || 0))
  }

  if (!updated) {
    loadConversationList()
  }

  if (updated && String(currentConv.value?.id || '') === targetId) {
    currentConv.value = {
      ...(currentConv.value || {}),
      ...(bumpedName ? { name: bumpedName } : {}),
      ...(bumpedAvatar ? { avatar: bumpedAvatar } : {}),
    }
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
  if (isMine(m)) return preferredDisplayName(me.value || {})[0] || 'M'
  const sender = resolveSender(m.senderId, m)
  const name = preferredDisplayName(sender || {})
  return (name[0] || 'U').toUpperCase()
}

function displayNameFor(m) {
  const s = resolveSender(m.senderId, m)
  if (String(m.senderId) === meId.value) return preferredDisplayName(me.value || {})
  return s ? preferredDisplayName(s) : 'Unknown'
}
function nameForSender(userId, msg = null) {
  if (!userId && !msg?.senderId) return ''
  if (String(userId || msg?.senderId) === meId.value) return preferredDisplayName(me.value || {})

  const s = resolveSender(userId, msg)
  return s ? preferredDisplayName(s) : 'Unknown'
}

// Header avatar and title helpers.
const headerTitle = computed(() => {
  if (!currentConv.value) return 'Chat'
  
  const title = titleForConversation(currentConv.value, meId.value)
  if (title && title !== 'Chat') return title

  if (!isGroup.value) return preferredDisplayName(peer.value || {}) || 'Unknown'
  return preferredDisplayName(currentConv.value || {})
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

  const replyFileRel =
    raw.replyTo?.fileUrl ||
    raw.replyTo?.file_url ||
    raw.replyTo?.imageUrl ||
    raw.replyTo?.image_url ||
    raw.replyTo?.file ||
    (replyType === 'image' && looksLikeFileUrl(raw.replyTo?.content) ? raw.replyTo?.content : null)
  const replyImageAbs = replyFileRel ? absUrl(replyFileRel) : ''

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

  const senderKey = (c) =>
    String(c?.senderId || c?.userId || c?.user_id || c?.authorId || c?.author_id || '')

  const typeFor = (c) => (c?.type || c?.Type || '').toLowerCase()
  const contentFor = (c) => c?.content || c?.Content || ''

  const byMe = commentList.filter(c => senderKey(c) === meId.value)

  const myEmojis = byMe
    .filter(c => typeFor(c) === 'emoji' && contentFor(c))
    .map(c => contentFor(c))

  const allEmojis = commentList
    .filter(c => typeFor(c) === 'emoji' && contentFor(c))
    .map(c => contentFor(c))

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
    _reactions: Array.from(new Set(allEmojis)),
    _replyPreview: replyPreview,
    _replyImage: replyImageAbs,
    _replyFrom: preferredDisplayName(replySenderProfile),
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
          m._replyImage = target.fileAbsUrl || m._replyImage
          m._replyFrom = nameForSender(target.senderId, target)
        }
      }

      if (m._replyPreview && !m._replyFrom && m.replyToId) {
        const target = byId.get(String(m.replyToId))
        if (target) {
          m._replyFrom = nameForSender(target.senderId, target)
          m._replyImage = target.fileAbsUrl || m._replyImage
        }
      }
    })

    messages.value = mapped

    await scrollToBottom(true)
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load messages'
  } finally {
    loading.value = false
  }
}

function normalizeMembers(list) {
  return normalizeParticipantList(list).map(u => ({
    id: u.id,
    name: preferredDisplayName(u),
    avatar: u.avatar,
    username: safeUsername(u),
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
        name: preferredDisplayName(hit),
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
      name: preferredDisplayName(detail?.group || detail || {}),
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
    await scrollToBottom(true)

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
      await uncommentMessage(m.id)
      m._myReactions = []
    } else {
      await commentMessage(m.id, {
        type: 'emoji',
        content: emoji,
      })
      m._myReactions = [emoji]
    }
    await loadMessages()
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to react'
  }
}

async function loadForwardList() {
  forwardLoading.value = true
  forwardError.value = ''
  try {
    const raw = await getMyConversations()
    const items = raw?.data?.items || raw?.items || (Array.isArray(raw) ? raw : []) || []
    forwardList.value = normalizeConversationList(items)
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
    await scrollToBottom(true)
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
    updateConversationMeta({ name: groupInfo.value?.name, avatar: groupInfo.value?.avatar })
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
    groupNotice.value = 'Group name saved.'
    await loadGroupPanel()
    updateConversationMeta({ name: groupInfo.value?.name, avatar: groupInfo.value?.avatar })
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

function updateConversationMeta(extra = {}) {
  if (!convId.value) return
  window.dispatchEvent(
    new CustomEvent('conversations:refresh', {
      detail: {
        conversationId: convId.value,
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
  window.addEventListener('conversations:refresh', handleConversationRefresh)
  bootstrap()
})

onBeforeUnmount(() => {
  window.removeEventListener('conversations:refresh', handleConversationRefresh)
})

watch(
  () => messages.value.length,
  () => {
    scrollToBottom()
  }
)

watch(convId, async () => {
  pinnedToBottom.value = true
  forwardPanelOpen.value = false
  forwardTargetMessage.value = null
  forwardSearch.value = ''
  forwardError.value = ''
  replyTarget.value = null
  clearImageSelection()
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
  font-weight: 700;
  color: #1e293b;
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-width: 0;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-start;
}

.header-avatar {
  background: var(--avatar-bg);
  color: var(--avatar-text);
  border-color: var(--avatar-border);
}

.muted {
  font-size: 0.75rem;
  color: #64748b;
}

.content {
  width: 100%;
  padding: 16px;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  overflow: hidden;
}

.content-inner {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.chat-layout {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  gap: 16px;
  align-items: stretch;
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
}

.chat-layout.has-group {
  grid-template-columns: 320px minmax(0, 1fr) 300px;
}

.conv-rail {
  background: var(--panel);
  border: 1px solid var(--border);
  padding: 16px;
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

.conv-rail__badge {
  background: #ecfeff;
  color: #0f172a;
  border-radius: 12px;
  padding: 4px 10px;
  font-weight: 700;
}

.conv-rail__search input {
  width: 100%;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-control);
  padding: 10px 12px;
  font-size: var(--font-primary);
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
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 12px;
  background: var(--panel);
  cursor: pointer;
  transition: border 0.2s, background 0.2s, box-shadow 0.2s;
}

.conv-pill:hover {
  border-color: #d5dde7;
  background: #f3f4f6;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
}

.conv-pill.active {
  border-color: #9ae6b4;
  background: #eefbf3;
}

.conv-pill__avatar {
  background: #e2e8f0;
  color: #475569;
  border-color: #d1d5db;
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
  color: #9aa1ad;
  font-size: var(--font-secondary);
}

.conv-pill__bottom {
  color: #738396;
  font-size: var(--font-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-main {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
  min-height: 0;
  flex: 1 1 auto;
  overflow: hidden;
}

.group-panel {
  align-self: stretch;
}

.group-card {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  max-height: 100%;
  overflow: auto;
}

.group-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 10px;
  align-items: center;
}

.group-avatar {
  background: var(--avatar-bg);
  border-color: var(--avatar-border);
  color: var(--avatar-text);
  font-weight: 700;
}

.group-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.group-name {
  font-weight: 700;
  color: #0f172a;
  font-size: var(--font-title);
}

.group-sub {
  color: #94a3b8;
  font-size: var(--font-secondary);
}

.group-rename {
  background: #f8fafc;
  border: 1px solid #edf2f7;
  padding: 10px 10px 10px 12px;
}

.group-rename .btn.sm {
  background: #e5e7eb;
  color: #1f2937;
  box-shadow: none;
}

.group-rename .btn.sm:hover {
  background: #e8ecf1;
  filter: none;
}

.field.inline {
  display: flex;
  gap: 8px;
  align-items: center;
}

.field .input {
  flex: 1;
  background: transparent;
  border-color: transparent;
  box-shadow: none;
  padding-left: 0;
  padding-right: 0;
}

.field .input:focus {
  border-color: #cbd5e1;
  box-shadow: 0 0 0 2px rgba(34, 197, 94, 0.14);
}

.members-block,
.members-add {
  background: transparent;
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 12px;
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

.section-label {
  font-size: var(--font-secondary);
  font-weight: 600;
  color: #475569;
}

.member-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 6px;
}

.member-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 6px 10px;
  background: #fff;
}

.member-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.member-avatar {
  background: var(--avatar-bg);
  border-color: var(--avatar-border);
  font-weight: 700;
  color: var(--avatar-text);
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
  color: #4b5563;
  cursor: pointer;
  padding: 0;
}

.link.muted {
  color: #6b7280;
}

.link:hover {
  color: #334155;
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
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  background: transparent;
  border-radius: 0;
  padding: 4px 6px 12px;
}

.empty-thread {
  min-height: 50vh;
  display: grid;
  place-items: center;
  text-align: center;
  color: #0f172a;
}

.row {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  margin: var(--space-3) 0;
}

.row.mine {
  justify-content: flex-end;
  align-items: flex-end;
}

.avatar {
  flex-shrink: 0;
  background: var(--avatar-bg);
  border-color: var(--avatar-border);
}

.avatar.mine {
  margin-left: 4px;    /* Keep a small gap between my avatar and bubble */
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
  font-size: var(--font-secondary);
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
  font-size: var(--font-primary);
  line-height: 1.5;
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

.reply-body {
  display: flex;
  align-items: center;
  gap: 8px;
}

.reply-thumb {
  width: 34px;
  height: 34px;
  border-radius: 6px;
  object-fit: cover;
  border: 1px solid #e2e8f0;
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
  font-size: var(--font-secondary);
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
  font-size: var(--font-secondary);
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

.reactions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 6px;
}

.reaction-pill {
  border-radius: 999px;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  padding: 2px 8px;
  font-size: 0.9rem;
  line-height: 1.2;
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
  gap: var(--space-2);
  margin-top: var(--space-2);
  align-items: center;
}

.icon-btn {
  border: none;
  background: none;
  cursor: pointer;
  font-size: 0.95rem;
  line-height: 1;
  padding: 0;
  border-radius: var(--radius-control);
  width: 28px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #334155;
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
  border: 1px solid var(--border);
  border-radius: 0;
  background: var(--panel);
  cursor: pointer;
  transition: 0.2s;
}

.forward-item:hover {
  background: #eef2ff;
  border-color: #c7d2fe;
}

.forward-avatar {
  background: #e2e8f0;
  font-weight: 700;
  color: #475569;
}

.forward-meta {
  text-align: left;
}

.forward-name {
  font-weight: 700;
  color: #0f172a;
  font-size: var(--font-primary);
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
