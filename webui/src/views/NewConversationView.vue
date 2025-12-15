<!-- src/views/NewConversationView.vue: Lightweight flow to start a private chat. -->
<template>
  <div class="wrap">
    <h2 class="title">Start a New Conversation</h2>

    <ErrorMsg v-if="err" :text="err" class="mb-3" />

    <!-- Search and pick a user -->
    <section class="card">
      <h3 class="h6">Find a user</h3>
      <!-- Reuses the shared search component (emits @select when a user is chosen). -->
      <UserSearch class="mb-2" @select="onPick" @error="onChildError" />

      <div v-if="picked" class="picked">
        <div class="who">
          <div class="name">
            {{ picked.username || picked.email || picked.name || ('User ' + (picked.id || '').slice(0,8)) }}
          </div>
          <div class="sub">id: {{ picked.id }}</div>
        </div>
        <button class="btn" :disabled="loading" @click="start">
          <span v-if="loading">Starting…</span>
          <span v-else>Start conversation</span>
        </button>
      </div>

      <p v-else class="muted">Search and select a user above to start a chat.</p>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import UserSearch from '@/components/UserSearch.vue'
import { isAuthed, startPrivateConversation } from '@/services/api'

const router = useRouter()
const picked = ref(null)
const loading = ref(false)
const err = ref('')

onMounted(() => {
  if (!isAuthed()) router.replace('/login')
})

function onPick(u) {
  picked.value = u
  err.value = ''
}

function onChildError(msg) {
  err.value = msg || ''
}

// Create a private chat and navigate using api.startPrivateConversation.
async function start () {
  const id = String(picked.value?.id || '')
  if (!id) return
  loading.value = true
  err.value = ''
  try {
    const res = await startPrivateConversation({ id }) // Passing the raw id string also works.
    const cid =
      res?.conversation?.id ||
      res?.conversationId ||
      res?.conversation_id ||
      res?.id ||
      res?.cid ||
      String(res)

    if (!cid) throw new Error('No conversation id returned')
    router.push({ name: 'chat', params: { type: 'conv', id: cid } })
  } catch (e) {
    if (e?.response?.status === 401) {
      err.value = 'Unauthorized. Please login again.'
      router.push('/login')
    } else {
      err.value = e?.response?.data?.message || e?.message || 'Failed to start conversation'
    }
  } finally {
    loading.value = false
  }
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

.picked{ display:flex; align-items:center; justify-content:space-between; gap:12px; margin-top:6px; }
.who .name{ font-weight:600; color:#0f172a }
.who .sub{ color:#64748b; font-size:.9rem }
.muted{ margin-top:.35rem; color:#64748b; font-size:.9rem }

.btn{
  border:0; border-radius:10px; color:#fff; padding:.55rem .9rem;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65 }
</style>
