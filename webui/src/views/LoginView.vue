<!-- src/views/LoginView.vue: Sign-in form with minimal session bootstrap. -->
<template>
  <div class="auth-wrap">
    <div class="auth-card">
      <div class="brand">WASA <span class="grad">Chat</span></div>
      <h1 class="subtitle">Sign in</h1>

      <ErrorMsg v-if="err" :text="err" class="mb-3" />

      <form @submit.prevent="login" novalidate>
        <label class="form-label field-label">Name</label>
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
          @input="onUserInput"
        />
        <p class="field-helper small text-muted mb-0">
          This name is shown to others when you start chats.
        </p>
        <p v-if="nameError" class="text-danger small mt-2 mb-0">
          {{ nameError }}
        </p>

        <button
          type="submit"
          class="btn btn-gradient btn-lg w-100 mt-2"
          :class="{ 'is-loading': busy }"
          :disabled="busy || name.length < 1"
          :aria-busy="busy"
        >
          <span
            v-if="busy"
            class="spinner-border spinner-border-sm me-2"
            role="status"
            aria-hidden="true"
          ></span>
          {{ busy ? 'Logging in…' : 'Login' }}
        </button>
      </form>

      <p class="text-center login-hint mb-0">
        Press <kbd>Enter</kbd> to login
      </p>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ErrorMsg from '@/components/ErrorMsg.vue'
import { ensureAuthReady, isAuthenticated, login as performLogin } from '@/services/auth'

const router = useRouter()
const route = useRoute()

const name = ref('')
const err  = ref('')
const nameError = ref('')
const busy = ref(false)

onMounted(async () => {
  document.body.classList.add('theme-light-login')

  // Prefill the input when a name query parameter is provided.
  const preset = String(route.query.name || '').trim()
  if (preset) name.value = preset

  // Skip the form when a valid token is already stored.
  await ensureAuthReady()
  if (isAuthenticated.value) router.replace('/conversations')
})

onUnmounted(() => {
  document.body.classList.remove('theme-light-login')
})

function onUserInput() {
  if (err.value || nameError.value) {
    err.value = ''
    nameError.value = ''
  }
}

async function login () {
  if (busy.value) return
  if (!name.value) {
    nameError.value = 'Please enter a name to continue.'
    return
  }
  err.value = ''
  nameError.value = ''
  busy.value = true

  try {
    await performLogin(name.value)

    // Honor a redirect query parameter when present.
    const next = (route.query.redirect && String(route.query.redirect)) || '/conversations'
    router.replace(next)
  } catch (e) {
    const status = e?.response?.status
    if (status === 409) {
      err.value = 'Name already in use.'
    } else if (status === 400 || status === 422) {
      err.value = 'Invalid name.'
    } else if (status === 500) {
      err.value = 'Server error — try again later.'
    } else {
      err.value = 'Login failed — please try again.'
    }
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
  align-items: flex-start;
  justify-content: center;
  padding: 24px 16px 32px;
}
.auth-card{
  width: 100%;
  max-width: 460px;
  background: #fff;
  border: 1px solid #dbe4f0;
  border-radius: 0;
  box-shadow: 0 24px 70px rgba(2,6,23,.12);
  padding: 28px 24px;
  color: #0f172a;
  margin-top: 16px;
}
.brand{
  font-weight: 700;
  font-size: 22px;
  letter-spacing: .4px;
  text-align: center;
  color: #1f2937;
}
.brand .grad{
  background: linear-gradient(90deg, #22c55e, #3b82f6);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}
.subtitle{
  text-align: center;
  color: #0f172a;
  font-weight: 800;
  font-size: 32px;
  margin: 8px 0 18px;
}
.form-label{ color:#334155 }
.field-label{
  text-transform: uppercase;
  font-size: .72rem;
  letter-spacing: .08em;
  font-weight: 700;
  margin-bottom: 6px;
  color: #64748b;
}
.field-helper{
  margin-top: 6px;
  color: #94a3b8;
}
.form-control{
  background:#fff;
  border-color: #cbd5e1;
  color:#0f172a;
  transition: border-color .2s ease, box-shadow .2s ease;
}
.form-control::placeholder{ color:#94a3b8 }
.form-control:focus{
  border-color:#22c55e;
  box-shadow: 0 0 0 .25rem rgba(34,197,94,.18);
}
.btn-gradient{
  background-image: linear-gradient(135deg, #22c55e 0%, #16a34a 45%, #3b82f6 120%);
  color:#fff;
  border:0;
  box-shadow:0 .6rem 1.4rem rgba(34,197,94,.25);
  transition: transform .15s ease, box-shadow .2s ease, filter .2s ease;
}
.btn-gradient:hover{
  filter: brightness(1.03);
  box-shadow: 0 .75rem 1.6rem rgba(34,197,94,.3);
}
.btn-gradient:active{
  transform: translateY(1px);
  box-shadow: 0 .4rem 1rem rgba(34,197,94,.2);
}
.btn-gradient.is-loading{
  cursor: progress;
  filter: saturate(.9);
}
.btn-gradient:disabled{
  opacity:.55;
  box-shadow:none;
  cursor: not-allowed;
}
.login-hint{
  margin-top: 14px;
  font-size: .8rem;
  color: #94a3b8;
}
</style>