<!-- src/views/LoginView.vue -->
<template>
  <div class="auth-wrap">
    <div class="auth-card">
      <div class="brand">WASA <span class="grad">Chat</span></div>
      <h1 class="subtitle">Sign in</h1>

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <form @submit.prevent="login" novalidate>
        <label class="form-label fw-medium">Name</label>
        <input
          v-model.trim="name"
          type="text"
          class="form-control form-control-lg"
          placeholder="Type your name"
          autocomplete="username"
          required
          minlength="1"
          maxlength="16"
          autofocus
          :disabled="busy"
        />

        <button
          type="submit"
          class="btn btn-gradient btn-lg w-100 mt-3"
          :disabled="busy || name.length < 1"
        >
          <span
            v-if="busy"
            class="spinner-border spinner-border-sm me-2"
            role="status"
            aria-hidden="true"
          ></span>
          {{ busy ? 'Signing in…' : 'Login' }}
        </button>
      </form>

      <p class="text-center text-muted small mt-3 mb-0">
        Press <kbd>Enter</kbd> to login
      </p>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import api from '@/services/api'

const router = useRouter()
const route = useRoute()

const name = ref('')
const err  = ref('')
const busy = ref(false)

onMounted(() => {
  document.body.classList.add('theme-light-login')

  // Allow prefill from ?name=xxx query param
  const preset = String(route.query.name || '').trim()
  if (preset) name.value = preset

  // Redirect immediately if already logged in
  const authed = !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))
  if (authed) router.replace('/conversations')
})

onUnmounted(() => {
  document.body.classList.remove('theme-light-login')
})

async function login () {
  if (!name.value || busy.value) return
  err.value = ''
  busy.value = true

  try {
    // Clear any stale tokens
    localStorage.removeItem('token')
    sessionStorage.removeItem('authToken')

    // Login (api client stores the token internally)
    await api.doLogin(name.value)

    // Keep other views in sync with the chosen username
    localStorage.setItem('username', name.value)
    localStorage.setItem('name', name.value)
    localStorage.setItem('me', JSON.stringify({ username: name.value }))

    // Notify other views to refresh
    window.dispatchEvent(new Event('auth:changed'))

    // Support ?redirect=/xxx
    const next = (route.query.redirect && String(route.query.redirect)) || '/conversations'
    router.replace(next)
  } catch (e) {
    err.value =
      e?.response?.data?.error ||
      e?.response?.data?.message ||
      e?.message ||
      'Login failed'
  } finally {
    busy.value = false
  }
}
</script>

<!-- Global background (not scoped) -->
<style>
body.theme-light-login{
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe) !important;
  color: #0f172a;
}
</style>

<!-- Component styles -->
<style scoped>
.auth-wrap{
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
}
.auth-card{
  width: 100%;
  max-width: 460px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(2,6,23,.08);
  padding: 28px 24px;
  color: #0f172a;
}
.brand{
  font-weight: 800;
  font-size: 28px;
  letter-spacing: .5px;
  text-align: center;
}
.brand .grad{
  background: linear-gradient(90deg, #22c55e, #3b82f6);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.subtitle{
  text-align: center;
  color: #475569;
  font-weight: 700;
  margin: 6px 0 18px;
}
.form-label{ color:#334155 }
.form-control{
  background:#fff;
  border-color: #cbd5e1;
  color:#0f172a;
}
.form-control::placeholder{ color:#94a3b8 }
.form-control:focus{
  border-color:#22c55e;
  box-shadow: 0 0 0 .25rem rgba(34,197,94,.15);
}
.btn-gradient{
  background-image: linear-gradient(135deg, #22c55e 0%, #16a34a 45%, #3b82f6 120%);
  color:#fff;
  border:0;
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
}
.btn-gradient:disabled{ opacity:.6 }
</style>
