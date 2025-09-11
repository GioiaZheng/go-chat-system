<template>
  <div>
    <h1>Login</h1>
    <form @submit.prevent="login">
      <input v-model="name" placeholder="alice" />
      <button>Login</button>
    </form>
    <p v-if="err" style="color:red">{{ err }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import api from '@/services/axios';

const router = useRouter();
const name = ref('alice');
const err = ref('');

async function login() {
  err.value = '';
  try {
    const resp = await api.post('/session', { name: name.value.trim() });
    const token = resp?.data?.data?.token;
    if (!token) throw new Error('No token in response');
    // Store as Bearer token for subsequent API calls
    localStorage.setItem('token', token);
    await router.push('/chat');
  } catch (e) {
    err.value = e?.response?.data?.message || e.message || 'Login failed';
  }
}
</script>
