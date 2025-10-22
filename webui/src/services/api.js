// src/services/api.js
// Unified API client aligned with backend RegisterRoutes().
// - Uses canonical /messages & /groups endpoints (no 404 aliases)
// - Automatically attaches Authorization header if token exists
// - unwraps { data } payloads for simpler usage

import axios from './axios'

/* -------------------- helpers -------------------- */

function auth() {
  const token = localStorage.getItem('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

function unwrap(res) {
  if (res?.data && typeof res.data === 'object' && 'data' in res.data) {
    return res.data.data
  }
  return res.data
}

/* -------------------- public / utility -------------------- */

/** POST /session — simplified login */
export async function doLogin(name) {
  const res = await axios.post('/session', { name })
  const payload = unwrap(res)
  const identifier =
    payload?.identifier ?? payload?.token ?? payload?.id ?? payload?.user_id ?? payload

  if (!identifier) throw new Error('Login response missing identifier')

  localStorage.setItem('token', identifier)
  sessionStorage.setItem('authToken', identifier)
  return identifier
}

/** POST /register */
export async function registerUser({ name, email } = {}) {
  const res = await axios.post('/register', { name, email })
  return unwrap(res)
}

/** GET /liveness */
export async function liveness() {
  const res = await axios.get('/liveness')
  return unwrap(res)
}

/** GET /healthz */
export async function healthz() {
  const res = await axios.get('/healthz')
  return unwrap(res)
}

/* -------------------- users -------------------- */

export async function getMyProfile() {
  const res = await axios.get('/users/me', { headers: auth() })
  return unwrap(res)
}

export async function setMyUserName(username) {
  const body = { username, name: username }
  const res = await axios.put('/users/set_username', body, { headers: auth() })
  return unwrap(res)
}

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
  if (/^[a-f0-9-]{8,}$/i.test(userIdOrName)) {
    const res = await axios.get(`/users/profile/${encodeURIComponent(userIdOrName)}`, {
      headers: auth(),
    })
    return unwrap(res)
  }
  const res = await axios.get('/users/profile/0', {
    params: { username: userIdOrName },
    headers: auth(),
  })
  return unwrap(res)
}

/* -------------------- conversations -------------------- */

export async function startConversation({ user_id, group_id } = {}) {
  const res = await axios.post('/conversations', { user_id, group_id }, { headers: auth() })
  return unwrap(res)
}

export async function getMyConversations() {
  const res = await axios.get('/conversations', { headers: auth() })
  return unwrap(res)
}

/** 兼容旧函数：内部改走 /messages?conversationId=... */
export async function getConversation(conversationId, { limit, beforeCursor, afterCursor } = {}) {
  return getMessages({ conversationId, limit, beforeCursor, afterCursor })
}

/** 兼容旧函数：内部改走 POST /messages */
export async function sendConversationMessage(
  conversationId,
  content,
  { type = 'text', status = 'sent' } = {}
) {
  return sendMessage({ conversationId, content, type, status })
}

/* -------------------- messages -------------------- */

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

export async function addToGroup(groupId, userOrUsers) {
  const body = Array.isArray(userOrUsers)
    ? { members: userOrUsers }
    : { user_id: userOrUsers }
  const res = await axios.post(`/groups/${encodeURIComponent(groupId)}/members`, body, {
    headers: auth(),
  })
  return unwrap(res)
}

export async function leaveGroup(groupId) {
  const res = await axios.delete(`/groups/${encodeURIComponent(groupId)}/members`, {
    headers: auth(),
  })
  return unwrap(res)
}

/* -------------------- compat / alias helpers -------------------- */

export async function getConversationMessages(opt = {}) {
  const mapped = {
    conversationId: opt.conversation_id ?? opt.conversationId,
    limit: opt.limit,
    beforeCursor: opt.before ?? opt.beforeCursor,
    afterCursor: opt.afterCursor,
  }
  return getMessages(mapped)
}

export function sendMessageToConversation(conversationId, content, opt) {
  return sendMessage({ conversationId, content, ...(opt || {}) })
}

/* -------------------- export -------------------- */

export default {
  doLogin,
  registerUser,
  liveness,
  healthz,
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
  // alias
  getConversationMessages,
  sendMessageToConversation,
}
