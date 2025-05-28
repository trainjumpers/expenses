"use client";

/**
 * Retrieves the value of a cookie by its name.
 *
 * @param name - The name of the cookie to retrieve.
 * @returns The cookie value if found, or `undefined` if the cookie does not exist.
 */
export function getCookie(name: string): string | undefined {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(";").shift();
}

/**
 * Sets a browser cookie with the specified name and value.
 *
 * By default, the cookie is set with path `/`, `SameSite=Strict`, and the `Secure` attribute unless explicitly disabled. Optional parameters allow customization of the cookie's maximum age, security, and SameSite policy.
 *
 * @param name - The name of the cookie to set.
 * @param value - The value to assign to the cookie.
 * @param options.maxAge - Optional. The lifetime of the cookie in seconds.
 * @param options.secure - Optional. If `false`, omits the `Secure` attribute; otherwise, the cookie is set as secure.
 * @param options.sameSite - Optional. Controls the `SameSite` attribute; defaults to `"Strict"`.
 */
export function setCookie(
  name: string,
  value: string,
  options: {
    maxAge?: number;
    secure?: boolean;
    sameSite?: "Strict" | "Lax" | "None";
  } = {}
) {
  let cookie = `${name}=${value}`;
  if (options.maxAge) {
    cookie += `; max-age=${options.maxAge}`;
  }
  cookie += "; path=/";
  const sameSite = options.sameSite || "Strict";
  cookie += `; SameSite=${sameSite}`;
  if (options.secure !== false) {
    cookie += "; Secure";
  }
  document.cookie = cookie;
}

/**
 * Deletes a cookie by name, expiring it immediately.
 *
 * The cookie is removed by setting its value to empty, `max-age` to 0, and applying `path=/`, `SameSite=Strict`, and `Secure` attributes.
 *
 * @param name - The name of the cookie to delete.
 */
export function deleteCookie(name: string) {
  document.cookie = `${name}=; max-age=0; path=/; SameSite=Strict; Secure`;
}
