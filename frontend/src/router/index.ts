import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("../views/Auth/login/LoginView.vue"),
    },
    {
      path: "/register",
      name: "register",
      component: () => import("../views/Auth/signup/SignupView.vue"),
    },
  ],
});

export default router;
