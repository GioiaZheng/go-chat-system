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
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'
import { ensureAuthReady, isAuthenticated, isReady } from '@/services/auth'


const route = useRoute()
const router = useRouter()

const checking = computed(() => !isReady.value)
const isAuthed = isAuthenticated

ensureAuthReady()

watch(
  () => [route.fullPath, isAuthenticated.value, isReady.value],
  () => {
    if (isReady.value && !isAuthenticated.value && route.path !== '/login') {
      const query = route.fullPath ? { redirect: route.fullPath } : undefined
      router.replace({ path: '/login', query })
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
