// src/services/api.js
import axios from './axios'

/* -------------------- helpers -------------------- */

// 读 token（兼容 localStorage / sessionStorage）
const readToken = () =>
  localStorage.getItem('token') || sessionStorage.getItem('authToken') || ''

// 是否已登录
export function isAuthed() {
  return !!readToken()
}

// 统一在请求里加 Authorization 头
function withAuthConfig(cfg = {}) {
  const token = readToken()
  const headers = { ...(cfg?.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  return { ...(cfg || {}), headers }
}

// 只剥掉一层 { code, message, data }，不破坏内部结构
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




// 快捷请求
const get   = (url, cfg)       => axios.get(url,    withAuthConfig(cfg))
const del   = (url, cfg)       => axios.delete(url, withAuthConfig(cfg))
const post  = (url, data, cfg) => axios.post(url,   data, withAuthConfig(cfg))
const put   = (url, data, cfg) => axios.put(url,    data, withAuthConfig(cfg))

/* -------------------- base URL helpers -------------------- */

// axios.baseURL > window.__API_URL__ > import.meta.env.VITE_API_BASE_URL
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

/* -------------------- public / utility -------------------- */

export async function doLogin(name) {
  // 你的后端 /session 简化登录（只用 name）
  const res = await axios.post('/session', { name })

  // axios 的响应：res.data = { code, message, data: { user, token } }
  const root = res && res.data ? res.data : res

  // 尽可能多地尝试几种包裹方式：
  const token =
    root?.data?.token ||   // 正常情况：{code, message, data:{user, token}}
    root?.token           // 万一后端直接返回 {user, token}
  
  if (!token) {
    console.error('Login response payload =', root)   // 方便你在控制台看
    throw new Error('Login response missing token')
  }

  // 存 token
  localStorage.setItem('token', token)
  sessionStorage.setItem('authToken', token)

  try {
    window.dispatchEvent(new Event('auth:changed'))
  } catch {}

  return token
}


export async function liveness() { return unwrap(await get('/liveness')) }
export async function healthz()  { return unwrap(await get('/healthz')) }

/* -------------------- logout helper -------------------- */

/**
 * Logout: clear all tokens and cached info.
 * (前端清除登录状态，后端无需处理)
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

/* -------------------- users -------------------- */

export async function getMyProfile() {
  const u = unwrap(await get('/users/me')) || {}
  // 归一化到前端常用键
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

/** 头像上传（字段名 'upload'） */
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

/** 统一头像 URL + 缓存破坏参数 */
export function getAvatarUrl(userLike) {
  const u = userLike || {}
  const raw =
    u.avatarUri || u.avatar_uri || u.avatar_url || u.avatar || u.photo_url || u.photo || ''
  const url = raw ? absUrl(raw) : ''
  const v = encodeURIComponent(u.updatedAt || u.updated_at || Date.now())
  return url ? `${url}${url.includes('?') ? '&' : '?'}v=${v}` : ''
}

/** 搜索用户（/users/search?q） */
export async function searchUsers(q = '') {
  const v = unwrap(await get('/users/search', { params: { q } }))
  return Array.isArray(v) ? v : (v?.items ?? v?.users ?? v?.list ?? [])
}

/** 兼容 ContactsView：用 search 聚合成联系人列表 */
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
  // 过滤掉自己
  let meId = ''
  try { const me = await getMyProfile(); meId = String(me.id || '') } catch {}
  return Array.from(uniq.values()).filter(u => String(u?.id ?? '') !== meId)
}

/* -------------------- conversations helpers -------------------- */

export function isGroupConversation(c) {
  return c?.type === 'group'
}


/** 私聊标题 = 对方名字；群聊 = 群名 */
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


/** 会话头像：群=群头像；私聊=对方头像 */
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



/* -------------------- conversations -------------------- */

export async function startConversation(payload = {}) {
  // OpenAPI：POST /conversations
  // 接受多种键名并归一化：name / memberIds / members / user_id / userId
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
    name: 'private',   // ✔ 必须加这个！
    memberIds: [id]
  }))
}



export async function getMyConversations() {
  return unwrap(await get('/conversations'))
}

// 会话成员：优先 /conversations/:id/members；兜底用 groups 反查
export async function getConversationMembers(conversationId) {
  const id = String(conversationId || '').trim();
  if (!id) throw new Error('conversationId required');

  // 1) 直连 conversations 成员端点（将来如果后端补上，这里优先命中）
  try {
    const v = unwrap(await get(`/conversations/${encodeURIComponent(id)}/members`));
    return Array.isArray(v) ? v : (v?.items ?? v?.members ?? v?.list ?? []);
  } catch {}

  // 2) 兜底：在 /groups 里按 conversationId 反查群，再取群成员
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

/** 后端的 DELETE /groups/:id/members：自退。
 * 如果需要“移除某个成员”，后端应支持读取 body {userId}；
 * 这里按此约定传递，若未实现，前端请隐藏“移除成员”按钮或在后端补路由。
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

export async function sendMessage({ conversationId, content, type = 'text', replyTo } = {}) {
  const body = {
    conversationId: String(conversationId || ''),
    content: String(content || ''),
    type,
  }

  // 支持回复字段：兼容可能的键名
  if (replyTo) {
    body.replyTo = replyTo
    body.replyToId = replyTo
    body.reply_to_id = replyTo
  }

  return unwrap(await post('/messages', body))
}

export async function sendImageMessage({ conversationId, file }) {
  if (!file) throw new Error('No file selected')

  // 1) 先上传文件
  const form = new FormData()
  form.append('upload', file, file.name)

  const r = await post('/messages/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

  const fileUrl  = r?.fileUrl  || r?.file_url  || ''
  const filename = r?.filename || file.name

  if (!fileUrl) {
    console.error('upload response =', r)
    throw new Error('Upload failed: no fileUrl')
  }

  // 2) 再发一条 type=image 的普通消息
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

/* ---- 评论 / 回复 / 表情 ---- */

// 读取评论列表
export async function getMessageComments(id) {
  const mid = encodeURIComponent(String(id))
  return unwrap(await get(`/messages/${mid}/comment`))
}

// 写评论
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



// 删除评论：后端不读取 body，只需要 POST 即可
export async function uncommentMessage(id) {
  const mid = encodeURIComponent(String(id))
  return unwrap(await post(`/messages/${mid}/uncomment`))
}

// 语义别名
export const replyToMessage     = commentMessage
export const getMessageReplies  = getMessageComments
export const removeMessageReply = uncommentMessage

// 表情：将内容写成 :react:emoji，并标记类型为 emoji
export async function reactToMessage(id, emoji) {
  return commentMessage(id, { type: 'emoji', content: emoji })
}

// 删除表情：后端目前无法删除指定 emoji，因此全删
export async function unreactToMessage(id, emoji) {
  return uncommentMessage(id)
}

/* ---- 已读对勾（0=排队 1=已发 2=已达 3=已读；非本人发的返回 -1） ---- */
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
