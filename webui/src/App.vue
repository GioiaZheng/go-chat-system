<!-- App.vue
     Dashboard shell ONLY after auth; plain login screen before auth.
     English-only UI & comments.
-->
<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const authed = ref(false)
const username = ref('')

// sync auth state from localStorage
function syncAuth() {
  const token = localStorage.getItem('token')
  authed.value = !!token
  try {
    const me = JSON.parse(localStorage.getItem('me') || '{}')
    username.value = me?.name || me?.username || ''
  } catch { username.value = '' }
}

function logout() {
  localStorage.clear()
  sessionStorage.clear()
  location.href = '/login'
}

onMounted(() => {
  syncAuth()
  window.addEventListener('storage', syncAuth)
})
onBeforeUnmount(() => window.removeEventListener('storage', syncAuth))
</script>

<template>
  <!-- BEFORE LOGIN: just render the route, no navbar, no sidebar -->
  <div v-if="!authed" class="min-vh-100 d-flex align-items-center justify-content-center bg-light">
    <router-view />
  </div>

  <!-- AFTER LOGIN: show the original Dashboard shell -->
  <div v-else>
    <header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow">
      <a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6" href="#/">WASAText</a>
      <div class="navbar-nav ms-auto">
        <div class="nav-item text-nowrap">
          <span class="nav-link px-3 text-white-50">Signed in as {{ username || 'user' }}</span>
        </div>
      </div>
    </header>

    <div class="container-fluid">
      <div class="row">
        <!-- hide the WHOLE sidebar when not logged in; here authed is true -->
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

        <main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">
          <router-view />
        </main>
      </div>
    </div>
  </div>
</template>

<style scoped>
.nav-link i { margin-right: 8px; }
</style>
