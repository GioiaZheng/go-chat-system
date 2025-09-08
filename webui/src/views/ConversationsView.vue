<template>
  <div class="page">
    <header class="bar">
      <h2 class="title">Conversations</h2>

      <!-- User search to start a new conversation -->
      <UserSearchBox
        :busy="searching"
        :results="results"
        @query="onSearch"
        @pick="onPickUser"
        @clear="clearResults"
      />
    </header>

    <!-- Loading state -->
    <LoadingSpinner v-if="loading" text="Loading conversations..." />

    <!-- Empty state -->
    <div v-else-if="!sortedConversations.length" class="empty">
      No conversations yet. Try searching a user to start one.
    </div>

    <!-- Conversation list -->
    <ul v-else class="list">
      <ConversationItem
        v-for="c in sortedConversations"
        :key="c.id || c.conversationId"
        :item="c"
        @open="openConversation"
      />
    </ul>
  </div>
</template>

<script>
/**
 * ConversationsView
 * - English comments for TA/reviewers.
 * - Fetches user's conversations and displays the latest first (desc by last message time).
 * - Allows searching users and starting a new conversation with one click.
 * - Clicking a conversation navigates to /chat/:type/:id.
 *
 * Data shape assumptions (align with your backend's models.go / handler):
 * Each conversation item may include:
 *   - id (or conversationId)
 *   - type: "private" | "group"
 *   - targetId: userId (for private) or groupId (for group)
 *   - name: username or group name
 *   - avatarUrl: optional avatar
 *   - lastMessage: preview text (if image, we show an icon)
 *   - lastMessageAt: timestamp (ISO string or unix)
 *
 * If your backend uses slightly different property names, we normalize in computed getter.
 */

import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import {
  getMyConversations,
  startConversation,
  searchUsers,
} from "../services/api";

import ConversationItem from "../components/ConversationItem.vue";
import UserSearchBox from "../components/UserSearchBox.vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";

export default {
  name: "ConversationsView",
  components: { ConversationItem, UserSearchBox, LoadingSpinner },
  setup() {
    const router = useRouter();

    // --- State ---
    const loading = ref(true);     // fetch conversations loading
    const conversations = ref([]); // raw conversations from API
    const searching = ref(false);  // search loading
    const results = ref([]);       // user search results

    // --- Lifecycle: initial fetch ---
    onMounted(async () => {
      await refresh();
    });

    // Fetch conversations from backend
    async function refresh() {
      loading.value = true;
      try {
        const res = await getMyConversations();
        // Expected: { data: { conversations: [...] } } or { data: [...] }
        conversations.value =
          res?.data?.conversations ||
          res?.data ||
          [];
      } catch (e) {
        console.error("getMyConversations failed:", e);
        conversations.value = [];
      } finally {
        loading.value = false;
      }
    }

    // Normalize & sort conversations by last message time (desc)
    const sortedConversations = computed(() => {
      const list = (conversations.value || []).map((c) => {
        // Normalize fields to be robust against backend variations
        const id = c.id ?? c.conversationId ?? c.convId ?? c.targetId;
        const type = c.type ?? c.chatType ?? (c.groupId ? "group" : "private");
        const targetId = c.targetId ?? c.userId ?? c.groupId ?? c.id;
        const name = c.name ?? c.username ?? c.groupName ?? String(targetId);
        const avatarUrl = c.avatarUrl ?? c.photoUrl ?? null;
        const lastMessage = c.lastMessage ?? c.preview ?? "";
        const lastMessageAt =
          c.lastMessageAt ??
          c.updatedAt ??
          c.lastAt ??
          c.time ??
          0;

        return { ...c, id, type, targetId, name, avatarUrl, lastMessage, lastMessageAt };
      });

      // Sort by lastMessageAt desc
      return list.sort((a, b) => {
        const ta = new Date(a.lastMessageAt).getTime() || 0;
        const tb = new Date(b.lastMessageAt).getTime() || 0;
        return tb - ta;
      });
    });

    // --- Open a conversation -> navigate to /chat/:type/:id ---
    function openConversation(item) {
      const type = item.type || "private";
      const id = item.targetId ?? item.id;
      router.push(`/chat/${type}/${encodeURIComponent(id)}`);
    }

    // --- Search a user by username (debounced externally if needed) ---
    let searchTimer = null;
    function onSearch(q) {
      clearTimeout(searchTimer);
      if (!q) {
        results.value = [];
        return;
      }
      searchTimer = setTimeout(async () => {
        searching.value = true;
        try {
          const res = await searchUsers(q);
          // Expected: { data: { users: [...] } } or { data: [...] }
          results.value = res?.data?.users || res?.data || [];
        } catch (e) {
          console.error("searchUsers failed:", e);
          results.value = [];
        } finally {
          searching.value = false;
        }
      }, 250);
    }

    function clearResults() {
      results.value = [];
    }

    // --- Pick a user from search results & start a private conversation ---
    async function onPickUser(user) {
      try {
        // Private conversation: one member (the other side). Server adds me automatically.
        await startConversation({ name: null, memberIds: [user.id] });
        await refresh(); // refresh list to show the new conversation

        // Find the conversation that targets this user (fallback to first if not found)
        const conv =
          sortedConversations.value.find(
            (c) => c.type === "private" && String(c.targetId) === String(user.id)
          ) || sortedConversations.value[0];

        if (conv) openConversation(conv);
      } catch (e) {
        console.error("startConversation failed:", e);
      }
    }

    return {
      loading,
      conversations,
      sortedConversations,
      searching,
      results,
      onSearch,
      onPickUser,
      clearResults,
      openConversation,
    };
  },
};
</script>

<style scoped>
.page {
  display: grid;
  grid-template-rows: auto 1fr;
  height: 100vh;
}
.bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-bottom: 1px solid #eee;
}
.title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
}
.empty {
  padding: 16px;
  opacity: 0.7;
}
.list {
  list-style: none;
  margin: 0;
  padding: 0;
  overflow: auto;
}
</style>
