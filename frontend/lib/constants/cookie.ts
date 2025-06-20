export const ACCESS_TOKEN_EXPIRY =
  Number(process.env.NEXT_PUBLIC_ACCESS_TOKEN_EXPIRY) || 12 * 60 * 60; // 12 hours in seconds
export const REFRESH_TOKEN_EXPIRY =
  Number(process.env.NEXT_PUBLIC_REFRESH_TOKEN_EXPIRY) || 7 * 24 * 60 * 60; // 7 days in seconds
export const ACCESS_TOKEN_NAME = "access_token";
export const REFRESH_TOKEN_NAME = "refresh_token";
