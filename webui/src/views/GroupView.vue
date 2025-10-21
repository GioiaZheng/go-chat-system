<template>
  <div class="layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="brand">WASAText</div>

      <nav class="nav">
        <div class="sec">CONVERSATIONS</div>
        <RouterLink class="link" to="/conversations">All Conversations</RouterLink>
        <RouterLink class="link active" to="/groups">Groups</RouterLink>

        <div class="sec mt">SETTINGS</div>
        <RouterLink class="link" to="/profile">Profile</RouterLink>
        <button class="link link-btn" @click="logout">Logout</button>
      </nav>
    </aside>

    <!-- Main -->
    <main class="main">
      <header class="topbar">
        <div class="title">Groups</div>
        <div class="who">Signed in as {{ meName }}</div>
      </header>

      <section class="content">
        <ErrorMsg v-if="err" :text="err" class="mb-2" />

        <section class="card">
          <h3 class="h6">Create Group</h3>
          <div class="field">
            <input v-model.trim="groupName" class="input" placeholder="Group name" />
          </div>
          <div class="field">
            <input v-model.trim="memberIdsRaw" class="input" placeholder="Member IDs (comma separated)" />
          </div>
          <div class="actions">
            <button class="btn" :disabled="loading || !groupName || !memberIdsRaw" @click="createGroup">
              {{ loading ? 'Creating…' : 'Create' }}
            </button>
            <span class="muted">Note: include yourself in the list (或我们会自动补上自己)。</span>
          </div>
        </section>

        <LoadingSpinner v-if="loadingList" />
        <ul v-else class="list">
          <li v-for="g in groups" :key="g.id" class="item">
            <div class="info">
              <div class="name">{{ g.name || ('Group ' + (g.id || '').slice(0,8)) }}</div>
              <div class="sub">conversation_id: {{ g.conversation_id || '(unknown)' }}</div>
            </div>
            <RouterLink
              v-if="g.conversation_id"
              class="open"
              :to="{ name:'chat', params:{ type:'conv', id:g.conversation_id } }"
            >Open</RouterLink>
          </li>
          <li v-if="!groups.length" class="empty">No groups yet. Create one above.</li>
        </ul>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { getGroupsList, createGroup as apiCreateGroup, getMyProfile } from '@/services/api'

const router = useRouter()
const meName = ref(localStorage.getItem('name') || 'user')
const meId = ref('')

const groups = ref([])
const loading = ref(false)
const loadingList = ref(false)
const err = ref('')

const groupName = ref('')
const memberIdsRaw = ref('')

const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

function logout(){
  localStorage.clear(); sessionStorage.clear()
  router.replace('/login')
}

function normalizeGroups(list){
  const arr = Array.isArray(list) ? list : (list?.items ?? list?.groups ?? [])
  return (arr || []).map(g => ({
    id: g.id ?? g.group_id ?? g._id,
    name: g.name ?? g.title ?? '',
    conversation_id: g.conversation_id ?? g.conversationId ?? g.cid ?? null,
  })).filter(g => !!g.id)
}

async function loadMe(){
  try {
    const data = await getMyProfile()
    const u = data?.user ?? data
    meId.value = String(u?.id || '')
  } catch {}
}

async function loadList(){
  loadingList.value = true; err.value = ''; groups.value = []
  try {
    const data = await getGroupsList()
    groups.value = normalizeGroups(data)
  } catch(e){
    if (e?.response?.status === 401) { err.value='Unauthorized. Please login again.'; router.push('/login') }
    else err.value = e?.response?.data?.message || e?.message || 'Failed to load groups'
  } finally {
    loadingList.value = false
  }
}

onMounted(async () => {
  if (!authed()) { router.replace('/login'); return }
  await loadMe()
  await loadList()
})

async function createGroup(){
  err.value = ''
  const list = memberIdsRaw.value
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)

  // 自动把自己也加入（如果没写）
  if (meId.value && !list.includes(meId.value)) list.unshift(meId.value)

  if (!groupName.value || list.length === 0) {
    err.value = 'Group name and member list are required.'
    return
  }

  loading.value = true
  try {
    await apiCreateGroup({ name: groupName.value, members: list })
    groupName.value = ''
    memberIdsRaw.value = ''
    await loadList()
  } catch(e){
    if (e?.response?.status === 401) { err.value='Unauthorized. Please login again.'; router.push('/login') }
    else err.value = e?.response?.data?.message || e?.message || 'Failed to create group'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* Layout 同 ChatView / ConversationsView */
.layout{
  min-height:100vh;
  display:grid;
  grid-template-columns:240px 1fr;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
  color:#0f172a;
}
.sidebar{ background:#f8fafc; border-right:1px solid #e2e8f0; padding:14px 12px; }
.brand{ height:44px; display:flex; align-items:center; padding:0 8px; font-weight:800; }
.nav{ padding:8px 4px; display:flex; flex-direction:column; gap:6px; }
.sec{ font-size:.78rem; color:#64748b; padding:6px 8px; }
.mt{ margin-top:8px; }
.link{ display:block; padding:8px 10px; border-radius:8px; color:#0f172a; text-decoration:none; }
.link:hover{ background:#eef2f7; }
.link-btn{ background:none; border:0; text-align:left; }
.link.active{ background:#eef2f7; }

.main{ display:flex; flex-direction:column; }
.topbar{
  height:56px; display:flex; align-items:center; justify-content:space-between;
  padding:0 18px; border-bottom:1px solid rgba(20,100,60,.08); background:#fff8; backdrop-filter:blur(6px);
}
.title{ font-size:1.1rem; font-weight:800; color:#0f172a; }
.who{ color:#64748b; font-size:.95rem }
.content{ max-width:960px; margin:0 auto; padding:18px; }

.card{
  background:#fff; border:1px solid #e2e8f0; border-radius:14px; padding:14px;
  box-shadow:0 6px 18px rgba(2,6,23,.06); margin-bottom:12px;
}
.h6{ margin:0 0 8px; font-weight:700; color:#0f172a; }
.field{ margin-bottom:8px }
.input{ width:100%; border:1px solid #cbd5e1; border-radius:10px; padding:.55rem .75rem; outline:none; }
.actions{ display:flex; align-items:center; gap:10px; }
.btn{
  border:0; border-radius:10px; color:#fff; padding:.55rem .9rem;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.muted{ color:#64748b; font-size:.9rem }

.list{ list-style:none; padding:0; margin:10px 0 0; }
.item{
  background:#fff; border:1px solid #e2e8f0; border-radius:12px; padding:12px;
  display:flex; align-items:center; justify-content:space-between; margin-bottom:10px;
}
.info{ min-width:0 }
.name{ font-weight:600; color:#0f172a }
.sub{ color:#64748b; font-size:.9rem }
.open{ color:#2563eb; text-decoration:none }
.open:hover{ text-decoration:underline }
.empty{ text-align:center; color:#64748b }
</style>
