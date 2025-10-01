// src/services/api.js
// Full frontend API client that matches ALL routes registered in your backend.
// It does NOT modify axios.js. Every call injects Authorization locally.
// Return values unwrap { code, data } => return data only.

import axios from './axios';

/* -------------------- helpers -------------------- */

function auth() {
  const token = localStorage.getItem('token');
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function unwrap(res) {
  if (res?.data && typeof res.data === 'object' && 'data' in res.data) return res.data.data;
  return res.data;
}

/* -------------------- public / utility -------------------- */

/** POST /session  -- doLogin (PDF “simplified login”) */
export async function doLogin(name) {
  const res = await axios.post('/session', { name });
  const payload = unwrap(res);
  const identifier = payload?.identifier || payload?.id || payload;
  if (!identifier) throw new Error('Login response missing identifier');
  localStorage.setItem('token', identifier);
  return identifier;
}

/** POST /register  -- optional helper (kept since router registers it) */
export async function registerUser({ name, email } = {}) {
  const res = await axios.post('/register', { name, email });
  return unwrap(res);
}

/** GET /liveness */
export async function liveness() {
  const res = await axios.get('/liveness');
  return unwrap(res);
}

/* -------------------- users -------------------- */
/* Canonical paths per api-handler.go:
   PUT /users/set_username
   PUT /users/set_photo
   GET /users/me
   GET /users/search
   GET /users/profile/:user_id
*/

export async function getMyProfile() {
  const res = await axios.get('/users/me', { headers: auth() });
  return unwrap(res);
}

export async function setMyUserName(username) {
  const res = await axios.put('/users/set_username', { username }, { headers: auth() });
  return unwrap(res);
}

/** setMyPhoto – supports preset (?preset=avatar7) or multipart upload (field: upload) */
export async function setMyPhoto({ preset, file }) {
  if (preset) {
    const res = await axios.put(`/users/set_photo?preset=${encodeURIComponent(preset)}`, null, {
      headers: auth(),
    });
    return unwrap(res);
  }
  if (file) {
    const form = new FormData();
    form.append('upload', file, file.name);
    const res = await axios.put('/users/set_photo', form, {
      headers: { ...auth(), 'Content-Type': 'multipart/form-data' },
    });
    return unwrap(res);
  }
  throw new Error('Provide either { preset } or { file }');
}

export async function searchUsers(q) {
  const res = await axios.get('/users/search', { params: { q }, headers: auth() });
  return unwrap(res);
}

export async function getUserProfile(userIdOrName) {
  // Prefer path param by id; if you only have username, backend also accepts ?username=
  if (/^[a-f0-9-]{8,}$/i.test(userIdOrName)) {
    const res = await axios.get(`/users/profile/${encodeURIComponent(userIdOrName)}`, {
      headers: auth(),
    });
    return unwrap(res);
  }
  const res = await axios.get('/users/profile/0', {
    // dummy path id; backend resolves via ?username
    params: { username: userIdOrName },
    headers: auth(),
  });
  return unwrap(res);
}

/* -------------------- conversations -------------------- */
/* Canonical:
   POST /conversations           (startConversation)
   GET  /conversations           (getMyConversations)
   Aliases:
   GET  /conversations/:conversationId                -> maps to GET /messages?conversation_id=...
   POST /conversations/:conversationId/messages       -> maps to POST /messages (with conversation_id)
*/

export async function startConversation({ user_id, group_id } = {}) {
  // Either start a private (to user) or ensure a group conversation
  const res = await axios.post('/conversations', { user_id, group_id }, { headers: auth() });
  return unwrap(res);
}

export async function getMyConversations() {
  const res = await axios.get('/conversations', { headers: auth() });
  return unwrap(res);
}

/** Equivalent to getConversation via alias route */
export async function getConversation(conversationId) {
  const res = await axios.get(`/conversations/${encodeURIComponent(conversationId)}`, {
    headers: auth(),
  });
  return unwrap(res);
}

/** Send inside a conversation via alias route */
export async function sendConversationMessage(conversationId, content, { type = 'text', status = 'sent' } = {}) {
  const res = await axios.post(
    `/conversations/${encodeURIComponent(conversationId)}/messages`,
    { content, type, status },
    { headers: auth() }
  );
  return unwrap(res);
}

/* -------------------- messages -------------------- */
/* Canonical:
   GET    /messages                      (supports ?conversation_id=...)
   POST   /messages
   GET    /messages/:id
   DELETE /messages/:id
   POST   /messages/:id/forward
   GET    /messages/:id/comment          (list comments)
   POST   /messages/:id/comment          (add comment)
   POST   /messages/:id/uncomment        (remove my comment)
*/

export async function getMessages({ conversation_id, limit, before } = {}) {
  const res = await axios.get('/messages', {
    params: { conversation_id, limit, before },
    headers: auth(),
  });
  return unwrap(res);
}

export async function sendMessage({ content, to_user_id, conversation_id, type = 'text', status = 'sent' }) {
  const res = await axios.post(
    '/messages',
    { content, to_user_id, conversation_id, type, status },
    { headers: auth() }
  );
  return unwrap(res);
}

export async function getMessageById(messageId) {
  const res = await axios.get(`/messages/${encodeURIComponent(messageId)}`, { headers: auth() });
  return unwrap(res);
}

export async function deleteMessage(messageId) {
  const res = await axios.delete(`/messages/${encodeURIComponent(messageId)}`, { headers: auth() });
  return unwrap(res);
}

export async function forwardMessage(messageId, { to_user_id, to_group_id }) {
  const res = await axios.post(
    `/messages/${encodeURIComponent(messageId)}/forward`,
    { to_user_id, to_group_id },
    { headers: auth() }
  );
  return unwrap(res);
}

export async function getMessageComments(messageId) {
  const res = await axios.get(`/messages/${encodeURIComponent(messageId)}/comment`, { headers: auth() });
  return unwrap(res);
}

export async function commentMessage(messageId, comment) {
  const res = await axios.post(
    `/messages/${encodeURIComponent(messageId)}/comment`,
    { comment },
    { headers: auth() }
  );
  return unwrap(res);
}

export async function uncommentMessage(messageId) {
  const res = await axios.post(`/messages/${encodeURIComponent(messageId)}/uncomment`, null, {
    headers: auth(),
  });
  return unwrap(res);
}

/* -------------------- groups -------------------- */
/* Canonical:
   POST   /groups
   GET    /groups
   GET    /groups/:id
   PUT    /groups/:id/name
   PUT    /groups/:id/photo
   POST   /groups/:id/members      (add members)
   DELETE /groups/:id/members      (leave current user)
*/

export async function createGroup({ name, members }) {
  const res = await axios.post('/groups', { name, members }, { headers: auth() });
  return unwrap(res);
}

export async function getGroupsList() {
  const res = await axios.get('/groups', { headers: auth() });
  return unwrap(res);
}

export async function getGroup(groupId) {
  const res = await axios.get(`/groups/${encodeURIComponent(groupId)}`, { headers: auth() });
  return unwrap(res);
}

export async function setGroupName(groupId, name) {
  const res = await axios.put(
    `/groups/${encodeURIComponent(groupId)}/name`,
    { name },
    { headers: auth() }
  );
  return unwrap(res);
}

/** setGroupPhoto – preset (?preset=avatar7) or multipart (field: upload) */
export async function setGroupPhoto(groupId, { preset, file }) {
  if (preset) {
    const res = await axios.put(
      `/groups/${encodeURIComponent(groupId)}/photo?preset=${encodeURIComponent(preset)}`,
      null,
      { headers: auth() }
    );
    return unwrap(res);
  }
  if (file) {
    const form = new FormData();
    form.append('upload', file, file.name);
    const res = await axios.put(`/groups/${encodeURIComponent(groupId)}/photo`, form, {
      headers: { ...auth(), 'Content-Type': 'multipart/form-data' },
    });
    return unwrap(res);
  }
  throw new Error('Provide either { preset } or { file }');
}

/** addToGroup – supports 1 or多成员；后端按 body 解析 */
export async function addToGroup(groupId, userOrUsers) {
  const body = Array.isArray(userOrUsers)
    ? { members: userOrUsers }
    : { user_id: userOrUsers };
  const res = await axios.post(`/groups/${encodeURIComponent(groupId)}/members`, body, {
    headers: auth(),
  });
  return unwrap(res);
}

/** leaveGroup – current user leaves */
export async function leaveGroup(groupId) {
  const res = await axios.delete(`/groups/${encodeURIComponent(groupId)}/members`, {
    headers: auth(),
  });
  return unwrap(res);
}

/* -------------------- compat / alias helpers -------------------- */
/* 这些是便捷别名；如果你组件里更喜欢“会话口径”，可以用它们 */

export const getConversationMessages = getMessages; // with { conversation_id }
export function sendMessageToConversation(conversationId, content, opt) {
  return sendMessage({ conversation_id: conversationId, content, ...(opt || {}) });
}
