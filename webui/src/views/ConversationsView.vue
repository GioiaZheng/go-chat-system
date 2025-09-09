<template>
  <div class="max-w-3xl mx-auto p-4">
    <h2 class="text-xl font-semibold mb-4">My Conversations</h2>
    <ErrorMsg v-if="err" :text="err" class="mb-3" />
    <LoadingSpinner v-if="loading" />
    <ul v-else class="space-y-2">
      <li v-for="c in items" :key="c.id" class="p-3 bg-white rounded border">
        <div class="flex items-center justify-between">
          <div>
            <div class="font-medium">{{ c.name || ('Conversation ' + c.id.slice(0,8)) }}</div>
            <div class="text-sm text-gray-500">id: {{ c.id }}</div>
          </div>
          <router-link class="text-blue-600"
            :to="{ name: 'chat', params: { type: 'conv', id: c.id } }">
            Open
          </router-link>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup>
// English: GET /conversations; route to /chat/conv/:id
import { ref, onMounted } from "vue";
import axios from "../services/axios";

const items = ref([]);
const loading = ref(false);
const err = ref("");

onMounted(load);

async function load() {
  loading.value = true; err.value = "";
  try {
    const res = await axios.get("/conversations");
    items.value = res.data?.items || res.data?.data?.items || [];
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to load conversations";
  } finally {
    loading.value = false;
  }
}
</script>
