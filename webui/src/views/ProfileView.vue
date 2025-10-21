<!-- src/views/ProfileView.vue -->
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
            <div class="row"><span class="key">id:</span><span class="val">{{ me.id }}</span></div>
            <div class="row"><span class="key">username:</span><span class="val">{{ me.username }}</span></div>
          </div>
        </div>

        <div class="grid">
          <!-- Change username -->
          <div class="field">
            <label class="label">Change username</label>
            <div class="hstack">
              <input v-model.trim="newUsername" class="input" placeholder="new username" />
              <button class="btn" :disabled="loading || !newUsername" @click="setUsername">
                <span v-if="loading" class="spinner"></span>
                Save
              </button>
            </div>
          </div>

          <!-- Set photo -->
          <div class="field">
            <label class="label">Set photo</label>
            <div class="hstack">
              <input type="file" @change="onFile" />
              <button class="btn" :disabled="loading || !file" @click="setPhoto">
                <span v-if="loading" class="spinner"></span>
                Upload
              </button>
              <button class="btn" :disabled="loading" @click="setPreset('avatar7')">
                Use preset avatar7
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
import ErrorMsg from '../components/ErrorMsg.vue'
import { getMyProfile, setMyUserName, setMyPhoto } from '../services/api'

const router = useRouter()
const me = ref(null)
const newUsername = ref('')
const file = ref(null)
const loading = ref(false)
const err = ref('')
const imgBroken = ref(false)

const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

const avatarUrl = computed(() => {
  const raw = me.value?.photo_url || ''
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw)) return raw
  const base = typeof __API_URL__ !== 'undefined' ? __API_URL__ : ''
  return raw.startsWith('/') ? `${base}${raw}` : `${base}/${raw}`
})

const initials = computed(() => (me.value?.username?.[0]?.toUpperCase() || 'U'))

onMounted(() => {
  if (!authed()) { router.replace('/login'); return }
  loadProfile()
})

async function loadProfile () {
  err.value = ''
  try {
    const data = await getMyProfile()
    me.value = data?.user ?? data ?? null
    if (me.value) {
      localStorage.setItem('username', me.value.username)
      localStorage.setItem('me', JSON.stringify(me.value))
      window.dispatchEvent(new Event('auth:changed')) // 顶栏立即更新
    }
    imgBroken.value = false
  } catch (e) { handleError(e, 'Failed to load profile') }
}

async function setUsername () {
  if (!newUsername.value) return
  loading.value = true; err.value = ''
  try {
    await setMyUserName(newUsername.value)
    await loadProfile()
    newUsername.value = ''
  } catch (e) { handleError(e, 'Failed to set username') }
  finally { loading.value = false }
}

function onFile (e) { file.value = e.target.files?.[0] || null }

async function setPhoto () {
  if (!file.value) return
  loading.value = true; err.value = ''
  try {
    await setMyPhoto({ file: file.value })
    file.value = null
    await loadProfile()
  } catch (e) { handleError(e, 'Failed to upload photo') }
  finally { loading.value = false }
}

async function setPreset (preset) {
  loading.value = true; err.value = ''
  try {
    await setMyPhoto({ preset })
    await loadProfile()
  } catch (e) { handleError(e, 'Failed to set preset photo') }
  finally { loading.value = false }
}

function handleError (e, fallback) {
  if (e?.response?.status === 401) { err.value = 'Unauthorized. Please login again.'; router.push('/login') }
  else { err.value = e?.response?.data?.message || e?.message || fallback }
}
</script>

<style scoped>
.page{
  min-height:100vh;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
}
/* 顶栏已交由 App.vue 统一渲染，这里不需要再放 header */

.wrap{ max-width:900px; margin:0 auto; padding:18px; }
.title{ font-size:1.5rem; font-weight:800; color:#334155; margin:10px 0 14px; }

.card{
  background:#fff; border:1px solid #e2e8f0; border-radius:16px; padding:16px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
}

.head{ display:flex; align-items:center; gap:14px; padding-bottom:12px; border-bottom:1px solid #e2e8f0; margin-bottom:12px; }
.avatar{ width:64px; height:64px; border-radius:50%; object-fit:cover; border:1px solid #e2e8f0; }
.avatar.placeholder{
  width:64px; height:64px; border-radius:50%; display:grid; place-items:center;
  background:#e0f7ee; border:1px solid #a7f3d0; color:#0f766e; font-weight:800;
}
.idblock .row{ display:flex; gap:8px; line-height:1.8; }
.key{ color:#64748b; width:100px; }
.val{ color:#0f172a; word-break:break-all; }

.grid{ display:grid; grid-template-columns: 1fr 1fr; gap:14px; }
.field{ display:flex; flex-direction:column; gap:8px; }
.label{ font-weight:600; color:#334155; }
.hstack{ display:flex; gap:10px; align-items:center; flex-wrap:wrap; }

.input{
  flex:1 1 220px;
  border:1px solid #cbd5e1; border-radius:10px; padding:.55rem .75rem; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }

.btn{
  border:0; border-radius:10px; color:#fff; padding:.55rem .9rem; white-space:nowrap;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65; cursor:not-allowed; }

.spinner{
  display:inline-block; width:1em; height:1em; margin-right:.4em;
  border:2px solid rgba(255,255,255,.6); border-top-color:transparent; border-radius:50%;
  animation: spin .7s linear infinite; vertical-align:-2px;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
