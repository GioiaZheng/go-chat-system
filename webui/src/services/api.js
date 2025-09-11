import axios from "./axios";

// Helper: read local token (assignment requires using userID as Bearer token)
function getAuthHeader() {
  const token = localStorage.getItem("token");
  return token ? { Authorization: `Bearer ${token}` } : {};
}

// User authentication
export async function doLogin(name, password = "pass") {
  const res = await axios.post("/session", { name, password });
  return res;
}

// Get current user info
export async function me() {
  const res = await axios.get("/me");
  return res;
}

// Get conversations list
export async function getMyConversations() {
  const res = await axios.get("/conversations");
  return res;
}

// Search users
export async function searchUsers(query) {
  const res = await axios.get(`/users?q=${encodeURIComponent(query)}`);
  return res;
}

// Start conversation
export async function startConversation(data) {
  const res = await axios.post("/conversations", data);
  return res;
}

// Get conversation messages
export async function getConversation({ chatType, targetId }) {
  const res = await axios.get(`/conversations/${chatType}/${targetId}`);
  return res;
}

// Send message
export async function sendMessage({ chatType, targetId, content }) {
  const res = await axios.post(`/conversations/${chatType}/${targetId}`, { content });
  return res;
}

// Delete message
export async function deleteMessage(messageId) {
  const res = await axios.delete(`/messages/${messageId}`);
  return res;
}

// Comment/reply to message
export async function commentMessage(messageId, content) {
  const res = await axios.post(`/messages/${messageId}/comments`, { content });
  return res;
}

// Remove comment/reply
export async function uncommentMessage(messageId) {
  const res = await axios.delete(`/messages/${messageId}/comments`);
  return res;
}

// Forward message
export async function forwardMessage(messageId, { toUserId, toGroupId }) {
  const res = await axios.post(`/messages/${messageId}/forward`, { toUserId, toGroupId });
  return res;
}

// Get message comments
export async function getMessageComments(messageId) {
  const res = await axios.get(`/messages/${messageId}/comments`);
  return res;
}

// Get groups list
export async function listGroups() {
  const res = await axios.get("/groups");
  return res;
}

// Create group
export async function createGroup(data) {
  const res = await axios.post("/groups", data);
  return res;
}

// Set username
export async function setMyUserName(username) {
  const res = await axios.put("/me/username", { username });
  return res;
}

// Set preset avatar
export async function setMyPhotoPreset(preset) {
  const res = await axios.put("/me/photo/preset", { preset });
  return res;
}

// Upload avatar
export async function setMyPhotoUpload(file) {
  const formData = new FormData();
  formData.append("photo", file);
  const res = await axios.put("/me/photo/upload", formData, {
    headers: {
      "Content-Type": "multipart/form-data",
    },
  });
  return res;
}