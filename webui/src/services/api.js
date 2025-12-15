// src/services/api.js
import axios from './axios'

/* -------------------- Authentication helpers -------------------- */

// Read token from either localStorage or sessionStorage.
const readToken = () =>
  localStorage.getItem('token') || sessionStorage.getItem('authToken') || ''

// Determine whether the user is currently authenticated.
export function isAuthed() {
  return !!readToken()
}

// Attach the Authorization header when a token is present.
function withAuthConfig(cfg = {}) {
  const token = readToken()
  const headers = { ...(cfg?.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  return { ...(cfg || {}), headers }
}

// Unwrap one layer of { code, message, data } without altering deeper payloads.
function unwrap(res) {
  const root = res?.data ?? res

  if (root == null || typeof root !== 'object') {
    return root
  }

  if ('data' in root) {
    return root.data
  }
  return root
}




// Shorthand wrappers that automatically include auth headers.
const get   = (url, cfg)       => axios.get(url,    withAuthConfig(cfg))
const del   = (url, cfg)       => axios.delete(url, withAuthConfig(cfg))
const post  = (url, data, cfg) => axios.post(url,   data, withAuthConfig(cfg))
const put   = (url, data, cfg) => axios.put(url,    data, withAuthConfig(cfg))

/* -------------------- API base URL helpers -------------------- */

// Resolve API base URL using axios defaults, runtime override, or Vite env values.
export function resolveApiBase() {
  return (axios?.defaults?.baseURL)
      || (typeof window !== 'undefined' ? window.__API_URL__ : '')
      || (import.meta?.env?.VITE_API_BASE_URL)
      || ''
}
export function absUrl(path) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  const base = resolveApiBase() || ''
  if (path.startsWith('/')) return `${base}${path}`
  return `${base}/${path}`
}

/* -------------------- Public endpoints and utilities -------------------- */

export async function doLogin(name) {
  // Basic login that posts the username to /session.
  const res = await axios.post('/session', { name })

  // Typical axios response: res.data = { code, message, data: { user, token } }
  const root = res && res.data ? res.data : res

  // Accept multiple wrapper shapes to keep the client tolerant.
  const token =
    root?.data?.token ||   // {code, message, data:{user, token}}
    root?.token           // {user, token}
  
  if (!token) {
    console.error('Login response payload =', root)
    throw new Error('Login response missing token')
  }

  // Persist token in both storage locations for compatibility.
  localStorage.setItem('token', token)
  sessionStorage.setItem('authToken', token)

  try {
    window.dispatchEvent(new Event('auth:changed'))
  } catch {}

  return token
}


export async function liveness() { return unwrap(await get('/liveness')) }
export async function healthz()  { return unwrap(await get('/healthz')) }

/* -------------------- Logout helper -------------------- */

/**
 * Logout: clear all tokens and cached info.
 * Client-side only; the backend does not need to revoke anything.
 */
export function doLogout() {
  try {
    localStorage.removeItem('token')
    sessionStorage.removeItem('authToken')
    localStorage.removeItem('username')
    localStorage.removeItem('name')
    localStorage.removeItem('me')
    window.dispatchEvent(new Event('auth:changed'))
  } catch {}
}

/* -------------------- User profile -------------------- */

export async function getMyProfile() {
  const u = unwrap(await get('/users/me')) || {}
  // Normalize to the keys expected by the UI.
  return {
    ...u,
    id: String(u.id ?? ''),
    name: u.name ?? '',
    avatarUri: u.avatarUri ?? u.avatar_uri ?? u.avatar_url ?? u.avatar ?? '',
    updatedAt: u.updatedAt ?? u.updated_at ?? Date.now(),
  }
}

export async function getUserProfile(userId) {
  const id = String(userId || '').trim()
  if (!id) throw new Error('userId required')
  return unwrap(await get(`/users/profile/${encodeURIComponent(id)}`))
}

export async function setMyUserName(name) {
  if (!name) throw new Error('empty name')
  try {
    return unwrap(await put('/users/set_username', { name }))
  } catch (e) {
    if (e?.response?.status === 409) {
      throw new Error('Username already taken. Please choose another.')
    }
    const msg = e?.response?.data?.message || e?.message || 'Failed to set username'
    throw new Error(msg)
  }
}

/** Upload avatar file (field name: 'upload'). */
export async function setMyPhotoFile(file) {
  if (!file) throw new Error('No file selected')
  const form = new FormData()
  form.append('upload', file, file.name)
  try {
    return unwrap(await put('/users/set_photo', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    }))
  } catch (e) {
    const msg = e?.response?.data?.message || e?.message || 'Failed to upload photo'
    throw new Error(msg)
  }
}

/** Build avatar URL with a cache-busting parameter. */
export function getAvatarUrl(userLike) {
  const u = userLike || {}
  const raw =
    u.avatarUri || u.avatar_uri || u.avatar_url || u.avatar || u.photo_url || u.photo || ''
  const url = raw ? absUrl(raw) : ''
  const v = encodeURIComponent(u.updatedAt || u.updated_at || Date.now())
  return url ? `${url}${url.includes('?') ? '&' : '?'}v=${v}` : ''
}

/** Search users via /users/search?q. */
export async function searchUsers(q = '') {
  const v = unwrap(await get('/users/search', { params: { q } }))
  return Array.isArray(v) ? v : (v?.items ?? v?.users ?? v?.list ?? [])
}

/**
 * Build the contacts list by aggregating search results across alphabets.
 * Keeps compatibility with the ContactsView component.
 */
export async function listContacts() {
  const letters = 'abcdefghijklmnopqrstuvwxyz'.split('')
  const uniq = new Map()
  for (const ch of letters) {
    try {
      const arr = await searchUsers(ch)
      for (const u of arr) {
        const id = String(u?.id ?? '')
        if (id && !uniq.has(id)) uniq.set(id, u)
      }
    } catch {}
  }
  // Filter out the current user
  let meId = ''
  try { const me = await getMyProfile(); meId = String(me.id || '') } catch {}
  return Array.from(uniq.values()).filter(u => String(u?.id ?? '') !== meId)
}

/* -------------------- Conversation helpers -------------------- */

export function isGroupConversation(c) {
  return c?.type === 'group'
}


/** Conversation title: peer name for private chats, group name otherwise. */
export function titleForConversation(c, myId = '') {
  if (!c) return 'Chat'

  // private chat: participants = 2
  const isPrivate =
    Array.isArray(c.participants) &&
    c.participants.length === 2 &&
    c.type !== 'group'

  if (isPrivate) {
    const list = c.participants
    const other = list.find(u => String(u.id) !== String(myId))
    return other?.name || 'Chat'
  }

  // group
  return c.name || 'Group'
}


/** Conversation avatar: group avatar for groups, peer avatar for private chats. */
export function avatarForConversation(c, myId) {
  if (!c) return ''

  const isPrivate =
    Array.isArray(c.participants) &&
    c.participants.length === 2 &&
    c.type !== 'group'

  if (isPrivate) {
    const list = c.participants
    const other = list.find(u => String(u.id) !== String(myId))
    return getAvatarUrl(other || {})
  }

  return getAvatarUrl(c)
}



/* -------------------- Conversation endpoints -------------------- */

export async function startConversation(payload = {}) {
  // OpenAPI: POST /conversations
  // Accept multiple key names and normalize to memberIds.
  const body = {}
  let memberIds = []
  if (Array.isArray(payload.memberIds)) memberIds = payload.memberIds
  else if (Array.isArray(payload.members)) memberIds = payload.members
  else if (payload.user_id || payload.userId) memberIds = [payload.user_id || payload.userId]

  if (memberIds.length) body.memberIds = memberIds.map(String)
  return unwrap(await post('/conversations', body))
}

export async function startPrivateConversation(userOrId) {
  const id = String(userOrId?.id ?? userOrId)
  if (!id) throw new Error('userId required')

  return unwrap(await post('/conversations', {
    name: 'private',   // Required to indicate a private conversation
    memberIds: [id]
  }))
}



export async function getMyConversations() {
  return unwrap(await get('/conversations'))
}

// Fetch members directly; fall back to group lookup when the endpoint is missing.
export async function getConversationMembers(conversationId) {
  const id = String(conversationId || '').trim();
  if (!id) throw new Error('conversationId required');

  // 1) Prefer the dedicated conversations members endpoint when available.
  try {
    const v = unwrap(await get(`/conversations/${encodeURIComponent(id)}/members`));
    return Array.isArray(v) ? v : (v?.items ?? v?.members ?? v?.list ?? []);
  } catch {}

  // 2) Fallback: locate the group by conversationId, then read its members.
  try {
    const glist = unwrap(await get('/groups'));
    const arr = Array.isArray(glist) ? glist : (glist?.items ?? glist?.groups ?? glist?.list ?? []);
    const hit = (arr || []).find(g =>
      String(g?.conversationId ?? g?.conversation_id ?? '') === id
    );
    if (hit?.id) {
      const gd = unwrap(await get(`/groups/${encodeURIComponent(String(hit.id))}`));
      const members = gd?.members ?? hit?.members ?? [];
      return Array.isArray(members) ? members : (members?.items ?? []);
    }
  } catch {}

  return [];
}

export async function deleteConversation(id) {
  const cid = encodeURIComponent(String(id))
  return unwrap(await del(`/conversations/${cid}`))
}



/* -------------------- groups -------------------- */

export async function createGroup({ name, memberIds = [] }) {
  return unwrap(await post('/groups', { name, memberIds: memberIds.map(String) }))
}

export async function getGroupsList() {
  return unwrap(await get('/groups'))
}

export async function getGroupDetail(id) {
  return unwrap(await get(`/groups/${encodeURIComponent(String(id))}`))
}

export async function setGroupName(id, name) {
  return unwrap(await put(`/groups/${encodeURIComponent(String(id))}/name`, { name }))
}

export async function setGroupPhoto(id, file) {
  if (!file) throw new Error('No file selected')
  const form = new FormData()
  form.append('upload', file, file.name)
  return unwrap(await put(`/groups/${encodeURIComponent(String(id))}/photo`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }))
}

export async function addToGroup(id, memberIds = []) {
  return unwrap(await post(`/groups/${encodeURIComponent(String(id))}/members`, {
    memberIds: memberIds.map(String),
  }))
}

/**
 * DELETE /groups/:id/members implements self-removal.
 * To remove another member, the backend should read body { userId }.
 * The client sends that shape here; hide the UI action if the backend lacks support.
 */
export async function removeFromGroup(id, userId) {
  return unwrap(await del(`/groups/${encodeURIComponent(String(id))}/members`, {
    data: { userId: String(userId) },
  }))
}

export async function leaveGroup(id) {
  return unwrap(await del(`/groups/${encodeURIComponent(String(id))}/members`))
}

/* -------------------- messages -------------------- */

export async function getMessages({ conversationId, limit = 50, beforeCursor, afterCursor } = {}) {
  const params = { conversationId, limit }
  if (beforeCursor) params.beforeCursor = beforeCursor
  if (afterCursor)  params.afterCursor  = afterCursor
  return unwrap(await get('/messages', { params }))
}

export async function sendMessage({
  conversationId,
  content,
  type = 'text',
  replyTo,
  replyToId,
} = {}) {
  const body = {
    conversationId: String(conversationId || ''),
    content: String(content || ''),
    type,
  }

  // Support reply metadata: accept various keys; replyToId is the backend field.
  const reply = replyToId || replyTo
  if (reply) {
    body.replyToId = reply
  }

  return unwrap(await post('/messages', body))
}

export async function sendImageMessage({ conversationId, file }) {
  if (!file) throw new Error('No file selected')
  
  // 1) Upload the binary first
  const form = new FormData()
  form.append('upload', file, file.name)

  const r = unwrap(await post('/upload/messages', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }))

  const fileUrl  = r?.fileUrl  || r?.file_url  || ''
  const filename = r?.filename || file.name

  if (!fileUrl) {
    console.error('upload response =', r)
    throw new Error('Upload failed: no fileUrl')
  }

  // 2) Send a normal message referencing the uploaded image URL
  return sendMessage({
    conversationId,
    content: fileUrl,
    type: 'image',
    filename,
  })
}


export async function getMessageById(id) {
  return unwrap(await get(`/messages/${encodeURIComponent(String(id))}`))
}

export async function deleteMessage(id) {
  return unwrap(await del(`/messages/${encodeURIComponent(String(id))}`))
}

export async function forwardMessage(messageId, conversationId) {
  return unwrap(await post(
    `/messages/${encodeURIComponent(String(messageId))}/forward`,
    { conversationId: String(conversationId) }
  ))
}

/* ---- Comments, replies, and reactions ---- */

// Retrieve comment list for a message
export async function getMessageComments(id) {
  const mid = encodeURIComponent(String(id))
  return unwrap(await get(`/messages/${mid}/comment`))
}

// Create a comment or reply
export async function commentMessage(msgId, payload) {
  const mid = encodeURIComponent(String(msgId))
  let body = { type: 'text', content: '' }

  if (payload && typeof payload === 'object' && !Array.isArray(payload)) {
    body = {
      type: payload.type || 'text',
      content: payload.content ?? '',
    }
  } else {
    body = {
      type: payload || 'text',
      content: arguments[2] ?? '',
    }
  }

  return unwrap(await post(`/messages/${mid}/comment`, body))
}



// Remove comments: backend ignores the body and only expects a POST
export async function uncommentMessage(id) {
  const mid = encodeURIComponent(String(id))
  return unwrap(await post(`/messages/${mid}/uncomment`))
}

// Semantic aliases for convenience
export const replyToMessage     = commentMessage
export const getMessageReplies  = getMessageComments
export const removeMessageReply = uncommentMessage

// React with an emoji (content stored as :react:emoji with type=emoji)
export async function reactToMessage(id, emoji) {
  return commentMessage(id, { type: 'emoji', content: emoji })
}

// Remove reactions: backend currently deletes all emojis at once
export async function unreactToMessage(id, emoji) {
  return uncommentMessage(id)
}

/* ---- Read receipt ticks (0=queued 1=sent 2=delivered 3=read; -1 for others) ---- */
export function ticksFor(m, myId) {
  if (!m || String(m.senderId) !== String(myId)) return -1
  const s = (m.status || '').toLowerCase()
  if (s === 'seen' || s === 'read') return 3
  if (s === 'delivered') return 2
  if (s === 'sent') return 1
  return 0
}

/* -------------------- default export -------------------- */

const api = {
  // auth / ping / base
  isAuthed,
  doLogin,
  liveness,
  healthz,
  resolveApiBase,
  absUrl,
  doLogout,

  // users
  getMyProfile,
  getUserProfile,
  setMyUserName,
  setMyPhotoFile,
  getAvatarUrl,
  searchUsers,
  listContacts,

  // conversations
  isGroupConversation,
  titleForConversation,
  avatarForConversation,
  startConversation,
  startPrivateConversation,
  getMyConversations,
  deleteConversation,

  // groups
  createGroup,
  getGroupsList,
  getGroupDetail,
  setGroupName,
  setGroupPhoto,
  addToGroup,
  removeFromGroup,
  leaveGroup,

  // messages
  getMessages,
  sendMessage,
  sendImageMessage,
  getMessageById,
  deleteMessage,
  forwardMessage,
  getMessageComments,
  commentMessage,
  uncommentMessage,
  replyToMessage,
  getMessageReplies,
  removeMessageReply,
  reactToMessage,
  unreactToMessage,

  // ticks
  ticksFor,
}

export default api
