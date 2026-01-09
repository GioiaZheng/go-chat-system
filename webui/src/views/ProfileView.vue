<!-- src/views/ProfileView.vue: Profile settings for account identity and avatar -->
<template>
  <div class="page">
    <main class="wrap">
      <div class="page-title">
        <div>
          <h2 class="title">My Profile</h2>
          <p class="subtitle">Manage how you appear across chats and groups.</p>
        </div>
      </div>

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <section v-if="me" class="profile-card">
        <header class="profile-header">
          <div class="avatar-section">
            <div
              class="avatar-container"
              tabindex="0"
              role="button"
              @click="triggerFile"
              @keydown.enter.prevent="triggerFile"
            >
              <img
                v-if="avatarPreview && !imgBroken"
                :src="avatarPreview"
                alt="Profile photo"
                class="avatar avatar-circle profile-avatar"
                @error="imgBroken = true"
              />
              <div v-else class="avatar-fallback avatar-circle profile-avatar">{{ initials }}</div>
              <div class="avatar-overlay">
                <span class="overlay-icon" aria-hidden="true">📷</span>
                <span class="overlay-text">Change photo</span>
              </div>
            </div>
            <input
              ref="fileInput"
              type="file"
              class="hidden-input"
              accept="image/*"
              @change="onFile"
            />
            <p class="avatar-hint">Profile photo is visible in chats and groups.</p>
            <div v-if="avatarDirty" class="inline-actions">
              <button
                class="btn btn-primary"
                :disabled="loading"
                @click="uploadPhoto"
              >
                <span v-if="loading && pendingAction === 'photo'" class="spinner"></span>
                Save photo
              </button>
              <button class="btn btn-secondary" :disabled="loading" @click="resetPhotoSelection">Cancel</button>
            </div>
          </div>

          <div class="identity-section">
            <p class="eyebrow">Name</p>
            <div v-if="!editingName" class="name-display">
              <h3 class="profile-name">{{ me.name || 'Name not set' }}</h3>
              <button class="icon-btn" type="button" aria-label="Edit name" @click="startNameEdit">
                <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                  <path
                    d="M4 17.25V20h2.75L17.81 8.94l-2.75-2.75L4 17.25zm3.92.75H6v-1.92l7.06-7.06 1.92 1.92L7.92 18zM20.71 7.04a1.003 1.003 0 0 0 0-1.42l-2.34-2.34a1.003 1.003 0 0 0-1.42 0l-1.83 1.83 3.76 3.76 1.83-1.83z"
                  />
                </svg>
              </button>
            </div>
            <div v-else class="name-edit">
              <input v-model.trim="newName" class="input" placeholder="Your name" />
              <div class="inline-actions">
                <button
                  class="btn btn-primary"
                  :disabled="loading || !newName"
                  @click="saveName"
                >
                  <span v-if="loading && pendingAction === 'name'" class="spinner"></span>
                  Save
                </button>
                <button class="btn btn-secondary" :disabled="loading" @click="cancelNameEdit">Cancel</button>
              </div>
            </div>
            <p class="muted">This is how you appear in chats and groups.</p>
          </div>
        </header>

        <div v-if="toastMessage" class="toast" role="status" aria-live="polite">
          <span class="checkmark">✓</span>
          <span>{{ toastMessage }}</span>
        </div>

        <section class="info-stack">
          <div class="accordion-card">
            <button
              class="accordion-toggle"
              type="button"
              :aria-expanded="accountOpen"
              @click="accountOpen = !accountOpen"
            >
              <span>Account</span>
              <span class="accordion-icon" :class="{ open: accountOpen }">▸</span>
            </button>
            <div v-if="accountOpen" class="accordion-body">
              <div class="system-row">
                <div class="muted">User ID</div>
                <div class="system-value">{{ me.id || me.userId || me.user_id || '—' }}</div>
              </div>
            </div>
          </div>
        </section>
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import { getAvatarUrl, setMyPhotoFile, setMyUserName } from '@/services/api'
import { currentUser, ensureAuthReady, isAuthenticated, refreshProfile } from '@/services/auth'

const router = useRouter()
const me = ref(null)
const newName = ref('')
const file = ref(null)
const fileInput = ref(null)
const previewUrl = ref('')
const avatarDirty = ref(false)
const loading = ref(false)
const err = ref('')
const imgBroken = ref(false)
const editingName = ref(false)
const toastMessage = ref('')
const pendingAction = ref('')
const accountOpen = ref(true)
let toastTimer = null

const avatarUrl = computed(() => getAvatarUrl(me.value || {}))
const avatarPreview = computed(() => previewUrl.value || (avatarUrl.value && !imgBroken.value ? avatarUrl.value : ''))
const initials = computed(() => ((me.value?.name || 'U')[0] || 'U').toUpperCase())

onMounted(async () => {
  await ensureAuthReady()
  if (!isAuthenticated.value) {
    router.replace('/login')
    return
  }
  loadProfile()
})

async function loadProfile() {
  err.value = ''
  try {
    await refreshProfile()
    me.value = currentUser.value
    newName.value = me.value?.name || ''
    localStorage.setItem('name', me.value?.name || '')
    localStorage.setItem('me', JSON.stringify(me.value || {}))
    window.dispatchEvent(new Event('auth:changed'))
    imgBroken.value = false
  } catch (e) {
    handleError(e, 'Failed to load profile')
  }
}

function startNameEdit() {
  if (avatarDirty.value) resetPhotoSelection()
  editingName.value = true
  toastMessage.value = ''
}

function cancelNameEdit() {
  editingName.value = false
  newName.value = me.value?.name || ''
}

async function saveName() {
  if (!newName.value) return
  loading.value = true
  pendingAction.value = 'name'
  err.value = ''
  toastMessage.value = ''
  try {
    await setMyUserName(newName.value)
    await loadProfile()
    editingName.value = false
    showToast('Name updated')
    window.dispatchEvent(new CustomEvent('conversations:reload'))
  } catch (e) {
    handleError(e, 'Failed to set name')
  } finally {
    loading.value = false
    pendingAction.value = ''
  }
}

function triggerFile() {
  fileInput.value?.click()
}

function onFile(e) {
  const selected = e.target.files?.[0] || null
  if (!selected) return
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  file.value = selected
  previewUrl.value = URL.createObjectURL(selected)
  avatarDirty.value = true
  editingName.value = false
  toastMessage.value = ''
}

function resetPhotoSelection() {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = ''
  file.value = null
  avatarDirty.value = false
}

async function uploadPhoto() {
  if (!file.value) return
  loading.value = true
  pendingAction.value = 'photo'
  err.value = ''
  toastMessage.value = ''
  try {
    await setMyPhotoFile(file.value)
    await loadProfile()
    resetPhotoSelection()
    showToast('Avatar updated')
    me.value = { ...(me.value || {}), updatedAt: Date.now() }
    if (me.value) {
      localStorage.setItem('me', JSON.stringify(me.value))
      window.dispatchEvent(new Event('auth:changed'))
    }
    window.dispatchEvent(new CustomEvent('conversations:reload'))
  } catch (e) {
    handleError(e, 'Failed to upload photo')
  } finally {
    loading.value = false
    pendingAction.value = ''
  }
}

function showToast(message) {
  toastMessage.value = message
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
    toastTimer = null
  }, 3000)
}

function handleError(e, fallback) {
  if (e?.response?.status === 401) {
    err.value = 'Unauthorized. Please login again.'
    router.push('/login')
  } else if (e?.response?.status === 409) {
    err.value = 'Name is already in use. Please choose another.'
  } else {
    err.value = e?.response?.data?.message || e?.message || fallback
  }
}
</script>

<style scoped>
.page {
  min-height: 100%;
  height: 100%;
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}
.wrap {
  flex: 1 1 auto;
  max-width: 720px;
  margin: 0 auto;
  padding: 24px;
  width: 100%;
}
.page-title {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}
.title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 4px;
}
.subtitle {
  margin: 0;
  color: #64748b;
  font-size: 0.95rem;
}

.card {
  background: #fff;
  border: 0;
  border-radius: 18px;
  padding: 20px 22px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.04);
}
.profile-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  box-shadow: 0 10px 24px rgba(2, 6, 23, 0.08);
  padding: 18px;
}
.profile-header {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 16px;
  align-items: center;
  padding: 12px 10px;
  border-radius: 12px;
  background: transparent;
}
@media (max-width: 640px) {
  .profile-header {
    grid-template-columns: 1fr;
  }
}
.avatar-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-start;
}
.avatar-container {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  outline: none;
}
.avatar-container:focus-visible {
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.4);
  border-radius: 50%;
}
.profile-avatar {
  width: 32px !important;
  height: 32px !important;
  font-size: 0.9rem;
  border: 2px solid #d9f99d;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.avatar-container:hover .profile-avatar {
  transform: scale(1.01);
  box-shadow: 0 8px 18px rgba(34, 197, 94, 0.22);
  border-color: #22c55e;
}
.avatar-fallback {
  background: #e0f7ee;
  color: #0f766e;
  font-weight: 700;
}
.avatar-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border-radius: 50%;
  background: rgba(34, 197, 94, 0.8);
  color: #ecfdf3;
  opacity: 0;
  transition: opacity 0.2s ease;
}
.avatar-container:hover .avatar-overlay,
.avatar-container:focus-visible .avatar-overlay {
  opacity: 1;
}
.overlay-icon {
  font-size: 1rem;
}
.overlay-text {
  font-size: 0.9rem;
  font-weight: 600;
}
.avatar-hint {
  margin: 0;
  color: #94a3b8;
  font-size: 0.88rem;
}
.identity-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.eyebrow {
  margin: 0;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #94a3b8;
  font-weight: 700;
}
.profile-name {
  margin: 0;
  font-size: 1.65rem;
  color: #0f172a;
  font-weight: 800;
}
.name-display {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.name-edit {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-start;
}
.muted {
  color: #94a3b8;
  margin: 0;
  font-size: 0.9rem;
}
.inline-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.icon-btn {
  border: 1px solid rgba(34, 197, 94, 0.25);
  background: rgba(34, 197, 94, 0.12);
  color: #15803d;
  border-radius: 12px;
  padding: 8px 10px;
  min-width: 40px;
  min-height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
  outline: none;
}
.icon-btn svg {
  width: 20px;
  height: 20px;
  fill: currentColor;
}
.icon-btn:hover {
  background: rgba(34, 197, 94, 0.22);
  box-shadow: 0 10px 18px rgba(34, 197, 94, 0.3);
  transform: translateY(-1px);
}
.icon-btn:focus-visible {
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.3);
}
.icon-btn:active {
  transform: translateY(0);
}
.input {
  border: 1px solid #cbd5e1;
  border-radius: var(--radius-control);
  padding: 0.55rem 0.75rem;
  outline: none;
  background: #fff;
  min-width: 240px;
}
.input:focus {
  border-color: #22c55e;
  box-shadow: 0 0 0 0.18rem rgba(34, 197, 94, 0.2);
}
.btn {
  border: 0;
  border-radius: var(--radius-control);
  padding: 0.55rem 0.9rem;
  white-space: nowrap;
  font-weight: 700;
  cursor: pointer;
}
.btn-primary {
  color: #fff;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 60%, #0f766e 120%);
  box-shadow: 0 10px 20px rgba(34, 197, 94, 0.3);
}
.btn-primary:not(:disabled):hover {
  box-shadow: 0 12px 24px rgba(34, 197, 94, 0.35);
  transform: translateY(-1px);
}
.btn-secondary {
  background: #fff;
  color: #0f172a;
  border: 1px solid #cbd5e1;
}
.btn-secondary:not(:disabled):hover {
  border-color: #22c55e;
  color: #166534;
  box-shadow: 0 8px 16px rgba(15, 23, 42, 0.06);
}
.btn:focus-visible {
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.28);
  outline: none;
}
.btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}
.spinner {
  display: inline-block;
  width: 1em;
  height: 1em;
  margin-right: 0.4em;
  border: 2px solid rgba(255, 255, 255, 0.6);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  vertical-align: -2px;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
.hidden-input {
  display: none;
}
.info-stack {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px;
  box-shadow: 0 6px 16px rgba(2, 6, 23, 0.06);
}
.accordion-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.accordion-toggle {
  border: 0;
  background: transparent;
  color: #475569;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  cursor: pointer;
  padding: 6px 0;
}
.accordion-icon {
  transition: transform 0.2s ease;
}
.accordion-icon.open {
  transform: rotate(90deg);
}
.accordion-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.system-row {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.system-value {
  word-break: break-all;
  color: #0f172a;
  font-weight: 600;
  font-size: 0.95rem;
}
.toast {
  position: fixed;
  top: 20px;
  right: 20px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  background: #ecfdf3;
  border: 1px solid #bbf7d0;
  color: #166534;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 600;
  font-size: 0.95rem;
  box-shadow: 0 12px 26px rgba(34, 197, 94, 0.18);
  z-index: 10;
}
.checkmark {
  font-size: 1.1rem;
}
</style>