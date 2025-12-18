// Client-side API helper functions used by the web UI.
import axios from './axios'

/* Authentication helpers */

let activeAbortController = createAbortController()

function createAbortController() {
  return new AbortController()
}

function getRequestSignal(customSignal) {
  if (customSignal) return customSignal

  if (!activeAbortController || activeAbortController.signal.aborted) {
    activeAbortController = createAbortController()
  }
  return activeAbortController.signal
}

export function abortActiveRequests(reason = 'Canceled due to logout') {
  try {
    if (activeAbortController) {
      activeAbortController.abort(reason)
    }
  } finally {
    activeAbortController = createAbortController()
  }
}

export function isAbortError(err) {
  const code = err?.code || err?.name || ''
  const msg = String(err?.message || '').toLowerCase()
  return (
    code === 'ERR_CANCELED' ||
    code === 'CanceledError' ||
    code === 'AbortError' ||
    msg === 'canceled' ||
    msg === 'cancelled' ||
    msg === 'aborted' ||
    msg.includes('canceled') ||
    msg.includes('aborted')
  )
}

const AUTH_EVENT = 'auth:changed'

function emitAuthChanged() {
  try {
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new Event(AUTH_EVENT))
    }
  } catch {}
}

// Read the persisted token from localStorage.
export const readToken = () => {
  if (typeof window === 'undefined') return ''

  const rawToken = localStorage.getItem('token') || ''
  const token = String(rawToken).trim()

  // Treat placeholder values (e.g., "null" or "undefined") as missing tokens so we fall
  // back to the login page instead of mistakenly considering the user authenticated.
  return token && token !== 'null' && token !== 'undefined' ? token : ''
}

// Backward-compatible alias used by legacy callers; ensures any lingering
// references resolve to the single canonical accessor above instead of
// throwing a ReferenceError at runtime.
export const readTokens = () => readToken()
// Determine whether the user is currently authenticated.
export function isAuthed() {
  return !!readToken()
}

/**
 * Remove all persisted authentication artifacts.
 * Useful for both logout and resetting state before a new login attempt.
 */
export function clearAuthArtifacts({ emit = true } = {}) {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    localStorage.removeItem('name')
    localStorage.removeItem('me')
  }

  if (axios?.defaults?.headers?.common) {
    delete axios.defaults.headers.common.Authorization
  }

  if (emit) emitAuthChanged()
}

/**
 * Persist a valid auth token across storage layers and axios defaults.
 */
export function persistAuthToken(token, { emit = true } = {}) {
  const value = typeof token === 'string' ? token.trim() : String(token || '').trim()
  if (!value) throw new Error('Invalid token provided')

  if (typeof window !== 'undefined') {
    localStorage.setItem('token', value)
  }

  if (axios?.defaults?.headers?.common) {
    axios.defaults.headers.common.Authorization = `Bearer ${value}`
  }

  if (emit) emitAuthChanged()
  return value
}

// Attach the Authorization header when a token is present.
export function withAuthConfig(cfg = {}) {
  const { authRequired = true, signal, ...rest } = cfg || {}
  const token = readToken()

  if (authRequired && !token) {
    throw new Error('Not authenticated')
  }

  const headers = { ...(rest.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`

  return {
    ...rest,
    headers,
    signal: getRequestSignal(signal),
  }
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

/**
 * Normalize user-like payloads to a consistent shape consumed by the UI.
 * Ensures we never surface placeholder handles like "-" by falling back to
 * id/"unknown" when the username is missing.
 */
export function normalizeUser(u = {}) {
  const rawId =
    u.id ?? u.user_id ?? u.userId ?? u._id ?? u.uid ?? u.uuid ?? u.userID ?? null
  const id = String(rawId || '').trim() || 'unknown'

  const usernameCandidate = String(
    u.username ?? u.user_name ?? u.userName ?? u.handle ?? ''
  ).trim()
  const nameCandidate = String(u.name ?? u.full_name ?? u.fullName ?? '').trim()

  const username = usernameCandidate && usernameCandidate !== '-' ? usernameCandidate : null
  const name = nameCandidate || username || null

  const avatarUri =
    u.avatarUri || u.avatar_uri || u.avatar_url || u.avatar || u.photo_url || u.photo || ''
  const updatedAt = u.updatedAt ?? u.updated_at ?? Date.now()

  return {
    ...u,
    id,
    username,
    name,
    avatarUri,
    avatarUrl: getAvatarUrl({ ...u, avatarUri, updatedAt }),
    updatedAt,
  }
}

/* API base URL helpers */

// Resolve the API base URL using axios defaults, runtime overrides, or Vite env values.
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

/* Public endpoints and utilities */

export async function doLogin(name, { persist = true } = {}) {
  // Basic login that posts the username to /session.
  const res = await axios.post('/session', { name })

  // Typical axios response: res.data = { code, message, data: { user, token } }.
  const root = res && res.data ? res.data : res

  // Accept multiple wrapper shapes to keep the client tolerant.
  const token =
    root?.data?.token ||   // {code, message, data:{user, token}}
    root?.token           // {user, token}

  if (!token) {
    console.error('Login response payload =', root)
    throw new Error('Login response missing token')
  }

  if (persist) {
    // Persist the token, signaling auth changes only once the flow succeeds.
    clearAuthArtifacts({ emit: false })
    persistAuthToken(token, { emit: false })
    emitAuthChanged()
  }

  return { token, user: root?.data?.user || root?.user || null }
}


export async function liveness() { return unwrap(await get('/liveness', { authRequired: false })) }
export async function healthz()  { return unwrap(await get('/healthz', { authRequired: false })) }

/* Logout helper */

/**
 * Logout: clear all tokens and cached info.
 * Client-side only; the backend does not need to revoke anything.
 */
export function doLogout() {
  try {
    abortActiveRequests('User logged out')
    clearAuthArtifacts()
  } catch {}
}

/* User profile helpers */

export async function getMyProfile() {
  const u = unwrap(await get('/users/me')) || {}
  return normalizeUser(u)
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
  const list = Array.isArray(v) ? v : (v?.items ?? v?.users ?? v?.list ?? [])
  return list.map(normalizeUser)
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
  return Array.from(uniq.values())
}

/* Conversation helpers */

export function isGroupConversation(c) {
  return c?.type === 'group' || !!(c?.groupId || c?.group_id || c?.group?.id)
}

/** Conversation title: peer name for private chats, group name otherwise. */
export function titleForConversation(c, myId = '') {
  if (!c) return 'Chat'

  // Private chat: participants = 2.
  const isPrivate =
    Array.isArray(c.participants) &&
    c.participants.length === 2 &&
    !isGroupConversation(c)

  if (isPrivate) {
    const list = c.participants
    const other = list.find(u => String(u.id) !== String(myId))
    return other?.name || 'Chat'
  }

  // Group conversation fallback.
  return c.name || 'Group'
}

/** Conversation avatar: group avatar for groups, peer avatar for private chats. */
export function avatarForConversation(c, myId) {
  if (!c) return ''

  const isPrivate =
    Array.isArray(c.participants) &&
    c.participants.length === 2 &&
    !isGroupConversation(c)

  if (isPrivate) {
    const list = c.participants
    const other = list.find(u => String(u.id) !== String(myId))
    return getAvatarUrl(other || {})
  }

  return getAvatarUrl(c)
}

/* Conversation endpoints */

export async function startConversation(payload = {}) {
  // OpenAPI: POST /conversations.
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
    name: 'private',   // Required to indicate a private conversation.
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

  // Step 1: Prefer the dedicated conversations members endpoint when available.
  try {
    const v = unwrap(await get(`/conversations/${encodeURIComponent(id)}/members`));
    return Array.isArray(v) ? v : (v?.items ?? v?.members ?? v?.list ?? []);
  } catch {}

  // Step 2: Locate the group by conversationId, then read its members.
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



/* Group helpers */

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

/* Message helpers */

export async function getMessages({ conversationId, limit = 50, beforeCursor, afterCursor } = {}) {
  const params = { conversationId, limit }
  if (beforeCursor) params.beforeCursor = beforeCursor
  if (afterCursor)  params.afterCursor  = afterCursor
  return unwrap(await get('/messages', { params }))
}

export async function sendMessage({
  conversationId,
  content,
  fileUrl,
  type = 'text',
  replyTo,
  replyToId,
} = {}) {
  const body = {
    conversationId: String(conversationId || ''),
    content: String(content || ''),
    type,
  }

  if (fileUrl !== undefined) {
    body.fileUrl = fileUrl
  }

  // Support reply metadata: accept various keys; replyToId is the backend field.
  const reply = replyToId || replyTo
  if (reply) {
    body.replyToId = reply
  }

  return unwrap(await post('/messages', body))
}

export async function sendImageMessage({ conversationId, file, caption = '', replyToId }) {
  if (!file) throw new Error('No file selected')
  
  // 1) Upload the binary first
  const form = new FormData()
  form.append('upload', file, file.name)

  const r = unwrap(await post('/upload/messages', form))

  const fileUrl  = r?.fileUrl  || r?.file_url  || ''
  const filename = r?.filename || file.name

  if (!fileUrl) {
    console.error('upload response =', r)
    throw new Error('Upload failed: no fileUrl')
  }

  // Step 2: Send a normal message referencing the uploaded image URL.
  return sendMessage({
    conversationId,
    content: caption,
    type: 'image',
    fileUrl,
    replyToId,
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

/* Comments, replies, and reactions */

// Retrieve the comment list for a message.
export async function getMessageComments(id) {
  const mid = encodeURIComponent(String(id))
  return unwrap(await get(`/messages/${mid}/comment`))
}

// Create a comment or reply.
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



// Remove comments: backend ignores the body and only expects a POST.
export async function uncommentMessage(id) {
  const mid = encodeURIComponent(String(id))
  return unwrap(await post(`/messages/${mid}/uncomment`))
}

// Semantic aliases for convenience.
export const replyToMessage     = commentMessage
export const getMessageReplies  = getMessageComments
export const removeMessageReply = uncommentMessage

// React with an emoji (content stored as :react:emoji with type=emoji).
export async function reactToMessage(id, emoji) {
  return commentMessage(id, { type: 'emoji', content: emoji })
}

// Remove reactions: backend currently deletes all emojis at once.
export async function unreactToMessage(id, emoji) {
  return uncommentMessage(id)
}

/* Read receipt ticks (0=queued 1=sent 2=delivered 3=read; -1 for others) */
export function ticksFor(m, myId) {
  if (!m || String(m.senderId) !== String(myId)) return -1
  if (m.read) return 2
  const s = (m.status || '').toLowerCase()
  if (s === 'seen' || s === 'read') return 3
  if (s === 'delivered') return 2
  if (s === 'sent') return 1
  return 0
}

/* Default export for convenience */

const api = {
  // auth / ping / base
  isAuthed,
  readToken,
  readTokens,
  doLogin,
  abortActiveRequests,
  isAbortError,
  persistAuthToken,
  clearAuthArtifacts,
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
