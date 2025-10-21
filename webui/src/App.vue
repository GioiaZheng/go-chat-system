<!-- App.vue -->
<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const authed = ref(false)
const username = ref('')

// 同步登录态 & 用户名
function syncAuth () {
  const token = localStorage.getItem('token') || sessionStorage.getItem('authToken')
  authed.value = !!token
  try {
    // 优先拿你在 Profile 保存下来的 username
    username.value =
      localStorage.getItem('username') ||
      JSON.parse(localStorage.getItem('me') || '{}')?.username ||
      localStorage.getItem('name') ||
      ''
  } catch { username.value = '' }
}

function logout () {
  localStorage.clear()
  sessionStorage.clear()
  router.replace('/login')
}

onMounted(() => {
  syncAuth()
  window.addEventListener('storage', syncAuth)
})
onBeforeUnmount(() => window.removeEventListener('storage', syncAuth))

// 不用外壳的页面（登录 / 注册）
const noShellRoutes = ['login', 'register']
</script>

<template>
  <!-- ① 登录页/注册页：不套外壳 -->
  <div v-if="noShellRoutes.includes(route.name)">
    <router-view />
  </div>

  <!-- ② 其他路由：未登录就直接只渲染当前页面（路由守卫也会重定向到 /login） -->
  <div v-else-if="!authed">
    <router-view />
  </div>

  <!-- ③ 已登录：统一外壳 -->
  <div v-else>
    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
      <div class="navbar-nav ms-auto">
        <div class="nav-item text-nowrap">
          <span class="nav-link px-3 text-white-50">
            Signed in as {{ username || 'user' }}
          </span>
        </div>
      </div>
    </header>

    <div class="container-fluid">
      <div class="row">
        <!-- 统一侧边栏 -->
        <nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-light sidebar collapse">
          <div class="position-sticky pt-3 sidebar-sticky">
            <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
              <span>Conversations</span>
            </h6>
            <ul class="nav flex-column">
              <li class="nav-item">
                <RouterLink to="/conversations" class="nav-link">
                  <i class="bi bi-chat-dots"></i> All Conversations
                </RouterLink>
              </li>
              <li class="nav-item">
                <RouterLink to="/groups" class="nav-link">
                  <i class="bi bi-people"></i> Groups
                </RouterLink>
              </li>
            </ul>

            <h6 class="sidebar-heading d-flex justify-content-between align-items-center px-3 mt-4 mb-1 text-muted text-uppercase">
              <span>Settings</span>
            </h6>
            <ul class="nav flex-column">
              <li class="nav-item">
                <RouterLink to="/profile" class="nav-link">
                  <i class="bi bi-person"></i> Profile
                </RouterLink>
              </li>
              <li class="nav-item">
                <a href="#" class="nav-link" @click.prevent="logout">
                  <i class="bi bi-box-arrow-right"></i> Logout
                </a>
              </li>
            </ul>
          </div>
        </nav>

        <!-- 受保护页面都在这里渲染（样式统一） -->
        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <router-view />
        </main>
      </div>
    </div>
  </div>
</template>

<style scoped>
.nav-link i { margin-right: 8px; }
/* 高亮当前路由（可选） */
.nav-link.router-link-active { font-weight: 600; color:#0f172a; }
</style>
