<!-- Notes:
     - Centered login card using Bootstrap (kept minimal).
     - POST /session { name }; stores token in localStorage as Bearer token.
     - On success, redirects to /conversations.
     - Shows normalized error message when request fails. -->

<template>
  <div class="container d-flex align-items-center justify-content-center min-vh-100">
    <div class="card shadow-sm" style="max-width: 420px; width: 100%;">
      <div class="card-body">
        <h1 class="h4 mb-3 text-center">Sign in</h1>

        <ErrorMsg v-if="err" :text="err" class="mb-3" />

        <form @submit.prevent="login" novalidate>
          <div class="mb-3">
            <label class="form-label">Name</label>
            <input
              v-model="name"
              type="text"
              class="form-control"
              placeholder="alice"
              autocomplete="username"
              required
            />
          </div>

          <button class="btn btn-dark w-100" :disabled="busy">
            <span v-if="busy" class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
            {{ busy ? 'Signing in…' : 'Login' }}
          </button>
        </form>

        <p class="text-muted mt-3 mb-0 small text-center">
          This is a minimal login for evaluation purposes.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import api from '@/services/axios';
import ErrorMsg from '@/components/ErrorMsg.vue';

const router = useRouter();
const name = ref('alice');
const err = ref('');
const busy = ref(false);

async function login() {
  err.value = '';
  busy.value = true;
  try {
    const resp = await api.post('/session', { name: name.value.trim() });
    const token = resp?.data?.data?.token;
    if (!token) throw new Error('No token in response');
    localStorage.setItem('token', token);
    await router.push('/conversations');
  } catch (e) {
    err.value = e.uiMessage || e?.response?.data?.message || e.message || 'Login failed';
  } finally {
    busy.value = false;
  }
}
</script>
