<!-- src/views/NewGroupView.vue: Guided flow for creating a new group chat. -->
<template>
  <div class="page">
    <section class="content">
      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <div class="panel new-group-panel">
        <div class="form">
          <!-- Group name input -->
          <label class="label">Group Name</label>
          <input v-model.trim="groupName" placeholder="Enter group name" class="input" />

          <!-- Optional avatar upload -->
          <label class="label mt">Group Avatar (optional)</label>
          <div class="avatar-row">
            <AvatarUpload
              :src="avatarPreview"
              fallback-text="+"
              overlay-text="Upload"
              alt="Group avatar preview"
              :size="48"
              @select="onPickFile"
            />
            <button
              v-if="avatarFile"
              class="btn-outline"
              type="button"
              @click="clearAvatar"
            >
              Clear
            </button>
          </div>

          <!-- Member search and selection -->
          <label class="label mt">Add Members</label>
          <UserSearch
            placeholder="Search users by name/username"
            class="mb-2"
            @select="onPickUser"
            @error="onChildError"
          />

          <div class="selected-box">
            <div class="selected-header">
              <span>Selected members</span>
              <span class="selected-count">{{ pickedUsers.length }}</span>
            </div>
            <div class="picked" v-if="pickedUsers.length">
              <div
                v-for="u in pickedUsers"
                :key="String(u.id)"
                class="chip"
                :title="u.username || u.name || ''"
              >
                <div v-if="!avatar(u)" class="chip-avatar fallback">{{ initials(u) }}</div>
                <img v-else :src="avatar(u)" class="chip-avatar" alt="avatar" />
                <span class="chip-name">{{ u.name || u.username || '(user)' }}</span>
                <button class="chip-x" @click="removePicked(u)">×</button>
              </div>
            </div>
            <p v-else class="selected-empty">No members selected yet. Use the search above to add people.</p>
          </div>
          <div class="form-actions">
            <button
              class="btn btn-primary"
              :disabled="!groupName.trim() || pickedUsers.length < 2 || loading"
              @click="create"
            >
              {{ loading ? 'Creating…' : 'Create Group' }}
            </button>
            <p v-if="showHelper" class="helper">
              A group must include at least 3 people.
            </p>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import AvatarUpload from '@/components/AvatarUpload.vue'
import UserSearch from '@/components/UserSearch.vue'
import { getAvatarUrl, createGroup as apiCreateGroup, setGroupPhoto } from '@/services/api'
import { ensureAuthReady, isAuthenticated, currentUser, refreshProfile } from '@/services/auth'

const router = useRouter()

const err = ref('')
const loading = ref(false)
const me = ref(null)

const groupName = ref('')

// Selected users displayed as chips with avatar/name metadata.
const pickedUsers = ref([])
const avatarFile = ref(null)
const avatarPreview = ref('')
const attemptedCreate = ref(false)

const avatar = (u) => getAvatarUrl(u)
const initials = (u) => {
  const name = u?.name || u?.username || 'U'
  const match = String(name).match(/\b\w/g) || ['U']
  return match.slice(0, 2).join('').toUpperCase()
}
const memberIds = computed(() => {
  const ids = new Set(pickedUsers.value.map(u => String(u.id)).filter(Boolean))
  if (me.value?.id) ids.add(String(me.value.id))
  return Array.from(ids)
})

onMounted(async () => {
  await ensureAuthReady()
  if (!isAuthenticated.value) {
    router.replace('/login')
    return
  }
  await refreshProfile().catch(() => {})
  me.value = currentUser.value
})

function onChildError(msg) {
  err.value = msg || ''
}

function onPickUser(u) {
  err.value = ''
  const id = String(u?.id || '')
  if (!id) return
  if (!pickedUsers.value.some(x => String(x.id) === id)) {
    pickedUsers.value.push(u)
  }
}

function removePicked(u) {
  const id = String(u?.id || '')
  pickedUsers.value = pickedUsers.value.filter(x => String(x.id) !== id)
}

function onPickFile(f) {
  if (avatarPreview.value) {
    URL.revokeObjectURL(avatarPreview.value)
  }
  avatarFile.value = f || null
  if (f) {
    const url = URL.createObjectURL(f)
    avatarPreview.value = url
  } else {
    avatarPreview.value = ''
  }
}

function clearAvatar() {
  if (avatarPreview.value) {
    URL.revokeObjectURL(avatarPreview.value)
  }
  avatarFile.value = null
  avatarPreview.value = ''
}

const showHelper = computed(() => {
  if (!attemptedCreate.value) return false
  return !groupName.value.trim() || pickedUsers.value.length < 2
})

async function create() {
  err.value = ''
  attemptedCreate.value = true
  if (!groupName.value.trim() || pickedUsers.value.length < 2) return

  loading.value = true
  try {
    // 1) Create the group (backend accepts any users as long as you have their IDs)
    const res = await apiCreateGroup({ name: groupName.value.trim(), memberIds: memberIds.value })
    // Handle different response shapes
    const groupId =
      res?.group?.id || res?.id || res?.group_id || res?._id || res?.gid || null
    const conversationId =
      res?.conversation_id ||
      res?.conversationId ||
      res?.conversation?.id ||
      res?.group?.conversation_id ||
      res?.group?.conversationId ||
      null

    if (!groupId) throw new Error('No group id returned')
    // 2) Upload the group avatar immediately when provided
    if (avatarFile.value) {
      await setGroupPhoto(groupId, avatarFile.value)
    }
    window.dispatchEvent(new CustomEvent('conversations:reload'))

    // 3) Navigate to the group chat using conversationId
    if (conversationId) {
      router.push({ name: 'chat', params: { type: 'group', id: String(conversationId) } })
    } else {
      // Fallback: return to the group list if no conversation id is present
      router.push('/groups')
    }
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to create group'
  } finally {
    loading.value = false
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
  background: #e5e7eb;
}
.content {
  flex: 1 1 auto;
  padding: 16px;
  width: 100%;
  display: flex;
  justify-content: center;
}

.panel.new-group-panel {
  width: min(100%, 760px);
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: var(--shadow);
  padding: 20px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.label { font-weight: 600; color: var(--muted); }
.mt { margin-top: 8px; }

.input {
  border: 1px solid var(--border);
  border-radius: var(--radius-control);
  padding: 10px 12px;
  outline: none;
  background: #fff;
  font-size: var(--font-primary);
}
.input:focus { border-color: #22c55e; box-shadow: 0 0 0 3px rgba(34,197,94,.15); }

.avatar-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.btn {
  border: 0;
  border-radius: var(--radius-control);
  color: #fff;
  padding: 12px 16px;
  background: var(--grad);
  box-shadow: 0 10px 24px rgba(34, 197, 94, 0.2);
}
.btn:disabled { opacity: 0.65; }
.btn-primary {
  align-self: center;
  width: min(100%, 240px);
}
.btn-outline {
  border: 1px solid var(--border);
  background: #f1f5f9;
  color: #0f172a;
  border-radius: var(--radius-control);
  padding: 8px 12px;
}

.muted { color: var(--muted); font-size: .9rem; }
.form-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
}
.helper {
  color: var(--muted);
  font-size: 0.9rem;
  text-align: center;
}

.selected-box {
  border: 1px dashed #e2e8f0;
  border-radius: 12px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: #fff;
}

.selected-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  color: #0f172a;
}

.selected-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 22px;
  border-radius: 999px;
  border: 1px solid #e2e8f0;
  font-size: 0.8rem;
  color: #475569;
}

.selected-empty {
  margin: 0;
  color: #64748b;
  font-size: 0.9rem;
}

.picked { display: flex; gap: 8px; flex-wrap: wrap; }
.chip {
  display: inline-flex; align-items: center; gap: 6px; padding: 4px 8px;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 999px;
}
.chip-avatar { width: 22px; height: 22px; border-radius: 50%; object-fit: cover; border: 1px solid #cbd5e1; }
.chip-avatar.fallback {
  display: inline-flex; align-items: center; justify-content: center;
  background: #e0f7ee; color: #0f766e; font-weight: 700; border: 1px solid #a7f3d0;
}
.chip-name { color: #0f172a; font-weight: 600; }
.chip-x {
  border: 0; background: transparent; cursor: pointer; font-size: 1rem; line-height: 1;
  color: #64748b;
}
</style>
