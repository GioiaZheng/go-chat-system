<template>
  <div class="wrap">
    <h2 class="title">Groups</h2>

    <ErrorMsg v-if="err" :text="err" class="mb-3" />

    <section class="card">
      <h3 class="h6">Create Group</h3>
      <div class="row">
        <input v-model="groupName" class="input" placeholder="Group name" />
        <input v-model="memberIdsRaw" class="input" placeholder="Member IDs (comma separated)" />
        <button class="btn" @click="createGroup" :disabled="loading">Create</button>
      </div>
      <p class="muted">Note: include yourself in the member list; otherwise the server may reject the request.</p>
    </section>

    <LoadingSpinner v-if="loading" />
    <ul v-else class="ul">
      <li v-for="g in groups" :key="g.id" class="li">
        <div class="li-main">
          <div class="name">{{ g.name || ('Group ' + (g.id || '').slice(0, 8)) }}</div>
          <div class="sub">conversation_id: {{ g.conversation_id || '(unknown)' }}</div>
        </div>
        <router-link
          v-if="g.conversation_id"
          class="link"
          :to="{ name: 'chat', params: { type: 'conv', id: g.conversation_id } }"
        >
          Open
        </router-link>
      </li>
      <li v-if="!groups.length" class="li empty">No groups yet. Create one above.</li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '../services/axios'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ErrorMsg from '../components/ErrorMsg.vue'

const router = useRouter()
const groups = ref([])
const loading = ref(false)
const err = ref('')

const groupName = ref('')
const memberIdsRaw = ref('')

function auth() {
  const t = localStorage.getItem('token')
  return t ? { Authorization: `Bearer ${t}` } : {}
}
function unwrap(res){ const d=res?.data; return (d && typeof d==='object' && 'data' in d) ? d.data : d }

onMounted(load)

async function load () {
  loading.value = true; err.value = ''; groups.value = []
  try {
    const res = await axios.get('/groups', { headers: auth() })
    const payload = unwrap(res)
    groups.value = payload?.items || payload?.groups || (Array.isArray(payload) ? payload : []) || []
  } catch (e) {
    if (e?.response?.status === 401) { err.value='Unauthorized. Please login again.'; router.push('/login') }
    else { err.value = e?.response?.data?.message || e?.message || 'Failed to load groups' }
  } finally {
    loading.value = false
  }
}

async function createGroup () {
  err.value = ''
  const members = memberIdsRaw.value.split(',').map(s => s.trim()).filter(Boolean)
  if (!groupName.value || members.length === 0) { err.value = 'Group name and member list are required.'; return }
  loading.value = true
  try {
    await axios.post('/groups', { name: groupName.value, members, member_ids: members }, { headers: auth() })
    await load(); groupName.value=''; memberIdsRaw.value=''
  } catch (e) {
    if (e?.response?.status === 401) { err.value='Unauthorized. Please login again.'; router.push('/login') }
    else { err.value = e?.response?.data?.message || e?.message || 'Failed to create group' }
  } finally { loading.value = false }
}
</script>

<style scoped>
.wrap{ max-width:900px; margin:0 auto; padding:18px; }
.title{ font-size:1.5rem; font-weight:800; color:#334155; margin:6px 0 14px; }

.card{
  background:#fff; border:1px solid #e2e8f0; border-radius:14px; padding:14px;
  box-shadow:0 6px 18px rgba(2,6,23,.06); margin-bottom:12px;
}
.h6{ margin:0 0 8px; font-weight:700; color:#0f172a; }
.row{ display:flex; gap:10px; flex-wrap:wrap; }
.input{
  flex:1 1 240px; border:1px solid #cbd5e1; border-radius:10px; padding:.55rem .75rem; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }
.btn{
  border:0; border-radius:10px; color:#fff; padding:.55rem .9rem;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65 }
.muted{ margin:.35rem 0 0; color:#64748b; font-size:.85rem; }

.ul{ list-style:none; padding:0; margin:0; }
.li{
  background:#fff; border:1px solid #e2e8f0; border-radius:12px; padding:12px;
  display:flex; align-items:center; justify-content:space-between; margin-bottom:10px;
}
.li.empty{ text-align:center; color:#64748b }
.name{ font-weight:600; color:#0f172a }
.sub{ color:#64748b; font-size:.9rem }
.link{ color:#2563eb; text-decoration:none }
.link:hover{ text-decoration:underline }
</style>
