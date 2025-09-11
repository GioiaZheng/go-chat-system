<template>
  <div>
    <div class="flex gap-2">
      <input
        v-model="q"
        @keyup.enter="search"
        class="input flex-1"
        placeholder="Search users..."
      />
      <button class="btn" @click="search">Search</button>
    </div>
    <ul v-if="items.length" class="mt-2 space-y-1">
      <li
        v-for="u in items"
        :key="u.id"
        class="p-2 bg-white border rounded flex justify-between"
      >
        <span>{{ u.username || u.email }}</span>
        <button class="text-blue-600" @click="$emit('select', u)">Select</button>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref } from "vue";
import axios from "../services/axios";

const q = ref("");
const items = ref([]);

async function search() {
  if (!q.value.trim()) {
    items.value = [];
    return;
  }
  const res = await axios.get("/users/search", { params: { q: q.value } });
  items.value = res.data?.items || res.data?.data?.items || [];
}
</script>

<style scoped>
.input {
  width: 100%;
  border: 1px solid #ddd;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
}
.btn {
  background: #111827;
  color: #fff;
  padding: 0.5rem 0.75rem;
  border-radius: 0.375rem;
}
</style>
