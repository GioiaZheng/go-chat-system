// src/services/api.js
// Full frontend API client that matches major routes in backend.
// Every call injects Authorization locally.
// Return values unwrap { code, data } => return data only.

import axios from './axios'

/* -------------------- helpers -------------------- */

function auth() {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

function unwrap(res) {
  // 支持 { code, data } 或直接 payload
  if (res?.data && typeof res.data === 'object' && 'data' in res.data) {
    return res.data.data
  }
  return res.data
}

/* -------------------- public / utility -------------------- */

/**
 * POST /session  -- simplified login
 * 兼容后端需要 username 的情况：同时传 name/username/display_name
 * 返回中兼容 identifier/token/id/user_id 等多种写法
 */
export async function doLogin(name) {
  const res = await axios.post('/session', { name })
  const payload = unwrap(res)

  const identifier =
    payload?.identifier ??
    payload?.token ??
    payload?.id ??
    payload?.user_id ??
    payload

  if (!identifier) throw new Error('Login response missing identifier')

  // 同时写入 localStorage + sessionStorage
  localStorage.setItem('token', identifier)
  sessionStorage.setItem('authToken', identifier)

  return identifier
}

/** POST /register  -- optional helper */
export async function registerUser({ name, email } = {}) {
  const res = await axios.post('/register', { name, email })
  return unwrap(res)
}

/** GET /liveness */
export async function liveness() {
  const res = await axios.get('/liveness')
  return unwrap(res)
}

/* -------------------- users -------------------- */
/* Canonical:
   PUT /users/set_username
   PUT /users/set_photo
   GET /users/me
   GET /users/search
   GET /users/profile/:user_id
*/

export async function getMyProfile() {
  const res = await axios.get('/users/me', { headers: auth() })
  return unwrap(res)
}

export async function setMyUserName (username) {
  // 兼容后端把“用户名”写到 name 或 username 的两种实现
  const body = { username, name: username }
  const res = await axios.put('/users/set_username', body, { headers: auth() })
  return unwrap(res)
}

/** setMyPhoto – preset (?preset=avatar7) or multipart upload (field: upload) */
export async function setMyPhoto({ preset, file }) {
  if (preset) {
    const res = await axios.put(`/users/set_photo?preset=${encodeURIComponent(preset)}`, null, {
      headers: auth(),
    })
    return unwrap(res)
  }
  if (file) {
    const form = new FormData()
    form.append('upload', file, file.name)
    const res = await axios.put('/users/set_photo', form, {
      headers: { ...auth(), 'Content-Type': 'multipart/form-data' },
    })
    return unwrap(res)
  }
  throw new Error('Provide either { preset } or { file }')
}

export async function searchUsers(q) {
  const res = await axios.get('/users/search', { params: { q }, headers: auth() })
  return unwrap(res)
}

export async function getUserProfile(userIdOrName) {
  // Prefer path param by id; if only username, backend also accepts ?username=
  if (/^[a-f0-9-]{8,}$/i.test(userIdOrName)) {
    const res = await axios.get(`/users/profile/${encodeURIComponent(userIdOrName)}`, {
      headers: auth(),
    })
    return unwrap(res)
  }
  const res = await axios.get('/users/profile/0', {
    // dummy path id; backend resolves via ?username
    params: { username: userIdOrName },
    headers: auth(),
  })
  return unwrap(res)
}

/* -------------------- conversations -------------------- */
/* Canonical:
   POST /conversations
   GET  /conversations
   Aliases (optional compat):
   GET  /conversations/:conversationId
   POST /conversations/:conversationId/messages
*/

export async function startConversation({ user_id, group_id } = {}) {
  const res = await axios.post('/conversations', { user_id, group_id }, { headers: auth() })
  return unwrap(res)
}

export async function getMyConversations() {
  const res = await axios.get('/conversations', { headers: auth() })
  return unwrap(res)
}

export async function getConversation(conversationId) {
  const res = await axios.get(`/conversations/${encodeURIComponent(conversationId)}`, {
    headers: auth(),
  })
  return unwrap(res)
}

export async function sendConversationMessage(
  conversationId,
  content,
  { type = 'text', status = 'sent' } = {}
) {
  const res = await axios.post(
    `/conversations/${encodeURIComponent(conversationId)}/messages`,
    { content, type, status },
    { headers: auth() }
  )
  return unwrap(res)
}

/* -------------------- messages -------------------- */
/* Canonical (对齐后端的 camelCase 参数名):
   GET    /messages                      (supports ?conversationId=...&limit=&beforeCursor=&afterCursor=)
   POST   /messages                      (body: { conversationId, content, type? })
   GET    /messages/:id
   DELETE /messages/:id
   POST   /messages/:id/forward
   GET    /messages/:id/comment
   POST   /messages/:id/comment
   POST   /messages/:id/uncomment
*/

export async function getMessages({ conversationId, limit, beforeCursor, afterCursor } = {}) {
  const res = await axios.get('/messages', {
    params: { conversationId, limit, beforeCursor, afterCursor },
    headers: auth(),
  })
  return unwrap(res)
}

export async function sendMessage({ content, conversationId, type = 'text', status = 'sent' }) {
  const res = await axios.post(
    '/messages',
    { content, conversationId, type, status },
    { headers: auth() }
  )
  return unwrap(res)
}

export async function getMessageById(messageId) {
  const res = await axios.get(`/messages/${encodeURIComponent(messageId)}`, { headers: auth() })
  return unwrap(res)
}

export async function deleteMessage(messageId) {
  const res = await axios.delete(`/messages/${encodeURIComponent(messageId)}`, { headers: auth() })
  return unwrap(res)
}

export async function forwardMessage(messageId, { conversationId }) {
  const res = await axios.post(
    `/messages/${encodeURIComponent(messageId)}/forward`,
    { conversationId },
    { headers: auth() }
  )
  return unwrap(res)
}

export async function getMessageComments(messageId) {
  const res = await axios.get(`/messages/${encodeURIComponent(messageId)}/comment`, {
    headers: auth(),
  })
  return unwrap(res)
}

export async function commentMessage(messageId, comment) {
  const res = await axios.post(
    `/messages/${encodeURIComponent(messageId)}/comment`,
    { comment },
    { headers: auth() }
  )
  return unwrap(res)
}

export async function uncommentMessage(messageId) {
  const res = await axios.post(`/messages/${encodeURIComponent(messageId)}/uncomment`, null, {
    headers: auth(),
  })
  return unwrap(res)
}

/* -------------------- groups -------------------- */
/* Canonical:
   POST   /groups
   GET    /groups
   GET    /groups/:id
   PUT    /groups/:id/name
   PUT    /groups/:id/photo
   POST   /groups/:id/members
   DELETE /groups/:id/members
*/

export async function createGroup({ name, members }) {
  const res = await axios.post('/groups', { name, members }, { headers: auth() })
  return unwrap(res)
}

export async function getGroupsList() {
  const res = await axios.get('/groups', { headers: auth() })
  return unwrap(res)
}

export async function getGroup(groupId) {
  const res = await axios.get(`/groups/${encodeURIComponent(groupId)}`, { headers: auth() })
  return unwrap(res)
}

export async function setGroupName(groupId, name) {
  const res = await axios.put(
    `/groups/${encodeURIComponent(groupId)}/name`,
    { name },
    { headers: auth() }
  )
  return unwrap(res)
}

/** setGroupPhoto – preset (?preset=avatar7) or multipart (field: upload) */
export async function setGroupPhoto(groupId, { preset, file }) {
  if (preset) {
    const res = await axios.put(
      `/groups/${encodeURIComponent(groupId)}/photo?preset=${encodeURIComponent(preset)}`,
      null,
      { headers: auth() }
    )
    return unwrap(res)
  }
  if (file) {
    const form = new FormData()
    form.append('upload', file, file.name)
    const res = await axios.put(`/groups/${encodeURIComponent(groupId)}/photo`, form, {
      headers: { ...auth(), 'Content-Type': 'multipart/form-data' },
    })
    return unwrap(res)
  }
  throw new Error('Provide either { preset } or { file }')
}

/** addToGroup – supports 1 or many members; backend parses body */
export async function addToGroup(groupId, userOrUsers) {
  const body = Array.isArray(userOrUsers)
    ? { members: userOrUsers }
    : { user_id: userOrUsers }
  const res = await axios.post(`/groups/${encodeURIComponent(groupId)}/members`, body, {
    headers: auth(),
  })
  return unwrap(res)
}

/** leaveGroup – current user leaves */
export async function leaveGroup(groupId) {
  const res = await axios.delete(`/groups/${encodeURIComponent(groupId)}/members`, {
    headers: auth(),
  })
  return unwrap(res)
}

/* -------------------- compat / alias helpers -------------------- */

// 兼容旧签名：getConversationMessages({ conversation_id, limit, before })
export async function getConversationMessages(opt = {}) {
  const mapped = {
    conversationId: opt.conversation_id ?? opt.conversationId,
    limit: opt.limit,
    beforeCursor: opt.before ?? opt.beforeCursor,
    afterCursor: opt.afterCursor,
  }
  return getMessages(mapped)
}

// 兼容旧签名：sendMessageToConversation(conversationId, content, opt)
export function sendMessageToConversation(conversationId, content, opt) {
  return sendMessage({ conversationId, content, ...(opt || {}) })
}

/* Default export (方便旧代码以 `import api from ...` 使用) */
export default {
  // public
  doLogin,
  registerUser,
  liveness,
  // users
  getMyProfile,
  setMyUserName,
  setMyPhoto,
  searchUsers,
  getUserProfile,
  // conversations
  startConversation,
  getMyConversations,
  getConversation,
  sendConversationMessage,
  // messages
  getMessages,
  sendMessage,
  getMessageById,
  deleteMessage,
  forwardMessage,
  getMessageComments,
  commentMessage,
  uncommentMessage,
  // groups
  createGroup,
  getGroupsList,
  getGroup,
  setGroupName,
  setGroupPhoto,
  addToGroup,
  leaveGroup,
  // alias / compat
  getConversationMessages,
  sendMessageToConversation,
}
