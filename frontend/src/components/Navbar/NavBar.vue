<script setup lang="ts">
import { NAVBAR_ICON_SIZE, THEMES } from "@/components/Navbar/constants";
import { useThemeStore } from "@/stores/theme";
import { useUserStore } from "@/stores/user";
import { checkIfAuth } from "@/utils/auth";
import { removeUserToken } from "@/utils/cookies";
import {
  PhGear,
  PhPalette,
  PhSignOut,
  PhTextIndent,
  PhUser,
} from "@phosphor-icons/vue";
import { onMounted } from "vue";
import { useRouter } from "vue-router";

const { theme, getTheme, setTheme } = useThemeStore();

const { getUser } = useUserStore();
const user = await getUser();

const router = useRouter();

const handleLogout = () => {
  removeUserToken();
  router.push("/login");
};

const handleTheme = (theme: string) => {
  setTheme(theme);
}

const handleProfile = () => {
}

const handleSettings = () => {
}

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
        <!-- Theme dropdown -->
        <div class="flex items-center">
          <div class="dropdown dropdown-end">
            <div tabindex="0" role="button" class="btn btn-ghost rounded-btn">
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
              tabindex="0"
              class="menu dropdown-content z-[1] bg-base-200 rounded-box mt-4 w-96 overflow-y-auto p-2 shadow"
              style="max-height: calc(100vh - 10rem)"
            >
              <li v-for="(theme, index) in THEMES" :key="index">
                <a @click="handleTheme(theme)">
                  {{ theme.charAt(0).toUpperCase() + theme.slice(1) }}
                </a>
              </li>
            </ul>
          </div>

          <!-- User dropdown -->
          <div class="flex items-center">
            <div class="dropdown dropdown-end ml-4">
              <div
                tabindex="0"
                role="button"
                class="btn btn-ghost btn-circle avatar"
              >
                <div class="w-10 rounded-full">
                  <img
                    src="https://api.dicebear.com/9.x/personas/svg?seed=Jack"
                    alt="User profile"
                  />
                </div>
              </div>
              <div
                tabindex="0"
                class="dropdown-content z-[1] menu p-2 shadow bg-base-200 rounded-box w-52"
              >
                <div class="px-4 py-2 text-center">
                  <div class="font-bold">{{ user.name }}</div>
                  <div class="text-sm opacity-50">{{ user.email }}</div>
                </div>
                <div class="divider my-0"></div>
                <li>
                  <a
                    ><PhUser
                      class="mr-2"
                      :size="NAVBAR_ICON_SIZE"
                      weight="duotone"
                      @click="handleProfile"
                    />Profile</a
                  >
                </li>
                <li>
                  <a
                    ><PhGear
                      class="mr-2"
                      :size="NAVBAR_ICON_SIZE"
                      weight="duotone"
                      @click="handleSettings"
                    />Settings</a
                  >
                </li>
                <li>
                  <a @click="handleLogout"
                    ><PhSignOut
                      class="mr-2"
                      :size="NAVBAR_ICON_SIZE"
                      weight="duotone"
                    />Logout</a
                  >
                </li>
              </div>
            </div>
          </div>
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
