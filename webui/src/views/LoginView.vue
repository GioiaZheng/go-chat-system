<template>
  <div class="center-screen">
    <div class="card">
      <h1 class="h1">Login</h1>
      <ErrorMsg v-if="err" :text="err" class="mt-2" />
      <form @submit.prevent="login" class="mt-3">
        <label class="label">Name</label>
        <input v-model.trim="name" class="input" placeholder="Your name" />
        <button class="btn block mt-3" :disabled="loading">
          <LoadingSpinner v-if="loading" style="margin-right:8px" />
          Sign in
        </button>
      </form>
      <div class="text-muted mt-3">Tip: Any name will create / login the account.</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "../services/axios";
import ErrorMsg from "../components/ErrorMsg.vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";

const router = useRouter();
const name = ref("");
const loading = ref(false);
const err = ref("");

async function login() {
  err.value = "";
  if (!name.value) { err.value = "Name is required"; return; }
  loading.value = true;
  try {
    const res = await axios.post("/session", { name: name.value });
    const token = res.data?.data?.token;
    if (!token) throw new Error("Invalid login response");
    localStorage.setItem("token", token);
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
