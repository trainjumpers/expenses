<script setup lang="ts">
import { NAVBAR_ICON_SIZE, THEMES } from "@/constants/navbar";
import { checkIfAuth } from "@/utils/auth";
import {
  getTheme as getCookieTheme,
  setTheme as setCookieTheme,
} from "@/utils/cookies";
import { PhPalette, PhTextIndent } from "@phosphor-icons/vue";
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

// Composables
const router = useRouter();
const theme = ref("light");

const setTheme = (newTheme: string) => {
  theme.value = newTheme;
  document.documentElement.setAttribute("data-theme", newTheme);
  setCookieTheme(newTheme);
};

const getTheme = () => {
  theme.value = getCookieTheme();
  document.documentElement.setAttribute("data-theme", theme.value);
};

onMounted(() => {
  getTheme();
  if (!checkIfAuth()) {
    router.push("/login");
  }
});
</script>

<template>
  <div className="drawer">
    <input id="navbar-drawer" type="checkbox" className="drawer-toggle" />
    <div className="drawer-content flex flex-col">
      <div className="navbar bg-base-300 w-full">
        <div className="flex-none">
          <label
            htmlFor="navbar-drawer"
            aria-label="open sidebar"
            className="btn btn-square btn-ghost"
          >
            <PhTextIndent :size="NAVBAR_ICON_SIZE" weight="duotone" />
          </label>
        </div>
        <div className="mx-2 flex-1 px-2 text-xl font-bold">
          Expense Tracker
        </div>
        <div className="dropdown dropdown-end">
          <div tabindex="0" role="button" className="btn btn-ghost rounded-btn">
            <PhPalette :size="NAVBAR_ICON_SIZE" weight="duotone" />
            <div>
              {{
                theme
                  .split(" ")
                  .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
                  .join(" ")
              }}
            </div>
          </div>
          <ul
            tabIndex="0"
            className="menu z-[1] dropdown-content bg-base-200 rounded-box mt-4 w-96 overflow-y-auto p-2 shadow"
            style="max-height: calc(100vh - 10rem)"
          >
            <li v-for="(theme, index) in THEMES" :key="index">
              <a @click="setTheme(theme)">
                {{ theme.charAt(0).toUpperCase() + theme.slice(1) }}
              </a>
            </li>
          </ul>
        </div>
      </div>
    </div>
    <div className="drawer-side">
      <label
        htmlFor="navbar-drawer"
        aria-label="close sidebar"
        className="drawer-overlay"
      ></label>
      <ul className="menu bg-base-200 min-h-full w-80 p-4">
        <li><a>Sidebar Item 1</a></li>
        <li><a>Sidebar Item 2</a></li>
      </ul>
    </div>
  </div>
</template>
