<!-- Notes:
     - Centered, polished login screen inspired by a green/eco theme.
     - Kept backend contract minimal: POST /session { name } only (no password field).
     - Stores the returned token in localStorage and redirects to /conversations.
     - Uses <ErrorMsg/> for normalized error display and preserves existing axios service.
     - Pure CSS inside <style scoped>, no global changes required.
-->

<template>
  <div class="page">
    <div class="container">
      <div class="login-card">
        <div class="topbar"></div>

        <div class="logo">
          <i class="fas fa-leaf"></i>
          <h1>自然生态</h1>
          <p>欢迎回来，请登录您的账户</p>
        </div>

        <ErrorMsg v-if="err" :text="err" class="mb-12" />

        <form @submit.prevent="login" class="form" novalidate>
          <div class="input-group">
            <i class="fas fa-user"></i>
            <input
              v-model="name"
              type="text"
              placeholder="用户名（示例：alice）"
              autocomplete="username"
              required
            />
          </div>

          <button class="btn-login" :disabled="busy">
            <span v-if="busy" class="spinner" aria-hidden="true"></span>
            {{ busy ? '登录中…' : '登录' }}
          </button>
        </form>

        <div class="separator"><span>或</span></div>

        <div class="social-login" aria-hidden="true">
          <div class="social-btn"><i class="fab fa-google"></i></div>
          <div class="social-btn"><i class="fab fa-facebook-f"></i></div>
          <div class="social-btn"><i class="fab fa-twitter"></i></div>
        </div>

        <p class="hint">提示：课堂环境中仅需输入名字（例如 <strong>alice</strong>）即可登录。</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/services/axios'
import ErrorMsg from '@/components/ErrorMsg.vue'

const router = useRouter()
const name = ref('alice')
const err = ref('')
const busy = ref(false)

async function login() {
  err.value = ''
  const n = name.value.trim()
  if (!n) {
    err.value = '请输入用户名'
    return
  }
  busy.value = true
  try {
    // Backend contract: POST /session { name }
    const resp = await axios.post('/session', { name: n })
    const token = resp?.data?.data?.token
    if (!token) throw new Error('No token in response')
    localStorage.setItem('token', token)
    await router.push('/conversations')
  } catch (e) {
    err.value = e?.response?.data?.message || e.message || 'Network error'
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
/* Layout */
.page {
  min-height: 100vh;
  background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.container {
  width: 100%;
  max-width: 420px;
}
.login-card {
  position: relative;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 10px 30px rgba(56, 142, 60, 0.15);
  padding: 40px 35px;
  backdrop-filter: blur(10px);
  animation: fadeIn 0.8s ease-out;
  overflow: hidden;
}
.topbar {
  position: absolute;
  top: 0; left: 0;
  width: 100%; height: 6px;
  background: linear-gradient(90deg, #4caf50, #2e7d32);
}

/* Logo / heading */
.logo { text-align: center; margin-bottom: 28px; }
.logo i { font-size: 48px; color: #2e7d32; margin-bottom: 12px; }
.logo h1 { color: #1b5e20; font-weight: 600; font-size: 28px; letter-spacing: .5px; }
.logo p { color: #689f38; font-size: 15px; margin-top: 8px; }

/* Form */
.form { display: grid; gap: 16px; }
.input-group { position: relative; }
.input-group i {
  position: absolute; left: 18px; top: 50%; transform: translateY(-50%);
  color: #4caf50; font-size: 18px;
}
.input-group input {
  width: 100%; padding: 16px 20px 16px 50px;
  border: 1px solid #e0e0e0; background: #f9f9f9; border-radius: 10px;
  outline: none; font-size: 16px; color: #333;
  transition: all .3s ease;
}
.input-group input:focus {
  border-color: #4caf50; background: #fff;
  box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.1);
}
.input-group input::placeholder { color: #9e9e9e; }

/* Button */
.btn-login {
  width: 100%; padding: 14px 16px; border: none; cursor: pointer;
  background: linear-gradient(135deg, #4caf50 0%, #2e7d32 100%);
  color: #fff; border-radius: 10px; font-size: 16px; font-weight: 600;
  letter-spacing: 0.5px;
  transition: transform .25s ease, box-shadow .25s ease, opacity .2s ease;
}
.btn-login:hover { transform: translateY(-2px); box-shadow: 0 6px 15px rgba(76,175,80,.3); }
.btn-login:active { transform: translateY(0); box-shadow: 0 3px 8px rgba(76,175,80,.3); }
.btn-login[disabled] { opacity: .7; cursor: not-allowed; }

/* Spinner (pure CSS, no dependency) */
@keyframes spin { to { transform: rotate(360deg); } }
.spinner {
  display: inline-block; width: 1em; height: 1em; margin-right: .5rem;
  border: 2px solid rgba(255,255,255,.6);
  border-top-color: #fff; border-radius: 50%;
  animation: spin 1s linear infinite;
  vertical-align: -2px;
}

/* Divider + socials */
.separator {
  display: flex; align-items: center; justify-content: center;
  gap: 16px; color: #9e9e9e; margin: 22px 0;
}
.separator::before, .separator::after {
  content: ''; flex: 1; border-bottom: 1px solid #e0e0e0;
}
.separator span { font-size: 14px; }
.social-login { display: flex; justify-content: center; gap: 12px; margin-bottom: 10px; }
.social-btn {
  width: 48px; height: 48px; border-radius: 50%;
  display: grid; place-items: center;
  background: #f5f5f5; border: 1px solid #eee; color: #4caf50;
  transition: transform .2s ease, box-shadow .2s ease;
}
.social-btn:hover { transform: translateY(-3px); box-shadow: 0 5px 10px rgba(0,0,0,.08); }
.social-btn i { font-size: 18px; }

/* Hint */
.hint {
  margin-top: 8px; text-align: center; color: #616161; font-size: 13px;
}

/* Animations + responsive */
@keyframes fadeIn { from { opacity: 0; transform: translateY(20px);} to { opacity: 1; transform: translateY(0);} }
@media (max-width: 480px) {
  .login-card { padding: 30px 25px; }
}
</style>
