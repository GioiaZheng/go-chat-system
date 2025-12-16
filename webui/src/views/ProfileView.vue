<!-- src/views/ProfileView.vue: Profile settings for account identity and avatar. -->
<template>
  <div class="page">
    <main class="wrap">
      <h2 class="title">My Profile</h2>

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <section v-if="me" class="card">
        <div class="head">
          <img
            v-if="avatarUrl && !imgBroken"
            :src="avatarUrl"
            alt="avatar"
            class="avatar"
            @error="imgBroken = true"
          />
          <div v-else class="avatar placeholder">{{ initials }}</div>

          <div class="idblock">
            <div class="row">
              <span class="key">id:</span>
              <span class="val">{{ me.id }}</span>
            </div>
            <div class="row">
              <span class="key">name:</span>
              <span class="val">{{ me.name || '(unset)' }}</span>
            </div>
          </div>
        </div>


        <div class="summary-grid">
          <div class="summary-card">
            <p class="summary-label">Display name</p>
            <p class="summary-value">{{ me.name || '(unset)' }}</p>
            <p class="summary-hint">Shown to friends across private and group chats.</p>
          </div>
          <div class="summary-card">
            <p class="summary-label">Account ID</p>
            <p class="summary-value">{{ me.id }}</p>
            <p class="summary-hint">Share this if someone needs to add you directly.</p>
          </div>
          <div class="summary-card">
            <p class="summary-label">Profile photo</p>
            <p class="summary-value">{{ avatarUrl && !imgBroken ? 'Custom image' : 'Default badge' }}</p>
            <p class="summary-hint">Use a friendly photo to keep your chats recognizable.</p>
          </div>
        </div>

        <div class="grid">
          <!-- Display name editor -->
          <div class="field">
            <label class="label">Change name</label>
            <div class="hstack">
              <input v-model.trim="newName" class="input" placeholder="new name" />
              <button class="btn" :disabled="loading || !newName" @click="saveName">
                <span v-if="loading" class="spinner"></span>
                Save
              </button>
            </div>
          </div>

          <!-- Profile photo uploader -->
          <div class="field">
            <label class="label">Set photo</label>
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
import { getMyProfile, setMyUserName, setMyPhotoFile, getAvatarUrl } from '@/services/api'

const router = useRouter()
const me = ref(null)
const newName = ref('')
const file = ref(null)
const loading = ref(false)
const err = ref('')
const imgBroken = ref(false)

// Lightweight auth helper to detect an existing token.
const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

// Resolve avatar URLs through getAvatarUrl to avoid relative paths.
const avatarUrl = computed(() => getAvatarUrl(me.value || {}))

const initials = computed(() => ((me.value?.name || me.value?.username || 'U')[0] || 'U').toUpperCase())

onMounted(() => {
  if (!authed()) {
    router.replace('/login')
    return
  }
  loadProfile()
})

async function loadProfile() {
  err.value = ''
  try {
    const data = await getMyProfile()
    me.value = data?.data || data?.user || data || null
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
  padding: 18px;
  width: 100%;
}
.title {
  font-size: 1.5rem;
  font-weight: 800;
  color: #334155;
  margin: 10px 0 14px;
}

.card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 6px 18px rgba(2, 6, 23, 0.06);
}
.head {
  display: flex;
  align-items: center;
  gap: 14px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e2e8f0;
  margin-bottom: 12px;
}
.avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #e2e8f0;
}
.avatar.placeholder {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: #e0f7ee;
  border: 1px solid #a7f3d0;
  color: #0f766e;
  font-weight: 800;
}
.idblock .row {
  display: flex;
  gap: 8px;
  line-height: 1.8;
}
.key {
  color: #64748b;
  width: 100px;
}
.val {
  color: #0f172a;
  word-break: break-all;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.summary-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 10px 12px;
  box-shadow: 0 6px 18px rgba(2, 6, 23, 0.04);
}

.summary-label {
  margin: 0;
  color: #64748b;
  font-size: 0.9rem;
}

.summary-value {
  margin: 4px 0;
  font-weight: 700;
  color: #0f172a;
}

.summary-hint {
  margin: 0;
  color: #94a3b8;
  font-size: 0.9rem;
}

.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.label {
  font-weight: 600;
  color: #334155;
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
  border-radius: 10px;
  padding: 0.55rem 0.75rem;
  outline: none;
}
.input:focus {
  border-color: #22c55e;
  box-shadow: 0 0 0 0.2rem rgba(34, 197, 94, 0.15);
}

.btn {
  border: 0;
  border-radius: 10px;
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
