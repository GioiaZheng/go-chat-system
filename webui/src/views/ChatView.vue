<template>
  <div class="wrap">
    <div class="bar">
      <h2 class="h5">Chat <small>({{ type }} / {{ id }})</small></h2>
      <RouterLink class="link" to="/conversations">Back</RouterLink>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />

    <section class="panel">
      <div ref="scrollbox" class="scroll">
        <LoadingSpinner v-if="loading" />
        <div
          v-for="m in messages"
          :key="m.id"
          class="msg"
          :class="{ mine: m.sender_id === meId }"
        >
          <div class="bubble" :class="{ mine: m.sender_id === meId }">
            {{ m.content }}
          </div>
          <div class="meta">{{ formatMeta(m) }}</div>
        </div>
      </div>

      <div class="composer">
        <textarea
          v-model="draft"
          class="input"
          placeholder="Type a message…"
          rows="1"
          @keyup.enter.exact.prevent="onSend"
        ></textarea>
        <button class="btn" :disabled="!draft.trim()" @click="onSend">Send</button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from '../services/axios'
import ErrorMsg from '../components/ErrorMsg.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'

const route = useRoute()
const router = useRouter()
const type = computed(() => route.params.type) // conv | private | group
const id = computed(() => route.params.id)

const messages = ref([])
const loading = ref(false)
const err = ref('')
const draft = ref('')
const scrollbox = ref(null)

const meId = ref('')
try { meId.value = JSON.parse(localStorage.getItem('me') || '{}')?.id || '' } catch {}

function authHeaders () {
  const t = localStorage.getItem('token')
  return t ? { Authorization: `Bearer ${t}` } : {}
}

watch([type, id], load, { immediate: true })

function formatMeta(m){
  return `${(m.sender_id || '').slice(0,8)} · ${m.created_at || ''}`
}

async function load(){
  if (!localStorage.getItem('token')) { router.replace('/login'); return }
  loading.value = true; err.value = ''; messages.value = []
  try {
    if (type.value === 'conv') {
      const res = await axios.get('/messages', {
        params: { conversation_id: id.value },
        headers: authHeaders()
      })
      messages.value = res.data?.data?.messages || res.data?.messages || []
    } else if (type.value === 'private') {
      const res = await axios.get('/messages', {
        params: { chat_type: 'private', target_id: id.value },
        headers: authHeaders()
      })
      messages.value = res.data?.data?.messages || res.data?.messages || []
    } else if (type.value === 'group') {
      const res = await axios.get('/messages', {
        params: { chat_type: 'group', target_id: id.value },
        headers: authHeaders()
      })
      messages.value = res.data?.data?.messages || res.data?.messages || []
    } else {
      err.value = 'Unknown chat type'
    }
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to load messages'
  } finally {
    loading.value = false
    await nextTick()
    if (scrollbox.value) scrollbox.value.scrollTop = scrollbox.value.scrollHeight
  }
}

async function onSend(){
  const text = draft.value.trim()
  if (!text) return
  try {
    if (type.value === 'conv') {
      await axios.post('/messages',
        { conversation_id: id.value, content: text },
        { headers: authHeaders() }
      )
    } else if (type.value === 'private') {
      await axios.post('/messages',
        { chat_type: 'private', target_id: id.value, content: text },
        { headers: authHeaders() }
      )
    } else if (type.value === 'group') {
      await axios.post('/messages',
        { chat_type: 'group', target_id: id.value, content: text },
        { headers: authHeaders() }
      )
    }
    draft.value = ''
    await load()
  } catch (e) {
    err.value = e?.response?.data?.message || e?.message || 'Failed to send message'
  }
}
</script>

<style scoped>
.wrap{ max-width:900px; margin:0 auto; padding:12px 16px; }
.bar{
  display:flex; align-items:center; justify-content:space-between;
  padding:8px 0; border-bottom:1px solid #e2e8f0; margin-bottom:8px;
}
.link{ color:#2563eb; text-decoration:none; }
.link:hover{ text-decoration:underline; }

.panel{
  background:#fff; border:1px solid #e2e8f0; border-radius:14px;
  box-shadow:0 6px 18px rgba(2,6,23,.06);
  overflow:hidden;
}
.scroll{ height:64vh; overflow:auto; background:#f8fafc; padding:12px; }

/* messages */
.msg{ display:flex; gap:.5rem; margin:.25rem 0; align-items:flex-end; }
.msg.mine{ flex-direction: row-reverse; }
.bubble{
  max-width:70ch; padding:.5rem .75rem; border-radius:.75rem;
  background:#ffffff; box-shadow:0 1px 1px rgba(0,0,0,.06); word-break:break-word;
}
.bubble.mine{
  background: linear-gradient(135deg,#22c55e,#3b82f6); color:#fff;
}
.meta{ font-size:.75rem; color:#6b7280; margin:0 .25rem; user-select:none; }

.composer{
  display:grid; grid-template-columns:1fr auto; gap:10px;
  padding:10px; border-top:1px solid #e2e8f0; background:#fff;
}
.input{
  min-height:44px; resize:vertical; border-radius:10px; border:1px solid #cbd5e1;
  padding:10px 12px; outline:none;
}
.input:focus{ border-color:#22c55e; box-shadow:0 0 0 .2rem rgba(34,197,94,.15); }
.btn{
  border:0; border-radius:10px; color:#fff; padding:10px 16px;
  background-image: linear-gradient(135deg,#22c55e 0%, #16a34a 45%, #3b82f6 120%);
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn:disabled{ opacity:.65 }
</style>
