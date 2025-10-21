<template>
  <div class="page">
    <div class="d-flex align-items-center justify-content-between mt-3 mb-2">
      <h2 class="h5 m-0 fw-bold">My Conversations</h2>
      <button class="btn btn-success" @click="startNewChat">＋ New Chat</button>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />

    <div v-if="loading" class="text-center py-5">
      <LoadingSpinner />
    </div>

    <div v-else-if="!conversations.length" class="alert alert-light border-dashed">
      No conversations yet. Start a new chat from the button above.
    </div>

    <ul v-else class="list-unstyled">
      <li v-for="c in conversations" :key="c.id || c.conversation_id"
          class="p-3 mb-2 bg-white border rounded d-flex justify-content-between align-items-center shadow-sm"
          @click="openChat(c)" style="cursor:pointer">
        <div class="fw-semibold">{{ c.name || ('Conversation ' + (c.id || '').slice(0,8)) }}</div>
        <div class="text-muted small">{{ fmtTime(c.updated_at || c.created_at) }}</div>
      </li>
    </ul>

    <!-- New Chat Modal -->
    <div v-if="showModal" class="modal-backdrop">
      <div class="modal-card">
        <h3 class="h6 fw-bold">Start New Chat</h3>
        <p class="text-muted small mb-2">Enter user ID or username:</p>
        <input v-model.trim="newTarget" class="form-control" placeholder="e.g. u_176051..." />
        <div class="d-flex gap-2 mt-2">
          <button class="btn btn-success" @click="confirmNewChat" :disabled="loading">Start</button>
          <button class="btn btn-outline-secondary" @click="showModal=false">Cancel</button>
        </div>
        <p v-if="errModal" class="text-danger small mt-2">{{ errModal }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { getMyConversations, startConversation } from '@/services/api'

const router = useRouter()
const conversations = ref([])
const loading = ref(false)
const err = ref('')
let timer = null

const showModal = ref(false)
const newTarget = ref('')
const errModal = ref('')

function fmtTime (t) { return (t || '').toString().slice(0,19) }

async function loadConversations () {
  err.value = ''
  if (loading.value) return
  loading.value = true
  try {
    const data = await getMyConversations()
    const arr = Array.isArray(data) ? data : (data?.items || data?.conversations || [])
    conversations.value = (arr || []).filter(Boolean)
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load conversations'
  } finally {
    loading.value = false
  }
}

function openChat (c) {
  const cid = c.id || c.conversation_id
  if (cid) router.push(`/chat/conv/${cid}`)
}

function startNewChat () {
  showModal.value = true
  newTarget.value = ''
  errModal.value = ''
}

async function confirmNewChat () {
  if (!newTarget.value) { errModal.value = 'Please enter a user ID or username'; return }
  errModal.value = ''
  loading.value = true
  try {
    const data = await startConversation({ user_id: newTarget.value })
    const cid = data?.conversationId ?? data?.id ?? data?.conversation_id ?? null
    if (!cid) throw new Error('Invalid response')
    showModal.value = false
    router.push(`/chat/conv/${cid}`)
  } catch (e) {
    errModal.value = e?.response?.data?.message || e?.message || 'Failed to start chat'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadConversations()
  timer = setInterval(loadConversations, 5000)
})
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<style scoped>
.border-dashed { border:1px dashed #cbd5e1; }
.modal-backdrop{
  position:fixed; inset:0; background:rgba(0,0,0,.4);
  display:flex; align-items:center; justify-content:center; z-index:1000;
}
.modal-card{
  background:#fff; border-radius:12px; padding:20px; width:320px;
  box-shadow:0 10px 30px rgba(0,0,0,.2);
}
</style>
