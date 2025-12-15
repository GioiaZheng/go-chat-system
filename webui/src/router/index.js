import { createRouter, createWebHistory } from 'vue-router'

// View components used in the route map.
import LoginView from '../views/LoginView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import ContactsView from '../views/ContactsView.vue'
import ChatView from '../views/ChatView.vue'
import ProfileView from '../views/ProfileView.vue'
import NewGroupView from '../views/NewGroupView.vue'

// Routes grouped by access level and legacy redirects.
const routes = [
  // Root path redirects to login; the guard will handle auth flow.
  {
    path: '/',
    redirect: '/login',
  },

  // Lightweight route to verify router wiring during development.
  {
    path: '/__check',
    component: { template: '<div style="padding:12px">Router OK</div>' },
  },

  // Public routes
  {
    path: '/login',
    name: 'login',
    component: LoginView,
    meta: {
      public: true,
      hideSidebar: true, // Hide the sidebar on the login page.
    },
  },

  // Authenticated routes
  {
    path: '/conversations',
    name: 'conversations',
    component: ConversationsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/contacts',
    name: 'contacts',
    component: ContactsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/new-group',
    name: 'new-group',
    component: NewGroupView,
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'profile',
    component: ProfileView,
    meta: { requiresAuth: true },
  },

  // Chat page (private/group conversation)
  {
    path: '/chat/:type/:id',
    name: 'chat',
    component: ChatView,
    props: true,
    meta: { requiresAuth: true },
  },

  // Legacy and redirect helpers
  {
    path: '/conversations/:id',
    redirect: (to) => ({
      name: 'chat',
      params: { type: 'conv', id: to.params.id },
    }),
  },
  { path: '/new-conversation', redirect: '/contacts' },
  { path: '/groups', redirect: '/new-group' },
  { path: '/me', redirect: '/profile' },

  // Fallback for unknown routes
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

// Router instance with scroll reset behavior.
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

// Navigation guard enforcing authentication and login redirects.
const isAuthed = () =>
  !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

router.beforeEach((to, from, next) => {
  const needsAuth = to.matched.some((r) => r.meta?.requiresAuth)
  const authed = !!(
    sessionStorage.getItem('authToken') || localStorage.getItem('token')
  )

  // Step 1: Redirect unauthenticated users away from protected routes.
  if (needsAuth && !authed) {
    next({
      path: '/login',
      query: { redirect: to.fullPath },
    })
    return
  }

  // Step 2: Send authenticated users away from the login page.
  if (to.path === '/login' && authed) {
    next('/conversations')
    return
  }

  // Step 3: Emit a global event when the route actually changes (optional).
  if (from.name && from.name !== to.name) {
    try {
      window.dispatchEvent(new Event('auth:changed'))
    } catch {}
  }

  next()
})


export default router
