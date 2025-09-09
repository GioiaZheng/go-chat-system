<template>
  <div class="max-w-md mx-auto pt-20">
    <h1 class="text-2xl font-semibold mb-6">Login</h1>
    <ErrorMsg v-if="err" :text="err" class="mb-3" />
    <form @submit.prevent="login">
      <label class="block mb-2 text-sm">Name</label>
      <input v-model.trim="name" class="input" placeholder="Your name" />
      <button class="btn mt-4 w-full" :disabled="loading">
        <LoadingSpinner v-if="loading" class="inline-block mr-2" />
        Sign in
      </button>
    </form>
  </div>
</template>

<script setup>
// English: login to /session; store data.token + goto conversations
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "../services/axios";

const router = useRouter();
const name = ref("");
const loading = ref(false);
const err = ref("");

async function login() {
  err.value = "";
  if (!name.value) {
    err.value = "Name is required";
    return;
  }
  loading.value = true;
  try {
    const res = await axios.post("/session", { name: name.value });
    const token = res.data?.data?.token;
    if (!token) throw new Error("Invalid login response");
    localStorage.setItem("token", token);
    // optional: store user profile
    const user = res.data?.data?.user;
    if (user) localStorage.setItem("me", JSON.stringify(user));
    router.push({ name: "conversations" });
  } catch (e) {
    err.value = e?.response?.data?.message || e.message || "Login failed";
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.input { width:100%; border:1px solid #ddd; padding:.5rem .75rem; border-radius:.375rem; }
.btn { background:#111827; color:#fff; padding:.5rem .75rem; border-radius:.375rem; }
</style>
