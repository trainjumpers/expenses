"use client";

export function getCookie(name: string): string | undefined {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(";").shift();
}

export function setCookie(
  name: string,
  value: string,
  options: { maxAge?: number } = {}
) {
  let cookie = `${name}=${value}`;
  if (options.maxAge) {
    cookie += `; max-age=${options.maxAge}`;
  }
  cookie += "; path=/";
  document.cookie = cookie;
}

export function deleteCookie(name: string) {
  document.cookie = `${name}=; max-age=0; path=/`;
}
