<template>
  <div class="max-w-xl mx-auto p-4">
    <h2 class="text-xl font-semibold mb-4">My Profile</h2>
    <ErrorMsg v-if="err" :text="err" class="mb-3" />
    <div v-if="me" class="space-y-2 bg-white p-4 rounded border">
      <div><span class="font-medium">id:</span> {{ me.id }}</div>
      <div><span class="font-medium">username:</span> {{ me.username }}</div>
      <div><span class="font-medium">name:</span> {{ me.name }}</div>
      <div><span class="font-medium">email:</span> {{ me.email }}</div>
      <div><span class="font-medium">gender:</span> {{ me.gender }}</div>

      <div class="mt-4">
        <label class="block mb-1 text-sm">Change username</label>
        <input v-model="newName" class="input" placeholder="new username" />
        <button class="btn mt-2" @click="setUsername" :disabled="loading">Save</button>
      </div>

      <div class="mt-4">
        <label class="block mb-1 text-sm">Set photo</label>
        <input type="file" @change="onFile" />
        <button class="btn mt-2" @click="setPhoto" :disabled="loading || !file">Upload</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "../services/axios";
import ErrorMsg from "../components/ErrorMsg.vue";

const me = ref(null);
const newName = ref("");
const file = ref(null);
const loading = ref(false);
const err = ref("");

onMounted(async () => {
  try {
    const res = await axios.get("/users/me");
    me.value = res.data?.data?.user || res.data?.user || null;
    if (me.value) localStorage.setItem("me", JSON.stringify(me.value));
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to load profile";
  }
});

function onFile(e) {
  file.value = e.target.files?.[0] || null;
}

async function setUsername() {
  if (!newName.value) return;
  loading.value = true; err.value = "";
  try {
    await axios.put("/users/set_username", { username: newName.value });
    const res = await axios.get("/users/me");
    me.value = res.data?.data?.user || me.value;
    localStorage.setItem("me", JSON.stringify(me.value));
    newName.value = "";
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to set username";
  } finally {
    loading.value = false;
  }
}

async function setPhoto() {
  if (!file.value) return;
  loading.value = true; err.value = "";
  try {
    const form = new FormData();
    form.append("upload", file.value);
    await axios.put("/users/set_photo", form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    const res = await axios.get("/users/me");
    me.value = res.data?.data?.user || me.value;
    localStorage.setItem("me", JSON.stringify(me.value));
    file.value = null;
  } catch (e) {
    err.value = e?.response?.data?.message || "Failed to set photo";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.input { width:100%; border:1px solid #ddd; padding:.5rem .75rem; border-radius:.375rem; }
.btn { background:#111827; color:#fff; padding:.5rem .75rem; border-radius:.375rem; }
</style>
