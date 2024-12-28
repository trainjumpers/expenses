const THEME_COOKIE_NAME = "theme";
const THEME_COOKIE_EXPIRY_DAYS = 365;

const TOKEN_COOKIE_NAME = "token";
const TOKEN_COOKIE_EXPIRY_DAYS = 7;

/* ############################################################################### */
/* ############################################################################### */
/* ############################ THEME COOKIE ##################################### */
/* ############################################################################### */
/* ############################################################################### */
/**
 * Gets the user's preferred theme from a cookie, or uses the system's
 * preferred color scheme if no cookie is found.
 *
 * @returns The user's preferred theme, either 'light' or 'dark'.
 */
export const getTheme = () => {
  const cookieTheme = getCookie(THEME_COOKIE_NAME);
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
  setCookie(THEME_COOKIE_NAME, theme, THEME_COOKIE_EXPIRY_DAYS);
};

/* ############################################################################### */
/* ############################################################################### */
/* ############################ USER COOKIE ###################################### */
/* ############################################################################### */
/* ############################################################################### */

export const setUserToken = (token: string) => {
  setCookie(TOKEN_COOKIE_NAME, token, TOKEN_COOKIE_EXPIRY_DAYS);
};

export const getUserToken = () => {
  return getCookie(TOKEN_COOKIE_NAME);
};

export const removeUserToken = () => {
  removeCookie(TOKEN_COOKIE_NAME);
};

/* ############################################################################### */
/* ############################################################################### */
/* ############################ COOKIE UTIL ###################################### */
/* ############################################################################### */
/* ############################################################################### */

/**
 * Sets a cookie with the specified name and value.
 * The cookie will expire after 7 days.
 *
 * @param name The name of the cookie
 * @param value The value to store in the cookie
 */
const setCookie = (name: string, value: string, days = 7) => {
  if (!name || !value) return;
  document.cookie = `${name}=${value}; path=/; max-age=${60 * 60 * 24 * days}`;
};

/**
 * Gets the value of a cookie by its name.
 *
 * @param name The name of the cookie to retrieve
 * @returns The value of the cookie if found, null otherwise
 */
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
 * Removes a cookie with the specified name by setting its expiration
 * to a past date.
 *
 * @param name The name of the cookie to remove
 */
const removeCookie = (name: string) => {
  if (!name) return;
  document.cookie = `${name}=; path=/; max-age=0`;
};
