<template>
  <div class="page">
    <!-- Top bar -->
    <header class="bar">
      <h2 class="title">Groups</h2>

      <!-- Create a new group -->
      <form class="new" @submit.prevent="create">
        <input
          v-model="newName"
          placeholder="New group name"
          :disabled="creating"
          required
        />
        <button :disabled="creating">{{ creating ? "Creating…" : "Create" }}</button>
      </form>
    </header>

    <!-- Loading -->
    <LoadingSpinner v-if="loading" text="Loading groups..." />

    <!-- Empty -->
    <div v-else-if="!groups.length" class="empty">No groups yet.</div>

    <!-- List -->
    <ul v-else class="list">
      <li v-for="g in groups" :key="g.id">
        <button class="row" @click="open(g)">
          <img v-if="g.avatarUrl" :src="g.avatarUrl" class="avatar" alt="group" />
          <div class="main">
            <div class="top">
              <strong class="name">{{ g.name || ("Group #" + g.id) }}</strong>
              <span class="id">#{{ g.id }}</span>
            </div>
            <div class="meta">
              {{ (g.members || []).length }} members
            </div>
          </div>
        </button>
      </li>
    </ul>

    <!-- Drawer / modal style editor -->
    <GroupEditor
      v-if="active"
      :group-id="active.id"
      @close="active = null; refresh()"
    />
  </div>
</template>

<script>
/**
 * GroupsView
 * - English comments for TA/teacher clarity.
 * - Displays groups the user belongs to.
 * - Allows creating a new group (server auto-adds the creator as a member).
 * - Clicking a group opens GroupEditor (rename, add members, change photo, leave).
 */

import { ref, onMounted } from "vue";
import { listGroups, createGroup } from "../services/api";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import GroupEditor from "../components/GroupEditor.vue";

export default {
  name: "GroupsView",
  components: { LoadingSpinner, GroupEditor },
  setup() {
    const loading = ref(true);
    const groups = ref([]);
    const newName = ref("");
    const creating = ref(false);
    const active = ref(null); // currently opened group

    onMounted(refresh);

    async function refresh() {
      loading.value = true;
      try {
        const r = await listGroups();
        // Expected: { data: [...] } or { data: { groups: [...] } }
        groups.value = r?.data?.groups || r?.data || [];
      } catch (e) {
        console.error("listGroups failed:", e);
        groups.value = [];
      } finally {
        loading.value = false;
      }
    }

    async function create() {
      if (!newName.value.trim()) return;
      creating.value = true;
      try {
        // Create an empty group first (members can be added in editor)
        await createGroup({ name: newName.value.trim(), members: [] });
        newName.value = "";
        await refresh();
      } catch (e) {
        console.error("createGroup failed:", e);
      } finally {
        creating.value = false;
      }
    }

    function open(g) {
      active.value = g;
    }

    return { loading, groups, newName, creating, active, refresh, create, open };
  },
};
</script>

<style scoped>
.page { display: grid; grid-template-rows: auto 1fr; height: 100vh; }
.bar {
  display: flex; align-items: center; gap: 12px;
  padding: 12px; border-bottom: 1px solid #eee;
}
.title { margin: 0; font-size: 18px; font-weight: 700; }
.new { display: flex; gap: 8px; }
.empty { padding: 16px; opacity: 0.7; }
.list { list-style: none; margin: 0; padding: 0; overflow: auto; }
.row {
  width: 100%; display: grid; grid-template-columns: 40px 1fr;
  gap: 12px; align-items: center; padding: 12px;
  border: 0; border-bottom: 1px solid #f0f0f0; background: transparent; text-align: left;
}
.avatar { width: 40px; height: 40px; border-radius: 12px; object-fit: cover; }
.top { display: flex; justify-content: space-between; gap: 8px; }
.name { overflow: hidden; text-overflow: ellipsis; }
.id { font-size: 12px; opacity: 0.6; }
.meta { font-size: 12px; opacity: 0.8; }
</style>
