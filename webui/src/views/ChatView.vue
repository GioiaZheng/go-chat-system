<template>
  <div class="py-3">
    <!-- Header card -->
    <div class="d-flex align-items-center justify-content-between mb-3">
      <h2 class="h5 mb-0">Chat ({{ type }} / {{ id }})</h2>
      <RouterLink class="text-decoration-none" to="/conversations">Back</RouterLink>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />

    <!-- Messages panel -->
    <div class="card shadow-sm">
      <div class="card-body p-0 d-flex flex-column" style="height: 64vh;">
        <div class="chat-scroll flex-grow-1 p-3">
          <LoadingSpinner v-if="loading" />
          <MessageBubble
            v-for="m in messages"
            :key="m.id"
            :mine="m.sender_id === meId"
            :text="m.content"
            :meta="formatMeta(m)"
          />
        </div>

        <!-- Composer -->
        <div class="composer border-top p-2">
          <div class="d-grid gap-2" style="grid-template-columns: 1fr auto;">
            <textarea
              v-model="draft"
              class="form-control"
              placeholder="Type a message…"
              rows="1"
              @keyup.enter.exact.prevent="onSend"
            ></textarea>
            <button class="btn btn-success fw-semibold" :disabled="!draft.trim()" @click="onSend">Send</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
// Dashboard-styled chat view (English-only)
import { ref, computed, watch } from "vue";
import { useRoute } from "vue-router";
import axios from "../services/axios";
import MessageBubble from "../components/MessageBubble.vue";
import ErrorMsg from "../components/ErrorMsg.vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";

const route = useRoute();
const type = computed(() => route.params.type); // conv | private | group
const id = computed(() => route.params.id);

const messages = ref([]);
const loading = ref(false);
const err = ref("");
const draft = ref("");

const meId = ref("");
try {
  meId.value = JSON.parse(localStorage.getItem("me") || "{}")?.id || "";
} catch {}

watch([type, id], load, { immediate: true });

function formatMeta(m) {
  return `${(m.sender_id || "").slice(0, 8)} · ${m.created_at || ""}`;
}

async function load() {
  loading.value = true; err.value = ""; messages.value = [];
  try {
    if (type.value === "conv") {
      const res = await axios.get("/messages", { params: { conversation_id: id.value } });
      messages.value = res.data?.data?.messages || res.data?.messages || [];
    } else if (type.value === "private") {
      const res = await axios.get("/messages", { params: { chat_type: "private", target_id: id.value } });
      messages.value = res.data?.data?.messages || res.data?.messages || [];
    } else if (type.value === "group") {
      const res = await axios.get("/messages", { params: { chat_type: "group", target_id: id.value } });
      messages.value = res.data?.data?.messages || res.data?.messages || [];
    } else {
      err.value = "Unknown chat type";
    }
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to load messages";
  } finally {
    loading.value = false;
  }
}

async function onSend() {
  const text = draft.value.trim();
  if (!text) return;
  try {
    if (type.value === "conv") {
      await axios.post("/messages", { conversation_id: id.value, content: text });
    } else if (type.value === "private") {
      await axios.post("/messages", { chat_type: "private", target_id: id.value, content: text });
    } else if (type.value === "group") {
      await axios.post("/messages", { chat_type: "group", target_id: id.value, content: text });
    }
    draft.value = "";
    await load();
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to send message";
  }
}
</script>

<style scoped>
.chat-scroll {
  background: #f7f7f8;
  overflow: auto;
}

/* Bubbles */
.msg {
  display: flex;
  gap: .5rem;
  margin: .25rem 0;
  align-items: flex-end;
}
.msg.mine { flex-direction: row-reverse; }

.bubble {
  max-width: 70ch;
  padding: .5rem .75rem;
  border-radius: .75rem;
  background: #fff;
  box-shadow: 0 1px 1px rgba(0,0,0,.06);
  word-break: break-word;
}
.msg.mine .bubble { background: #95ec69; }
.meta {
  font-size: .75rem;
  color: #6b7280;
  margin: 0 .25rem;
  user-select: none;
}

.composer textarea { resize: vertical; min-height: 44px; }
</style>
