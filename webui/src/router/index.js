import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import GroupsView from '../views/GroupView.vue'
import ProfileView from '../views/ProfileView.vue'
import ChatView from '../views/ChatView.vue'

const routes = [
  {
    path: '/',
    redirect: '/conversations'
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView
  },
  {
    path: '/conversations',
    name: 'conversations',
    component: ConversationsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/groups',
    name: 'groups',
    component: GroupsView,
    meta: { requiresAuth: true }
  },
  {
    path: '/me',
    name: 'profile',
    component: ProfileView,
    meta: { requiresAuth: true }
  },
  {
    path: '/chat/:type/:id',
    name: 'chat',
    component: ChatView,
    props: true,
    meta: { requiresAuth: true }
  },
  // --- Compatibility redirects (no new views required) ---
  {
    path: '/conversations/:id',
    redirect: to => `/chat/conv/${to.params.id}`
  },
  {
    path: '/new-conversation',
    redirect: '/conversations'
  },
  {
    path: '/new-group',
    redirect: '/groups'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Global navigation guard: check requiresAuth meta
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/conversations')
  } else {
    next()
  }
})

export default router
