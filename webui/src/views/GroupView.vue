<!-- src/views/GroupView.vue: Group directory with creation and management panels. -->
<template>
  <div class="page">
    <header class="topbar">
      <div class="title">Groups</div>
    </header>

    <section class="content">
      <ErrorMsg v-if="err" :text="err" class="mb-2" />
      <p v-else-if="notice" class="notice">{{ notice }}</p>

      <!-- Creation form -->
      <section class="card">
        <h3 class="h6">Create Group</h3>
        <div class="field">
          <input v-model.trim="groupName" class="input" placeholder="Group name" />
        </div>
        <div class="field">
          <input
            v-model.trim="memberIdsRaw"
            class="input"
            placeholder="Member IDs (comma separated)"
          />
          <div class="muted tip">
            You can enter any user IDs (not limited to your contacts).<br />
            Your own ID is appended automatically if it is missing.
          </div>
        </div>
        <div class="actions">
          <button
            class="btn"
            :disabled="creating || !groupName"
            @click="createGroup"
          >
            {{ creating ? 'Creating…' : 'Create' }}
          </button>
        </div>
      </section>

      <!-- List -->
      <LoadingSpinner v-if="loadingList" />
      <ul v-else class="list">
        <li v-for="g in groups" :key="g.id" class="item">
          <div class="left">
            <span v-if="!g.avatar" class="avatar-fallback">{{ initials(g) }}</span>
            <img
              v-else
              class="avatar"
              :src="g.avatar"
              alt="group avatar"
            />
            <div class="info">
              <div class="name">{{ g.name || ('Group ' + (g.id || '').slice(0,8)) }}</div>
              <div class="sub">conversation_id: {{ g.conversation_id || '(unknown)' }}</div>
            </div>
          </div>

          <div class="right">
            <RouterLink
              v-if="g.conversation_id"
              class="open"
              :to="{ name: 'chat', params: { type: 'conv', id: g.conversation_id } }"
            >
              Open
            </RouterLink>
            <button class="link" @click="toggleManage(g.id)">
              {{ manageId === g.id ? 'Close' : 'Manage' }}
            </button>
          </div>

          <!-- Inline management panel -->
          <div v-if="manageId === g.id" class="manage">
            <div class="row">
              <label class="lbl">Rename</label>
              <input v-model.trim="editName" class="input" placeholder="New group name" />
              <button class="btn sm" :disabled="busy" @click="onRename(g.id)">Save</button>
            </div>

            <div class="row">
              <label class="lbl">Avatar</label>
              <input ref="filePick" type="file" accept="image/*" @change="onPickPhoto(g.id, $event)" />
              <button class="btn sm" :disabled="busy" @click="triggerPick">Upload</button>
            </div>

            <div class="row">
              <label class="lbl">Add members</label>
              <input
                v-model.trim="addIds"
                class="input"
                placeholder="User IDs (comma separated)"
              />
              <button class="btn sm" :disabled="busy || !addIds" @click="onAddMembers(g.id)">Add</button>
            </div>

            <div class="row leave">
              <button class="btn danger" :disabled="busy" @click="onLeave(g.id)">Leave Group</button>
            </div>

            <div v-if="panelErr" class="panel-err">{{ panelErr }}</div>
          </div>
        </li>

        <li v-if="!groups.length" class="empty">No groups yet. Create one above.</li>
      </ul>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import {
  getGroupsList,
  createGroup as apiCreateGroup,
  getGroupDetail,
  setGroupName,
  setGroupPhoto,
  addToGroup,
  leaveGroup,
  getMyProfile,
  getAvatarUrl,
} from '@/services/api'

const router = useRouter()

// Authentication helpers and cached profile details.
const authed = () => !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))
const meId = ref('')
async function loadMe () {
  try { const me = await getMyProfile(); meId.value = String(me?.id || '') } catch {}
}

// Page-level error, notice, and loading indicators.
const err = ref('')
const notice = ref('')
const creating = ref(false)
const loadingList = ref(false)
const groups = ref([])

// Create form inputs.
const groupName = ref('')
const memberIdsRaw = ref('')

// Manage panel state scoped per group.
const manageId = ref('')
const editName = ref('')
const addIds = ref('')
const busy = ref(false)
const panelErr = ref('')
const filePick = ref(null)

// Helper utilities for normalizing API payloads.
function normalizeGroups(list) {
  const arr = Array.isArray(list) ? list : (list?.items ?? list?.groups ?? [])
  return (arr || [])
    .map(g => {
      const avatar = g.avatarUri || g.avatar_uri || g.avatar_url || g.avatar || ''
      return {
        id: g.id ?? g.group_id ?? g._id,
        name: g.name ?? g.title ?? '',
        conversation_id: g.conversation_id ?? g.conversationId ?? g.cid ?? null,
        avatar: avatar ? getAvatarUrl({ avatarUri: avatar, updatedAt: g.updatedAt || g.updated_at }) : '',
      }
    })
    .filter(g => !!g.id)
}

function initials(g) {
  const src = g?.name || g?.title || 'G'
  const match = String(src).match(/\b\w/g) || ['G']
  return match.slice(0, 2).join('').toUpperCase()
}

async function loadList() {
  loadingList.value = true
  err.value = ''
  notice.value = ''
  try {
    const data = await getGroupsList()
    groups.value = normalizeGroups(data)
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to load groups'
    }
  } finally {
    loadingList.value = false
  }
}

// Group creation workflow.
async function createGroup() {
  err.value = ''
  notice.value = ''
  const list = memberIdsRaw.value
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)

  // Always include the current user when creating the group.
  if (meId.value && !list.includes(meId.value)) list.unshift(meId.value)

  if (!groupName.value) {
    err.value = 'Group name is required.'
    return
  }

  creating.value = true
  try {
    await apiCreateGroup({ name: groupName.value, memberIds: list })
    groupName.value = ''
    memberIdsRaw.value = ''
    notice.value = 'Group created successfully.'
    await loadList()
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to create group'
    }
  } finally {
    creating.value = false
  }
}

// Inline management actions for each group card.
function toggleManage(id) {
  panelErr.value = ''
  if (manageId.value === id) {
    manageId.value = ''
    editName.value = ''
    addIds.value = ''
    return
  }
  manageId.value = id
  // Refresh the panel with the latest group details before editing.
  hydrateManage(id)
}

async function hydrateManage(id) {
  try {
    const g = await getGroupDetail(id)
    editName.value = g?.name || ''
  } catch {}
}

async function onRename(id) {
  panelErr.value = ''
  if (!editName.value) { panelErr.value = 'Name is required.'; return }
  busy.value = true
  try {
    await setGroupName(id, editName.value)
    await loadList()
  } catch (e) {
    panelErr.value = e?.response?.data?.message || e?.message || 'Failed to rename group'
  } finally {
    busy.value = false
  }
}

function triggerPick () {
  if (filePick.value) filePick.value.click()
}
async function onPickPhoto(id, e) {
  const file = e?.target?.files?.[0]
  e.target.value = ''
  if (!file) return
  panelErr.value = ''
  busy.value = true
  try {
    await setGroupPhoto(id, file)
    await loadList()
  } catch (er) {
    panelErr.value = er?.response?.data?.message || er?.message || 'Failed to upload photo'
  } finally {
    busy.value = false
  }
}

async function onAddMembers(id) {
  panelErr.value = ''
  const list = addIds.value.split(',').map(s => s.trim()).filter(Boolean)
  if (!list.length) { panelErr.value = 'Please enter user IDs.'; return }
  busy.value = true
  try {
    await addToGroup(id, list)
    addIds.value = ''
    // Reloading the list here keeps the on-screen data in sync.
    await loadList()
  } catch (e) {
    panelErr.value = e?.response?.data?.message || e?.message || 'Failed to add members'
  } finally {
    busy.value = false
  }
}

async function onLeave(id) {
  panelErr.value = ''
  if (!confirm('Are you sure you want to leave this group?')) return
  busy.value = true
  try {
    await leaveGroup(id)
    if (manageId.value === id) manageId.value = ''
    await loadList()
  } catch (e) {
    panelErr.value = e?.response?.data?.message || e?.message || 'Failed to leave group'
  } finally {
    busy.value = false
  }
}

// Bootstrap the page once authenticated.
onMounted(async () => {
  if (!authed()) { router.replace('/login'); return }
  await loadMe()
  await loadList()
})
</script>

<style scoped>
.page{
  min-height:100vh;
  display:flex;
  flex-direction:column;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg,#ffffff,#f7fafe);
  color:#0f172a;
}
.topbar{
  height:56px; display:flex; align-items:center; justify-content:space-between;
  padding:0 18px; border-bottom:1px solid rgba(20,100,60,.08); background:#fff8; backdrop-filter: blur(6px);
}
.title{ font-weight:800; color:#0f172a; }
.content{ max-width:1100px; margin:0 auto; padding:16px; }

.card{
  background:#fff; border:1px solid #e2e8f0; border-radius:14px; padding:14px; margin-bottom:12px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
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
.btn.sm{ padding:.45rem .7rem; }
.btn.danger{
  background-image:linear-gradient(135deg,#ef4444,#f97316);
}

.muted{ color:#64748b; font-size:.9rem }
.tip{ margin-top:6px }
.notice{ background:#ecfdf3; color:#166534; border:1px solid #bbf7d0; padding:8px 10px; border-radius:10px; }

.list{ list-style:none; padding:0; margin:10px 0 0; display:grid; gap:10px }
.item{
  background:#fff; border:1px solid #e2e8f0; border-radius:12px; padding:12px;
  display:flex; flex-direction:column; gap:10px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
}
.left{ display:flex; align-items:center; gap:10px }
.right{ display:flex; gap:12px; align-items:center; margin-left:auto }
.link{ background:none; border:0; color:#2563eb; cursor:pointer }
.open{ color:#2563eb; text-decoration:none }
.open:hover{ text-decoration:underline }
.info{ min-width:0 }
.name{ font-weight:600; color:#0f172a }
.sub{ color:#64748b; font-size:.9rem }
.empty{ text-align:center; color:#64748b }

.avatar{
  width:36px; height:36px; border-radius:50%; object-fit:cover;
  border:1px solid #e2e8f0; background:#fff;
}

.avatar-fallback{
  width:36px; height:36px; border-radius:50%; display:inline-flex;
  align-items:center; justify-content:center;
  background:#e0f7ee; color:#0f766e; font-weight:700; border:1px solid #a7f3d0;
}

.manage{
  border-top:1px dashed #e2e8f0; padding-top:10px; display:grid; gap:10px;
}
.row{ display:grid; grid-template-columns: 90px 1fr auto; gap:8px; align-items:center }
.lbl{ color:#475569; font-size:.92rem }
.leave{ margin-top:4px; grid-template-columns: 1fr auto; }
.panel-err{ color:#dc2626; font-size:.9rem }
</style>
