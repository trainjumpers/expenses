const setCookie = (name: string, value: string) => {
  if (!name || !value) return;
  document.cookie = `${name}=${value}; path=/; max-age=${60 * 60 * 24 * 7}`;
};

const getCookie = (name: string) => {
  const nameEQ = `${name}=`;
  if (typeof document === "undefined") return null;
  const ca = document.cookie.split(";");
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) === " ") c = c.substring(1, c.length);
    if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
  }
  return null;
};

/**
 * Gets the user's preferred theme from a cookie, or uses the system's
 * preferred color scheme if no cookie is found.
 *
 * @returns The user's preferred theme, either 'light' or 'dark'.
 */
export const getTheme = (isClient: boolean) => {
  if (!isClient) return "light";
  const cookieTheme = getCookie("theme");
  if (cookieTheme) {
    return cookieTheme;
  }
  if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
    return "dark";
  }
  return "light";
};

/**
 * Sets a cookie to save the user's preferred theme.
 *
 * @param theme The desired theme, either 'light' or 'dark'.
 */
export const setTheme = (theme: string) => {
  document.documentElement.setAttribute("data-theme", theme);
  setCookie("theme", theme);
};

export const setUserToken = (token: string) => {
  setCookie("token", token);
};

export const getUserToken = () => {
  return getCookie("token");
};
