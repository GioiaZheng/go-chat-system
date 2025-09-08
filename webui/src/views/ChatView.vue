<template>
  <div class="chat-page">
    <header class="bar">
      <button class="back" @click="$router.back()">← Back</button>
      <h2 class="title">{{ convTitle }}</h2>
    </header>

    <section class="msgs">
      <LoadingSpinner v-if="loading" text="Loading messages..." />
      <div v-else-if="!messages.length" class="empty">No messages yet.</div>

      <MessageBubble
        v-for="m in messages"
        :key="m.id"
        :msg="m"
        :me-id="meId"
        :show-comments="Boolean(m._showComments)"
        :comments="m._comments || []"
        @toggle-comments="toggleComments"
        @comment="onComment"
        @uncomment="onUncomment"
        @forward="onForward"
        @delete="onDelete"
      />
    </section>

    <footer class="input">
      <MessageInput @send="onSend" />
    </footer>
  </div>
</template>

<script>
/**
 * ChatView
 * - Adds lazy-loaded comments (getMessageComments) and actions:
 *   * commentMessage / uncommentMessage
 *   * forwardMessage (to userId or groupId)
 */

import { ref, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import {
  getConversation,
  sendMessage,
  deleteMessage,
  commentMessage,
  uncommentMessage,
  forwardMessage,
  getMessageComments,
  me,
} from "../services/api";

import LoadingSpinner from "../components/LoadingSpinner.vue";
import MessageBubble from "../components/MessageBubble.vue";
import MessageInput from "../components/MessageInput.vue";

export default {
  name: "ChatView",
  components: { LoadingSpinner, MessageBubble, MessageInput },
  setup() {
    const route = useRoute();

    const chatType = ref(route.params.type);
    const targetId = ref(route.params.id);

    const loading = ref(true);
    const messages = ref([]);
    const convTitle = ref("");
    const meId = ref(localStorage.getItem("token"));

    onMounted(async () => {
      try {
        const r = await me();
        meId.value = r?.data?.user?.id || meId.value;
      } catch (e) {
        console.warn("Failed to fetch me:", e);
      }
      await fetchMsgs();
    });

    watch(
      () => route.fullPath,
      async () => {
        chatType.value = route.params.type;
        targetId.value = route.params.id;
        await fetchMsgs();
      }
    );

    async function fetchMsgs() {
      loading.value = true;
      try {
        const r = await getConversation({
          chatType: chatType.value,
          targetId: targetId.value,
        });
        const list = r?.data?.messages || r?.data || [];
        // Reset local comment toggles for fresh load
        messages.value = list.map((m) => ({ ...m, _showComments: false, _comments: [] }));
        convTitle.value = r?.data?.name || `${chatType.value} #${targetId.value}`;
      } catch (e) {
        console.error("getConversation failed:", e);
        messages.value = [];
      } finally {
        loading.value = false;
      }
    }

    async function onSend(text) {
      try {
        await sendMessage({ chatType: chatType.value, targetId: targetId.value, content: text });
        await fetchMsgs();
      } catch (e) {
        console.error("sendMessage failed:", e);
      }
    }

    async function onDelete(m) {
      try {
        await deleteMessage(m.id);
        await fetchMsgs();
      } catch (e) {
        console.error("deleteMessage failed:", e);
      }
    }

    // Toggle comments display for a message (lazy load)
    async function toggleComments(m) {
      m._showComments = !m._showComments;
      if (m._showComments && (!m._comments || !m._comments.length)) {
        try {
          const r = await getMessageComments(m.id);
          // Expected: { data: { comments: [...] } } or { data: [...] }
          m._comments = r?.data?.comments || r?.data || [];
        } catch (e) {
          console.error("getMessageComments failed:", e);
          m._comments = [];
        }
      }
    }

    // Add a comment/reaction
    async function onComment(m) {
      try {
        const text = prompt("Enter comment/reaction:");
        if (!text) return;
        await commentMessage(m.id, text);
        // Refresh comments if panel is open
        if (m._showComments) {
          const r = await getMessageComments(m.id);
          m._comments = r?.data?.comments || r?.data || [];
        }
      } catch (e) {
        console.error("commentMessage failed:", e);
      }
    }

    // Remove my comment/reaction
    async function onUncomment(m) {
      try {
        await uncommentMessage(m.id);
        if (m._showComments) {
          const r = await getMessageComments(m.id);
          m._comments = r?.data?.comments || r?.data || [];
        }
      } catch (e) {
        console.error("uncommentMessage failed:", e);
      }
    }

    // Forward message: ask whether to a user or a group (simple prompt UX)
    async function onForward(m) {
      try {
        const dest = prompt("Forward to: type 'user:<id>' or 'group:<id>'");
        if (!dest) return;
        let toUserId, toGroupId;
        if (dest.startsWith("user:")) {
          toUserId = dest.split(":")[1];
        } else if (dest.startsWith("group:")) {
          toGroupId = dest.split(":")[1];
        } else {
          alert("Invalid format. Use 'user:123' or 'group:456'.");
          return;
        }
        await forwardMessage(m.id, { toUserId, toGroupId });
        // Optional: show a toast or small confirmation
        console.log("Message forwarded.");
      } catch (e) {
        console.error("forwardMessage failed:", e);
      }
    }

    return {
      chatType,
      targetId,
      loading,
      messages,
      convTitle,
      meId,
      onSend,
      onDelete,
      toggleComments,
      onComment,
      onUncomment,
      onForward,
    };
  },
};
</script>

<style scoped>
.chat-page { display: grid; grid-template-rows: auto 1fr auto; height: 100vh; }
.bar { display: flex; align-items: center; gap: 12px; padding: 12px; border-bottom: 1px solid #eee; }
.back { border: 0; background: transparent; cursor: pointer; font-size: 16px; }
.title { margin: 0; font-size: 18px; font-weight: 700; }
.msgs { overflow: auto; padding: 12px; display: flex; flex-direction: column; gap: 8px; }
.empty { padding: 16px; opacity: 0.7; }
.input { padding: 12px; border-top: 1px solid #eee; }
</style>
