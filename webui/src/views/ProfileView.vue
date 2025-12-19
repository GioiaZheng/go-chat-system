<!-- src/views/ProfileView.vue: Profile settings for account identity and avatar. -->
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

      <section v-if="me" class="card profile-card">
        <header class="profile-header">
          <div class="avatar-wrap">
            <img
              v-if="avatarUrl && !imgBroken"
              :src="avatarUrl"
              alt="avatar"
              class="avatar avatar-circle profile-avatar"
              @error="imgBroken = true"
            />
            <div v-else class="avatar-fallback avatar-circle profile-avatar">{{ initials }}</div>
          </div>

          <div class="profile-meta">
            <p class="eyebrow">Account</p>
            <h3 class="profile-name">{{ me.name || 'Unnamed' }}</h3>
            <div class="profile-id">
              <span class="id-label">User ID</span>
              <span class="id-value">{{ me.id || me.userId || me.user_id || '—' }}</span>
            </div>
            <div class="profile-id">
              <span class="id-label">Username</span>
              <span class="id-value">{{ me.username || me.handle || '—' }}</span>
            </div>
          </div>

          <div class="profile-status">
            <div class="status-card">
              <p class="status-label">Profile photo</p>
              <p class="status-value">{{ avatarUrl && !imgBroken ? 'Custom image' : 'Default badge' }}</p>
              <p class="status-hint">Keep your chats recognizable.</p>
            </div>
            <div class="status-card">
              <p class="status-label">Display name</p>
              <p class="status-value">{{ me.name || '(unset)' }}</p>
              <p class="status-hint">Shown in private and group chats.</p>
            </div>
          </div>
        </header>

        <div class="panel-grid">
          <div class="panel">
            <h4 class="panel-title">Update display name</h4>
            <p class="panel-subtitle">Choose a friendly name for your profile.</p>
            <div class="hstack">
              <input v-model.trim="newName" class="input" placeholder="New display name" />
              <button class="btn" :disabled="loading || !newName" @click="saveName">
                <span v-if="loading" class="spinner"></span>
                Save
              </button>
            </div>
          </div>

          <div class="panel">
            <h4 class="panel-title">Update profile photo</h4>
            <p class="panel-subtitle">Upload a square image for best results.</p>
            <div class="hstack">
              <input type="file" accept="image/*" @change="onFile" />
              <button class="btn" :disabled="loading || !file" @click="uploadPhoto">
                <span v-if="loading" class="spinner"></span>
                Upload
              </button>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import { setMyUserName, setMyPhotoFile, getAvatarUrl } from '@/services/api'
import { ensureAuthReady, isAuthenticated, refreshProfile, currentUser } from '@/services/auth'

const router = useRouter()
const me = ref(null)
const newName = ref('')
const file = ref(null)
const loading = ref(false)
const err = ref('')
const imgBroken = ref(false)

// Resolve avatar URLs through getAvatarUrl to avoid relative paths.
const avatarUrl = computed(() => getAvatarUrl(me.value || {}))

const initials = computed(() => ((me.value?.name || me.value?.username || 'U')[0] || 'U').toUpperCase())

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

async function saveName() {
  if (!newName.value) return
  loading.value = true
  err.value = ''
  try {
    await setMyUserName(newName.value)
    await loadProfile()
  } catch (e) {
    handleError(e, 'Failed to set name')
  } finally {
    loading.value = false
  }
}

function onFile(e) {
  file.value = e.target.files?.[0] || null
}

async function uploadPhoto() {
  if (!file.value) return
  loading.value = true
  err.value = ''
  try {
    await setMyPhotoFile(file.value) // Multipart field name is "upload"; handled in the API layer.
    file.value = null
    await loadProfile()
    me.value = { ...(me.value || {}), updatedAt: Date.now() }
    if (me.value) {
      localStorage.setItem('me', JSON.stringify(me.value))
      window.dispatchEvent(new Event('auth:changed'))
    }
  } catch (e) {
    handleError(e, 'Failed to upload photo')
  } finally {
    loading.value = false
  }
}

function handleError(e, fallback) {
  if (e?.response?.status === 401) {
    err.value = 'Unauthorized. Please login again.'
    router.push('/login')
  } else if (
    e?.response?.status === 409 ||
    /exists|taken|already/i.test(e?.response?.data?.message)
  ) {
    err.value = 'Username already taken. Please choose another.'
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
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
}
.wrap {
  flex: 1 1 auto;
  max-width: 900px;
  margin: 0 auto;
  padding: 24px;
  width: 100%;
}
.page-title {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
}
.title {
  font-size: 1.65rem;
  font-weight: 800;
  color: #0f172a;
  margin: 0 0 6px;
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
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.08);
}
.profile-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
}
.profile-header {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
  padding: 18px;
  border-radius: 16px;
  background: linear-gradient(135deg, #eefbf3, #eef2ff);
}
.avatar-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
}
.profile-avatar {
  width: 72px !important;
  height: 72px !important;
  font-size: 1.4rem;
  border: 2px solid #bbf7d0;
}
.profile-meta {
  flex: 1 1 240px;
  min-width: 200px;
}
.eyebrow {
  margin: 0 0 6px;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #94a3b8;
  font-weight: 700;
}
.profile-name {
  margin: 0 0 10px;
  font-size: 1.3rem;
  color: #0f172a;
  font-weight: 700;
}
.profile-id {
  display: flex;
  gap: 10px;
  align-items: center;
  font-size: 0.92rem;
  color: #475569;
  margin-bottom: 6px;
}
.id-label {
  min-width: 90px;
  color: #64748b;
  font-weight: 600;
}
.id-value {
  color: #0f172a;
  font-weight: 600;
  word-break: break-all;
}
.profile-status {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  min-width: 220px;
  flex: 1 1 260px;
}

.status-card {
  background: #fff;
  border-radius: 14px;
  padding: 12px 14px;
  box-shadow: 0 10px 20px rgba(15, 23, 42, 0.06);
}
.status-label {
  margin: 0;
  color: #64748b;
  font-size: 0.85rem;
}
.status-value {
  margin: 4px 0;
  font-weight: 700;
  color: #0f172a;
}
.status-hint {
  margin: 0;
  color: #94a3b8;
  font-size: 0.85rem;
}

.panel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 16px;
}
.panel {
  background: #f8fafc;
  border-radius: 16px;
  padding: 16px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.panel-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: #0f172a;
}
.panel-subtitle {
  margin: 0;
  color: #64748b;
  font-size: 0.88rem;
}
.hstack {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.input {
  flex: 1 1 220px;
  border: 1px solid #cbd5e1;
  border-radius: var(--radius-control);
  padding: 0.55rem 0.75rem;
  outline: none;
  background: #fff;
}

.input:focus {
  border-color: #22c55e;
  box-shadow: 0 0 0 0.2rem rgba(34, 197, 94, 0.15);
}

.btn {
  border: 0;
  border-radius: var(--radius-control);
  color: #fff;
  padding: 0.55rem 0.9rem;
  white-space: nowrap;
  background-image: linear-gradient(135deg, #22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow: 0 0.6rem 1.4rem rgba(34, 197, 94, 0.25);
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
</style>
