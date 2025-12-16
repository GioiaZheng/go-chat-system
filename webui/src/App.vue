<template>
  <div class="app-shell">
    <RouterView v-slot="{ Component, route }">
      <template v-if="!route.meta?.hideSidebar && isAuthed">
        <AppLayout>
          <div v-if="checking" class="checking">
            <div class="spinner"></div>
            <p>Checking session…</p>
          </div>

          <div v-else-if="!isAuthed" class="unauth">
            <p>You are not logged in. Redirecting to login…</p>
          </div>

          <component v-else :is="Component" />
        </AppLayout>
      </template>

      <template v-else>
        <div v-if="checking" class="checking">
          <div class="spinner"></div>
          <p>Checking session…</p>
        </div>

        <div v-else-if="!isAuthed && route.path !== '/login'" class="unauth">
          <p>You are not logged in. Redirecting to login…</p>
        </div>

        <component v-else :is="Component" />
      </template>
    </RouterView>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'

const route = useRoute()
const router = useRouter()

/* Track authentication state across navigation. */
const hasToken = () =>
  !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

const isAuthed = ref(false)
const checking = ref(true)

function refreshAuth() {
  isAuthed.value = hasToken()
}

/* Lifecycle hooks manage authentication redirects and listeners. */
onMounted(() => {
  refreshAuth()

  // Redirect unauthenticated visitors to the login page.
  if (!hasToken() && route.path !== '/login') {
    router.replace('/login')
  }

  // Listen for login/logout events emitted elsewhere in the app.
  window.addEventListener('auth:changed', refreshAuth)

  // Refresh auth status when the page becomes visible again.
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) refreshAuth()
  })

  checking.value = false
})

onBeforeUnmount(() => {
  window.removeEventListener('auth:changed', refreshAuth)
})

/* Validate authentication whenever the route changes. */
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

</script>

<style>
.app-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Unauthenticated notice */
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

/* Loading state while auth is checked */
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
</style>
