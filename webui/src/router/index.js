// src/router/index.js
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
    name: 'Login',
    component: LoginView
  },
  {
    path: '/conversations',
    name: 'Conversations',
    component: ConversationsView
  },
  {
    path: '/groups',
    name: 'Groups',
    component: GroupsView
  },
  {
    path: '/me',
    name: 'Profile',
    component: ProfileView
  },
  {
    path: '/chat/:type/:id',
    name: 'Chat',
    component: ChatView,
    props: true
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫 - 检查认证
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/conversations')
  } else {
    next()
  }
})

export default router