<template>
  <div class="py-3">
    <div class="d-flex align-items-center justify-content-between mb-3">
      <h2 class="h5 mb-0">My Conversations</h2>
      <button class="btn btn-outline-secondary btn-sm" @click="load">Refresh</button>
    </div>

    <ErrorMsg v-if="err" :text="err" class="mb-3" />
    <LoadingSpinner v-if="loading" />

    <div v-else class="card shadow-sm">
      <ul class="list-group list-group-flush">
        <li
          v-for="c in items"
          :key="c.id"
          class="list-group-item d-flex justify-content-between align-items-center"
        >
          <div>
            <div class="fw-semibold">{{ c.name || ('Conversation ' + c.id.slice(0,8)) }}</div>
            <div class="text-muted small">id: {{ c.id }}</div>
          </div>
          <RouterLink
            class="text-decoration-none"
            :to="{ name: 'chat', params: { type: 'conv', id: c.id } }"
          >
            Open
          </RouterLink>
        </li>
        <li v-if="!items.length" class="list-group-item text-muted small">
          No conversations yet. Start a new chat from the Groups page.
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "../services/axios";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import ErrorMsg from "../components/ErrorMsg.vue";

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
    err.value = e?.response?.data?.message || "failed to fetch conversations";
  } finally {
    loading.value = false;
  }
}
</script>
