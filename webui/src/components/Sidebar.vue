<!-- src/components/Sidebar.vue -->
<template>
  <nav class="sidebar">
    <!-- Brand -->
    <div class="brand">Wa<span class="accent">SA</span>Text</div>

    <!-- Navigation -->
    <div class="nav">
      <RouterLink to="/conversations" class="link" active-class="active">💬 Chats</RouterLink>
      <RouterLink to="/contacts" class="link" active-class="active">👥 Contacts</RouterLink>
      <RouterLink to="/new-group" class="link" active-class="active">🗂 New Group</RouterLink>
      <RouterLink to="/profile" class="link" active-class="active">🪪 Profile</RouterLink>
    </div>

    <!-- Footer -->
    <div class="footer">
      <button class="logout" @click="logout">🚪 Logout</button>
    </div>
  </nav>
</template>

<script setup>
import { useRouter } from 'vue-router'
import api from '@/services/api' // ✅ 使用统一的 doLogout

const router = useRouter()

function logout() {
  api.doLogout()               // 清理 token、name、me，并广播 auth:changed
  router.replace('/login')     // 回到登录页
}
</script>

<style scoped>
.sidebar {
  width: 230px;
  min-height: 100vh;
  background: linear-gradient(180deg, #ffffff, #f8fafc);
  border-right: 1px solid #e2e8f0;
  padding: 18px 14px;
  display: flex;
  flex-direction: column;
  box-shadow: 4px 0 12px rgba(0,0,0,0.03);
}

.brand {
  font-weight: 800;
  font-size: 1.4rem;
  letter-spacing: .4px;
  text-align: center;
  color: #0f172a;
  margin-bottom: 24px;
}
.accent { color: #3b82f6; }

.nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-grow: 1;
}

.link {
  display: block;
  padding: 10px 12px;
  border-radius: 10px;
  color: #0f172a;
  font-weight: 500;
  text-decoration: none;
  transition: 0.2s;
}
.link:hover {
  background: #f1f5f9;
}
.link.active {
  background: #3b82f620;
  color: #2563eb;
  font-weight: 600;
}

.footer {
  border-top: 1px solid #e2e8f0;
  padding-top: 16px;
  margin-top: auto;
}
.logout {
  width: 100%;
  border: 0;
  border-radius: 10px;
  padding: 8px 0;
  background: #ef4444;
  color: #fff;
  font-weight: 600;
  transition: 0.25s;
}
.logout:hover {
  background: #dc2626;
}
</style>
