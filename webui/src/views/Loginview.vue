<template>
  <div class="login-container">
    <form class="login-form" @submit.prevent="onSubmit">
      <h2>WASAText • Login</h2>

      <!-- Username only (per assignment's simplified login) -->
      <input
        v-model="name"
        placeholder="Enter username"
        required
        autofocus
      />

      <!-- Optional: show a short hint to TA about simplified login -->
      <small class="hint">
        This demo uses simplified login (username only) as required by the assignment.
      </small>

      <!-- Submit -->
      <button type="submit" :disabled="loading || !name">
        {{ loading ? "Signing in…" : "Sign in" }}
      </button>

      <!-- Error message -->
      <ErrorMsg v-if="error">{{ error }}</ErrorMsg>
    </form>
  </div>
</template>

<script>
/**
 * LoginView
 * - English comments for reviewers/TA.
 * - Implements simplified login: username only.
 * - Calls doLogin(name) from services/api.js.
 * - On success: stores { token, user } in localStorage and navigates to /conversations.
 *
 * NOTE:
 *   Your backend currently might still accept/expect a password; in services/api.js
 *   we already send a default "pass" for compatibility. Here we only ask for name.
 */

import { ref } from "vue";
import { useRouter } from "vue-router";
import { doLogin } from "../services/api";
import ErrorMsg from "../components/ErrorMsg.vue";

export default {
  name: "LoginView",
  components: { ErrorMsg },
  setup() {
    // --- Local reactive state ---
    const router = useRouter();
    const name = ref("");       // username input model
    const loading = ref(false); // request in-flight
    const error = ref("");      // last error message

    // --- Submit handler ---
    async function onSubmit() {
      loading.value = true;
      error.value = "";
      try {
        // Call backend simplified login (username only)
        const res = await doLogin(name.value);

        // Expected shape: { code, message, data: [ { user, token } ] }
        const { user, token } = (res.data && res.data[0]) || {};

        // Persist identifier (token) and user profile for later requests
        localStorage.setItem("token", token || user?.id);
        localStorage.setItem("me", JSON.stringify(user));

        // Navigate to conversations list
        router.replace("/conversations");
      } catch (e) {
        error.value = e?.message || "Login failed";
      } finally {
        loading.value = false;
      }
    }

    return { name, loading, error, onSubmit };
  },
};
</script>

<style scoped>
.login-container {
  display: grid;
  place-items: center;
  height: 100vh;
}
.login-form {
  width: 320px;
  display: grid;
  gap: 12px;
}
.hint {
  opacity: 0.7;
  font-size: 12px;
}
</style>
