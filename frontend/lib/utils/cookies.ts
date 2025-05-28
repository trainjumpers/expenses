"use client";

export function getCookie(name: string): string | undefined {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(";").shift();
}

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

export function deleteCookie(
  name: string,
  options: {
    secure?: boolean;
    sameSite?: "Strict" | "Lax" | "None";
  } = {}
) {
  let cookie = `${name}=; max-age=0; path=/`;
  const sameSite = options.sameSite || "Strict";
  cookie += `; SameSite=${sameSite}`;
  if (options.secure !== false) {
    cookie += "; Secure";
  }
  document.cookie = cookie;
}
