<template>
  <div class="app-shell" :class="{ 'no-sidebar': hideSidebar }">
    <!-- ✅ 已登录才显示 Sidebar -->
    <Sidebar v-if="!hideSidebar" />

    <main class="flex-fill p-3">
      <!-- ✅ 未登录时显示过渡提示 -->
      <div v-if="checking" class="checking">
        <div class="spinner"></div>
        <p>Checking session…</p>
      </div>

      <div v-else-if="!isAuthed && route.path !== '/login'" class="unauth">
        <p>You are not logged in. Redirecting to login…</p>
      </div>

      <!-- ✅ 正常内容 -->
      <RouterView v-else />
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Sidebar from '@/components/Sidebar.vue'

const route = useRoute()
const router = useRouter()

/* ---------- Auth 状态检测 ---------- */
const hasToken = () =>
  !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

const isAuthed = ref(false)
const checking = ref(true)

function refreshAuth() {
  isAuthed.value = hasToken()
}

/* ---------- 生命周期 ---------- */
onMounted(() => {
  refreshAuth()

  // 检查当前路由：若未登录则跳转 /login
  if (!hasToken() && route.path !== '/login') {
    router.replace('/login')
  }

  // 监听登录/登出事件
  window.addEventListener('auth:changed', refreshAuth)

  // 页面可见时自动刷新状态
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) refreshAuth()
  })

  checking.value = false
})

onBeforeUnmount(() => {
  window.removeEventListener('auth:changed', refreshAuth)
})

/* ---------- 路由变化时再验证一次 ---------- */
watch(
  () => route.fullPath,
  () => {
    refreshAuth()
    if (!hasToken() && route.path !== '/login') {
      router.replace('/login')
    }
  },
  { immediate: true }
)

/* ---------- Sidebar 是否隐藏 ---------- */
const hideSidebar = computed(() => !!route.meta?.hideSidebar || !isAuthed.value)
</script>

<style>
/* 布局控制 */
.app-shell {
  display: flex;
  min-height: 100vh;
}

.app-shell main {
  flex: 1 1 auto;
  min-width: 0;
}

.no-sidebar main {
  width: 100%;
}

/* 未登录提示 */
.unauth {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
  color: #64748b;
  font-weight: 500;
  font-size: 1rem;
}

/* 检查中过渡状态 */
.checking {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
  color: #64748b;
  gap: 10px;
}

.spinner {
  width: 1.2rem;
  height: 1.2rem;
  border: 3px solid rgba(100,116,139,.25);
  border-top-color: #22c55e;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 992px) {
  .app-shell {
    flex-direction: column;
  }

  .app-shell main {
    width: 100%;
    padding: 1rem !important;
  }
}
</style>
