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
      <div class="chat-layout">
        <div class="chat-main">
          <div ref="scrollbox" class="scroll">
            <div
              v-for="m in messages"
              :key="m.id"
              class="row"
              :id="`msg-${m.id}`"
              :class="{ mine: isMine(m), highlight: replyHighlightId === String(m.id) }"
            >
              <!-- LEFT AVATAR  -->
              <div
                v-if="!isMine(m)"
                class="avatar"
                :class="{ placeholder: !avatarFor(m) }"
                :style="avatarBg(avatarFor(m))"
              >
                <span v-if="!avatarFor(m)" class="avatar-initial">{{ avatarInitial(m) }}</span>
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
              <div
                v-if="isMine(m)"
                class="avatar mine"
                :class="{ placeholder: !myAvatar }"
                :style="avatarBg(myAvatar)"
              >
                <span v-if="!myAvatar" class="avatar-initial">{{ avatarInitial(m) }}</span>
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
              <ul v-else class="member-list">
                <li v-for="u in groupMembers" :key="u.id" class="member-item">
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
            <p v-if="groupNotice" class="notice small">{{ groupNotice }}</p>
          </div>
        </aside>
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
  isAuthed,
  getMyProfile,
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

function avatarBg(src) {
  return src ? { backgroundImage: `url('${src}')` } : {}
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
    if (isGroup.value) {
      await loadGroupPanel()
    } else {
      groupInfo.value = null
    }
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

function normalizeMembers(list) {
  return (list || [])
    .map(u => ({
      id: String(u?.id ?? u?.userId ?? u?.user_id ?? ''),
      name: u?.name || u?.username || 'User',
      avatar: getAvatarUrl({
        avatarUri: u?.avatarUri ?? u?.avatar_uri ?? u?.avatar_url ?? u?.avatar,
        updatedAt: u?.updatedAt ?? u?.updated_at ?? Date.now(),
      }),
    }))
    .filter(u => u.id)
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
    console.error('resolveGroupId failed', e)
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
      currentConv.value = { ...(currentConv.value || {}), participants: membersRaw }
    }

    groupInfo.value = {
      id: String(detail?.id ?? detail?.group_id ?? gid),
      name: detail?.name || detail?.title || detail?.group?.name || headerTitle.value,
      avatar: getAvatarUrl(detail || detail?.group || {}),
      members,
    }

    groupNameDraft.value = groupInfo.value.name || ''
  } catch (e) {
    groupErr.value = e?.response?.data?.message || e?.message || 'Failed to load group info'
  } finally {
    groupLoading.value = false
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
  const v = ticksFor(m, meId.value)
  if (v === 3) return '✓✓ read'
  if (v === 2) return '✓✓ delivered'
  if (v === 1) return '✓ sent'
  if (v === 0) return '…'
  return ''
}

// ---- Group management ----
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
  max-width: 1100px;
  margin: 0 auto;
  padding: 16px;
}

.chat-layout {
  display: grid;
  grid-template-columns: 2.2fr 1fr;
  gap: 14px;
  align-items: start;
}

.chat-main {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.group-panel {
  position: sticky;
  top: 68px;
}

.group-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  padding: 12px;
  box-shadow: 0 6px 18px rgba(2, 6, 23, 0.06);
  display: grid;
  gap: 10px;
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
  border-radius: 14px;
  background: #e2e8f0;
  display: grid;
  place-items: center;
  background-size: cover;
  background-position: center;
  color: #1e293b;
  font-weight: 700;
}

.group-avatar.placeholder {
  border: 1px solid #e2e8f0;
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

.members-block {
  border-top: 1px dashed #e2e8f0;
  padding-top: 6px;
  display: grid;
  gap: 8px;
}

.members-add {
  border-top: 1px dashed #e2e8f0;
  padding-top: 8px;
  display: grid;
  gap: 6px;
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
  border-radius: 12px;
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
  border-radius: 12px;
  background: #e2e8f0;
  display: grid;
  place-items: center;
  background-size: cover;
  background-position: center;
  font-weight: 700;
  color: #1f2937;
}

.member-avatar.placeholder {
  border: 1px solid #e2e8f0;
}

.member-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
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
  width: 42px;
  height: 42px;
  border-radius: 50%;
  overflow: hidden;

  flex-shrink: 0;
  display: grid;
  place-items: center;
  background: #e2e8f0;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
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
  border-radius: 14px;
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
