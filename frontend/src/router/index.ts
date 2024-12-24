import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: () => import("../views/Home/HomeView.vue"),
    },
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
