// src/router/index.js
import { createRouter, createWebHistory } from "vue-router";
import LoginView from "../views/LoginView.vue";
import ConversationsView from "../views/ConversationsView.vue";
import ChatView from "../views/ChatView.vue";
import GroupsView from "../views/GroupsView.vue";
import ProfileView from "../views/ProfileView.vue";

/**
 * Router configuration
 * - English comments for TA/teacher clarity.
 * - Defines assignment-required routes: login, conversations, chat, groups, profile.
 * - Includes a navigation guard that redirects to /login if no token is present.
 */

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: "/login", component: LoginView },
    { path: "/conversations", component: ConversationsView },
    { path: "/chat/:type/:id", component: ChatView, props: true },
    { path: "/groups", component: GroupsView },
    { path: "/me", component: ProfileView },
    { path: "/:pathMatch(.*)*", redirect: "/conversations" }, // fallback
  ],
});

// Navigation guard: check authentication before each route
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem("token");
  if (!token && to.path !== "/login") {
    // No token → must login
    next("/login");
  } else {
    next();
  }
});

export default router;
