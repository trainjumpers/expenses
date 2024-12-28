import {
  getTheme as getCookieTheme,
  setTheme as setCookieTheme,
} from "@/utils/cookies";
import { defineStore } from "pinia";
import { ref } from "vue";

export const useThemeStore = defineStore("theme", () => {
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

  return { theme, setTheme, getTheme };
});
