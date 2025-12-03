<!-- src/views/NewGroupView.vue -->
<template>
  <div class="page">
    <header class="topbar">
      <div class="title">Create Group</div>
    </header>

    <section class="content">
      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <div class="form">
        <!-- Group name -->
        <label class="label">Group Name</label>
        <input v-model.trim="groupName" placeholder="Enter group name" class="input" />

        <!-- Avatar upload (optional) -->
        <label class="label mt">Group Avatar (optional)</label>
        <div class="avatar-row">
          <div class="avatar-preview" v-if="avatarPreview">
            <img :src="avatarPreview" alt="preview" />
          </div>
          <input
            ref="fileInput"
            type="file"
            accept="image/*"
            @change="onPickFile"
            class="file"
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

        <!-- Member search + picked list -->
        <label class="label mt">Add Members</label>
        <UserSearch
          placeholder="Search users by name/username"
          class="mb-2"
          @select="onPickUser"
          @error="onChildError"
        />

        <div class="picked" v-if="pickedUsers.length">
          <div
            v-for="u in pickedUsers"
            :key="String(u.id)"
            class="chip"
            :title="u.username || u.name || ''"
          >
            <img :src="avatar(u)" class="chip-avatar" alt="avatar" />
            <span class="chip-name">{{ u.name || u.username || '(user)' }}</span>
            <button class="chip-x" @click="removePicked(u)">×</button>
          </div>
        </div>
        <p v-else class="muted">Pick at least one user to create the group.</p>

        <button
          class="btn mt"
          :disabled="!groupName.trim() || pickedUsers.length === 0 || loading"
          @click="create"
        >
          {{ loading ? 'Creating…' : 'Create Group' }}
        </button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import UserSearch from '@/components/UserSearch.vue'
import {
  isAuthed,
  getAvatarUrl,
  createGroup as apiCreateGroup,
  setGroupPhoto,
} from '@/services/api'

const router = useRouter()

const err = ref('')
const loading = ref(false)

const groupName = ref('')

const pickedUsers = ref([]) // [{id, name, username, ...}]
const avatarFile = ref(null)
const avatarPreview = ref('')
const fileInput = ref(null)

const avatar = (u) => getAvatarUrl(u)
const memberIds = computed(() => pickedUsers.value.map(u => String(u.id)).filter(Boolean))

onMounted(() => {
  if (!isAuthed()) {
    router.replace('/login')
  }
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

function onPickFile(e) {
  const f = e?.target?.files?.[0]
  avatarFile.value = f || null
  if (f) {
    const url = URL.createObjectURL(f)
    avatarPreview.value = url
  } else {
    avatarPreview.value = ''
  }
}

function clearAvatar() {
  avatarFile.value = null
  avatarPreview.value = ''
  if (fileInput.value) fileInput.value.value = ''
}

async function create() {
  err.value = ''
  if (!groupName.value.trim() || memberIds.value.length === 0) return

  loading.value = true
  try {
    // 1) 创建群组（后端允许任意用户，只要你有他们的ID即可）
    const res = await apiCreateGroup({ name: groupName.value.trim(), memberIds: memberIds.value })
    // 兼容返回结构
    const groupId =
      res?.group?.id || res?.id || res?.group_id || res?._id || res?.gid || null
    const conversationId =
      res?.conversation_id || res?.conversationId || res?.conversation?.id || null

    if (!groupId) throw new Error('No group id returned')
    // 2) 如选择了群头像，立即上传
    if (avatarFile.value) {
      await setGroupPhoto(groupId, avatarFile.value)
    }

    // 3) 跳到群聊（用 conversationId）
    if (conversationId) {
      router.push({ name: 'chat', params: { type: 'conv', id: String(conversationId) } })
    } else {
      // 如果返回没有会话ID，就回到群列表（兜底）
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
  min-height: 100vh;
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe);
}
.topbar {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 18px;
  border-bottom: 1px solid #e2e8f0;
  background: #fff;
}
.title { font-weight: 800; color: #0f172a; }
.content { max-width: 900px; margin: 0 auto; padding: 18px; }

.form { display: flex; flex-direction: column; gap: 10px; }
.label { font-weight: 600; color: #334155; }
.mt { margin-top: 8px; }

.input {
  border: 1px solid #cbd5e1; border-radius: 10px; padding: .6rem .8rem; outline: none;
}
.input:focus { border-color: #22c55e; box-shadow: 0 0 0 .2rem rgba(34,197,94,.15); }

.avatar-row {
  display: flex; align-items: center; gap: 10px; flex-wrap: wrap;
}
.avatar-preview {
  width: 56px; height: 56px; border-radius: 50%; overflow: hidden; border: 1px solid #e2e8f0;
}
.avatar-preview img { width: 100%; height: 100%; object-fit: cover; }
.file { max-width: 280px; }

.btn {
  border: 0; border-radius: 10px; color: #fff; padding: .65rem 1rem;
  background-image: linear-gradient(135deg, #22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow: 0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled { opacity: .65; }
.btn-outline {
  border: 1px solid #cbd5e1; background: #fff; color: #334155; border-radius: 10px; padding: .55rem .9rem;
}

.muted { color: #64748b; font-size: .9rem; }

.picked { display: flex; gap: 8px; flex-wrap: wrap; }
.chip {
  display: inline-flex; align-items: center; gap: 6px; padding: 4px 8px;
  background: #fff; border: 1px solid #e2e8f0; border-radius: 999px;
}
.chip-avatar { width: 22px; height: 22px; border-radius: 50%; object-fit: cover; }
.chip-name { color: #0f172a; font-weight: 600; }
.chip-x {
  border: 0; background: transparent; cursor: pointer; font-size: 1rem; line-height: 1;
  color: #64748b;
}
</style>
