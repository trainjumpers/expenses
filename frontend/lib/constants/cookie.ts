export const ACCESS_TOKEN_EXPIRY =
  Number(process.env.NEXT_PUBLIC_ACCESS_TOKEN_EXPIRY) || 60 * 60; // 1 hour in seconds
export const REFRESH_TOKEN_EXPIRY =
  Number(process.env.NEXT_PUBLIC_REFRESH_TOKEN_EXPIRY) || 90 * 24 * 60 * 60; // 90 days in seconds
export const ACCESS_TOKEN_NAME = "access_token";
export const REFRESH_TOKEN_NAME = "refresh_token";
