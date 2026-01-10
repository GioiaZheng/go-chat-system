<!-- src/views/ProfileView.vue: Profile settings for account identity and avatar -->
<template>
  <div class="page">
    <main class="wrap">

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <div v-if="showWelcomeBanner" class="welcome-banner">
        <div>
          <strong>Welcome!</strong> You can customize your profile here.
        </div>
        <button class="banner-dismiss" type="button" @click="dismissWelcome">Got it</button>
      </div>

      <section v-if="me" class="profile-card">
        <header class="profile-header">
          <div class="avatar-section">
            <AvatarUpload
              :src="avatarPreview && !imgBroken ? avatarPreview : ''"
              :fallback-text="initials"
              overlay-text="Change photo"
              alt="Profile photo"
              :size="72"
              :disabled="loading"
              @select="onAvatarSelect"
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
            <button v-if="!editingName" class="name-display" type="button" @click="startNameEdit">
              <h3 class="profile-name">{{ me.name || 'Name not set' }}</h3>
              <span class="icon-btn" aria-hidden="true">
                <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                  <path
                    d="M4 17.25V20h2.75L17.81 8.94l-2.75-2.75L4 17.25zm3.92.75H6v-1.92l7.06-7.06 1.92 1.92L7.92 18zM20.71 7.04a1.003 1.003 0 0 0 0-1.42l-2.34-2.34a1.003 1.003 0 0 0-1.42 0l-1.83 1.83 3.76 3.76 1.83-1.83z"
                  />
                </svg>
              </span>
            </button>
            <div v-else class="name-edit">
              <input v-model.trim="newName" class="input" placeholder="Your name" />
              <div class="inline-actions inline-actions-right">
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
          <div class="account-card">
            <div class="account-header">Account</div>
            <div class="account-body">
              <div class="system-row">
                <div class="muted">User ID</div>
                <div class="system-value">
                  <span class="system-value-text" :title="userIdFull || ''">{{ userIdShort }}</span>
                  <button
                    v-if="canCopyUserId"
                    class="copy-btn"
                    type="button"
                    aria-label="Copy User ID"
                    @click="copyUserId"
                  >
                    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                      <path
                        d="M9 9h10v10H9zM5 5h10v2H7v8H5z"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.6"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                  </button>
                </div>
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
import AvatarUpload from '@/components/AvatarUpload.vue'
import { getAvatarUrl, setMyPhotoFile, setMyUserName } from '@/services/api'
import { currentUser, ensureAuthReady, isAuthenticated, refreshProfile } from '@/services/auth'

const router = useRouter()
const me = ref(null)
const newName = ref('')
const file = ref(null)
const previewUrl = ref('')
const avatarDirty = ref(false)
const loading = ref(false)
const err = ref('')
const imgBroken = ref(false)
const editingName = ref(false)
const toastMessage = ref('')
const pendingAction = ref('')
const showWelcomeBanner = ref(false)
let toastTimer = null

const avatarUrl = computed(() => getAvatarUrl(me.value || {}))
const avatarPreview = computed(() => previewUrl.value || (avatarUrl.value && !imgBroken.value ? avatarUrl.value : ''))
const initials = computed(() => ((me.value?.name || 'U')[0] || 'U').toUpperCase())
const userIdFull = computed(() => String(me.value?.id || me.value?.userId || me.value?.user_id || ''))
const userIdShort = computed(() => {
  const value = userIdFull.value
  if (!value) return '—'
  if (value.length <= 16) return value
  return `${value.slice(0, 6)}…${value.slice(-4)}`
})
const canCopyUserId = computed(() => !!userIdFull.value)

onMounted(async () => {
  await ensureAuthReady()
  if (!isAuthenticated.value) {
    router.replace('/login')
    return
  }
  if (typeof window !== 'undefined') {
    const seen = localStorage.getItem('profile-welcome-seen')
    if (!seen) {
      showWelcomeBanner.value = true
      localStorage.setItem('profile-welcome-seen', 'true')
    }
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
    showToast('Profile updated')
    window.dispatchEvent(new CustomEvent('conversations:reload'))
  } catch (e) {
    handleError(e, 'Failed to set name')
  } finally {
    loading.value = false
    pendingAction.value = ''
  }
}

function onAvatarSelect(selected) {
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
    showToast('Profile updated')
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

function dismissWelcome() {
  showWelcomeBanner.value = false
}

async function copyUserId() {
  const value = userIdFull.value
  if (!value) return
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(value)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = value
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    showToast('User ID copied')
  } catch (e) {
    handleError(e, 'Failed to copy User ID')
  }
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
  max-width: 760px;
  margin: 0;
  padding: 24px 16px;
  width: 100%;
  align-self: flex-start;
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
  padding: 16px;
}
.profile-header {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 16px;
  align-items: center;
  padding: 8px 10px;
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
  display: inline-flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  border: 0;
  background: transparent;
  padding: 6px 8px;
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.2s ease, box-shadow 0.2s ease;
}
.name-display:hover {
  background: rgba(34, 197, 94, 0.08);
}
.name-display:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.3);
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
.inline-actions-right {
  align-self: flex-end;
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
.info-stack {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 12px;
  box-shadow: 0 6px 16px rgba(2, 6, 23, 0.06);
}
.account-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.account-header {
  color: #475569;
  font-weight: 700;
  padding: 6px 0;
}
.account-body {
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
  display: flex;
  align-items: center;
  gap: 8px;
  color: #0f172a;
  font-weight: 600;
  font-size: 0.95rem;
}
.system-value-text {
  max-width: 260px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.copy-btn {
  border: 1px solid #cbd5e1;
  background: #fff;
  color: #334155;
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease, box-shadow 0.2s ease;
}
.copy-btn:hover {
  border-color: #22c55e;
  color: #166534;
  box-shadow: 0 6px 16px rgba(34, 197, 94, 0.2);
}
.copy-btn svg {
  width: 16px;
  height: 16px;
}
.welcome-banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #166534;
  padding: 10px 14px;
  border-radius: 12px;
  font-weight: 600;
  margin-bottom: 12px;
}
.banner-dismiss {
  border: 1px solid #86efac;
  background: #ffffff;
  color: #166534;
  padding: 6px 10px;
  border-radius: 999px;
  font-weight: 700;
  cursor: pointer;
}
.banner-dismiss:hover {
  border-color: #22c55e;
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
