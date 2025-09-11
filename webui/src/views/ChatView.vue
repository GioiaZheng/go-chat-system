<template> 
  <div class="max-w-3xl mx-auto p-4 flex flex-col h-[90vh]">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-lg font-semibold">Chat ({{ type }} / {{ id }})</h2>
      <router-link class="text-blue-600" to="/conversations">Back</router-link>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-2" />
    <div class="flex-1 overflow-auto p-3 bg-white rounded border space-y-3">
      <LoadingSpinner v-if="loading" />
      <MessageBubble
        v-for="m in messages"
        :key="m.id"
        :mine="m.sender_id === meId"
        :text="m.content"
        :meta="formatMeta(m)"
      />
    </div>

    <MessageInput class="mt-3" @send="onSend" />
  </div>
</template>

<script setup>
// English: messages via conversation_id (new) with legacy fallback
import { ref, computed, watch } from "vue";
import { useRoute } from "vue-router";
import axios from "../services/axios";
import MessageBubble from "../components/MessageBubble.vue";
import MessageInput from "../components/MessageInput.vue";
import ErrorMsg from "../components/ErrorMsg.vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";

const route = useRoute();
const type = computed(() => route.params.type); // conv | private | group
const id = computed(() => route.params.id);

const messages = ref([]);
const loading = ref(false);
const err = ref("");

const meId = ref("");
try {
  const me = JSON.parse(localStorage.getItem("me") || "{}");
  meId.value = me?.id || "";
} catch {}

watch([type, id], load, { immediate: true });

function formatMeta(m) {
  return `${m.sender_id?.slice(0,8)} · ${m.created_at || ""}`;
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

async function onSend(text) {
  if (!text?.trim()) return;
  try {
    if (type.value === "conv") {
      await axios.post("/messages", { conversation_id: id.value, content: text });
    } else if (type.value === "private") {
      await axios.post("/messages", { chat_type: "private", target_id: id.value, content: text });
    } else if (type.value === "group") {
      await axios.post("/messages", { chat_type: "group", target_id: id.value, content: text });
    }
    await load();
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to send message";
  }
}
</script>
