// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router'

// views
import LoginView from '../views/LoginView.vue'
import ConversationsView from '../views/ConversationsView.vue'
import ChatView from '../views/ChatView.vue'
import GroupView from '../views/GroupView.vue'
import ProfileView from '../views/ProfileView.vue'

const routes = [
  // 根路径：按是否登录进行动态重定向
  {
    path: '/',
    redirect: () =>
      (sessionStorage.getItem('authToken') || localStorage.getItem('token'))
        ? '/conversations'
        : '/login',
  },

  // 快速自检页（开发调试用，可删）
  { path: '/__check', component: { template: '<div style="padding:12px">Router OK ✅</div>' } },

  // 公开页
  { path: '/login', name: 'login', component: LoginView },

  // 受保护页
  {
    path: '/conversations',
    name: 'conversations',
    component: ConversationsView,
    meta: { requiresAuth: true },
  },
  {
    path: '/groups',
    name: 'groups',
    component: GroupView,
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'profile',
    component: ProfileView,
    meta: { requiresAuth: true },
  },

  // 聊天页（会话 / 私聊 / 群聊）
  {
    path: '/chat/:type/:id',
    name: 'chat',
    component: ChatView,
    props: true,
    meta: { requiresAuth: true },
  },

  // 兼容跳转（旧路径适配）
  {
    path: '/conversations/:id',
    redirect: to => ({ name: 'chat', params: { type: 'conv', id: to.params.id } }),
  },
  { path: '/new-conversation', redirect: '/conversations' },
  { path: '/new-group', redirect: '/groups' },
  { path: '/me', redirect: '/profile' },

  // 兜底
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

// ✅ 登录检测：兼容 sessionStorage + localStorage
const isAuthed = () =>
  !!(sessionStorage.getItem('authToken') || localStorage.getItem('token'))

router.beforeEach((to) => {
  const needsAuth = to.matched.some(r => r.meta?.requiresAuth)

  // 受保护：未登录 -> 登录页，并附带回跳地址
  if (needsAuth && !isAuthed()) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  // 已登录：不允许再进登录页
  if (to.name === 'login' && isAuthed()) {
    return { name: 'conversations' }
  }

  return true
})

export default router
