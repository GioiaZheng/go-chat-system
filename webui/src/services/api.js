// src/services/api.js
// Unified API wrapper (Vue + JS)
// Depends on: src/axios.js (make sure baseURL is set or use Vite proxy to http://localhost:3000)

import axios from "../axios";

// Helper: read local token (assignment requires using userID as Bearer token)
function auth() {
  const t = localStorage.getItem("token");
  return t ? { Authorization: `Bearer ${t}` } : {};
}

// Generic request helper
async function req({ method = "GET", url, params, data, headers, responseType }) {
  const res = await axios.request({
    method,
    url,
    params,
    data,
    headers: { "Content-Type": "application/json", ...auth(), ...(headers || {}) },
    responseType,
  });
  return res.data;
}

/* -------------------- Auth -------------------- */

// doLogin (operationId: doLogin)
// Assignment spec: only requires username, but your backend still asks for password.
// To stay compatible, we send { name, password } and fallback to default "pass".
export async function doLogin(name, password) {
  const body = { name: String(name || "").trim(), password: password ?? "pass" };
  // POST /api/v1/session
  // Expected response: { code, message, data: [ { user, token } ] }
  return req({ method: "POST", url: "/api/v1/session", data: body });
}

// doRegister (not explicitly required by assignment, but useful for user creation)
export async function doRegister({ username, email, password, gender }) {
  return req({
    method: "POST",
    url: "/api/v1/register",
    data: { username, email, password, gender },
  });
}

/* -------------------- Users -------------------- */

// me (operationId: getUserInfo / getMyProfile depending on spec)
// Returns the current logged-in user's profile.
export async function me() {
  return req({ method: "GET", url: "/api/v1/users/me" });
}

// setMyUserName (operationId: setMyUserName)
// Updates the username of the current user.
export async function setMyUserName(username) {
  return req({
    method: "PUT",
    url: "/api/v1/users/set_username",
    data: { username },
  });
}

// setMyPhoto (operationId: setMyPhoto)
// Two versions supported: preset-based or file upload.

// Preset version (use ?preset=avatar6)
export async function setMyPhotoPreset(preset /* e.g. "avatar6" */) {
  return req({
    method: "PUT",
    url: "/api/v1/users/set_photo",
    params: { preset },
  });
}

// Upload version (multipart/form-data)
export async function setMyPhotoUpload(file /* File */) {
  const fd = new FormData();
  fd.append("upload", file);
  const res = await axios.request({
    method: "PUT",
    url: "/api/v1/users/set_photo",
    data: fd,
    headers: { ...auth() }, // Let browser set multipart boundary
  });
  return res.data;
}

// searchUsers (operationId: searchUsers)
// Finds users by username, name or email.
export async function searchUsers(q) {
  return req({ method: "GET", url: "/api/v1/users/search", params: { q } });
}

// getUserProfile (operationId: getUserProfile)
// Retrieves profile of another user by ID.
export async function getUserProfile(userId) {
  return req({ method: "GET", url: `/api/v1/users/profile/${encodeURIComponent(userId)}` });
}

/* -------------------- Conversations -------------------- */

// getMyConversations (operationId: getMyConversations)
// Returns list of conversations with last message info.
export async function getMyConversations() {
  return req({ method: "GET", url: "/api/v1/conversations" });
}

// startConversation (operationId: startConversation)
// Creates a new conversation (private or group).
export async function startConversation({ name, memberIds }) {
  return req({
    method: "POST",
    url: "/api/v1/conversations",
    data: { name, memberIds },
  });
}

/* -------------------- Messages -------------------- */

// getConversation (operationId: getConversation)
// Retrieve messages of a conversation by type (private/group) and targetId.
export async function getConversation({ chatType, targetId }) {
  return req({
    method: "GET",
    url: "/api/v1/messages",
    params: { chat_type: chatType, target_id: targetId },
  });
}

// sendMessage (operationId: sendMessage)
// Sends a message to a private or group chat.
export async function sendMessage({ chatType, targetId, content }) {
  return req({
    method: "POST",
    url: "/api/v1/messages",
    data: { chat_type: chatType, target_id: targetId, content },
  });
}

// getMessageById (operationId: getMessageById)
export async function getMessageById(messageId) {
  return req({ method: "GET", url: `/api/v1/messages/${encodeURIComponent(messageId)}` });
}

// deleteMessage (operationId: deleteMessage)
export async function deleteMessage(messageId) {
  return req({ method: "DELETE", url: `/api/v1/messages/${encodeURIComponent(messageId)}` });
}

// forwardMessage (operationId: forwardMessage)
export async function forwardMessage(messageId, { toUserId, toGroupId }) {
  return req({
    method: "POST",
    url: `/api/v1/messages/${encodeURIComponent(messageId)}/forward`,
    data: { toUserId, toGroupId },
  });
}

// getMessageComments (helper for commentMessage/uncommentMessage)
export async function getMessageComments(messageId) {
  return req({
    method: "GET",
    url: `/api/v1/messages/${encodeURIComponent(messageId)}/comment`,
  });
}

// commentMessage (operationId: commentMessage)
// Add a comment/reaction to a message.
export async function commentMessage(messageId, comment) {
  return req({
    method: "POST",
    url: `/api/v1/messages/${encodeURIComponent(messageId)}/comment`,
    data: { comment },
  });
}

// uncommentMessage (operationId: uncommentMessage)
// Remove a comment/reaction from a message.
export async function uncommentMessage(messageId) {
  return req({
    method: "POST",
    url: `/api/v1/messages/${encodeURIComponent(messageId)}/uncomment`,
  });
}

/* -------------------- Groups -------------------- */

// listGroups (operationId: getGroupsList)
// Returns all groups the user is a member of.
export async function listGroups() {
  return req({ method: "GET", url: "/api/v1/groups" });
}

// getGroup (operationId: getGroup)
// Retrieve detail of a group by ID.
export async function getGroup(id) {
  return req({ method: "GET", url: `/api/v1/groups/${encodeURIComponent(id)}` });
}

// createGroup (operationId: createGroup)
export async function createGroup({ name, members }) {
  return req({ method: "POST", url: "/api/v1/groups", data: { name, members } });
}

// setGroupName (operationId: setGroupName)
// Update the name of a group.
export async function setGroupName(id, name) {
  return req({
    method: "PUT",
    url: `/api/v1/groups/${encodeURIComponent(id)}/name`,
    data: { name },
  });
}

// setGroupPhoto (operationId: setGroupPhoto)
// Two versions: preset or upload

// Preset-based
export async function setGroupPhotoPreset(id, preset /* e.g. "group6" */) {
  return req({
    method: "PUT",
    url: `/api/v1/groups/${encodeURIComponent(id)}/set_photo`,
    params: { preset },
  });
}

// File upload version
export async function setGroupPhotoUpload(id, file /* File */) {
  const fd = new FormData();
  fd.append("upload", file);
  const res = await axios.request({
    method: "PUT",
    url: `/api/v1/groups/${encodeURIComponent(id)}/photo`,
    data: fd,
    headers: { ...auth() },
  });
  return res.data;
}

// addToGroup (operationId: addToGroup)
export async function addToGroup(id, members /* array of userIds */) {
  return req({
    method: "POST",
    url: `/api/v1/groups/${encodeURIComponent(id)}/members`,
    data: { members },
  });
}

// leaveGroup (operationId: leaveGroup)
export async function leaveGroup(id) {
  return req({
    method: "DELETE",
    url: `/api/v1/groups/${encodeURIComponent(id)}/members`,
  });
}
