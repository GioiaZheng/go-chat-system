<template>
  <div class="max-w-3xl mx-auto p-4">
    <h2 class="text-xl font-semibold mb-4">Groups</h2>
    <ErrorMsg v-if="err" :text="err" class="mb-3" />
    <div class="mb-4 p-3 bg-white rounded border">
      <h3 class="font-medium mb-2">Create Group</h3>
      <div class="flex gap-2">
        <input v-model="groupName" class="input flex-1" placeholder="Group name" />
        <input v-model="memberIdsRaw" class="input flex-1" placeholder="member_ids (comma separated)" />
        <button class="btn" @click="createGroup" :disabled="loading">Create</button>
      </div>
      <div class="text-xs text-gray-500 mt-1">English: member_ids must include at least yourself</div>
    </div>

    <LoadingSpinner v-if="loading" />
    <ul v-else class="space-y-2">
      <li v-for="g in groups" :key="g.id" class="p-3 bg-white rounded border">
        <div class="flex items-center justify-between">
          <div>
            <div class="font-medium">{{ g.name }}</div>
            <div class="text-sm text-gray-500">conversation_id: {{ g.conversation_id || '(unknown)' }}</div>
          </div>
          <router-link
            v-if="g.conversation_id"
            class="text-blue-600"
            :to="{ name: 'chat', params: { type: 'conv', id: g.conversation_id } }"
          >
            Open
          </router-link>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "../services/axios";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import ErrorMsg from "../components/ErrorMsg.vue";

const groups = ref([]);
const loading = ref(false);
const err = ref("");

const groupName = ref("");
const memberIdsRaw = ref("");

onMounted(load);

async function load() {
  loading.value = true; err.value = "";
  try {
    const res = await axios.get("/groups");
    groups.value = res.data?.items || res.data?.data?.items || [];
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to load groups";
  } finally {
    loading.value = false;
  }
}

async function createGroup() {
  err.value = "";
  const member_ids = memberIdsRaw.value
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

  if (!groupName.value || member_ids.length === 0) {
    err.value = "Group name & member_ids are required";
    return;
  }
  loading.value = true;
  try {
    await axios.post("/groups", { name: groupName.value, member_ids });
    await load();
    groupName.value = "";
    memberIdsRaw.value = "";
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to create group";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.input { width:100%; border:1px solid #ddd; padding:.5rem .75rem; border-radius:.375rem; }
.btn { background:#111827; color:#fff; padding:.5rem .75rem; border-radius:.375rem; }
</style>
