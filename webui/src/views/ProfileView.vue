<template>
  <div class="page">
    <header class="topbar">
      <div class="brand">WASA <span class="grad">Chat</span></div>
      <router-link class="nav" to="/conversations">Conversations</router-link>
    </header>

    <main class="wrap">
      <h2 class="title">My Profile</h2>

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <section v-if="me" class="card">
        <div class="head">
          <img
            v-if="me.photo_url || me.avatarUri"
            :src="me.photo_url || me.avatarUri"
            alt="avatar"
            class="avatar"
          />
          <div v-else class="avatar placeholder">{{ initials }}</div>

          <div class="idblock">
            <div class="row"><span class="key">id:</span><span class="val">{{ me.id }}</span></div>
            <div class="row"><span class="key">username:</span><span class="val">{{ me.username }}</span></div>
            <div class="row"><span class="key">name:</span><span class="val">{{ me.name }}</span></div>
            <div class="row"><span class="key">email:</span><span class="val">{{ me.email }}</span></div>
            <div class="row"><span class="key">gender:</span><span class="val">{{ me.gender }}</span></div>
          </div>
        </div>

        <div class="grid">
          <!-- Change username -->
          <div class="field">
            <label class="label">Change username</label>
            <div class="hstack">
              <input v-model.trim="newName" class="input" placeholder="new username" />
              <button class="btn" :disabled="loading || !newName" @click="setUsername">
                <span v-if="loading && savingKind==='name'" class="spinner"></span>
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
                <span v-if="loading && savingKind==='photo'" class="spinner"></span>
                Upload
              </button>
              <!-- 可选：使用预设头像（示例） -->
              <button class="btn" :disabled="loading" @click="setPreset('avatar7')">
                <span v-if="loading && savingKind==='photo-preset'" class="spinner"></span>
                Use preset avatar7
              </button>
            </div>
            <p class="hint">You can also use preset mode via backend (e.g., <code>?preset=avatar7</code>), if enabled.</p>
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
import api, { getMyProfile, setMyUserName, setMyPhoto } from '../services/api' // ✅ 用你自己的 api.js

const router = useRouter()

const me = ref(null)
const newName = ref('')
const file = ref(null)
const loading = ref(false)
const savingKind = ref('') // 'name' | 'photo' | 'photo-preset'
const err = ref('')

const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

const initials = computed(() => {
  const s = (me.value?.username || me.value?.name || 'U').toString().trim()
  const parts = s.split(/\s+/)
  return (parts[0]?.[0] || 'U').toUpperCase() + (parts[1]?.[0] || '')
})

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
    // 后端可能返回 {user: {...}} 或直接 {...}
    me.value = data?.user ?? data ?? null
    if (me.value) localStorage.setItem('me', JSON.stringify(me.value))
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to load profile'
    }
  }
}

function onFile(e) { file.value = e.target.files?.[0] || null }

async function setUsername() {
  if (!newName.value) return
  loading.value = true
  savingKind.value = 'name'
  err.value = ''
  try {
    const next = newName.value
    await setMyUserName(newName.value)
    newName.value = ''
    await loadProfile()
    if (me.value && me.value.username !== next) me.value.username = next
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'; router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to set username'
    }
  } finally {
    loading.value = false
    savingKind.value = ''
  }
}

async function setPhoto() {
  if (!file.value) return
  loading.value = true
  savingKind.value = 'photo'
  err.value = ''
  try {
    await setMyPhoto({ file: file.value }) // api.js 内部自动用字段名 upload
    file.value = null
    await loadProfile()
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'; router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to set photo'
    }
  } finally {
    loading.value = false
    savingKind.value = ''
  }
}

// 可选：使用后端预设头像
async function setPreset(presetName) {
  loading.value = true
  savingKind.value = 'photo-preset'
  err.value = ''
  try {
    await setMyPhoto({ preset: presetName })
    await loadProfile()
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'; router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to set photo'
    }
  } finally {
    loading.value = false
    savingKind.value = ''
  }
}
</script>

<style scoped>
/* Page scaffold (white / green / blue) */
.page{
  min-height:100vh;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
  color:#0f172a;
}
.topbar{
  height:56px; display:flex; align-items:center; justify-content:space-between;
  padding:0 18px; border-bottom:1px solid rgba(20,100,60,.08); background:#fff8; backdrop-filter: blur(6px);
}
.brand{ font-weight:800; letter-spacing:.4px; }
.grad{ background: linear-gradient(90deg,#22c55e,#3b82f6); -webkit-background-clip:text; background-clip:text; color:transparent; }
.nav{ color:#2563eb; text-decoration:none; }
.nav:hover{ text-decoration:underline; }

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

.grid{ display:grid; grid-template-columns: 1fr; gap:14px; }
@media (min-width: 720px){ .grid{ grid-template-columns: 1fr 1fr; } }

.field{ display:flex; flex-direction:column; gap:8px; }
.label{ font-weight:600; color:#334155; }
.hstack{ display:flex; gap:10px; align-items:center; flex-wrap:wrap; }

.input{
  flex:1 1 220px; min-width: 0;
  border:1px solid #cbd5e1; border-radius:10px; padding:.55rem .75rem; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }

.btn{
  border:0; border-radius:10px; color:#fff; padding:.55rem .9rem; white-space:nowrap;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65; cursor:not-allowed; }

.hint{ margin:.25rem 0 0; color:#64748b; font-size:.85rem; }

.spinner{
  display:inline-block; width:1em; height:1em; margin-right:.4em;
  border:2px solid rgba(255,255,255,.6); border-top-color:transparent; border-radius:50%;
  animation: spin .7s linear infinite; vertical-align:-2px;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
