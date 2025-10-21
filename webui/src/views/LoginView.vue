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
        />

        <button
          type="submit"
          class="btn btn-gradient btn-lg w-100 mt-3"
          :disabled="busy || (name?.length ?? 0) < 1"
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
import ErrorMsg from '../components/ErrorMsg.vue'
import api from '../services/api'  // 继续用你自己的 api.js（默认导出里含 doLogin）

const router = useRouter()
const route = useRoute()

const name = ref('')
const err = ref('')
const busy = ref(false)

onMounted(() => {
  // 登录页浅色背景
  document.body.classList.add('theme-light-login')

  // 已登录则跳转
  const authed = !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))
  if (authed) {
    router.replace('/conversations')
  }
})

onUnmounted(() => {
  document.body.classList.remove('theme-light-login')
})

async function login () {
  err.value = ''
  if (!name.value) return
  busy.value = true
  try {
    // 后端登录（内部已保存 token / authToken）
    await api.doLogin(name.value)

    // 显示用途：保存用户名（顶栏依赖这个键）
    localStorage.setItem('username', name.value)
    // 保留你之前的 name 键（如果其它页面用了）
    localStorage.setItem('name', name.value || '')

    // 可选：给 me 做个最小缓存，便于其它页面兜底读取
    const me = { username: name.value }
    localStorage.setItem('me', JSON.stringify(me))

    // 同标签页通知 App.vue 立即刷新用户名
    window.dispatchEvent(new Event('auth:changed'))

    // 回跳
    const next = (route.query.redirect && String(route.query.redirect)) || '/conversations'
    router.replace(next)
  } catch (e) {
    err.value = e?.response?.data?.error || e?.response?.data?.message || e?.message || 'Login failed'
  } finally {
    busy.value = false
  }
}
</script>

<!-- 不加 scoped，覆盖 body 背景 -->
<style>
body.theme-light-login{
  background:
    radial-gradient(1200px 800px at 10% -10%, #f3f8ff 0, transparent 60%),
    radial-gradient(1000px 700px at 110% 0%, #eef6ff 0, transparent 55%),
    linear-gradient(180deg, #ffffff, #f7fafe) !important;
  color: #0f172a;
}
</style>

<!-- 组件内样式 -->
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
