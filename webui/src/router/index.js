// English: routes + auth guard
import { createRouter, createWebHistory } from "vue-router";

const LoginView = () => import("../views/LoginView.vue");
const ConversationsView = () => import("../views/ConversationsView.vue");
const ChatView = () => import("../views/ChatView.vue");
const GroupView = () => import("../views/GroupView.vue");
const ProfileView = () => import("../views/ProfileView.vue");

const routes = [
  { path: "/login", name: "login", component: LoginView },
  { path: "/conversations", name: "conversations", component: ConversationsView },
  // type: conv | private | group
  { path: "/chat/:type/:id", name: "chat", component: ChatView, props: true },
  { path: "/groups", name: "groups", component: GroupView },
  { path: "/me", name: "me", component: ProfileView },
  { path: "/:pathMatch(.*)*", redirect: "/conversations" },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const token = localStorage.getItem("token");
  if (!token && to.name !== "login") return { name: "login" };
  return true;
});

export default router;
