import { createRouter, createWebHistory } from 'vue-router'

// imports are from src/router/ -> back to src/
import LoginView from '../views/LoginView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import ChatView from '../views/ChatView.vue'
import GroupView from '../views/GroupView.vue'
import ProfileView from '../views/ProfileView.vue'

const routes = [
  // Root -> login or conversations
  { path: '/', redirect: () => (localStorage.getItem('token') ? '/conversations' : '/login') },

  // Quick check route: if you see this text, router/app mounted OK
  { path: '/__check', component: { template: '<div style="padding:12px">Router OK ✅</div>' } },

  { path: '/login', name: 'login', component: LoginView },
  { path: '/conversations', name: 'conversations', component: ConversationsView, meta: { requiresAuth: true } },
  { path: '/groups', name: 'groups', component: GroupView, meta: { requiresAuth: true } },
  { path: '/profile', name: 'profile', component: ProfileView, meta: { requiresAuth: true } },

  // Chat route (conv | private | group)
  { path: '/chat/:type/:id', name: 'chat', component: ChatView, props: true, meta: { requiresAuth: true } },

  // Compatibility redirects
  { path: '/conversations/:id', redirect: to => ({ name: 'chat', params: { type: 'conv', id: to.params.id } }) },
  { path: '/new-conversation', redirect: '/conversations' },
  { path: '/new-group', redirect: '/groups' },
  { path: '/me', redirect: '/profile' },

  // Fallback
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta?.requiresAuth && !token) return next('/login')
  if (to.name === 'login' && token)    return next('/conversations')
  next()
})

export default router
